package template

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Node defines a single element within a template, such as text, variable, or control structure.
type Node struct {
	Type       string
	Text       string
	Variable   string
	Collection string
	Filters    []Filter
	Children   []*Node
	EndText    string
}

// Template represents a structured template that can be executed with a given context.
type Template struct {
	Nodes []*Node
}

// NewTemplate creates an empty template, ready to be populated with nodes.
func NewTemplate() *Template {
	return &Template{Nodes: []*Node{}}
}

// Execute combines template data with the provided context to produce a string.
func (t *Template) Execute(ctx Context) (string, error) {
	var builder strings.Builder
	forLayers := 0
	if err := executeNodes(t.Nodes, ctx, &builder, forLayers); err != nil {
		return builder.String(), err
	}
	return builder.String(), nil
}

// MustExecute combines template data with the provided context to produce a string, ignoring errors.
func (t *Template) MustExecute(ctx Context) string {
	result, _ := t.Execute(ctx)
	return result
}

// executeNodes recursively processes a slice of nodes, appending the result to the builder.
func executeNodes(nodes []*Node, ctx Context, builder *strings.Builder, forLayers int) error {
	var firstErr error
	for _, node := range nodes {
		err := executeNode(node, ctx, builder, forLayers)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// executeNode executes a single node, handling text and variable nodes differently.
func executeNode(node *Node, ctx Context, builder *strings.Builder, forLayers int) error {
	switch node.Type {
	case "text":
		builder.WriteString(node.Text)
	case "variable":
		value, err := executeVariableNode(node, ctx)
		builder.WriteString(value)
		if err != nil {
			return err
		}
	case "if":
		if err := executeIfNode(node, ctx, builder, forLayers); err != nil {
			return err
		}
	case "for":
		forLayers++
		if err := executeForNode(node, ctx, builder, forLayers); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnknownNodeType, node.Type)
	}
	return nil
}

// executeVariableNode resolves and processes a variable node, applying any filters.
func executeVariableNode(node *Node, ctx Context) (string, error) {
	value, err := resolveVariable(node.Variable, ctx)
	if err != nil {
		// Instead of returning an error, return the original variable placeholder.
		return node.Text, err
	}

	// Apply filters to the resolved value.
	if len(node.Filters) > 0 {
		value, err = ApplyFilters(value, node.Filters, ctx)
		if err != nil {
			return node.Text, err
		}
	}

	result, err := convertToString(value)
	if err != nil {
		return node.Text, nil //nolint: nilerr // Return the original variable placeholder.
	}

	return result, nil
}

// resolveVariable retrieves and formats a variable's value from the context, supporting nested keys.
func resolveVariable(variable string, ctx Context) (interface{}, error) {
	// Directly return string literals.
	if strings.HasPrefix(variable, "'") && strings.HasSuffix(variable, "'") {
		return strings.Trim(variable, "'"), nil
	}

	value, err := ctx.Get(variable)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// convertToString attempts to convert various types to a string, handling common and complex types distinctly.
func convertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []string:
		return fmt.Sprintf("[%s]", strings.Join(v, ", ")), nil
	case []int, []int64, []float64, []bool:
		formatted := fmt.Sprint(v)                           // Convert slice to string
		formatted = strings.Trim(formatted, "[]")            // Remove square brackets
		formatted = strings.ReplaceAll(formatted, " ", ", ") // Replace spaces with commas
		return fmt.Sprintf("[%s]", formatted), nil
	case time.Time:
		// Customize the time format as needed
		return v.Format("2006-01-02 15:04:05"), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		// Fallback for more complex or unknown types: use JSON serialization
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("could not convert value to string: %w", err)
		}
		return string(jsonBytes), nil
	}
}

// executeIfNode handles conditional rendering
func executeIfNode(node *Node, ctx Context, builder *strings.Builder, forLayers int) error {
	// Parse condition expression
	expression := disassembleExpression(node.Text)
	lexer := &Lexer{input: expression}
	tokens, err := lexer.Lex()
	if err != nil {
		return err
	}

	grammar := NewGrammar(tokens)
	ast, err := grammar.Parse()
	if err != nil {
		return err
	}
	condition, err := ast.Evaluate(ctx)
	if err != nil {
		return err
	}

	// Convert condition to boolean value
	conditionMet, err := condition.toBool()
	if err != nil {
		return err
	}

	next := 0
	ElseExists := false
	for next < len(node.Children) {
		if node.Children[next].Text == "{% else %}" {
			ElseExists = true
			break
		}
		next++
	}

	// Execute corresponding child nodes based on condition
	switch {
	case conditionMet:
		return executeNodes(node.Children[:next], ctx, builder, forLayers)
	case ElseExists:
		return executeNodes(node.Children[next].Children, ctx, builder, forLayers)
	default:
		return nil
	}
}

// executeForNode handles loop rendering
func executeForNode(node *Node, ctx Context, builder *strings.Builder, forLayers int) error {
	// Get collection data
	collection, err := resolveVariable(node.Collection, ctx)
	if err != nil {
		return err
	}
	// Create new context only for the outer loop
	loopCtx := Context{}
	if forLayers < 2 {
		loopCtx = deepCopy(ctx)
	}
	// Handle different types of collections
	val := reflect.ValueOf(collection)
	kind := val.Kind()

	if kind == reflect.String {
		str := val.String()
		for i, ch := range str {
			updateLoopContext(loopCtx, node.Variable, "", string(ch), i)
			err := executeNodes(node.Children, loopCtx, builder, forLayers)
			forLayers--
			if err != nil {
				return err
			}
		}
		return nil
	}

	if kind == reflect.Slice || kind == reflect.Array {
		length := val.Len()
		for i := 0; i < length; i++ {
			item := val.Index(i).Interface()
			updateLoopContext(loopCtx, node.Variable, "", item, i)
			err := executeNodes(node.Children, loopCtx, builder, forLayers)
			forLayers--
			if err != nil {
				return err
			}
		}
		return nil
	}

	if kind == reflect.Map {
		keys := val.MapKeys()

		for i, key := range keys {
			keyStr := fmt.Sprint(key.Interface())
			valueValue := val.MapIndex(key).Interface()

			// Check if we have a single variable name or multiple variable names
			varNames := strings.Split(node.Variable, ",")
			for j := range varNames {
				varNames[j] = strings.TrimSpace(varNames[j])
			}

			var item interface{}
			if len(varNames) == 1 {
				// For single variable, create an object with key and value
				item = map[string]interface{}{
					"key":   keyStr,
					"value": valueValue,
				}
				updateLoopContext(loopCtx, node.Variable, "", item, i)
			} else {
				// For two variables, use the original behavior
				updateLoopContext(loopCtx, node.Variable, keyStr, valueValue, i)
			}

			err := executeNodes(node.Children, loopCtx, builder, forLayers)
			forLayers--
			if err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("%w: %T", ErrUnsupportedCollectionType, collection)
}

// deepCopy deep copy context data
func deepCopy(ctx Context) Context {
	newCtx := make(Context)
	for k, v := range ctx {
		newCtx[k] = deepCopyValue(v)
	}
	return newCtx
}

// deepCopyValue deep copy any value
func deepCopyValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for k, v := range val {
			newMap[k] = deepCopyValue(v)
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(val))
		for i, v := range val {
			newSlice[i] = deepCopyValue(v)
		}
		return newSlice
	case Context:
		return deepCopy(val)
	case map[interface{}]interface{}:
		newMap := make(map[interface{}]interface{})
		for k, v := range val {
			newMap[k] = deepCopyValue(v)
		}
		return newMap
	case []string:
		newSlice := make([]string, len(val))
		copy(newSlice, val)
		return newSlice
	case []int:
		newSlice := make([]int, len(val))
		copy(newSlice, val)
		return newSlice
	case []float64:
		newSlice := make([]float64, len(val))
		copy(newSlice, val)
		return newSlice
	case []bool:
		newSlice := make([]bool, len(val))
		copy(newSlice, val)
		return newSlice
	default:
		// For basic types (string, int, float64, bool, etc.), return directly
		return val
	}
}

// updateLoopContext updates loop context information
func updateLoopContext(ctx Context, varName string, keyStr string, item interface{}, index int) {
	// Split variable name into multiple variables if needed
	varNames := strings.Split(varName, ",")
	// Trim spaces from each variable name
	for i := range varNames {
		varNames[i] = strings.TrimSpace(varNames[i])
	}

	if len(varNames) == 1 {
		ctx.Set(varNames[0], item)
	} else if len(varNames) == 2 {
		if keyStr != "" {
			ctx.Set(varNames[0], keyStr)
			ctx.Set(varNames[1], item)
		} else {
			ctx.Set(varNames[0], index)
			ctx.Set(varNames[1], item)
		}
	}
}

// Disassemble expression
func disassembleExpression(expression string) string {
	n := len(expression)
	prev := 0
	next := 0
	for next < n {
		tokenPrev := expression[0:prev]
		tokenNext := expression[next:n]

		if tokenPrev != "{% if " && tokenPrev != "{%if " {
			prev++
		}
		if tokenNext != " %}" {
			next++
		}
		if (tokenPrev == "{% if " || tokenPrev == "{%if ") && tokenNext == " %}" {
			break
		}
	}
	return expression[prev:next]
}
