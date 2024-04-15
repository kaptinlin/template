package template

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Node defines a single element within a template, such as text, variable, or control structure.
type Node struct {
	Type     string
	Text     string
	Variable string
	Filters  []Filter
	Children []*Node
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
	if err := executeNodes(t.Nodes, ctx, &builder); err != nil {
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
func executeNodes(nodes []*Node, ctx Context, builder *strings.Builder) error {
	var firstErr error
	for _, node := range nodes {
		err := executeNode(node, ctx, builder)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// executeNode executes a single node, handling text and variable nodes differently.
func executeNode(node *Node, ctx Context, builder *strings.Builder) error {
	switch node.Type {
	case "text":
		builder.WriteString(node.Text)
	case "variable":
		value, err := executeVariableNode(node, ctx)
		builder.WriteString(value)
		if err != nil {
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
		jsonBytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("could not convert value to string: %w", err)
		}
		return string(jsonBytes), nil
	}
}
