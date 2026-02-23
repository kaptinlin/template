package template

import (
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
			name:     "StripFilter",
			template: "{{ '  hello  ' | strip }}",
			expected: "hello",
		},
		{
			name:     "LstripFilter",
			template: "{{ '  hello  ' | lstrip }}",
			expected: "hello  ",
		},
		{
			name:     "TrimLeftFilter",
			template: "{{ '  hello  ' | trim_left }}",
			expected: "hello  ",
		},
		{
			name:     "RstripFilter",
			template: "{{ '  hello  ' | rstrip }}",
			expected: "  hello",
		},
		{
			name:     "TrimRightFilter",
			template: "{{ '  hello  ' | trim_right }}",
			expected: "  hello",
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
			name:     "ReplaceFirstFilter",
			template: "{{ 'aabbcc' | replace_first:'b','x' }}",
			expected: "aaxbcc",
		},
		{
			name:     "ReplaceLastFilter",
			template: "{{ 'aabbcc' | replace_last:'b','x' }}",
			expected: "aabxcc",
		},
		{
			name:     "RemoveFilter",
			template: "{{ 'hello world' | remove:' world' }}",
			expected: "hello",
		},
		{
			name:     "RemoveFirstFilter",
			template: "{{ 'abcabc' | remove_first:'b' }}",
			expected: "acabc",
		},
		{
			name:     "RemoveLastFilter",
			template: "{{ 'abcabc' | remove_last:'b' }}",
			expected: "abcac",
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
			name:     "UpcaseFilter",
			template: "{{ 'hello' | upcase }}",
			expected: "HELLO",
		},
		{
			name:     "LowerFilter",
			template: "{{ 'HELLO' | lower }}",
			expected: "hello",
		},
		{
			name:     "DowncaseFilter",
			template: "{{ 'HELLO' | downcase }}",
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
			name:     "TruncateFilterDefault",
			template: "{{ text | truncate }}",
			context:  map[string]any{"text": "This is a short string"},
			expected: "This is a short string",
		},
		{
			name:     "TruncateFilter",
			template: "{{ 'hello world' | truncate:5 }}",
			expected: "he...",
		},
		{
			name:     "TruncateFilterCustomEllipsis",
			template: "{{ 'hello world' | truncate:5,'--' }}",
			expected: "hel--",
		},
		{
			name:     "TruncatewordsFilter",
			template: "{{ 'hello beautiful world' | truncatewords:2 }}",
			expected: "hello beautiful...",
		},
		{
			name:     "TruncateWordsSnakeCase",
			template: "{{ 'hello beautiful world' | truncate_words:2 }}",
			expected: "hello beautiful...",
		},
		{
			name:     "EscapeFilter",
			template: "{{ text | escape }}",
			context:  map[string]any{"text": "<b>bold</b>"},
			expected: "&lt;b&gt;bold&lt;/b&gt;",
		},
		{
			name:     "EscapeFilterAlias",
			template: "{{ text | h }}",
			context:  map[string]any{"text": "<b>bold</b>"},
			expected: "&lt;b&gt;bold&lt;/b&gt;",
		},
		{
			name:     "EscapeOnceFilter",
			template: "{{ text | escape_once }}",
			context:  map[string]any{"text": "&lt;b&gt;bold&lt;/b&gt;"},
			expected: "&lt;b&gt;bold&lt;/b&gt;",
		},
		{
			name:     "StripHTMLFilter",
			template: "{{ text | strip_html }}",
			context:  map[string]any{"text": "<p>Hello <b>World</b></p>"},
			expected: "Hello World",
		},
		{
			name:     "StripNewlinesFilter",
			template: "{{ text | strip_newlines }}",
			context:  map[string]any{"text": "hello\nworld\n"},
			expected: "helloworld",
		},
		{
			name:     "URLEncodeFilter",
			template: "{{ text | url_encode }}",
			context:  map[string]any{"text": "hello world&foo=bar"},
			expected: "hello+world%26foo%3Dbar",
		},
		{
			name:     "URLDecodeFilter",
			template: "{{ text | url_decode }}",
			context:  map[string]any{"text": "hello+world%26foo%3Dbar"},
			expected: "hello world&foo=bar",
		},
		{
			name:     "Base64EncodeFilter",
			template: "{{ text | base64_encode }}",
			context:  map[string]any{"text": "hello world"},
			expected: "aGVsbG8gd29ybGQ=",
		},
		{
			name:     "Base64DecodeFilter",
			template: "{{ text | base64_decode }}",
			context:  map[string]any{"text": "aGVsbG8gd29ybGQ="},
			expected: "hello world",
		},
		{
			name:     "SliceFilterString",
			template: "{{ 'hello' | slice:1,3 }}",
			expected: "ell",
		},
		{
			name:     "SliceFilterStringNoLength",
			template: "{{ 'hello' | slice:1 }}",
			expected: "e",
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
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("ReplaceMissingArgs", func(t *testing.T) {
		_, err := replaceFilter("hello", "old")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("ReplaceFirstMissingArgs", func(t *testing.T) {
		_, err := replaceFirstFilter("hello", "old")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("ReplaceLastMissingArgs", func(t *testing.T) {
		_, err := replaceLastFilter("hello", "old")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("RemoveMissingSubstring", func(t *testing.T) {
		_, err := removeFilter("hello")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("RemoveFirstMissingSubstring", func(t *testing.T) {
		_, err := removeFirstFilter("hello")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("RemoveLastMissingSubstring", func(t *testing.T) {
		_, err := removeLastFilter("hello")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("AppendMissingArg", func(t *testing.T) {
		_, err := appendFilter("hello")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("PrependMissingArg", func(t *testing.T) {
		_, err := prependFilter("hello")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("PluralizeMissingArgs", func(t *testing.T) {
		_, err := pluralizeFilter(1, "apple")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("TruncateDefaultLength", func(t *testing.T) {
		got, err := truncateFilter("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", got) // shorter than default 50
	})

	t.Run("TruncateWordsDefaultCount", func(t *testing.T) {
		got, err := truncateWordsFilter("hello world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", got) // fewer than default 15 words
	})

	t.Run("TruncateInvalidLength", func(t *testing.T) {
		_, err := truncateFilter("hello", "abc")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrFilterInputNotNumeric)
	})

	t.Run("PluralizeInvalidCount", func(t *testing.T) {
		_, err := pluralizeFilter("not_a_number", "apple", "apples")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrFilterInputNotNumeric)
	})

	t.Run("DefaultFilterNoArgs", func(t *testing.T) {
		got, err := defaultFilter("")
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("OrdinalizeValidInt", func(t *testing.T) {
		got, err := ordinalizeFilter(1)
		require.NoError(t, err)
		assert.Equal(t, "1st", got)
	})

	t.Run("OrdinalizeNonNumeric", func(t *testing.T) {
		_, err := ordinalizeFilter("abc")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrFilterInputNotNumeric)
	})

	t.Run("TruncateWordsValid", func(t *testing.T) {
		got, err := truncateWordsFilter("one two three four", 2)
		require.NoError(t, err)
		assert.Equal(t, "one two...", got)
	})

	t.Run("TruncateWordsNoArgs", func(t *testing.T) {
		got, err := truncateWordsFilter("hello world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", got)
	})

	t.Run("TruncateWordsInvalidArg", func(t *testing.T) {
		_, err := truncateWordsFilter("hello world", "abc")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrFilterInputNotNumeric)
	})

	t.Run("TruncateWordsEmpty", func(t *testing.T) {
		got, err := truncateWordsFilter("", 5)
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("SliceMissingOffset", func(t *testing.T) {
		_, err := sliceFilter("hello")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientArgs)
	})

	t.Run("SliceInvalidOffset", func(t *testing.T) {
		_, err := sliceFilter("hello", "abc")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrFilterInputNotNumeric)
	})

	t.Run("SliceInvalidLength", func(t *testing.T) {
		_, err := sliceFilter("hello", 1, "abc")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrFilterInputNotNumeric)
	})
}
