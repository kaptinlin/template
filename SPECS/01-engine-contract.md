# Engine Contract

## Overview

This package is one template engine, not a collection of parallel product
models. The long-term contract is:

```text
New Engine.
Choose loader.
Choose output format.
Enable layout features only when needed.
Register engine-local filters before compiling.
Render.
Everything else is implementation.
```

The specification protects the concepts that must stay stable across future
implementation changes: engine lifecycle, loader identity, public surface,
layout feature gating, trusted HTML output, value semantics, and verification.

## Scenarios

- A caller renders a source string with a small data map and no loader.
- A caller renders named templates through a loader, with cache reuse and
  deterministic dependency handling.
- A caller enables layout syntax for includes, inheritance, raw blocks, block
  overrides, and trusted HTML marking.
- A caller chooses HTML output and expects user data to be escaped unless it is
  explicitly trusted.
- A caller registers a filter on one engine and expects other engines and
  already-compiled templates to keep their meaning.
- A caller uses `Data.Get` and template lookup on the same data shape and
  expects the two surfaces to agree where their error contracts allow it.

Out of scope:

- Untrusted-template sandboxing.
- Context-aware escaping for JavaScript, CSS, URLs, or JSON-in-HTML.
- Public custom tag authoring.
- Runtime mutation of compiled template meaning.
- Global mutable configuration as the normal extension path.

## Concept Model

### Engine

- **Definition**: The owner of loader-backed rendering state, output format,
  optional features, defaults, filter layers, template cache, and compile
  lifecycle.
- **Includes**: Source-string parsing, named template loading, rendering,
  cache ownership, feature gating, and engine-local filters.
- **Excludes**: Global product configuration, public tag/parser access, and
  caller-owned compile caches.
- **Lifecycle**: Construct with options, optionally register filters, compile
  or load templates, then render immutable templates. After the first compile
  or load, filter mutation is rejected.
- **Identity**: An engine instance plus its loader, format, feature set,
  defaults, filter layer, and cache.
- **Invariants**:
  - A compiled template's filter meaning does not change.
  - Loader-backed cache entries are keyed by resolved template identity.
  - Optional syntax appears only through explicit feature choices.
- **Owner / Stage Source**: `engine.go`, `template.go`, and this spec.

### Template

- **Definition**: An immutable compiled program produced by an engine.
- **Includes**: Source-string templates and loader-backed named templates.
- **Excludes**: Public AST, parser, token stream, or mutation hooks.
- **Lifecycle**: Created by compilation, then rendered any number of times.
- **Identity**: For loader-backed templates, the loader-resolved name; for
  source strings, the template value itself.
- **Invariants**:
  - Rendering must not mutate the compiled program.
  - Loader-backed templates carry enough engine context for includes,
    inheritance, and output format.
- **Owner / Stage Source**: `template.go`.

### Loader Identity

- **Definition**: The stable name returned by `Loader.Open` alongside template
  source.
- **Includes**: Prefixes or normalized names added by loaders to distinguish
  override layers and physical origins.
- **Excludes**: The caller-provided request name as a cache or cycle key.
- **Lifecycle**: Produced on open, then used for cache keys, in-flight compile
  de-duplication, parsing-stack checks, and dependency recursion decisions.
- **Identity**: Exact resolved string.
- **Invariants**:
  - Two request names resolving to the same identity share one compiled
    template.
  - Two different physical templates must not collide in the cache.
  - Include and inheritance dependency checks use resolved identity.
- **Owner / Stage Source**: `loader.go` and `engine.go`.

### Feature Layout

- **Definition**: The optional language feature set for multi-template layout
  rendering.
- **Includes**: `include`, `extends`, `block`, `raw`, `safe`, and block super.
- **Excludes**: Core conditionals, loops, ordinary filters, and source-string
  parsing without layout enabled.
- **Lifecycle**: Enabled at engine construction through `WithLayout()` or the
  matching feature option.
- **Invariants**:
  - Layout-only syntax is unavailable unless enabled.
  - Include recursion is bounded.
  - Inheritance cycles fail.
  - Raw blocks are recognized only when layout behavior is enabled.
- **Owner / Stage Source**: `tags.go`, `tag_include.go`, `tag_extends.go`,
  `tag_block.go`, and `tag_raw.go`.

### SafeHTML

- **Definition**: A Go value marking text as trusted for final HTML output.
- **Includes**: Values explicitly constructed as `SafeHTML`, values returned by
  the `safe` filter, and HTML-mode escaping filters that have already escaped
  the content.
- **Excludes**: Sanitization, non-HTML trust, cross-context trust, and
  automatic preservation through ordinary filters.
- **Lifecycle**: Created by Go code or layout-enabled filters, then consumed
  by the HTML output sink.
- **Identity**: The Go dynamic type at output time.
- **Invariants**:
  - `FormatHTML` bypasses final escaping only for `SafeHTML`.
  - `FormatText` treats `SafeHTML` as an ordinary string.
  - Ordinary filters downgrade trust by returning ordinary values.
- **Owner / Stage Source**: `safe.go`, `filter_string.go`, and `nodes.go`.

### Value Law

- **Definition**: The internal rules for lookup, truthiness, conversion,
  stringification, length, indexing, iteration, and missing values.
- **Includes**: Template expressions, conditionals, loops, filters, output, and
  `Data.Get`.
- **Excludes**: Public value wrapper APIs and global strict-mode switches.
- **Lifecycle**: Applied during expression evaluation, rendering, filtering,
  and explicit data lookup.
- **Invariants**:
  - Missing map keys and struct fields are nil-like in templates.
  - Struct lookup honors JSON tags and exported fields only.
  - Map lookup uses map keys; sequence lookup uses indexes.
  - Strings are indexed, iterated, sliced, and measured by runes.
  - Integer contexts require exact finite integer values.
  - Map and slice stringification is deterministic enough for template output
    and filter coercion.
- **Owner / Stage Source**: `value.go`, `data.go`, `expr.go`,
  `filter_string.go`, and `nodes.go`.

## Contracts

### Engine-First Public API

- **Decision**: Keep normal usage centered on `Engine`, `Template`, `Data`,
  `Loader`, `Format`, `FeatureLayout`, `SafeHTML`, and `FilterFunc`.
- **Why**: The engine is the only place that can keep loader identity, output
  format, optional syntax, defaults, and extension meaning coherent.
- **Rejected**: Public lexer/parser/token/AST/render-context/value/tag-registry
  APIs. They freeze implementation machinery and make future simplification
  harder.
- **Contract Impact**:
  - Public examples must use `Engine`.
  - Custom tag authoring remains private.
  - Tests may inspect internals inside the package, but godoc must not present
    internals as product concepts.

### Compile-Time Filter Binding

- **Decision**: Templates bind filter functions when compiled.
- **Why**: A compiled template should have stable meaning. Render-time registry
  lookup makes cached templates depend on later mutation.
- **Rejected**: Clearing caches on mutation, render-time lookup, and global
  filter override as the normal path.
- **Contract Impact**:
  - `WithFilter` is the preferred construction-time extension path.
  - `RegisterFilter` and `ReplaceFilter` are allowed only before first compile
    or load.
  - After compile/load, those methods return `ErrEngineCompiled`.
  - `Clone(...)` creates a fresh configurable engine with an empty cache.
  - Nil filter functions are configuration errors, not production panics.

### Loader-Resolved Identity

- **Decision**: Use loader-resolved identity for cache keys, in-flight compile
  de-duplication, include recursion decisions, and inheritance cycle checks.
- **Why**: Request names can be aliases. Loader identity is the only stable
  representation of the template that was actually opened.
- **Rejected**: Keying dependency state by request name.
- **Contract Impact**:
  - `Loader.Open` must return a resolved name that distinguishes different
    physical templates.
  - Chain loaders must prevent same-name files in different layers from
    colliding.
  - Include recursion may downgrade to lazy rendering when the resolved
    dependency is already parsing.
  - Inheritance recursion returns `ErrCircularExtends`.

### Layout Feature Gate

- **Decision**: Layout-only syntax is feature-gated.
- **Why**: Source-string rendering should stay small by default, while
  multi-file rendering can opt into the extra language surface.
- **Rejected**: Enabling all tags in every engine.
- **Contract Impact**:
  - Engines without layout enabled reject layout-only tags as unknown.
  - `safe` is not a global filter.
  - Layout-enabled engines layer the needed tags and filters privately.

### HTML Trust At The Sink

- **Decision**: Trust is checked at final HTML output by Go dynamic type.
- **Why**: The output sink is the only place that knows the render format.
- **Rejected**: Broad safe-string semantics, filter-safety metadata, and
  automatic trust preservation through ordinary filters.
- **Contract Impact**:
  - `FormatHTML` escapes ordinary `{{ expr }}` output.
  - `SafeHTML` bypasses final HTML escaping.
  - HTML-mode `escape`, `escape_once`, and `h` return `SafeHTML` after
    escaping so output is not escaped twice.
  - Ordinary filters return ordinary values and therefore lose trust.
  - Producing `SafeHTML` from untrusted input is caller error.

### One Value Law

- **Decision**: Template lookup and `Data.Get` share object-shaped semantics
  where their public contracts overlap.
- **Why**: Template languages fail in small edge cases when lookup, filters,
  conditionals, and explicit data access each invent their own rules.
- **Rejected**: Delegating explicit data lookup to a separate path with
  incompatible map, struct, string, or missing-value behavior.
- **Contract Impact**:
  - Missing fields and map misses render empty, are false in conditionals, and
    are caught by `default`.
  - `Data.Get` preserves sentinel errors for explicit caller handling.
  - Strings use rune semantics for indexing, length, slicing, and iteration.
  - Fractional, infinite, overflowing, and boolean values are invalid in
    integer contexts.

## Failure Semantics

- Loader name validation failures return loader/template sentinel errors and
  must not weaken path safety.
- Missing templates return `ErrTemplateNotFound`, with template context attached
  where rendering or loading can provide it.
- Include recursion is limited by depth and reports `ErrIncludeDepthExceeded`.
- Inheritance cycles report `ErrCircularExtends`; excessive inheritance depth
  reports `ErrExtendsDepthExceeded`.
- Filter lookup failures report `ErrFilterNotFound`.
- Integer conversion failures in filter arguments and subscripts report the
  appropriate numeric or index sentinel.
- `Data.Get` reports `ErrContextKeyNotFound`,
  `ErrContextIndexOutOfRange`, or `ErrContextInvalidKeyType` according to the
  explicit lookup failure.
- Render failures should preserve sentinel causes for `errors.Is` and template
  location for diagnostics.

## Security Semantics

- Template names are untrusted input. Every loader must validate names before
  opening source.
- Directory loading must preserve its root boundary, including symlink escape
  attempts.
- Loader-backed identity must prevent cache collisions across override layers.
- HTML escaping is opt-in through `FormatHTML`; text rendering does not escape.
- `SafeHTML` is a trust marker, not a sanitizer.
- This package does not claim to safely execute untrusted template source.

## Forbidden

- Do not expose lexer, parser, token, AST, render context, raw value, or tag
  registry types as public API.
- Do not add a public custom-tag API by exporting internal parser/runtime
  structures.
- Do not add package-global mutable configuration as the normal extension path.
- Do not allow filter mutation to change already-compiled templates.
- Do not key loader-backed caches, in-flight compiles, or dependency cycle
  checks by request name when a resolved identity exists.
- Do not weaken template-name validation, directory sandbox behavior, include
  depth limits, inheritance cycle checks, or HTML escaping semantics.
- Do not preserve HTML trust through ordinary filters.
- Do not introduce strict-mode switches before a concrete caller contract proves
  that missing-value policy belongs in the engine.

## Acceptance Criteria

- Loader tests prove cache identity, include recursion, and inheritance cycles
  use resolved names.
- Engine tests prove compiled templates keep filter meaning after attempted
  mutation and that cloning is the fresh-configuration path.
- Package godoc exposes the engine-first public model and not internal parser,
  lexer, AST, render context, tag registry, or value APIs.
- Value-law tests cover missing fields, map key mismatch, struct tags,
  hidden fields, integer conversion, stringification, rune indexing, rune
  length, rune slicing, string iteration, filters, template output, conditionals,
  loops, and `Data.Get`.
- HTML trust tests prove ordinary output is escaped in `FormatHTML`, escaped
  filters do not double-escape, `SafeHTML` bypasses final escaping, and ordinary
  filters downgrade trust.
- README, docs, examples, and `ARCHITECTURE.md` point to the same public API
  shape and do not describe internals as caller contracts.
- `task test` and `task lint` pass.
