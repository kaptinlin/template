// Package main demonstrates typical template usage.
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kaptinlin/template"
)

func main() {
	runMain(os.Stdout, log.Fatal)
}

func runMain(out io.Writer, fatal func(...any)) {
	if err := run(out); err != nil {
		fatal(err)
	}
}

func run(out io.Writer) error {
	source := `Hello, {{ name|upper }}!
{% if items %}Items:
{% for i, item in items %}  {{ i }}: {{ item }}
{% endfor %}{% else %}No items.
{% endif %}`

	return render(out, source, template.Data{
		"name":  "alice",
		"items": []string{"foo", "bar", "baz"},
	})
	// Hello, ALICE!
	// Items:
	//   0: foo
	//   1: bar
	//   2: baz
}

func render(out io.Writer, source string, data template.Data) error {
	engine := template.New()

	tpl, err := engine.ParseString(source)
	if err != nil {
		return err
	}

	rendered, err := tpl.Render(data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(out, rendered)
	return err
}
