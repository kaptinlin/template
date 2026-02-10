// Package template provides a lightweight template engine with Django/Jinja-style syntax.
package template

import (
	"fmt"
)

// Compile compiles a template from source code and returns an executable Template.
//
// This is the main entry point for template compilation. It performs:
//  1. Lexical analysis (tokenization)
//  2. Parsing (AST generation)
//  3. Template creation
//
// Parameters:
//   - source: Template source code
//
// Returns:
//   - *Template: Compiled template ready for execution
//   - error: Any compilation error (lexer or parser errors)
//
// Example:
//
//	tmpl, err := template.Compile("Hello {{ name }}!")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	output, _ := tmpl.Render(map[string]interface{}{"name": "World"})
//	fmt.Println(output) // Output: Hello World!
func Compile(source string) (*Template, error) {
	// Step 1: Lexical analysis
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("lexer error: %w", err)
	}

	// Step 2: Parsing
	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parser error: %w", err)
	}

	// Step 3: Create template
	return NewTemplate(ast), nil
}
