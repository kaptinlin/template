// Package main demonstrates registering custom filters.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/kaptinlin/template"
)

func main() {
	// Register a "repeat" filter: {{ text|repeat:3 }} → "texttexttext"
	template.RegisterFilter("repeat", func(value interface{}, args ...string) (interface{}, error) {
		s := fmt.Sprintf("%v", value)
		n := 2
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &n)
		}
		return strings.Repeat(s, n), nil
	})

	output, err := template.Render(`{{ word|repeat:3 }}`, map[string]interface{}{
		"word": "ha",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output) // hahaha
}
