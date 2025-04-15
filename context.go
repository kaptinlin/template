package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/kaptinlin/filter"
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
	value = recursivelyProcessValue(value)

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

// structToMap converts a struct to map[string]interface{}, facilitating dot-notation access in templates.
// Uses JSON marshaling and unmarshaling for conversion, utilizing json tags for field names.
// If conversion fails, the original object is returned.
func structToMap(obj interface{}) interface{} {
	v := reflect.ValueOf(obj)

	// If not a struct type, return the original value
	if v.Kind() != reflect.Struct {
		return obj
	}

	// First serialize to JSON
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		// Serialization failed, return the original object
		return obj
	}

	// Then deserialize to map
	var resultMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &resultMap); err != nil {
		// Deserialization failed, return the original object
		return obj
	}

	return resultMap
}

// recursivelyProcessValue is the entry function for recursive value processing.
// Converts structs to maps while keeping other types unchanged, for easier template access.
// Returns nil directly for nil values.
func recursivelyProcessValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	// Process the value and return the result
	return processValueDeep(value)
}

// processValueDeep deeply processes values, maintaining their original structure, only converting structs to maps.
// Specially handles complex types like pointers, structs, slices/arrays, and maps.
// Basic types remain unchanged and are returned directly.
func processValueDeep(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)

	// Handle pointers, get the value they point to
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Special handling for time type, avoid converting time.Time to map
	if t, ok := v.Interface().(time.Time); ok {
		return t
	}

	//nolint:exhaustive
	switch v.Kind() {
	case reflect.Struct:
		// Check again if it's a time.Time type, avoid converting it to map
		if v.Type().String() == "time.Time" {
			return v.Interface()
		}
		// Convert struct to map for easier dot-notation access in templates
		return structToMap(v.Interface())

	case reflect.Slice, reflect.Array:
		// Process slices/arrays, maintaining the original type
		length := v.Len()
		// Create a new slice of the same type
		newSlice := reflect.MakeSlice(v.Type(), length, length)

		// Process each element
		for i := 0; i < length; i++ {
			elemValue := v.Index(i).Interface()
			processedElem := processValueDeep(elemValue)

			// Set the element in the new slice
			if reflect.TypeOf(processedElem) == reflect.TypeOf(elemValue) {
				// Same type, set directly
				newSlice.Index(i).Set(reflect.ValueOf(processedElem))
			} else {
				// Type changed (e.g., struct->map), create a new interface{} slice
				// Because the original slice type might not accommodate map types
				interfaceSlice := make([]interface{}, length)
				for j := 0; j < length; j++ {
					if j == i {
						interfaceSlice[j] = processedElem
					} else {
						// Process other elements
						interfaceSlice[j] = processValueDeep(v.Index(j).Interface())
					}
				}
				return interfaceSlice
			}
		}
		return newSlice.Interface()

	case reflect.Map:
		// Check if the map contains struct values
		containsStruct := false
		mapIter := v.MapRange()
		for mapIter.Next() {
			mapVal := mapIter.Value()
			valKind := getBaseKind(mapVal)
			if valKind == reflect.Struct {
				containsStruct = true
				break
			}
		}

		// Choose processing strategy based on whether the map contains structs
		return processMap(v, containsStruct)

	default:
		// Basic types remain unchanged
		return value
	}
}

// processMap processes map-type values, choosing different strategies based on whether they contain structs.
// Parameters:
//   - v: The reflection value of the map to process
//   - containsStruct: Whether the map contains struct values
//
// Returns:
//   - If it contains structs, returns a converted map[string]interface{}
//   - Otherwise, attempts to maintain the original map type
func processMap(v reflect.Value, containsStruct bool) interface{} {
	// If it contains structs, convert directly to map[string]interface{}
	if containsStruct {
		return convertMapToStringMap(v)
	}

	// No structs, try to maintain the original map type
	return processRegularMap(v)
}

// convertMapToStringMap converts any map to map[string]interface{}.
// Suitable for maps containing structs, ensuring keys are converted to strings and values are recursively processed.
// Parameters:
//   - v: The reflection value of the map to convert
//
// Returns: The converted map[string]interface{}
func convertMapToStringMap(v reflect.Value) map[string]interface{} {
	resultMap := make(map[string]interface{})
	mapIter := v.MapRange()

	for mapIter.Next() {
		mapKey := mapIter.Key()
		mapVal := mapIter.Value().Interface()
		keyStr := fmt.Sprint(mapKey.Interface())

		// Recursively process the value
		processedVal := processValueDeep(mapVal)
		resultMap[keyStr] = processedVal
	}

	return resultMap
}

// processRegularMap processes maps that don't contain structs, attempting to maintain the original map type.
// Recursively processes each value in the map. If the processed value type matches the original type,
// maintains the original map type; otherwise, converts the entire map to map[string]interface{}.
// Parameters:
//   - v: The reflection value of the map to process
//
// Returns: The processed map, which may be of the original type or map[string]interface{}
func processRegularMap(v reflect.Value) interface{} {
	newMap := reflect.MakeMap(v.Type())
	mapIter := v.MapRange()

	for mapIter.Next() {
		mapKey := mapIter.Key()
		mapVal := mapIter.Value().Interface()

		// Recursively process the value
		processedVal := processValueDeep(mapVal)

		// Check if the processed value type has changed
		if reflect.TypeOf(processedVal) == reflect.TypeOf(mapVal) {
			// Same type, set to the new map
			newMap.SetMapIndex(mapKey, reflect.ValueOf(processedVal))
		} else {
			// Type changed, need to convert the entire map to map[string]interface{}
			return convertEntireMapDueToTypeChange(v, mapKey, processedVal)
		}
	}

	return newMap.Interface()
}

// convertEntireMapDueToTypeChange converts the entire map to map[string]interface{} when a value's type changes.
// This is necessary because Go's type system doesn't allow a strongly-typed map to contain values of different types.
// Parameters:
//   - v: The reflection value of the original map
//   - changedKey: The key whose value type has changed
//   - processedVal: The new processed value
//
// Returns: The converted map[string]interface{}
func convertEntireMapDueToTypeChange(v reflect.Value, changedKey reflect.Value, processedVal interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{})
	mapIter := v.MapRange()

	for mapIter.Next() {
		k := mapIter.Key()
		val := mapIter.Value().Interface()
		keyStr := fmt.Sprint(k.Interface())

		if k.Interface() == changedKey.Interface() {
			resultMap[keyStr] = processedVal
		} else {
			resultMap[keyStr] = processValueDeep(val)
		}
	}

	return resultMap
}

// getBaseKind gets the basic Kind type of a value, handling pointer cases.
// If it's a pointer type, recursively gets the Kind of the value it points to.
// Parameters:
//   - v: The reflection value to get the Kind for
//
// Returns: The basic Kind type of the value
func getBaseKind(v reflect.Value) reflect.Kind {
	kind := v.Kind()
	// If it's a pointer, get the Kind of the element it points to
	if kind == reflect.Ptr {
		if v.IsNil() {
			return reflect.Ptr
		}
		return getBaseKind(v.Elem())
	}
	return kind
}

// Get retrieves a value from the Context for the specified key, supporting nested key access.
// Uses filter.Extract to handle complex key paths, such as array indices and nested properties.
// Parameters:
//   - key: The key to retrieve, can be a dot-separated nested key like "user.profile.name"
//
// Returns:
//   - interface{}: The retrieved value
//   - error: Returns an appropriate error if the key doesn't exist or access fails
func (c Context) Get(key string) (interface{}, error) {
	value, err := filter.Extract(c, key)
	if err != nil {
		switch {
		case errors.Is(err, filter.ErrKeyNotFound):
			return nil, fmt.Errorf("%w: key '%s' not found in context", ErrContextKeyNotFound, key)
		case errors.Is(err, filter.ErrInvalidKeyType):
			return nil, fmt.Errorf("%w: invalid type for key '%s', cannot navigate", ErrContextInvalidKeyType, key)
		case errors.Is(err, filter.ErrIndexOutOfRange):
			return nil, fmt.Errorf("%w: index out of range for key '%s'", ErrContextIndexOutOfRange, key)
		}

		return nil, fmt.Errorf("unknown error while accessing key '%s': %w", key, err)
	}
	return value, nil
}
