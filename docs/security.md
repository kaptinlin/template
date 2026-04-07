# Security Model

This page documents the threat model and defenses for the multi-file
template system. Single-string `template.Compile(src)` has **no loader
and no input other than the source string**, so this page applies only
to code that uses `NewHTMLSet` or `NewTextSet`.

## Threat model

**Trusted**: template authors. Templates you ship with your application
are assumed to be code you control. This library does **not** sandbox
template-level logic (infinite loops in `{% for %}`, large string
manipulations, etc.) — Django, Jinja2, and Liquid all take the same
position.

**Untrusted**: everything that flows **into** a template through data
or dynamic paths. Specifically:

| Source | Why it matters |
|---|---|
| User-submitted content inside `{{ expr }}` | XSS if rendered as HTML without escape |
| Dynamic include paths (`{% include page.widget %}`) | Path traversal if the value comes from untrusted input |
| Template names passed to `Set.Render(name, ...)` | Same — the caller may derive `name` from URL parameters or frontmatter |
| Symbolic links in the templates directory | Accidental or malicious escape of the root |
| File names on case-insensitive filesystems | Cache collisions in multi-layer chains |
| Pre-rendered HTML from upstream services | Must be explicitly marked `SafeString` — never infer "trusted" |

## Threat matrix

| # | Threat | Example | Defense |
|---|---|---|---|
| T1 | Relative path escape | `{% include "../../etc/passwd" %}` | `ValidateName` rejects `..` |
| T2 | Absolute path | `{% include "/etc/passwd" %}` | `ValidateName` rejects `/`-prefixed names |
| T3 | Symlink escape | File in root is a symlink to `/etc/hosts` | `NewDirLoader` uses `os.Root` — syscall-level refuse |
| T4 | Path injection characters | NUL byte, backslash, newline | `ValidateName` rejects NUL and backslash |
| T5 | Windows path aliases | `C:\...`, `\\server\share`, `\\?\...` | `ValidateName` rejects backslash universally |
| T6 | Case-insensitive cache collision | `Header.html` vs `header.html` on macOS | `ChainLoader` adds layer prefixes; multiple cache entries coexist safely |
| T7 | TOCTOU | File replaced between check and open | `os.Root` uses `openat2` / `O_NOFOLLOW` — no separate check step |
| T8 | Mutual include cycle | A → B → A | Parse-time detection downgrades to lazy; runtime hits depth cap |
| T9 | Self-include / deep chain | A → A or 50-level nesting | Hard-coded `maxIncludeDepth = 32`, `maxExtendsDepth = 10` |
| T10 | HTML injection (XSS) | `{{ user_title }}` contains `<script>` | `NewHTMLSet` auto-escapes by default |
| T11 | Double-escape bypass | `{{ x \| safe \| some_filter }}` | Non-safe-aware filters downgrade `SafeString` back to plain |

## Layered defense

### Layer 1 — `ValidateName`

Every `Loader.Open` implementation must call `ValidateName(name)` as
the first thing it does. This is a stricter version of `fs.ValidPath`:

```go
func ValidateName(name string) error {
    if !fs.ValidPath(name) {      // rejects "", "..", absolute, trailing /
        return ErrInvalidTemplateName
    }
    if strings.ContainsAny(name, "\\\x00") {  // plus backslash and NUL
        return ErrInvalidTemplateName
    }
    return nil
}
```

Why stricter than `fs.ValidPath`? `fs.ValidPath` allows backslash and
NUL (they're valid UTF-8 code points that Go considers non-path-
significant on Unix). We reject them universally to:

- Prevent Windows path aliases (`C:\foo`, `\\server\share`) from ever
  reaching an FS that might interpret them.
- Prevent NUL-byte injection attacks that truncate paths when passed
  to C syscalls.

### Layer 2 — `os.Root` sandbox (DirLoader)

`NewDirLoader` uses Go 1.24+'s `os.Root` primitive:

```go
root, err := os.OpenRoot(dir)
// ...
f, err := root.Open(name)
```

`os.Root` guarantees:

- **Symbolic links cannot cross the root boundary** at the syscall
  level. If a file inside the root is a symlink pointing outside, the
  `Open` call fails.
- `..` and absolute paths are rejected again at the OS layer, forming
  a second line of defense after Layer 1.
- TOCTOU is closed: it uses `openat2` or `O_NOFOLLOW` primitives,
  not "check then open".

This is stronger than `os.DirFS`, which Go's documentation explicitly
notes "does not prevent symbolic links from referring to files outside
the directory."

### Layer 3 — FS loader contract

`NewFSLoader(fs.FS)` assumes the caller has already picked a sandboxed
filesystem. Safe sources include:

- `embed.FS` — compiled into the binary, immutable
- `fstest.MapFS` — in-memory test double
- `archive/zip.Reader` — bounded archive

**Explicitly non-safe** source: `os.DirFS(path)`. If you pass this,
the library cannot prevent symlink escape — you are opting into the
less-strict Go primitive. The documented use case is developer
workflows (theme dev, monorepo sharing) where you trust the local
filesystem.

### Layer 4 — Resolved-name cache keys

`ChainLoader` prepends `layer0:`, `layer1:`, ... to each layer's
resolved name, so:

- User `layouts/blog.html` and theme `layouts/blog.html` get distinct
  cache entries (`layer0:layouts/blog.html` vs `layer1:layouts/blog.html`)
- A macOS case-insensitive collision (`Header.html` vs `header.html`)
  produces two cache entries referring to the same underlying file —
  a small amount of wasted memory, but no correctness bug.

### Layer 5 — Cycle and depth caps

Multiple defenses against runaway recursion:

1. **Parse-time circular detection** (`Set.parsing` map). When
   parsing template A, any static `{% include "B" %}` in A where B is
   already mid-parse is downgraded to lazy (runtime) mode. This
   enables recursive tree-walk patterns while preventing parse-time
   infinite recursion.
2. **Parse-time extends circular detection**. Unlike include, extends
   cannot be lazy — circular extends returns `ErrCircularExtends`.
3. **Runtime include depth cap**: `maxIncludeDepth = 32`. Deeper
   nesting returns `ErrIncludeDepthExceeded` before the Go stack
   overflows.
4. **Runtime extends depth cap**: `maxExtendsDepth = 10`. Returns
   `ErrExtendsDepthExceeded`.

Both caps are hard-coded constants. They are safety nets, not
performance tunables. If your architecture legitimately needs more,
something is probably wrong.

### Layer 6 — HTML auto-escape (HTMLSet only)

`NewHTMLSet` wires `ec.autoescape = true`, which causes
`OutputNode.Execute` to pipe every `{{ expr }}` output through
`filter.Escape` before writing — **unless** the underlying value is a
`SafeString`.

`SafeString` is opt-in, not opt-out:

- The `safe` filter wraps a value: `{{ content | safe }}`
- Go code can construct one directly:
  `template.SafeString("<p>trusted</p>")`
- The HTMLSet overrides of `escape`, `escape_once`, and `h` return
  `SafeString`, so `{{ x | escape }}` doesn't double-escape.

**Conservative downgrade**: if any non-safe-aware filter runs on a
`SafeString`, the wrapper is stripped and the value becomes plain
string again (subject to auto-escape). This prevents "I thought I was
safe" XSS bugs like `{{ user_input | safe | upper }}`. The only
safe-aware filters in v1 are `safe` and (in HTMLSet) the `escape`
family.

## What this library does NOT defend against

### Template author malice

Template authors can write infinite loops (`{% for x in items %}` where
`items` is mutated to never terminate), allocate arbitrary amounts of
memory via filters, or traverse the data context to leak information.
**This is by design** — template authors are trusted.

If you need to run untrusted templates (a hosted CMS, a sandboxed
scripting feature), this library is not a safe choice. Use a dedicated
sandbox (e.g. `sanitize + timeout + memory cap` around a separate
process).

### Complex JavaScript / CSS escape contexts

`NewHTMLSet` auto-escapes for an HTML text context: `<`, `>`, `&`,
`'`, `"`. It does **not** do context-aware escaping for:

- JavaScript strings (`<script>var x = "{{ user }}"</script>`)
- CSS identifiers (`<style>.{{ class }} { ... }</style>`)
- URL components (`<a href="{{ url }}">` is OK for text-context, but
  `javascript:` schemes are not filtered)
- JSON embedded in HTML attributes

For these contexts, use `filter.Escape` only as a baseline. Treat
embedded JS/CSS/URL values as separate escape problems and solve them
at the data layer before passing to the template.

### CSRF, authentication, authorization

This is a template engine. It doesn't know who the user is, what
they're allowed to see, or whether this request is authorized. Do not
use `{% if user.is_admin %}` as a security boundary — check at the
controller layer before rendering.

## Testing your security posture

The library ships a security matrix covering T1–T11 in
`security_test.go`. When writing a custom loader, run the same matrix
against it:

```go
func TestMyLoader_Security(t *testing.T) {
    t.Parallel()

    loader := NewMyLoader(...)
    invalid := []string{
        "../etc/passwd",
        "/etc/passwd",
        "a\\b",
        "a\x00b",
        "C:\\x",
        "\\\\server\\share",
        "",
        ".",
    }
    for _, name := range invalid {
        if _, _, err := loader.Open(name); !errors.Is(err, ErrInvalidTemplateName) {
            t.Errorf("Open(%q) should return ErrInvalidTemplateName, got %v", name, err)
        }
    }
}
```

If your loader touches the real filesystem, also add the symlink-
escape test from `security_test.go` (`TestDirLoader_SymlinkEscape_Rejected`).
