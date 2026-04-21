# template

## Project Overview

`github.com/kaptinlin/template` is a Go template engine with Django-style control flow, Liquid-compatible filter syntax, and optional loader-backed layout rendering.

Normal usage is engine-first: configure an `Engine` with `WithLoader(...)`, `WithFormat(...)`, and `WithLayout()` when needed, then parse source strings or render named templates. Advanced types such as `Loader`, `Template`, `RenderContext`, `Registry`, `TagRegistry`, `Statement`, and `Expression` stay public as extension points, not as separate product models.

For usage examples, installation, and user-facing guides, see [README.md](README.md).

## Commands

```bash
task test           # Run all tests with race detection
task test-coverage  # Generate coverage.out and coverage.html
task bench          # Run benchmarks
task fmt            # Run go fmt ./...
task vet            # Run go vet ./...
task lint           # Run golangci-lint and go mod tidy checks
task verify         # Run deps, fmt, vet, lint, and test
```

CI runs `task test` and `task lint` on pushes and pull requests to `main`.

## Architecture

```text
template/
├── engine.go        # Engine configuration, loader-backed rendering, cache, registry layering
├── template.go      # Compiled Template execution API
├── data.go          # Data, DataBuilder, and RenderContext
├── loader.go        # Loader interface and memory/dir/fs/chain implementations
├── filter_*.go      # Built-in filters
├── filters.go       # Filter registry and engine-local overrides
├── tag_*.go         # Built-in tags and layout features
├── tags.go          # Tag registry and feature-gated tag registration
├── docs/            # User-facing guides
└── examples/        # Runnable examples
```

Core execution flow:

```text
Source -> Lexer -> Parser -> AST -> Template.Execute
                                  ^
                     Engine.Load / Engine.Render compile and cache named templates here
```

Key invariants:

- `Engine` is the primary entry point for normal use.
- Loader-backed templates are cached by loader-resolved name.
- Engine-local tag and filter registries layer on top of built-in registries.
- `FeatureLayout` gates `include`, `extends`, `block`, `raw`, and `safe` behavior.
- `FormatHTML` enables auto-escape for `{{ expr }}` output; `FormatText` does not.

## Agent Workflow

### Implementation Phase — Find 2 References First

Before writing implementation code, find at least 2 relevant references in [`.references/`](.references/) and study their API and syntax decisions first. If the relevant reference slots are empty, ask the user before inventing a new pattern.

## References Index

Reference directories in [`.references/`](.references/):

| Category | Path | Purpose |
|---|---|---|
| Jinja | [`.references/jinja/`](.references/jinja/) | Reference slot for Jinja-compatible syntax and semantics |
| Liquid | [`.references/liquid/`](.references/liquid/) | Reference slot for Liquid filter and template behavior |
| Pongo2 | [`.references/pongo2/`](.references/pongo2/) | Reference slot for Django-style Go template behavior |

## Design Philosophy

- **KISS** — Day-to-day usage should start with one mental model: configure an `Engine`, then parse or render.
- **OCP** — Extend behavior through loaders, filters, tags, statements, and expressions instead of forking the core execution path.
- **Precision over cleverness** — HTML escaping, safe output, and loader boundaries must stay explicit and visible in the API.
- **Errors as teachers** — Lexer, parser, loader, and runtime failures should point to the exact category of mistake with position info or sentinel errors.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.

## API Design Principles

- **Progressive Disclosure** — Most callers should only need `Engine`, `Data`, `WithFormat(...)`, `WithLoader(...)`, and `WithLayout()`. Lower-level registries and execution types exist for advanced integrations.

## Coding Rules

### Must Follow

- Go 1.26.2 — use modern stdlib features already present here when they simplify code.
- Follow Google Go Best Practices: https://google.github.io/go-style/best-practices
- Follow Google Go Style Decisions: https://google.github.io/go-style/decisions
- Keep the public model engine-first — prefer `Engine`, `Format`, and `FeatureLayout` over adding parallel entry points for the same use case.
- Keep optional behavior engine-local — new feature tags and filters belong in engine-local registries unless they are true language defaults.
- Use sentinel errors from `errors.go` and wrap additional context with `fmt.Errorf("%w: ...")`.
- Preserve loader safety — validate every template name and keep `NewDirLoader` sandbox behavior intact.
- Use `t.Parallel()` in tests and `b.Loop()` in benchmarks to match repository conventions.

### Forbidden

- No `panic` in production code — return `error` instead.
- No new mutable package-global configuration surface for tags or filters — engine-local registration is the default extension path.
- No weakening of template-name validation, include or extends depth limits, or HTML escaping semantics.
- No premature abstraction — three similar lines are better than a helper used once.
- No feature creep — implement only the behavior required now.
- No working around dependency bugs — if a dependency blocks progress, create a report in `reports/` instead of reimplementing it inline.

## Dependency Issue Reporting

When you encounter a bug, limitation, or unexpected behavior in a dependency library:

1. Do not work around it by reimplementing the dependency's functionality.
2. Do not skip the dependency and quietly add a local replacement.
3. Create a report file at `reports/<dependency-name>.md`.
4. Include the dependency name and version, the trigger scenario, expected vs actual behavior, relevant errors, and any non-code workaround suggestion.
5. Continue with other work that does not depend on the broken behavior.

## Testing

- Tests use the standard library `testing` package with table-driven subtests and `t.Parallel()`.
- Loader and sandbox behavior are covered in `loader_test.go` and `security_test.go`.
- Benchmarks live in `benchmark_test.go` and use `b.Loop()`.
- Example-based API checks may live in `*_test.go` files and should exercise the public API directly.

```bash
go test -race ./...                  # Full test suite
go test -run '^Example' ./...        # Example-based API checks
go test -run '^TestDirLoader' ./...  # Targeted loader tests
go test -bench=. -benchmem ./...     # Benchmarks
```

## Dependencies

| Dependency | Purpose |
|---|---|
| `github.com/kaptinlin/filter` | Shared implementations for built-in string, array, math, date, and number filters |
| `github.com/kaptinlin/jsonpointer` | Dot-path access for `Data` and render context lookups |
| `github.com/go-json-experiment/json` | JSON-backed data shaping and JSON-oriented filter behavior |

## Error Handling

- Sentinel errors are defined in `errors.go` for loader, parser, filter, context, and runtime failures.
- Lexer and parser failures include line and column information.
- Match sentinel errors with `errors.Is` in tests and callers.

## Performance

- `Engine` caches compiled loader-backed templates by resolved name.
- Benchmarks in `benchmark_test.go` cover parsing, execution, caching, data access, and filter application.

## Linting

- `task lint` runs `golangci-lint` plus a `go mod tidy` diff check.
- Lint configuration lives in `.golangci.yml`.

## CI

- `.github/workflows/ci.yml` runs `task test` and `task lint` on `main` and pull requests.

## Agent Skills

Specialized skills in [`.agents/skills/`](.agents/skills/):

| Skill | When to Use |
|---|---|
| [`go-template-templating`](.agents/skills/go-template-templating/) | Adding or reviewing filters, tags, loaders, or template-engine examples |
| [`agent-md-writing`](.agents/skills/agent-md-writing/) | Regenerating `CLAUDE.md` and `AGENTS.md` from the current codebase |
| [`readme-writing`](.agents/skills/readme-writing/) | Refreshing `README.md` examples and user-facing documentation |
| [`golangci-linting`](.agents/skills/golangci-linting/) | Running or tuning golangci-lint and fixing lint failures |
| [`go-best-practices`](.agents/skills/go-best-practices/) | Applying repository-aligned Go style and testing practices |
| [`committing`](.agents/skills/committing/) | Creating a conventional commit after lint and tests are green |
