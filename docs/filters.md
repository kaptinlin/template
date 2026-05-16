# Filters

Filters transform values inside output expressions. Chain them with `|` and pass
arguments after `:`.

```plaintext
{{ variable | filter }}
{{ variable | filter:arg1,"arg2" | other_filter }}
```

Examples in this guide use variables supplied by render data. The template
language supports string, number, boolean, and nil literals in expressions, but
collections such as arrays and maps should be passed through `template.Data`.

Most built-in filters delegate their core behavior to
[github.com/kaptinlin/filter](https://github.com/kaptinlin/filter).

## String Filters

### Defaults and Whitespace

| Filter | Aliases | Description | Example |
|---|---|---|---|
| `default` | | Returns the fallback when the value is falsey | `{{ name | default:"Guest" }}` -> `Guest` |
| `strip` | `trim` | Trims leading and trailing whitespace | `{{ "  hi  " | strip }}` -> `hi` |
| `lstrip` | `trim_left` | Trims leading whitespace | `{{ "  hi  " | lstrip }}` -> `hi  ` |
| `rstrip` | `trim_right` | Trims trailing whitespace | `{{ "  hi  " | rstrip }}` -> `  hi` |
| `strip_newlines` | | Removes newline characters | `{{ text | strip_newlines }}` |

### Text Changes

| Filter | Aliases | Description | Example |
|---|---|---|---|
| `append` | | Appends a string | `{{ "hello" | append:"!" }}` -> `hello!` |
| `prepend` | | Prepends a string | `{{ "world" | prepend:"hello " }}` -> `hello world` |
| `replace` | | Replaces all matches | `{{ "a-b-a" | replace:"a","x" }}` -> `x-b-x` |
| `replace_first` | | Replaces the first match | `{{ "a-b-a" | replace_first:"a","x" }}` -> `x-b-a` |
| `replace_last` | | Replaces the last match | `{{ "a-b-a" | replace_last:"a","x" }}` -> `a-b-x` |
| `remove` | | Removes all matches | `{{ "a-b-a" | remove:"-" }}` -> `aba` |
| `remove_first` | | Removes the first match | `{{ "abcabc" | remove_first:"b" }}` -> `acabc` |
| `remove_last` | | Removes the last match | `{{ "abcabc" | remove_last:"b" }}` -> `abcac` |
| `split` | | Splits a string into a slice | `{{ "a,b,c" | split:"," | join:"/" }}` -> `a/b/c` |
| `slice` | | Returns a substring or sub-slice by offset and optional length | `{{ "hello" | slice:1,3 }}` -> `ell` |

### Case and Identifiers

| Filter | Aliases | Description | Example |
|---|---|---|---|
| `upcase` | `upper` | Uppercases text | `{{ "hello" | upcase }}` -> `HELLO` |
| `downcase` | `lower` | Lowercases text | `{{ "HELLO" | downcase }}` -> `hello` |
| `capitalize` | | Capitalizes the first letter and lowercases the rest | `{{ "hELLO" | capitalize }}` -> `Hello` |
| `titleize` | | Capitalizes words | `{{ "hello world" | titleize }}` -> `Hello World` |
| `camelize` | | Converts to camelCase | `{{ "hello_world" | camelize }}` -> `helloWorld` |
| `pascalize` | | Converts to PascalCase | `{{ "hello_world" | pascalize }}` -> `HelloWorld` |
| `dasherize` | | Converts to dash-separated lowercase | `{{ "hello world" | dasherize }}` -> `hello-world` |
| `slugify` | | Creates a URL-friendly slug | `{{ "Hello World & Friends" | slugify }}` -> `hello-world-and-friends` |

### HTML, URL, and Encoding

| Filter | Aliases | Description | Example |
|---|---|---|---|
| `escape` | `h` | Escapes HTML special characters | `{{ html | escape }}` |
| `escape_once` | | Escapes HTML without double-escaping existing entities | `{{ html | escape_once }}` |
| `strip_html` | | Removes HTML tags | `{{ html | strip_html }}` |
| `url_encode` | | Percent-encodes text for URLs | `{{ "hello world" | url_encode }}` -> `hello+world` |
| `url_decode` | | Decodes percent-encoded text | `{{ encoded | url_decode }}` |
| `base64_encode` | | Encodes text as Base64 | `{{ "hello" | base64_encode }}` -> `aGVsbG8=` |
| `base64_decode` | | Decodes Base64 text | `{{ value | base64_decode }}` |

Under `FormatHTML`, `escape`, `escape_once`, and `h` return trusted escaped
content so the automatic escaper does not escape them a second time.

### Length and English Helpers

| Filter | Aliases | Description | Example |
|---|---|---|---|
| `length` | | Length of a string, array, slice, or map | `{{ "hello" | length }}` -> `5` |
| `truncate` | | Truncates to a character count; optional custom ellipsis | `{{ "hello world" | truncate:5 }}` -> `he...` |
| `truncatewords` | `truncate_words` | Truncates to a word count; optional custom ellipsis | `{{ "hello beautiful world" | truncatewords:2 }}` -> `hello beautiful...` |
| `pluralize` | | Chooses singular or plural form from a numeric value | `{{ count | pluralize:"item","items" }}` |
| `ordinalize` | | Formats an integer ordinal | `{{ 2 | ordinalize }}` -> `2nd` |

## Array Filters

The examples below assume render data such as `numbers`, `words`, `users`,
`products`, and `more_numbers`.

| Filter | Aliases | Description | Example |
|---|---|---|---|
| `join` | | Joins items; default separator is a space | `{{ words | join:", " }}` |
| `first` | | Returns the first item | `{{ words | first }}` |
| `last` | | Returns the last item | `{{ words | last }}` |
| `reverse` | | Reverses item order | `{{ numbers | reverse | join:"," }}` |
| `size` | | Length of a string, array, slice, or map | `{{ users | size }}` |
| `uniq` | `unique` | Removes duplicates; optional property key keeps the first item per key | `{{ users | uniq:"role" | map:"name" | join:"," }}` |
| `sort` | | Sorts ascending; optional property key | `{{ users | sort:"name" | map:"name" | join:"," }}` |
| `sort_natural` | | Case-insensitive natural sort; optional property key | `{{ words | sort_natural | join:"," }}` |
| `compact` | | Removes nil items; optional property key removes items with nil property values | `{{ products | compact:"image" | size }}` |
| `concat` | | Appends another array or slice | `{{ numbers | concat:more_numbers | join:"," }}` |
| `map` | | Extracts a property from every item | `{{ users | map:"name" | join:", " }}` |
| `where` | | Selects items where a property is truthy or equals a value | `{{ users | where:"active","true" | map:"name" | join:"," }}` |
| `reject` | | Rejects items where a property is truthy or equals a value | `{{ users | reject:"active","false" | map:"name" | join:"," }}` |
| `find` | | Returns the first item where a property equals a value | `{{ users | find:"name","Bob" | extract:"age" }}` |
| `find_index` | | Returns the first matching item index | `{{ users | find_index:"name","Bob" }}` |
| `has` | | Reports whether any item matches | `{{ users | has:"name","Alice" }}` |
| `sum` | | Sums numeric items; optional property key | `{{ products | sum:"price" }}` |
| `max` | | Maximum numeric item | `{{ numbers | max }}` |
| `min` | | Minimum numeric item | `{{ numbers | min }}` |
| `average` | | Average of numeric items | `{{ numbers | average }}` |
| `random` | | Returns a random item | `{{ words | random }}` |
| `shuffle` | | Returns items in random order | `{{ words | shuffle | join:"," }}` |

## Date Filters

Date filters accept `time.Time`, supported date strings, and supported numeric
timestamps through the underlying filter package.

| Filter | Aliases | Description | Example |
|---|---|---|---|
| `date` | | Formats with PHP-style date tokens | `{{ current | date:"Y-m-d" }}` -> `2024-03-30` |
| `date` | | Unix timestamp token | `{{ current | date:"U" }}` -> `1711811045` |
| `day` | | Day of month | `{{ current | day }}` -> `30` |
| `month` | | Month number | `{{ current | month }}` -> `3` |
| `month_full` | | Full month name | `{{ current | month_full }}` -> `March` |
| `year` | | Year | `{{ current | year }}` -> `2024` |
| `week` | | ISO week number | `{{ current | week }}` -> `13` |
| `weekday` | | Weekday name | `{{ current | weekday }}` -> `Saturday` |
| `time_ago` | `timeago` | Human-readable relative time | `{{ past | time_ago }}` |

## Number and Math Filters

| Filter | Aliases | Description | Example |
|---|---|---|---|
| `number` | | Formats a number with a pattern | `{{ 1234567.89 | number:"#,###.##" }}` -> `1,234,567.89` |
| `bytes` | | Formats a byte count | `{{ 2048 | bytes }}` -> `2.0 KB` |
| `abs` | | Absolute value | `{{ -5 | abs }}` -> `5` |
| `at_least` | | Clamps to a minimum | `{{ 3 | at_least:5 }}` -> `5` |
| `at_most` | | Clamps to a maximum | `{{ 10 | at_most:8 }}` -> `8` |
| `round` | | Rounds; optional precision defaults to 0 | `{{ 3.142 | round:2 }}` -> `3.14` |
| `floor` | | Rounds down | `{{ 3.99 | floor }}` -> `3` |
| `ceil` | | Rounds up | `{{ 3.01 | ceil }}` -> `4` |
| `plus` | | Adds a number | `{{ 5 | plus:3 }}` -> `8` |
| `minus` | | Subtracts a number | `{{ 10 | minus:4 }}` -> `6` |
| `times` | | Multiplies by a number | `{{ 5 | times:4 }}` -> `20` |
| `divided_by` | `divide` | Divides by a number | `{{ 20 | divided_by:5 }}` -> `4` |
| `modulo` | | Remainder after division | `{{ 10 | modulo:3 }}` -> `1` |

Division and modulo by zero return render errors that match the numeric
sentinels in `errors.go`.

## Format and Map Filters

| Filter | Description | Example |
|---|---|---|
| `json` | Serializes the value to deterministic JSON | `{{ data | json }}` |
| `extract` | Retrieves a nested value from a map, slice, or array by dot path | `{{ data | extract:"user.profile.age" }}` |

`extract` returns an empty string for missing keys or out-of-range indexes so
templates can stay concise when optional data is absent.
