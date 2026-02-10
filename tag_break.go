package template

// parseBreakTag parses the break tag.
// {% break %}
func parseBreakTag(_ *Parser, start *Token, arguments *Parser) (Statement, error) {
	// break does not accept arguments.
	if arguments.Remaining() > 0 {
		return nil, arguments.Error("break does not take arguments")
	}

	return &BreakNode{
		Line: start.Line,
		Col:  start.Col,
	}, nil
}
