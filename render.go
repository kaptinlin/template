package template

// Render is a convenience function that compiles and renders a template in one step.
//
// This is the simplest way to use the template engine - just provide the template source
// and data, and get the rendered output.
//
// Example:
//
//	output, err := template.Render("Hello {{ name }}!", map[string]interface{}{
//	    "name": "World",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(output) // Output: Hello World!
func Render(source string, data map[string]interface{}) (string, error) {
	tmpl, err := Compile(source)
	if err != nil {
		return "", err
	}
	return tmpl.Render(data)
}
