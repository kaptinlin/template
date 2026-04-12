package template

func parseSourceTemplate(source string) (*Template, error) {
	return New().ParseString(source)
}

func renderSourceTemplate(source string, data Data) (string, error) {
	tpl, err := parseSourceTemplate(source)
	if err != nil {
		return "", err
	}
	return tpl.Render(data)
}
