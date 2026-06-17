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
	root []statement

	// name is the loader-resolved name (including any loader prefix);
	// "" for templates parsed from an unnamed source string.
	name string

	// engine is the owning Engine; nil for templates parsed without an engine.
	engine *Engine

	// parent is the template this one extends, nil otherwise.
	parent *Template

	// blocks is this template's own block definitions, keyed by block name.
	// Populated only when the template contains {% block %} tags.
	blocks map[string]*blockNode
}

// blockNode is a forward declaration stub filled in when {% block %} is
// implemented. Defined here so Template.blocks can reference it without
// import cycles.
type blockNode struct {
	Name string
	Body []node
	Line int
	Col  int
}

func newTemplate(root []statement) *Template {
	return &Template{root: root}
}

// execute writes the template output to w using the given render context.
//
// When the template extends a parent (via {% extends %}), execute walks up
// to the root of the extends chain and runs that template's body. The
// current template is recorded as ctx.currentLeaf so blockNode.Execute
// can resolve overrides across the chain. For templates without a parent
// the loop is a no-op and the template runs its own body.
func (t *Template) execute(ctx *renderContext, w io.Writer) error {
	root := t
	for root.parent != nil {
		root = root.parent
	}
	prevEngine := ctx.engine
	prevAutoescape := ctx.autoescape
	if ctx.engine == nil && t.engine != nil {
		ctx.engine = t.engine
		ctx.autoescape = t.engine.format == FormatHTML
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
	return attachTemplate(t.name, root.executeRoot(ctx, w))
}

// executeRoot runs this template's own top-level statements without
// walking the extends chain.
func (t *Template) executeRoot(ctx *renderContext, w io.Writer) error {
	for _, stmt := range t.root {
		if err := stmt.Execute(ctx, w); err != nil {
			return wrapRender(stmt, err)
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
func (t *Template) RenderTo(w io.Writer, data Data) error {
	ctx := newRenderContext(data)
	return t.execute(ctx, w)
}
