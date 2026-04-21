package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// stringerImpl implements fmt.Stringer for testing toString.
type stringerImpl struct {
	value string
}

func (s stringerImpl) String() string {
	return s.value
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "string value",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with spaces",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "Stringer interface",
			input:    stringerImpl{value: "custom string"},
			expected: "custom string",
		},
		{
			name:     "Stringer interface empty",
			input:    stringerImpl{value: ""},
			expected: "",
		},
		{
			name:     "integer",
			input:    42,
			expected: "42",
		},
		{
			name:     "negative integer",
			input:    -10,
			expected: "-10",
		},
		{
			name:     "zero integer",
			input:    0,
			expected: "0",
		},
		{
			name:     "float64",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "boolean true",
			input:    true,
			expected: "true",
		},
		{
			name:     "boolean false",
			input:    false,
			expected: "false",
		},
		{
			name:     "nil value",
			input:    nil,
			expected: "<nil>",
		},
		{
			name:     "slice of ints",
			input:    []int{1, 2, 3},
			expected: "[1 2 3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToInteger(t *testing.T) {
	ptrInt := 42

	tests := []struct {
		name        string
		input       any
		expected    int
		expectError bool
	}{
		{
			name:        "int value",
			input:       42,
			expected:    42,
			expectError: false,
		},
		{
			name:        "int zero",
			input:       0,
			expected:    0,
			expectError: false,
		},
		{
			name:        "negative int",
			input:       -5,
			expected:    -5,
			expectError: false,
		},
		{
			name:        "float64 truncated",
			input:       3.14,
			expected:    3,
			expectError: false,
		},
		{
			name:        "float64 whole number",
			input:       5.0,
			expected:    5,
			expectError: false,
		},
		{
			name:        "float64 negative",
			input:       -2.9,
			expected:    -2,
			expectError: false,
		},
		{
			name:        "string numeric",
			input:       "123",
			expected:    123,
			expectError: false,
		},
		{
			name:        "string negative",
			input:       "-10",
			expected:    -10,
			expectError: false,
		},
		{
			name:        "string zero",
			input:       "0",
			expected:    0,
			expectError: false,
		},
		{
			name:        "string non-numeric errors",
			input:       "abc",
			expected:    0,
			expectError: true,
		},
		{
			name:        "string float format errors",
			input:       "3.14",
			expected:    0,
			expectError: true,
		},
		{
			name:        "string empty errors",
			input:       "",
			expected:    0,
			expectError: true,
		},
		{
			name:        "boolean type errors",
			input:       true,
			expected:    0,
			expectError: true,
		},
		{
			name:        "nil pointer errors",
			input:       (*int)(nil),
			expected:    0,
			expectError: true,
		},
		{
			name:        "pointer to int dereferences",
			input:       &ptrInt,
			expected:    42,
			expectError: false,
		},
		{
			name:        "slice type errors",
			input:       []int{1, 2},
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toInteger(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDereferenceIfNeeded(t *testing.T) {
	strVal := "test"
	intVal := 123
	floatVal := 3.14

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "non-pointer string unchanged",
			input:    "test",
			expected: "test",
		},
		{
			name:     "pointer to string dereferenced",
			input:    &strVal,
			expected: "test",
		},
		{
			name:     "nil string pointer unchanged",
			input:    (*string)(nil),
			expected: (*string)(nil),
		},
		{
			name:     "non-pointer int unchanged",
			input:    123,
			expected: 123,
		},
		{
			name:     "pointer to int dereferenced",
			input:    &intVal,
			expected: 123,
		},
		{
			name:     "nil int pointer unchanged",
			input:    (*int)(nil),
			expected: (*int)(nil),
		},
		{
			name:     "pointer to float dereferenced",
			input:    &floatVal,
			expected: 3.14,
		},
		{
			name:     "boolean unchanged",
			input:    true,
			expected: true,
		},
		{
			name:     "slice unchanged",
			input:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dereferenceIfNeeded(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
