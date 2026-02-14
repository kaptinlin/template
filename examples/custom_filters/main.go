// Package main demonstrates registering custom filters.
package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/kaptinlin/template"
)

func main() {
	// Register a "repeat" filter: {{ text|repeat:3 }} → "texttexttext"
	template.RegisterFilter("repeat", func(value any, args ...string) (any, error) {
		s := fmt.Sprintf("%v", value)
		n := 2
		if len(args) > 0 {
			if parsed, err := strconv.Atoi(args[0]); err == nil {
				n = parsed
			}
		}
		return strings.Repeat(s, n), nil
	})

	output, err := template.Render(`{{ word|repeat:3 }}`, map[string]any{
		"word": "ha",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output) // hahaha
}
