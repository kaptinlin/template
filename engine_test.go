package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockUserProfileContext() Context {
	return Context{
		"userName": "JaneDoe",
		"profile": map[string]any{
			"age": 29,
			"bio": "Software developer with a passion for open source.",
			"contacts": map[string]any{
				"email": "jane.doe@example.com",
			},
		},
		"tasks": []string{"Code Review", "Write Documentation", "Update Dependencies"},
	}
}

func TestExecuteWithErrorHandling(t *testing.T) {
	ctx := mockUserProfileContext()
	// Template that includes a variable and filter that will not cause an error,
	// followed by a non-existent variable that will cause an error.
	tplStr := "Hello, {{userName}}! Missing: {{nonExistentVariable}}"

	parser := NewParser()
	tpl, err := parser.Parse(tplStr)
	require.NoError(t, err, "Failed to parse template")

	// Execute the template.
	output, err := Execute(tpl, ctx)
	assert.Error(t, err, "Expected an error due to non-existent variable")
	assert.Contains(t, output, "Hello, JaneDoe!", "Expected partial output before error")
}

func TestMustExecutePanicsOnError(t *testing.T) {
	ctx := mockUserProfileContext()
	// Template with a non-existent variable that will cause an error.
	tplStr := "Hello, {{userName}}! Missing: {{nonExistentVariable}}"

	parser := NewParser()
	tpl, err := parser.Parse(tplStr)
	require.NoError(t, err, "Failed to parse template")

	// MustExecute should panic when an error occurs.
	assert.Panics(t, func() {
		MustExecute(tpl, ctx)
	}, "Expected MustExecute to panic on error")
}

func TestMustExecuteSuccess(t *testing.T) {
	ctx := mockUserProfileContext()
	tplStr := "Hello, {{userName}}!"

	parser := NewParser()
	tpl, err := parser.Parse(tplStr)
	require.NoError(t, err, "Failed to parse template")

	output := MustExecute(tpl, ctx)
	assert.Equal(t, "Hello, JaneDoe!", output)
}

func TestNestedVariableRetrieval(t *testing.T) {
	ctx := mockUserProfileContext()
	parser := NewParser()
	src := "Contact: {{profile.contacts.email}}"
	expected := "Contact: jane.doe@example.com"

	tmpl, err := parser.Parse(src)
	require.NoError(t, err, "Failed to parse template")

	result, err := tmpl.Execute(ctx)
	require.NoError(t, err, "Failed to execute template")

	assert.Equal(t, expected, result)
}

func TestApplyUpperCaseFilter(t *testing.T) {
	ctx := mockUserProfileContext()
	parser := NewParser()
	src := "Username: {{userName|upper}}"
	expected := "Username: JANEDOE"

	tmpl, err := parser.Parse(src)
	require.NoError(t, err, "Failed to parse template")

	result, err := tmpl.Execute(ctx)
	require.NoError(t, err, "Failed to execute template")

	assert.Equal(t, expected, result)
}

func TestChainFiltersForTitleCase(t *testing.T) {
	ctx := mockUserProfileContext()
	parser := NewParser()
	src := "Bio: {{profile.bio|capitalize}}"
	expected := "Bio: Software developer with a passion for open source."

	tmpl, err := parser.Parse(src)
	require.NoError(t, err, "Failed to parse template")

	result, err := tmpl.Execute(ctx)
	require.NoError(t, err, "Failed to execute template")

	assert.Equal(t, expected, result)
}

func TestVariableNotFoundReturnOriginal(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected string
	}{
		{
			"SingleVariableNotFound",
			"Hello, {{missing}}!",
			"Hello, {{missing}}!",
		},
		{
			"SingleVariableWithFilterNotFound",
			"User: {{missing|lower}}",
			"User: {{missing|lower}}",
		},
		{
			"VariableWithMultipleFiltersNotFound",
			"{{missing|capitalize|append:', welcome back'}}",
			"{{missing|capitalize|append:', welcome back'}}",
		},
		{
			"MixedTextAndNonexistentVariable",
			"Welcome, {{missing}}! How are you?",
			"Welcome, {{missing}}! How are you?",
		},
		{
			"MultipleVariablesSomeNotFound",
			"User: {{user}}, Email: {{email|lower}}, Location: {{location}}",
			"User: {{user}}, Email: {{email|lower}}, Location: {{location}}",
		},
		{
			"ExistingAndNonexistentVariables",
			"User: {{userName}}, Unknown: {{unknown}}",
			"User: JaneDoe, Unknown: {{unknown}}",
		},
		{
			"VariableNotFoundFollowedByExisting",
			"Welcome, {{unknown}}! User: {{userName}}.",
			"Welcome, {{unknown}}! User: JaneDoe.",
		},
		{
			"NestedVariableNotExistsWithFilter",
			"Location: {{profile.location|capitalize}}",
			"Location: {{profile.location|capitalize}}",
		},
		{
			"MixedExistingAndNonexistentVariablesAndText",
			"Hello, {{userName}}! Missing: {{missing}}, Task: {{tasks.0}}, No Task: {{tasks.3}}.",
			"Hello, JaneDoe! Missing: {{missing}}, Task: Code Review, No Task: {{tasks.3}}.",
		},
		{
			"MultipleFiltersOnExistingAndNonexistent",
			"{{userName|lower}}, {{unknown|capitalize|append:' world'}}",
			"janedoe, {{unknown|capitalize|append:' world'}}",
		},
		{
			"ExistingVariableWithNonexistentNested",
			"Profile update: {{profile.|lower}}",
			"Profile update: {{profile.|lower}}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := mockUserProfileContext()
			parser := NewParser()
			tmpl, err := parser.Parse(tc.source)
			require.NoError(t, err, "Unexpected error in %s", tc.name)

			result, err := tmpl.Execute(ctx)
			assert.Error(t, err, "Expected an error")

			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestFilterErrorsReturnOriginalVariableText ensures that when a filter encounters an error,
// the template system gracefully handles it by returning the original variable text.
func TestFilterErrorsReturnOriginalVariableText(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:     "FilterNotFound",
			source:   "Welcome, {{userName|nonExistentFilter}}!",
			expected: "Welcome, {{userName|nonExistentFilter}}!",
		},
		{
			name:     "PlusFilterWithString",
			source:   "Total: {{tasks | plus:' extra'}}", // Plus expects numerical arguments
			expected: "Total: {{tasks | plus:' extra'}}",
		},
		{
			name:     "FirstFilterOnString",
			source:   "First Task: {{userName|first}}", // First filter applied on a string, not an array
			expected: "First Task: {{userName|first}}",
		},
		{
			name:     "IndexFilterOutOfRange",
			source:   "Task: {{tasks|index:10}}",
			expected: "Task: {{tasks|index:10}}",
		},
		{
			name:     "MultipleFiltersWithOneInvalid",
			source:   "Email: {{profile.contacts.email|lower|nonExistentFilter}}",
			expected: "Email: {{profile.contacts.email|lower|nonExistentFilter}}",
		},
		{
			name:     "ValidAndInvalidFiltersMixed",
			source:   "Bio: {{profile.bio|capitalize|nonExistentFilter}}, Age: {{profile.age|plus:1}}",
			expected: "Bio: {{profile.bio|capitalize|nonExistentFilter}}, Age: 30",
		},
		{
			name:     "LastFilterOnNonArray",
			source:   "Contact: {{profile.contacts|last}}", // Last filter expects an array
			expected: "Contact: {{profile.contacts|last}}",
		},
		{
			name:     "InvalidNestedVariableWithFilter",
			source:   "Missing: {{profile.missingDetail|lower}}",
			expected: "Missing: {{profile.missingDetail|lower}}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := mockUserProfileContext()
			parser := NewParser()
			tmpl, err := parser.Parse(tc.source)
			require.NoError(t, err, "Unexpected error in %s", tc.name)

			result, err := tmpl.Execute(ctx)
			assert.Error(t, err, "Expected an error in %s", tc.name)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVariablesWithPunctuation(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		context  Context
		expected string
	}{
		{
			name:     "VariableFollowedByExclamation",
			source:   "Hello, {{userName}}!",
			expected: "Hello, JaneDoe!",
		},
		{
			name:     "VariableFollowedByComma",
			source:   "User: {{userName}}, welcome back!",
			expected: "User: JaneDoe, welcome back!",
		},
		{
			name:     "VariableFollowedByPeriod",
			source:   "Your name is {{userName}}.",
			expected: "Your name is JaneDoe.",
		},
		{
			name:     "VariableFollowedByQuestionMark",
			source:   "Is {{userName}} your name?",
			expected: "Is JaneDoe your name?",
		},
		{
			name:     "VariableInsideQuotes",
			source:   "\"{{userName}}\" is the username.",
			expected: "\"JaneDoe\" is the username.",
		},
		{
			name:     "MultipleVariablesSeparatedByPunctuation",
			source:   "{{userName}}, your age is {{profile.age}}.",
			expected: "JaneDoe, your age is 29.",
		},
		{
			name:     "VariableFollowedByFilterAndPunctuation",
			source:   "Welcome, {{userName|lower}}!",
			expected: "Welcome, janedoe!",
		},
	}
	ctx := mockUserProfileContext()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser()
			tmpl, err := parser.Parse(tc.source)
			require.NoError(t, err, "Unexpected error in %s", tc.name)

			result := tmpl.MustExecute(ctx)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVariablesWithPunctuationAndErrors(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:     "FilterNotFoundWithPunctuation",
			source:   "Hello, {{userName|nonExistentFilter}}!",
			expected: "Hello, {{userName|nonExistentFilter}}!",
		},
		{
			name:     "FilterArgumentMismatchWithPunctuation",
			source:   "Result: {{profile.age|plus}}.",
			expected: "Result: {{profile.age|plus}}.",
		},
	}
	ctx := mockUserProfileContext()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser()
			tmpl, err := parser.Parse(tc.source)
			require.NoError(t, err, "Unexpected error in %s", tc.name)

			result, err := tmpl.Execute(ctx)
			assert.Error(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFiltersWithVariableArguments(t *testing.T) {
	ctx := mockUserProfileContext()
	cases := []struct {
		name     string
		source   string
		expected string
	}{
		{
			"PlusFilterWithVariableArgument",
			"New age: {{ profile.age|plus:profile.age }}",
			"New age: 58",
		},
		{
			"MinusFilterWithVariableArgument",
			"Half age: {{ profile.age|minus:profile.age|plus:2 }}",
			"Half age: 2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser()
			tmpl, err := parser.Parse(tc.source)
			require.NoError(t, err, "Unexpected error in %s", tc.name)

			result, err := tmpl.Execute(ctx)
			require.NoError(t, err, "Failed to execute template in %s", tc.name)

			assert.Equal(t, tc.expected, result, "Test case %s", tc.name)
		})
	}
}
func TestFiltersWithNumericArguments(t *testing.T) {
	ctx := mockUserProfileContext()
	cases := []struct {
		name     string
		source   string
		expected string
	}{
		{
			"PlusFilterWithNumericArgument",
			"Age next year: {{ profile.age|plus:1 }}",
			"Age next year: 30",
		},
		{
			"TimesFilterWithNumericArgument",
			"Double age: {{ profile.age|times:2 }}",
			"Double age: 58",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser()
			tmpl, err := parser.Parse(tc.source)
			require.NoError(t, err, "Unexpected error in %s", tc.name)

			result, err := tmpl.Execute(ctx)
			require.NoError(t, err, "Failed to execute template in %s", tc.name)

			assert.Equal(t, tc.expected, result, "Test case %s", tc.name)
		})
	}
}
