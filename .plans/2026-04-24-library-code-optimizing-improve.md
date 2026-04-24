# library-code-optimizing — examples pass 2

## Baseline
- `task test`: green (`go test -race ./...`)
- `task lint`: green (`golangci-lint run --timeout=10m --path-prefix .`, 0 issues)
- `.cov.before` total statements: 92.1%
- Reference slots checked: `.references/jinja`, `.references/liquid`, `.references/pongo2` are present but empty; no template syntax or semantic changes in this pass.

## Package order
1. examples/usage
2. examples/custom_filters
3. examples/custom_tags

## Per-package items
### examples/usage
- Restating comments to delete: `main.go:33` expected output comment after `return render(...)` is covered by runnable example tests.

### examples/custom_filters
- Other simplifying / go-best-practices targets: `main.go:29`, `main.go:32` uses `fmt.Sprintf("%v", ...)` where `fmt.Sprint(...)` expresses the same conversion directly.
- Restating comments to delete: `main.go:48` expected output comment is covered by runnable example tests.

### examples/custom_tags
- Restating comments to delete: `main.go:64`, `main.go:59` describe exactly what the registration/output code does and are covered by runnable example tests.

## Exported-symbol exclusions
- `examples/custom_tags.SetNode`: exported example extension type, retained as public API demonstration.
