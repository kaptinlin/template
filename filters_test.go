package template

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// mockToUpper is a simple filter function that converts a string to uppercase.
func mockToUpper(value interface{}, args ...string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, ErrFilterInputInvalid
	}
	return strings.ToUpper(str), nil
}

// mockAppend is a filter function that appends given args to the string.
func mockAppend(value interface{}, args ...string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, ErrFilterInputInvalid
	}
	if len(args) < 1 {
		return nil, ErrInsufficientArgs
	}
	return fmt.Sprintf("%s%s", str, args[0]), nil
}

func TestApplyFilters(t *testing.T) {
	// Register mock filters for testing
	err := RegisterFilter("mockToUpper", mockToUpper)
	if err != nil {
		t.Fatalf("Failed to register filter: %v", err)
	}
	err = RegisterFilter("mockAppend", mockAppend)
	if err != nil {
		t.Fatalf("Failed to register filter: %v", err)
	}

	cases := []struct {
		name     string
		value    interface{}
		filters  []Filter
		expected interface{}
		err      error
	}{
		{
			name:     "SingleFilterToUpper",
			value:    "hello",
			filters:  []Filter{{Name: "mockToUpper", Args: []string{}}},
			expected: "HELLO",
			err:      nil,
		},
		{
			name:     "MultipleFiltersAppendThenToUpper",
			value:    "hello",
			filters:  []Filter{{Name: "mockAppend", Args: []string{" world"}}, {Name: "mockToUpper", Args: []string{}}},
			expected: "HELLO WORLD",
			err:      nil,
		},
		{
			name:     "FilterNotFound",
			value:    "test",
			filters:  []Filter{{Name: "nonexistent", Args: []string{}}},
			expected: "test",
			err:      ErrFilterNotFound,
		},
		{
			name:     "FilterInvalidInput",
			value:    123,
			filters:  []Filter{{Name: "mockToUpper", Args: []string{}}},
			expected: nil,
			err:      ErrFilterInputInvalid,
		},
		{
			name:     "FilterInsufficientArgs",
			value:    "test",
			filters:  []Filter{{Name: "mockAppend", Args: []string{}}},
			expected: nil,
			err:      ErrInsufficientArgs,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ApplyFilters(tc.value, tc.filters, NewContext())

			// Check for expected error
			if !errors.Is(err, tc.err) {
				t.Errorf("%s: expected error %v, got %v", tc.name, tc.err, err)
			}

			// Check for expected result
			if result != tc.expected {
				t.Errorf("%s: expected result %v, got %v", tc.name, tc.expected, result)
			}
		})
	}
}
