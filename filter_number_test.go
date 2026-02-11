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
		context  map[string]any
		expected string
	}{
		{
			name:     "NumberFilterWithFormat",
			template: "{{ value | number:'#,###.##' }}",
			context:  map[string]any{"value": 1234567.89},
			expected: "1,234,567.89",
		},
		{
			name:     "BytesFilterForKilobytes",
			template: "{{ value | bytes }}",
			context:  map[string]any{"value": 1024},
			expected: "1.0 kB",
		},
		{
			name:     "BytesFilterForMegabytes",
			template: "{{ value | bytes }}",
			context:  map[string]any{"value": 1048576},
			expected: "1.0 MB",
		},
		{
			name:     "BytesFilterForGigabytes",
			template: "{{ value | bytes }}",
			context:  map[string]any{"value": 1073741824},
			expected: "1.1 GB",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := Compile(tc.template)
			require.NoError(t, err)

			ctx := NewContext()
			for k, v := range tc.context {
				ctx.Set(k, v)
			}

			got, err := tpl.Render(map[string]any(ctx))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestNumberFilterErrors(t *testing.T) {
	t.Run("NumberMissingFormat", func(t *testing.T) {
		_, err := numberFilter(1234)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})
}
