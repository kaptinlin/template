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

// toFloat64 attempts to convert an any value to float64.
func toFloat64(input any) (float64, error) {
	input = dereferenceIfNeeded(input)
	switch v := input.(type) {
	case int:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, nil
		}
		return 0, fmt.Errorf("%w: unable to parse '%v' as float", ErrFilterInputNotNumeric, input)
	default:
		return 0, fmt.Errorf("%w: received %T", ErrFilterInputNotNumeric, input)
	}
}

// toSlice converts an any value to []any using reflection.
func toSlice(input any) ([]any, error) {
	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, fmt.Errorf("%w: got %T", ErrExpectedSliceOrArray, input)
	}
	result := make([]any, v.Len())
	for i := range v.Len() {
		result[i] = v.Index(i).Interface()
	}
	return result, nil
}

// dereferenceIfNeeded checks if the input is a pointer and dereferences it if it's not nil.
func dereferenceIfNeeded(input any) any {
	valRef := reflect.ValueOf(input)
	if valRef.Kind() == reflect.Pointer && !valRef.IsNil() {
		return valRef.Elem().Interface()
	}
	return input
}
