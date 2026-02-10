package template

import (
	"bytes"
	"io"
)

// Template represents a compiled template ready for execution.
// It contains the parsed AST and provides methods for rendering.
//
// Template instances are safe for concurrent use by multiple goroutines
// once created, as they are immutable after compilation.
type Template struct {
	root []Statement // Root AST nodes (the compiled template)
}

// NewTemplate creates a new Template from parsed AST nodes.
// The AST nodes should be obtained from the parse package.
//
// Example:
//
//	lexer := template.NewLexer(content)
//	tokens, _ := lexer.Tokenize()
//	parser := template.NewParser(tokens)
//	ast, _ := parser.Parse()
//	tmpl := template.NewTemplate(ast)
func NewTemplate(root []Statement) *Template {
	return &Template{
		root: root,
	}
}

// Execute executes the template with the given context and writes output to the writer.
//
// This is the low-level execution method that gives full control over the execution context
// and output destination. It executes each statement in the AST sequentially and writes
// the output to the provided writer.
//
// Parameters:
//   - ctx: The execution context containing variables and internal state
//   - writer: The output writer (can be any io.Writer)
//
// Returns:
//   - error: Any error that occurred during execution
//
// Example:
//
//	var buf bytes.Buffer
//	ctx := exec.NewContext(map[string]interface{}{"name": "Alice"})
//	err := tmpl.Execute(ctx, &buf)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(buf.String())
func (t *Template) Execute(ctx *ExecutionContext, writer io.Writer) error {
	// Execute each statement in the root AST
	for _, stmt := range t.root {
		if err := stmt.Execute(ctx, writer); err != nil {
			return err
		}
	}
	return nil
}

// Render executes the template with the given data and returns the output as a string.
//
// This is a convenience method that wraps Execute for the common case where you want
// a string result. It creates a new execution context from the provided data,
// executes the template into a buffer, and returns the result as a string.
//
// Parameters:
//   - data: A map of variable names to values that will be available in the template
//
// Returns:
//   - string: The rendered template output
//   - error: Any error that occurred during execution
//
// Example:
//
//	output, err := tmpl.Render(map[string]interface{}{
//	    "name": "Alice",
//	    "age": 30,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(output) // Output: rendered template
func (t *Template) Render(data map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	ctx := NewExecutionContext(data)

	if err := t.Execute(ctx, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
