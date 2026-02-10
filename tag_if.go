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
		return nil, arguments.Error("unexpected tokens after condition")
	}

	// 2. Parse the if body (until elif/else/endif).
	body, endTag, argParser, err := doc.ParseUntilWithArgs("elif", "else", "endif")
	if err != nil {
		return nil, err
	}

	// Convert []Statement to []Node.
	bodyNodes := convertStatementsToNodes(body)

	branches = append(branches, IfBranch{
		Condition: condition,
		Body:      bodyNodes,
	})

	// 3. Process elif and else branches.
	hasElse := false
	for endTag == "elif" || endTag == "else" {
		switch endTag {
		case "elif":
			// elif cannot appear after else.
			if hasElse {
				return nil, doc.Error("elif cannot appear after else")
			}

			// Parse the elif condition.
			condition, err := argParser.ParseExpression()
			if err != nil {
				return nil, err
			}

			if argParser.Remaining() > 0 {
				return nil, argParser.Error("unexpected tokens after condition")
			}

			// Parse the elif body.
			body, nextTag, nextArgParser, err := doc.ParseUntilWithArgs("elif", "else", "endif")
			if err != nil {
				return nil, err
			}

			// Convert to []Node.
			elifBodyNodes := convertStatementsToNodes(body)

			branches = append(branches, IfBranch{
				Condition: condition,
				Body:      elifBodyNodes,
			})

			endTag = nextTag
			argParser = nextArgParser

		case "else":
			// Ensure there is only one else branch.
			if hasElse {
				return nil, doc.Error("multiple 'else' statements found in if block. Use 'elif' for additional conditions")
			}
			hasElse = true

			// else does not accept arguments.
			if argParser.Remaining() > 0 {
				return nil, argParser.Error("else does not take arguments")
			}

			// Parse the else body (must end with endif).
			body, nextTag, nextArgParser, err := doc.ParseUntilWithArgs("endif")
			if err != nil {
				return nil, err
			}

			// Convert to []Node.
			elseBody = convertStatementsToNodes(body)

			endTag = nextTag
			argParser = nextArgParser
		}
	}

	// 4. Validate that the final closing tag is endif.
	if endTag != "endif" {
		return nil, doc.Errorf("expected endif, got %s", endTag)
	}

	if argParser.Remaining() > 0 {
		return nil, argParser.Error("endif does not take arguments")
	}

	// 5. Return the parsed IfNode.
	return &IfNode{
		Branches: branches,
		ElseBody: elseBody,
		Line:     start.Line,
		Col:      start.Col,
	}, nil
}
