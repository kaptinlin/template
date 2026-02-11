package template

// parseIfTag parses an if-elif-else-endif block into an IfNode.
//
// Syntax:
//
//	{% if condition %}...{% elif condition %}...{% else %}...{% endif %}
func parseIfTag(doc *Parser, start *Token, args *Parser) (Statement, error) {
	var branches []IfBranch
	var elseBody []Node

	// Parse the first condition.
	cond, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	if args.Remaining() > 0 {
		return nil, args.Error(ErrUnexpectedTokensAfterCondition.Error())
	}

	// Parse the if body until elif/else/endif.
	body, tag, ap, err := doc.ParseUntilWithArgs("elif", "else", "endif")
	if err != nil {
		return nil, err
	}
	branches = append(branches, IfBranch{
		Condition: cond,
		Body:      convertStatementsToNodes(body),
	})

	// Process elif and else branches.
	hasElse := false
	for tag == "elif" || tag == "else" {
		switch tag {
		case "elif":
			if hasElse {
				return nil, doc.Error(ErrElifAfterElse.Error())
			}
			cond, err = ap.ParseExpression()
			if err != nil {
				return nil, err
			}
			if ap.Remaining() > 0 {
				return nil, ap.Error(ErrUnexpectedTokensAfterCondition.Error())
			}
			next, nextTag, nextAP, err := doc.ParseUntilWithArgs("elif", "else", "endif")
			if err != nil {
				return nil, err
			}
			branches = append(branches, IfBranch{
				Condition: cond,
				Body:      convertStatementsToNodes(next),
			})
			tag = nextTag
			ap = nextAP

		case "else":
			if hasElse {
				return nil, doc.Error(ErrMultipleElseStatements.Error())
			}
			hasElse = true
			if ap.Remaining() > 0 {
				return nil, ap.Error(ErrElseNoArgs.Error())
			}
			stmts, nextTag, nextAP, err := doc.ParseUntilWithArgs("endif")
			if err != nil {
				return nil, err
			}
			elseBody = convertStatementsToNodes(stmts)
			tag = nextTag
			ap = nextAP
		}
	}

	if tag != "endif" {
		return nil, doc.Errorf("expected endif, got %s", tag)
	}
	if ap.Remaining() > 0 {
		return nil, ap.Error(ErrEndifNoArgs.Error())
	}

	return &IfNode{
		Branches: branches,
		ElseBody: elseBody,
		Line:     start.Line,
		Col:      start.Col,
	}, nil
}
