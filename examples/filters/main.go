package main

import (
	"fmt"

	"github.com/kaptinlin/template" // Import the template package
)

func main() {
	// Define a template with filters applied to variables
	source := `Hello, {{ user.firstName | capitalize }} {{ user.lastName | capitalize }}! 
You have {{ unreadMessages | pluralize:"%d message","%d messages" }}.`

	// Parse the template string
	tpl, err := template.Parse(source)
	if err != nil {
		// In real applications, handle errors appropriately
		panic(err)
	}

	// Create a context and add complex structured data
	context := template.NewContext()
	user := map[string]interface{}{
		"firstName": "john",
		"lastName":  "doe",
	}
	context.Set("user", user)
	context.Set("unreadMessages", 1) // Changed to 1 to demonstrate pluralize filter

	// Execute the template with the provided context
	output, err := template.Execute(tpl, context)
	if err != nil {
		// In real applications, handle errors appropriately
		panic(err)
	}

	// Print the result of the template execution
	fmt.Println(output) // Expected Output: Hello, John Doe! You have 1 message.
}
