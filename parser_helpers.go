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

// peek returns the token at a relative offset without advancing.
func (p *Parser) peek(offset int) *Token {
	pos := p.pos + offset
	if pos >= len(p.tokens) {
		return nil
	}
	return p.tokens[pos]
}

// Remaining returns the number of unconsumed tokens.
func (p *Parser) Remaining() int {
	return len(p.tokens) - p.pos
}

// expect checks whether the current token matches tokenType.
// If it matches, it consumes and returns the token; otherwise it returns an error.
func (p *Parser) expect(tokenType TokenType) (*Token, error) {
	tok := p.Current()
	if tok == nil {
		return nil, p.Error("unexpected end of input")
	}
	if tok.Type != tokenType {
		return nil, p.Errorf("expected %s, got %s", tokenType, tok.Type)
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

// collectUntil collects tokens until tokenType is reached (exclusive).
func (p *Parser) collectUntil(tokenType TokenType) []*Token {
	// Estimate remaining tokens for preallocation.
	remaining := p.Remaining()
	if remaining <= 0 {
		return nil
	}
	tokens := make([]*Token, 0, remaining)
	for tok := p.Current(); tok != nil && tok.Type != tokenType; tok = p.Current() {
		tokens = append(tokens, tok)
		p.Advance()
	}
	return tokens
}

// isEndTag reports whether the current position points to one of the given end tags.
// End-tag format: {% tagname %}
func (p *Parser) isEndTag(endTags ...string) bool {
	tok := p.Current()
	if tok == nil || tok.Type != TokenTagBegin {
		return false
	}

	// Check whether the next token is a tag name.
	next := p.peek(1)
	if next == nil || next.Type != TokenIdentifier {
		return false
	}

	// Check whether it matches one of the requested end tags.
	for _, endTag := range endTags {
		if next.Value == endTag {
			return true
		}
	}

	return false
}

// endTagName returns the current end-tag name.
// It assumes the current token is {%.
func (p *Parser) endTagName() string {
	next := p.peek(1)
	if next != nil && next.Type == TokenIdentifier {
		return next.Value
	}
	return ""
}

// Error creates a parse error at the current token position.
func (p *Parser) Error(msg string) error {
	tok := p.Current()
	if tok != nil {
		return &ParseError{
			Message: msg,
			Line:    tok.Line,
			Col:     tok.Col,
		}
	}
	return &ParseError{Message: msg}
}

// Errorf creates a formatted parse error.
func (p *Parser) Errorf(format string, args ...any) error {
	return p.Error(fmt.Sprintf(format, args...))
}

func convertStatementsToNodes(stmts []Statement) []Node {
	if len(stmts) == 0 {
		return nil
	}
	nodes := make([]Node, 0, len(stmts))
	for _, stmt := range stmts {
		if stmt != nil {
			nodes = append(nodes, stmt)
		}
	}
	return nodes
}
