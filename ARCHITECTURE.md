# Template Engine Architecture

This document walks through a complete example to demonstrate the full pipeline
from source to output, including actual output at each stage. It then covers
precise error reporting and the plugin-based extensibility design.

---

## Part 1: Source to Output — Full Pipeline

### Example Template and Data

```
Hello {{ name|upper }}!
{% if score > 80 %}Grade: A{% else %}Grade: B{% endif %}
```

```go
data := map[string]interface{}{
    "name":  "alice",
    "score": 95,
}
```

The engine processes templates in three stages:

```
Source string ──► Lexer ──► Token stream ──► Parser ──► AST ──► Execute ──► Output string
```

### Stage 1: Lexical Analysis (Lexer)

The lexer (`lexer.go`) scans the source string character by character,
producing a flat **token stream**. Each token carries a type, value,
and source position (line:col).

```go
lexer := template.NewLexer(source)
tokens, err := lexer.Tokenize()
```

Actual output (JSON format):

```json
[
  { "type": "TEXT",       "value": "Hello ", "line": 1, "col": 1 },
  { "type": "VAR_BEGIN",  "value": "{{",     "line": 1, "col": 7 },
  { "type": "IDENTIFIER", "value": "name",   "line": 1, "col": 10 },
  { "type": "SYMBOL",     "value": "|",      "line": 1, "col": 14 },
  { "type": "IDENTIFIER", "value": "upper",  "line": 1, "col": 15 },
  { "type": "VAR_END",    "value": "}}",     "line": 1, "col": 21 },
  { "type": "TEXT",       "value": "!\n",    "line": 1, "col": 23 },
  { "type": "TAG_BEGIN",  "value": "{%",     "line": 2, "col": 1 },
  { "type": "IDENTIFIER", "value": "if",     "line": 2, "col": 4 },
  { "type": "IDENTIFIER", "value": "score",  "line": 2, "col": 7 },
  { "type": "SYMBOL",     "value": ">",      "line": 2, "col": 13 },
  { "type": "NUMBER",     "value": "80",     "line": 2, "col": 15 },
  { "type": "TAG_END",    "value": "%}",     "line": 2, "col": 18 },
  { "type": "TEXT",       "value": "Grade: A", "line": 2, "col": 20 },
  { "type": "TAG_BEGIN",  "value": "{%",     "line": 2, "col": 28 },
  { "type": "IDENTIFIER", "value": "else",   "line": 2, "col": 31 },
  { "type": "TAG_END",    "value": "%}",     "line": 2, "col": 36 },
  { "type": "TEXT",       "value": "Grade: B", "line": 2, "col": 38 },
  { "type": "TAG_BEGIN",  "value": "{%",     "line": 2, "col": 46 },
  { "type": "IDENTIFIER", "value": "endif",  "line": 2, "col": 49 },
  { "type": "TAG_END",    "value": "%}",     "line": 2, "col": 55 },
  { "type": "EOF",        "value": "",       "line": 2, "col": 57 }
]
```

Key observations:

- Plain text `Hello ` becomes a single `TEXT` token
- `{{ ... }}` is split into `VAR_BEGIN` → inner tokens → `VAR_END`
- `{% ... %}` is split into `TAG_BEGIN` → inner tokens → `TAG_END`
- Variable names `name`, keywords `if`, `else`, `endif` are all `IDENTIFIER` tokens
- Operators `|`, `>` are `SYMBOL` tokens
- Number `80` is a `NUMBER` token
- Every token carries precise `line` and `col` (e.g. `score` is at line 2, col 7)

Core tokenization rules:

| Source pattern | Produced tokens |
|----------------|-----------------|
| Plain text | `TEXT` |
| `{{ ... }}` | `VAR_BEGIN`, inner tokens, `VAR_END` |
| `{% ... %}` | `TAG_BEGIN`, inner tokens, `TAG_END` |
| `{# ... #}` | Discarded (comment, no tokens produced) |
| Identifiers | `IDENTIFIER` (variable names, keywords like `if`/`for`) |
| `"..."` / `'...'` | `STRING` |
| `42`, `3.14` | `NUMBER` |
| `+`, `==`, `\|` | `SYMBOL` |

### Stage 2: Parsing (Parser + ExprParser)

The parser (`parser.go`) consumes the token stream and builds an **abstract syntax tree** (AST).

```go
parser := template.NewParser(tokens)
ast, err := parser.Parse()
```

Actual output (JSON format, showing the full recursive AST):

```json
[
  {
    "type": "Text",
    "text": "Hello ",
    "pos": { "line": 1, "col": 1 }
  },
  {
    "type": "Output",
    "expression": {
      "type": "Filter",
      "filter": "upper",
      "expression": {
        "type": "Variable",
        "name": "name",
        "pos": { "line": 1, "col": 10 }
      },
      "pos": { "line": 1, "col": 14 }
    },
    "pos": { "line": 1, "col": 7 }
  },
  {
    "type": "Text",
    "text": "!\n",
    "pos": { "line": 1, "col": 23 }
  },
  {
    "type": "If",
    "branches": [
      {
        "condition": {
          "type": "BinaryOp",
          "operator": ">",
          "left": {
            "type": "Variable",
            "name": "score",
            "pos": { "line": 2, "col": 7 }
          },
          "right": {
            "type": "Literal",
            "value": 80,
            "pos": { "line": 2, "col": 15 }
          },
          "pos": { "line": 2, "col": 13 }
        },
        "body": [
          {
            "type": "Text",
            "text": "Grade: A",
            "pos": { "line": 2, "col": 20 }
          }
        ]
      }
    ],
    "else_body": [
      {
        "type": "Text",
        "text": "Grade: B",
        "pos": { "line": 2, "col": 38 }
      }
    ],
    "pos": { "line": 2, "col": 4 }
  }
]
```

The AST root contains four elements:

- **`Text`** — Plain text node `"Hello "`, written directly to output
- **`Output`** — Variable output node containing a nested `Filter` expression: apply `upper` to variable `name`
- **`Text`** — Plain text node `"!\n"`
- **`If`** — Conditional node with 1 branch (`score > 80`) and an `else_body`

`Parse()` reads tokens one by one and dispatches by type:

- **`TEXT`** → Create a `Text` node
- **`VAR_BEGIN`** → Collect tokens until `VAR_END`, pass to expression parser `ExprParser`, create an `Output` node
- **`TAG_BEGIN`** → Read the tag name, look up the corresponding `TagParser` in the **tag registry**, delegate

Parsing `{% if score > 80 %}`:

1. Read tag name `"if"`, find `parseIfTag` in the registry
2. `parseIfTag` calls `arguments.ParseExpression()` to parse `score > 80` → produces a `BinaryOp` node
3. Calls `doc.ParseUntilWithArgs("elif", "else", "endif")` to parse the if-branch body
4. Encounters `else`, continues parsing the else body until `endif`
5. Returns an `If` node

### Stage 3: Execution

`Template.Execute()` (`template.go`) traverses the AST and produces output:

```go
tmpl := template.NewTemplate(ast)
var buf bytes.Buffer
ctx := template.NewExecutionContext(data)
tmpl.Execute(ctx, &buf)
```

The executor walks the root node list in order:

| Step | Node | Action | Output |
|------|------|--------|--------|
| 1 | `Text("Hello ")` | Write text directly | `Hello ` |
| 2 | `Output(Filter(Var(name)\|upper))` | Resolve `name` = `"alice"` → apply `upper` → `"ALICE"` | `ALICE` |
| 3 | `Text("!\n")` | Write text directly | `!\n` |
| 4 | `If(1 branches)` | Evaluate `score > 80` → `95 > 80` = `true` → execute first branch | `Grade: A` |

Actual output:

```
Hello ALICE!
Grade: A
```

### Convenience Wrappers

The three stages above are wrapped into convenience functions:

```go
// Compile = Lex + Parse + create Template
tmpl, err := template.Compile("Hello {{ name|upper }}!")

// Render = create ExecutionContext + Execute + extract string
output, err := tmpl.Render(map[string]interface{}{"name": "alice"})

// One-step shortcut:
output, err := template.Render("Hello {{ name|upper }}!", data)
```

---

## Part 2: Precise Error Reporting

The engine provides **line and column accurate** error messages during both
lexing and parsing, making it easy to locate problems in templates.

### Error Types

Two position-aware error types are defined:

```go
// lexer.go
type LexerError struct {
    Message string
    Line    int
    Col     int
}
// Format: "lexer error at line %d, col %d: %s"

// parser_helpers.go
type ParseError struct {
    Message string
    Line    int
    Col     int
}
// Format: "parse error at line %d, col %d: %s"
```

### Error Reporting Examples

All examples below are actual runtime output:

**Unclosed variable tag**

```
Template: "Hello {{ name"
Error:    lexer error at line 1, col 7: unclosed variable tag, expected '}}'
```

The lexer remembers the `{{` start position (1:7) and reports the unclosed tag at EOF.

**Unclosed string**

```
Template: '{{ "hello }}'
Error:    lexer error at line 1, col 4: unclosed string, expected "
```

Points precisely to the opening quote position (1:4).

**Unknown tag**

```
Template: "{% unknown %}"
Error:    parse error at line 1, col 4: unknown tag: unknown
```

Points to the tag name `unknown` at position (1:4).

**Friendly hints for common mistakes**

```
Template: "{% elif x %}"
Error:    parse error at line 1, col 4: unknown tag: elif (elif must be used inside an if block, not standalone)
```

For commonly misused tags like `elif`, `else`, `endif`, `endfor`, the parser provides
contextual hints rather than just "unknown tag".

**Unclosed tag block**

```
Template: "{% if true %}hello"
Error:    parse error at line 1, col 19: unexpected EOF, expected one of: [elif else endif]
```

At EOF, the parser reports the list of expected closing tags.

**Unclosed comment**

```
Template: "{# this is a comment"
Error:    lexer error at line 1, col 1: unclosed comment, expected '#}'
```

Points to the comment start position (1:1).

**Precise location in multi-line templates**

```
Template: "line 1\nline 2\n{{ name @ }}"
Error:    lexer error at line 3, col 9: unexpected character: @
```

In multi-line templates, the error pinpoints the illegal character `@` at line 3, col 9.

### Position Tracking

The lexer maintains three state variables:

```go
type Lexer struct {
    input  string   // Input template string
    pos    int      // Current position
    line   int      // Current line number (starts at 1)
    col    int      // Current column number (starts at 1)
    tokens []*Token // Collected tokens
}
```

On each `\n`, `line` is incremented and `col` resets to 1; otherwise `col` is incremented.
Each token records the current `line` and `col` at creation time, so subsequent parser
and executor stages can provide precise error reporting via token positions.

---

## Part 3: Plugin Architecture — Extending Without Modifying Existing Code

The engine uses a **registry pattern** for both tags and filters. Adding new functionality
requires only registering a parse function or filter function — no existing source files
need to be modified.

### 3.1 Tag Registry

Tags are stored in a `*TagRegistry` value type that supports a parent fallback:

```go
// tags.go
type TagRegistry struct {
    mu     sync.RWMutex
    tags   map[string]TagParser
    parent *TagRegistry  // optional fallback
}

var defaultTagRegistry = NewTagRegistry()
```

`TagRegistry.Get(name)` first checks its own map; on a miss, it consults
`parent` (if set). This layered design lets a `Set` layer its own private
tags over the global registry without touching `defaultTagRegistry` — the
foundation of the multi-file mode covered in Part 4.

`TagParser` function signature:

```go
type TagParser func(doc *Parser, start *Token, arguments *Parser) (Statement, error)
```

Parameters:
- `doc` — Document-level parser, used to parse nested content within tag bodies via `ParseUntil` / `ParseUntilWithArgs`. Tag parsers can reach the owning `Set` via `doc.Set()` (returns `nil` for standalone `Compile(src)`).
- `start` — Tag name token, carries source position for error reporting
- `arguments` — Dedicated parser scoped to the tokens between the tag name and `%}`

When the parser encounters `{% tagname ... %}`, it first checks `p.set.tags`
(the per-Set private registry) and falls back to `defaultTagRegistry`. If
no parser is found, it reports a parse error.

Built-in tags register themselves via `init()`:

```go
// tags.go
var builtinTags = []builtinTag{
    {"if", parseIfTag},           // tag_if.go
    {"for", parseForTag},         // tag_for.go
    {"break", parseBreakTag},     // tag_break.go
    {"continue", parseContinueTag}, // tag_continue.go
}

var layoutTags = []builtinTag{
    {"include", parseIncludeTag}, // tag_include.go — Set-only
    {"extends", parseExtendsTag}, // tag_extends.go — Set-only
    {"block", parseBlockTag},     // tag_block.go — Set-only
}

func init() {
    for _, bt := range builtinTags {
        _ = RegisterTag(bt.name, bt.parser)  // global registry
    }
    // layoutTags are NOT registered globally — they are layered into
    // Set-private registries by newSet() in set.go.
}
```

### 3.2 Filter Registry

Filters follow the same layered design:

```go
// filters.go
type Registry struct {
    mu      sync.RWMutex
    filters map[string]FilterFunc
    parent  *Registry  // optional fallback
}

var defaultRegistry = NewRegistry()

type FilterFunc func(value any, args ...any) (any, error)
```

Built-in filters register themselves via `init()` across multiple files
(`filter_string.go`, `filter_math.go`, `filter_array.go`, etc.). The
`safe` filter and HTMLSet's `SafeString`-returning `escape`/`escape_once`
overrides are **not** in the global registry — they live in Set-private
layers, same as layout tags.

`FilterNode.Evaluate` consults `ctx.set.filters` first (if the template
was loaded via a Set), falling back to the global `defaultRegistry`.

### 3.3 Example: Adding a `{% set %}` Tag

The `Statement` and `Expression` interfaces are fully open to external packages.
External packages can implement custom AST nodes directly and register them
via `RegisterTag()`.

Here is a complete example showing how to add a `{% set x = expr %}` tag
from an external package:

```go
package main

import (
    "fmt"
    "io"

    "github.com/kaptinlin/template"
)

// SetNode represents a {% set varname = expr %} statement.
type SetNode struct {
    VarName    string
    Expression template.Expression
    Line       int
    Col        int
}

func (n *SetNode) Position() (int, int) { return n.Line, n.Col }
func (n *SetNode) String() string       { return fmt.Sprintf("Set(%s)", n.VarName) }

// Execute evaluates the expression and stores the result in the context.
func (n *SetNode) Execute(ctx *template.ExecutionContext, _ io.Writer) error {
    val, err := n.Expression.Evaluate(ctx)
    if err != nil {
        return err
    }
    ctx.Set(n.VarName, val.Interface())
    return nil
}

func main() {
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

        if arguments.Remaining() > 0 {
            return nil, arguments.Error("unexpected tokens after expression")
        }

        return &SetNode{
            VarName:    varToken.Value,
            Expression: expr,
            Line:       start.Line,
            Col:        start.Col,
        }, nil
    })

    source := `{% set greeting = "Hello" %}{{ greeting }}, {{ name }}!`
    output, _ := template.Render(source, map[string]interface{}{
        "name": "World",
    })
    fmt.Println(output) // Hello, World!
}
```

#### Parser API Available to Tag Parsers

| Method | Purpose | Example |
|--------|---------|---------|
| `arguments.ParseExpression()` | Parse a complete expression | `{% if expr %}`, `{% set x = expr %}` |
| `arguments.ExpectIdentifier()` | Consume an identifier token | Read `item` in `{% for item in ... %}` |
| `arguments.Match(type, value)` | Consume current token if it matches, return nil otherwise | Check for `,`, `=`, `in`, etc. |
| `arguments.Remaining()` | Return the number of unconsumed tokens | Check for extra arguments |
| `arguments.Error(msg)` | Create a parse error at the current token position | All error reporting |
| `doc.ParseUntilWithArgs(tags...)` | Parse tag body until a closing tag is found | `{% if %}` parsing to `elif`/`else`/`endif` |

### 3.4 Example: Adding a Custom Filter

Adding filters is even simpler — just a single function call, and it works from external packages.

`FilterFunc` signature:

```go
type FilterFunc func(value any, args ...any) (any, error)
```

- `value` — The evaluated result of the expression to the left of `|`
- `args` — Evaluated arguments after `:` (e.g. in `{{ x|truncate:10 }}`, args = `[10]`)

#### Registering a Filter from an External Package

```go
package myfilters

import (
    "fmt"
    "strings"

    "github.com/kaptinlin/template"
)

func init() {
    // Register a repeat filter: {{ text|repeat:3 }} -> "texttexttext"
    template.RegisterFilter("repeat", func(value any, args ...any) (any, error) {
        s := fmt.Sprintf("%v", value)
        n := 2 // default: repeat 2 times
        if len(args) > 0 {
            if parsed, ok := args[0].(int); ok {
                n = parsed
            }
        }
        return strings.Repeat(s, n), nil
    })
}
```

#### Usage

```go
import _ "myproject/myfilters" // init() registers automatically

func main() {
    output, _ := template.Render(`{{ word|repeat:3 }}`, map[string]interface{}{
        "word": "ha",
    })
    fmt.Println(output) // hahaha
}
```

Using Go's `init()` mechanism and blank imports (`import _ "pkg"`),
filters are registered before the first template is compiled.

### 3.5 Architecture Overview

```
+----------------------------------------------------------+
|                    template.Compile()                     |
|                                                          |
|  Source --> Lexer --> Tokens --> Parser --> AST            |
|                                  |                       |
|                   +--------------+                       |
|                   v                                      |
|           tagRegistry[name]  <-- RegisterTag()           |
|                   |                                      |
|                   v                                      |
|           TagParser(doc, start, args) -> Statement       |
|                                                          |
+----------------------------------------------------------+
|                    template.Render()                      |
|                                                          |
|  AST --> Execute(ctx, writer)                            |
|               |                                          |
|               +-- TextNode     -> write text directly    |
|               +-- OutputNode   -> evaluate + write       |
|               |      |                                   |
|               |      +-- FilterNode                      |
|               |             |                            |
|               |      filterRegistry[name]                |
|               |             |  <-- RegisterFilter()      |
|               |             v                            |
|               |      FilterFunc(value, args) -> result   |
|               |                                          |
|               +-- IfNode       -> evaluate branches      |
|               +-- ForNode      -> iterate collection     |
|               +-- SetNode      -> your custom tag        |
+----------------------------------------------------------+
```

### 3.6 Core Design Principles

1. **Open/Closed Principle** — Open for extension, closed for modification. Add new tags by calling `RegisterTag()`; add new filters by calling `RegisterFilter()`.
2. **Unified Interface** — Custom tags use the exact same `Parser` and `ExecutionContext` API as built-in tags, with full access to expression parsing, nested body parsing, and error reporting.
3. **One Tag Per File** — The engine itself follows this convention. New tags simply follow the same pattern.
4. **Fully Open Extension Points**:
   - **Tags**: The `Statement` and `Expression` interfaces are open to external packages. External packages can implement `Statement` directly and register via `RegisterTag`.
   - **Filters**: `FilterFunc` is a plain function type that any external package can register freely.
   - **Loaders**: The `Loader` interface (Part 4) is open to external implementations.

---

## Part 4: Multi-file Template System (Set)

When a project outgrows single-string rendering — layouts, partials, theme
overrides, HTML auto-escape — the engine provides a second API path: the
`Set`. This part explains how the Set layers on top of the primitives
covered in Parts 1–3 **without disturbing the single-string contract**.

### 4.1 Two API paths, one package

```
+----------------------------------------------------+
|  Path A: Compile(src) / Render(src, ctx)           |
|  ─────────────────────────────────────────────────  |
|  For log lines, config snippets, simple text.      |
|  Lexer:    allowRaw = false                        |
|  Parser:   p.set = nil                             |
|  Tags:     defaultTagRegistry only                 |
|  Filters:  defaultRegistry only                    |
|  Escape:   off                                     |
|                                                    |
|  → Sees only: if / for / break / continue          |
|    + all global filters (escape returns string)    |
+----------------------------------------------------+

+----------------------------------------------------+
|  Path B: NewTextSet(loader) / NewHTMLSet(loader)   |
|  ─────────────────────────────────────────────────  |
|  For multi-file layouts, includes, inheritance.    |
|  Lexer:    allowRaw = true                         |
|  Parser:   p.set = the owning Set                  |
|  Tags:     Set.tags (private) → defaultTagRegistry |
|  Filters:  Set.filters (private) → defaultRegistry |
|  Escape:   HTMLSet only                            |
|                                                    |
|  → Adds: include / extends / block / raw / safe    |
|    HTMLSet also overrides escape/escape_once/h     |
|    to return SafeString.                           |
+----------------------------------------------------+
```

This mirrors Go's own `text/template` vs `html/template` split, inside a
single package via the **layered registry pattern**.

### 4.2 The Set type

```go
// set.go
type Set struct {
    loader     Loader
    autoescape bool
    globals    Context

    tags    *TagRegistry  // private, parent = defaultTagRegistry
    filters *Registry     // private, parent = defaultRegistry

    mu      sync.RWMutex
    cache   map[string]*Template  // loaded & compiled templates
    parsing map[string]bool       // resolved names currently mid-compile
}

func NewHTMLSet(loader Loader, opts ...SetOption) *Set  // autoescape = true
func NewTextSet(loader Loader, opts ...SetOption) *Set  // autoescape = false
```

The Set's three responsibilities:

1. **Template resolution** — `Get(name)` loads the source via `loader`,
   compiles it, caches the result, and returns it.
2. **Dependency graph** — During parse, `{% include "x" %}` and
   `{% extends "x" %}` call `Set.Get(x)` recursively. The `parsing`
   map detects cycles so `include` can downgrade to lazy mode
   (supports recursive tree rendering) and `extends` can error out
   (cycles in inheritance are never valid).
3. **Registry layering** — At construction, the Set builds private tag
   and filter registries layered over the global ones, then registers
   layout tags and (for HTMLSet) the `SafeString`-returning escape
   variants.

### 4.3 Loader interface

```go
// loader.go
type Loader interface {
    Open(name string) (source string, resolved string, err error)
}
```

Every loader must call `ValidateName(name)` before doing I/O.
`ValidateName` goes beyond `fs.ValidPath` by also rejecting backslash
and NUL bytes.

**Built-in loaders:**

| Loader | Purpose | Security |
|---|---|---|
| `NewMemoryLoader(map[string]string)` | Tests, small pre-registered sets | Pure in-memory |
| `NewFSLoader(fs.FS)` | `embed.FS`, `fstest.MapFS`, `zip.Reader` | `ValidateName` only — caller provides the sandbox |
| `NewDirLoader(dir)` | Local directory | `os.Root` — symbolic links cannot escape the root |
| `NewChainLoader(loaders...)` | User > theme > builtin overrides | First hit wins; resolved names get `layerN:` prefix to avoid cache collisions |

The `NewDirLoader` + `os.Root` pairing is critical: it uses Go 1.24+'s
root-relative filesystem primitive, which refuses symlinks that point
outside the root at the syscall level. This closes path-traversal
attacks even when the loader handles untrusted names (e.g. from
frontmatter or URL parameters).

For dev workflows that deliberately need symlink following (e.g.
monorepo theme sharing), the escape hatch is
`NewFSLoader(os.DirFS(dir))` — the caller explicitly opts into Go's
non-sandboxed primitive.

### 4.4 Layout syntax overview

All four layout features (`include`, `extends`, `block`, `raw`) are
**only** available when the template is loaded through a Set. From
`Compile(src)` they remain unknown tags.

#### `{% include %}`

```django
{% include "partials/header.html" %}
{% include "card.html" with title="Hi" count=3 %}
{% include "card.html" with title="Hi" only %}
{% include "sidebar.html" only %}
{% include "optional.html" if_exists %}
{% include page.widget %}                          {# dynamic path #}
```

- String-literal paths resolve at parse time for fail-fast errors.
- Expression paths are evaluated at runtime and re-validated against
  `fs.ValidPath`.
- `with k=v` keyword arguments are evaluated in the **parent**
  context, then injected into the child's Private scope.
- `only` isolates the child from the parent and from `WithGlobals`.
- `if_exists` turns a missing template into a silent no-op.
- Parse-time circular references (A → B → A) automatically downgrade
  to lazy mode so recursive tree-walk templates work.
- Runtime include depth is capped at 32 (`ErrIncludeDepthExceeded`).

#### `{% extends %}` + `{% block %}` + `{{ block.super }}`

```django
{# parent layouts/base.html #}
<html>
<head>{% block head %}<title>{{ site.title }}</title>{% endblock %}</head>
<body>{% block content %}{% endblock %}</body>
</html>

{# child layouts/blog.html #}
{% extends "layouts/base.html" %}
{% block head %}
  {{ block.super }}
  <meta name="description" content="{{ page.description }}">
{% endblock %}
{% block content %}
  <article>{{ page.content | safe }}</article>
{% endblock content %}
```

**Rules:**

- `{% extends %}` must be the first non-whitespace, non-comment tag.
  Violations return `ErrExtendsNotFirst`.
- `{% extends %}` path must be a string literal (dynamic parents are a
  Go-level concern). Violations return `ErrExtendsPathNotLiteral`.
- Circular extends chains return `ErrCircularExtends`; max chain depth
  is 10 (`ErrExtendsDepthExceeded`).
- Child content outside `{% block %}` tags is discarded (Django DTL
  semantics).
- `{% endblock %}` may optionally carry the block name for readability;
  mismatches are errors.
- Duplicate block names in the same template return `ErrBlockRedefined`.

**Execution model:**

```
                                   ┌────────────────┐
                                   │ leaf Template  │ ctx.currentLeaf
                                   └───────┬────────┘
                                           │ .parent
                                   ┌───────┴────────┐
                                   │ middle         │
                                   └───────┬────────┘
                                           │ .parent
                                   ┌───────┴────────┐
                                   │ root Template  │ runs this body
                                   └────────────────┘
                                           │
                                           ▼
                                   each BlockNode.Execute walks
                                   leaf→root collecting same-name
                                   blocks; deepest (child-most) wins
```

`{{ block.super }}` is implemented by pre-rendering the parent chain
into a buffer and injecting it as `SafeString` under the `block.super`
variable before executing the active override. Recursive rendering
naturally supports multi-level super calls.

#### `{% raw %}...{% endraw %}`

Implemented in the lexer. When `Lexer.allowRaw` is true (set only by
`compileForSet(src, set)` when `set != nil`), the lexer intercepts
`{% raw %}` at the token level and emits everything up to `{% endraw %}`
as a single `TEXT` token. No parser or tag machinery sees the raw body.

### 4.5 HTML auto-escape (HTMLSet only)

`OutputNode.Execute` checks two things before writing:

1. Is the raw value a `SafeString`? → write as-is, no escape.
2. Is `ctx.autoescape` true? → run through `filter.Escape` before writing.

`ctx.autoescape` is set by `Set.Render` to mirror `s.autoescape`.
`NewHTMLSet` sets it to `true`, `NewTextSet` to `false`.

**Filter chain safety:**

- `{{ x | safe }}` wraps the value in `SafeString`.
- `{{ x | escape }}` runs HTML-escape. In HTMLSet, the override returns
  `SafeString` so the auto-escape path won't double-escape.
- `{{ x | safe | upper }}` — `upper` is not safe-aware, so it
  stringifies the input and returns plain `string`. The
  `SafeString` tag is lost and the output is re-escaped. This conservative
  downgrade prevents "I thought I was safe" bugs (matches Jinja2's `Markup`
  semantics).

### 4.6 Set-level options

```go
func WithGlobals(g Context) SetOption    // shared variables for every Render
func WithFilters(r *Registry) SetOption  // custom filter registry (replaces Set.filters)
func WithTags(r *TagRegistry) SetOption  // custom tag registry (replaces Set.tags)
```

`WithGlobals` merges into every render's context. Render-time ctx keys
take precedence over globals.

### 4.7 Concurrency and hot reload

Compiled `*Template` instances are treated as read-only after `Set.Get`
returns them. Multiple goroutines can render the same cached template
concurrently without locks.

`Set.Reset()` clears the cache. Dev servers wire it into a file watcher
to pick up template edits on the next render.

### 4.8 Full Set pipeline diagram

```
+--------------------------------------------------------------+
|  set := NewHTMLSet(loader, WithGlobals(...))                 |
|                                                              |
|  set.Render("layouts/blog.html", ctx, w)                     |
|        │                                                     |
|        ▼                                                     |
|  Set.Get("layouts/blog.html")                                |
|        │                                                     |
|        ├── cache hit? → return cached *Template              |
|        │                                                     |
|        ├── loader.Open(name) → source + resolved name        |
|        │      │                                              |
|        │      └── ValidateName, os.Root (dir), ChainLoader   |
|        │                                                     |
|        ├── s.markParsing(resolved) = true                    |
|        │                                                     |
|        ├── compileForSet(source, s)                          |
|        │      │                                              |
|        │      ├── Lexer (allowRaw=true) → Tokens             |
|        │      │                                              |
|        │      ├── Parser (p.set = s)                         |
|        │      │      │                                       |
|        │      │      ├── {% include %}                       |
|        │      │      │     → s.tags.Get("include")            |
|        │      │      │     → parseIncludeTag resolves        |
|        │      │      │       sub-templates via s.Get()       |
|        │      │      │                                       |
|        │      │      ├── {% extends %}                       |
|        │      │      │     → parseExtendsTag loads parent    |
|        │      │      │       via s.Get(); sets p.parent      |
|        │      │      │                                       |
|        │      │      └── {% block %}                         |
|        │      │            → parseBlockTag stores in         |
|        │      │              p.blocks                        |
|        │      │                                              |
|        │      └── NewTemplate(ast); tpl.parent=p.parent;     |
|        │          tpl.blocks=p.blocks; tpl.set=s             |
|        │                                                     |
|        ├── cache[name] = tpl                                 |
|        ├── s.markParsing(resolved) = false                   |
|        │                                                     |
|        ▼                                                     |
|  tpl.Execute(ec, w)  where ec.set=s, ec.autoescape=true      |
|        │                                                     |
|        ├── walk tpl.parent chain to find root                |
|        ├── ec.currentLeaf = tpl                              |
|        ├── root.executeRoot(ec, w)                           |
|        │                                                     |
|        ├── for each statement:                               |
|        │     ├── TextNode       → write directly             |
|        │     ├── OutputNode     → evaluate, SafeString check,|
|        │     │                    then auto-escape if needed |
|        │     ├── IfNode         → evaluate branches          |
|        │     ├── ForNode        → iterate collection         |
|        │     ├── IncludeNode    → resolveChild +             |
|        │     │                    buildChildContext +        |
|        │     │                    child.Execute              |
|        │     ├── ExtendsNode    → no-op (handled in          |
|        │     │                    Template.Execute)          |
|        │     └── BlockNode      → walk ec.currentLeaf chain, |
|        │                          render deepest override    |
|        │                          with block.super pre-      |
|        │                          rendered as SafeString     |
|        │                                                     |
|        └── output                                            |
+--------------------------------------------------------------+
```

### 4.9 Security model

The multi-file mode's threat model assumes template authors are trusted
but **input paths are not** (they may come from frontmatter, URL
parameters, database rows, or untrusted expressions). Defense is
layered:

| Layer | Where | Defends against |
|---|---|---|
| 1. `ValidateName` | Every `Loader.Open` | `..`, absolute paths, backslash, NUL |
| 2. `os.Root` | `NewDirLoader` only | Symlink escape across root boundary |
| 3. FS loader contract | `NewFSLoader` docs | Caller must supply a sandboxed `fs.FS` |
| 4. Resolved-name cache keys | `Set.cache`, `ChainLoader` prefix | Cross-layer cache collisions (macOS case-insensitivity) |
| 5. Parse-time circular set + runtime depth cap | `Set.parsing`, `maxIncludeDepth`, `maxExtendsDepth` | Stack exhaustion, infinite recursion |
| 6. HTML auto-escape + `SafeString` | `OutputNode.Execute` | XSS via `{{ user_input }}` |

These layers are enforced unconditionally for Set-loaded templates.
They do not apply to `Compile(src)` because that path has no loader,
no HTML context, and no input other than the source string — the
caller is directly responsible for whatever they feed in.
