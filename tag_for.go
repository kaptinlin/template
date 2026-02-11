package template

// parseForTag parses a for-endfor loop block into a ForNode.
//
// Syntax:
//
//	{% for item in items %}...{% endfor %}
//	{% for key, value in dict %}...{% endfor %}
func parseForTag(doc *Parser, start *Token, args *Parser) (Statement, error) {
	// Parse the first loop variable.
	first, err := args.ExpectIdentifier()
	if err != nil {
		return nil, args.Error(ErrExpectedVariable.Error())
	}
	vars := []string{first.Value}

	// Optional second variable (key, value).
	if args.Match(TokenSymbol, ",") != nil {
		second, err := args.ExpectIdentifier()
		if err != nil {
			return nil, args.Error(ErrExpectedSecondVariable.Error())
		}
		vars = append(vars, second.Value)
	}

	// Expect the "in" keyword.
	if cur := args.Current(); cur == nil || cur.Type != TokenIdentifier || cur.Value != "in" {
		return nil, args.Error(ErrExpectedInKeyword.Error())
	}
	args.Advance()

	// Parse the iterable expression.
	collection, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	if args.Remaining() > 0 {
		return nil, args.Error(ErrUnexpectedTokensAfterCollection.Error())
	}

	// Parse the loop body until endfor.
	body, tag, ap, err := doc.ParseUntilWithArgs("endfor")
	if err != nil {
		return nil, err
	}
	if tag != "endfor" {
		return nil, doc.Errorf("expected endfor, got %s", tag)
	}
	if ap.Remaining() > 0 {
		return nil, ap.Error(ErrEndforNoArgs.Error())
	}

	return &ForNode{
		Vars:       vars,
		Collection: collection,
		Body:       convertStatementsToNodes(body),
		Line:       start.Line,
		Col:        start.Col,
	}, nil
}
