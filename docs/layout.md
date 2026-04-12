# Layout & Inheritance

This page covers the multi-file layout features: `{% include %}`,
`{% extends %}`, `{% block %}`, `{{ block.super }}`, and `{% raw %}`.

> **Scope note**: these are `FeatureLayout` features. They are available
> only when an engine enables `template.WithFeatures(template.FeatureLayout)`.

## Quick example

```go
loader, _ := template.NewDirLoader("./templates")
engine := template.New(
    template.WithLoader(loader),
    template.WithFormat(template.FormatHTML),
    template.WithFeatures(template.FeatureLayout),
)
_ = engine.Render("layouts/blog.html", template.Context{
    "page": pageData,
}, os.Stdout)
```

Templates in `./templates/`:

```django
{# layouts/base.html — the shared skeleton #}
<!DOCTYPE html>
<html>
<head>{% block head %}<title>{{ site.title }}</title>{% endblock %}</head>
<body>
  {% include "partials/header.html" %}
  <main>{% block content %}{% endblock %}</main>
  {% include "partials/footer.html" %}
</body>
</html>
```

```django
{# layouts/blog.html — inherits base.html #}
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

---

## `{% include %}`

Renders another template at the current position.

### Basic forms

```django
{% include "partials/header.html" %}                  {# static literal path  #}
{% include page.widget %}                              {# dynamic expression path #}
```

**Static paths** are resolved at parse time — missing templates fail
immediately. **Dynamic paths** are re-validated and loaded on each
render, so they're slightly slower but enable data-driven composition.

### Passing variables with `with`

```django
{% include "partials/card.html" with title="Hi" count=3 %}
{% include "partials/card.html" with greeting=page.hello items=user.roles %}
```

- The values are evaluated in the **parent** context (so `page.hello`
  means the outer template's `page`, not the included template's).
- Bindings land in the included template's `Private` scope — they do
  not mutate the parent.
- Child execution preserves runtime state from the parent render:
  engine-local filters/tags, auto-escape mode, include depth, and the
  current extends leaf all stay intact.

### Isolating context with `only`

```django
{% include "partials/card.html" only %}                {# no parent vars visible #}
{% include "partials/card.html" with title="Hi" only %} {# only "title" is visible #}
```

`only` fully isolates the included template. It does not inherit the
parent's context **and** it does not see `WithGlobals`-set values. The
only variables available to the child are those passed via `with`.

This isolation affects data visibility only. Rendering semantics still
come from the parent engine, so HTML mode stays HTML mode and include
depth still advances.

This matches Django DTL and Pongo2 semantics. If you need the child to
see `site.*` from globals while still hiding the page-level state, pass
it explicitly:

```django
{% include "partials/card.html" with site=site title="Hi" only %}
```

### Optional includes with `if_exists`

```django
{% include "partials/sidebar.html" if_exists %}
```

If the template cannot be loaded (not found in any loader layer), the
tag silently renders nothing instead of erroring. Useful for optional
theme hooks.

### Combined options

```django
{% include "card.html" with title="Hi" only if_exists %}
```

Order: `with` pairs come first, then `only`, then `if_exists`. `with`
may be omitted, in which case `only` / `if_exists` can appear standalone
in any combination.

### Runtime safeguards

- **Depth cap**: include nesting is limited to 32 levels. Exceeding this
  returns `ErrIncludeDepthExceeded`.
- **Path validation**: dynamic path results are re-checked against
  `fs.ValidPath`; `..`, absolute paths, backslashes, and NUL bytes are
  rejected.
- **Circular recursion support**: parse-time detection of `A includes B`
  and `B includes A` automatically downgrades one side to lazy mode, so
  recursive data-driven rendering (tree walks) works. Runtime recursion
  still hits the depth cap.

---

## `{% extends %}` + `{% block %}`

Template inheritance lets a child template override named regions of a
parent. The parent's body is what actually renders; the child just
provides the overrides.

### Basic inheritance

```django
{# parent.html #}
<h1>{% block title %}Default{% endblock %}</h1>
<main>{% block content %}{% endblock %}</main>

{# child.html #}
{% extends "parent.html" %}
{% block content %}<p>Hello, world</p>{% endblock %}
```

Rendering `child.html` produces:

```html
<h1>Default</h1>
<main><p>Hello, world</p></main>
```

- The `title` block is not overridden, so the parent's default renders.
- The `content` block is overridden by the child.
- Any text or tag **outside** blocks in the child is discarded. Only
  `{% extends %}` + `{% block %}` nodes matter at the child's top level.

### Multi-level inheritance

```django
{# a.txt #}
{% block x %}A{% endblock %}

{# middle.txt #}
{% extends "a.txt" %}
{% block x %}M{% endblock %}

{# leaf.txt #}
{% extends "middle.txt" %}
{% block x %}L{% endblock %}
```

Rendering `leaf.txt` produces `L` — the deepest (most-child) override
always wins. Rendering `middle.txt` produces `M`. Rendering `a.txt`
produces `A`.

Max chain depth is 10 layers.

### `{{ block.super }}` — calling the parent block

Inside an overriding block, `{{ block.super }}` renders the parent
block's content. It works through any number of layers:

```django
{# a.txt #}
{% block x %}A{% endblock %}

{# middle.txt #}
{% extends "a.txt" %}
{% block x %}M({{ block.super }}){% endblock %}

{# leaf.txt #}
{% extends "middle.txt" %}
{% block x %}L[{{ block.super }}]{% endblock %}
```

Rendering `leaf.txt` → `L[M(A)]`.

**Safety**: in an engine using `FormatHTML`, the super output is already rendered HTML
and wrapped in `SafeString`, so it is **not** re-escaped when
interpolated.

### Constraints

| Rule | Error |
|---|---|
| `{% extends %}` must be the first non-whitespace, non-comment tag | `ErrExtendsNotFirst` |
| `{% extends %}` path must be a string literal | `ErrExtendsPathNotLiteral` |
| Duplicate block names within a single template | `ErrBlockRedefined` |
| Circular extends (`A extends B`, `B extends A`) | `ErrCircularExtends` |
| Chain deeper than 10 levels | `ErrExtendsDepthExceeded` |
| Missing parent template | `ErrTemplateNotFound` |

### Optional endblock name

```django
{% block content %}
  <article>...</article>
{% endblock content %}
```

The name after `endblock` is optional but must match if present. This
makes large templates easier to scan.

### Blocks inside `{% include %}`

Blocks inside an included template do **not** participate in the
extends chain. They simply render their own body inline. The included
template is a self-contained unit — it has no relationship to the
outer template's inheritance hierarchy.

```django
{# partial.txt contains a block for structure #}
{% block widget %}<div>widget</div>{% endblock %}

{# page.txt uses include, not extends #}
Page: {% include "partial.txt" %}
{# → "Page: <div>widget</div>" — the block is rendered inline #}
```

---

## `{% raw %}...{% endraw %}`

Outputs a literal block of template-like text without interpretation.

```django
{% raw %}
  Template syntax demo: {{ variable }} and {% for x in items %}
{% endraw %}
```

Renders literally (including the braces and percent signs). Useful for:

- Generating templates for other template engines (Taskfile, Helm,
  Jinja2, etc.) whose syntax clashes with ours.
- Showing example template code inside documentation pages.
- Outputting string literals that happen to contain `{{` or `{%`.

**Errors**: missing `{% endraw %}` returns `ErrUnclosedRaw`.

**Engine note**: `{% raw %}` is lexer-level, so it requires
`FeatureLayout` to be enabled on the owning engine.

---

## Writing HTML: `safe` and auto-escape

An engine using `FormatHTML` auto-escapes every `{{ expr }}` output:

```django
{{ page.title }}
```

If `page.title` is `"Hello <world>"`, the output is `Hello &lt;world&gt;`.
This is the XSS defense for HTML rendering.

To output pre-rendered HTML without escaping, either:

**1. Use the `safe` filter** in the template:

```django
{{ page.content | safe }}
```

**2. Wrap the value in `SafeString`** in Go code:

```go
engine.Render("page.html", template.Context{
    "title":   "Hello <world>",                         // escaped
    "content": template.SafeString("<p>trusted</p>"),   // raw
}, w)
```

### Filter chain safety

`safe` status survives only as long as every filter in the chain is
safe-aware. `safe` and the `FormatHTML` override of `escape` are the only
safe-aware filters shipped. Any other filter downgrades the value:

```django
{{ x | safe }}                  → kept as SafeString → NOT escaped
{{ x | safe | upper }}          → upper returns string → RE-escaped
{{ x | upper | safe }}          → safe at terminal → NOT escaped
```

This conservative downgrade matches Jinja2's `Markup` semantics and
prevents "I thought I was safe" XSS bugs.

### Text mode

An engine using `FormatText` does **not** auto-escape. `SafeString` and the `safe`
filter still exist but are no-ops: they just produce the underlying
string. The `escape` filter in text mode falls through to the global
version and still returns plain `string`.

---

## Comparison to Django DTL / Jinja2 / Pongo2

This engine implements **Django DTL's** inheritance semantics:

- `extends` must be first (Django, not Jinja2's "can be anywhere")
- `block.super` is an attribute access (Django, not Jinja2's `super()`
  function call)
- `include ... with k=v` keyword arguments (Django, not Jinja2's
  `with/without context` boolean)
- `only` fully isolates, matching Django and Pongo2
- Multi-level inheritance and `block.super` chaining (all three)

The only deliberate deviation from Django DTL:

- `{% extends %}` path must be a string literal. Dynamic inheritance
  should be handled in Go code (picking the right template name before
  calling `Render`). This matches Pongo2 and enables fail-fast parse-
  time errors and parse-time dependency graph analysis.
