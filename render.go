package template

// Render compiles and renders a template in one step.
//
// This is a shorthand for calling [Compile] followed by [Template.Render].
// For repeated rendering of the same template, compile once with [Compile]
// and call [Template.Render] to avoid redundant compilation.
func Render(source string, data map[string]interface{}) (string, error) {
	tmpl, err := Compile(source)
	if err != nil {
		return "", err
	}
	return tmpl.Render(data)
}
