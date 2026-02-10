package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNumberFilters(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "NumberFilterWithFormat",
			template: "{{ value | number:'#,###.##' }}",
			context:  map[string]interface{}{"value": 1234567.89},
			expected: "1,234,567.89",
		},
		{
			name:     "BytesFilterForKilobytes",
			template: "{{ value | bytes }}",
			context:  map[string]interface{}{"value": 1024},
			expected: "1.0 kB",
		},
		{
			name:     "BytesFilterForMegabytes",
			template: "{{ value | bytes }}",
			context:  map[string]interface{}{"value": 1048576},
			expected: "1.0 MB",
		},
		{
			name:     "BytesFilterForGigabytes",
			template: "{{ value | bytes }}",
			context:  map[string]interface{}{"value": 1073741824},
			expected: "1.1 GB",
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
