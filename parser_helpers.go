package template

import "fmt"

// Current returns the current token without advancing the parser.
func (p *Parser) Current() *Token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	return p.tokens[p.pos]
}

// current is the internal alias of Current.
func (p *Parser) current() *Token {
	return p.Current()
}

// Advance moves to the next token.
func (p *Parser) Advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

// advance is the internal alias of Advance.
func (p *Parser) advance() {
	p.Advance()
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
	tok := p.current()
	if tok == nil {
		return nil, p.Error("unexpected end of input")
	}
	if tok.Type != tokenType {
		return nil, p.Errorf("expected %s, got %s", tokenType, tok.Type)
	}
	p.advance()
	return tok, nil
}

// ExpectIdentifier expects and consumes an identifier token.
func (p *Parser) ExpectIdentifier() (*Token, error) {
	return p.expect(TokenIdentifier)
}

// Match consumes and returns the current token if type and value match.
// It returns nil when there is no match.
func (p *Parser) Match(tokenType TokenType, value string) *Token {
	tok := p.current()
	if tok != nil && tok.Type == tokenType && tok.Value == value {
		p.advance()
		return tok
	}
	return nil
}

// collectUntil collects tokens until tokenType is reached (exclusive).
func (p *Parser) collectUntil(tokenType TokenType) []*Token {
	var tokens []*Token
	for p.current() != nil && p.current().Type != tokenType {
		tokens = append(tokens, p.current())
		p.advance()
	}
	return tokens
}

// isEndTag reports whether the current position points to one of the given end tags.
// End-tag format: {% tagname %}
func (p *Parser) isEndTag(endTags ...string) bool {
	if p.current() == nil || p.current().Type != TokenTagBegin {
		return false
	}

	// Check whether the next token is a tag name.
	nextToken := p.peek(1)
	if nextToken == nil || nextToken.Type != TokenIdentifier {
		return false
	}

	// Check whether it matches one of the requested end tags.
	for _, endTag := range endTags {
		if nextToken.Value == endTag {
			return true
		}
	}

	return false
}

// getEndTagName returns the current end-tag name.
// It assumes the current token is {%.
func (p *Parser) getEndTagName() string {
	nextToken := p.peek(1)
	if nextToken != nil && nextToken.Type == TokenIdentifier {
		return nextToken.Value
	}
	return ""
}

// Error creates a parse error at the current token position.
func (p *Parser) Error(msg string) error {
	tok := p.current()
	if tok != nil {
		return &ParseError{
			Message: msg,
			Line:    tok.Line,
			Col:     tok.Col,
		}
	}
	return &ParseError{
		Message: msg,
		Line:    0,
		Col:     0,
	}
}

// Errorf creates a formatted parse error.
func (p *Parser) Errorf(format string, args ...interface{}) error {
	return p.Error(fmt.Sprintf(format, args...))
}

// errorf is the internal alias of Errorf.
func (p *Parser) errorf(format string, args ...interface{}) error {
	return p.Errorf(format, args...)
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
