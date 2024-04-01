package template

// Parse parses a template string and returns a Template instance.
func Parse(source string) (*Template, error) {
	parser := NewParser()
	return parser.Parse(source)
}

// Execute renders the template with provided context.
func Execute(tpl *Template, ctx Context) (string, error) {
	return tpl.Execute(ctx)
}

// MustExecute renders the template with provided context, ignoring errors.
func MustExecute(tpl *Template, ctx Context) string {
	return tpl.MustExecute(ctx)
}

// Render combines parsing and executing a template with the given context for convenience.
func Render(source string, ctx Context) (string, error) {
	tpl, err := Parse(source)
	if err != nil {
		return "", err
	}
	return Execute(tpl, ctx)
}
