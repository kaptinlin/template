// Package main demonstrates multi-file HTML templating with layout
// inheritance, includes, block.super, and HTML auto-escape.
//
// Run:
//
//	go run ./examples/layout
package main

import (
	"embed"
	"io/fs"
	"log"
	"os"

	"github.com/kaptinlin/template"
)

//go:embed templates/*
var templateFS embed.FS

func main() {
	// Strip the "templates/" prefix so templates can reference
	// "layouts/base.html" instead of "templates/layouts/base.html".
	rooted, err := fs.Sub(templateFS, "templates")
	if err != nil {
		log.Fatal(err)
	}

	// Build a loader chain. In a real project you would also include a
	// writable user-overrides directory:
	//
	//   user,   _ := template.NewDirLoader("./user-templates")
	//   loader    := template.NewChainLoader(user, template.NewFSLoader(rooted))
	//
	// Here we just use the embedded FS for simplicity.
	loader := template.NewFSLoader(rooted)

	// HTMLSet enables auto-escape for every {{ expr }} output.
	set := template.NewHTMLSet(loader,
		template.WithGlobals(template.Context{
			"site": map[string]any{
				"title": "Example Blog",
				"url":   "https://example.com",
			},
		}),
	)

	// Render a blog post. Values flow through auto-escape except where
	// explicitly marked with SafeString or the | safe filter.
	err = set.Render("layouts/blog.html", template.Context{
		"page": map[string]any{
			"title":  "Hello <world> & friends",
			"author": "Alice",
			"date":   "2026-04-08",
			// SafeString marks this as pre-rendered HTML so it is not
			// escaped. In a real app, this would be the output of a
			// Markdown-to-HTML pipeline.
			"content": template.SafeString(`
    <p>This is the <strong>trusted</strong> body,
    already rendered to HTML by an upstream Markdown pipeline.</p>`),
			"tags": []string{"golang", "templates", "<example>"},
		},
	}, os.Stdout)

	if err != nil {
		log.Fatal(err)
	}
}
