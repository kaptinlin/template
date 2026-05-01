# library-code-optimizing — repos/template pass 2

## Baseline
- `task test`: green (`go test -race ./...`)
- `task lint`: green, 0 issues
- `.cov.before`: total statements 92.3%

## Package order
1. `.`

## Per-package items
### `.`
- Dead unexported branches: `tag_if.go:32` | `hasElse` | unreachable duplicate-else guards after `ParseUntilWithArgs("endif")`
- Single-caller helpers to inline: `compile.go:40` | `compileForEngine` | call site `engine.go:243`; update stale internal reference in `tag_extends.go:42`
- Single-caller helpers to inline: `filters.go:32` | `(*Registry).validate` | call site `filters.go:51`
- Single-purpose helper to inline: `engine.go:166` | `(*Engine).autoescape` | call sites `engine.go:372`, `template.go:70`
- Restating comments to delete: `filter_date.go:5`, `filter_date.go:20`, `filter_date.go:30`, `filter_date.go:35`, `filter_date.go:40`, `filter_date.go:45`, `filter_date.go:50`, `filter_date.go:55`, `filter_date.go:60`
- Restating comments to delete: `filter_number.go:9`, `filter_number.go:15`, `filter_number.go:23`

## Exported-symbol exclusions
- None.

## Drift items appended 2026-05-02
- Baseline update: `go test ./...`, `task lint`, and `go test -coverprofile=.cov.before ./...` passed; total statements 92.3%.
- Dead unexported branch: `value.go:69` | `(*Value).IsTrue` | remove unreachable `reflect.Invalid` switch case because invalid reflected values return before the switch.
