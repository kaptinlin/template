package template

// parseBreakTag parses a break statement into a BreakNode.
// The break tag accepts no arguments: {% break %}
func parseBreakTag(_ *Parser, start *Token, args *Parser) (Statement, error) {
	if args.Remaining() > 0 {
		return nil, args.Error(ErrBreakNoArgs.Error())
	}
	return &BreakNode{Line: start.Line, Col: start.Col}, nil
}
