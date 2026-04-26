# library-code-optimizing — template pass 2

## Baseline
- `go test ./...`: passed
- `golangci-lint run`: 0 issues
- `go test -coverprofile=.cov.before ./...`: passed
- `go tool cover -func=.cov.before`: total statements 92.3%

## Package order
1. `.`

## Per-package items
### `.`
- Dead unexported fields: `loader.go:84` | `dirLoader.dir` | remove unused field and constructor assignment.
- Dead unexported fields: `tag_include.go:37` | `IncludeNode.lazy` | remove unused field and assignments; behavior is already determined by `prepared == nil` and `pathExpr != nil`.

## Exported-symbol exclusions
- `DataBuilder`, loader constructors, token helpers, and exported sentinel errors reported by deadcode are public API and out of scope for this pass.
