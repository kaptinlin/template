package template

import (
	"bytes"
	"io"
)

// Template represents a compiled template ready for execution.
// A Template is immutable after compilation.
//
// Templates compiled via Compile(src) have nil set/name and cannot use
// include/extends. Templates loaded via Set.Get carry a set reference
// and a resolved name for caching, dependency tracking, and multi-file
// rendering.
type Template struct {
	root []Statement

	// name is the loader-resolved name (including any loader prefix);
	// "" for templates created via Compile(src).
	name string

	// set is the owning Set; nil for templates created via Compile(src).
	set *Set

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
// Most callers should use [Compile] instead, which handles
// lexing and parsing automatically.
func NewTemplate(root []Statement) *Template {
	return &Template{root: root}
}

// Execute writes the template output to w using the given execution context.
//
// For most use cases, [Template.Render] is simpler. Use Execute when you need
// control over the output destination or execution context.
//
// When the template extends a parent (via {% extends %}), Execute walks up
// to the root of the extends chain and runs that template's body. The
// current template is recorded as ctx.currentLeaf so BlockNode.Execute
// can resolve overrides across the chain. For templates without a parent
// the loop is a no-op and the template runs its own body.
func (t *Template) Execute(ctx *ExecutionContext, w io.Writer) error {
	root := t
	for root.parent != nil {
		root = root.parent
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
func (t *Template) executeRoot(ctx *ExecutionContext, w io.Writer) error {
	for _, stmt := range t.root {
		if err := stmt.Execute(ctx, w); err != nil {
			return err
		}
	}
	return nil
}

// Render executes the template with data and returns the output as a string.
//
// Render is a convenience wrapper around [Template.Execute] for the common case
// where a string result is needed.
func (t *Template) Render(data Context) (string, error) {
	var buf bytes.Buffer
	ctx := NewExecutionContext(data)
	if err := t.Execute(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
