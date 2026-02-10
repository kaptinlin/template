package template

// parseContinueTag parses the continue tag.
// {% continue %}
func parseContinueTag(_ *Parser, start *Token, arguments *Parser) (Statement, error) {
	if arguments.Remaining() > 0 {
		return nil, arguments.Error(ErrContinueNoArgs.Error())
	}

	return &ContinueNode{
		Line: start.Line,
		Col:  start.Col,
	}, nil
}
