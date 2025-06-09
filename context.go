package template

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kaptinlin/jsonpointer"
)

// Context stores template variables in a map structure, used for passing and accessing variables during template execution.
// Keys are strings, and values can be of any type, supporting dot-notation (.) for nested access.
type Context map[string]interface{}

// NewContext creates and returns a new empty Context instance.
// Example usage: ctx := NewContext()
func NewContext() Context {
	return make(Context)
}

// Set inserts a value into the Context with the specified key, supporting dot-notation (.) for nested keys.
// For example: ctx.Set("user.profile.name", "John") creates a nested map structure.
// Values are recursively processed, with structs converted to maps to enable dot-notation access in templates.
func (c Context) Set(key string, value interface{}) {
	// Process the value, ensuring structs are converted to maps

	// Split the key by dot to create nested structure
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		c[key] = value
		return
	}

	// Create or update nested structure
	var currentMap map[string]interface{} = c
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part, set the actual value
			currentMap[part] = value
		} else {
			// Intermediate parts, ensure they are map types
			if _, exists := currentMap[part]; !exists {
				currentMap[part] = make(map[string]interface{})
			} else if _, ok := currentMap[part].(map[string]interface{}); !ok {
				// If exists but not a map type, override with a new map
				currentMap[part] = make(map[string]interface{})
			}
			// Descend to the next level map
			currentMap = currentMap[part].(map[string]interface{})
		}
	}
}

// Get retrieves a value from the Context for the specified key, supporting nested key access.
// Uses jsonpointer.Get to handle complex key paths, such as array indices and nested properties.
// Parameters:
//   - key: The key to retrieve, can be a dot-separated nested key like "user.profile.name"
//
// Returns:
//   - interface{}: The retrieved value
//   - error: Returns an appropriate error if the key doesn't exist or access fails
func (c Context) Get(key string) (interface{}, error) {
	// Handle empty key
	if key == "" {
		return map[string]interface{}(c), nil
	}

	// Convert dot notation to path components
	parts := convertDotToPath(key)

	// Validate each part (basic validation before calling jsonpointer)
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("%w: empty path component in key '%s'", ErrContextInvalidKeyType, key)
		}
	}

	// Use jsonpointer.Get with variadic arguments (more efficient than string concatenation)
	value, err := jsonpointer.Get(c, parts...)
	if err != nil {
		switch {
		case errors.Is(err, jsonpointer.ErrNotFound) || errors.Is(err, jsonpointer.ErrKeyNotFound) || errors.Is(err, jsonpointer.ErrFieldNotFound):
			return nil, fmt.Errorf("%w: key '%s' not found in context", ErrContextKeyNotFound, key)
		case errors.Is(err, jsonpointer.ErrIndexOutOfBounds) || errors.Is(err, jsonpointer.ErrInvalidIndex):
			return nil, fmt.Errorf("%w: index out of range for key '%s'", ErrContextIndexOutOfRange, key)
		case errors.Is(err, jsonpointer.ErrInvalidPath) || errors.Is(err, jsonpointer.ErrInvalidPathStep):
			return nil, fmt.Errorf("%w: invalid type for key '%s', cannot navigate", ErrContextInvalidKeyType, key)
		}

		return nil, fmt.Errorf("unknown error while accessing key '%s': %w", key, err)
	}
	return value, nil
}

// convertDotToPath converts dot notation to path components array
func convertDotToPath(dotNotation string) []string {
	if dotNotation == "" {
		return []string{}
	}
	return strings.Split(dotNotation, ".")
}
