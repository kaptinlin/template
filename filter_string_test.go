package template

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringFilters(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "DefaultFilter",
			template: "{{ name | default:'Unknown' }}",
			context:  map[string]any{"name": ""},
			expected: "Unknown",
		},
		{
			name:     "DefaultFilterWithValue",
			template: "{{ name | default:'Unknown' }}",
			context:  map[string]any{"name": "Alice"},
			expected: "Alice",
		},
		{
			name:     "TrimFilter",
			template: "{{ '  hello  ' | trim }}",
			expected: "hello",
		},
		{
			name:     "SplitFilter",
			template: "{{ 'one,two,three' | split:',' | size }}",
			expected: "3",
		},
		{
			name:     "ReplaceFilter",
			template: "{{ 'hello world' | replace:'world','there' }}",
			expected: "hello there",
		},
		{
			name:     "RemoveFilter",
			template: "{{ 'hello world' | remove:' world' }}",
			expected: "hello",
		},
		{
			name:     "AppendFilter",
			template: "{{ 'hello' | append:' world' }}",
			expected: "hello world",
		},
		{
			name:     "PrependFilter",
			template: "{{ 'world' | prepend:'hello ' }}",
			expected: "hello world",
		},
		{
			name:     "LengthFilter",
			template: "{{ 'hello' | length }}",
			expected: "5",
		},
		{
			name:     "UpperFilter",
			template: "{{ 'hello' | upper }}",
			expected: "HELLO",
		},
		{
			name:     "LowerFilter",
			template: "{{ 'HELLO' | lower }}",
			expected: "hello",
		},
		{
			name:     "TitleizeFilter",
			template: "{{ 'hello world' | titleize }}",
			expected: "Hello World",
		},
		{
			name:     "CapitalizeFilter",
			template: "{{ 'hello' | capitalize }}",
			expected: "Hello",
		},
		{
			name:     "CamelizeFilter",
			template: "{{ 'hello_world' | camelize }}",
			expected: "helloWorld",
		},
		{
			name:     "PascalizeFilter",
			template: "{{ 'hello_world' | pascalize }}",
			expected: "HelloWorld",
		},
		{
			name:     "DasherizeFilter",
			template: "{{ 'hello world' | dasherize }}",
			expected: "hello-world",
		},
		{
			name:     "SlugifyFilter",
			template: "{{ 'Hello Wörld & Friends' | slugify }}",
			expected: "hello-world-and-friends",
		},
		{
			name:     "PluralizeFilterSingular",
			template: "{{ count | pluralize:'apple','apples' }}",
			context:  map[string]any{"count": 1},
			expected: "apple",
		},
		{
			name:     "PluralizeFilterPlural",
			template: "{{ count | pluralize:'apple','apples' }}",
			context:  map[string]any{"count": 2},
			expected: "apples",
		},
		{
			name:     "OrdinalizeFilter",
			template: "{{ '1' | ordinalize }}",
			expected: "1st",
		},
		{
			name:     "TruncateFilter",
			template: "{{ 'hello world' | truncate:5 }}",
			expected: "hello...",
		},
		{
			name:     "TruncateWordsFilter",
			template: "{{ 'hello beautiful world' | truncateWords:2 }}",
			expected: "hello beautiful...",
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

func TestStringFilterErrors(t *testing.T) {
	t.Run("SplitMissingDelimiter", func(t *testing.T) {
		_, err := splitFilter("hello")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("ReplaceMissingArgs", func(t *testing.T) {
		_, err := replaceFilter("hello", "old")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("RemoveMissingSubstring", func(t *testing.T) {
		_, err := removeFilter("hello")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("AppendMissingArg", func(t *testing.T) {
		_, err := appendFilter("hello")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("PrependMissingArg", func(t *testing.T) {
		_, err := prependFilter("hello")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("PluralizeMissingArgs", func(t *testing.T) {
		_, err := pluralizeFilter(1, "apple")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("TruncateMissingLength", func(t *testing.T) {
		_, err := truncateFilter("hello")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("TruncateWordsMissingCount", func(t *testing.T) {
		_, err := truncateWordsFilter("hello world")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInsufficientArgs))
	})

	t.Run("TruncateInvalidLength", func(t *testing.T) {
		_, err := truncateFilter("hello", "abc")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrFilterInputNotNumeric))
	})

	t.Run("PluralizeInvalidCount", func(t *testing.T) {
		_, err := pluralizeFilter("not_a_number", "apple", "apples")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrFilterInputNotNumeric))
	})

	t.Run("DefaultFilterNoArgs", func(t *testing.T) {
		got, err := defaultFilter("")
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})
}
