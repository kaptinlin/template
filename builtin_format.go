package template

import (
	"fmt"

	"github.com/go-json-experiment/json"
)

func init() {
	mustRegisterFilters(map[string]FilterFunc{
		"json": jsonFilter,
	})
}

// jsonFilter serializes an input object into its JSON representation.
func jsonFilter(input any, _ ...string) (any, error) {
	// Use deterministic mode to ensure consistent key ordering
	jsonBytes, err := json.Marshal(input, json.Deterministic(true))
	if err != nil {
		return nil, fmt.Errorf("json marshal failed: %w", err)
	}
	return string(jsonBytes), nil
}
