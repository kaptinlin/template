package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMathFilters(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "AbsFilterPositive",
			template: "{{ value | abs }}",
			context:  map[string]interface{}{"value": -42},
			expected: "42",
		},
		{
			name:     "AtLeastFilter",
			template: "{{ value | atLeast:10 }}",
			context:  map[string]interface{}{"value": 5},
			expected: "10",
		},
		{
			name:     "AtMostFilter",
			template: "{{ value | atMost:10 }}",
			context:  map[string]interface{}{"value": 15},
			expected: "10",
		},
		{
			name:     "RoundFilter",
			template: "{{ value | round:2 }}",
			context:  map[string]interface{}{"value": 3.14159},
			expected: "3.14",
		},
		{
			name:     "FloorFilter",
			template: "{{ value | floor }}",
			context:  map[string]interface{}{"value": 3.99},
			expected: "3",
		},
		{
			name:     "CeilFilter",
			template: "{{ value | ceil }}",
			context:  map[string]interface{}{"value": 3.01},
			expected: "4",
		},
		{
			name:     "PlusFilter",
			template: "{{ value | plus:3 }}",
			context:  map[string]interface{}{"value": 7},
			expected: "10",
		},
		{
			name:     "MinusFilter",
			template: "{{ value | minus:2 }}",
			context:  map[string]interface{}{"value": 10},
			expected: "8",
		},
		{
			name:     "TimesFilter",
			template: "{{ value | times:2 }}",
			context:  map[string]interface{}{"value": 5},
			expected: "10",
		},
		{
			name:     "DivideFilter",
			template: "{{ value | divide:4 }}",
			context:  map[string]interface{}{"value": 20},
			expected: "5",
		},
		{
			name:     "ModuloFilter",
			template: "{{ value | modulo:3 }}",
			context:  map[string]interface{}{"value": 10},
			expected: "1",
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
