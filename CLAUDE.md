# CLAUDE.md

## Project Overview

A Go template engine at `github.com/kaptinlin/template`. Django-inspired control flow + Liquid-compatible filter syntax. Two modes coexist in one package:

- **Single-string mode** — `Compile(src)` / `Render(src, ctx)`. Minimal feature set, same behavior since the library was born.
- **Multi-file mode** — `NewHTMLSet(loader)` / `NewTextSet(loader)`. Adds `include` / `extends` / `block` / `{{ block.super }}` / `raw` / `safe` / auto-HTML-escape. All of these live in Set-scoped registries, so the single-string mode sees **none of them**.

This split mirrors Go's own `text/template` vs `html/template`. Touch the single-string path with great caution — it's the backward-compat contract.

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
- Receiver names: short and consistent (`p` for Parser, `l` for Lexer, `v` for Value, `s` for Set, `n` for Node types, `ec` for ExecutionContext)
- Sentinel errors wrapped with `fmt.Errorf("%w: %s", ErrSomething, detail)`
- All errors defined as sentinel `var` in `errors.go`

## Architecture

```
Source → Lexer → Tokens → Parser → AST → Template.Execute → Output
                                       ↑
                   Set.Get caches Templates here and
                   threads its private tag/filter
                   registries + Loader through
```

### File map

| File | Purpose |
|---|---|
| `compile.go` | `Compile(src)` + internal `compileForSet(src, set)` |
| `render.go` | `Render(src, ctx)` shorthand |
| `template.go` | `Template` struct, `Execute` with extends-chain walk, `executeRoot` |
| `lexer.go` | Tokenizer; `allowRaw` opt-in for `{% raw %}` block |
| `token.go` | Token types and keywords |
| `parser.go` | Main parser; `Parser.set`/`parent`/`blocks`/`hasNonTrivialContent` fields |
| `parser_helpers.go` | `ParseError`, Match/Expect helpers |
| `expr.go` | Expression parser |
| `nodes.go` | AST nodes; `OutputNode.Execute` auto-escape path; `FilterNode.Evaluate` set-aware lookup |
| `value.go` | `Value` wrapper |
| `context.go` | `Context`, `ExecutionContext` (with `set`/`autoescape`/`includeDepth`/`currentLeaf`) |
| `errors.go` | All sentinel errors |
| `filters.go` | `Registry` (with `parent` fallback), global `defaultRegistry` |
| `filter_string.go` | Built-in string filters; `escape`/`escape_once` (global `string`) + `escapeFilterSafe`/`escapeOnceFilterSafe` (HTMLSet `SafeString`) |
| `filter_*.go` | Other built-in filters (math, array, map, date, number, format) |
| `tags.go` | `TagRegistry` (with `parent` fallback); `builtinTags` (4) vs `layoutTags` (3) |
| `tag_if.go`/`tag_for.go`/`tag_break.go`/`tag_continue.go` | Global built-in tag parsers |
| `tag_include.go` | `{% include %}` parser + `IncludeNode`; set-scoped |
| `tag_extends.go` | `{% extends %}` parser + `ExtendsNode`; set-scoped; first-tag constraint |
| `tag_block.go` | `{% block %}` parser + `BlockNode`; set-scoped; `block.super` chain |
| `safe.go` | `SafeString` type + `safeFilter` (set-scoped only) |
| `loader.go` | `Loader` interface + `ValidateName` + `memoryLoader`/`dirLoader` (os.Root)/`fsLoader`/`chainLoader` |
| `set.go` | `Set` struct + `NewHTMLSet`/`NewTextSet` + cache + parsing-set + `Reset` + `WithGlobals`/`WithFilters`/`WithTags` |
| `utils.go` | `toString`, `toInteger` internal helpers |

### The Set-private registry pattern

Both `*TagRegistry` and `*Registry` (filter) carry an optional `parent` field. When a lookup misses locally, the parent is consulted. `Set` constructs its own private registry with `parent = defaultXxxRegistry`, then registers layout tags / safe filters into the private layer only.

**This is the key invariant**: layout features live exclusively in Set-private layers. The global registries are never touched by Set construction, so `Compile(src)` continues to see only the pre-layout world.

- `parser.go:parseTag()` consults `p.set.tags` first, then `defaultTagRegistry`
- `nodes.go:FilterNode.Evaluate()` consults `ctx.set.filters` first, then global
- `lexer.go`'s `allowRaw` is set by `compileForSet` only when `set != nil`

### Two execution paths

| Path | `Compile(src)` / `Render(src)` | `NewXxxSet(loader).Render(...)` |
|---|---|---|
| Enters via | `parser.go:parseTag()` with `p.set == nil` | `parser.go:parseTag()` with `p.set != nil` |
| Raw lexer | Off (`allowRaw = false`) | On |
| Available tags | `if`/`for`/`break`/`continue` | + `include`/`extends`/`block` |
| Available filters | Global `defaultRegistry` only | + `safe` (always) + `escape`/`h` overrides (HTMLSet) |
| `OutputNode.Execute` escape | `ctx.autoescape = false` → no escape | HTMLSet: on; TextSet: off |
| `escape` filter return | `string` | `SafeString` (HTMLSet only) |

If you add a feature that could affect the single-string path, ask "does this leak through `Compile(src)`?" The answer MUST be "no" unless the feature is explicitly in `builtinTags` / `defaultRegistry`.

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

### New layout tag (Set-scoped)
Create `tag_<name>.go`, add to `layoutTags` slice in `tags.go`. The parser can access the owning `*Set` via `args.Set()` to resolve referenced templates at parse time.

### New error
Add to `errors.go` with a doc comment, wrap at use-sites with `fmt.Errorf("%w: ...", ErrName, ...)`.

### Tests
Same file as source with `_test.go` suffix, table-driven subtests with `t.Parallel()`, direct `if got != want` assertions, `errors.Is` for error matching.

**Regression guard**: `compile_compat_test.go` asserts that `Compile(src)` sees zero layout tags and that the `escape` filter returns plain `string`. Do not weaken these tests — they are the backward-compat contract.

## Interfaces open for external implementation

- `Statement` — `Execute(ctx *ExecutionContext, w io.Writer) error` + `Position()` + `String()`
- `Expression` — `Evaluate(ctx *ExecutionContext) (*Value, error)` + `Position()` + `String()`
- `Loader` — `Open(name string) (source, resolved string, err error)`

All three are intentionally open so external packages can implement custom tags, nodes, or loaders.

## Commit style

Conventional commits: `type: description`

Types: `feat`, `fix`, `refactor`, `style`, `test`, `chore`, `build`, `docs`

Examples from history:
- `feat: add multi-file template system (NewHTMLSet, NewTextSet, Loader)`
- `refactor: layer layout tags in Set-private registry for backward compat`
- `test: add security matrix for path traversal and symlink escape`
