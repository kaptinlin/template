package template

import (
	"fmt"
	"log"

	"github.com/go-json-experiment/json"
)

func init() {
	// Register all format filters
	filtersToRegister := map[string]FilterFunc{
		"json": jsonFilter,
	}

	for name, filterFunc := range filtersToRegister {
		if err := RegisterFilter(name, filterFunc); err != nil {
			log.Printf("Error registering filter %s: %v", name, err)
		}
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
