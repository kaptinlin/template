# Contributing to `template`

Thanks for contributing to `github.com/kaptinlin/template`.

This project is intentionally small on the surface and strict about API focus. The main rule is simple: keep the public model centered on `Engine`, `Format`, and `Feature`.

## Development Setup

Common commands:

```sh
task test
task lint
task verify
go test ./...
```

## Design Rules

Before opening a PR, check these invariants:

1. `Engine` must remain the only public entry point.
2. Layout features must stay behind `FeatureLayout`.
3. HTML auto-escape must remain opt-in through `FormatHTML`.
4. New tags and filters must belong to an engine instance, not mutable package-global APIs.
5. Cache identity must continue to be based on loader `resolved` names.

If a change widens the public surface area, call that out explicitly in the PR and justify why the existing engine model is not enough.

## Feature Checklist

For any new tag, filter, loader behavior, or execution change, answer these questions before merging:

1. Does this fit naturally inside `Engine`, `Format`, or `Feature`?
2. Should this be global, or should it live only inside an engine feature?
3. Does this affect HTML escaping or `SafeString` semantics?
4. Does this need a new sentinel error or better wrapped context?
5. Should registration use `Register`, `Replace`, or `MustRegister`?
6. What regression test proves the design boundary still holds?

## Code Style

Follow the conventions already used in the repository:

- Keep package APIs small and explicit.
- Prefer sentinel errors wrapped with context.
- Use table-driven tests with clear case names.
- Keep exported docs short and direct.
- Preserve the separation between built-in registries and engine-local registries.
- Prefer explicit registry intent: `Register` for new names, `Replace` for overrides, `MustRegister` for bootstrap code.

## Documentation

Documentation changes are first-class contributions.

When you add or change behavior:

- Update `README.md` if the user-facing entry points changed.
- Update package docs in `doc.go` if the public contract changed.
- Update focused docs in `docs/` for syntax, security, or loader behavior.
- Update `ARCHITECTURE.md` when cache identity, execution context inheritance, or execution flow changes.

Prefer explaining when to use a feature and when not to use it. This library aims to make the correct path obvious.

## Pull Requests

When submitting a PR:

1. Describe the user-visible change in one short paragraph.
2. Call out whether `Engine`, `Format`, or `FeatureLayout` behavior is affected.
3. Mention the tests you added or updated.
4. Mention any security or escaping implications.
5. Mention whether cache identity, loop scope, or child-context inheritance changed.

Conventional commit style is used in history:

- `feat: description`
- `fix: description`
- `refactor: description`
- `test: description`
- `docs: description`

## Questions

If something is unclear, open an issue or draft PR with the design question first. For this project, preserving a small and coherent engine model is usually more important than adding one more feature quickly.
