package template

func parseBreakTag(_ *parser, start *token, args *parser) (statement, error) {
	if args.Remaining() > 0 {
		return nil, args.Error(errBreakNoArgs.Error())
	}
	return &breakNode{Line: start.Line, Col: start.Col}, nil
}
