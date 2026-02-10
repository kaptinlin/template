package template

import (
	"fmt"

	"github.com/kaptinlin/filter"
)

// registerNumberFilters registers all number-related filters.
func registerNumberFilters() {
	filters := map[string]FilterFunc{
		"number": numberFilter,
		"bytes":  bytesFilter,
	}

	for name, fn := range filters {
		RegisterFilter(name, fn)
	}
}

// numberFilter formats a numeric value according to the specified format string.
func numberFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: number filter requires a format string", ErrInsufficientArgs)
	}
	format := args[0]
	return filter.Number(value, format)
}

// bytesFilter converts a numeric value into a human-readable byte format.
func bytesFilter(value interface{}, _ ...string) (interface{}, error) {
	return filter.Bytes(value)
}
