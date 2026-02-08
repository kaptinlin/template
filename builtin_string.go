package template

import (
	"fmt"

	"github.com/kaptinlin/filter"
)

func init() {
	mustRegisterFilters(map[string]FilterFunc{
		"default":       defaultFilter,
		"trim":          trimFilter,
		"split":         splitFilter,
		"replace":       replaceFilter,
		"remove":        removeFilter,
		"append":        appendFilter,
		"prepend":       prependFilter,
		"length":        lengthFilter,
		"upper":         upperFilter,
		"lower":         lowerFilter,
		"titleize":      titleizeFilter,
		"capitalize":    capitalizeFilter,
		"camelize":      camelizeFilter,
		"pascalize":     pascalizeFilter,
		"dasherize":     dasherizeFilter,
		"slugify":       slugifyFilter,
		"pluralize":     pluralizeFilter,
		"ordinalize":    ordinalizeFilter,
		"truncate":      truncateFilter,
		"truncateWords": truncateWordsFilter,
	})
}

// defaultFilter returns a default value if the input string is empty.
func defaultFilter(value any, args ...string) (any, error) {
	defaultValue := ""
	if len(args) > 0 {
		defaultValue = args[0]
	}
	return filter.Default(toString(value), defaultValue), nil
}

// trimFilter removes leading and trailing whitespace from the string.
func trimFilter(value any, _ ...string) (any, error) {
	return filter.Trim(toString(value)), nil
}

// splitFilter divides a string into a slice of strings based on a specified delimiter.
func splitFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: split filter requires a delimiter argument", ErrInsufficientArgs)
	}
	delimiter := args[0]
	return filter.Split(toString(value), delimiter), nil
}

// replaceFilter substitutes all instances of a specified substring with another string.
func replaceFilter(value any, args ...string) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: replace filter requires two arguments (old and new substrings)", ErrInsufficientArgs)
	}
	oldStr, newStr := args[0], args[1]
	return filter.Replace(toString(value), oldStr, newStr), nil
}

// removeFilter eliminates all occurrences of a specified substring.
func removeFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: remove filter requires a substring argument", ErrInsufficientArgs)
	}
	toRemove := args[0]
	return filter.Remove(toString(value), toRemove), nil
}

// appendFilter adds characters to the end of a string.
func appendFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: append filter requires a string to append", ErrInsufficientArgs)
	}
	toAppend := args[0]
	return filter.Append(toString(value), toAppend), nil
}

// prependFilter adds characters to the beginning of a string.
func prependFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: prepend filter requires a string to prepend", ErrInsufficientArgs)
	}
	toPrepend := args[0]
	return filter.Prepend(toString(value), toPrepend), nil
}

// lengthFilter returns the number of characters in a string.
func lengthFilter(value any, _ ...string) (any, error) {
	return filter.Length(toString(value)), nil
}

// upperFilter converts all characters in a string to uppercase.
func upperFilter(value any, _ ...string) (any, error) {
	return filter.Upper(toString(value)), nil
}

// lowerFilter converts all characters in a string to lowercase.
func lowerFilter(value any, _ ...string) (any, error) {
	return filter.Lower(toString(value)), nil
}

// titleizeFilter capitalizes the first letter of each word in a string.
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

// slugifyFilter converts a string into a URL-friendly "slug".
func slugifyFilter(value any, _ ...string) (any, error) {
	return filter.Slugify(toString(value)), nil
}

// pluralizeFilter determines the singular or plural form of a word based on a numeric value.
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
	number, err := toInteger(value)
	if err != nil {
		return nil, err
	}
	return filter.Ordinalize(number), nil
}

// truncateFilter shortens a string to a specified length and appends "..." if it exceeds that length.
func truncateFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: truncate filter requires a length argument", ErrInsufficientArgs)
	}
	maxLength, err := toInteger(args[0])
	if err != nil {
		return nil, err
	}
	return filter.Truncate(toString(value), maxLength), nil
}

// truncateWordsFilter truncates a string to a specified number of words, appending "..." if it exceeds that limit.
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
