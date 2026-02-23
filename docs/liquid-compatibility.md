# Liquid Standard Compatibility

This document describes the compatibility between `kaptinlin/template` and the [Shopify Liquid](https://shopify.github.io/liquid/) template language standard (v5.x).

## Overview

| Metric | Count |
|--------|-------|
| Liquid standard filters | 46 |
| Fully compliant | 41 |
| Behavioral differences | 2 |
| Missing | 3 |
| Extension filters | 20 |
| Convenience aliases | 9 |

This engine implements **89%** of the Liquid standard with full behavioral compliance, and provides 20 additional extension filters for common use cases.

## Syntax Compatibility

This engine supports Liquid-standard template syntax:

| Feature | Liquid | This Engine |
|---------|--------|-------------|
| Variable output | `{{ variable }}` | `{{ variable }}` |
| Filters with pipe | `{{ value \| filter }}` | `{{ value \| filter }}` |
| Filter arguments | `{{ value \| filter: arg1, arg2 }}` | `{{ value \| filter:arg1,arg2 }}` |
| Filter chaining | `{{ value \| a \| b }}` | `{{ value \| a \| b }}` |
| Tags | `{% tag %}` | `{% tag %}` |
| Comments | `{% comment %}...{% endcomment %}` | `{# ... #}` |

**Syntax differences:**
- Comments use `{# #}` (Django-style) instead of `{% comment %}{% endcomment %}`.
- Filter arguments omit spaces after `:` and between `,` separators.

---

## Liquid Standard Filters

### String Filters

| Filter | Liquid | Status |
|--------|--------|--------|
| `append` | `{{ "Hello" \| append: " World" }}` | Compliant |
| `capitalize` | `{{ "hello" \| capitalize }}` | Compliant |
| `downcase` | `{{ "HELLO" \| downcase }}` | Compliant |
| `escape` | `{{ "<b>bold</b>" \| escape }}` | Compliant |
| `escape_once` | `{{ "&lt;b&gt;" \| escape_once }}` | Compliant |
| `lstrip` | `{{ "  hello  " \| lstrip }}` | Compliant |
| `prepend` | `{{ "World" \| prepend: "Hello " }}` | Compliant |
| `remove` | `{{ "Hello World" \| remove: "World" }}` | Compliant |
| `remove_first` | `{{ "abcabc" \| remove_first: "b" }}` | Compliant |
| `remove_last` | `{{ "abcabc" \| remove_last: "b" }}` | Compliant |
| `replace` | `{{ "Hello" \| replace: "Hello", "Hi" }}` | Compliant |
| `replace_first` | `{{ "aabbcc" \| replace_first: "b", "x" }}` | Compliant |
| `replace_last` | `{{ "aabbcc" \| replace_last: "b", "x" }}` | Compliant |
| `rstrip` | `{{ "  hello  " \| rstrip }}` | Compliant |
| `slice` | `{{ "hello" \| slice: 1, 3 }}` | Compliant |
| `split` | `{{ "a,b,c" \| split: "," }}` | Compliant |
| `strip` | `{{ "  hello  " \| strip }}` | Compliant |
| `strip_html` | `{{ "<p>text</p>" \| strip_html }}` | Compliant |
| `strip_newlines` | `{{ text \| strip_newlines }}` | Compliant |
| `truncate` | `{{ text \| truncate: 20 }}` | Compliant (defaults to 50) |
| `truncatewords` | `{{ text \| truncatewords: 5 }}` | Compliant (defaults to 15) |
| `upcase` | `{{ "hello" \| upcase }}` | Compliant |
| `url_decode` | `{{ text \| url_decode }}` | Compliant |
| `url_encode` | `{{ text \| url_encode }}` | Compliant |
| `base64_decode` | `{{ text \| base64_decode }}` | Compliant |
| `base64_encode` | `{{ text \| base64_encode }}` | Compliant |
| `base64_url_safe_decode` | `{{ text \| base64_url_safe_decode }}` | **Missing** |
| `base64_url_safe_encode` | `{{ text \| base64_url_safe_encode }}` | **Missing** |
| `newline_to_br` | `{{ text \| newline_to_br }}` | **Missing** |

### Math Filters

| Filter | Liquid | Status |
|--------|--------|--------|
| `abs` | `{{ -5 \| abs }}` | Compliant |
| `at_least` | `{{ 3 \| at_least: 5 }}` | Compliant |
| `at_most` | `{{ 10 \| at_most: 8 }}` | Compliant |
| `ceil` | `{{ 3.01 \| ceil }}` | Compliant |
| `divided_by` | `{{ 20 \| divided_by: 4 }}` | Compliant |
| `floor` | `{{ 3.99 \| floor }}` | Compliant |
| `minus` | `{{ 10 \| minus: 2 }}` | Compliant |
| `modulo` | `{{ 10 \| modulo: 3 }}` | Compliant |
| `plus` | `{{ 7 \| plus: 3 }}` | Compliant |
| `round` | `{{ 3.14 \| round: 2 }}` | Compliant (defaults to 0) |
| `times` | `{{ 5 \| times: 2 }}` | Compliant |

### Array Filters

| Filter | Liquid | Status |
|--------|--------|--------|
| `compact` | `{{ array \| compact }}` | Compliant |
| `concat` | `{{ array \| concat: other }}` | Compliant |
| `find` | `{{ array \| find: "name", "Bob" }}` | Compliant |
| `find_index` | `{{ array \| find_index: "name", "Bob" }}` | Compliant |
| `first` | `{{ array \| first }}` | Compliant |
| `has` | `{{ array \| has: "name", "Alice" }}` | Compliant |
| `join` | `{{ array \| join: ", " }}` | Compliant (defaults to `" "`) |
| `last` | `{{ array \| last }}` | Compliant |
| `map` | `{{ array \| map: "name" }}` | Compliant |
| `reject` | `{{ array \| reject: "active", "false" }}` | Compliant |
| `reverse` | `{{ array \| reverse }}` | Compliant |
| `size` | `{{ array \| size }}` | Compliant (works on strings, arrays, and maps) |
| `sort` | `{{ array \| sort }}` | Compliant |
| `sort_natural` | `{{ array \| sort_natural }}` | Compliant |
| `sum` | `{{ array \| sum }}` | Compliant (supports `sum:'property'`) |
| `uniq` | `{{ array \| uniq }}` | Compliant (supports `uniq:'property'`) |
| `where` | `{{ array \| where: "active", "true" }}` | Compliant |

### Other Filters

| Filter | Liquid | Status |
|--------|--------|--------|
| `date` | `{{ timestamp \| date: "%Y-%m-%d" }}` | Behavioral difference |
| `default` | `{{ value \| default: "fallback" }}` | Behavioral difference |

---

## Behavioral Differences

### `default`

| | Liquid | This Engine |
|---|--------|-------------|
| Signature | `default: fallback, allow_false: true` | `default:'fallback'` |
| Named parameter `allow_false` | Supported. When `true`, only `nil` and empty string trigger the fallback; `false` passes through. | Not supported. The parser does not support named parameters. Uses Go/Django-style truthiness evaluation (empty string, `nil`, `0`, `false` all trigger the fallback). |

### `date`

| | Liquid | This Engine |
|---|--------|-------------|
| Format syntax | strftime (`%Y-%m-%d %H:%M`) | PHP-style (`Y-m-d H:i`) |
| `"now"` / `"today"` input | Supported as special string inputs | Not supported |

**Format specifier mapping:**

| Liquid (strftime) | This Engine (PHP) | Description |
|---|---|---|
| `%Y` | `Y` | 4-digit year |
| `%m` | `m` | 2-digit month |
| `%d` | `d` | 2-digit day |
| `%H` | `H` | 24-hour hour |
| `%M` | `i` | Minutes |
| `%S` | `s` | Seconds |

---

## Missing Filters

These 3 Liquid standard filters are not yet implemented:

| Filter | Description |
|--------|-------------|
| `newline_to_br` | Replaces `\n` with `<br />\n` |
| `base64_url_safe_encode` | URL-safe Base64 encoding (uses `-` and `_` instead of `+` and `/`) |
| `base64_url_safe_decode` | Decodes URL-safe Base64 |

---

## Extension Filters

These filters are not part of the Liquid standard but are provided as useful additions.

### String Extensions

| Filter | Description | Example |
|--------|-------------|---------|
| `titleize` | Capitalize first letter of each word | `{{ "hello world" \| titleize }}` |
| `camelize` | Convert to camelCase | `{{ "hello_world" \| camelize }}` |
| `pascalize` | Convert to PascalCase | `{{ "hello_world" \| pascalize }}` |
| `dasherize` | Convert to dash-separated lowercase | `{{ "hello world" \| dasherize }}` |
| `slugify` | URL-friendly slug | `{{ "Hello World & Friends" \| slugify }}` |
| `pluralize` | Singular/plural form by count | `{{ count \| pluralize:'apple','apples' }}` |
| `ordinalize` | Ordinal English form | `{{ 1 \| ordinalize }}` |
| `length` | String/array/map length | `{{ "hello" \| length }}` |

### Array Extensions

| Filter | Description | Example |
|--------|-------------|---------|
| `random` | Random element from array | `{{ items \| random }}` |
| `shuffle` | Randomly rearrange elements | `{{ items \| shuffle }}` |
| `max` | Maximum value from numeric array | `{{ scores \| max }}` |
| `min` | Minimum value from numeric array | `{{ scores \| min }}` |
| `average` | Average of numeric array | `{{ scores \| average }}` |

### Date Extensions

| Filter | Description | Example |
|--------|-------------|---------|
| `day` | Extract day of month | `{{ timestamp \| day }}` |
| `month` | Extract month number | `{{ timestamp \| month }}` |
| `month_full` | Full month name | `{{ timestamp \| month_full }}` |
| `year` | Extract year | `{{ timestamp \| year }}` |
| `week` | ISO week number | `{{ timestamp \| week }}` |
| `weekday` | Day of week name | `{{ timestamp \| weekday }}` |
| `time_ago` | Human-readable relative time | `{{ timestamp \| time_ago }}` |

### Number Extensions

| Filter | Description | Example |
|--------|-------------|---------|
| `number` | Format number with pattern | `{{ 1234.5 \| number:"#,###.##" }}` |
| `bytes` | Human-readable byte size | `{{ 2048 \| bytes }}` |

### Other Extensions

| Filter | Description | Example |
|--------|-------------|---------|
| `json` | Serialize to JSON string | `{{ data \| json }}` |
| `extract` | Nested value by dot-path key | `{{ data \| extract:"user.name" }}` |

---

## Convenience Aliases

These aliases are provided for developer convenience. Both the alias and the primary Liquid name work identically.

| Alias | Primary (Liquid) Name | Note |
|-------|-----------------------|------|
| `trim` | `strip` | |
| `trim_left` | `lstrip` | |
| `trim_right` | `rstrip` | |
| `upper` | `upcase` | |
| `lower` | `downcase` | |
| `h` | `escape` | Also a standard Liquid alias |
| `truncate_words` | `truncatewords` | |
| `unique` | `uniq` | |
| `divide` | `divided_by` | |
| `timeago` | `time_ago` | Extension filter alias |
