# Architecture

The durable architecture contract lives in
[SPECS/01-engine-contract.md](SPECS/01-engine-contract.md).

This file is intentionally short. The project no longer treats lexer, parser,
token stream, AST nodes, render context, value wrappers, or tag registries as
public product concepts. They are implementation machinery behind the engine.

## System Shape

```text
source string or loader name
        |
        v
      Engine
        |
        +-- compile source into immutable Template
        |
        +-- cache loader-backed templates by loader-resolved identity
        |
        +-- execute Template with Data and engine-local output semantics
```

## Public Center

Normal use starts with `Engine`:

- `New(...)`
- `WithLoader(...)`
- `WithFormat(...)`
- `WithLayout()`
- `WithDefaults(...)`
- `WithFilter(...)`
- `Engine.ParseString(...)`
- `Engine.Render(...)`
- `Template.Render(...)`

The stable extension points are custom loaders and engine-local filters.
Custom tag authoring is not public.

## Protected Decisions

- Loader-backed templates are cached and de-duplicated by loader-resolved name,
  not by caller-provided name.
- Compiled templates bind filter functions at compile time.
- Filter mutation is rejected after an engine has compiled or loaded a template;
  use `Engine.Clone(...)` for a fresh configurable engine.
- `FeatureLayout` gates layout-only syntax: `include`, `extends`, `block`,
  `raw`, and `safe`.
- `FormatHTML` enables final output escaping for `{{ expr }}`; `FormatText`
  does not.
- `SafeHTML` is trusted only at the HTML output sink. Ordinary filters downgrade
  trust by returning ordinary values.
- Missing fields and map misses evaluate as nil-like values in templates:
  empty output, false in conditionals, and eligible for `default`.
- `Data.Get` follows the same object-shaped lookup law as template lookup while
  preserving its sentinel error contract.

## File Map

```text
engine.go      Engine lifecycle, loader-backed rendering, cache, filter sealing
template.go    Immutable compiled template execution
data.go        Data, DataBuilder, Data.Get, internal render context
loader.go      Loader interface and built-in loader implementations
value.go       Internal value coercion, lookup, truthiness, iteration
filters.go     Internal filter registry and engine-local layering
tags.go        Internal tag registry and feature-gated built-in tags
safe.go        SafeHTML trust marker
errors.go      Sentinel errors
SPECS/         Durable architecture and behavior contracts
docs/          User-facing guides
examples/      Runnable usage examples
```

## Verification

The architecture contract is protected by:

- resolved-identity loader tests for cache, include recursion, and extends
  cycles
- engine lifecycle tests for compile-time filter binding
- public API checks that keep internals out of package godoc
- value-law tests across template output, conditionals, loops, filters, and
  `Data.Get`
- HTML trust tests covering escaping, non-double-escaping, and ordinary filter
  downgrade
- `task test` and `task lint`
