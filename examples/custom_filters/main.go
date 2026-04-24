// Package main demonstrates registering custom filters.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

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
	engine := template.New()

	// Register a "repeat" filter: {{ text|repeat:3 }} → "texttexttext"
	engine.RegisterFilter("repeat", func(value any, args ...any) (any, error) {
		s := fmt.Sprintf("%v", value)
		n := 2
		if len(args) > 0 {
			if parsed, err := strconv.Atoi(fmt.Sprintf("%v", args[0])); err == nil {
				n = parsed
			}
		}
		return strings.Repeat(s, n), nil
	})

	tpl, err := engine.ParseString(`{{ word|repeat:3 }}`)
	if err != nil {
		return err
	}

	rendered, err := tpl.Render(template.Data{"word": "ha"})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, rendered) // hahaha
	return err
}
