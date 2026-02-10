// Package template provides a lightweight template engine with Django/Jinja-style syntax.
package template

import (
	"fmt"
)

// Compile compiles a template source string and returns an executable [Template].
//
// It performs lexical analysis, parsing, and template creation in sequence.
// On failure, the returned error wraps the underlying lexer or parser error.
//
//	tmpl, err := template.Compile("Hello {{ name }}!")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	output, _ := tmpl.Render(map[string]interface{}{"name": "World"})
func Compile(source string) (*Template, error) {
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("tokenize: %w", err)
	}

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return NewTemplate(ast), nil
}
