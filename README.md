# template

[![Go Version](https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A Go template engine with engine-first configuration, Liquid-style filters, and optional layout-aware HTML rendering

For development guidelines and project conventions, see [AGENTS.md](AGENTS.md).

## Features

- **Engine-first API**: Configure one `Engine` for source-string rendering or loader-backed template rendering.
- **Two output modes**: Use `FormatText` for raw text generation and `FormatHTML` for automatic HTML escaping.
- **Layout features on demand**: `WithLayout()` enables `include`, `extends`, `block`, `raw`, and `safe`.
- **Composable loaders**: Use memory, directory, `fs.FS`, and chained loaders for tests, embedded assets, and override layers.
- **Extension points**: Register engine-local filters and tags or implement custom `Loader`, `Statement`, and `Expression` types.
- **Sandboxed local loading**: `NewDirLoader` uses `os.Root` to keep local template reads inside the configured root.

## Installation

```bash
go get github.com/kaptinlin/template
```

Requires **Go 1.26+**.

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/template"
)

func main() {
	engine := template.New()
	tmpl, err := engine.ParseString("Hello, {{ name|upcase }}!")
	if err != nil {
		log.Fatal(err)
	}

	out, err := tmpl.Render(template.Data{"name": "alice"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
}
```

## Core Concepts

| Task | API | Notes |
|---|---|---|
| Parse a source string | `Engine.ParseString` + `Template.Render` / `Template.RenderTo` | Best for one-off templates |
| Render named templates | `WithLoader(...)` + `Engine.Render` / `Engine.RenderTo` | Compiles and caches by loader-resolved name |
| Choose output semantics | `WithFormat(FormatText)` / `WithFormat(FormatHTML)` | `FormatHTML` auto-escapes `{{ expr }}` output |
| Enable layout syntax | `WithLayout()` | Turns on `include`, `extends`, `block`, `raw`, and `safe` |
| Provide shared defaults | `WithDefaults(Data)` | Render-time keys override engine defaults |
| Extend behavior | `RegisterFilter`, `ReplaceFilter`, `RegisterTag`, `ReplaceTag` | Extensions stay local to the engine instance |

## Loader-Backed Rendering

Use a loader when templates live outside the source string:

```go
loader := template.NewMemoryLoader(map[string]string{
	"base.html": `<h1>{{ page.title }}</h1>{% block content %}{% endblock %}`,
	"page.html": `{% extends "base.html" %}{% block content %}{{ page.content }}{% endblock %}`,
})

engine := template.New(
	template.WithLoader(loader),
	template.WithFormat(template.FormatHTML),
	template.WithLayout(),
)

out, err := engine.Render("page.html", template.Data{
	"page": map[string]any{
		"title":   "Hello <world>",
		"content": "<p>escaped</p>",
	},
})
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
// <h1>Hello &lt;world&gt;</h1>&lt;p&gt;escaped&lt;/p&gt;
```

For full runnable programs, see the example directories below.

## Extension Points

| Need | API |
|---|---|
| Custom filter | `Engine.RegisterFilter`, `Engine.ReplaceFilter`, `Engine.MustRegisterFilter` |
| Custom tag | `Engine.RegisterTag`, `Engine.ReplaceTag`, `Engine.MustRegisterTag` |
| Custom loader | Implement `Loader` and pass it through `WithLoader(...)` |
| Direct runtime control | `Template.Execute` with `RenderContext` |
| Trusted HTML | `SafeString` or `| safe` with `FormatHTML` |

## Examples

Runnable examples live in [`examples/`](examples/):

| Example | Path | What it shows |
|---|---|---|
| Basic usage | [`examples/usage`](examples/usage) | Parse and render a source string through `Engine` |
| Custom filters | [`examples/custom_filters`](examples/custom_filters) | Register an engine-local filter |
| Custom tags | [`examples/custom_tags`](examples/custom_tags) | Register an engine-local tag parser |
| HTML layouts | [`examples/layout`](examples/layout) | `WithLayout()`, `FormatHTML`, includes, extends, and `block.super` |
| Text generation | [`examples/multifile_text`](examples/multifile_text) | `FormatText` with loader-backed multi-file output |

Run any example with:

```bash
go run ./examples/<name>
```

## Security Model

- Use `NewDirLoader` for local directories when you want path validation and symlink-escape protection.
- Use `NewFSLoader` for already-sandboxed filesystems such as `embed.FS` or `testing/fstest.MapFS`.
- Use `FormatHTML` when rendering HTML; every `{{ expr }}` output is escaped unless the value is trusted.
- Use `SafeString` or the `safe` filter only for content you already trust.

## Documentation

- [docs/layout.md](docs/layout.md) — Includes, extends, blocks, raw blocks, and `block.super`
- [docs/loaders.md](docs/loaders.md) — Loader types and cache behavior
- [docs/filters.md](docs/filters.md) — Built-in filter reference
- [docs/variables.md](docs/variables.md) — Variable access and dot-path lookup
- [docs/control-structure.md](docs/control-structure.md) — `if`, `for`, `break`, and `continue`
- [docs/security.md](docs/security.md) — Loader boundaries and escaping rules
- [docs/liquid-compatibility.md](docs/liquid-compatibility.md) — Compatibility notes and differences

## Development

```bash
task test           # Run all tests with race detection
task test-coverage  # Generate coverage.out and coverage.html
task bench          # Run benchmarks
task fmt            # Run go fmt ./...
task vet            # Run go vet ./...
task lint           # Run golangci-lint and go mod tidy checks
task verify         # Run deps, fmt, vet, lint, and test
```

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for design boundaries, documentation expectations, and pull request guidance.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
