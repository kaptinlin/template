package template

import (
	"fmt"

	"github.com/go-json-experiment/json"
)

// registerFormatFilters registers all format-related filters.
func registerFormatFilters() {
	filters := map[string]FilterFunc{
		"json": jsonFilter,
	}

	for name, fn := range filters {
		RegisterFilter(name, fn)
	}
}

// jsonFilter serializes an input object into its JSON representation.
func jsonFilter(input interface{}, _ ...string) (interface{}, error) {
	// Use deterministic mode to ensure consistent key ordering
	jsonBytes, err := json.Marshal(input, json.Deterministic(true))
	if err != nil {
		return nil, fmt.Errorf("error marshaling to JSON: %w", err)
	}
	return string(jsonBytes), nil
}
