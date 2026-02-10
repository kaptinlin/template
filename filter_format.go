package template

import (
	"fmt"

	"github.com/go-json-experiment/json"
)

// registerFormatFilters registers all format-related filters.
func registerFormatFilters() {
	RegisterFilter("json", jsonFilter)
}

// jsonFilter serializes an input object into its JSON representation.
func jsonFilter(input any, _ ...string) (any, error) {
	// Use deterministic mode to ensure consistent key ordering
	b, err := json.Marshal(input, json.Deterministic(true))
	if err != nil {
		return nil, fmt.Errorf("marshal to JSON: %w", err)
	}
	return string(b), nil
}
