package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArrayFilters(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "UniqueFilter",
			template: "{{ value | unique | join:',' }}",
			context:  map[string]any{"value": []int{1, 2, 2, 3}},
			expected: "1,2,3",
		},
		{
			name:     "UniqFilter",
			template: "{{ value | uniq | join:',' }}",
			context:  map[string]any{"value": []int{1, 2, 2, 3}},
			expected: "1,2,3",
		},
		{
			name:     "JoinFilter",
			template: "{{ value | join:'-' }}",
			context:  map[string]any{"value": []string{"one", "two", "three"}},
			expected: "one-two-three",
		},
		{
			name:     "FirstFilter",
			template: "{{ value | first }}",
			context:  map[string]any{"value": []string{"first", "second"}},
			expected: "first",
		},
		{
			name:     "LastFilter",
			template: "{{ value | last }}",
			context:  map[string]any{"value": []string{"penultimate", "last"}},
			expected: "last",
		},
		{
			name:     "RandomFilter",
			template: "{{ value | random }}",
			context:  map[string]any{"value": []int{1}},
			expected: "1",
		},
		{
			name:     "ReverseFilter",
			template: "{{ value | reverse | join:',' }}",
			context:  map[string]any{"value": []int{1, 2, 3}},
			expected: "3,2,1",
		},
		{
			name:     "ShuffleFilter",
			template: "{{ value | shuffle | join:',' }}",
			context:  map[string]any{"value": []int{1, 1, 1}},
			expected: "1,1,1",
		},
		{
			name:     "SizeFilter",
			template: "{{ value | size }}",
			context:  map[string]any{"value": []int{1, 2, 3}},
			expected: "3",
		},
		{
			name:     "SizeFilterString",
			template: "{{ 'hello' | size }}",
			expected: "5",
		},
		{
			name:     "JoinFilterDefault",
			template: "{{ value | join }}",
			context:  map[string]any{"value": []string{"one", "two", "three"}},
			expected: "one two three",
		},
		{
			name:     "MaxFilter",
			template: "{{ value | max }}",
			context:  map[string]any{"value": []float64{1.1, 2.2, 3.3}},
			expected: "3.3",
		},
		{
			name:     "MinFilter",
			template: "{{ value | min }}",
			context:  map[string]any{"value": []float64{1.1, 2.2, 3.3}},
			expected: "1.1",
		},
		{
			name:     "SumFilter",
			template: "{{ value | sum }}",
			context:  map[string]any{"value": []int{1, 2, 3}},
			expected: "6",
		},
		{
			name:     "SumFilterProperty",
			template: "{{ value | sum:'age' }}",
			context: map[string]any{
				"value": []map[string]any{
					{"name": "Alice", "age": 30},
					{"name": "Bob", "age": 25},
				},
			},
			expected: "55",
		},
		{
			name:     "UniqFilterProperty",
			template: "{{ value | uniq:'role' | map:'name' | join:',' }}",
			context: map[string]any{
				"value": []map[string]any{
					{"name": "Alice", "role": "admin"},
					{"name": "Bob", "role": "user"},
					{"name": "Charlie", "role": "admin"},
				},
			},
			expected: "Alice,Bob",
		},
		{
			name:     "AverageFilter",
			template: "{{ value | average }}",
			context:  map[string]any{"value": []int{1, 2, 3}},
			expected: "2",
		},
		{
			name:     "MapFilter",
			template: "{{ value | map:'name' | join:', ' }}",
			context: map[string]any{
				"value": []map[string]any{
					{"name": "John", "age": 30},
					{"name": "Jane", "age": 25},
				},
			},
			expected: "John, Jane",
		},
		{
			name:     "SortFilter",
			template: "{{ value | sort | join:',' }}",
			context:  map[string]any{"value": []int{3, 1, 2}},
			expected: "1,2,3",
		},
		{
			name:     "SortFilterByKey",
			template: "{{ value | sort:'name' | map:'name' | join:',' }}",
			context: map[string]any{
				"value": []map[string]any{
					{"name": "Charlie"},
					{"name": "Alice"},
					{"name": "Bob"},
				},
			},
			expected: "Alice,Bob,Charlie",
		},
		{
			name:     "SortNaturalFilter",
			template: "{{ value | sort_natural | join:',' }}",
			context:  map[string]any{"value": []string{"Banana", "apple", "Cherry"}},
			expected: "apple,Banana,Cherry",
		},
		{
			name:     "CompactFilter",
			template: "{{ value | compact | join:',' }}",
			context:  map[string]any{"value": []any{1, nil, 2, nil, 3}},
			expected: "1,2,3",
		},
		{
			name:     "WhereFilter",
			template: "{{ value | where:'active','true' | map:'name' | join:',' }}",
			context: map[string]any{
				"value": []map[string]any{
					{"name": "Alice", "active": "true"},
					{"name": "Bob", "active": "false"},
					{"name": "Charlie", "active": "true"},
				},
			},
			expected: "Alice,Charlie",
		},
		{
			name:     "RejectFilter",
			template: "{{ value | reject:'active','false' | map:'name' | join:',' }}",
			context: map[string]any{
				"value": []map[string]any{
					{"name": "Alice", "active": "true"},
					{"name": "Bob", "active": "false"},
					{"name": "Charlie", "active": "true"},
				},
			},
			expected: "Alice,Charlie",
		},
		{
			name:     "FindFilter",
			template: "{{ value | find:'name','Bob' | extract:'age' }}",
			context: map[string]any{
				"value": []map[string]any{
					{"name": "Alice", "age": 30},
					{"name": "Bob", "age": 25},
				},
			},
			expected: "25",
		},
		{
			name:     "FindIndexFilter",
			template: "{{ value | find_index:'name','Bob' }}",
			context: map[string]any{
				"value": []map[string]any{
					{"name": "Alice"},
					{"name": "Bob"},
					{"name": "Charlie"},
				},
			},
			expected: "1",
		},
		{
			name:     "HasFilter",
			template: "{{ value | has:'name','Alice' }}",
			context: map[string]any{
				"value": []map[string]any{
					{"name": "Alice"},
					{"name": "Bob"},
				},
			},
			expected: "true",
		},
		{
			name:     "HasFilterFalse",
			template: "{{ value | has:'name','Unknown' }}",
			context: map[string]any{
				"value": []map[string]any{
					{"name": "Alice"},
					{"name": "Bob"},
				},
			},
			expected: "false",
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

func TestArrayFilterErrors(t *testing.T) {
	t.Run("JoinDefaultSeparator", func(t *testing.T) {
		got, err := joinFilter([]string{"a", "b"})
		require.NoError(t, err)
		assert.Equal(t, "a b", got)
	})

	t.Run("MapMissingKey", func(t *testing.T) {
		_, err := mapFilter([]map[string]any{{"name": "John"}})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("ConcatMissingArg", func(t *testing.T) {
		_, err := concatFilter([]int{1, 2})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("WhereMissingKey", func(t *testing.T) {
		_, err := whereFilter([]map[string]any{{"name": "John"}})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("RejectMissingKey", func(t *testing.T) {
		_, err := rejectFilter([]map[string]any{{"name": "John"}})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("FindMissingArgs", func(t *testing.T) {
		_, err := findFilter([]map[string]any{{"name": "John"}}, "name")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("FindIndexMissingArgs", func(t *testing.T) {
		_, err := findIndexFilter([]map[string]any{{"name": "John"}}, "name")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("HasMissingKey", func(t *testing.T) {
		_, err := hasFilter([]map[string]any{{"name": "John"}})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})
}
