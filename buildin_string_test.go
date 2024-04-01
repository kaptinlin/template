package template

import (
	"testing"
)

func TestStringFilters(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "DefaultFilter",
			template: "{{ name | default:'Unknown' }}",
			context:  map[string]interface{}{"name": ""},
			expected: "Unknown",
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
			context:  map[string]interface{}{"count": 1},
			expected: "apple",
		},
		{
			name:     "PluralizeFilterPlural",
			template: "{{ count | pluralize:'apple','apples' }}",
			context:  map[string]interface{}{"count": 2},
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
			// Parse the template
			tpl, err := Parse(tc.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			// Create a context and add variables
			context := NewContext()
			for k, v := range tc.context {
				context.Set(k, v)
			}

			// Execute the template
			output, err := Execute(tpl, context)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			// Verify the output
			if output != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, output)
			}
		})
	}
}
