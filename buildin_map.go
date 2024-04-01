package template

import (
	"errors"
	"fmt"

	"github.com/kaptinlin/filter"
)

func init() {
	// Register the 'extract' filter to handle nested data extraction
	RegisterFilter("extract", extractFilter)
}

// extractFilter retrieves a nested value from a map, slice, or array using a dot-separated key path.
func extractFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: extract filter requires a key path argument", ErrInsufficientArgs)
	}
	keyPath := args[0]
	result, err := filter.Extract(value, keyPath)

	if err != nil {
		if errors.Is(err, filter.ErrKeyNotFound) {
			return nil, ErrContextKeyNotFound
		} else if errors.Is(err, filter.ErrInvalidKeyType) {
			return nil, ErrContextInvalidKeyType
		} else if errors.Is(err, filter.ErrIndexOutOfRange) {
			return nil, ErrContextIndexOutOfRange
		}

		return nil, err
	}
	return result, nil
}
