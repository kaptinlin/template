package template

import (
	"fmt"
	"reflect"

	"github.com/kaptinlin/filter"
)

// registerStringFilters registers all string-related filters.
func registerStringFilters() {
	// Liquid-standard primary names
	RegisterFilter("default", defaultFilter)
	RegisterFilter("strip", trimFilter)
	RegisterFilter("lstrip", trimLeftFilter)
	RegisterFilter("rstrip", trimRightFilter)
	RegisterFilter("split", splitFilter)
	RegisterFilter("replace", replaceFilter)
	RegisterFilter("replace_first", replaceFirstFilter)
	RegisterFilter("replace_last", replaceLastFilter)
	RegisterFilter("remove", removeFilter)
	RegisterFilter("remove_first", removeFirstFilter)
	RegisterFilter("remove_last", removeLastFilter)
	RegisterFilter("append", appendFilter)
	RegisterFilter("prepend", prependFilter)
	RegisterFilter("length", lengthFilter)
	RegisterFilter("upcase", upperFilter)
	RegisterFilter("downcase", lowerFilter)
	RegisterFilter("capitalize", capitalizeFilter)
	RegisterFilter("escape", escapeFilter)
	RegisterFilter("escape_once", escapeOnceFilter)
	RegisterFilter("strip_html", stripHTMLFilter)
	RegisterFilter("strip_newlines", stripNewlinesFilter)
	RegisterFilter("url_encode", urlEncodeFilter)
	RegisterFilter("url_decode", urlDecodeFilter)
	RegisterFilter("base64_encode", base64EncodeFilter)
	RegisterFilter("base64_decode", base64DecodeFilter)
	RegisterFilter("truncate", truncateFilter)
	RegisterFilter("truncatewords", truncateWordsFilter)
	RegisterFilter("slice", sliceFilter)

	// Extension filters (non-Liquid)
	RegisterFilter("titleize", titleizeFilter)
	RegisterFilter("camelize", camelizeFilter)
	RegisterFilter("pascalize", pascalizeFilter)
	RegisterFilter("dasherize", dasherizeFilter)
	RegisterFilter("slugify", slugifyFilter)
	RegisterFilter("pluralize", pluralizeFilter)
	RegisterFilter("ordinalize", ordinalizeFilter)

	// Aliases
	RegisterFilter("trim", trimFilter)
	RegisterFilter("trim_left", trimLeftFilter)
	RegisterFilter("trim_right", trimRightFilter)
	RegisterFilter("upper", upperFilter)
	RegisterFilter("lower", lowerFilter)
	RegisterFilter("h", escapeFilter)
	RegisterFilter("truncate_words", truncateWordsFilter)
}

// defaultFilter returns a default value if the input is falsy.
func defaultFilter(value any, args ...any) (any, error) {
	if NewValue(value).IsTrue() {
		return value, nil
	}
	if len(args) > 0 {
		return args[0], nil
	}
	return "", nil
}

// trimFilter removes leading and trailing whitespace.
func trimFilter(value any, _ ...any) (any, error) {
	return filter.Trim(toString(value)), nil
}

// trimLeftFilter removes leading whitespace.
func trimLeftFilter(value any, _ ...any) (any, error) {
	return filter.TrimLeft(toString(value)), nil
}

// trimRightFilter removes trailing whitespace.
func trimRightFilter(value any, _ ...any) (any, error) {
	return filter.TrimRight(toString(value)), nil
}

// splitFilter divides a string by a delimiter.
func splitFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: split filter requires a delimiter argument", ErrInsufficientArgs)
	}
	return filter.Split(toString(value), toString(args[0])), nil
}

// replaceFilter substitutes all instances of a substring.
func replaceFilter(value any, args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: replace filter requires two arguments (old and new substrings)", ErrInsufficientArgs)
	}
	oldStr, newStr := toString(args[0]), toString(args[1])
	return filter.Replace(toString(value), oldStr, newStr), nil
}

// replaceFirstFilter substitutes the first instance of a substring.
func replaceFirstFilter(value any, args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: replace_first filter requires two arguments (old and new substrings)", ErrInsufficientArgs)
	}
	return filter.ReplaceFirst(toString(value), toString(args[0]), toString(args[1])), nil
}

// replaceLastFilter substitutes the last instance of a substring.
func replaceLastFilter(value any, args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: replace_last filter requires two arguments (old and new substrings)", ErrInsufficientArgs)
	}
	return filter.ReplaceLast(toString(value), toString(args[0]), toString(args[1])), nil
}

// removeFilter eliminates all occurrences of a substring.
func removeFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: remove filter requires a substring argument", ErrInsufficientArgs)
	}
	return filter.Remove(toString(value), toString(args[0])), nil
}

// removeFirstFilter eliminates the first occurrence of a substring.
func removeFirstFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: remove_first filter requires a substring argument", ErrInsufficientArgs)
	}
	return filter.RemoveFirst(toString(value), toString(args[0])), nil
}

// removeLastFilter eliminates the last occurrence of a substring.
func removeLastFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: remove_last filter requires a substring argument", ErrInsufficientArgs)
	}
	return filter.RemoveLast(toString(value), toString(args[0])), nil
}

// appendFilter adds characters to the end of a string.
func appendFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: append filter requires a string to append", ErrInsufficientArgs)
	}
	return filter.Append(toString(value), toString(args[0])), nil
}

// prependFilter adds characters to the beginning of a string.
func prependFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: prepend filter requires a string to prepend", ErrInsufficientArgs)
	}
	return filter.Prepend(toString(value), toString(args[0])), nil
}

// lengthFilter returns the length of a string, slice, array, or map.
func lengthFilter(value any, _ ...any) (any, error) {
	v := reflect.ValueOf(value)
	//exhaustive:ignore
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len(), nil
	}
	return filter.Length(toString(value)), nil
}

// upperFilter converts all characters to uppercase.
func upperFilter(value any, _ ...any) (any, error) {
	return filter.Upper(toString(value)), nil
}

// lowerFilter converts all characters to lowercase.
func lowerFilter(value any, _ ...any) (any, error) {
	return filter.Lower(toString(value)), nil
}

// titleizeFilter capitalizes the first letter of each word.
func titleizeFilter(value any, _ ...any) (any, error) {
	return filter.Titleize(toString(value)), nil
}

// capitalizeFilter capitalizes the first letter of a string and lowercases the rest.
func capitalizeFilter(value any, _ ...any) (any, error) {
	return filter.Capitalize(toString(value)), nil
}

// camelizeFilter converts a string to camelCase.
func camelizeFilter(value any, _ ...any) (any, error) {
	return filter.Camelize(toString(value)), nil
}

// pascalizeFilter converts a string to PascalCase.
func pascalizeFilter(value any, _ ...any) (any, error) {
	return filter.Pascalize(toString(value)), nil
}

// dasherizeFilter transforms a string into a lowercased, dash-separated format.
func dasherizeFilter(value any, _ ...any) (any, error) {
	return filter.Dasherize(toString(value)), nil
}

// slugifyFilter converts a string into a URL-friendly slug.
func slugifyFilter(value any, _ ...any) (any, error) {
	return filter.Slugify(toString(value)), nil
}

// pluralizeFilter determines the singular or plural form based on a numeric value.
func pluralizeFilter(value any, args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: pluralize filter requires two arguments (singular and plural forms)", ErrInsufficientArgs)
	}
	count, err := toInteger(value)
	if err != nil {
		return nil, err
	}
	singular, plural := toString(args[0]), toString(args[1])
	return filter.Pluralize(count, singular, plural), nil
}

// ordinalizeFilter converts a number to its ordinal English form.
func ordinalizeFilter(value any, _ ...any) (any, error) {
	n, err := toInteger(value)
	if err != nil {
		return nil, err
	}
	return filter.Ordinalize(n), nil
}

// truncateFilter shortens a string to a specified length (default 50).
// An optional second argument specifies the ellipsis string (default "...").
func truncateFilter(value any, args ...any) (any, error) {
	maxLen := 50
	if len(args) >= 1 {
		n, err := toInteger(args[0])
		if err != nil {
			return nil, err
		}
		maxLen = n
	}
	if len(args) >= 2 {
		return filter.Truncate(toString(value), maxLen, toString(args[1])), nil
	}
	return filter.Truncate(toString(value), maxLen), nil
}

// truncateWordsFilter truncates a string to a specified number of words (default 15).
// An optional second argument specifies the ellipsis string (default "...").
func truncateWordsFilter(value any, args ...any) (any, error) {
	maxWords := 15
	if len(args) >= 1 {
		n, err := toInteger(args[0])
		if err != nil {
			return nil, err
		}
		maxWords = n
	}
	if len(args) >= 2 {
		return filter.TruncateWords(toString(value), maxWords, toString(args[1])), nil
	}
	return filter.TruncateWords(toString(value), maxWords), nil
}

// escapeFilter escapes HTML special characters.
func escapeFilter(value any, _ ...any) (any, error) {
	return filter.Escape(toString(value)), nil
}

// escapeOnceFilter escapes HTML without double-escaping already-escaped entities.
func escapeOnceFilter(value any, _ ...any) (any, error) {
	return filter.EscapeOnce(toString(value)), nil
}

// stripHTMLFilter removes all HTML tags from a string.
func stripHTMLFilter(value any, _ ...any) (any, error) {
	return filter.StripHTML(toString(value)), nil
}

// stripNewlinesFilter removes all newline characters from a string.
func stripNewlinesFilter(value any, _ ...any) (any, error) {
	return filter.StripNewlines(toString(value)), nil
}

// urlEncodeFilter percent-encodes a string for use in URLs.
func urlEncodeFilter(value any, _ ...any) (any, error) {
	return filter.URLEncode(toString(value)), nil
}

// urlDecodeFilter decodes a percent-encoded string.
func urlDecodeFilter(value any, _ ...any) (any, error) {
	return filter.URLDecode(toString(value))
}

// base64EncodeFilter encodes a string to Base64.
func base64EncodeFilter(value any, _ ...any) (any, error) {
	return filter.Base64Encode(toString(value)), nil
}

// base64DecodeFilter decodes a Base64-encoded string.
func base64DecodeFilter(value any, _ ...any) (any, error) {
	return filter.Base64Decode(toString(value))
}

// sliceFilter extracts a substring or sub-slice by offset and optional length.
func sliceFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: slice filter requires an offset argument", ErrInsufficientArgs)
	}
	offset, err := toInteger(args[0])
	if err != nil {
		return nil, err
	}
	if len(args) >= 2 {
		length, err := toInteger(args[1])
		if err != nil {
			return nil, err
		}
		return filter.Slice(value, offset, length)
	}
	return filter.Slice(value, offset)
}
