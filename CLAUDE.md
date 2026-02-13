# CLAUDE.md

## Project Overview

A lightweight Go template engine (`github.com/kaptinlin/template`) with Liquid/Django-style syntax. Supports variable interpolation, filters, conditionals (`if/elif/else`), loops (`for/break/continue`), and comments. Go 1.25+.

## Build & Test Commands

```sh
make test          # Run all tests (96.3% coverage)
make lint          # Run golangci-lint + go mod tidy check
make verify        # deps + fmt + vet + lint + test (full CI pipeline)
make fmt           # Format code
make vet           # Run go vet
go test ./...      # Run tests directly
go test -cover ./. # Run tests with coverage
```

CI runs `make test` and `make lint` on push/PR to main.

## Code Style

- **Google Go style guide**: imports grouped (stdlib, external, internal), doc comments start with function name, table-driven subtests
- **golangci-lint v2.7.2**: errcheck, govet, staticcheck, revive, gosec, exhaustive, err113, errorlint, and more (see `.golangci.yml`)
- **Formatters**: gofmt, goimports, gci
- Receiver names: short, consistent (e.g., `p` for Parser, `l` for Lexer, `v` for Value)
- Sentinel errors with `fmt.Errorf` wrapping: `fmt.Errorf("%w: %s", ErrSomething, detail)`
- All errors are defined as sentinel `var` in `errors.go`

## Architecture

```
Source → Lexer (lexer.go) → Tokens → Parser (parser.go) → AST → Execute (nodes.go) → Output
```

### Key Files

| File | Purpose |
|------|---------|
| `compile.go` | `Compile()` function (Lex + Parse + Template) |
| `render.go` | `Render()` shorthand |
| `template.go` | `Template` type, `Execute()`, `Render()` |
| `lexer.go` | Tokenizer, `LexerError` with line:col |
| `token.go` | Token types and keywords |
| `parser.go` | Main parser, `ParseUntil`, `ParseUntilWithArgs` |
| `parser_helpers.go` | `ParseError`, helper methods (Match, ExpectIdentifier, etc.) |
| `expr.go` | Expression parser with operator precedence |
| `nodes.go` | All AST node types, `Statement`/`Expression` interfaces |
| `value.go` | `Value` type wrapper for expression results |
| `context.go` | `Context`, `ContextBuilder`, `ExecutionContext` |
| `errors.go` | All sentinel errors |
| `filters.go` | Filter registry (`Registry`, thread-safe) |
| `tags.go` | Tag registry (thread-safe) |
| `filter_*.go` | Built-in filter implementations (string, math, array, map, date, number, format) |
| `tag_*.go` | Built-in tag parsers (if, for, break, continue) |
| `utils.go` | Internal helpers (toString, toInteger) |

### Conventions for Adding Features

- **New filter**: Add to the appropriate `filter_*.go` file, register in its `register*Filters()` function. `FilterFunc` signature: `func(value any, args ...string) (any, error)`
- **New tag**: Create `tag_<name>.go`, add to `builtinTags` slice in `tags.go`. `TagParser` signature: `func(doc *Parser, start *Token, arguments *Parser) (Statement, error)`
- **New error**: Add to `errors.go` with doc comment, wrap with `fmt.Errorf("%w: ...", ErrName, ...)`
- **Tests**: Same file as source with `_test.go` suffix, use table-driven subtests

### Interfaces (Open for External Implementation)

- `Statement` — `Execute(ctx *ExecutionContext, w io.Writer) error` + `Position()` + `String()`
- `Expression` — `Evaluate(ctx *ExecutionContext) (*Value, error)` + `Position()` + `String()`

Both are intentionally open (no unexported methods) so external packages can implement custom tags/nodes.

## Commit Style

Conventional commits: `type: description`

Types: `feat`, `fix`, `refactor`, `style`, `test`, `chore`, `build`, `docs`

Examples from history:
- `test: add comprehensive edge case tests across all modules`
- `style: apply Google Go formatting rules across all modules`
- `refactor: rewrite template engine with dedicated lexer, parser, AST nodes`
