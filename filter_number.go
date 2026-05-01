package template

import (
	"fmt"

	"github.com/kaptinlin/filter"
)

func registerNumberFilters() {
	defaultRegistry.MustRegister("number", numberFilter)
	defaultRegistry.MustRegister("bytes", bytesFilter)
}

func numberFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: number filter requires a format string", ErrInsufficientArgs)
	}
	return filter.Number(value, toString(args[0]))
}

func bytesFilter(value any, _ ...any) (any, error) {
	return filter.Bytes(value)
}
