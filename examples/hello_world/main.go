package main

import (
	"fmt"

	"github.com/kaptinlin/template"
)

func main() {
	// Define a simple template
	source := "Hello, {{ name }}!"

	// Parse the template string
	tpl, err := template.Parse(source)
	if err != nil {
		// Handle the error appropriately in real applications
		panic(err)
	}

	// Create a context and add a variable named 'name'
	context := template.NewContext()
	context.Set("name", "World")

	// Execute the template with the provided context
	output, err := template.Execute(tpl, context)
	if err != nil {
		// Handle the error appropriately in real applications
		panic(err)
	}

	// Print the result of the template execution
	fmt.Println(output) // Output: Hello, World!
}
