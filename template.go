package template

import (
	"bytes"
	"io"
)

// Template represents a compiled template ready for execution.
// A Template is immutable after compilation.
//
// Templates parsed without a loader-backed engine have no resolved name or
// engine reference. Templates loaded via Engine.Load carry an engine
// reference and a resolved name for caching, dependency tracking, and
// multi-file rendering.
type Template struct {
	root []Statement

	// name is the loader-resolved name (including any loader prefix);
	// "" for templates parsed from an unnamed source string.
	name string

	// engine is the owning Engine; nil for templates parsed without an engine.
	engine *Engine

	// parent is the template this one extends, nil otherwise.
	parent *Template

	// blocks is this template's own block definitions, keyed by block name.
	// Populated only when the template contains {% block %} tags.
	blocks map[string]*BlockNode
}

// BlockNode is a forward declaration stub filled in when {% block %} is
// implemented. Defined here so Template.blocks can reference it without
// import cycles.
type BlockNode struct {
	Name string
	Body []Node
	Line int
	Col  int
}

// NewTemplate creates a new Template from parsed AST nodes.
//
// Most callers should use [Engine.ParseString] or [Engine.Load], which handle
// lexing and parsing automatically.
func NewTemplate(root []Statement) *Template {
	return &Template{root: root}
}

// Execute writes the template output to w using the given render context.
//
// For most use cases, [Template.Render] is simpler. Use Execute when you need
// control over the output destination or render context.
//
// When the template extends a parent (via {% extends %}), Execute walks up
// to the root of the extends chain and runs that template's body. The
// current template is recorded as ctx.currentLeaf so BlockNode.Execute
// can resolve overrides across the chain. For templates without a parent
// the loop is a no-op and the template runs its own body.
func (t *Template) Execute(ctx *RenderContext, w io.Writer) error {
	root := t
	for root.parent != nil {
		root = root.parent
	}
	prevEngine := ctx.engine
	prevAutoescape := ctx.autoescape
	if ctx.engine == nil && t.engine != nil {
		ctx.engine = t.engine
		ctx.autoescape = t.engine.autoescape()
		defer func() {
			ctx.engine = prevEngine
			ctx.autoescape = prevAutoescape
		}()
	}
	// Preserve any outer currentLeaf (for nested include+extends
	// scenarios) and restore on return.
	prevLeaf := ctx.currentLeaf
	ctx.currentLeaf = t
	defer func() { ctx.currentLeaf = prevLeaf }()
	return root.executeRoot(ctx, w)
}

// executeRoot runs this template's own top-level statements without
// walking the extends chain.
func (t *Template) executeRoot(ctx *RenderContext, w io.Writer) error {
	for _, stmt := range t.root {
		if err := stmt.Execute(ctx, w); err != nil {
			return err
		}
	}
	return nil
}

// Render executes the template with data and returns the output as a string.
//
// Render is a convenience wrapper around [Template.RenderTo] for the common
// case where a string result is needed.
func (t *Template) Render(data Data) (string, error) {
	var buf bytes.Buffer
	if err := t.RenderTo(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderTo writes the template output to w using plain render data.
//
// Use RenderTo for the common writer-based path. Reach for [Template.Execute]
// only when you need direct control over [RenderContext].
func (t *Template) RenderTo(w io.Writer, data Data) error {
	ctx := NewRenderContext(data)
	return t.Execute(ctx, w)
}
