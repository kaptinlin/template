package template

// parseForTag parses the for tag.
// {% for item in items %}...{% endfor %}
// {% for key, value in dict %}...{% endfor %}
func parseForTag(doc *Parser, start *Token, arguments *Parser) (Statement, error) {
	// 1. Parse loop variable names.
	var loopVars []string

	firstVar, err := arguments.ExpectIdentifier()
	if err != nil {
		return nil, arguments.Error(ErrExpectedVariable.Error())
	}
	loopVars = append(loopVars, firstVar.Value)

	// Optional second variable (key, value).
	if arguments.Match(TokenSymbol, ",") != nil {
		secondVar, err := arguments.ExpectIdentifier()
		if err != nil {
			return nil, arguments.Error(ErrExpectedSecondVariable.Error())
		}
		loopVars = append(loopVars, secondVar.Value)
	}

	// 2. Expect the "in" keyword.
	cur := arguments.Current()
	if cur == nil || cur.Type != TokenIdentifier || cur.Value != "in" {
		return nil, arguments.Error(ErrExpectedInKeyword.Error())
	}
	arguments.Advance()

	// 3. Parse the iterable expression.
	collection, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error(ErrUnexpectedTokensAfterCollection.Error())
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
		return nil, argParser.Error(ErrEndforNoArgs.Error())
	}

	return &ForNode{
		LoopVars:   loopVars,
		Collection: collection,
		Body:       convertStatementsToNodes(body),
		Line:       start.Line,
		Col:        start.Col,
	}, nil
}
