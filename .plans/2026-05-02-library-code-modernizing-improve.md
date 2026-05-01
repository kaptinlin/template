# library-code-modernizing — repos/template

## Baseline
- `go test ./...`: green.
- `task lint`: green, 0 golangci-lint issues, tidy diff clean.
- `go test -coverprofile=.cov.before ./...`: green; root package coverage 92.5%, examples 88.9%–100.0%.
- `go fix -diff ./... > .fix.before.diff || true`: no hunks.

## Package order
1. `.`

## Per-package items
### `.`
- Category swaps: none. The remaining non-test Go files already use applicable modern Go 1.20–1.26 idioms where behavior-preserving.
- Judgment-required deferrals:
  - `engine.go` duplicate-load coordination remains custom; replacing it would require a new dependency or behavior-sensitive timing changes.
  - `filters.go` and `tags.go` nil-registration panics remain unchanged because converting them to returned errors would alter public API contracts.
  - `parser_helpers.go` statement filtering remains a direct loop; iterator conversion would be speculative.
  - `value.go` deterministic map rendering still collects keys; reflection iterators do not remove the ordering requirement.
  - `value.go` unreachable `reflect.Invalid` cleanup belongs to the separate `library-code-optimizing` pass and is excluded from this modernization pass.

## Public-API-affecting items
- None.
