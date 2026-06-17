package template

import (
	"fmt"
	"slices"
)

// Current returns the current token without advancing the parser.
func (p *parser) Current() *token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	return p.tokens[p.pos]
}

// Advance moves to the next token.
func (p *parser) Advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

// notEOF reports whether the parser has more non-EOF tokens.
func (p *parser) notEOF() bool {
	tok := p.Current()
	return tok != nil && tok.Type != tokenEOF
}

// peek returns the token at a relative offset without advancing.
func (p *parser) peek(offset int) *token {
	pos := p.pos + offset
	if pos < 0 || pos >= len(p.tokens) {
		return nil
	}
	return p.tokens[pos]
}

// Remaining returns the number of unconsumed tokens.
func (p *parser) Remaining() int {
	return len(p.tokens) - p.pos
}

// expect checks whether the current token matches the given type.
// If it matches, it consumes and returns the token; otherwise it returns an error.
func (p *parser) expect(typ tokenType) (*token, error) {
	tok := p.Current()
	if tok == nil {
		return nil, p.Error("unexpected end of input")
	}
	if tok.Type != typ {
		return nil, p.Errorf("expected %s, got %s", typ, tok.Type)
	}
	p.Advance()
	return tok, nil
}

// ExpectIdentifier expects and consumes an identifier token.
func (p *parser) ExpectIdentifier() (*token, error) {
	return p.expect(tokenIdentifier)
}

// Match consumes and returns the current token if type and value match.
// It returns nil when there is no match.
func (p *parser) Match(tokenType tokenType, value string) *token {
	tok := p.Current()
	if tok != nil && tok.Type == tokenType && tok.value == value {
		p.Advance()
		return tok
	}
	return nil
}

// collectUntil collects tokens until the given type is reached (exclusive).
func (p *parser) collectUntil(typ tokenType) []*token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	tokens := make([]*token, 0, 4)
	for tok := p.Current(); tok != nil && tok.Type != typ; tok = p.Current() {
		tokens = append(tokens, tok)
		p.Advance()
	}
	if len(tokens) == 0 {
		return nil
	}
	return tokens
}

// isEndTag reports whether the current position is one of the given end tags.
// End-tag format: {% tagname ... %}
func (p *parser) isEndTag(tags ...string) bool {
	cur := p.Current()
	if cur == nil || cur.Type != tokenTagBegin {
		return false
	}
	next := p.peek(1)
	if next == nil || next.Type != tokenIdentifier {
		return false
	}
	return slices.Contains(tags, next.value)
}

// endTagName returns the tag name at the current position.
// It assumes the current token is {% (tokenTagBegin).
func (p *parser) endTagName() string {
	if next := p.peek(1); next != nil && next.Type == tokenIdentifier {
		return next.value
	}
	return ""
}

// consumeEndTag consumes an end tag ({% name ... %}) and returns
// the tag name and a parser over any arguments.
func (p *parser) consumeEndTag() (string, *parser, error) {
	p.Advance() // Skip {%.

	name := p.Current()
	tag := name.value
	p.Advance() // Skip tag name.

	args := p.collectUntil(tokenTagEnd)

	if cur := p.Current(); cur == nil || cur.Type != tokenTagEnd {
		return "", nil, newParseError("expected %}", name.Line, name.Col)
	}
	p.Advance() // Consume %}.

	argParser := newParser(args)
	argParser.engine = p.engine
	argParser.anchorLine = name.Line
	argParser.anchorCol = name.Col
	return tag, argParser, nil
}

// Error creates a parse error at the current token position.
func (p *parser) Error(msg string) error {
	if tok := p.Current(); tok != nil {
		return newParseError(msg, tok.Line, tok.Col)
	}
	if p.anchorLine > 0 || p.anchorCol > 0 {
		return newParseError(msg, p.anchorLine, p.anchorCol)
	}
	return &ParseError{Message: msg}
}

// Errorf creates a formatted parse error at the current token position.
func (p *parser) Errorf(format string, args ...any) error {
	return p.Error(fmt.Sprintf(format, args...))
}

// newParseError creates a ParseError with the given message and position.
func newParseError(msg string, line, col int) *ParseError {
	return &ParseError{Message: msg, Line: line, Col: col}
}

// convertStatementsToNodes converts statements to nodes, filtering nil entries.
func convertStatementsToNodes(stmts []statement) []node {
	if len(stmts) == 0 {
		return nil
	}

	nodes := make([]node, 0, len(stmts))
	for _, s := range stmts {
		if s != nil {
			nodes = append(nodes, s)
		}
	}

	if len(nodes) == 0 {
		return nil
	}
	return nodes
}
