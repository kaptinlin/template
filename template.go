package template

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-json-experiment/json"
)

// String conversion constants
const (
	estimatedCharsPerArrayItem = 20 // Estimate ~20 chars per item when converting arrays to strings
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

// ControlFlow represents the type of control flow operation
type ControlFlow int

const (
	// ControlFlowNone indicates no control flow change
	ControlFlowNone ControlFlow = iota
	// ControlFlowBreak indicates a break control flow
	ControlFlowBreak
	// ControlFlowContinue indicates a continue control flow
	ControlFlowContinue
)

// Node type constants
const (
	NodeTypeBreak    = "break"
	NodeTypeContinue = "continue"
)

// LoopContext represents the loop information available in templates
type LoopContext struct {
	Index    int  // Current index (starting from 0)
	Revindex int  // Reverse index (length-1 to 0)
	First    bool // Whether this is the first iteration
	Last     bool // Whether this is the last iteration
	Length   int  // Total length of the collection
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
	_, err := executeNodes(t.Nodes, ctx, &builder, forLayers)
	if err != nil {
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
func executeNodes(nodes []*Node, ctx Context, builder *strings.Builder, forLayers int) (ControlFlow, error) {
	var errs []error
	for _, node := range nodes {
		controlFlow, err := executeNode(node, ctx, builder, forLayers)
		if err != nil {
			errs = append(errs, err)
		}
		// If a control flow signal is received, return it immediately with aggregated errors
		if controlFlow != ControlFlowNone {
			return controlFlow, errors.Join(errs...)
		}
	}
	return ControlFlowNone, errors.Join(errs...)
}

// executeNode executes a single node, handling text and variable nodes differently.
func executeNode(node *Node, ctx Context, builder *strings.Builder, forLayers int) (ControlFlow, error) {
	switch node.Type {
	case "text":
		builder.WriteString(node.Text)
	case "variable":
		value, err := executeVariableNode(node, ctx)
		builder.WriteString(value)
		if err != nil {
			return ControlFlowNone, err
		}
	case "if":
		controlFlow, err := executeIfNode(node, ctx, builder, forLayers)
		if err != nil {
			return ControlFlowNone, err
		}
		if controlFlow != ControlFlowNone {
			return controlFlow, nil
		}
	case "for":
		forLayers++
		controlFlow, err := executeForNode(node, ctx, builder, forLayers)
		if err != nil {
			return ControlFlowNone, err
		}
		if controlFlow != ControlFlowNone {
			return controlFlow, nil
		}
	case NodeTypeBreak:
		// Check if we're in a loop
		if forLayers <= 0 {
			return ControlFlowNone, ErrBreakOutsideLoop
		}
		return ControlFlowBreak, nil
	case NodeTypeContinue:
		// Check if we're in a loop
		if forLayers <= 0 {
			return ControlFlowNone, ErrContinueOutsideLoop
		}
		return ControlFlowContinue, nil
	case "elif", "else":
		// For elif/else nodes, execute their children but skip the tag itself
		return executeNodes(node.Children, ctx, builder, forLayers)
	default:
		return ControlFlowNone, fmt.Errorf("%w: %s", ErrUnknownNodeType, node.Type)
	}
	return ControlFlowNone, nil
}

// executeVariableNode resolves and processes a variable node, applying any filters.
// If an error occurs, it returns the original variable placeholder along with the error
// for debugging purposes, but allows template execution to continue.
func executeVariableNode(node *Node, ctx Context) (string, error) {
	value, err := resolveVariable(node.Variable, ctx)
	if err != nil {
		// Return fallback value with error for debugging
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
		// Return fallback value with error for debugging
		return node.Text, err
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

// convertToString attempts to convert various types to a string, handling all types including aliases uniformly.
func convertToString(value interface{}) (string, error) {
	// Handle special interfaces first
	if t, ok := value.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05"), nil
	}
	if s, ok := value.(fmt.Stringer); ok {
		return s.String(), nil
	}

	// Use reflect for uniform handling of all types including aliases
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return "", nil
	}

	//nolint:exhaustive // Only handle types supported by the template engine
	switch rv.Kind() {
	case reflect.String:
		return rv.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", rv.Float()), nil
	case reflect.Bool:
		return fmt.Sprintf("%t", rv.Bool()), nil
	case reflect.Slice, reflect.Array:
		length := rv.Len()
		if length == 0 {
			return "[]", nil
		}

		var result strings.Builder
		result.Grow(length * estimatedCharsPerArrayItem)
		result.WriteByte('[')

		for i := 0; i < length; i++ {
			if i > 0 {
				result.WriteByte(',')
			}
			item := rv.Index(i).Interface()
			str, err := convertToString(item)
			if err != nil {
				// Fallback to JSON for complex items
				jsonBytes, err := json.Marshal(item, json.Deterministic(true))
				if err != nil {
					return "", fmt.Errorf("could not convert slice item to string: %w", err)
				}
				str = string(jsonBytes)
			}
			result.WriteString(str)
		}
		result.WriteByte(']')
		return result.String(), nil
	default:
		// Fallback for complex types: use JSON serialization
		jsonBytes, err := json.Marshal(value, json.Deterministic(true))
		if err != nil {
			return "", fmt.Errorf("could not convert value to string: %w", err)
		}
		return string(jsonBytes), nil
	}
}

// parseCondition parses a condition expression and returns a boolean value
func parseCondition(text string, ctx Context) (bool, error) {
	expression := disassembleExpression(text)
	lexer := &Lexer{input: expression}
	tokens, err := lexer.Lex()
	if err != nil {
		return false, err
	}

	grammar := NewGrammar(tokens)
	ast, err := grammar.Parse()
	if err != nil {
		return false, err
	}

	condition, err := ast.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	return condition.toBool()
}

// executeIfBranch executes if branch content until the first else/elif
func executeIfBranch(node *Node, ctx Context, builder *strings.Builder, forLayers int) (ControlFlow, error) {
	for _, child := range node.Children {
		if child.Type == "else" || child.Type == "elif" {
			break
		}
		controlFlow, err := executeNode(child, ctx, builder, forLayers)
		if err != nil {
			return ControlFlowNone, err
		}
		if controlFlow != ControlFlowNone {
			return controlFlow, nil
		}
	}
	return ControlFlowNone, nil
}

// validateIfElseStructure validates that an if node has at most one else branch
func validateIfElseStructure(node *Node) error {
	elseCount := 0
	for _, child := range node.Children {
		if child.Type == "else" {
			elseCount++
			if elseCount > 1 {
				return ErrMultipleElseStatements
			}
		}
	}
	return nil
}

// executeIfNode handles conditional rendering with support for elif branches
func executeIfNode(node *Node, ctx Context, builder *strings.Builder, forLayers int) (ControlFlow, error) {
	// Validate the if-else structure before execution
	if err := validateIfElseStructure(node); err != nil {
		return ControlFlowNone, err
	}

	// 1. Evaluate main if condition
	conditionMet, err := parseCondition(node.Text, ctx)
	if err != nil {
		return ControlFlowNone, err
	}

	if conditionMet {
		// Execute if branch content (until first else/elif)
		return executeIfBranch(node, ctx, builder, forLayers)
	}

	// 2. Evaluate elif conditions (short-circuit mechanism)
	for _, child := range node.Children {
		if child.Type == "elif" {
			elifCondition, err := parseCondition(child.Text, ctx)
			if err != nil {
				return ControlFlowNone, err
			}
			if elifCondition {
				return executeNodes(child.Children, ctx, builder, forLayers)
			}
		}
	}

	// 3. Check for else branch
	for _, child := range node.Children {
		if child.Type == "else" {
			return executeNodes(child.Children, ctx, builder, forLayers)
		}
	}

	return ControlFlowNone, nil
}

// executeForNode handles loop rendering
func executeForNode(node *Node, ctx Context, builder *strings.Builder, forLayers int) (ControlFlow, error) {
	// 1. Backup existing loop object (if exists)
	var backupLoop interface{}
	var hasBackup bool
	if existingLoop, exists := ctx["loop"]; exists {
		backupLoop = existingLoop
		hasBackup = true
	}

	// 2. Get collection data and calculate length
	collection, err := resolveVariable(node.Collection, ctx)
	if err != nil {
		return ControlFlowNone, err
	}
	length := calculateCollectionLength(collection)

	// 3. Create new context only for the outer loop (preserve original logic)
	var loopCtx Context
	if forLayers < 2 {
		loopCtx = deepCopy(ctx)
		// Ensure backup state is correctly set in new context
		if hasBackup {
			loopCtx["loop"] = backupLoop
		}
	} else {
		loopCtx = ctx
	}

	// 4. Handle different types of collections
	val := reflect.ValueOf(collection)
	kind := val.Kind()

	if kind == reflect.String {
		str := val.String()
		for i, ch := range str {
			// Create current iteration's LoopContext
			currentLoop := &LoopContext{
				Index:    i,
				Revindex: length - 1 - i,
				First:    i == 0,
				Last:     i == length-1,
				Length:   length,
			}

			// Set loop object to context
			loopCtx.Set("loop", currentLoop)

			// Update loop variables (preserve original logic)
			updateLoopContext(loopCtx, node.Variable, "", string(ch), i)

			// Execute loop body
			controlFlow, err := executeNodes(node.Children, loopCtx, builder, forLayers)
			if err != nil {
				// Restore backup on error
				restoreLoopBackup(loopCtx, backupLoop, hasBackup)
				return ControlFlowNone, err
			}

			switch controlFlow {
			case ControlFlowBreak:
				restoreLoopBackup(loopCtx, backupLoop, hasBackup)
				return ControlFlowNone, nil
			case ControlFlowContinue:
				continue
			case ControlFlowNone:
				// Normal flow, continue iteration
			}
		}

		// String iteration ended, restore backup
		restoreLoopBackup(loopCtx, backupLoop, hasBackup)
		return ControlFlowNone, nil
	}

	if kind == reflect.Slice || kind == reflect.Array {
		arrayLength := val.Len()
		for i := 0; i < arrayLength; i++ {
			// Create current iteration's LoopContext
			currentLoop := &LoopContext{
				Index:    i,
				Revindex: arrayLength - 1 - i,
				First:    i == 0,
				Last:     i == arrayLength-1,
				Length:   arrayLength,
			}

			// Set loop object to context
			loopCtx.Set("loop", currentLoop)

			item := val.Index(i).Interface()
			updateLoopContext(loopCtx, node.Variable, "", item, i)

			controlFlow, err := executeNodes(node.Children, loopCtx, builder, forLayers)
			if err != nil {
				restoreLoopBackup(loopCtx, backupLoop, hasBackup)
				return ControlFlowNone, err
			}
			switch controlFlow {
			case ControlFlowBreak:
				restoreLoopBackup(loopCtx, backupLoop, hasBackup)
				return ControlFlowNone, nil
			case ControlFlowContinue:
				continue
			case ControlFlowNone:
				// Normal flow, continue iteration
			}
		}
		restoreLoopBackup(loopCtx, backupLoop, hasBackup)
		return ControlFlowNone, nil
	}

	if kind == reflect.Map {
		keys := val.MapKeys()

		// Sort keys by their string representation for stable iteration order
		slices.SortFunc(keys, func(a, b reflect.Value) int {
			return cmp.Compare(
				fmt.Sprint(a.Interface()),
				fmt.Sprint(b.Interface()),
			)
		})

		for i, key := range keys {
			// Create current iteration's LoopContext
			currentLoop := &LoopContext{
				Index:    i,
				Revindex: length - 1 - i,
				First:    i == 0,
				Last:     i == length-1,
				Length:   length,
			}

			// Set loop object to context
			loopCtx.Set("loop", currentLoop)

			keyStr := fmt.Sprint(key.Interface())
			valueValue := val.MapIndex(key).Interface()

			// Check if we have a single variable name or multiple variable names
			varNames := strings.Split(node.Variable, ",")
			for j := range varNames {
				varNames[j] = strings.TrimSpace(varNames[j])
			}

			if len(varNames) == 1 {
				// For single variable, create an object with key and value
				item := map[string]interface{}{
					"key":   keyStr,
					"value": valueValue,
				}
				updateLoopContext(loopCtx, node.Variable, "", item, i)
			} else {
				// For two variables, use the original behavior
				updateLoopContext(loopCtx, node.Variable, keyStr, valueValue, i)
			}

			controlFlow, err := executeNodes(node.Children, loopCtx, builder, forLayers)
			if err != nil {
				restoreLoopBackup(loopCtx, backupLoop, hasBackup)
				return ControlFlowNone, err
			}
			switch controlFlow {
			case ControlFlowBreak:
				restoreLoopBackup(loopCtx, backupLoop, hasBackup)
				return ControlFlowNone, nil
			case ControlFlowContinue:
				continue
			case ControlFlowNone:
				// Normal flow, continue iteration
			}
		}
		restoreLoopBackup(loopCtx, backupLoop, hasBackup)
		return ControlFlowNone, nil
	}

	restoreLoopBackup(loopCtx, backupLoop, hasBackup)
	return ControlFlowNone, fmt.Errorf("%w: %T", ErrUnsupportedCollectionType, collection)
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
		return slices.Clone(val)
	case []int:
		return slices.Clone(val)
	case []float64:
		return slices.Clone(val)
	case []bool:
		return slices.Clone(val)
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

// Disassemble expression from if or elif statements
func disassembleExpression(expression string) string {
	n := len(expression)
	prev := 0
	next := 0
	for next < n {
		tokenPrev := expression[0:prev]
		tokenNext := expression[next:n]

		// Check for both if and elif patterns
		if tokenPrev != "{% if " && tokenPrev != "{%if " && tokenPrev != "{% elif " && tokenPrev != "{%elif " {
			prev++
		}
		if tokenNext != " %}" {
			next++
		}
		if (tokenPrev == "{% if " || tokenPrev == "{%if " || tokenPrev == "{% elif " || tokenPrev == "{%elif ") && tokenNext == " %}" {
			break
		}
	}
	return expression[prev:next]
}

// calculateCollectionLength calculates the length of a collection
func calculateCollectionLength(collection interface{}) int {
	val := reflect.ValueOf(collection)
	kind := val.Kind()

	//nolint:exhaustive
	switch kind {
	case reflect.String:
		return utf8.RuneCountInString(val.String())
	case reflect.Slice, reflect.Array, reflect.Map:
		return val.Len()
	default:
		return 0
	}
}

// restoreLoopBackup restores the backed up loop object
func restoreLoopBackup(ctx Context, backup interface{}, hasBackup bool) {
	if hasBackup {
		ctx.Set("loop", backup)
	} else {
		delete(ctx, "loop")
	}
}
