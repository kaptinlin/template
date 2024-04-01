package template

import (
	"strings"
	"testing"
)

func mockUserProfileContext() Context {
	return Context{
		"userName": "JaneDoe",
		"profile": map[string]interface{}{
			"age": 29,
			"bio": "Software developer with a passion for open source.",
			"contacts": map[string]interface{}{
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
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Execute the template.
	output, err := Execute(tpl, ctx)
	if err == nil {
		t.Errorf("Expected an error due to non-existent variable, but got nil")
	}
	if !strings.Contains(output, "Hello, JaneDoe!") {
		t.Errorf("Expected partial output before error, got: %s", output)
	}
}

func TestMustExecuteIgnoresError(t *testing.T) {
	ctx := mockUserProfileContext()
	// Similar template to above, which will encounter an error due to a non-existent variable.
	tplStr := "Hello, {{userName}}! Missing: {{nonExistentVariable}}"

	parser := NewParser()
	tpl, err := parser.Parse(tplStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// MustExecute should ignore errors and attempt to return any partial output.
	output := MustExecute(tpl, ctx)
	if !strings.Contains(output, "Hello, JaneDoe!") {
		t.Errorf("Expected partial output before error in MustExecute, got: %s", output)
	}
	// MustExecute should ignore the error and output the placeholder for missing variable.
	if !strings.Contains(output, "{{nonExistentVariable}}") {
		t.Errorf("Expected placeholder for missing variable, got: %s", output)
	}
}

func TestNestedVariableRetrieval(t *testing.T) {
	ctx := mockUserProfileContext()
	parser := NewParser()
	src := "Contact: {{profile.contacts.email}}"
	expected := "Contact: jane.doe@example.com"

	tmpl, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}

func TestApplyUpperCaseFilter(t *testing.T) {
	ctx := mockUserProfileContext()
	parser := NewParser()
	src := "Username: {{userName|upper}}"
	expected := "Username: JANEDOE"

	tmpl, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}

func TestChainFiltersForTitleCase(t *testing.T) {
	ctx := mockUserProfileContext()
	parser := NewParser()
	src := "Bio: {{profile.bio|capitalize}}"
	expected := "Bio: Software developer with a passion for open source."

	tmpl, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
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
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			result, err := tmpl.Execute(ctx)
			if err == nil {
				t.Errorf("Expected an error, but got nil")
			}

			if result != tc.expected {
				t.Errorf("Expected '%s', but got '%s'", tc.expected, result)
			}
		})
	}
}
