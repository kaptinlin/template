package template

import (
	"bytes"
	"io"
)

// Template represents a compiled template ready for execution.
// It is immutable after compilation.
type Template struct {
	root []Statement
}

// NewTemplate creates a new Template from parsed AST nodes.
//
// Most callers should use [Compile] instead, which handles
// lexing and parsing automatically.
func NewTemplate(root []Statement) *Template {
	return &Template{
		root: root,
	}
}

// Execute writes the template output to writer using the given execution context.
//
// For most use cases, [Template.Render] is simpler. Use Execute when you need
// control over the output destination or execution context.
func (t *Template) Execute(ctx *ExecutionContext, writer io.Writer) error {
	for _, stmt := range t.root {
		if err := stmt.Execute(ctx, writer); err != nil {
			return err
		}
	}
	return nil
}

// Render executes the template with data and returns the output as a string.
//
// This is a convenience wrapper around [Template.Execute] for the common case
// where a string result is needed.
func (t *Template) Render(data map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	ctx := NewExecutionContext(data)

	if err := t.Execute(ctx, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
