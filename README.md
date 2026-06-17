# template

[![Go Version](https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A Go template engine with Django-style control flow, Liquid-style filters, and optional layout-aware HTML rendering.

For development guidelines and project conventions, see [CLAUDE.md](CLAUDE.md).

## Features

- **One entry point**: Configure a single `Engine` for source-string parsing or loader-backed rendering.
- **Two output modes**: `FormatText` for raw text, `FormatHTML` for automatic escaping of `{{ expr }}` output.
- **Layout on demand**: `WithLayout()` enables `include`, `extends`, `block`, `raw`, and `safe`.
- **Composable loaders**: `NewMemoryLoader`, `NewDirLoader`, `NewFSLoader`, and `NewChainLoader` for tests, embedded assets, and override layers.
- **Sandboxed disk reads**: `NewDirLoader` uses `os.Root` so template lookups cannot escape the configured directory.
- **Typed diagnostics**: `*RenderError` carries the failing template name, line, column, and a sentinel cause for `errors.Is` / `errors.As`.
- **Engine-local extensions**: Register filters on a single engine and provide custom loaders without touching global state.

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
	tmpl, err := engine.ParseString("Hello, {{ name | upcase }}!")
	if err != nil {
		log.Fatal(err)
	}

	out, err := tmpl.Render(template.Data{"name": "alice"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out) // Hello, ALICE!
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
| Extend behavior | `WithFilter(...)` | Scoped to the engine instance |

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

For runnable programs, see the [examples](#examples) below.

## Errors and Diagnostics

Parse-time failures return `*ParseError`. Render-time failures return
`*RenderError`, which carries the failing template name, 1-based line and
column, and the underlying sentinel.

```go
out, err := engine.Render("page.html", data)
if err != nil {
	var re *template.RenderError
	if errors.As(err, &re) {
		log.Printf("%s:%d:%d: %v", re.Template, re.Line, re.Col, re.Cause)
	}
	if errors.Is(err, template.ErrFilterNotFound) {
		// branch on the sentinel category
	}
	return err
}
```

Sentinels live in [`errors.go`](errors.go) and are matched with `errors.Is`.
The human-readable `RenderError.Error()` format is not part of the contract;
read the fields for stable output.

## Extension Points

| Need | API |
|---|---|
| Custom filter | `WithFilter(...)` at construction; `Engine.RegisterFilter` / `Engine.ReplaceFilter` before first compile |
| Custom loader | Implement `Loader` and pass it through `WithLoader(...)` |
| Trusted HTML | `SafeHTML` or `| safe` under `FormatHTML` |

After an engine starts compiling templates, `RegisterFilter` and
`ReplaceFilter` return `ErrEngineCompiled`. Use `Clone(...)` to derive a fresh
configurable engine with an empty cache. Nil filter functions are reported as
configuration errors instead of panicking.

## Examples

Runnable examples live in [`examples/`](examples/):

| Example | Path | What it shows |
|---|---|---|
| Basic usage | [`examples/usage`](examples/usage) | Parse and render a source string through `Engine` |
| Custom filters | [`examples/custom_filters`](examples/custom_filters) | Configure an engine-local filter |
| HTML layouts | [`examples/layout`](examples/layout) | `WithLayout()`, `FormatHTML`, includes, extends, and `block.super` |
| Text generation | [`examples/multifile_text`](examples/multifile_text) | `FormatText` with loader-backed multi-file output |
| Secret redaction | [`examples/secret_redaction`](examples/secret_redaction) | Redact rendered output at the writer or filter boundary |

Run any example with:

```bash
go run ./examples/<name>
```

## Documentation

- [docs/layout.md](docs/layout.md) — Includes, extends, blocks, raw blocks, and `block.super`
- [docs/loaders.md](docs/loaders.md) — Loader types and cache behavior
- [docs/filters.md](docs/filters.md) — Built-in filter reference
- [docs/variables.md](docs/variables.md) — Variable access, dot paths, and strict mode
- [docs/control-structure.md](docs/control-structure.md) — `if`, `for`, `break`, and `continue`
- [docs/security.md](docs/security.md) — Loader sandbox, escaping rules, and redaction
- [docs/liquid-compatibility.md](docs/liquid-compatibility.md) — Compatibility notes and differences
- [SPECS/01-engine-contract.md](SPECS/01-engine-contract.md) — Durable engine, loader, value, and HTML trust contract

## Development

```bash
task test           # Run all tests with race detection
task test-coverage  # Generate coverage.out and coverage.html
task bench          # Run benchmarks
task fmt            # Run configured Go formatters
task vet            # Run go vet ./...
task lint           # Run golangci-lint and go mod tidy checks
task verify         # Run deps, fmt, vet, lint, and test
```

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for design boundaries, documentation expectations, and pull request guidance.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
