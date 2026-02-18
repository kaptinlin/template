/*
Package template provides a simple and efficient template engine for Go.

The template engine supports:
  - Variable interpolation: {{ variable }}
  - Filters: {{ variable|filter:arg }}
  - Control structures: {% if condition %}, {% for item in collection %}
  - Comments: {# comment #}

Basic Usage:

	tmpl, err := template.Compile("Hello, {{ name|upper }}!")
	if err != nil {
		panic(err)
	}

	output, err := tmpl.Render(map[string]any{"name": "world"})
	// Output: "Hello, WORLD!"

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
