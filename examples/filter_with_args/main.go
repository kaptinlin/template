// Package main demonstrates template filters with arguments.
package main

import (
	"fmt"

	"github.com/kaptinlin/template"
)

func main() {
	// Define a template string using the truncate filter with arguments
	source := `Here's a short summary: "{{ summary | truncate:10 }}"`

	// Parse the template string
	tpl, err := template.Parse(source)

	if err != nil {
		panic(err) // Handle errors appropriately
	}

	// Create a context and add data
	context := template.NewContext()
	context.Set("summary", "This is a long summary text that needs truncating.")

	// Execute the template with the provided context
	output, err := template.Execute(tpl, context)
	if err != nil {
		panic(err) // Handle errors appropriately
	}

	fmt.Println(output) // Output: Here's a short summary: "This is a..."
}
