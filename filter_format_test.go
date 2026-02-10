package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonFilter(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "SimpleMap",
			input:    map[string]interface{}{"name": "John", "age": 30},
			expected: `{"age":30,"name":"John"}`,
		},
		{
			name:     "SimpleSlice",
			input:    []string{"apple", "banana"},
			expected: `["apple","banana"]`,
		},
		{
			name:     "EmptyMap",
			input:    map[string]interface{}{},
			expected: `{}`,
		},
		{
			name:     "NestedMap",
			input:    map[string]interface{}{"a": map[string]interface{}{"b": 1}},
			expected: `{"a":{"b":1}}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := jsonFilter(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestJsonFilterViaTemplate(t *testing.T) {
	tpl, err := Compile("{{ data | json }}")
	require.NoError(t, err)

	ctx := NewContext()
	ctx.Set("data", map[string]interface{}{"key": "value"})

	output, err := tpl.Render(map[string]interface{}(ctx))
	require.NoError(t, err)
	assert.Equal(t, `{"key":"value"}`, output)
}
