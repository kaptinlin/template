/*
Package template provides a simple and efficient template engine for Go.

Single-string usage (backwards-compatible, minimal feature set):

	tmpl, err := template.Compile("Hello, {{ name|upper }}!")
	if err != nil {
		panic(err)
	}

	output, err := tmpl.Render(template.Context{"name": "world"})
	// Output: "Hello, WORLD!"

Multi-file usage (layout, inheritance, includes, HTML auto-escape):

	loader, _ := template.NewDirLoader("./templates")
	set := template.NewHTMLSet(loader,
		template.WithGlobals(template.Context{"site": siteData}),
	)
	_ = set.Render("layouts/blog.html", template.Context{"page": pageData}, os.Stdout)

Two worlds, one package:

  - [Compile] / [Render] — simple string templates. Supports {{ var }},
    filters, {% if %}, {% for %}, {% break %}, {% continue %}, {# ... #}.
    Does NOT support {% include %}, {% extends %}, {% block %}, {% raw %},
    or the "safe" filter. Does NOT auto-escape. Suited to log lines,
    config snippets, plain-text emails, and anything that is "interpolate
    a few values into a string".

  - [NewTextSet] / [NewHTMLSet] — multi-file template systems with
    Loader-backed inclusion, inheritance, block.super, raw blocks, and
    (HTMLSet only) automatic HTML escaping with the [SafeString]
    mechanism. Layout tags and the safe filter live in a per-Set
    registry so the Compile(src) path above is unaffected.

This mirrors Go's own text/template vs html/template split.

Supported syntax summary:

  - Variable interpolation: {{ variable }}
  - Filters: {{ variable | filter:arg }}
  - Control structures: {% if %}, {% for %}
  - Comments: {# ... #}
  - (Set only) {% include "x" [with k=v] [only] [if_exists] %}
  - (Set only) {% extends "parent" %} + {% block name %}...{% endblock %}
  - (Set only) {{ block.super }}
  - (Set only) {% raw %}...{% endraw %}
  - (Set only) {{ x | safe }} and (HTMLSet only) auto-escape of {{ x }}

Architecture:

The package is organized into several key components:
  - Lexer: Tokenizes template source into a token stream
  - Parser: Converts tokens into an AST (Abstract Syntax Tree)
  - Expression Parser: Parses expressions with operator precedence
  - Template: Executes the AST with a given context
  - Filters: Transforms values during template execution
  - Context: Stores and retrieves template variables

Control Flow:

The template engine supports break and continue statements within loops:

	{% for item in items %}
		{% if item == "skip" %}
			{% continue %}
		{% endif %}
		{% if item == "stop" %}
			{% break %}
		{% endif %}
		{{ item }}
	{% endfor %}

Loop Context:

Within loops, a special "loop" variable is available providing:
  - loop.Index: Current index (0-based)
  - loop.Revindex: Reverse index
  - loop.First: True if first iteration
  - loop.Last: True if last iteration
  - loop.Length: Total collection length

For detailed examples, see the examples/ directory.
*/
package template
