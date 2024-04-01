package template

import (
	"testing"
	"time"
)

// TestTemplateExecution verifies the correct execution of a template with various node types.
func TestTemplateExecution(t *testing.T) {
	ctx := mockUserProfileContext()
	// Define test cases
	cases := []struct {
		name     string
		nodes    []*Node
		expected string
	}{
		{
			"TextOnly",
			[]*Node{{Type: "text", Text: "Hello, world!"}},
			"Hello, world!",
		},
		{
			"SingleVariable",
			[]*Node{
				{Type: "text", Text: "User: "},
				{Type: "variable", Variable: "userName"},
			},
			"User: JaneDoe",
		},
		{
			"NestedVariable",
			[]*Node{
				{Type: "variable", Variable: "profile.age"},
			},
			"29",
		},
		{
			"VariableNotFound",
			[]*Node{
				{Type: "variable", Variable: "nonexistent", Text: "{{nonexistent}}"},
			},
			"{{nonexistent}}",
		},
		{
			"MixedContent",
			[]*Node{
				{Type: "text", Text: "Welcome, "},
				{Type: "variable", Variable: "userName"},
				{Type: "text", Text: "! Age: "},
				{Type: "variable", Variable: "profile.age"},
			},
			"Welcome, JaneDoe! Age: 29",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup template
			tmpl := &Template{Nodes: tc.nodes}
			// Execute template
			result := tmpl.MustExecute(ctx)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestConvertToString(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "String",
			input:    "Hello, world!",
			expected: "Hello, world!",
		},
		{
			name:     "SliceOfString",
			input:    []string{"apple", "banana", "cherry"},
			expected: "[apple, banana, cherry]",
		},
		{
			name:     "SliceOfInt",
			input:    []int{1, 2, 3},
			expected: "[1, 2, 3]",
		},
		{
			name:     "SliceOfFloat64",
			input:    []float64{1.1, 2.2, 3.3},
			expected: "[1.1, 2.2, 3.3]",
		},
		{
			name:     "SliceOfBool",
			input:    []bool{true, false, true},
			expected: "[true, false, true]",
		},
		{
			name:     "Time",
			input:    time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "2020-01-01 12:00:00",
		},
		{
			name:     "ComplexTypeWithJSONFallback",
			input:    map[string]interface{}{"name": "John Doe", "age": 30},
			expected: "{\n  \"age\": 30,\n  \"name\": \"John Doe\"\n}",
		},
		{
			name:     "HandleErrorInJSONFallback",
			input:    make(chan int),
			expected: "could not convert value to string: json: unsupported type: chan int",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := convertToString(tc.input)
			if err != nil {
				if err.Error() != tc.expected {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
