/*
Package template provides a Go template engine built around a single
core concept: [Engine].

Use [New] with [WithLoader], [WithFormat], [WithLayout], and [WithFeatures] to define
how templates are loaded, what output semantics they use, and which
optional language features are enabled.

	loader, _ := template.NewDirLoader("./templates")
	engine := template.New(
		template.WithLoader(loader),
		template.WithFormat(template.FormatHTML),
		template.WithLayout(),
		template.WithDefaults(template.Data{"site": siteData}),
	)
	_ = engine.RenderTo("layouts/blog.html", os.Stdout, template.Data{"page": pageData})

Core design rules:

  - [Engine] is the only public entry point.
  - [FormatHTML] and [FormatText] define output semantics.
  - [WithLayout] is optional and must be enabled explicitly.
  - HTML auto-escape exists only behind [FormatHTML].
  - Layout inheritance, include, raw, and safe-aware behavior exist only
    behind [WithLayout] (internally this maps to [FeatureLayout]).

Supported syntax summary:

  - Variable interpolation: {{ variable }}
  - Filters: {{ variable | filter:arg }}
  - Control structures: {% if %}, {% for %}
  - Comments: {# ... #}
  - (FeatureLayout only) {% include "x" [with k=v] [only] [if_exists] %}
  - (FeatureLayout only) {% extends "parent" %} + {% block name %}...{% endblock %}
  - (FeatureLayout only) {{ block.super }}
  - (FeatureLayout only) {% raw %}...{% endraw %}
  - (FeatureLayout only) {{ x | safe }} and ([FormatHTML] only) auto-escape of {{ x }}

For detailed examples, see the examples/ directory.
*/
package template
