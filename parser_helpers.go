package template

import "fmt"

// Current returns the current token without advancing the parser.
func (p *Parser) Current() *Token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	return p.tokens[p.pos]
}

// Advance moves to the next token.
func (p *Parser) Advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

// notEOF reports whether the parser has more non-EOF tokens.
func (p *Parser) notEOF() bool {
	tok := p.Current()
	return tok != nil && tok.Type != TokenEOF
}

// peek returns the token at a relative offset without advancing.
func (p *Parser) peek(offset int) *Token {
	pos := p.pos + offset
	if pos < 0 || pos >= len(p.tokens) {
		return nil
	}
	return p.tokens[pos]
}

// Remaining returns the number of unconsumed tokens.
func (p *Parser) Remaining() int {
	return len(p.tokens) - p.pos
}

// expect checks whether the current token matches the given type.
// If it matches, it consumes and returns the token; otherwise it returns an error.
func (p *Parser) expect(typ TokenType) (*Token, error) {
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
func (p *Parser) ExpectIdentifier() (*Token, error) {
	return p.expect(TokenIdentifier)
}

// Match consumes and returns the current token if type and value match.
// It returns nil when there is no match.
func (p *Parser) Match(tokenType TokenType, value string) *Token {
	tok := p.Current()
	if tok != nil && tok.Type == tokenType && tok.Value == value {
		p.Advance()
		return tok
	}
	return nil
}

// collectUntil collects tokens until the given type is reached (exclusive).
func (p *Parser) collectUntil(typ TokenType) []*Token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	// Preallocate with a small estimate to reduce grow-copies.
	tokens := make([]*Token, 0, 4)
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
func (p *Parser) isEndTag(tags ...string) bool {
	cur := p.Current()
	if cur == nil || cur.Type != TokenTagBegin {
		return false
	}
	next := p.peek(1)
	if next == nil || next.Type != TokenIdentifier {
		return false
	}
	for _, tag := range tags {
		if next.Value == tag {
			return true
		}
	}
	return false
}

// endTagName returns the tag name at the current position.
// It assumes the current token is {% (TokenTagBegin).
func (p *Parser) endTagName() string {
	if next := p.peek(1); next != nil && next.Type == TokenIdentifier {
		return next.Value
	}
	return ""
}

// consumeEndTag consumes an end tag ({% name ... %}) and returns
// the tag name and a parser over any arguments.
func (p *Parser) consumeEndTag() (string, *Parser, error) {
	p.Advance() // Skip {%.

	name := p.Current()
	tag := name.Value
	p.Advance() // Skip tag name.

	args := p.collectUntil(TokenTagEnd)

	if cur := p.Current(); cur == nil || cur.Type != TokenTagEnd {
		return "", nil, newParseError("expected %}", name.Line, name.Col)
	}
	p.Advance() // Consume %}.

	return tag, NewParser(args), nil
}

// Error creates a parse error at the current token position.
func (p *Parser) Error(msg string) error {
	if tok := p.Current(); tok != nil {
		return newParseError(msg, tok.Line, tok.Col)
	}
	return &ParseError{Message: msg}
}

// Errorf creates a formatted parse error at the current token position.
func (p *Parser) Errorf(format string, args ...any) error {
	return p.Error(fmt.Sprintf(format, args...))
}

// newParseError creates a ParseError with the given message and position.
func newParseError(msg string, line, col int) *ParseError {
	return &ParseError{Message: msg, Line: line, Col: col}
}

// convertStatementsToNodes converts statements to nodes, filtering nil entries.
func convertStatementsToNodes(stmts []Statement) []Node {
	if len(stmts) == 0 {
		return nil
	}
	nodes := make([]Node, 0, len(stmts))
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
