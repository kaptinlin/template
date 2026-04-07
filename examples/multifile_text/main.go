// Package main demonstrates multi-file text generation with NewTextSet.
//
// This is the "code / config generator" use case: the output is YAML
// meant for another template engine (Task runner), so HTML-escape would
// corrupt it ({{ becomes &#123;&#123;). NewTextSet never escapes.
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
	"io/fs"
	"log"
	"os"

	"github.com/kaptinlin/template"
)

//go:embed scaffold/*
var scaffoldFS embed.FS

func main() {
	// Strip the "scaffold/" prefix so includes can use short names.
	rooted, err := fs.Sub(scaffoldFS, "scaffold")
	if err != nil {
		log.Fatal(err)
	}
	loader := template.NewFSLoader(rooted)

	// NewTextSet — no HTML auto-escape. The generated Taskfile.yml
	// contains literal '{{.GOBIN}}' sequences meant for Task runner,
	// and must be emitted byte-for-byte.
	set := template.NewTextSet(loader,
		template.WithGlobals(template.Context{
			"project": map[string]any{
				"name":   "awesome-service",
				"binary": "awesomectl",
				"lint":   true,
			},
		}),
	)

	if err := set.Render("Taskfile.yml.tmpl", nil, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
