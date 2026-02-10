package template

// parseIfTag parses the if tag.
// {% if condition %}...{% elif condition %}...{% else %}...{% endif %}
func parseIfTag(doc *Parser, start *Token, arguments *Parser) (Statement, error) {
	var branches []IfBranch
	var elseBody []Node

	// 1. Parse the first if condition.
	condition, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error(ErrUnexpectedTokensAfterCondition.Error())
	}

	// 2. Parse the if body (until elif/else/endif).
	body, endTag, argParser, err := doc.ParseUntilWithArgs("elif", "else", "endif")
	if err != nil {
		return nil, err
	}

	branches = append(branches, IfBranch{
		Condition: condition,
		Body:      convertStatementsToNodes(body),
	})

	// 3. Process elif and else branches.
	hasElse := false
	for endTag == "elif" || endTag == "else" {
		switch endTag {
		case "elif":
			if hasElse {
				return nil, doc.Error(ErrElifAfterElse.Error())
			}

			// Parse the elif condition. Use = to avoid shadowing outer condition/err.
			condition, err = argParser.ParseExpression()
			if err != nil {
				return nil, err
			}

			if argParser.Remaining() > 0 {
				return nil, argParser.Error(ErrUnexpectedTokensAfterCondition.Error())
			}

			// Parse the elif body.
			var nextBody []Statement
			var nextTag string
			var nextArgParser *Parser
			nextBody, nextTag, nextArgParser, err = doc.ParseUntilWithArgs("elif", "else", "endif")
			if err != nil {
				return nil, err
			}

			branches = append(branches, IfBranch{
				Condition: condition,
				Body:      convertStatementsToNodes(nextBody),
			})

			endTag = nextTag
			argParser = nextArgParser

		case "else":
			if hasElse {
				return nil, doc.Error(ErrMultipleElseStatements.Error())
			}
			hasElse = true

			if argParser.Remaining() > 0 {
				return nil, argParser.Error(ErrElseNoArgs.Error())
			}

			// Parse the else body (must end with endif).
			var elseStmts []Statement
			var nextTag string
			var nextArgParser *Parser
			elseStmts, nextTag, nextArgParser, err = doc.ParseUntilWithArgs("endif")
			if err != nil {
				return nil, err
			}

			elseBody = convertStatementsToNodes(elseStmts)

			endTag = nextTag
			argParser = nextArgParser
		}
	}

	// 4. Validate that the final closing tag is endif.
	if endTag != "endif" {
		return nil, doc.Errorf("expected endif, got %s", endTag)
	}

	if argParser.Remaining() > 0 {
		return nil, argParser.Error(ErrEndifNoArgs.Error())
	}

	return &IfNode{
		Branches: branches,
		ElseBody: elseBody,
		Line:     start.Line,
		Col:      start.Col,
	}, nil
}
