package template

import (
	"fmt"
	"log"

	"github.com/kaptinlin/filter"
)

func init() {
	// Register number-related filters
	filtersToRegister := map[string]FilterFunc{
		"number": numberFilter,
		"bytes":  bytesFilter,
	}

	for name, filterFunc := range filtersToRegister {
		if err := RegisterFilter(name, filterFunc); err != nil {
			log.Printf("Error registering filter %s: %v", name, err)
		}
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
func bytesFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Bytes(value)
}
