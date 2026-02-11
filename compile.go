// Package template provides a lightweight template engine with Django/Jinja-style syntax.
package template

import "fmt"

// Compile compiles a template source string and returns an executable [Template].
//
// Compile performs lexical analysis, parsing, and template creation in sequence.
// On failure, the returned error wraps the underlying lexer or parser error.
//
//	tmpl, err := template.Compile("Hello {{ name }}!")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	output, _ := tmpl.Render(map[string]any{"name": "World"})
func Compile(source string) (*Template, error) {
	l := NewLexer(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("tokenizing template: %w", err)
	}

	p := NewParser(tokens)
	ast, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return NewTemplate(ast), nil
}
