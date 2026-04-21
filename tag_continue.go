package template

func parseContinueTag(_ *Parser, start *Token, args *Parser) (Statement, error) {
	if args.Remaining() > 0 {
		return nil, args.Error(ErrContinueNoArgs.Error())
	}
	return &ContinueNode{Line: start.Line, Col: start.Col}, nil
}
