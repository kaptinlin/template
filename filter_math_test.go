package template

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMathFilters(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "AbsFilterPositive",
			template: "{{ value | abs }}",
			context:  map[string]any{"value": -42},
			expected: "42",
		},
		{
			name:     "AtLeastFilter",
			template: "{{ value | atLeast:10 }}",
			context:  map[string]any{"value": 5},
			expected: "10",
		},
		{
			name:     "AtLeastFilterNoClamp",
			template: "{{ value | atLeast:3 }}",
			context:  map[string]any{"value": 5},
			expected: "5",
		},
		{
			name:     "AtMostFilter",
			template: "{{ value | atMost:10 }}",
			context:  map[string]any{"value": 15},
			expected: "10",
		},
		{
			name:     "AtMostFilterNoClamp",
			template: "{{ value | atMost:20 }}",
			context:  map[string]any{"value": 15},
			expected: "15",
		},
		{
			name:     "RoundFilter",
			template: "{{ value | round:2 }}",
			context:  map[string]any{"value": 3.14159},
			expected: "3.14",
		},
		{
			name:     "FloorFilter",
			template: "{{ value | floor }}",
			context:  map[string]any{"value": 3.99},
			expected: "3",
		},
		{
			name:     "CeilFilter",
			template: "{{ value | ceil }}",
			context:  map[string]any{"value": 3.01},
			expected: "4",
		},
		{
			name:     "PlusFilter",
			template: "{{ value | plus:3 }}",
			context:  map[string]any{"value": 7},
			expected: "10",
		},
		{
			name:     "MinusFilter",
			template: "{{ value | minus:2 }}",
			context:  map[string]any{"value": 10},
			expected: "8",
		},
		{
			name:     "TimesFilter",
			template: "{{ value | times:2 }}",
			context:  map[string]any{"value": 5},
			expected: "10",
		},
		{
			name:     "DivideFilter",
			template: "{{ value | divide:4 }}",
			context:  map[string]any{"value": 20},
			expected: "5",
		},
		{
			name:     "ModuloFilter",
			template: "{{ value | modulo:3 }}",
			context:  map[string]any{"value": 10},
			expected: "1",
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

func TestMathFilterErrors(t *testing.T) {
	t.Run("AtLeastMissingArg", func(t *testing.T) {
		_, err := atLeastFilter(5)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("AtMostMissingArg", func(t *testing.T) {
		_, err := atMostFilter(5)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("RoundMissingArg", func(t *testing.T) {
		_, err := roundFilter(3.14)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("PlusMissingArg", func(t *testing.T) {
		_, err := plusFilter(5)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("MinusMissingArg", func(t *testing.T) {
		_, err := minusFilter(5)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("TimesMissingArg", func(t *testing.T) {
		_, err := timesFilter(5)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("DivideMissingArg", func(t *testing.T) {
		_, err := divideFilter(5)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("ModuloMissingArg", func(t *testing.T) {
		_, err := moduloFilter(5)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})
}
