package template

// parseForTag parses a for-endfor loop block into a forNode.
//
// Syntax:
//
//	{% for item in items %}...{% endfor %}
//	{% for key, value in dict %}...{% endfor %}
func parseForTag(doc *parser, start *token, args *parser) (statement, error) {
	first, err := args.ExpectIdentifier()
	if err != nil {
		return nil, args.Error(errExpectedVariable.Error())
	}
	vars := []string{first.value}

	if args.Match(tokenSymbol, ",") != nil {
		second, err := args.ExpectIdentifier()
		if err != nil {
			return nil, args.Error(errExpectedSecondVariable.Error())
		}
		vars = append(vars, second.value)
	}

	if cur := args.Current(); cur == nil || cur.Type != tokenIdentifier || cur.value != "in" {
		return nil, args.Error(errExpectedInKeyword.Error())
	}
	args.Advance()

	collection, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	if args.Remaining() > 0 {
		return nil, args.Error(errUnexpectedTokensAfterCollection.Error())
	}

	body, tag, ap, err := doc.ParseUntilWithArgs("endfor")
	if err != nil {
		return nil, err
	}
	if tag != "endfor" {
		return nil, doc.Errorf("expected endfor, got %s", tag)
	}
	if ap.Remaining() > 0 {
		return nil, ap.Error(errEndforNoArgs.Error())
	}

	return &forNode{
		Vars:       vars,
		Collection: collection,
		Body:       convertStatementsToNodes(body),
		Line:       start.Line,
		Col:        start.Col,
	}, nil
}
