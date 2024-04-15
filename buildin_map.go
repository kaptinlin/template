package template

import (
	"errors"
	"fmt"
	"log"

	"github.com/kaptinlin/filter"
)

func init() {
	// Register the 'extract' filter to handle nested data extraction
	filtersToRegister := map[string]FilterFunc{
		"extract": extractFilter,
	}

	for name, filterFunc := range filtersToRegister {
		if err := RegisterFilter(name, filterFunc); err != nil {
			log.Printf("Error registering filter %s: %v", name, err)
		}
	}
}

// extractFilter retrieves a nested value from a map, slice, or array using a dot-separated key path.
func extractFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: extract filter requires a key path argument", ErrInsufficientArgs)
	}
	keyPath := args[0]
	result, err := filter.Extract(value, keyPath)

	if err != nil {
		switch {
		case errors.Is(err, filter.ErrKeyNotFound):
			return nil, ErrContextKeyNotFound
		case errors.Is(err, filter.ErrInvalidKeyType):
			return nil, ErrContextInvalidKeyType
		case errors.Is(err, filter.ErrIndexOutOfRange):
			return nil, ErrContextIndexOutOfRange
		}

		return nil, err
	}
	return result, nil
}
