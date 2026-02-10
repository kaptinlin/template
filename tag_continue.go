package template

// parseContinueTag parses the continue tag.
// {% continue %}
func parseContinueTag(_ *Parser, start *Token, arguments *Parser) (Statement, error) {
	// continue does not accept arguments.
	if arguments.Remaining() > 0 {
		return nil, arguments.Error("continue does not take arguments")
	}

	return &ContinueNode{
		Line: start.Line,
		Col:  start.Col,
	}, nil
}
