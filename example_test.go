package template_test

import (
	"fmt"
	"log"

	"github.com/kaptinlin/template"
)

func ExampleEngine_ParseString() {
	engine := template.New()
	tmpl, err := engine.ParseString("Hello, {{ name|upcase }}!")
	if err != nil {
		log.Fatal(err)
	}

	out, err := tmpl.Render(template.Data{"name": "alice"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
	// Output:
	// Hello, ALICE!
}

func ExampleEngine_Render() {
	loader := template.NewMemoryLoader(map[string]string{
		"base.html": `<h1>{{ page.title }}</h1>{% block content %}{% endblock %}`,
		"page.html": `{% extends "base.html" %}{% block content %}{{ page.content }} {{ page.safe | safe }}{% endblock %}`,
	})

	engine := template.New(
		template.WithLoader(loader),
		template.WithFormat(template.FormatHTML),
		template.WithLayout(),
	)

	out, err := engine.Render("page.html", template.Data{
		"page": map[string]any{
			"title":   "Hello <world>",
			"content": "<p>escaped</p>",
			"safe":    template.SafeString("<p>trusted</p>"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
	// Output:
	// <h1>Hello &lt;world&gt;</h1>&lt;p&gt;escaped&lt;/p&gt; <p>trusted</p>
}
