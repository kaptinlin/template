package template

// parseIfTag parses an if-elif-else-endif block into an ifNode.
//
// Syntax:
//
//	{% if condition %}...{% elif condition %}...{% else %}...{% endif %}
func parseIfTag(doc *parser, start *token, args *parser) (statement, error) {
	var branches []ifBranch
	var elseBody []node

	cond, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	if args.Remaining() > 0 {
		return nil, args.Error(errUnexpectedTokensAfterCondition.Error())
	}

	body, tag, ap, err := doc.ParseUntilWithArgs("elif", "else", "endif")
	if err != nil {
		return nil, err
	}
	branches = append(branches, ifBranch{
		Condition: cond,
		Body:      convertStatementsToNodes(body),
	})

	for tag == "elif" || tag == "else" {
		switch tag {
		case "elif":
			cond, err = ap.ParseExpression()
			if err != nil {
				return nil, err
			}
			if ap.Remaining() > 0 {
				return nil, ap.Error(errUnexpectedTokensAfterCondition.Error())
			}
			next, nextTag, nextAP, err := doc.ParseUntilWithArgs("elif", "else", "endif")
			if err != nil {
				return nil, err
			}
			branches = append(branches, ifBranch{
				Condition: cond,
				Body:      convertStatementsToNodes(next),
			})
			tag = nextTag
			ap = nextAP

		case "else":
			if ap.Remaining() > 0 {
				return nil, ap.Error(errElseNoArgs.Error())
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
		return nil, ap.Error(errEndifNoArgs.Error())
	}

	return &ifNode{
		Branches: branches,
		ElseBody: elseBody,
		Line:     start.Line,
		Col:      start.Col,
	}, nil
}
