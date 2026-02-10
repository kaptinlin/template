package template

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArrayFilters(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "UniqueFilter",
			template: "{{ value | unique | join:',' }}",
			context:  map[string]interface{}{"value": []int{1, 2, 2, 3}},
			expected: "1,2,3",
		},
		{
			name:     "JoinFilter",
			template: "{{ value | join:'-' }}",
			context:  map[string]interface{}{"value": []string{"one", "two", "three"}},
			expected: "one-two-three",
		},
		{
			name:     "FirstFilter",
			template: "{{ value | first }}",
			context:  map[string]interface{}{"value": []string{"first", "second"}},
			expected: "first",
		},
		{
			name:     "LastFilter",
			template: "{{ value | last }}",
			context:  map[string]interface{}{"value": []string{"penultimate", "last"}},
			expected: "last",
		},
		{
			name:     "RandomFilter",
			template: "{{ value | random }}",
			context:  map[string]interface{}{"value": []int{1}},
			expected: "1",
		},
		{
			name:     "ReverseFilter",
			template: "{{ value | reverse | join:',' }}",
			context:  map[string]interface{}{"value": []int{1, 2, 3}},
			expected: "3,2,1",
		},
		{
			name:     "ShuffleFilter",
			template: "{{ value | shuffle | join:',' }}",
			context:  map[string]interface{}{"value": []int{1, 1, 1}},
			expected: "1,1,1",
		},
		{
			name:     "SizeFilter",
			template: "{{ value | size }}",
			context:  map[string]interface{}{"value": []int{1, 2, 3}},
			expected: "3",
		},
		{
			name:     "MaxFilter",
			template: "{{ value | max }}",
			context:  map[string]interface{}{"value": []float64{1.1, 2.2, 3.3}},
			expected: "3.3",
		},
		{
			name:     "MinFilter",
			template: "{{ value | min }}",
			context:  map[string]interface{}{"value": []float64{1.1, 2.2, 3.3}},
			expected: "1.1",
		},
		{
			name:     "SumFilter",
			template: "{{ value | sum }}",
			context:  map[string]interface{}{"value": []int{1, 2, 3}},
			expected: "6",
		},
		{
			name:     "AverageFilter",
			template: "{{ value | average }}",
			context:  map[string]interface{}{"value": []int{1, 2, 3}},
			expected: "2",
		},
		{
			name:     "MapFilter",
			template: "{{ value | map:'name' | join:', ' }}",
			context: map[string]interface{}{
				"value": []map[string]interface{}{
					{"name": "John", "age": 30},
					{"name": "Jane", "age": 25},
				},
			},
			expected: "John, Jane",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := Compile(tc.template)
			require.NoError(t, err)

			context := NewContext()
			for k, v := range tc.context {
				context.Set(k, v)
			}

			output, err := tpl.Render(map[string]interface{}(context))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestArrayFilterErrors(t *testing.T) {
	t.Run("JoinMissingSeparator", func(t *testing.T) {
		_, err := joinFilter([]string{"a", "b"})
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("MapMissingKey", func(t *testing.T) {
		_, err := mapFilter([]map[string]interface{}{{"name": "John"}})
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})
}
