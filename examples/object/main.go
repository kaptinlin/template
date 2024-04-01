package main

import (
	"fmt"

	"github.com/kaptinlin/template" // Import the template package
)

func main() {
	// Define a template with a complex variable structure
	source := "Hello, {{ user.firstName }} {{ user.lastName }}! You have {{ unreadMessages }} new messages."

	// Parse the template string
	tpl, err := template.Parse(source)
	if err != nil {
		// In real applications, handle errors appropriately
		panic(err)
	}

	// Create a context and add complex structured data
	context := template.NewContext()
	user := map[string]interface{}{
		"firstName": "John",
		"lastName":  "Doe",
	}
	context.Set("user", user)
	context.Set("unreadMessages", 5)

	// Execute the template with the provided context
	output, err := template.Execute(tpl, context)
	if err != nil {
		// In real applications, handle errors appropriately
		panic(err)
	}

	// Print the result of the template execution
	fmt.Println(output) // Output: Hello, John Doe! You have 5 new messages.
}
