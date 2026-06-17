package template

import (
	"errors"
	"fmt"
	"io"
)

// extendsNode is a marker for {% extends "parent" %}. It produces no
// output directly; inheritance is handled by the template execution path when
// it detects a non-nil parent field.
type extendsNode struct {
	Line int
	Col  int
}

// Position returns the source position of the extends tag.
func (n *extendsNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation.
func (n *extendsNode) String() string { return "Extends" }

// Execute is a no-op. The parent relationship is established at parse time and
// consumed by the template execution path.
func (n *extendsNode) Execute(_ *renderContext, _ io.Writer) error {
	return nil
}

// maxExtendsDepth caps the depth of the extends chain. Defends against
// accidental over-nesting. Real-world chains rarely exceed 3.
const maxExtendsDepth = 10

// parseExtendsTag parses {% extends "parent.html" %}.
//
// Constraints:
//   - Must be the first non-whitespace, non-comment tag in the template.
//   - Parent path must be a string literal (dynamic extends is not
//     supported; use Go-level template selection instead).
//   - Parent template is loaded and compiled at parse time.
//   - Circular chains (A extends B, B extends A) are rejected.
//
// Sets doc.parent on the owning parser so compilation can transfer the
// reference to the final Template.
func parseExtendsTag(doc *parser, start *token, args *parser) (statement, error) {
	if doc.hasNonTrivialContent {
		return nil, fmt.Errorf("%w at line %d", ErrExtendsNotFirst, start.Line)
	}
	if doc.parent != nil {
		return nil, newParseError(
			"extends: template already extends another",
			start.Line, start.Col)
	}

	tok := args.Current()
	if tok == nil {
		return nil, newParseError("extends: expected parent path", start.Line, start.Col)
	}
	if tok.Type != tokenString {
		return nil, fmt.Errorf("%w at line %d", ErrExtendsPathNotLiteral, start.Line)
	}
	parentName := tok.value
	args.Advance()

	if args.Remaining() > 0 {
		return nil, newParseError("extends: unexpected tokens after path", start.Line, start.Col)
	}

	// parseExtendsTag is only reachable when a layout-enabled engine is present
	// (the tag is not in the global registry), so doc.Engine() is non-nil here.
	engine := doc.Engine()

	parentTpl, err := engine.loadExtends(parentName)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("extends: %w", err)
	}

	// Enforce depth limit by walking the parent chain.
	depth := 1
	for t := parentTpl; t != nil; t = t.parent {
		depth++
		if depth > maxExtendsDepth {
			return nil, fmt.Errorf("%w: %d", ErrExtendsDepthExceeded, depth)
		}
	}

	doc.parent = parentTpl
	return &extendsNode{Line: start.Line, Col: start.Col}, nil
}
