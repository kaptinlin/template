# library-code-simplifying — repos/template

## Baseline
- `task --dir repos/template test`: passed.
- `task --dir repos/template lint`: passed with 0 issues.
- `go -C repos/template test -coverprofile=.cov.before ./...`: passed.
- `.cov.before` total statements: 92.4%.

## Package order
1. `.`

## Per-package items
### `.`
- Nesting / early-return targets: [`tag_include.go:120` | default to inherited child context and override for `only` to remove the `else` branch]
- Duplicate-logic consolidations: none retained after evaluation.
- Restating / outdated comments to delete: [`tag_for.go:10`, `tag_for.go:17`, `tag_for.go:26`, `tag_for.go:32`, `tag_for.go:41`, `tag_if.go:12`, `tag_if.go:21`, `tag_if.go:31`, `value.go:91`]
- Name-clarity improvements (internal only): [`parser.go:153` | `fn` → `tagParser`, `ok` → `found`; `nodes.go:677` | `fn` → `filterFn`, `ok` → `found`]
