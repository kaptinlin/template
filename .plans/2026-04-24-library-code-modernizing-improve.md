# library-code-modernizing — template

## Baseline
- `go test ./...`: green.
- `task lint`: green, 0 golangci-lint issues, tidy diff clean.
- `.cov.before`: captured with all packages green.
- `go fix -diff ./...`: one hunk, `data.go` `strings.Split` iteration to `strings.SplitSeq`.

## Package order
1. `github.com/kaptinlin/template`

## Per-package items
### `github.com/kaptinlin/template`
- Category swaps: `filters.go:113` | `NewRegistry` + `maps.Copy` -> `maps.Clone` in `Registry.Clone`.
- Category swaps: `tags.go:119` | `NewTagRegistry` + `maps.Copy` -> `maps.Clone` in `TagRegistry.Clone`.
- Category swaps: `data.go:334` | `strings.Split` ranged slice -> `strings.SplitSeq`.
- Judgment-required deferrals: `engine.go:377` | keep map reallocation in `Reset`; `clear` would retain backing storage and change memory behavior.
- Judgment-required deferrals: `data.go:141` | keep `strings.Split`; callers need a `[]string` for validation and variadic jsonpointer lookup.

## Public-API-affecting items
- None.
