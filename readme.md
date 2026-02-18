# Template Engine

A lightweight Go template engine with Liquid/Django-style syntax, supporting variable interpolation, filters, conditionals, and loop control.

## Installation

```sh
go get github.com/kaptinlin/template
```

## Quick Start

### One-Step Rendering

```go
output, err := template.Render("Hello, {{ name|upper }}!", map[string]any{
    "name": "alice",
})
// output: "Hello, ALICE!"
```

### Compile and Reuse

For templates that need to be rendered multiple times, compile once and reuse:

```go
tmpl, err := template.Compile("Hello, {{ name }}!")
if err != nil {
    log.Fatal(err)
}

output, err := tmpl.Render(map[string]any{"name": "World"})
// output: "Hello, World!"
```

### Using io.Writer

```go
tmpl, _ := template.Compile("Hello, {{ name }}!")
ctx := template.NewExecutionContext(map[string]any{"name": "World"})
tmpl.Execute(ctx, os.Stdout)
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
{{ name|upper }}
{{ title|truncate:20 }}
{{ name|lower|capitalize }}
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

See [Control Structure Documentation](docs/control-structure.md) for details.

### Comments

```
{# This content will not appear in the output #}
```

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
| `default` | Return default value if empty | `{{ name\|default:"Anonymous" }}` |
| `upper` | Convert to uppercase | `{{ name\|upper }}` |
| `lower` | Convert to lowercase | `{{ name\|lower }}` |
| `capitalize` | Capitalize first letter | `{{ name\|capitalize }}` |
| `titleize` | Capitalize first letter of each word | `{{ title\|titleize }}` |
| `trim` | Remove leading/trailing whitespace | `{{ text\|trim }}` |
| `truncate` | Truncate to specified length | `{{ text\|truncate:20 }}` |
| `truncateWords` | Truncate to specified word count | `{{ text\|truncateWords:5 }}` |
| `replace` | Replace substring | `{{ text\|replace:"old","new" }}` |
| `remove` | Remove substring | `{{ text\|remove:"bad" }}` |
| `append` | Append string | `{{ name\|append:"!" }}` |
| `prepend` | Prepend string | `{{ name\|prepend:"Hi " }}` |
| `split` | Split by delimiter | `{{ csv\|split:"," }}` |
| `length` | Get length | `{{ name\|length }}` |
| `slugify` | Convert to URL-friendly format | `{{ title\|slugify }}` |
| `camelize` | Convert to camelCase | `{{ name\|camelize }}` |
| `pascalize` | Convert to PascalCase | `{{ name\|pascalize }}` |
| `dasherize` | Convert to dash-separated | `{{ name\|dasherize }}` |
| `pluralize` | Singular/plural selection | `{{ count\|pluralize:"item","items" }}` |
| `ordinalize` | Convert to ordinal | `{{ num\|ordinalize }}` |

### Math

| Filter | Description | Example |
|--------|-------------|---------|
| `plus` | Addition | `{{ price\|plus:10 }}` |
| `minus` | Subtraction | `{{ price\|minus:5 }}` |
| `times` | Multiplication | `{{ price\|times:2 }}` |
| `divide` | Division | `{{ total\|divide:3 }}` |
| `modulo` | Modulo | `{{ num\|modulo:2 }}` |
| `abs` | Absolute value | `{{ num\|abs }}` |
| `round` | Round | `{{ pi\|round:2 }}` |
| `floor` | Floor | `{{ num\|floor }}` |
| `ceil` | Ceiling | `{{ num\|ceil }}` |
| `atLeast` | Ensure minimum value | `{{ num\|atLeast:0 }}` |
| `atMost` | Ensure maximum value | `{{ num\|atMost:100 }}` |

### Array

| Filter | Description | Example |
|--------|-------------|---------|
| `join` | Join with delimiter | `{{ items\|join:", " }}` |
| `first` | First element | `{{ items\|first }}` |
| `last` | Last element | `{{ items\|last }}` |
| `size` | Collection length | `{{ items\|size }}` |
| `reverse` | Reverse order | `{{ items\|reverse }}` |
| `unique` | Remove duplicates | `{{ items\|unique }}` |
| `shuffle` | Random shuffle | `{{ items\|shuffle }}` |
| `random` | Random element | `{{ items\|random }}` |
| `max` | Maximum value | `{{ scores\|max }}` |
| `min` | Minimum value | `{{ scores\|min }}` |
| `sum` | Sum | `{{ scores\|sum }}` |
| `average` | Average | `{{ scores\|average }}` |
| `map` | Extract specified key from each element | `{{ users\|map:"name" }}` |

### Map

| Filter | Description | Example |
|--------|-------------|---------|
| `extract` | Extract nested value by dot path | `{{ data\|extract:"user.name" }}` |

### Date

| Filter | Description | Example |
|--------|-------------|---------|
| `date` | Format date | `{{ timestamp\|date:"Y-m-d" }}` |
| `year` | Extract year | `{{ timestamp\|year }}` |
| `month` | Extract month number | `{{ timestamp\|month }}` |
| `monthFull` / `month_full` | Full month name | `{{ timestamp\|monthFull }}` |
| `day` | Extract day | `{{ timestamp\|day }}` |
| `week` | ISO week number | `{{ timestamp\|week }}` |
| `weekday` | Day of week | `{{ timestamp\|weekday }}` |
| `timeAgo` / `timeago` | Relative time | `{{ timestamp\|timeAgo }}` |

### Number Formatting

| Filter | Description | Example |
|--------|-------------|---------|
| `number` | Number formatting | `{{ price\|number:"0.00" }}` |
| `bytes` | Convert to readable byte units | `{{ fileSize\|bytes }}` |

### Serialization

| Filter | Description | Example |
|--------|-------------|---------|
| `json` | Serialize to JSON | `{{ data\|json }}` |

## Extension

### Custom Filters

```go
template.RegisterFilter("repeat", func(value any, args ...string) (any, error) {
    s := fmt.Sprintf("%v", value)
    n := 2
    if len(args) > 0 {
        if parsed, err := strconv.Atoi(args[0]); err == nil {
            n = parsed
        }
    }
    return strings.Repeat(s, n), nil
})

// {{ "ha"|repeat:3 }} -> "hahaha"
```

### Custom Tags

Register custom tags by implementing the `Statement` interface and calling `RegisterTag`. Here's an example of a `{% set %}` tag:

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

template.RegisterTag("set", func(doc *template.Parser, start *template.Token, arguments *template.Parser) (template.Statement, error) {
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

See [examples](examples/) directory for more examples.

## Context Building

```go
// Using map directly
output, _ := template.Render(source, map[string]any{
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

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for details.

## Contributing

Contributions are welcome. Please see [Contributing Guide](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE) for details.
