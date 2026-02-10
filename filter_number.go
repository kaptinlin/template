package template

import (
	"fmt"

	"github.com/kaptinlin/filter"
)

// registerNumberFilters registers all number-related filters.
func registerNumberFilters() {
	RegisterFilter("number", numberFilter)
	RegisterFilter("bytes", bytesFilter)
}

// numberFilter formats a numeric value according to the specified format string.
func numberFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("number requires a format string: %w", ErrInsufficientArgs)
	}
	format := args[0]
	return filter.Number(value, format)
}

// bytesFilter converts a numeric value into a human-readable byte format.
func bytesFilter(value interface{}, _ ...string) (interface{}, error) {
	return filter.Bytes(value)
}
