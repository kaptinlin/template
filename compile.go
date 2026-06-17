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
	if parseErr, ok := errors.AsType[*ParseError](e.err); ok {
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
	if _, ok := errors.AsType[*templateSourceError](err); ok {
		return err
	}
	return &templateSourceError{name: name, err: err}
}

func compileNamedForEngine(name, source string, engine *Engine) (*Template, error) {
	l := newLexer(source)
	// Only enable raw-block lexer mode when layout is enabled.
	if engine != nil && engine.HasFeature(FeatureLayout) {
		l.allowRaw = true
	}
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, wrapTemplateSourceError(name, fmt.Errorf("tokenizing template: %w", err))
	}

	p := newParser(tokens)
	p.engine = engine
	ast, err := p.Parse()
	if err != nil {
		return nil, wrapTemplateSourceError(name, fmt.Errorf("parsing template: %w", err))
	}

	tpl := newTemplate(ast)
	// Transfer parse-time state onto the Template.
	tpl.parent = p.parent
	tpl.blocks = p.blocks
	return tpl, nil
}
