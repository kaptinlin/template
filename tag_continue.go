package template

// parseContinueTag parses a continue statement into a ContinueNode.
// The continue tag accepts no arguments: {% continue %}
func parseContinueTag(_ *Parser, start *Token, args *Parser) (Statement, error) {
	if args.Remaining() > 0 {
		return nil, args.Error(ErrContinueNoArgs.Error())
	}
	return &ContinueNode{Line: start.Line, Col: start.Col}, nil
}
