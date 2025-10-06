package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDereferenceIfNeeded(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		input       interface{}
		expectedNil bool // Expect the dereferenced value to be nil
	}{
		{
			name:        "Non-pointer string",
			input:       "test",
			expectedNil: false,
		},
		{
			name:        "Pointer to string",
			input:       new(string),
			expectedNil: false,
		},
		{
			name:        "Nil pointer to string",
			input:       (*string)(nil),
			expectedNil: true,
		},
		{
			name:        "Non-pointer int",
			input:       123,
			expectedNil: false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			derefed := dereferenceIfNeeded(tc.input)
			if tc.expectedNil {
				assert.Nil(t, derefed)
			} else {
				assert.NotNil(t, derefed)
				// Additional type assertion if needed
				if _, ok := tc.input.(*string); ok {
					_, ok := derefed.(string)
					assert.True(t, ok, "Expected a string type after dereferencing")
				}
			}
		})
	}
}
