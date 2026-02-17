package template

import (
	"errors"
	"fmt"

	"github.com/kaptinlin/filter"
)

// registerMapFilters registers all map-related filters.
func registerMapFilters() {
	RegisterFilter("extract", extractFilter)
}

// extractFilter retrieves a nested value from a map, slice, or array using a dot-separated key path.
// It returns an empty string for KeyNotFound and IndexOutOfRange errors to maintain backward compatibility.
func extractFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: extract filter requires a key path argument", ErrInsufficientArgs)
	}

	result, err := filter.Extract(value, args[0])
	if err == nil {
		return result, nil
	}

	if errors.Is(err, filter.ErrKeyNotFound) || errors.Is(err, filter.ErrIndexOutOfRange) {
		return "", nil
	}
	if errors.Is(err, filter.ErrInvalidKeyType) {
		return nil, ErrContextInvalidKeyType
	}
	return nil, err
}
