package template

import (
	"testing"
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
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if output != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, output)
			}
		})
	}
}
