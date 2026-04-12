// Package main demonstrates typical template usage.
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/template"
)

func main() {
	engine := template.New()

	source := `Hello, {{ name|upper }}!
{% if items %}Items:
{% for i, item in items %}  {{ i }}: {{ item }}
{% endfor %}{% else %}No items.
{% endif %}`

	tpl, err := engine.ParseString(source)
	if err != nil {
		log.Fatal(err)
	}

	rendered, err := tpl.Render(template.Data{
		"name":  "alice",
		"items": []string{"foo", "bar", "baz"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(rendered)
	// Hello, ALICE!
	// Items:
	//   0: foo
	//   1: bar
	//   2: baz
}
