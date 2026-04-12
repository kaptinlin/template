package template

import (
	"errors"
	"fmt"
	"io"
)

// IncludeNode represents an {% include %} statement.
//
// Shapes:
//   - Static path (string literal), resolved at parse time: prepared != nil.
//   - Parse-time circular static path: lazy = true, staticName holds the
//     literal for runtime lookup.
//   - Dynamic path (expression): lazy = true, pathExpr != nil.
//
// Options:
//   - withPairs: {% include "x" with k1=expr1 k2=expr2 %}
//   - only: {% include "x" only %} — fully isolates the child context,
//     excluding parent variables AND defaults.
//   - ifExists: {% include "x" if_exists %} — missing template is a no-op
//     instead of an error.
type IncludeNode struct {
	Line int
	Col  int

	// prepared is the pre-parsed template for static includes. nil when lazy.
	prepared *Template

	// staticName is the literal name for string-literal includes, used for
	// runtime lookup when lazy downgrade is active.
	staticName string

	// pathExpr is set for dynamic includes (non-string-literal path).
	pathExpr Expression

	// lazy indicates runtime template resolution.
	lazy bool

	// withPairs holds "with k=expr" bindings; may be nil.
	withPairs []withPair

	// only indicates the child template must not see the parent's context.
	only bool

	// ifExists indicates a missing template should silently render nothing.
	ifExists bool
}

// withPair is a single "key=expression" binding on an include tag.
type withPair struct {
	name string
	expr Expression
}

// Position returns the source position of the include tag.
func (n *IncludeNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation.
func (n *IncludeNode) String() string {
	if n.prepared != nil {
		return fmt.Sprintf("Include(%q)", n.prepared.name)
	}
	return fmt.Sprintf("Include(%q lazy)", n.staticName)
}

// maxIncludeDepth is the hard cap on {% include %} nesting depth. It
// defends against runaway recursion (self-include, mutual include, data-
// driven deep trees). 32 is well beyond any reasonable real-world need.
const maxIncludeDepth = 32

// Execute renders the included template. The parser only produces
// IncludeNode inside an Engine with layout enabled, so ctx.engine is guaranteed non-nil by the
// time we reach here.
func (n *IncludeNode) Execute(ctx *RenderContext, w io.Writer) error {
	if ctx.includeDepth >= maxIncludeDepth {
		return fmt.Errorf("%w at line %d (max %d)",
			ErrIncludeDepthExceeded, n.Line, maxIncludeDepth)
	}
	child, err := n.resolveChild(ctx)
	if err != nil {
		return err
	}
	if child == nil {
		// if_exists miss: render nothing, return cleanly.
		return nil
	}
	childCtx, err := n.buildChildContext(ctx)
	if err != nil {
		return err
	}
	return child.Execute(childCtx, w)
}

// resolveChild loads the sub-template this include points at. It
// returns (nil, nil) when the target is missing and if_exists is set,
// signalling "silently render nothing". Any other failure is returned
// as an error.
func (n *IncludeNode) resolveChild(ctx *RenderContext) (*Template, error) {
	if n.prepared != nil {
		return n.prepared, nil
	}
	name, err := n.resolveName(ctx)
	if err != nil {
		if n.ifExists && errors.Is(err, ErrTemplateNotFound) {
			return nil, nil
		}
		return nil, err
	}
	tpl, err := ctx.engine.Load(name)
	if err != nil {
		if n.ifExists && errors.Is(err, ErrTemplateNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return tpl, nil
}

// buildChildContext constructs the render context the sub-template
// will run in. It honors the "only" keyword (full isolation) and
// evaluates "with" bindings in the PARENT context.
func (n *IncludeNode) buildChildContext(ctx *RenderContext) (*RenderContext, error) {
	var childCtx *RenderContext
	if n.only {
		// Fully isolated: only the with-pairs are visible.
		childCtx = NewIsolatedChildContext(ctx)
	} else {
		// Inherit parent's render data and runtime state with isolated
		// locals for the child render.
		childCtx = NewChildContext(ctx)
	}
	childCtx.includeDepth = ctx.includeDepth + 1

	// Evaluate "with" bindings in the PARENT context, then insert them
	// into the CHILD context's Locals (which is always a fresh per-
	// child map, so this does not leak into the parent).
	for _, wp := range n.withPairs {
		val, err := wp.expr.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		childCtx.Set(wp.name, val.Interface())
	}
	return childCtx, nil
}

// resolveName returns the target template name for a lazy include,
// re-validating dynamic names against fs.ValidPath (defense in depth).
func (n *IncludeNode) resolveName(ctx *RenderContext) (string, error) {
	if n.pathExpr == nil {
		return n.staticName, nil
	}
	val, err := n.pathExpr.Evaluate(ctx)
	if err != nil {
		return "", err
	}
	name, ok := val.Interface().(string)
	if !ok {
		return "", fmt.Errorf("%w: got %T", ErrIncludePathNotString, val.Interface())
	}
	if err := ValidateName(name); err != nil {
		return "", err
	}
	return name, nil
}

// parseIncludeTag parses {% include %}.
//
// Syntax:
//
//	{% include "path/to/template.html" %}
//	{% include name_expr %}
//	{% include "card.html" with title="Hi" count=3 %}
//	{% include "card.html" with title="Hi" only %}
//	{% include "card.html" only %}
//	{% include "card.html" if_exists %}
//	{% include "card.html" with k=v only if_exists %}
//
// String literals are loaded and compiled at parse time (fast fail on
// missing templates). Expressions are resolved at runtime (lazy mode).
// Parse-time circular references are automatically downgraded to lazy
// to support recursive template patterns.
func parseIncludeTag(_ *Parser, start *Token, args *Parser) (Statement, error) {
	node := &IncludeNode{Line: start.Line, Col: start.Col}

	tok := args.Current()
	if tok == nil {
		return nil, newParseError("include: missing path", start.Line, start.Col)
	}

	if tok.Type == TokenString {
		node.staticName = tok.Value
		args.Advance()
	} else {
		// Dynamic path: parse as expression and mark lazy.
		expr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		node.pathExpr = expr
		node.lazy = true
	}

	if err := parseIncludeOptions(node, args); err != nil {
		return nil, err
	}

	// parseIncludeTag is only reachable when a layout-enabled engine is present
	// (the tag is not in the global registry), so args.Engine() is non-nil here.
	engine := args.Engine()

	// Dynamic paths always resolve at runtime.
	if node.pathExpr != nil {
		node.lazy = true
		return node, nil
	}

	// Detect parse-time circular reference and downgrade to lazy.
	// Without this, mutual includes (A→B→A) cause infinite parse recursion.
	if engine.isParsing(node.staticName) {
		node.lazy = true
		return node, nil
	}

	prepared, err := engine.Load(node.staticName)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			if node.ifExists {
				// Silently accept missing template — execute is a no-op.
				node.lazy = true
				return node, nil
			}
			return nil, fmt.Errorf("include: %w", err)
		}
		return nil, err
	}
	node.prepared = prepared
	return node, nil
}

// parseIncludeOptions consumes the "with k=v …", "only", and "if_exists"
// trailing options from an include tag's argument parser.
func parseIncludeOptions(node *IncludeNode, args *Parser) error {
	for args.Remaining() > 0 {
		tok := args.Current()
		if tok == nil {
			break
		}
		if tok.Type != TokenIdentifier {
			return newParseError(
				fmt.Sprintf("include: unexpected token %q", tok.Value),
				tok.Line, tok.Col)
		}

		switch tok.Value {
		case "with":
			args.Advance()
			if err := parseIncludeWithPairs(node, args); err != nil {
				return err
			}
		case "only":
			args.Advance()
			node.only = true
		case "if_exists":
			args.Advance()
			node.ifExists = true
		default:
			return newParseError(
				fmt.Sprintf("include: unknown option %q", tok.Value),
				tok.Line, tok.Col)
		}
	}
	return nil
}

// parseIncludeWithPairs consumes one or more "key=expression" bindings.
// It stops at "only", "if_exists", or end-of-args.
func parseIncludeWithPairs(node *IncludeNode, args *Parser) error {
	for args.Remaining() > 0 {
		tok := args.Current()
		if tok == nil {
			break
		}
		// Stop when we hit a trailing option keyword.
		if tok.Type == TokenIdentifier && (tok.Value == "only" || tok.Value == "if_exists") {
			return nil
		}
		if tok.Type != TokenIdentifier {
			return newParseError(
				fmt.Sprintf("include: expected identifier, got %q", tok.Value),
				tok.Line, tok.Col)
		}
		name := tok.Value
		args.Advance()

		if eq := args.Current(); eq == nil || eq.Type != TokenSymbol || eq.Value != "=" {
			return newParseError(
				fmt.Sprintf("include: expected '=' after %q", name),
				tok.Line, tok.Col)
		}
		args.Advance() // consume '='

		expr, err := args.ParseExpression()
		if err != nil {
			return err
		}
		node.withPairs = append(node.withPairs, withPair{name: name, expr: expr})
	}
	return nil
}
