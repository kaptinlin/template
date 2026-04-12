// Package template provides a lightweight template engine with Django/Jinja-style syntax.
package template

import (
	"errors"
	"fmt"
)

type templateSourceError struct {
	name string
	err  error
}

func (e *templateSourceError) Error() string {
	var parseErr *ParseError
	if errors.As(e.err, &parseErr) {
		return fmt.Sprintf("%s:%d:%d: %s", e.name, parseErr.Line, parseErr.Col, parseErr.Message)
	}
	return fmt.Sprintf("%s: %v", e.name, e.err)
}

func (e *templateSourceError) Unwrap() error {
	return e.err
}

func wrapTemplateSourceError(name string, err error) error {
	if err == nil || name == "" {
		return err
	}
	var sourceErr *templateSourceError
	if errors.As(err, &sourceErr) {
		return err
	}
	return &templateSourceError{name: name, err: err}
}

// compileForEngine compiles a template source, wiring the parser to the given
// Engine so tag parsers can resolve referenced templates at parse time.
//
// When FeatureLayout is enabled, the lexer also accepts {% raw %}...{% endraw %}
// blocks and the parser can resolve layout tags against the engine loader.
func compileForEngine(source string, engine *Engine) (*Template, error) {
	return compileNamedForEngine("", source, engine)
}

func compileNamedForEngine(name, source string, engine *Engine) (*Template, error) {
	l := NewLexer(source)
	// Only enable raw-block lexer mode when layout is enabled.
	if engine != nil && engine.HasFeature(FeatureLayout) {
		l.allowRaw = true
	}
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, wrapTemplateSourceError(name, fmt.Errorf("tokenizing template: %w", err))
	}

	p := NewParser(tokens)
	p.engine = engine
	ast, err := p.Parse()
	if err != nil {
		return nil, wrapTemplateSourceError(name, fmt.Errorf("parsing template: %w", err))
	}

	tpl := NewTemplate(ast)
	// Transfer parse-time state onto the Template.
	tpl.parent = p.parent
	tpl.blocks = p.blocks
	return tpl, nil
}
