# Loaders

A `Loader` resolves a template name to its source text. `NewHTMLSet`
and `NewTextSet` both take a loader as their first argument:

```go
set := template.NewHTMLSet(loader)
```

The package ships four loader constructors, each suited to a different
use case. Most real projects wire several together with
`NewChainLoader` to build an override chain (user > theme > builtin).

## The `Loader` interface

```go
type Loader interface {
    Open(name string) (source string, resolved string, err error)
}
```

- `source` — the raw template text
- `resolved` — a stable, unique key used for the Set's compile cache
  and circular-reference detection. For simple loaders this is usually
  just the input name; `ChainLoader` prepends a layer index.
- Errors:
  - `ErrInvalidTemplateName` — path failed validation
    (`fs.ValidPath`, no backslash, no NUL)
  - `ErrTemplateNotFound` — name is valid but not present

All built-in loaders call `ValidateName(name)` first. Custom loaders
must too.

---

## `NewMemoryLoader`

In-memory map. Intended for tests and small pre-registered sets:

```go
loader := template.NewMemoryLoader(map[string]string{
    "base.html":  `<html>{% block body %}{% endblock %}</html>`,
    "child.html": `{% extends "base.html" %}{% block body %}hi{% endblock %}`,
})

set := template.NewHTMLSet(loader)
out, _ := set.RenderString("child.html", nil)
// <html>hi</html>
```

The input map is copied on construction — later mutations do not
affect the loader.

---

## `NewFSLoader`

Wraps any `fs.FS`:

```go
//go:embed themes/default/*
var themeFS embed.FS

loader := template.NewFSLoader(themeFS)
set := template.NewHTMLSet(loader)
```

Works with:

- `embed.FS` — production theme assets compiled into the binary
- `fstest.MapFS` — test doubles
- `archive/zip.Reader` — loading templates from a zip archive
- `os.DirFS` — **explicitly non-sandboxed** local directory (see
  "Dev workflows" below)
- Any custom `fs.FS` implementation

### When to use FSLoader vs DirLoader

- `embed.FS` / `fstest.MapFS` / zip → `NewFSLoader` (no sandbox
  needed, the FS itself is the boundary)
- Real directory on disk with untrusted template names → `NewDirLoader`
  (uses `os.Root` to block symlink escape)
- Real directory on disk where you **want** symlinks to follow
  (theme dev, monorepo sharing) → `NewFSLoader(os.DirFS(dir))` —
  you are explicitly opting into the non-sandboxed primitive and
  taking responsibility for the environment.

---

## `NewDirLoader`

Local directory, sandboxed by `os.Root` (Go 1.24+):

```go
loader, err := template.NewDirLoader("./templates")
if err != nil {
    log.Fatal(err) // directory doesn't exist or is unreadable
}
```

### Security guarantees

- All template names go through `ValidateName`: no `..`, no absolute
  paths, no backslash, no NUL.
- All file access goes through `os.Root.Open`, which refuses to
  follow symbolic links that point outside the root. This closes path
  traversal attacks even when the template name comes from untrusted
  input (frontmatter, URL params, database rows).
- TOCTOU is closed by `os.Root` using `openat2` / `O_NOFOLLOW` primitives
  — not "check then open".

### When NOT to use DirLoader

If your workflow deliberately relies on symbolic links (theme dev,
monorepo theme sharing), use the documented escape hatch:

```go
loader := template.NewFSLoader(os.DirFS("./templates"))
```

`os.DirFS` does **not** sandbox symlinks — the Go standard library
documents this explicitly. By writing it out at the call site, you
make the tradeoff visible to reviewers.

---

## `NewChainLoader`

Consults a list of loaders in order, returning the first hit:

```go
user, _ := template.NewDirLoader("./templates")           // writable user overrides
theme, _ := template.NewDirLoader("./themes/default")     // the active theme

//go:embed themes/default/*
var defaultThemeFS embed.FS
builtin := template.NewFSLoader(defaultThemeFS)           // fallback baked into the binary

loader := template.NewChainLoader(user, theme, builtin)
```

### Override semantics

When the user requests `layouts/blog.html`:

1. `user.Open("layouts/blog.html")` — if present, returned immediately
2. Otherwise `theme.Open("layouts/blog.html")`
3. Otherwise `builtin.Open("layouts/blog.html")`
4. Otherwise `ErrTemplateNotFound`

Same-name files in different layers **do not collide** in the cache:
the `ChainLoader` prepends the layer index (`layer0:`, `layer1:`, ...)
to the resolved name. So even on case-insensitive filesystems (macOS
HFS+/APFS), templates with the same name in different layers get
distinct cache entries.

### Empty chains

`NewChainLoader()` (no arguments) is a valid loader — it simply
returns `ErrTemplateNotFound` for every request. Useful as a placeholder
during tests.

---

## Custom loaders

The `Loader` interface is open. Any type that implements
`Open(name string) (string, string, error)` and honors the
`ValidateName` contract qualifies.

### Example: HTTP loader

```go
type httpLoader struct {
    baseURL string
    client  *http.Client
}

func (l *httpLoader) Open(name string) (string, string, error) {
    if err := template.ValidateName(name); err != nil {
        return "", "", err
    }
    resp, err := l.client.Get(l.baseURL + "/" + name)
    if err != nil {
        return "", "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode == http.StatusNotFound {
        return "", "", fmt.Errorf("%w: %q", template.ErrTemplateNotFound, name)
    }
    if resp.StatusCode != http.StatusOK {
        return "", "", fmt.Errorf("http %d", resp.StatusCode)
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", "", err
    }
    return string(body), "http:" + name, nil
}
```

**Caveats for remote loaders:**

- Error on network failure — don't swallow as "not found" unless the
  HTTP status is actually 404.
- Prepend a scheme/host prefix to the resolved name so different
  backends get distinct cache keys in a `ChainLoader`.
- Consider a timeout; the calling `Set.Render` will block on `Open`.
- Remember that `Set.Get` caches the compiled template — the HTTP
  request happens only once per name until `Set.Reset()` is called.

### Example: Database loader

```go
type dbLoader struct {
    db *sql.DB
}

func (l *dbLoader) Open(name string) (string, string, error) {
    if err := template.ValidateName(name); err != nil {
        return "", "", err
    }
    var src string
    err := l.db.QueryRow("SELECT source FROM templates WHERE name = ?", name).Scan(&src)
    if errors.Is(err, sql.ErrNoRows) {
        return "", "", fmt.Errorf("%w: %q", template.ErrTemplateNotFound, name)
    }
    if err != nil {
        return "", "", err
    }
    return src, "db:" + name, nil
}
```

---

## Hot reload

`Set` caches compiled templates indefinitely. To pick up on-disk
changes, call `set.Reset()` after the file watcher fires:

```go
watcher, _ := fsnotify.NewWatcher()
watcher.Add("./templates")

go func() {
    for event := range watcher.Events {
        if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
            set.Reset()
        }
    }
}()
```

`Set.Reset()` clears the cache; in-flight renders finish using their
already-loaded templates, and the next `Render` call recompiles.

Do not call `Reset` from request-handler hot paths — it forces every
template to recompile on next access.
