// Package main demonstrates typical template usage.
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/template"
)

func main() {
	source := `Hello, {{ name|upper }}!
{% if items %}Items:
{% for i, item in items %}  {{ i }}: {{ item }}
{% endfor %}{% else %}No items.
{% endif %}`

	output, err := template.Render(source, map[string]interface{}{
		"name":  "alice",
		"items": []string{"foo", "bar", "baz"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(output)
	// Hello, ALICE!
	// Items:
	//   0: foo
	//   1: bar
	//   2: baz
}
