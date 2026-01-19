/*
Package template provides a simple and efficient template engine for Go.

The template engine supports:
  - Variable interpolation: {{ variable }}
  - Filters: {{ variable|filter:arg }}
  - Control structures: {% if condition %}, {% for item in collection %}
  - Comments: {# comment #}

Basic Usage:

	source := "Hello, {{ name|upper }}!"
	parser := template.NewParser()
	tpl, err := parser.Parse(source)
	if err != nil {
		panic(err)
	}

	ctx := template.NewContext()
	ctx.Set("name", "world")

	output, err := tpl.Execute(ctx)
	// Output: "Hello, WORLD!"

Architecture:

The package is organized into several key components:
  - Parser: Converts template strings into an AST (Abstract Syntax Tree)
  - Grammar: Parses and evaluates conditional expressions
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
