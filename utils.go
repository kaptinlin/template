package template

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/golang-module/carbon/v2"
)

var (
	// ErrFilterInputNotSlice indicates a filter expected a slice but received a different type.
	ErrFilterInputNotSlice = errors.New("filter input is not a slice")

	// ErrFilterInputNotNumeric indicates a filter expected a numeric value but received a different type.
	ErrFilterInputNotNumeric = errors.New("filter input is not numeric")

	// ErrFilterInputInvalidTimeFormat indicates a filter expected a valid time format but didn't receive it.
	ErrFilterInputInvalidTimeFormat = errors.New("filter input has an invalid time format")

	// ErrFilterInputUnsupportedType indicates the filter received a type it does not support.
	ErrFilterInputUnsupportedType = errors.New("filter input is of an unsupported type")
)

// toSlice attempts to convert an interface{} to a slice of interface{}.
func toSlice(input interface{}) ([]interface{}, error) {
	input = dereferenceIfNeeded(input)
	valRef := reflect.ValueOf(input)
	if valRef.Kind() != reflect.Slice {
		return nil, fmt.Errorf("%w: expected slice, got %T", ErrFilterInputNotSlice, input)
	}

	var result []interface{}
	for i := 0; i < valRef.Len(); i++ {
		result = append(result, valRef.Index(i).Interface())
	}
	return result, nil
}

func toFloat64(input interface{}) (float64, error) {
	input = dereferenceIfNeeded(input)
	switch v := input.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			// Here, it's important to include the original error message to provide more context.
			return 0, fmt.Errorf("%w: unable to parse '%v' as float64, error: %v", ErrFilterInputNotNumeric, input, err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("%w: received %T", ErrFilterInputNotNumeric, input)
	}
}

// Helper function to ensure the value is a string
func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// toInteger attempts to convert an interface{} to an integer.
func toInteger(input interface{}) (int, error) {
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

// toCarbon converts an input of type interface{} to a carbon.Carbon object.
func toCarbon(input interface{}) (carbon.Carbon, error) {
	input = dereferenceIfNeeded(input)
	switch v := input.(type) {
	case carbon.Carbon:
		return v, nil
	case time.Time:
		return carbon.CreateFromStdTime(v), nil
	case string:
		parsedTime := carbon.Parse(v)
		if parsedTime.Error != nil {
			return carbon.Carbon{}, fmt.Errorf("%w: %s", ErrFilterInputInvalidTimeFormat, parsedTime.Error)
		}
		return parsedTime, nil
	default:
		return carbon.Carbon{}, fmt.Errorf("%w: %T", ErrFilterInputUnsupportedType, input)
	}
}

// dereferenceIfNeeded checks if the input is a pointer and dereferences it if it's not nil.
func dereferenceIfNeeded(input interface{}) interface{} {
	valRef := reflect.ValueOf(input)
	if valRef.Kind() == reflect.Ptr && !valRef.IsNil() {
		return valRef.Elem().Interface()
	}
	return input
}
