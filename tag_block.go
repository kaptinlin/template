package template

import (
	"bytes"
	"fmt"
	"io"
	"slices"
)

// Position returns the source position of the block.
func (n *BlockNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation.
func (n *BlockNode) String() string { return fmt.Sprintf("Block(%s)", n.Name) }

// Execute renders this block, resolving overrides across the extends
// chain. If the current template is not part of a chain (or this block
// is inside an included partial), the block simply renders its own
// body inline.
//
// The {{ block.super }} expression is supported by injecting a "block"
// variable into the execution context whose "super" field is the
// pre-rendered parent block body (already a SafeString so it survives
// HTML auto-escape untouched).
func (n *BlockNode) Execute(ctx *ExecutionContext, w io.Writer) error {
	if ctx.currentLeaf == nil {
		return renderBlockWithSuper(n, nil, ctx, w)
	}
	chain := collectBlockChain(ctx.currentLeaf, n.Name)
	if len(chain) == 0 {
		return renderBlockWithSuper(n, nil, ctx, w)
	}
	active := chain[len(chain)-1]
	parents := chain[:len(chain)-1]
	return renderBlockWithSuper(active, parents, ctx, w)
}

// collectBlockChain walks the extends chain from leaf to root,
// collecting every template that defines a block with the given name.
// The result is ordered oldest-first (root → leaf), so the last element
// is the most-child override.
func collectBlockChain(leaf *Template, name string) []*BlockNode {
	var chain []*BlockNode
	for t := leaf; t != nil; t = t.parent {
		if t.blocks == nil {
			continue
		}
		if b, ok := t.blocks[name]; ok {
			chain = append(chain, b)
		}
	}
	slices.Reverse(chain)
	return chain
}

// renderBlockWithSuper executes block's body, pre-rendering the parent
// chain's output and exposing it as {{ block.super }}.
func renderBlockWithSuper(block *BlockNode, parents []*BlockNode, ctx *ExecutionContext, w io.Writer) error {
	var superVal SafeString
	if len(parents) > 0 {
		var buf bytes.Buffer
		parent := parents[len(parents)-1]
		rest := parents[:len(parents)-1]
		if err := renderBlockWithSuper(parent, rest, ctx, &buf); err != nil {
			return err
		}
		superVal = SafeString(buf.String())
	}

	if ctx.Private == nil {
		ctx.Private = NewContext()
	}
	prev, had := ctx.Private["block"]
	ctx.Private["block"] = map[string]any{"super": superVal}
	defer func() {
		if had {
			ctx.Private["block"] = prev
		} else {
			delete(ctx.Private, "block")
		}
	}()

	return executeBody(block.Body, ctx, w)
}

// parseBlockTag parses {% block name %}...{% endblock [name] %}.
//
// The block name is stored in the owning parser's block map so the
// final Template carries its own block definitions (used for override
// resolution in extends chains).
func parseBlockTag(doc *Parser, start *Token, args *Parser) (Statement, error) {
	nameTok := args.Current()
	if nameTok == nil || nameTok.Type != TokenIdentifier {
		return nil, newParseError("block: expected identifier", start.Line, start.Col)
	}
	args.Advance()
	blockName := nameTok.Value

	if args.Remaining() > 0 {
		return nil, newParseError("block: unexpected tokens after name", start.Line, start.Col)
	}

	body, endTag, endArgs, err := doc.ParseUntilWithArgs("endblock")
	if err != nil {
		return nil, err
	}
	if endTag != "endblock" {
		return nil, doc.Errorf("block %q: expected endblock, got %q", blockName, endTag)
	}
	// Optional endblock name must match the opening name.
	if endArgs.Remaining() > 0 {
		endNameTok := endArgs.Current()
		if endNameTok == nil || endNameTok.Type != TokenIdentifier {
			return nil, endArgs.Error("endblock: expected identifier or nothing")
		}
		if endNameTok.Value != blockName {
			return nil, endArgs.Errorf(
				"endblock name %q does not match block name %q",
				endNameTok.Value, blockName)
		}
		endArgs.Advance()
		if endArgs.Remaining() > 0 {
			return nil, endArgs.Error("endblock: unexpected tokens")
		}
	}

	node := &BlockNode{
		Name: blockName,
		Body: convertStatementsToNodes(body),
		Line: start.Line,
		Col:  start.Col,
	}

	// Register the block in the owning parser's map. Duplicates are an
	// error inside a single template (different templates may each
	// define a "content" block, which is the whole point of extends).
	if doc.blocks == nil {
		doc.blocks = make(map[string]*BlockNode)
	}
	if _, exists := doc.blocks[blockName]; exists {
		return nil, fmt.Errorf("%w: %q", ErrBlockRedefined, blockName)
	}
	doc.blocks[blockName] = node

	return node, nil
}
