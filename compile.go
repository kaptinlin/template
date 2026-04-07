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
	return compileForSet(source, nil)
}

// compileForSet compiles a template source, wiring the parser to the given
// Set so tag parsers (include, extends) can resolve referenced templates
// at parse time. Pass set = nil to compile a standalone template.
//
// Templates compiled via a Set have {% raw %}...{% endraw %} lexer
// support enabled, whereas standalone Compile(src) does not — this
// keeps the Compile(src) path byte-identical to its pre-layout
// behavior.
func compileForSet(source string, set *Set) (*Template, error) {
	l := NewLexer(source)
	// Only enable raw-block lexer mode when loading via a Set.
	// Compile(src) passes set=nil and keeps its original behavior.
	if set != nil {
		l.allowRaw = true
	}
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("tokenizing template: %w", err)
	}

	p := NewParser(tokens)
	p.set = set
	ast, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	tpl := NewTemplate(ast)
	// Transfer parse-time state onto the Template.
	tpl.parent = p.parent
	tpl.blocks = p.blocks
	return tpl, nil
}
