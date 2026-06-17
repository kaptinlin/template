package template

// parser consumes tokens and builds an AST.
type parser struct {
	tokens []*token
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
	blocks map[string]*blockNode

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

// newParser creates a new parser.
func newParser(tokens []*token) *parser {
	return &parser{tokens: tokens}
}

// Engine returns the template engine associated with this parser, if any.
// Tag parsers can use this to load referenced templates at parse time.
func (p *parser) Engine() *Engine {
	return p.engine
}

// Parse parses the entire template and returns AST statement nodes.
func (p *parser) Parse() ([]statement, error) {
	estimated := max(len(p.tokens)/4, 4)
	nodes := make([]statement, 0, estimated)

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
func (p *parser) parseNext() (statement, error) {
	tok := p.Current()
	if tok == nil || tok.Type == tokenEOF {
		return nil, nil
	}

	switch tok.Type {
	case tokenText:
		return p.parseText(), nil
	case tokenVarBegin:
		return p.parseVariable()
	case tokenTagBegin:
		return p.parseTag()
	case tokenError, tokenEOF, tokenVarEnd, tokenTagEnd,
		tokenIdentifier, tokenString, tokenNumber, tokenSymbol:
		return nil, p.Errorf("unexpected token: %s", tok.Type)
	}
	return nil, p.Errorf("unexpected token: %s", tok.Type)
}

// parseText parses a plain text node.
func (p *parser) parseText() statement {
	tok := p.Current()
	p.Advance()
	for _, r := range tok.value {
		switch r {
		case ' ', '\t', '\r', '\n':
			continue
		default:
			p.hasNonTrivialContent = true
			return newTextNode(tok.value, tok.Line, tok.Col)
		}
	}
	return newTextNode(tok.value, tok.Line, tok.Col)
}

// parseVariable parses a variable output: {{ expression }}.
func (p *parser) parseVariable() (statement, error) {
	p.hasNonTrivialContent = true
	start := p.Current()
	p.Advance() // Skip {{.

	tokens := p.collectUntil(tokenVarEnd)
	if len(tokens) == 0 {
		return nil, newParseError("empty variable expression", start.Line, start.Col)
	}

	cur := p.Current()
	if cur == nil || cur.Type != tokenVarEnd {
		return nil, newParseError("expected }}", start.Line, start.Col)
	}
	p.Advance() // Consume }}.

	ep := p.newExprParser(tokens)
	expr, err := ep.ParseExpression()
	if err != nil {
		return nil, err
	}

	return newOutputNode(expr, start.Line, start.Col), nil
}

// parseTag parses a tag: {% tag_name arguments %}.
func (p *parser) parseTag() (statement, error) {
	start := p.Current()
	p.Advance() // Skip {%.

	name := p.Current()
	if name == nil || name.Type != tokenIdentifier {
		return nil, newParseError("expected tag name", start.Line, start.Col)
	}
	p.Advance() // Consume tag name.

	// Any tag other than extends marks the template as "has content",
	// so a later {% extends %} will be rejected as not-first.
	if name.value != "extends" {
		p.hasNonTrivialContent = true
	}

	// Look up the tag parser. Templates compiled through an Engine consult
	// the engine-local registry first, layered over the built-in registry.
	var tagParser tagParser
	var found bool
	if p.engine != nil && p.engine.tags != nil {
		tagParser, found = p.engine.tags.Get(name.value)
	} else {
		tagParser, found = defaultTagRegistry.Get(name.value)
	}
	if !found {
		msg := "unknown tag: " + name.value
		if hint, found := misusedTagHints[name.value]; found {
			msg += " (" + hint + ")"
		}
		return nil, newParseError(msg, name.Line, name.Col)
	}

	args := p.collectUntil(tokenTagEnd)

	cur := p.Current()
	if cur == nil || cur.Type != tokenTagEnd {
		return nil, newParseError("expected %}", start.Line, start.Col)
	}
	p.Advance() // Consume %}.

	argParser := newParser(args)
	argParser.engine = p.engine
	argParser.anchorLine = name.Line
	argParser.anchorCol = name.Col
	return tagParser(p, name, argParser)
}

// ParseUntil parses nodes until one of the given end tags is encountered.
//
// Returns the parsed nodes, the matched end-tag name, and any error.
func (p *parser) ParseUntil(endTags ...string) ([]statement, string, error) {
	var nodes []statement

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
func (p *parser) ParseUntilWithArgs(endTags ...string) ([]statement, string, *parser, error) {
	var nodes []statement

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
func (p *parser) ParseExpression() (expression, error) {
	if p.Remaining() == 0 {
		return nil, p.Error("unexpected end of expression")
	}
	ep := p.newExprParser(p.tokens[p.pos:])
	expr, err := ep.ParseExpression()
	if err != nil {
		return nil, err
	}
	p.pos += ep.pos
	return expr, nil
}

func (p *parser) newExprParser(tokens []*token) *exprParser {
	ep := newExprParser(tokens)
	if p.engine != nil && p.engine.filters != nil {
		ep.filters = p.engine.filters
	}
	return ep
}
