package template

import (
	"reflect"
	"testing"
)

func TestParseTextNode(t *testing.T) {
	source := "Hello, world!"
	parser := NewParser()
	tpl, err := parser.Parse(source)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := &Template{
		Nodes: []*Node{{Type: "text", Text: "Hello, world!"}},
	}

	if !reflect.DeepEqual(tpl, expected) {
		t.Errorf("Expected %v, got %v", expected, tpl)
	}
}

func TestParseTextNodeWithWhitespace(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected *Template
	}{
		{
			"OnlySpaces",
			"    ",
			&Template{
				Nodes: []*Node{{Type: "text", Text: "    "}},
			},
		},
		{
			"OnlyLineBreaks",
			"\n\n\n",
			&Template{
				Nodes: []*Node{{Type: "text", Text: "\n\n\n"}},
			},
		},
		{
			"OnlyTabs",
			"\t\t\t",
			&Template{
				Nodes: []*Node{{Type: "text", Text: "\t\t\t"}},
			},
		},
		{
			"SpacesAndText",
			"  Hello, world!  ",
			&Template{
				Nodes: []*Node{{Type: "text", Text: "  Hello, world!  "}},
			},
		},
		{
			"Newlines",
			"\nHello,\nworld!\n",
			&Template{
				Nodes: []*Node{{Type: "text", Text: "\nHello,\nworld!\n"}},
			},
		},
		{
			"TabsAndSpaces",
			"\tHello,  world!\t",
			&Template{
				Nodes: []*Node{{Type: "text", Text: "\tHello,  world!\t"}},
			},
		},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			if !reflect.DeepEqual(tpl, tc.expected) {
				t.Errorf("Case %s: Expected %v, got %v", tc.name, tc.expected, tpl)
			}
		})
	}
}

func TestParseTextNodeWithMultipleLinesAndVariations(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected *Template
	}{
		{
			"MultipleLinesSimple",
			`Hello,
World!
This is a test.`,
			&Template{
				Nodes: []*Node{{Type: "text", Text: "Hello,\nWorld!\nThis is a test."}},
			},
		},
		{
			"MultipleLinesWithEmptyLines",
			`Hello,

World!


This is a test.`,
			&Template{
				Nodes: []*Node{{Type: "text", Text: "Hello,\n\nWorld!\n\n\nThis is a test."}},
			},
		},
		{
			"MultipleLinesWithVariable",
			`User: {{username}}
Welcome back!`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "User: "},
					{Type: "variable", Variable: "username", Text: "{{username}}"},
					{Type: "text", Text: "\nWelcome back!"},
				},
			},
		},
		{
			"MultipleLinesWithVariableAndText",
			`User: {{
username
}}
Welcome back!`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "User: "},
					{Type: "variable", Variable: "username", Text: "{{\nusername\n}}"},
					{Type: "text", Text: "\nWelcome back!"},
				},
			},
		},
		{
			"MultipleLinesWithVariableAndTextAndSpaces",
			`User: {{
	username
	}}
Welcome back!`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "User: "},
					{Type: "variable", Variable: "username", Text: `{{
	username
	}}`},
					{Type: "text", Text: "\nWelcome back!"},
				},
			},
		},
		{
			"MultipleLinesWithVariableAndTextAndTabs",
			"User: {{\t username \n}}\nWelcome back!",
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "User: "},
					{Type: "variable", Variable: "username", Text: "{{\t username \n}}"},
					{Type: "text", Text: "\nWelcome back!"},
				},
			},
		},
		{
			"MultipleLinesWithVariableAndFilters",
			`User: {{username|lower}}
Welcome back!`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "User: "},
					{Type: "variable", Variable: "username", Filters: []Filter{{Name: "lower"}}, Text: "{{username|lower}}"},
					{Type: "text", Text: "\nWelcome back!"},
				},
			},
		},
		{
			"MultipleLinesWithVariableAndFiltersAndText",
			`User: {{username|lower}}
Welcome back, {{username}}!`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "User: "},
					{Type: "variable", Variable: "username", Filters: []Filter{{Name: "lower"}}, Text: "{{username|lower}}"},
					{Type: "text", Text: "\nWelcome back, "},
					{Type: "variable", Variable: "username", Text: "{{username}}"},
					{Type: "text", Text: "!"},
				},
			},
		},
		{
			"MultipleLinesWithVariableAndFiltersAndTextAndSpaces",
			`User: {{ username | lower }}
Welcome back, {{ username }}!`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "User: "},
					{Type: "variable", Variable: "username", Filters: []Filter{{Name: "lower"}}, Text: "{{ username | lower }}"},
					{Type: "text", Text: "\nWelcome back, "},
					{Type: "variable", Variable: "username", Text: "{{ username }}"},
					{Type: "text", Text: "!"},
				},
			},
		},
		{
			"MultipleLinesWithVariableAndFiltersAndArgs",
			`User: {{username|replace:"Mr.","Mrs."}}
Welcome back!`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "User: "},
					{Type: "variable", Variable: "username", Filters: []Filter{{Name: "replace", Args: []string{"Mr.", "Mrs."}}}, Text: `{{username|replace:"Mr.","Mrs."}}`},
					{Type: "text", Text: "\nWelcome back!"},
				},
			},
		},
		{
			"MixedSpacesAndTabs",
			"\tHello,\n  World!  \n\tThis is a test.",
			&Template{
				Nodes: []*Node{{Type: "text", Text: "\tHello,\n  World!  \n\tThis is a test."}},
			},
		},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			if !reflect.DeepEqual(tpl, tc.expected) {
				t.Errorf("Case %s: Expected %v, got %v", tc.name, tc.expected, tpl)
			}
		})
	}
}

func TestParseVariableNode(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{"NoSpaces", "{{username}}"},
		{"SpacesBeforeVariable", "{{  username}}"},
		{"SpacesAfterVariable", "{{username  }}"},
		{"SpacesBeforeAndAfterVariable", "{{  username  }}"},
		{"LineBreakBeforeVariable", "{{\nusername}}"},
		{"LineBreakAfterVariable", "{{username\n}}"},
		{"LineBreakBeforeAndAfterVariable", "{{\nusername\n}}"},
		{"LineBreaksAndSpacesBeforeVariable", "{{  \nusername}}"},
		{"LineBreaksAndSpacesAfterVariable", "{{username\n  }}"},
		{"LineBreaksAndSpacesBeforeAndAfterVariable", "{{  \nusername\n  }}"},
		{"TabsBeforeVariable", "{{\tusername}}"},
		{"TabsAfterVariable", "{{username\t}}"},
		{"TabsBeforeAndAfterVariable", "{{\tusername\t}}"},
		{"TabsAndSpacesBeforeVariable", "{{\t  username}}"},
		{"TabsAndSpacesAfterVariable", "{{username  \t}}"},
		{"TabsAndSpacesBeforeAndAfterVariable", "{{\t  username  \t}}"},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			expected := &Template{
				Nodes: []*Node{{
					Type:     "variable",
					Variable: "username",
					Text:     tc.source,
				}},
			}

			if !reflect.DeepEqual(tpl, expected) {
				t.Errorf("Expected %v, got %v", expected, tpl)
			}
		})
	}
}
func TestParseNestedContextVariable(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{"DirectNestedVariable", "{{user.details.name}}"},
		{"SpacesInsideBraces", "{{ user.details.name }}"},
		{"SpacesBeforeVariable", "{{  user.details.name}}"},
		{"SpacesAfterVariable", "{{user.details.name  }}"},
		{"SpacesBeforeAndAfterVariable", "{{  user.details.name  }}"},
		{"TabsBeforeNestedVariable", "{{\tuser.details.name}}"},
		{"TabsAfterNestedVariable", "{{user.details.name\t}}"},
		{"TabsAroundNestedVariable", "{{\tuser.details.name\t}}"},
		{"LineBreakBeforeNestedVariable", "{{\nuser.details.name}}"},
		{"LineBreakAfterNestedVariable", "{{user.details.name\n}}"},
		{"LineBreaksAroundNestedVariable", "{{\nuser.details.name\n}}"},
		{"MixedWhitespaceAroundNestedVariable", "{{ \t\nuser.details.name\t\n }}"},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			expected := &Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "user.details.name",
						Text:     tc.source,
					},
				},
			}

			if !reflect.DeepEqual(tpl, expected) {
				t.Errorf("Case %s: Expected %+v, got %+v", tc.name, expected, tpl)
			}
		})
	}
}

func TestParseMixedTextAndVariableNodes(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected *Template
	}{
		{
			"BasicMixedContent",
			"Hello, {{username}}! Welcome to the site.",
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "Hello, "},
					{Type: "variable", Variable: "username", Text: "{{username}}"},
					{Type: "text", Text: "! Welcome to the site."},
				},
			},
		},
		{
			"SpacesInsideVariableBraces",
			"Hello, {{ username }}! Welcome to our world.",
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "Hello, "},
					{Type: "variable", Variable: "username", Text: "{{ username }}"},
					{Type: "text", Text: "! Welcome to our world."},
				},
			},
		},
		{
			"MultipleVariables",
			"User: {{ firstName }} {{ lastName }} - Welcome back!",
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "User: "},
					{Type: "variable", Variable: "firstName", Text: "{{ firstName }}"},
					{Type: "text", Text: " "},
					{Type: "variable", Variable: "lastName", Text: "{{ lastName }}"},
					{Type: "text", Text: " - Welcome back!"},
				},
			},
		},
		{
			"VariableStartOfLine",
			"{{ greeting }} John, have a great day!",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "greeting", Text: "{{ greeting }}"},
					{Type: "text", Text: " John, have a great day!"},
				},
			},
		},
		{
			"VariableEndOfLine",
			"Goodbye, {{ username }}",
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "Goodbye, "},
					{Type: "variable", Variable: "username", Text: "{{ username }}"},
				},
			},
		},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			if !reflect.DeepEqual(tpl, tc.expected) {
				t.Errorf("Case %s: Expected %v, got %v", tc.name, tc.expected, tpl)
			}
		})
	}
}

func TestParseVariableNodeWithFilterNoParams(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected *Template
	}{
		{
			name:   "SingleFilterNoSpace",
			source: "{{username|upper}}",
			expected: &Template{
				Nodes: []*Node{{Type: "variable", Variable: "username", Filters: []Filter{{Name: "upper"}}, Text: "{{username|upper}}"}},
			},
		},
		{
			name:   "SpaceBeforePipe",
			source: "{{username |upper}}",
			expected: &Template{
				Nodes: []*Node{{Type: "variable", Variable: "username", Filters: []Filter{{Name: "upper"}}, Text: "{{username |upper}}"}},
			},
		},
		{
			name:   "SpaceAfterPipe",
			source: "{{username| upper}}",
			expected: &Template{
				Nodes: []*Node{{Type: "variable", Variable: "username", Filters: []Filter{{Name: "upper"}}, Text: "{{username| upper}}"}},
			},
		},
		{
			name:   "SpacesAroundPipe",
			source: "{{username | upper}}",
			expected: &Template{
				Nodes: []*Node{{Type: "variable", Variable: "username", Filters: []Filter{{Name: "upper"}}, Text: "{{username | upper}}"}},
			},
		},
		{
			name:   "MultipleFiltersNoSpaces",
			source: "{{username|lower|capitalize}}",
			expected: &Template{
				Nodes: []*Node{{Type: "variable", Variable: "username", Filters: []Filter{{Name: "lower"}, {Name: "capitalize"}}, Text: "{{username|lower|capitalize}}"}},
			},
		},
		{
			name:   "SpacesAroundMultipleFilters",
			source: "{{username | lower | capitalize}}",
			expected: &Template{
				Nodes: []*Node{{Type: "variable", Variable: "username", Filters: []Filter{{Name: "lower"}, {Name: "capitalize"}}, Text: "{{username | lower | capitalize}}"}},
			},
		},
		{
			name:   "TextNodesAroundVariableWithFilter",
			source: "Hello {{name|capitalize}}, welcome!",
			expected: &Template{
				Nodes: []*Node{{Type: "text", Text: "Hello "}, {Type: "variable", Variable: "name", Filters: []Filter{{Name: "capitalize"}}, Text: "{{name|capitalize}}"}, {Type: "text", Text: ", welcome!"}},
			},
		},
		{
			name:   "TextNodeBeforeVariableMultipleFilters",
			source: "User: {{username|trim|lower}}",
			expected: &Template{
				Nodes: []*Node{{Type: "text", Text: "User: "}, {Type: "variable", Variable: "username", Filters: []Filter{{Name: "trim"}, {Name: "lower"}}, Text: "{{username|trim|lower}}"}},
			},
		},
		{
			name:   "TextNodeAfterVariableMultipleFilters",
			source: "{{username|trim|capitalize}} logged in",
			expected: &Template{
				Nodes: []*Node{{Type: "variable", Variable: "username", Filters: []Filter{{Name: "trim"}, {Name: "capitalize"}}, Text: "{{username|trim|capitalize}}"}, {Type: "text", Text: " logged in"}},
			},
		},
		{
			name:   "ComplexMixedTextAndVariables",
			source: "Dear {{name|capitalize}}, your score is {{score|round}}.",
			expected: &Template{
				Nodes: []*Node{{Type: "text", Text: "Dear "}, {Type: "variable", Variable: "name", Filters: []Filter{{Name: "capitalize"}}, Text: "{{name|capitalize}}"}, {Type: "text", Text: ", your score is "}, {Type: "variable", Variable: "score", Filters: []Filter{{Name: "round"}}, Text: "{{score|round}}"}, {Type: "text", Text: "."}},
			},
		},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			if !reflect.DeepEqual(tpl, tc.expected) {
				t.Errorf("Case %s: Expected %v, got %v", tc.name, tc.expected, tpl)
			}
		})
	}
}

func TestParseFilterWithStringLiteralArgument(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected *Template
	}{
		// Basic case with a single filter and string literal argument
		{
			"SingleFilterWithStringLiteral",
			`{{ value|append:"!" }}`,
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "value",
						Filters: []Filter{
							{Name: "append", Args: []string{"!"}},
						},
						Text: `{{ value|append:"!" }}`,
					},
				},
			},
		},
		// Spaces around the filter argument
		{
			"SpaceAroundArgument",
			`{{ value|append: "!" }}`,
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "value",
						Filters: []Filter{
							{Name: "append", Args: []string{"!"}},
						},
						Text: `{{ value|append: "!" }}`,
					},
				},
			},
		},
		// Multiple filters with string literal arguments
		{
			"MultipleFiltersWithStringLiteral",
			`{{ greeting|replace:"Hello","Hi"|append:" everyone" }}`,
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "greeting",
						Filters: []Filter{
							{Name: "replace", Args: []string{"Hello", "Hi"}},
							{Name: "append", Args: []string{" everyone"}},
						},
						Text: `{{ greeting|replace:"Hello","Hi"|append:" everyone" }}`,
					},
				},
			},
		},
		// Text nodes around variable with filter and argument
		{
			"TextNodesAroundVariableWithFilter",
			`Welcome, {{ user|prepend:"Mr. " }}!`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "Welcome, "},
					{
						Type:     "variable",
						Variable: "user",
						Filters: []Filter{
							{Name: "prepend", Args: []string{"Mr. "}},
						},
						Text: `{{ user|prepend:"Mr. " }}`,
					},
					{Type: "text", Text: "!"},
				},
			},
		},
		// Complex scenario with mixed text and multiple variables with filters and one string argument
		{
			"ComplexMixedTextAndVariables",
			`Welcome, {{ user|prepend:"Mr. " }}! your score is {{ score|round }}.`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "Welcome, "},
					{
						Type:     "variable",
						Variable: "user",
						Filters: []Filter{
							{Name: "prepend", Args: []string{"Mr. "}},
						},
						Text: `{{ user|prepend:"Mr. " }}`,
					},
					{Type: "text", Text: "! your score is "},
					{
						Type:     "variable",
						Variable: "score",
						Filters: []Filter{
							{Name: "round"},
						},
						Text: `{{ score|round }}`,
					},
					{Type: "text", Text: "."},
				},
			},
		},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			if !reflect.DeepEqual(tpl, tc.expected) {
				t.Errorf("Case %s: Expected %v, got %v", tc.name, tc.expected, tpl)
			}
		})
	}
}

func TestParseFilterWithMultipleParameters(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected *Template
	}{
		// Basic case with a single filter and multiple string literal arguments
		{
			"SingleFilterWithMultipleArgs",
			"{{ value|replace:'hello','world' }}",
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "value",
						Filters: []Filter{
							{Name: "replace", Args: []string{"hello", "world"}},
						},
						Text: `{{ value|replace:'hello','world' }}`,
					},
				},
			},
		},
		// Spaces around arguments
		{
			"SpacesAroundArguments",
			"{{ value|replace: 'hello', 'world' }}",
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "value",
						Filters: []Filter{
							{Name: "replace", Args: []string{"hello", "world"}},
						},
						Text: `{{ value|replace: 'hello', 'world' }}`,
					},
				},
			},
		},
		// Multiple filters with multiple arguments
		{
			"MultipleFiltersWithMultipleArgs",
			"{{ greeting|replace:'Hello','Hi'|append: '!', ' Have a great day' }}",
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "greeting",
						Filters: []Filter{
							{Name: "replace", Args: []string{"Hello", "Hi"}},
							{Name: "append", Args: []string{"!", " Have a great day"}},
						},
						Text: `{{ greeting|replace:'Hello','Hi'|append: '!', ' Have a great day' }}`,
					},
				},
			},
		},
		// Complex scenario with mixed text and multiple variables with filters and multiple arguments
		{
			"ComplexMixedTextAndVariables",
			`Hello {{ name|capitalize }}, you have {{ unread|pluralize:"message","messages" }}.`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "Hello "},
					{
						Type:     "variable",
						Variable: "name",
						Filters:  []Filter{{Name: "capitalize"}},
						Text:     "{{ name|capitalize }}",
					},
					{Type: "text", Text: ", you have "},
					{
						Type:     "variable",
						Variable: "unread",
						Filters:  []Filter{{Name: "pluralize", Args: []string{"message", "messages"}}},
						Text:     `{{ unread|pluralize:"message","messages" }}`,
					},
					{Type: "text", Text: "."},
				},
			},
		},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			if !reflect.DeepEqual(tpl, tc.expected) {
				t.Errorf("Case %s: Expected %v, got %v", tc.name, tc.expected, tpl)
			}
		})
	}
}

func TestParseVariableWithMultiplePipelineFiltersWithMultipleParameters(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected *Template
	}{
		// Original case
		{
			"ReplaceAndAppend",
			`{{ username|replace:"hello","world"|append:"!" }}`,
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "username",
						Filters: []Filter{
							{Name: "replace", Args: []string{"hello", "world"}},
							{Name: "append", Args: []string{"!"}},
						},
						Text: `{{ username|replace:"hello","world"|append:"!" }}`,
					},
				},
			},
		},
		// Additional case with space around pipe symbols
		{
			"SpacesAroundPipes",
			`{{ username | replace:"hello","world" | append:"!" }}`,
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "username",
						Filters: []Filter{
							{Name: "replace", Args: []string{"hello", "world"}},
							{Name: "append", Args: []string{"!"}},
						},
						Text: `{{ username | replace:"hello","world" | append:"!" }}`,
					},
				},
			},
		},
		// Multiple filters with varied arguments
		{
			"MultipleFiltersVariedArgs",
			`{{ date|date:"YYYY-MM-DD"|prepend:"Date: " }}`,
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "date",
						Filters: []Filter{
							{Name: "date", Args: []string{"YYYY-MM-DD"}},
							{Name: "prepend", Args: []string{"Date: "}},
						},
						Text: `{{ date|date:"YYYY-MM-DD"|prepend:"Date: " }}`,
					},
				},
			},
		},
		// Complex scenario mixing text and multiple variables with multiple filters
		{
			"MixedTextMultipleVarsFilters",
			`Hello {{ name|capitalize|append:"!" }}, you have {{ unread|pluralize:"1 message","%d messages"|replace:"%d","many" }}.`,
			&Template{
				Nodes: []*Node{
					{Type: "text", Text: "Hello "},
					{
						Type:     "variable",
						Variable: "name",
						Filters: []Filter{
							{Name: "capitalize"},
							{Name: "append", Args: []string{"!"}},
						},
						Text: `{{ name|capitalize|append:"!" }}`,
					},
					{Type: "text", Text: ", you have "},
					{
						Type:     "variable",
						Variable: "unread",
						Filters: []Filter{
							{Name: "pluralize", Args: []string{"1 message", "%d messages"}},
							{Name: "replace", Args: []string{"%d", "many"}},
						},
						Text: `{{ unread|pluralize:"1 message","%d messages"|replace:"%d","many" }}`,
					},
					{Type: "text", Text: "."},
				},
			},
		},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			if !reflect.DeepEqual(tpl, tc.expected) {
				t.Errorf("Case %s: Expected %v, got %v", tc.name, tc.expected, tpl)
			}
		})
	}
}

func TestParseMultipleAdjacentVariables(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected *Template
	}{
		{
			"TwoVariablesNoSpace",
			"{{firstName}}{{lastName}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Text: "{{firstName}}"},
					{Type: "variable", Variable: "lastName", Text: "{{lastName}}"},
				},
			},
		},
		{
			"TwoVariablesSpaceBetween",
			"{{firstName}} {{lastName}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Text: "{{firstName}}"},
					{Type: "text", Text: " "},
					{Type: "variable", Variable: "lastName", Text: "{{lastName}}"},
				},
			},
		},
		{
			"ThreeVariablesMixedWhitespace",
			"{{firstName}}\t{{lastName}}\n{{email}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Text: "{{firstName}}"},
					{Type: "text", Text: "\t"},
					{Type: "variable", Variable: "lastName", Text: "{{lastName}}"},
					{Type: "text", Text: "\n"},
					{Type: "variable", Variable: "email", Text: "{{email}}"},
				},
			},
		},
		{
			"FourVariablesLineBreaks",
			"{{firstName}}\n{{lastName}}\n{{email}}\n{{username}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Text: "{{firstName}}"},
					{Type: "text", Text: "\n"},
					{Type: "variable", Variable: "lastName", Text: "{{lastName}}"},
					{Type: "text", Text: "\n"},
					{Type: "variable", Variable: "email", Text: "{{email}}"},
					{Type: "text", Text: "\n"},
					{Type: "variable", Variable: "username", Text: "{{username}}"},
				},
			},
		},
		{
			"TwoVariablesOneFilter",
			"{{firstName|upper}}{{lastName}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Filters: []Filter{{Name: "upper"}}, Text: "{{firstName|upper}}"},
					{Type: "variable", Variable: "lastName", Text: "{{lastName}}"},
				},
			},
		},
		{
			"VariablesWithFilterAndArgument",
			"{{user|default:'Anonymous'}}{{age|default:18}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "user", Filters: []Filter{{Name: "default", Args: []string{"Anonymous"}}}, Text: "{{user|default:'Anonymous'}}"},
					{Type: "variable", Variable: "age", Filters: []Filter{{Name: "default", Args: []string{"18"}}}, Text: "{{age|default:18}}"},
				},
			},
		},
		{
			"VariablesFilterPipeline",
			"{{firstName|trim|capitalize}}{{lastName|lower}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Filters: []Filter{{Name: "trim"}, {Name: "capitalize"}}, Text: "{{firstName|trim|capitalize}}"},
					{Type: "variable", Variable: "lastName", Filters: []Filter{{Name: "lower"}}, Text: "{{lastName|lower}}"},
				},
			},
		},
		{
			"ThreeVariablesWithMultipleFilters",
			"{{firstName|trim}}{{lastName|lower|capitalize}}{{age|default:30}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Filters: []Filter{{Name: "trim"}}, Text: "{{firstName|trim}}"},
					{Type: "variable", Variable: "lastName", Filters: []Filter{{Name: "lower"}, {Name: "capitalize"}}, Text: "{{lastName|lower|capitalize}}"},
					{Type: "variable", Variable: "age", Filters: []Filter{{Name: "default", Args: []string{"30"}}}, Text: "{{age|default:30}}"},
				},
			},
		},
		{
			"TwoVariablesTabBetweenWithFilter",
			"{{firstName|capitalize}}\t{{lastName|upper}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Filters: []Filter{{Name: "capitalize"}}, Text: "{{firstName|capitalize}}"},
					{Type: "text", Text: "\t"},
					{Type: "variable", Variable: "lastName", Filters: []Filter{{Name: "upper"}}, Text: "{{lastName|upper}}"},
				},
			},
		},
		{
			"VariablesWithLineBreakAndFilter",
			"{{firstName|lower}}\n{{lastName|upper}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Filters: []Filter{{Name: "lower"}}, Text: "{{firstName|lower}}"},
					{Type: "text", Text: "\n"},
					{Type: "variable", Variable: "lastName", Filters: []Filter{{Name: "upper"}}, Text: "{{lastName|upper}}"},
				},
			},
		},
		{
			"ThreeVariablesSpaceAndFilter",
			"{{firstName|upper}} {{lastName|upper}} {{email|upper}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Filters: []Filter{{Name: "upper"}}, Text: "{{firstName|upper}}"},
					{Type: "text", Text: " "},
					{Type: "variable", Variable: "lastName", Filters: []Filter{{Name: "upper"}}, Text: "{{lastName|upper}}"},
					{Type: "text", Text: " "},
					{Type: "variable", Variable: "email", Filters: []Filter{{Name: "upper"}}, Text: "{{email|upper}}"},
				},
			},
		},
		{
			"NestedVariablesWithFilters",
			"{{user.firstName|capitalize}}{{user.lastName|upper}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "user.firstName", Filters: []Filter{{Name: "capitalize"}}, Text: "{{user.firstName|capitalize}}"},
					{Type: "variable", Variable: "user.lastName", Filters: []Filter{{Name: "upper"}}, Text: "{{user.lastName|upper}}"},
				},
			},
		},
		{
			"VariablesWithMultipleFiltersAndWhitespace",
			"{{ firstName | trim | capitalize }}\n{{ lastName | lower | trim }}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Filters: []Filter{{Name: "trim"}, {Name: "capitalize"}}, Text: "{{ firstName | trim | capitalize }}"},
					{Type: "text", Text: "\n"},
					{Type: "variable", Variable: "lastName", Filters: []Filter{{Name: "lower"}, {Name: "trim"}}, Text: "{{ lastName | lower | trim }}"},
				},
			},
		},
		{
			"FourVariablesWithMixedFiltersAndWhitespace",
			"{{firstName|capitalize}} {{lastName|lower}}\t{{email|upper}}\n{{username|reverse}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Filters: []Filter{{Name: "capitalize"}}, Text: "{{firstName|capitalize}}"},
					{Type: "text", Text: " "},
					{Type: "variable", Variable: "lastName", Filters: []Filter{{Name: "lower"}}, Text: "{{lastName|lower}}"},
					{Type: "text", Text: "\t"},
					{Type: "variable", Variable: "email", Filters: []Filter{{Name: "upper"}}, Text: "{{email|upper}}"},
					{Type: "text", Text: "\n"},
					{Type: "variable", Variable: "username", Filters: []Filter{{Name: "reverse"}}, Text: "{{username|reverse}}"},
				},
			},
		},
		{
			"VariablesWithSpecialCharactersAndFilters",
			"{{firstName|capitalize|replace:'John','Jonathan'}}{{lastName|append:' Smith'}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Filters: []Filter{{Name: "capitalize"}, {Name: "replace", Args: []string{"John", "Jonathan"}}}, Text: "{{firstName|capitalize|replace:'John','Jonathan'}}"},
					{Type: "variable", Variable: "lastName", Filters: []Filter{{Name: "append", Args: []string{" Smith"}}}, Text: "{{lastName|append:' Smith'}}"},
				},
			},
		},
		{
			"NestedAndComplexFilters",
			"{{user.details.address.city|capitalize}}\n{{user.details.phoneNumber|default:'N/A'}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "user.details.address.city", Filters: []Filter{{Name: "capitalize"}}, Text: "{{user.details.address.city|capitalize}}"},
					{Type: "text", Text: "\n"},
					{Type: "variable", Variable: "user.details.phoneNumber", Filters: []Filter{{Name: "default", Args: []string{"N/A"}}}, Text: "{{user.details.phoneNumber|default:'N/A'}}"},
				},
			},
		},
		{
			"ComplexNestedVariablesWithMultipleFilters",
			"{{user.address|trim}} {{user.phone|default:'Unknown'|upper}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "user.address", Filters: []Filter{{Name: "trim"}}, Text: "{{user.address|trim}}"},
					{Type: "text", Text: " "},
					{Type: "variable", Variable: "user.phone", Filters: []Filter{{Name: "default", Args: []string{"Unknown"}}, {Name: "upper"}}, Text: "{{user.phone|default:'Unknown'|upper}}"},
				},
			},
		},
		{
			"MultipleAdjacentVariablesWithMixedFilters",
			"{{firstName|lower}}{{middleName|capitalize}}{{lastName|upper}}",
			&Template{
				Nodes: []*Node{
					{Type: "variable", Variable: "firstName", Filters: []Filter{{Name: "lower"}}, Text: "{{firstName|lower}}"},
					{Type: "variable", Variable: "middleName", Filters: []Filter{{Name: "capitalize"}}, Text: "{{middleName|capitalize}}"},
					{Type: "variable", Variable: "lastName", Filters: []Filter{{Name: "upper"}}, Text: "{{lastName|upper}}"},
				},
			},
		},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			if !reflect.DeepEqual(tpl, tc.expected) {
				t.Errorf("Case %s: Expected %+v, got %+v", tc.name, tc.expected, tpl)
			}
		})
	}
}

func TestParseVariableWithFilterHavingCommaInArguments(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		expected *Template
	}{
		{
			"DateFilterWithComma",
			`{{ current | date:"F j, Y" }}`,
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "current",
						Filters: []Filter{
							{Name: "date", Args: []string{"F j, Y"}},
						},
						Text: `{{ current | date:"F j, Y" }}`,
					},
				},
			},
		},
		{
			"DateFilterWithQuotedComma",
			`{{ current | date:'F j, Y' }}`,
			&Template{
				Nodes: []*Node{
					{
						Type:     "variable",
						Variable: "current",
						Filters: []Filter{
							{Name: "date", Args: []string{"F j, Y"}},
						},
						Text: `{{ current | date:'F j, Y' }}`,
					},
				},
			},
		},
	}

	parser := NewParser()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			if !reflect.DeepEqual(tpl, tc.expected) {
				t.Errorf("Case %s: Expected %+v, got %+v", tc.name, tc.expected, tpl)
			}
		})
	}
}

func TestParseMalformedVariableNodeAsText(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{
			"MissingClosingBracket",
			"Welcome back, {{username",
		},
		{
			"MissingOpeningBracket",
			"Hello, username}}!",
		},
		{
			"UnfinishedFilter",
			"Your account balance is {{balance|}} today.",
		},
		{
			"PipeWithoutFilterName",
			"Good morning, {{name| . Have a nice day!",
		},
		{
			"MissingFilterNameWithArguments",
			"Record: {{record||upper}}",
		},
		{
			"NestedBracesMalformed",
			"Error: {{user.details.{name}}",
		},
		{
			"MissingVariableName",
			"New message: {{|capitalize}}",
		},
		{
			"MalformedWithTextAround",
			"Hello, {{user|trim} in the system.",
		},
		{
			"MultipleMalformedInText",
			"Start {{of something |middle|end}} incomplete.",
		},
		{
			"SpaceBeforeClosingBracket",
			"Attempt: {{attempt | }}",
		},
		{
			"RandomCharactersInBraces",
			"Code: {{1234abcd!@#$}}",
		},
		{
			"MalformedFilterSyntax",
			"Discount: {{price|*0.85}}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser()
			tpl, err := parser.Parse(tc.source)
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			// Expecting the problematic part, or in some cases the entire input, to be treated as a text node
			expected := &Template{
				Nodes: []*Node{
					{Type: "text", Text: tc.source},
				},
			}

			if !reflect.DeepEqual(tpl, expected) {
				t.Errorf("Case %s: Expected %v, got %v", tc.name, expected, tpl)
			}
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
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			result, err := tmpl.Execute(ctx)
			if err == nil {
				t.Fatalf("Expected an error in %s, but got nil", tc.name)
			}
			if result != tc.expected {
				t.Errorf("Expected '%s', but got '%s'", tc.expected, result)
			}
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
			if err != nil {
				t.Fatalf("Unexpected error in %s: %v", tc.name, err)
			}

			result := tmpl.MustExecute(ctx)

			if result != tc.expected {
				t.Errorf("Expected '%s', but got '%s'", tc.expected, result)
			}
		})
	}
}
