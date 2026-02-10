package template

import (
	"errors"
	"fmt"

	"github.com/kaptinlin/filter"
)

// registerMapFilters registers all map-related filters.
func registerMapFilters() {
	filters := map[string]FilterFunc{
		"extract": extractFilter,
	}

	for name, fn := range filters {
		RegisterFilter(name, fn)
	}
}

// extractFilter retrieves a nested value from a map, slice, or array using a dot-separated key path.
// Returns empty string for KeyNotFound and IndexOutOfRange errors (backward compatibility).
func extractFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: extract filter requires a key path argument", ErrInsufficientArgs)
	}
	keyPath := args[0]
	result, err := filter.Extract(value, keyPath)

	if err != nil {
		switch {
		case errors.Is(err, filter.ErrKeyNotFound):
			return "", nil // Return empty value instead of error for backward compatibility
		case errors.Is(err, filter.ErrIndexOutOfRange):
			return "", nil // Return empty value instead of error for backward compatibility
		case errors.Is(err, filter.ErrInvalidKeyType):
			return nil, ErrContextInvalidKeyType
		}

		return nil, err
	}
	return result, nil
}
