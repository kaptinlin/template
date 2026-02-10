package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonFilter(t *testing.T) {
	// Setup test cases
	cases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			"SimpleMap",
			map[string]interface{}{"name": "John", "age": 30},
			`{"age":30,"name":"John"}`, // Note: JSON serialization sorts keys alphabetically
		},
		{
			"SimpleSlice",
			[]string{"apple", "banana"},
			`["apple","banana"]`,
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
