// Package main demonstrates multi-file text generation with Engine.
//
// This is the "code / config generator" use case: the output is YAML
// meant for another template engine (Task runner), so HTML-escape would
// corrupt it ({{ becomes &#123;&#123;). FormatText never escapes.
//
// The scaffold directory contains a Taskfile.yml.tmpl that uses
// {% include %} to pull in header.tmpl and footer.tmpl, and emits
// literal Task-runner variables using string-literal interpolation
// (e.g. {{ "{{.GOBIN}}" }}).
//
// Run:
//
//	go run ./examples/multifile_text
package main

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"os"

	"github.com/kaptinlin/template"
)

//go:embed scaffold/*
var scaffoldFS embed.FS

func main() {
	runMain(os.Stdout, log.Fatal)
}

func runMain(out io.Writer, fatal func(...any)) {
	if err := run(out); err != nil {
		fatal(err)
	}
}

func run(out io.Writer) error {
	// Strip the "scaffold/" prefix so includes can use short names.
	rooted, err := fs.Sub(scaffoldFS, "scaffold")
	if err != nil {
		return err
	}
	loader := template.NewFSLoader(rooted)

	// Engine + FormatText — no HTML auto-escape. The generated Taskfile.yml
	// contains literal '{{.GOBIN}}' sequences meant for Task runner,
	// and must be emitted byte-for-byte.
	engine := template.New(
		template.WithLoader(loader),
		template.WithFormat(template.FormatText),
		template.WithLayout(),
		template.WithDefaults(template.Data{
			"project": map[string]any{
				"name":   "awesome-service",
				"binary": "awesomectl",
				"lint":   true,
			},
		}),
	)

	return engine.RenderTo("Taskfile.yml.tmpl", out, nil)
}
