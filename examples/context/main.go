// Package main demonstrates template context usage.
package main

import (
	"fmt"

	"github.com/kaptinlin/template"
)

func main() {
	context := template.NewContext()

	userInfo := map[string]interface{}{
		"firstName": "John",
		"lastName":  "Doe",
	}
	context.Set("user", userInfo)

	firstName, err := context.Get("user.firstName")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("User's first name: %v\n", firstName)
	}

	lastName, err := context.Get("user.lastName")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("User's last name: %v\n", lastName)
	}
}
