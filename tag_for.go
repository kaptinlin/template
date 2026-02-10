package template

// parseForTag parses the for tag.
// {% for item in items %}...{% endfor %}
// {% for key, value in dict %}...{% endfor %}
func parseForTag(doc *Parser, start *Token, arguments *Parser) (Statement, error) {
	// 1. Parse loop variable names.
	var loopVars []string

	// First variable.
	firstVar, err := arguments.ExpectIdentifier()
	if err != nil {
		return nil, arguments.Error("expected variable name")
	}
	loopVars = append(loopVars, firstVar.Value)

	// Optional second variable (key, value).
	if arguments.Match(TokenSymbol, ",") != nil {
		secondVar, err := arguments.ExpectIdentifier()
		if err != nil {
			return nil, arguments.Error("expected second variable name after comma")
		}
		loopVars = append(loopVars, secondVar.Value)
	}

	// 2. Expect the "in" keyword.
	if arguments.Current() == nil ||
		arguments.Current().Type != TokenIdentifier ||
		arguments.Current().Value != "in" {
		return nil, arguments.Error("expected 'in' keyword")
	}
	arguments.Advance() // Skip "in".

	// 3. Parse the iterable expression.
	collection, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("unexpected tokens after collection")
	}

	// 4. Parse the for body until endfor.
	body, endTag, argParser, err := doc.ParseUntilWithArgs("endfor")
	if err != nil {
		return nil, err
	}

	if endTag != "endfor" {
		return nil, doc.Errorf("expected endfor, got %s", endTag)
	}

	if argParser.Remaining() > 0 {
		return nil, argParser.Error("endfor does not take arguments")
	}

	// Convert []Statement to []Node.
	bodyNodes := convertStatementsToNodes(body)

	// 5. Return the parsed ForNode.
	return &ForNode{
		LoopVars:   loopVars,
		Collection: collection,
		Body:       bodyNodes,
		Line:       start.Line,
		Col:        start.Col,
	}, nil
}
