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

Tags are stored in a global map:

```go
// tags.go
var tagRegistry = make(map[string]TagParser)
```

`TagParser` function signature:

```go
type TagParser func(doc *Parser, start *Token, arguments *Parser) (Statement, error)
```

Parameters:
- `doc` — Document-level parser, used to parse nested content within tag bodies via `ParseUntil` / `ParseUntilWithArgs`
- `start` — Tag name token (e.g. `"if"`, `"for"`), carries source position for error reporting
- `arguments` — A dedicated parser scoped to the tokens between the tag name and `%}`,
  supporting `ParseExpression()`, `Match()`, `ExpectIdentifier()`, etc.

When the parser encounters `{% tagname ... %}`, it looks up `tagRegistry["tagname"]`
and invokes the corresponding function. If not found, it reports a parse error
(see the "Unknown tag" example in Part 2).

Built-in tags register themselves via `init()`:

```go
// tags.go
func init() {
    RegisterTag("if", parseIfTag)             // tag_if.go
    RegisterTag("for", parseForTag)           // tag_for.go
    RegisterTag("break", parseBreakTag)       // tag_break.go
    RegisterTag("continue", parseContinueTag) // tag_continue.go
}
```

### 3.2 Filter Registry

Filters use the same pattern:

```go
// filters.go
var filterRegistry = make(map[string]FilterFunc)

type FilterFunc func(value interface{}, args ...string) (interface{}, error)
```

Built-in filters register themselves via `init()` across multiple files
(`filter_string.go`, `filter_math.go`, `filter_array.go`, etc.).

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
type FilterFunc func(value interface{}, args ...string) (interface{}, error)
```

- `value` — The evaluated result of the expression to the left of `|`
- `args` — Arguments after `:` (e.g. in `{{ x|truncate:10 }}`, args = `["10"]`)

#### Registering a Filter from an External Package

```go
package myfilters

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/kaptinlin/template"
)

func init() {
    // Register a repeat filter: {{ text|repeat:3 }} -> "texttexttext"
    template.RegisterFilter("repeat", func(value interface{}, args ...string) (interface{}, error) {
        s := fmt.Sprintf("%v", value)
        n := 2 // default: repeat 2 times
        if len(args) > 0 {
            if parsed, err := strconv.Atoi(args[0]); err == nil {
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
3. **One Tag Per File** — The engine itself follows this convention (`tag_if.go`, `tag_for.go`, `tag_break.go`, `tag_continue.go`). New tags simply follow the same pattern.
4. **Fully Open Extension Points**:
   - **Tags** (external package registration supported): The `Statement` and `Expression` interfaces are fully open to external packages. External packages can implement `Statement` directly and register via `RegisterTag`.
   - **Filters** (external package registration supported): `FilterFunc` is a plain function type that any external package can register freely.
