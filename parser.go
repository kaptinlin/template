package template

// Parser consumes tokens and builds an AST.
type Parser struct {
	tokens []*Token // All tokens.
	pos    int      // Current cursor position.
}

// NewParser creates a new Parser.
func NewParser(tokens []*Token) *Parser {
	return &Parser{tokens: tokens}
}

// Parse parses the entire template and returns AST statement nodes.
func (p *Parser) Parse() ([]Statement, error) {
	// Preallocate with a reasonable estimate based on token count.
	estimated := len(p.tokens) / 4
	if estimated < 4 {
		estimated = 4
	}
	nodes := make([]Statement, 0, estimated)

	for tok := p.Current(); tok != nil && tok.Type != TokenEOF; tok = p.Current() {
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

// parseNext parses the next node.
func (p *Parser) parseNext() (Statement, error) {
	tok := p.Current()
	if tok == nil || tok.Type == TokenEOF {
		return nil, nil
	}

	switch tok.Type {
	case TokenText:
		return p.parseText(), nil

	case TokenVarBegin: // {{
		return p.parseVariable()

	case TokenTagBegin: // {%
		return p.parseTag()

	case TokenError, TokenVarEnd, TokenTagEnd, TokenIdentifier, TokenString, TokenNumber, TokenSymbol:
		return nil, p.Errorf("unexpected token: %s", tok.Type)

	case TokenEOF:
		return nil, nil
	}

	return nil, nil
}

// parseText parses a plain text node.
func (p *Parser) parseText() Statement {
	tok := p.Current()
	p.Advance()
	return NewTextNode(tok.Value, tok.Line, tok.Col)
}

// parseVariable parses a variable output: {{ expression }}.
func (p *Parser) parseVariable() (Statement, error) {
	startToken := p.Current()
	p.Advance() // Skip {{.

	// Collect expression tokens until }} (exclusive).
	exprTokens := p.collectUntil(TokenVarEnd)
	if len(exprTokens) == 0 {
		return nil, &ParseError{
			Message: "empty variable expression",
			Line:    startToken.Line,
			Col:     startToken.Col,
		}
	}

	// Expect }}.
	if tok := p.Current(); tok == nil || tok.Type != TokenVarEnd {
		return nil, &ParseError{
			Message: "expected }}",
			Line:    startToken.Line,
			Col:     startToken.Col,
		}
	}
	p.Advance() // Consume }}.

	// Parse expression tokens with ExprParser.
	exprParser := NewExprParser(exprTokens)
	expr, err := exprParser.ParseExpression()
	if err != nil {
		return nil, err
	}

	return NewOutputNode(expr, startToken.Line, startToken.Col), nil
}

// parseTag parses a tag: {% tag_name arguments %}.
// This is the core of the tag registration mechanism: lookup and invoke the TagParser.
func (p *Parser) parseTag() (Statement, error) {
	startToken := p.Current()
	p.Advance() // Skip {%.

	// Read the tag name.
	tagNameToken := p.Current()
	if tagNameToken == nil || tagNameToken.Type != TokenIdentifier {
		return nil, &ParseError{
			Message: "expected tag name",
			Line:    startToken.Line,
			Col:     startToken.Col,
		}
	}

	tagName := tagNameToken.Value
	p.Advance() // Consume tag name.

	// Lookup the tag parser in the registry.
	tagParser, ok := Tag(tagName)
	if !ok {
		// Provide more specific messages for common mistakes.
		var errorMsg string
		switch tagName {
		case "elif":
			errorMsg = "unknown tag: elif (elif must be used inside an if block, not standalone)"
		case "else":
			errorMsg = "unknown tag: else (else must be used inside an if block, not standalone)"
		case "endif":
			errorMsg = "unknown tag: endif (endif must match a corresponding if tag)"
		case "endfor":
			errorMsg = "unknown tag: endfor (endfor must match a corresponding for tag)"
		default:
			errorMsg = "unknown tag: " + tagName
		}
		return nil, &ParseError{
			Message: errorMsg,
			Line:    tagNameToken.Line,
			Col:     tagNameToken.Col,
		}
	}

	// Collect argument tokens until %} (exclusive).
	argTokens := p.collectUntil(TokenTagEnd)

	// Expect %}.
	if tok := p.Current(); tok == nil || tok.Type != TokenTagEnd {
		return nil, &ParseError{
			Message: "expected %}",
			Line:    startToken.Line,
			Col:     startToken.Col,
		}
	}
	p.Advance() // Consume %}.

	// Create a parser for tag arguments and invoke the tag parser.
	argParser := NewParser(argTokens)
	return tagParser(p, tagNameToken, argParser)
}

// ParseUntil parses nodes until one of the given end tags is encountered.
// It is used by TagParser implementations to parse tag bodies.
//
// Parameters:
//
//	endTags: Possible closing tags, for example:
//	  ParseUntil("elif", "else", "endif")
//
// Returns:
//
//	[]Statement: Parsed nodes before the closing tag.
//	string: The closing tag name that was encountered.
//	error: Parse error.
func (p *Parser) ParseUntil(endTags ...string) ([]Statement, string, error) {
	var nodes []Statement

	for tok := p.Current(); tok != nil && tok.Type != TokenEOF; tok = p.Current() {
		// Check whether an end tag is reached.
		if p.isEndTag(endTags...) {
			return nodes, p.endTagName(), nil
		}

		// Not an end tag; continue parsing.
		node, err := p.parseNext()
		if err != nil {
			return nil, "", err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	// Reached EOF without finding a closing tag.
	return nil, "", p.Errorf("unexpected EOF, expected one of: %v", endTags)
}

// ParseUntilWithArgs parses nodes until one of the given end tags is encountered,
// and also returns a parser for the end-tag arguments.
// It is used by TagParser implementations for closing tags that may carry arguments
// (for example, elif).
//
// Returns:
//
//	[]Statement: Parsed nodes before the closing tag.
//	string: The closing tag name that was encountered.
//	*Parser: A parser over end-tag arguments.
//	error: Parse error.
func (p *Parser) ParseUntilWithArgs(endTags ...string) ([]Statement, string, *Parser, error) {
	var nodes []Statement

	for tok := p.Current(); tok != nil && tok.Type != TokenEOF; tok = p.Current() {
		// Check whether an end tag is reached.
		if p.isEndTag(endTags...) {
			p.Advance() // Skip {%.

			endTagToken := p.Current()
			endTagName := endTagToken.Value
			p.Advance() // Skip tag name.

			// Collect end-tag argument tokens.
			argTokens := p.collectUntil(TokenTagEnd)

			// Expect %}.
			if cur := p.Current(); cur == nil || cur.Type != TokenTagEnd {
				return nil, "", nil, &ParseError{
					Message: "expected %}",
					Line:    endTagToken.Line,
					Col:     endTagToken.Col,
				}
			}
			p.Advance() // Consume %}.

			// Create an argument parser.
			argParser := NewParser(argTokens)

			return nodes, endTagName, argParser, nil
		}

		// Not an end tag; continue parsing.
		node, err := p.parseNext()
		if err != nil {
			return nil, "", nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	// Reached EOF without finding a closing tag.
	return nil, "", nil, p.Errorf("unexpected EOF, expected one of: %v", endTags)
}

// ParseExpression parses an expression from the current token position.
// It is intended for TagParser implementations.
func (p *Parser) ParseExpression() (Expression, error) {
	exprParser := NewExprParser(p.tokens[p.pos:])
	expr, err := exprParser.ParseExpression()
	if err != nil {
		return nil, err
	}

	// Advance parser position by consumed expression tokens.
	p.pos += exprParser.pos

	return expr, nil
}
