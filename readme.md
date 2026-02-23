# Template Engine

A lightweight Go template engine with Liquid/Django-style syntax, supporting variable interpolation, filters, conditionals, and loop control.

This engine implements the [Liquid](https://shopify.github.io/liquid/) filter standard with 41 out of 46 filters fully compliant, plus 20 extension filters. See [Liquid Compatibility](docs/liquid-compatibility.md) for details.

## Installation

```sh
go get github.com/kaptinlin/template
```

## Quick Start

### One-Step Rendering

```go
output, err := template.Render("Hello, {{ name|upcase }}!", map[string]any{
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
template.RegisterFilter("repeat", func(value any, args ...any) (any, error) {
    s := fmt.Sprintf("%v", value)
    n := 2
    if len(args) > 0 {
        if parsed, err := strconv.Atoi(fmt.Sprintf("%v", args[0])); err == nil {
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

## Liquid Compatibility

This engine targets compatibility with the [Liquid](https://shopify.github.io/liquid/) template standard. Filter names follow Liquid conventions (`upcase`, `downcase`, `strip`, `at_least`, `divided_by`, etc.).

For a complete comparison including behavioral differences, missing filters, extension filters, and convenience aliases, see [Liquid Compatibility](docs/liquid-compatibility.md).

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for details.

## Contributing

Contributions are welcome. Please see [Contributing Guide](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE) for details.
