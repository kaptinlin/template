package template

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockToUpper is a simple filter function that converts a string to uppercase.
func mockToUpper(value interface{}, _ ...string) (interface{}, error) {
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
	require.NoError(t, err, "Failed to register filter")
	err = RegisterFilter("mockAppend", mockAppend)
	require.NoError(t, err, "Failed to register filter")

	// Create a simple context for testing VariableArg
	ctx := NewContext()
	ctx.Set("user", "John Doe")

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
			filters:  []Filter{{Name: "mockToUpper", Args: []FilterArg{}}},
			expected: "HELLO",
			err:      nil,
		},
		{
			name:  "MultipleFiltersAppendThenToUpper",
			value: "hello",
			filters: []Filter{
				{Name: "mockAppend", Args: []FilterArg{StringArg{val: " world"}}},
				{Name: "mockToUpper", Args: []FilterArg{}},
			},
			expected: "HELLO WORLD",
			err:      nil,
		},
		{
			name:     "FilterNotFound",
			value:    "test",
			filters:  []Filter{{Name: "nonexistent", Args: []FilterArg{}}},
			expected: "test",
			err:      ErrFilterNotFound,
		},
		{
			name:     "FilterInvalidInput",
			value:    123,
			filters:  []Filter{{Name: "mockToUpper", Args: []FilterArg{}}},
			expected: nil,
			err:      ErrFilterInputInvalid,
		},
		{
			name:     "FilterInsufficientArgs",
			value:    "test",
			filters:  []Filter{{Name: "mockAppend", Args: []FilterArg{}}},
			expected: nil,
			err:      ErrInsufficientArgs,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ApplyFilters(tc.value, tc.filters, ctx)

			// Check for expected error
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err, "%s: expected error", tc.name)
			} else {
				assert.NoError(t, err, "%s: unexpected error", tc.name)
			}

			// Check for expected result
			assert.Equal(t, tc.expected, result, "%s: result mismatch", tc.name)
		})
	}
}
