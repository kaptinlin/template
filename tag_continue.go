package template

func parseContinueTag(_ *parser, start *token, args *parser) (statement, error) {
	if args.Remaining() > 0 {
		return nil, args.Error(errContinueNoArgs.Error())
	}
	return &continueNode{Line: start.Line, Col: start.Col}, nil
}
