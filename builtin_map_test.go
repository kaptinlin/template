package template

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFilter(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "ExtractFromMap",
			template: "{{ data | extract:'user.address.city' }}",
			context: map[string]any{
				"data": map[string]any{
					"user": map[string]any{
						"address": map[string]any{
							"city": "New York",
						},
					},
				},
			},
			expected: "New York",
		},
		{
			name:     "ExtractFromArray",
			template: "{{ data | extract:'0' }}",
			context: map[string]any{
				"data": []any{"First Element", "Second Element"},
			},
			expected: "First Element",
		},
		{
			name:     "ExtractFromNestedArray",
			template: "{{ data | extract:'1.0' }}",
			context: map[string]any{
				"data": []any{
					[]any{"Nested First Element"},
					[]any{"Nested Second Element"},
				},
			},
			expected: "Nested Second Element",
		},
		{
			name:     "KeyNotFound",
			template: "{{ data | extract:'nonexistent.key' }}",
			context: map[string]any{
				"data": map[string]any{
					"exists": "This exists",
				},
			},
			expected: "ErrContextKeyNotFound",
		},
		{
			name:     "IndexOutOfRange",
			template: "{{ data | extract:'2' }}",
			context: map[string]any{
				"data": []any{"First", "Second"},
			},
			expected: "ErrContextIndexOutOfRange",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the template
			tpl, err := Parse(tc.template)
			require.NoError(t, err, "Failed to parse template")

			// Create a context and add variables
			context := NewContext()
			for k, v := range tc.context {
				context.Set(k, v)
			}

			// Execute the template
			output, err := Execute(tpl, context)
			var finalOutput string
			if err != nil {
				switch {
				case errors.Is(err, ErrContextKeyNotFound):
					finalOutput = "ErrContextKeyNotFound"
				case errors.Is(err, ErrContextIndexOutOfRange):
					finalOutput = "ErrContextIndexOutOfRange"
				default:
					require.Fail(t, "Unexpected error during execution", err)
				}
			} else {
				finalOutput = output
			}

			// Verify the output matches the expected result
			assert.Equal(t, tc.expected, finalOutput, "Test case '%s'", tc.name)
		})
	}
}
