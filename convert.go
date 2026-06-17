package template

import (
	"fmt"
	"reflect"
	"strconv"
)

// toString converts the value to its string representation.
func toString(value any) string {
	return newValue(value).String()
}

// toInteger attempts to convert an any to an integer.
func toInteger(input any) (int, error) {
	input = dereferenceIfNeeded(input)
	if v, ok := input.(string); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i, nil
		}
		return 0, fmt.Errorf("%w: unable to parse '%v' as integer", ErrFilterInputNotNumeric, input)
	}

	i, err := newValue(input).Int()
	if err != nil {
		return 0, fmt.Errorf("%w: received %T", ErrFilterInputNotNumeric, input)
	}
	if !int64FitsInInt(i) {
		return 0, fmt.Errorf("%w: %v overflows int", ErrFilterInputNotNumeric, input)
	}
	return int(i), nil
}

// dereferenceIfNeeded checks if the input is a pointer and dereferences it if it's not nil.
func dereferenceIfNeeded(input any) any {
	valRef := reflect.ValueOf(input)
	if valRef.Kind() == reflect.Pointer && !valRef.IsNil() {
		return valRef.Elem().Interface()
	}
	return input
}
