package template

// Parser consumes tokens and builds an AST.
type Parser struct {
	tokens []*Token
	pos    int
}

// misusedTagHints maps end/branch tags to hints explaining correct usage.
var misusedTagHints = map[string]string{
	"elif":   "elif must be used inside an if block, not standalone",
	"else":   "else must be used inside an if block, not standalone",
	"endif":  "endif must match a corresponding if tag",
	"endfor": "endfor must match a corresponding for tag",
}

// NewParser creates a new Parser.
func NewParser(tokens []*Token) *Parser {
	return &Parser{tokens: tokens}
}

// Parse parses the entire template and returns AST statement nodes.
func (p *Parser) Parse() ([]Statement, error) {
	estimated := len(p.tokens) / 4
	if estimated < 4 {
		estimated = 4
	}
	nodes := make([]Statement, 0, estimated)

	for p.notEOF() {
		node, err := p.parseNext()
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	if len(nodes) == 0 {
		return nil, nil
	}
	return nodes, nil
}

// parseNext parses the next node based on the current token type.
func (p *Parser) parseNext() (Statement, error) {
	tok := p.Current()
	if tok == nil || tok.Type == TokenEOF {
		return nil, nil
	}

	switch tok.Type {
	case TokenText:
		return p.parseText(), nil
	case TokenVarBegin:
		return p.parseVariable()
	case TokenTagBegin:
		return p.parseTag()
	case TokenError, TokenEOF, TokenVarEnd, TokenTagEnd,
		TokenIdentifier, TokenString, TokenNumber, TokenSymbol:
		return nil, p.Errorf("unexpected token: %s", tok.Type)
	}
	return nil, p.Errorf("unexpected token: %s", tok.Type)
}

// parseText parses a plain text node.
func (p *Parser) parseText() Statement {
	tok := p.Current()
	p.Advance()
	return NewTextNode(tok.Value, tok.Line, tok.Col)
}

// parseVariable parses a variable output: {{ expression }}.
func (p *Parser) parseVariable() (Statement, error) {
	start := p.Current()
	p.Advance() // Skip {{.

	tokens := p.collectUntil(TokenVarEnd)
	if len(tokens) == 0 {
		return nil, newParseError("empty variable expression", start.Line, start.Col)
	}

	if cur := p.Current(); cur == nil || cur.Type != TokenVarEnd {
		return nil, newParseError("expected }}", start.Line, start.Col)
	}
	p.Advance() // Consume }}.

	ep := NewExprParser(tokens)
	expr, err := ep.ParseExpression()
	if err != nil {
		return nil, err
	}

	return NewOutputNode(expr, start.Line, start.Col), nil
}

// parseTag parses a tag: {% tag_name arguments %}.
func (p *Parser) parseTag() (Statement, error) {
	start := p.Current()
	p.Advance() // Skip {%.

	name := p.Current()
	if name == nil || name.Type != TokenIdentifier {
		return nil, newParseError("expected tag name", start.Line, start.Col)
	}
	p.Advance() // Consume tag name.

	// Look up the tag parser in the registry.
	fn, ok := Tag(name.Value)
	if !ok {
		msg := "unknown tag: " + name.Value
		if hint, found := misusedTagHints[name.Value]; found {
			msg += " (" + hint + ")"
		}
		return nil, newParseError(msg, name.Line, name.Col)
	}

	args := p.collectUntil(TokenTagEnd)

	if cur := p.Current(); cur == nil || cur.Type != TokenTagEnd {
		return nil, newParseError("expected %}", start.Line, start.Col)
	}
	p.Advance() // Consume %}.

	return fn(p, name, NewParser(args))
}

// ParseUntil parses nodes until one of the given end tags is encountered.
//
// Returns the parsed nodes, the matched end-tag name, and any error.
func (p *Parser) ParseUntil(endTags ...string) ([]Statement, string, error) {
	var nodes []Statement

	for p.notEOF() {
		if p.isEndTag(endTags...) {
			return nodes, p.endTagName(), nil
		}

		node, err := p.parseNext()
		if err != nil {
			return nil, "", err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	return nil, "", p.Errorf("unexpected EOF, expected one of: %v", endTags)
}

// ParseUntilWithArgs parses nodes until one of the given end tags is
// encountered, and also returns a parser for the end-tag arguments.
func (p *Parser) ParseUntilWithArgs(endTags ...string) ([]Statement, string, *Parser, error) {
	var nodes []Statement

	for p.notEOF() {
		if p.isEndTag(endTags...) {
			tag, argParser, err := p.consumeEndTag()
			if err != nil {
				return nil, "", nil, err
			}
			return nodes, tag, argParser, nil
		}

		node, err := p.parseNext()
		if err != nil {
			return nil, "", nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	return nil, "", nil, p.Errorf("unexpected EOF, expected one of: %v", endTags)
}

// ParseExpression parses an expression from the current token position.
func (p *Parser) ParseExpression() (Expression, error) {
	ep := NewExprParser(p.tokens[p.pos:])
	expr, err := ep.ParseExpression()
	if err != nil {
		return nil, err
	}
	p.pos += ep.pos
	return expr, nil
}
