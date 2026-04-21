package template

func parseBreakTag(_ *Parser, start *Token, args *Parser) (Statement, error) {
	if args.Remaining() > 0 {
		return nil, args.Error(ErrBreakNoArgs.Error())
	}
	return &BreakNode{Line: start.Line, Col: start.Col}, nil
}
