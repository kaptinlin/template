package template

import (
	"encoding/json"
	"fmt"
)

func init() {
	RegisterFilter("json", jsonFilter)
}

// jsonFilter serializes an input object into its JSON representation.
func jsonFilter(input interface{}, args ...string) (interface{}, error) {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("error marshaling to JSON: %v", err)
	}
	return string(jsonBytes), nil
}
