package template

// parseBreakTag parses the break tag.
// {% break %}
func parseBreakTag(_ *Parser, start *Token, arguments *Parser) (Statement, error) {
	if arguments.Remaining() > 0 {
		return nil, arguments.Error(ErrBreakNoArgs.Error())
	}

	return &BreakNode{
		Line: start.Line,
		Col:  start.Col,
	}, nil
}
