package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFilter(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "ExtractFromMap",
			template: "{{ data | extract:'user.address.city' }}",
			context: map[string]interface{}{
				"data": map[string]interface{}{
					"user": map[string]interface{}{
						"address": map[string]interface{}{
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
			context: map[string]interface{}{
				"data": []interface{}{"First Element", "Second Element"},
			},
			expected: "First Element",
		},
		{
			name:     "ExtractFromNestedArray",
			template: "{{ data | extract:'1.0' }}",
			context: map[string]interface{}{
				"data": []interface{}{
					[]interface{}{"Nested First Element"},
					[]interface{}{"Nested Second Element"},
				},
			},
			expected: "Nested Second Element",
		},
		{
			name:     "KeyNotFound",
			template: "{{ data | extract:'nonexistent.key' }}",
			context: map[string]interface{}{
				"data": map[string]interface{}{
					"exists": "This exists",
				},
			},
			expected: "", // Returns empty value for backward compatibility
		},
		{
			name:     "IndexOutOfRange",
			template: "{{ data | extract:'2' }}",
			context: map[string]interface{}{
				"data": []interface{}{"First", "Second"},
			},
			expected: "", // Returns empty value for backward compatibility
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the template
			tpl, err := Compile(tc.template)
			require.NoError(t, err, "Failed to parse template")

			// Create a context and add variables
			context := NewContext()
			for k, v := range tc.context {
				context.Set(k, v)
			}

			// Execute the template
			output, err := tpl.Render(map[string]interface{}(context))
			require.NoError(t, err, "Failed to execute template")

			// Verify the output matches the expected result
			assert.Equal(t, tc.expected, output, "Test case '%s'", tc.name)
		})
	}
}
