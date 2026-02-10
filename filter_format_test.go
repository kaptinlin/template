package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonFilter(t *testing.T) {
	cases := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "SimpleMap",
			input:    map[string]any{"name": "John", "age": 30},
			expected: `{"age":30,"name":"John"}`,
		},
		{
			name:     "SimpleSlice",
			input:    []string{"apple", "banana"},
			expected: `["apple","banana"]`,
		},
		{
			name:     "EmptyMap",
			input:    map[string]any{},
			expected: `{}`,
		},
		{
			name:     "NestedMap",
			input:    map[string]any{"a": map[string]any{"b": 1}},
			expected: `{"a":{"b":1}}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := jsonFilter(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestJsonFilterViaTemplate(t *testing.T) {
	tpl, err := Compile("{{ data | json }}")
	require.NoError(t, err)

	ctx := NewContext()
	ctx.Set("data", map[string]any{"key": "value"})

	got, err := tpl.Render(map[string]any(ctx))
	require.NoError(t, err)
	assert.Equal(t, `{"key":"value"}`, got)
}
