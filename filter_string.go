package template

import (
	"fmt"
	"reflect"

	"github.com/kaptinlin/filter"
)

// registerStringFilters registers all string-related filters.
func registerStringFilters() {
	RegisterFilter("default", defaultFilter)
	RegisterFilter("trim", trimFilter)
	RegisterFilter("split", splitFilter)
	RegisterFilter("replace", replaceFilter)
	RegisterFilter("remove", removeFilter)
	RegisterFilter("append", appendFilter)
	RegisterFilter("prepend", prependFilter)
	RegisterFilter("length", lengthFilter)
	RegisterFilter("upper", upperFilter)
	RegisterFilter("lower", lowerFilter)
	RegisterFilter("titleize", titleizeFilter)
	RegisterFilter("capitalize", capitalizeFilter)
	RegisterFilter("camelize", camelizeFilter)
	RegisterFilter("pascalize", pascalizeFilter)
	RegisterFilter("dasherize", dasherizeFilter)
	RegisterFilter("slugify", slugifyFilter)
	RegisterFilter("pluralize", pluralizeFilter)
	RegisterFilter("ordinalize", ordinalizeFilter)
	RegisterFilter("truncate", truncateFilter)
	RegisterFilter("truncateWords", truncateWordsFilter)
}

// defaultFilter returns a default value if the input is falsy.
func defaultFilter(value any, args ...string) (any, error) {
	fallback := ""
	if len(args) > 0 {
		fallback = args[0]
	}
	if !NewValue(value).IsTrue() {
		return fallback, nil
	}
	return value, nil
}

// trimFilter removes leading and trailing whitespace.
func trimFilter(value any, _ ...string) (any, error) {
	return filter.Trim(toString(value)), nil
}

// splitFilter divides a string by a delimiter.
func splitFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: split filter requires a delimiter argument", ErrInsufficientArgs)
	}
	delimiter := args[0]
	return filter.Split(toString(value), delimiter), nil
}

// replaceFilter substitutes all instances of a substring.
func replaceFilter(value any, args ...string) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: replace filter requires two arguments (old and new substrings)", ErrInsufficientArgs)
	}
	oldStr, newStr := args[0], args[1]
	return filter.Replace(toString(value), oldStr, newStr), nil
}

// removeFilter eliminates all occurrences of a substring.
func removeFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: remove filter requires a substring argument", ErrInsufficientArgs)
	}
	sub := args[0]
	return filter.Remove(toString(value), sub), nil
}

// appendFilter adds characters to the end of a string.
func appendFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: append filter requires a string to append", ErrInsufficientArgs)
	}
	suffix := args[0]
	return filter.Append(toString(value), suffix), nil
}

// prependFilter adds characters to the beginning of a string.
func prependFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: prepend filter requires a string to prepend", ErrInsufficientArgs)
	}
	prefix := args[0]
	return filter.Prepend(toString(value), prefix), nil
}

// lengthFilter returns the length of a string, slice, array, or map.
func lengthFilter(value any, _ ...string) (any, error) {
	v := reflect.ValueOf(value)
	//exhaustive:ignore
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len(), nil
	default:
		return filter.Length(toString(value)), nil
	}
}

// upperFilter converts all characters to uppercase.
func upperFilter(value any, _ ...string) (any, error) {
	return filter.Upper(toString(value)), nil
}

// lowerFilter converts all characters to lowercase.
func lowerFilter(value any, _ ...string) (any, error) {
	return filter.Lower(toString(value)), nil
}

// titleizeFilter capitalizes the first letter of each word.
func titleizeFilter(value any, _ ...string) (any, error) {
	return filter.Titleize(toString(value)), nil
}

// capitalizeFilter capitalizes the first letter of a string.
func capitalizeFilter(value any, _ ...string) (any, error) {
	return filter.Capitalize(toString(value)), nil
}

// camelizeFilter converts a string to camelCase.
func camelizeFilter(value any, _ ...string) (any, error) {
	return filter.Camelize(toString(value)), nil
}

// pascalizeFilter converts a string to PascalCase.
func pascalizeFilter(value any, _ ...string) (any, error) {
	return filter.Pascalize(toString(value)), nil
}

// dasherizeFilter transforms a string into a lowercased, dash-separated format.
func dasherizeFilter(value any, _ ...string) (any, error) {
	return filter.Dasherize(toString(value)), nil
}

// slugifyFilter converts a string into a URL-friendly slug.
func slugifyFilter(value any, _ ...string) (any, error) {
	return filter.Slugify(toString(value)), nil
}

// pluralizeFilter determines the singular or plural form based on a numeric value.
func pluralizeFilter(value any, args ...string) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: pluralize filter requires two arguments (singular and plural forms)", ErrInsufficientArgs)
	}
	count, err := toInteger(value)
	if err != nil {
		return nil, err
	}
	singular, plural := args[0], args[1]
	return filter.Pluralize(count, singular, plural), nil
}

// ordinalizeFilter converts a number to its ordinal English form.
func ordinalizeFilter(value any, _ ...string) (any, error) {
	n, err := toInteger(value)
	if err != nil {
		return nil, err
	}
	return filter.Ordinalize(n), nil
}

// truncateFilter shortens a string to a specified length and appends "...".
func truncateFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: truncate filter requires a length argument", ErrInsufficientArgs)
	}
	maxLen, err := toInteger(args[0])
	if err != nil {
		return nil, err
	}
	return filter.Truncate(toString(value), maxLen), nil
}

// truncateWordsFilter truncates a string to a specified number of words, appending "...".
func truncateWordsFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: truncateWords filter requires a word count argument", ErrInsufficientArgs)
	}
	maxWords, err := toInteger(args[0])
	if err != nil {
		return nil, err
	}
	return filter.TruncateWords(toString(value), maxWords), nil
}
