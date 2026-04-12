# Template Engine

A lightweight Go template engine with Django-inspired control flow and Liquid-compatible filter syntax. Supports variable interpolation, 80+ built-in filters, conditionals, loops, and — when you opt in — loader-backed templates with inheritance, includes, raw blocks, and HTML auto-escape.

This engine implements the [Liquid](https://shopify.github.io/liquid/) filter standard with 41 out of 46 filters fully compliant, plus 20+ extension filters. See [Liquid Compatibility](docs/liquid-compatibility.md) for details.

## Engine-First API

This package now exposes one public model: `Engine`.

| If you need to... | Use | Why |
|---|---|---|
| Render plain text, YAML, TOML, code, or Taskfiles from files | `template.New(template.WithLoader(...), template.WithFormat(template.FormatText))` | Unified engine API with text output semantics |
| Render HTML pages or HTML emails | `template.New(template.WithLoader(...), template.WithFormat(template.FormatHTML))` | Unified engine API with HTML output semantics |
| Enable layout features | `template.WithFeatures(template.FeatureLayout)` | Explicitly turns on `include` / `extends` / `block` / `raw` |
| Parse a one-off source string | `engine.ParseString(src)` | Same engine model, no second API surface |

Recommended defaults:

- Local filesystem templates: `NewDirLoader(...)`
- Embedded templates: `NewFSLoader(embed.FS)`
- HTML output: `New(..., WithFormat(FormatHTML))`
- Non-HTML output: `New(..., WithFormat(FormatText))`
- Layout features: `WithFeatures(FeatureLayout)` only when needed

## Design Rules

The public design is intentionally narrow:

```go
engine := template.New(
    template.WithLoader(loader),
    template.WithFormat(template.FormatHTML),
    template.WithFeatures(template.FeatureLayout),
)
```

1. `Engine` is the only entry point.
2. `FormatHTML` and `FormatText` define output semantics.
3. `FeatureLayout` is optional and must be enabled explicitly.
4. Tags and filters belong to an engine instance, not to mutable package-global APIs.

## Installation

```sh
go get github.com/kaptinlin/template
```

## Quick Start

### Parse and render a source string

```go
engine := template.New()
tmpl, err := engine.ParseString("Hello, {{ name|upcase }}!")
if err != nil {
    log.Fatal(err)
}

output, err := tmpl.Render(template.Context{"name": "alice"})
// output: "Hello, ALICE!"
```

### Multi-file HTML with layout inheritance

For web pages with shared layouts, includes, and auto-escape:

```go
import (
    "embed"
    "os"

    "github.com/kaptinlin/template"
)

//go:embed themes/default/*
var themeFS embed.FS

func main() {
    user, _ := template.NewDirLoader("./templates")               // os.Root sandbox
    theme  := template.NewFSLoader(themeFS)                        // embed.FS
    loader := template.NewChainLoader(user, theme)                 // user > theme

    engine := template.New(
        template.WithLoader(loader),
        template.WithFormat(template.FormatHTML),
        template.WithFeatures(template.FeatureLayout),
        template.WithGlobals(template.Context{"site": siteData}),
    )

    _ = engine.Render("layouts/blog.html", template.Context{
        "page": map[string]any{
            "title":   "Hello <world>",                            // auto-escaped
            "content": template.SafeString("<p>trusted HTML</p>"), // rendered as-is
        },
    }, os.Stdout)
}
```

The templates in `./themes/default/`:

```django
{# layouts/base.html #}
<!DOCTYPE html>
<html>
<head>{% block head %}<title>{{ site.title }}</title>{% endblock %}</head>
<body>
  {% include "partials/header.html" %}
  <main>{% block content %}{% endblock %}</main>
</body>
</html>
```

```django
{# layouts/blog.html #}
{% extends "layouts/base.html" %}

{% block head %}
  {{ block.super }}
  <meta name="description" content="{{ page.title }}">
{% endblock %}

{% block content %}
  <article>
    <h1>{{ page.title }}</h1>
    {{ page.content | safe }}
  </article>
{% endblock content %}
```

See [docs/layout.md](docs/layout.md) for the full layout guide, [docs/loaders.md](docs/loaders.md) for loader configuration, and [docs/security.md](docs/security.md) for the security model.

### Multi-file text generation

For generating code, config files, or anything that should NOT be HTML-escaped:

```go
loader, _ := template.NewDirLoader("./scaffold")

engine := template.New(
    template.WithLoader(loader),
    template.WithFormat(template.FormatText),
    template.WithFeatures(template.FeatureLayout),
    template.WithGlobals(template.Context{"project": projectMeta}),
)

// Generates Taskfile.yml with {{.GOBIN}}/bin intact — no HTML escape would
// turn & into &amp; and break the YAML.
_ = engine.Render("Taskfile.yml.tmpl", nil, &buf)
```

### Using io.Writer

```go
engine := template.New()
tmpl, _ := engine.ParseString("Hello, {{ name }}!")
ctx := template.NewExecutionContext(template.Context{"name": "World"})
_ = tmpl.Execute(ctx, os.Stdout)
```

## Template Syntax

### Variables

Use `{{ }}` to output variables, with dot notation for nested properties:

```
{{ user.name }}
{{ user.address.city }}
{{ items.0 }}
```

See [Variables Documentation](docs/variables.md) for details.

### Filters

Use pipe `|` to apply filters to variables, supporting chaining and arguments:

```
{{ name|upcase }}
{{ title|truncate:20 }}
{{ name|downcase|capitalize }}
{{ price|plus:10|times:2 }}
```

See [Filters Documentation](docs/filters.md) for details.

### Conditionals

```
{% if score > 80 %}
    Excellent
{% elif score > 60 %}
    Pass
{% else %}
    Fail
{% endif %}
```

### Loops

```
{% for item in items %}
    {{ item }}
{% endfor %}

{% for key, value in dict %}
    {{ key }}: {{ value }}
{% endfor %}
```

The `loop` variable is available inside loops:

| Property | Description |
|----------|-------------|
| `loop.Index` | Current index (starting from 0) |
| `loop.Revindex` | Reverse index |
| `loop.First` | Whether this is the first iteration |
| `loop.Last` | Whether this is the last iteration |
| `loop.Length` | Total length of the collection |

Supports `{% break %}` and `{% continue %}` for loop control.

Loop bindings are scoped to the loop body. After `{% endfor %}`, `loop`
and the loop variables are restored to their pre-loop values, or removed
if they did not exist before the loop.

See [Control Structure Documentation](docs/control-structure.md) for details.

### Comments

```
{# This content will not appear in the output #}
```

### Layout & inheritance (`FeatureLayout`)

Available when the engine enables `template.WithFeatures(template.FeatureLayout)`:

```django
{# Include a partial #}
{% include "partials/header.html" %}
{% include "partials/card.html" with title="Hi" count=3 %}      {# pass variables #}
{% include "partials/card.html" with title="Hi" only %}         {# isolate context #}
{% include "partials/optional.html" if_exists %}                {# missing is no-op #}
{% include page.widget %}                                        {# dynamic path #}

{# Inherit from a parent template #}
{% extends "layouts/base.html" %}

{# Define/override blocks #}
{% block content %}
  {{ block.super }}             {# call the parent block's output #}
  <p>My override</p>
{% endblock content %}

{# Output literal template syntax, no interpolation #}
{% raw %}
  Literal: {{ not_a_variable }}  {% for x in items %}
{% endraw %}
```

`{% include %}`, `{% extends %}`, `{% block %}`, and `{% raw %}` exist only when an engine enables `FeatureLayout`. See [docs/layout.md](docs/layout.md).

### Auto-escape & `safe` (`FormatHTML` only)

An engine using `FormatHTML` auto-escapes every `{{ expr }}` output:

```django
{{ page.title }}                 {# auto-escaped: < becomes &lt; #}
{{ page.content | safe }}        {# marked trusted — rendered as-is #}
{{ user_input | escape }}        {# explicit escape #}
```

Go code can mark values as trusted using `template.SafeString`:

```go
engine.Render("page.html", template.Context{
    "title":   "Hello <world>",                            // escaped
    "content": template.SafeString("<p>pre-rendered</p>"), // NOT escaped
}, w)
```

Engines using `FormatText` do not escape anything — they treat `SafeString` as a plain string.

### Expressions

Supports the following operators (from lowest to highest precedence):

| Operator | Description |
|----------|-------------|
| `or`, `\|\|` | Logical OR |
| `and`, `&&` | Logical AND |
| `==`, `!=`, `<`, `>`, `<=`, `>=` | Comparison |
| `+`, `-` | Addition, Subtraction |
| `*`, `/`, `%` | Multiplication, Division, Modulo |
| `not`, `-`, `+` | Unary operators |

Literal support: strings (`"text"` / `'text'`), numbers (`42`, `3.14`), booleans (`true` / `false`), null (`null`).

## Built-in Filters

### String

| Filter | Description | Example |
|--------|-------------|---------|
| `default` | Return default value if empty | `{{ name\|default:'Anonymous' }}` |
| `upcase` | Convert to uppercase | `{{ name\|upcase }}` |
| `downcase` | Convert to lowercase | `{{ name\|downcase }}` |
| `capitalize` | Capitalize first letter | `{{ name\|capitalize }}` |
| `strip` | Remove leading/trailing whitespace | `{{ text\|strip }}` |
| `lstrip` | Remove leading whitespace | `{{ text\|lstrip }}` |
| `rstrip` | Remove trailing whitespace | `{{ text\|rstrip }}` |
| `truncate` | Truncate to length (default 50) | `{{ text\|truncate:20 }}` |
| `truncatewords` | Truncate to word count (default 15) | `{{ text\|truncatewords:5 }}` |
| `replace` | Replace all occurrences | `{{ text\|replace:'old','new' }}` |
| `replace_first` | Replace first occurrence | `{{ text\|replace_first:'old','new' }}` |
| `replace_last` | Replace last occurrence | `{{ text\|replace_last:'old','new' }}` |
| `remove` | Remove all occurrences | `{{ text\|remove:'bad' }}` |
| `remove_first` | Remove first occurrence | `{{ text\|remove_first:'x' }}` |
| `remove_last` | Remove last occurrence | `{{ text\|remove_last:'x' }}` |
| `append` | Append string | `{{ name\|append:'!' }}` |
| `prepend` | Prepend string | `{{ name\|prepend:'Hi ' }}` |
| `split` | Split by delimiter | `{{ csv\|split:',' }}` |
| `slice` | Extract substring by offset and length | `{{ text\|slice:1,3 }}` |
| `escape` | Escape HTML characters | `{{ html\|escape }}` |
| `escape_once` | Escape without double-escaping | `{{ html\|escape_once }}` |
| `strip_html` | Remove HTML tags | `{{ html\|strip_html }}` |
| `strip_newlines` | Remove newline characters | `{{ text\|strip_newlines }}` |
| `url_encode` | Percent-encode for URLs | `{{ text\|url_encode }}` |
| `url_decode` | Decode percent-encoded string | `{{ text\|url_decode }}` |
| `base64_encode` | Encode to Base64 | `{{ text\|base64_encode }}` |
| `base64_decode` | Decode from Base64 | `{{ text\|base64_decode }}` |

**String extensions** (not in Liquid standard):

| Filter | Description | Example |
|--------|-------------|---------|
| `titleize` | Capitalize first letter of each word | `{{ title\|titleize }}` |
| `camelize` | Convert to camelCase | `{{ name\|camelize }}` |
| `pascalize` | Convert to PascalCase | `{{ name\|pascalize }}` |
| `dasherize` | Convert to dash-separated | `{{ name\|dasherize }}` |
| `slugify` | Convert to URL-friendly format | `{{ title\|slugify }}` |
| `pluralize` | Singular/plural selection | `{{ count\|pluralize:'item','items' }}` |
| `ordinalize` | Convert to ordinal | `{{ num\|ordinalize }}` |
| `length` | Get string/array/map length | `{{ name\|length }}` |

### Math

| Filter | Description | Example |
|--------|-------------|---------|
| `plus` | Addition | `{{ price\|plus:10 }}` |
| `minus` | Subtraction | `{{ price\|minus:5 }}` |
| `times` | Multiplication | `{{ price\|times:2 }}` |
| `divided_by` | Division | `{{ total\|divided_by:3 }}` |
| `modulo` | Modulo | `{{ num\|modulo:2 }}` |
| `abs` | Absolute value | `{{ num\|abs }}` |
| `round` | Round (default precision 0) | `{{ pi\|round:2 }}` |
| `floor` | Floor | `{{ num\|floor }}` |
| `ceil` | Ceiling | `{{ num\|ceil }}` |
| `at_least` | Ensure minimum value | `{{ num\|at_least:0 }}` |
| `at_most` | Ensure maximum value | `{{ num\|at_most:100 }}` |

### Array

| Filter | Description | Example |
|--------|-------------|---------|
| `join` | Join with separator (default `" "`) | `{{ items\|join:', ' }}` |
| `first` | First element | `{{ items\|first }}` |
| `last` | Last element | `{{ items\|last }}` |
| `size` | Collection or string length | `{{ items\|size }}` |
| `reverse` | Reverse order | `{{ items\|reverse }}` |
| `sort` | Sort elements | `{{ items\|sort }}` |
| `sort_natural` | Case-insensitive sort | `{{ items\|sort_natural }}` |
| `uniq` | Remove duplicates | `{{ items\|uniq }}` |
| `compact` | Remove nil values | `{{ items\|compact }}` |
| `concat` | Combine two arrays | `{{ items\|concat:more }}` |
| `map` | Extract key from each element | `{{ users\|map:'name' }}` |
| `where` | Select items matching key/value | `{{ users\|where:'active','true' }}` |
| `reject` | Reject items matching key/value | `{{ users\|reject:'active','false' }}` |
| `find` | Find first matching item | `{{ users\|find:'name','Bob' }}` |
| `find_index` | Find index of first match | `{{ users\|find_index:'name','Bob' }}` |
| `has` | Check if any item matches | `{{ users\|has:'name','Alice' }}` |
| `sum` | Sum values (supports property) | `{{ scores\|sum }}` |

**Array extensions** (not in Liquid standard):

| Filter | Description | Example |
|--------|-------------|---------|
| `shuffle` | Random shuffle | `{{ items\|shuffle }}` |
| `random` | Random element | `{{ items\|random }}` |
| `max` | Maximum value | `{{ scores\|max }}` |
| `min` | Minimum value | `{{ scores\|min }}` |
| `average` | Average | `{{ scores\|average }}` |

### Date

| Filter | Description | Example |
|--------|-------------|---------|
| `date` | Format date (PHP-style) | `{{ timestamp\|date:'Y-m-d' }}` |
| `day` | Extract day | `{{ timestamp\|day }}` |
| `month` | Extract month number | `{{ timestamp\|month }}` |
| `month_full` | Full month name | `{{ timestamp\|month_full }}` |
| `year` | Extract year | `{{ timestamp\|year }}` |
| `week` | ISO week number | `{{ timestamp\|week }}` |
| `weekday` | Day of week | `{{ timestamp\|weekday }}` |
| `time_ago` | Relative time | `{{ timestamp\|time_ago }}` |

### Number Formatting

| Filter | Description | Example |
|--------|-------------|---------|
| `number` | Number formatting | `{{ price\|number:'0.00' }}` |
| `bytes` | Convert to readable byte units | `{{ fileSize\|bytes }}` |

### Serialization

| Filter | Description | Example |
|--------|-------------|---------|
| `json` | Serialize to JSON | `{{ data\|json }}` |

### Map

| Filter | Description | Example |
|--------|-------------|---------|
| `extract` | Extract nested value by dot path | `{{ data\|extract:'user.name' }}` |

## Extension

### Custom Filters

```go
engine := template.New()
engine.Filters().Register("repeat", func(value any, args ...any) (any, error) {
    s := fmt.Sprintf("%v", value)
    n := 2
    if len(args) > 0 {
        if parsed, err := strconv.Atoi(fmt.Sprintf("%v", args[0])); err == nil {
            n = parsed
        }
    }
    return strings.Repeat(s, n), nil
})

tpl, _ := engine.ParseString(`{{ "ha"|repeat:3 }}`)
out, _ := tpl.Render(nil)
// out == "hahaha"
```

Use `Register` for new names, `Replace` when you intentionally override
an existing engine-local filter, and `MustRegister` in bootstrap-style
code where duplicate registration should panic.

### Custom Tags

Register custom tags by implementing the `Statement` interface and attaching the parser to an engine. Here's an example of a `{% set %}` tag:

```go
type SetNode struct {
    VarName    string
    Expression template.Expression
    Line, Col  int
}

func (n *SetNode) Position() (int, int) { return n.Line, n.Col }
func (n *SetNode) String() string       { return fmt.Sprintf("Set(%s)", n.VarName) }
func (n *SetNode) Execute(ctx *template.ExecutionContext, _ io.Writer) error {
    val, err := n.Expression.Evaluate(ctx)
    if err != nil {
        return err
    }
    ctx.Set(n.VarName, val.Interface())
    return nil
}

engine.Tags().Register("set", func(doc *template.Parser, start *template.Token, arguments *template.Parser) (template.Statement, error) {
    varToken, err := arguments.ExpectIdentifier()
    if err != nil {
        return nil, arguments.Error("expected variable name after 'set'")
    }
    if arguments.Match(template.TokenSymbol, "=") == nil {
        return nil, arguments.Error("expected '=' after variable name")
    }
    expr, err := arguments.ParseExpression()
    if err != nil {
        return nil, err
    }
    return &SetNode{
        VarName:    varToken.Value,
        Expression: expr,
        Line:       start.Line,
        Col:        start.Col,
    }, nil
})
```

As with filters, prefer `Register` for new names and `Replace` only when
you intentionally override an existing engine-local parser.

## Examples

Runnable examples live in the [`examples/`](examples/) directory:

| Directory | Demonstrates |
|---|---|
| [`examples/usage`](examples/usage) | Parsing and rendering source strings through `Engine` |
| [`examples/custom_filters`](examples/custom_filters) | Registering a custom filter on an `Engine` |
| [`examples/custom_tags`](examples/custom_tags) | Registering a custom tag on an `Engine` |
| [`examples/layout`](examples/layout) | Engine-based HTML site with `{% extends %}` / `{% block %}` / `{% include %}` / `block.super` / auto-escape |
| [`examples/multifile_text`](examples/multifile_text) | Engine-based text generation with `FormatText` and `FeatureLayout` |

Run any example with `go run ./examples/<name>`.

## Context Building

```go
engine := template.New()
tmpl, _ := engine.ParseString(source)

// Using map directly
output, _ := tmpl.Render(map[string]any{
    "name": "Alice",
    "age":  30,
})

// Using ContextBuilder (supports struct expansion)
ctx, err := template.NewContextBuilder().
    KeyValue("name", "Alice").
    Struct(user).
    Build()
output, _ := tmpl.Render(ctx)
```

## Error Reporting

All errors include precise line and column position information:

```
lexer error at line 1, col 7: unclosed variable tag, expected '}}'
parse error at line 1, col 4: unknown tag: unknown
parse error at line 1, col 19: unexpected EOF, expected one of: [elif else endif]
```

## Liquid Compatibility

This engine targets compatibility with the [Liquid](https://shopify.github.io/liquid/) template standard. Filter names follow Liquid conventions (`upcase`, `downcase`, `strip`, `at_least`, `divided_by`, etc.).

For a complete comparison including behavioral differences, missing filters, extension filters, and convenience aliases, see [Liquid Compatibility](docs/liquid-compatibility.md).

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for details.

## Contributing

Contributions are welcome. Please see [Contributing Guide](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE) for details.
