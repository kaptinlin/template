package template

// Parser consumes tokens and builds an AST.
type Parser struct {
	tokens []*Token
	pos    int

	// anchorLine/anchorCol provide a fallback source position when this parser
	// is asked to report an error after its token stream has been fully consumed.
	anchorLine int
	anchorCol  int

	// engine is the owning template engine, if this parser is being used by
	// Engine.Load or Engine.ParseString. It is nil only for internal parsers
	// constructed without an owning engine.
	engine *Engine

	// parent is populated by parseExtendsTag when the template being
	// compiled starts with {% extends "name" %}.
	parent *Template

	// blocks collects {% block %} definitions as they are parsed.
	blocks map[string]*BlockNode

	// hasNonTrivialContent tracks whether any non-whitespace text,
	// variable output, or non-extends tag has been seen. Used to enforce
	// the "extends must be first" rule.
	hasNonTrivialContent bool
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

// Engine returns the template engine associated with this parser, if any.
// Tag parsers can use this to load referenced templates at parse time.
func (p *Parser) Engine() *Engine {
	return p.engine
}

// Parse parses the entire template and returns AST statement nodes.
func (p *Parser) Parse() ([]Statement, error) {
	estimated := max(len(p.tokens)/4, 4)
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
	if !isAllWhitespace(tok.Value) {
		p.hasNonTrivialContent = true
	}
	return NewTextNode(tok.Value, tok.Line, tok.Col)
}

func isAllWhitespace(s string) bool {
	for _, r := range s {
		switch r {
		case ' ', '\t', '\r', '\n':
			continue
		default:
			return false
		}
	}
	return true
}

// parseVariable parses a variable output: {{ expression }}.
func (p *Parser) parseVariable() (Statement, error) {
	p.hasNonTrivialContent = true
	start := p.Current()
	p.Advance() // Skip {{.

	tokens := p.collectUntil(TokenVarEnd)
	if len(tokens) == 0 {
		return nil, newParseError("empty variable expression", start.Line, start.Col)
	}

	cur := p.Current()
	if cur == nil || cur.Type != TokenVarEnd {
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

	// Any tag other than extends marks the template as "has content",
	// so a later {% extends %} will be rejected as not-first.
	if name.Value != "extends" {
		p.hasNonTrivialContent = true
	}

	// Look up the tag parser. Templates compiled through an Engine consult
	// the engine-local registry first, layered over the built-in registry.
	var fn TagParser
	var ok bool
	if p.engine != nil && p.engine.tags != nil {
		fn, ok = p.engine.tags.Get(name.Value)
	} else {
		fn, ok = defaultTagRegistry.Get(name.Value)
	}
	if !ok {
		msg := "unknown tag: " + name.Value
		if hint, found := misusedTagHints[name.Value]; found {
			msg += " (" + hint + ")"
		}
		return nil, newParseError(msg, name.Line, name.Col)
	}

	args := p.collectUntil(TokenTagEnd)

	cur := p.Current()
	if cur == nil || cur.Type != TokenTagEnd {
		return nil, newParseError("expected %}", start.Line, start.Col)
	}
	p.Advance() // Consume %}.

	argParser := NewParser(args)
	argParser.engine = p.engine
	argParser.anchorLine = name.Line
	argParser.anchorCol = name.Col
	return fn(p, name, argParser)
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
	if p.Remaining() == 0 {
		return nil, p.Error("unexpected end of expression")
	}
	ep := NewExprParser(p.tokens[p.pos:])
	expr, err := ep.ParseExpression()
	if err != nil {
		return nil, err
	}
	p.pos += ep.pos
	return expr, nil
}
