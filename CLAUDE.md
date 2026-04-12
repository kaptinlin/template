# CLAUDE.md

## Project Overview

A Go template engine at `github.com/kaptinlin/template`. Django-inspired control flow + Liquid-compatible filter syntax. The public design is centered on one concept:

- **Engine** — `New(...)` with `WithLoader(...)`, `WithFormat(...)`, and optional `WithLayout()`. It handles source parsing, named-template loading, `include` / `extends` / `block` / `{{ block.super }}` / `raw` / `safe`, and HTML auto-escape.

Avoid rebuilding parallel entry points. New behavior should slot into `Engine`, `Format`, or `Feature`.

Go 1.26+.

## Build & Test Commands

```sh
task test          # go test -race ./...
task lint          # golangci-lint + go mod tidy check
task verify        # deps + fmt + vet + lint + test
go test ./...      # direct test run
go test -cover ./...
```

CI runs `task test` and `task lint` on push/PR to main.

## Code Style

- **Google Go style guide**: imports grouped (stdlib, external, internal), doc comments start with function name, table-driven subtests
- **golangci-lint**: errcheck, govet, staticcheck, revive, gosec, exhaustive, err113, errorlint, gci (see `.golangci.yml`)
- **Formatters**: gofmt, goimports, gci
- Receiver names: short and consistent (`p` for Parser, `l` for Lexer, `v` for Value, `e` for Engine, `n` for Node types, `ec` for RenderContext)
- Sentinel errors wrapped with `fmt.Errorf("%w: %s", ErrSomething, detail)`
- All errors defined as sentinel `var` in `errors.go`

## Architecture

```
Source → Lexer → Tokens → Parser → AST → Template.Execute → Output
                                       ↑
                   Engine.Load caches Templates here and
                   threads its private tag/filter
                   registries + Loader through
```

### File map

| File | Purpose |
|---|---|
| `compile.go` | Internal `compileForEngine(src, engine)` |
| `template.go` | `Template` struct, `Execute` with extends-chain walk, `executeRoot` |
| `lexer.go` | Tokenizer; `allowRaw` opt-in for `{% raw %}` block |
| `token.go` | Token types and keywords |
| `parser.go` | Main parser; `Parser.engine`/`parent`/`blocks`/`hasNonTrivialContent` fields |
| `parser_helpers.go` | `ParseError`, Match/Expect helpers |
| `expr.go` | Expression parser |
| `nodes.go` | AST nodes; `OutputNode.Execute` auto-escape path; `FilterNode.Evaluate` engine-aware lookup |
| `value.go` | `Value` wrapper |
| `data.go` | `Data`, `RenderContext` (with `engine`/`autoescape`/`includeDepth`/`currentLeaf`) |
| `errors.go` | All sentinel errors |
| `filters.go` | `Registry` (with `parent` fallback), global `defaultRegistry` |
| `filter_string.go` | Built-in string filters; `escape`/`escape_once` (global `string`) + `escapeFilterSafe`/`escapeOnceFilterSafe` (`FormatHTML` `SafeString`) |
| `filter_*.go` | Other built-in filters (math, array, map, date, number, format) |
| `tags.go` | `TagRegistry` (with `parent` fallback); `builtinTags` (4) vs `layoutTags` (3) |
| `tag_if.go`/`tag_for.go`/`tag_break.go`/`tag_continue.go` | Global built-in tag parsers |
| `tag_include.go` | `{% include %}` parser + `IncludeNode`; `FeatureLayout` only |
| `tag_extends.go` | `{% extends %}` parser + `ExtendsNode`; `FeatureLayout` only; first-tag constraint |
| `tag_block.go` | `{% block %}` parser + `BlockNode`; `FeatureLayout` only; `block.super` chain |
| `safe.go` | `SafeString` type + `safeFilter` (`FeatureLayout` only) |
| `loader.go` | `Loader` interface + `ValidateName` + `memoryLoader`/`dirLoader` (os.Root)/`fsLoader`/`chainLoader` |
| `engine.go` | `Engine` struct + `New` + `Format`/`Feature` + cache + parsing-set + `Reset` + `WithLoader`/`WithFormat`/`WithFeatures`/`WithDefaults`/`WithFilters`/`WithTags` |
| `utils.go` | `toString`, `toInteger` internal helpers |

### The Engine-private registry pattern

Both `*TagRegistry` and `*Registry` (filter) carry an optional `parent` field. When a lookup misses locally, the parent is consulted. `Engine` constructs its own private registry with `parent = defaultXxxRegistry`, then registers feature-gated tags / safe filters into the private layer only.

**This is the key invariant**: optional behavior lives in engine-private layers. Built-in registries provide the base language; features and overrides stay local to an engine instance.

- `parser.go:parseTag()` consults `p.engine.tags` first, then `defaultTagRegistry`
- `nodes.go:FilterNode.Evaluate()` consults `ctx.engine.filters` first, then global
- `lexer.go`'s `allowRaw` is set by `compileForEngine` only when `FeatureLayout` is enabled

### Engine execution model

| Concern | Behavior |
|---|---|
| Raw lexer | Enabled only when `FeatureLayout` is on |
| Available tags | Built-ins by default; `include`/`extends`/`block` gated by `FeatureLayout` |
| Available filters | Built-ins by default; `safe` gated by `FeatureLayout`; `escape`/`h` overrides gated by `FormatHTML` |
| `OutputNode.Execute` escape | `FormatHTML`: on; `FormatText`: off |
| `escape` filter return | `string` by default; `SafeString` in `FormatHTML` engines |

If you add a feature, ask "does this belong in `Engine`, `Format`, or `Feature`?" Avoid creating a second mental model.

## Adding features

### New global filter
Add to the appropriate `filter_*.go` file, register in its `register*Filters()` function. `FilterFunc` signature:

```go
func(value any, args ...any) (any, error)
```

### New global tag
Create `tag_<name>.go`, add to `builtinTags` slice in `tags.go`. **For global tags only — layout tags go in `layoutTags`.**

```go
func(doc *Parser, start *Token, arguments *Parser) (Statement, error)
```

### New layout tag (`FeatureLayout`)
Create `tag_<name>.go`, add to `layoutTags` slice in `tags.go`. The parser can access the owning `*Engine` via `args.Engine()` to resolve referenced templates at parse time.

### New error
Add to `errors.go` with a doc comment, wrap at use-sites with `fmt.Errorf("%w: ...", ErrName, ...)`.

### Tests
Same file as source with `_test.go` suffix, table-driven subtests with `t.Parallel()`, direct `if got != want` assertions, `errors.Is` for error matching.

Prefer assertions around `Engine`, `FormatHTML`, `FormatText`, and `FeatureLayout`. Test adapters may exist for legacy scenarios, but they are not design targets.

## Interfaces open for external implementation

- `Statement` — `Execute(ctx *RenderContext, w io.Writer) error` + `Position()` + `String()`
- `Expression` — `Evaluate(ctx *RenderContext) (*Value, error)` + `Position()` + `String()`
- `Loader` — `Open(name string) (source, resolved string, err error)`

All three are intentionally open so external packages can implement custom tags, nodes, or loaders.

## Commit style

Conventional commits: `type: description`

Types: `feat`, `fix`, `refactor`, `style`, `test`, `chore`, `build`, `docs`

Examples from history:
- `feat: add multi-file template system (Loader, layout tags, autoescape)`
- `refactor: unify template execution around Engine`
- `test: add security matrix for path traversal and symlink escape`
