package template

import (
	"fmt"

	"github.com/kaptinlin/filter"
)

// registerNumberFilters registers all number-related filters.
func registerNumberFilters() {
	defaultRegistry.MustRegister("number", numberFilter)
	defaultRegistry.MustRegister("bytes", bytesFilter)
}

// numberFilter formats a numeric value according to the specified format string.
func numberFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: number filter requires a format string", ErrInsufficientArgs)
	}
	return filter.Number(value, toString(args[0]))
}

// bytesFilter converts a numeric value into a human-readable byte format.
func bytesFilter(value any, _ ...any) (any, error) {
	return filter.Bytes(value)
}
