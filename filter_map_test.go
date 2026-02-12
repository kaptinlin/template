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
			expected: "",
		},
		{
			name:     "IndexOutOfRange",
			template: "{{ data | extract:'2' }}",
			context: map[string]any{
				"data": []any{"First", "Second"},
			},
			expected: "",
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

func TestExtractFilterErrors(t *testing.T) {
	t.Run("MissingKeyPath", func(t *testing.T) {
		_, err := extractFilter(map[string]any{"key": "value"})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("NonMapInput", func(_ *testing.T) {
		// Extracting from a non-map/non-slice returns empty or error.
		_, _ = extractFilter(42, "key")
	})
}
