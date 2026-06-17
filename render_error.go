package template

import (
	"errors"
	"fmt"
)

// RenderError carries machine-readable context for a render-time failure.
//
// It always wraps an underlying sentinel from errors.go (directly or
// through fmt.Errorf wrapping). Callers identify the failure mode with
// errors.Is and recover position information with errors.As:
//
//	var re *template.RenderError
//	if errors.As(err, &re) {
//	    log.Printf("%s:%d:%d: %v", re.Template, re.Line, re.Col, re.Cause)
//	}
//	if errors.Is(err, template.ErrFilterNotFound) {
//	    // ...
//	}
//
// The Error() string is human-readable and may evolve; the Template, Line,
// Col, and Cause fields are part of the public contract and will not change
// shape across minor versions.
type RenderError struct {
	// Template is the loader-resolved template name where the failure
	// originated; empty for templates parsed via Engine.ParseString.
	Template string

	// Line is the 1-based line where the failing node starts; 0 if unknown.
	Line int

	// Col is the 1-based column where the failing node starts; 0 if unknown.
	Col int

	// Cause is the underlying error. It is always non-nil for a
	// well-formed RenderError and is suitable for errors.Is / errors.As
	// against the sentinels in errors.go.
	Cause error
}

// Error returns a human-readable representation. The format is not part
// of the public contract; consume Template, Line, Col, and Cause directly
// for stable output.
func (e *RenderError) Error() string {
	switch {
	case e.Template != "" && e.Line > 0:
		return fmt.Sprintf("%s:%d:%d: %v", e.Template, e.Line, e.Col, e.Cause)
	case e.Template != "":
		return fmt.Sprintf("%s: %v", e.Template, e.Cause)
	case e.Line > 0:
		return fmt.Sprintf("%d:%d: %v", e.Line, e.Col, e.Cause)
	default:
		return e.Cause.Error()
	}
}

// Unwrap exposes the underlying cause for errors.Is / errors.As.
func (e *RenderError) Unwrap() error { return e.Cause }

// wrapRender attaches node position context to a render-time error.
// It returns nil unchanged, leaves an existing *RenderError alone (so the
// deepest position wins), and never wraps loop-control sentinels.
func wrapRender(n node, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[*RenderError](err); ok {
		return err
	}
	if _, ok := errors.AsType[*breakError](err); ok {
		return err
	}
	if _, ok := errors.AsType[*continueError](err); ok {
		return err
	}
	line, col := n.Position()
	return &RenderError{Line: line, Col: col, Cause: err}
}

// attachTemplate fills in the Template field on a *RenderError if it is
// not already set. Other error types pass through untouched. The original
// pointer is preserved when no change is needed; otherwise a copy is
// returned to keep RenderError values free of unexpected mutation.
func attachTemplate(name string, err error) error {
	if err == nil || name == "" {
		return err
	}
	re, ok := errors.AsType[*RenderError](err)
	if !ok || re.Template != "" {
		return err
	}
	return &RenderError{
		Template: name,
		Line:     re.Line,
		Col:      re.Col,
		Cause:    re.Cause,
	}
}
