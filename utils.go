package template

import (
	"fmt"
	"reflect"
	"strconv"
)

// toString converts the value to its string representation.
func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// toInteger attempts to convert an any to an integer.
func toInteger(input any) (int, error) {
	input = dereferenceIfNeeded(input)
	switch v := input.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, nil
		}
		return 0, fmt.Errorf("%w: unable to parse '%v' as integer", ErrFilterInputNotNumeric, input)
	default:
		return 0, fmt.Errorf("%w: received %T", ErrFilterInputNotNumeric, input)
	}
}

// dereferenceIfNeeded checks if the input is a pointer and dereferences it if it's not nil.
func dereferenceIfNeeded(input any) any {
	valRef := reflect.ValueOf(input)
	if valRef.Kind() == reflect.Ptr && !valRef.IsNil() {
		return valRef.Elem().Interface()
	}
	return input
}
