package template

import (
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterTag(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := tagRegistry
	defer func() { tagRegistry = originalRegistry }()

	tests := []struct {
		name          string
		setupFunc     func()
		tagName       string
		parser        TagParser
		expectedError bool
		errorMessage  string
	}{
		{
			name: "register new tag successfully",
			setupFunc: func() {
				tagRegistry = make(map[string]TagParser)
			},
			tagName:       "customtag",
			parser:        parseIfTag, // Use existing parser as dummy
			expectedError: false,
		},
		{
			name: "register duplicate tag returns error",
			setupFunc: func() {
				tagRegistry = make(map[string]TagParser)
				tagRegistry["duplicate"] = parseIfTag
			},
			tagName:       "duplicate",
			parser:        parseForTag,
			expectedError: true,
			errorMessage:  fmt.Sprintf("%s: %q", ErrTagAlreadyRegistered, "duplicate"),
		},
		{
			name: "register multiple different tags",
			setupFunc: func() {
				tagRegistry = make(map[string]TagParser)
			},
			tagName:       "tag1",
			parser:        parseIfTag,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()
			err := RegisterTag(tt.tagName, tt.parser)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
				// Verify tag was registered
				parser, exists := tagRegistry[tt.tagName]
				assert.True(t, exists)
				assert.NotNil(t, parser)
			}
		})
	}
}

func TestGetTagParser(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := tagRegistry
	defer func() { tagRegistry = originalRegistry }()

	tests := []struct {
		name           string
		setupFunc      func()
		tagName        string
		expectedExists bool
	}{
		{
			name: "get existing tag",
			setupFunc: func() {
				tagRegistry = make(map[string]TagParser)
				tagRegistry["if"] = parseIfTag
			},
			tagName:        "if",
			expectedExists: true,
		},
		{
			name: "get non-existing tag",
			setupFunc: func() {
				tagRegistry = make(map[string]TagParser)
			},
			tagName:        "nonexistent",
			expectedExists: false,
		},
		{
			name: "get built-in for tag",
			setupFunc: func() {
				tagRegistry = make(map[string]TagParser)
				tagRegistry["for"] = parseForTag
			},
			tagName:        "for",
			expectedExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()
			parser, exists := GetTagParser(tt.tagName)
			assert.Equal(t, tt.expectedExists, exists)
			if tt.expectedExists {
				assert.NotNil(t, parser)
			} else {
				assert.Nil(t, parser)
			}
		})
	}
}

func TestParseIfTag(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedError  bool
		errorContains  string
		validateResult func(*testing.T, Statement)
	}{
		{
			name:          "simple if",
			template:      "{% if x %}yes{% endif %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				ifNode, ok := stmt.(*IfNode)
				assert.True(t, ok, "expected IfNode")
				assert.Equal(t, 1, len(ifNode.Branches))
				assert.Equal(t, 0, len(ifNode.ElseBody))

				// Check condition
				varNode, ok := ifNode.Branches[0].Condition.(*VariableNode)
				assert.True(t, ok, "expected VariableNode")
				assert.Equal(t, "x", varNode.Name)

				// Check body has content
				assert.Equal(t, 1, len(ifNode.Branches[0].Body))
			},
		},
		{
			name:          "if-else",
			template:      "{% if x %}yes{% else %}no{% endif %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				ifNode, ok := stmt.(*IfNode)
				assert.True(t, ok)
				assert.Equal(t, 1, len(ifNode.Branches))
				assert.Equal(t, 1, len(ifNode.ElseBody))
			},
		},
		{
			name:          "if-elif-else",
			template:      "{% if x > 5 %}big{% elif x > 0 %}small{% else %}zero{% endif %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				ifNode, ok := stmt.(*IfNode)
				assert.True(t, ok)
				assert.Equal(t, 2, len(ifNode.Branches))
				assert.Equal(t, 1, len(ifNode.ElseBody))
			},
		},
		{
			name:          "multiple elif",
			template:      "{% if x == 1 %}one{% elif x == 2 %}two{% elif x == 3 %}three{% endif %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				ifNode, ok := stmt.(*IfNode)
				assert.True(t, ok)
				assert.Equal(t, 3, len(ifNode.Branches))
				assert.Equal(t, 0, len(ifNode.ElseBody))
			},
		},
		{
			name:          "if with comparison",
			template:      "{% if count > 10 %}many{% endif %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				ifNode, ok := stmt.(*IfNode)
				assert.True(t, ok)
				assert.Equal(t, 1, len(ifNode.Branches))

				// Check it's a binary operation (comparison)
				_, ok = ifNode.Branches[0].Condition.(*BinaryOpNode)
				assert.True(t, ok, "expected BinaryOpNode for comparison")
			},
		},
		{
			name:          "if with logical and",
			template:      "{% if x > 0 and x < 10 %}valid{% endif %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				ifNode, ok := stmt.(*IfNode)
				assert.True(t, ok)
				assert.Equal(t, 1, len(ifNode.Branches))

				// Check it's a binary operation (logical and)
				_, ok = ifNode.Branches[0].Condition.(*BinaryOpNode)
				assert.True(t, ok, "expected BinaryOpNode for logical operation")
			},
		},
		{
			name:          "empty if body",
			template:      "{% if x %}{% endif %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				ifNode, ok := stmt.(*IfNode)
				assert.True(t, ok)
				assert.Equal(t, 1, len(ifNode.Branches))
				assert.Equal(t, 0, len(ifNode.Branches[0].Body))
			},
		},
		{
			name:          "extra tokens after condition",
			template:      "{% if x y %}yes{% endif %}",
			expectedError: true,
			errorContains: "unexpected tokens after condition",
		},
		{
			name:          "else with arguments",
			template:      "{% if x %}yes{% else x %}no{% endif %}",
			expectedError: true,
			errorContains: "else does not take arguments",
		},
		{
			name:          "endif with arguments",
			template:      "{% if x %}yes{% endif x %}",
			expectedError: true,
			errorContains: "endif does not take arguments",
		},
		{
			name:          "missing endif",
			template:      "{% if x %}yes",
			expectedError: true,
		},
		{
			name:          "elif after else",
			template:      "{% if x %}a{% else %}b{% elif y %}c{% endif %}",
			expectedError: true,
			errorContains: "unknown tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.template)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tmpl)

				if tt.validateResult != nil {
					assert.Equal(t, 1, len(tmpl.root))
					tt.validateResult(t, tmpl.root[0])
				}
			}
		})
	}
}

func TestParseForTag(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedError  bool
		errorContains  string
		validateResult func(*testing.T, Statement)
	}{
		{
			name:          "simple for loop",
			template:      "{% for item in items %}{{ item }}{% endfor %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				forNode, ok := stmt.(*ForNode)
				assert.True(t, ok, "expected ForNode")
				assert.Equal(t, []string{"item"}, forNode.LoopVars)
				assert.Equal(t, 1, len(forNode.Body))

				// Check collection
				varNode, ok := forNode.Collection.(*VariableNode)
				assert.True(t, ok, "expected VariableNode")
				assert.Equal(t, "items", varNode.Name)
			},
		},
		{
			name:          "for loop with key-value",
			template:      "{% for key, value in dict %}{{ key }}: {{ value }}{% endfor %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				forNode, ok := stmt.(*ForNode)
				assert.True(t, ok)
				assert.Equal(t, []string{"key", "value"}, forNode.LoopVars)
				assert.Equal(t, 3, len(forNode.Body)) // key, text(": "), value
			},
		},
		{
			name:          "for loop with index-item",
			template:      "{% for i, item in list %}{{ i }}{% endfor %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				forNode, ok := stmt.(*ForNode)
				assert.True(t, ok)
				assert.Equal(t, []string{"i", "item"}, forNode.LoopVars)
			},
		},
		{
			name:          "empty for body",
			template:      "{% for item in items %}{% endfor %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				forNode, ok := stmt.(*ForNode)
				assert.True(t, ok)
				assert.Equal(t, 0, len(forNode.Body))
			},
		},
		{
			name:          "for with complex collection",
			template:      "{% for item in user.items %}{{ item }}{% endfor %}",
			expectedError: false,
			validateResult: func(t *testing.T, stmt Statement) {
				forNode, ok := stmt.(*ForNode)
				assert.True(t, ok)

				// Check collection is property access
				_, ok = forNode.Collection.(*PropertyAccessNode)
				assert.True(t, ok, "expected PropertyAccessNode")
			},
		},
		{
			name:          "missing variable name",
			template:      "{% for in items %}{% endfor %}",
			expectedError: true,
			errorContains: "expected 'in' keyword",
		},
		{
			name:          "missing in keyword",
			template:      "{% for item items %}{% endfor %}",
			expectedError: true,
			errorContains: "expected 'in' keyword",
		},
		{
			name:          "missing second variable after comma",
			template:      "{% for item, in items %}{% endfor %}",
			expectedError: true,
			errorContains: "expected 'in' keyword",
		},
		{
			name:          "extra tokens after collection",
			template:      "{% for item in items extra %}{% endfor %}",
			expectedError: true,
			errorContains: "unexpected tokens after collection",
		},
		{
			name:          "endfor with arguments",
			template:      "{% for item in items %}{% endfor x %}",
			expectedError: true,
			errorContains: "endfor does not take arguments",
		},
		{
			name:          "missing endfor",
			template:      "{% for item in items %}{{ item }}",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.template)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tmpl)

				if tt.validateResult != nil {
					assert.Equal(t, 1, len(tmpl.root))
					tt.validateResult(t, tmpl.root[0])
				}
			}
		})
	}
}

func TestParseBreakTag(t *testing.T) {
	tests := []struct {
		name          string
		template      string
		expectedError bool
		errorContains string
	}{
		{
			name:          "break without arguments",
			template:      "{% for item in items %}{% break %}{% endfor %}",
			expectedError: false,
		},
		{
			name:          "break with arguments should fail",
			template:      "{% for item in items %}{% break now %}{% endfor %}",
			expectedError: true,
			errorContains: "break does not take arguments",
		},
		{
			name:          "multiple breaks in loop",
			template:      "{% for item in items %}{% if item %}{% break %}{% endif %}{% break %}{% endfor %}",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.template)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tmpl)

				// Validate break node exists in the tree
				forNode, ok := tmpl.root[0].(*ForNode)
				assert.True(t, ok, "expected ForNode")

				// Check if break node is in the body
				hasBreak := false
				for _, node := range forNode.Body {
					if _, ok := node.(*BreakNode); ok {
						hasBreak = true
						break
					}
					// Break might be inside an if
					if ifNode, ok := node.(*IfNode); ok {
						for _, branch := range ifNode.Branches {
							for _, n := range branch.Body {
								if _, ok := n.(*BreakNode); ok {
									hasBreak = true
									break
								}
							}
						}
					}
				}
				assert.True(t, hasBreak, "expected to find BreakNode in for loop")
			}
		})
	}
}

func TestParseContinueTag(t *testing.T) {
	tests := []struct {
		name          string
		template      string
		expectedError bool
		errorContains string
	}{
		{
			name:          "continue without arguments",
			template:      "{% for item in items %}{% continue %}{% endfor %}",
			expectedError: false,
		},
		{
			name:          "continue with arguments should fail",
			template:      "{% for item in items %}{% continue now %}{% endfor %}",
			expectedError: true,
			errorContains: "continue does not take arguments",
		},
		{
			name:          "multiple continues in loop",
			template:      "{% for item in items %}{% if item %}{% continue %}{% endif %}{% continue %}{% endfor %}",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.template)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tmpl)

				// Validate continue node exists in the tree
				forNode, ok := tmpl.root[0].(*ForNode)
				assert.True(t, ok, "expected ForNode")

				// Check if continue node is in the body
				hasContinue := false
				for _, node := range forNode.Body {
					if _, ok := node.(*ContinueNode); ok {
						hasContinue = true
						break
					}
					// Continue might be inside an if
					if ifNode, ok := node.(*IfNode); ok {
						for _, branch := range ifNode.Branches {
							for _, n := range branch.Body {
								if _, ok := n.(*ContinueNode); ok {
									hasContinue = true
									break
								}
							}
						}
					}
				}
				assert.True(t, hasContinue, "expected to find ContinueNode in for loop")
			}
		})
	}
}

func TestIfTagExecution(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "if true condition",
			template: "{% if x %}yes{% endif %}",
			data:     map[string]interface{}{"x": true},
			expected: "yes",
		},
		{
			name:     "if false condition",
			template: "{% if x %}yes{% endif %}",
			data:     map[string]interface{}{"x": false},
			expected: "",
		},
		{
			name:     "if-else true",
			template: "{% if x %}yes{% else %}no{% endif %}",
			data:     map[string]interface{}{"x": true},
			expected: "yes",
		},
		{
			name:     "if-else false",
			template: "{% if x %}yes{% else %}no{% endif %}",
			data:     map[string]interface{}{"x": false},
			expected: "no",
		},
		{
			name:     "if-elif first true",
			template: "{% if first %}one{% elif second %}two{% else %}other{% endif %}",
			data:     map[string]interface{}{"first": true, "second": false},
			expected: "one",
		},
		{
			name:     "if-elif second true",
			template: "{% if first %}one{% elif second %}two{% else %}other{% endif %}",
			data:     map[string]interface{}{"first": false, "second": true},
			expected: "two",
		},
		{
			name:     "if-elif else",
			template: "{% if first %}one{% elif second %}two{% else %}other{% endif %}",
			data:     map[string]interface{}{"first": false, "second": false},
			expected: "other",
		},
		{
			name:     "if with comparison",
			template: "{% if count > 5 %}many{% else %}few{% endif %}",
			data:     map[string]interface{}{"count": 10},
			expected: "many",
		},
		{
			name:     "nested if",
			template: "{% if outer %}{% if inner %}yes{% endif %}{% endif %}",
			data:     map[string]interface{}{"outer": true, "inner": true},
			expected: "yes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestForTagExecution(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "simple for loop",
			template: "{% for item in items %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3}},
			expected: "123",
		},
		{
			name:     "for loop with separator",
			template: "{% for item in items %}{{ item }},{% endfor %}",
			data:     map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "a,b,c,",
		},
		{
			name:     "empty collection",
			template: "{% for item in items %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{}},
			expected: "",
		},
		{
			name:     "for with key-value",
			template: "{% for k, v in dict %}{{ k }}:{{ v }};{% endfor %}",
			data:     map[string]interface{}{"dict": map[string]int{"a": 1, "b": 2}},
			expected: "a:1;b:2;", // Note: map iteration order is not guaranteed in Go
		},
		{
			name:     "for map single var binds key",
			template: "{% for item in dict %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"dict": map[string]int{"a": 1, "b": 2}},
			expected: "ab",
		},
		{
			name:     "nested for loops",
			template: "{% for i in outer %}{% for j in inner %}{{ i }}{{ j }}{% endfor %}{% endfor %}",
			data:     map[string]interface{}{"outer": []int{1, 2}, "inner": []string{"a", "b"}},
			expected: "1a1b2a2b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)

			// For map iteration, order is not guaranteed, so we check if result contains expected parts
			if tt.name == "for with key-value" {
				assert.Contains(t, result, "a:1")
				assert.Contains(t, result, "b:2")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestBreakTagExecution(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "break immediately",
			template: "{% for item in items %}{% break %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3}},
			expected: "",
		},
		{
			name:     "break after first item",
			template: "{% for item in items %}{{ item }}{% break %}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3}},
			expected: "1",
		},
		{
			name:     "conditional break",
			template: "{% for item in items %}{% if item > 2 %}{% break %}{% endif %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3, 4}},
			expected: "12",
		},
		{
			name:     "break in nested loop",
			template: "{% for i in outer %}{% for j in inner %}{{ i }}{{ j }}{% if j == 'b' %}{% break %}{% endif %}{% endfor %}{% endfor %}",
			data:     map[string]interface{}{"outer": []int{1, 2}, "inner": []string{"a", "b", "c"}},
			expected: "1a1b2a2b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContinueTagExecution(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "continue immediately",
			template: "{% for item in items %}{% continue %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3}},
			expected: "",
		},
		{
			name:     "continue after output",
			template: "{% for item in items %}{{ item }}{% continue %}X{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3}},
			expected: "123",
		},
		{
			name:     "conditional continue",
			template: "{% for item in items %}{% if item > 1 and item < 3 %}{% continue %}{% endif %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3}},
			expected: "13",
		},
		{
			name:     "continue in nested loop",
			template: "{% for i in outer %}{% for j in inner %}{% if j %}{% continue %}{% endif %}X{% endfor %}{% endfor %}",
			data:     map[string]interface{}{"outer": []int{1, 2}, "inner": []bool{true, true}},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComplexTagCombinations(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "if inside for",
			template: `{% for item in items %}{% if item > 2 %}{{ item }}{% endif %}{% endfor %}`,
			data: map[string]interface{}{
				"items": []int{1, 2, 3, 4},
			},
			expected: "34",
		},
		{
			name:     "for inside if",
			template: `{% if show %}{% for item in items %}{{ item }}{% endfor %}{% endif %}`,
			data: map[string]interface{}{
				"show":  true,
				"items": []int{1, 2, 3},
			},
			expected: "123",
		},
		{
			name:     "nested if-for with break",
			template: `{% if show %}{% for item in items %}{% if item > 2 %}{% break %}{% endif %}{{ item }}{% endfor %}{% endif %}`,
			data: map[string]interface{}{
				"show":  true,
				"items": []int{1, 2, 3, 4},
			},
			expected: "12",
		},
		{
			name:     "nested for with continue",
			template: `{% for i in outer %}{% for j in inner %}{% if j > 1 and j < 3 %}{% continue %}{% endif %}{{ i }}{{ j }}{% endfor %}{% endfor %}`,
			data: map[string]interface{}{
				"outer": []int{1, 2},
				"inner": []int{1, 2, 3},
			},
			expected: "11132123", // Skip j=2 for both i=1 and i=2
		},
		{
			name:     "multiple elif with for",
			template: `{% if first %}one{% elif second %}{% for i in items %}{{ i }}{% endfor %}{% else %}other{% endif %}`,
			data: map[string]interface{}{
				"first":  false,
				"second": true,
				"items":  []string{"a", "b"},
			},
			expected: "ab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTagNodeStructure(t *testing.T) {
	tests := []struct {
		name     string
		template string
		validate func(*testing.T, *Template)
	}{
		{
			name:     "if node structure",
			template: "{% if x %}yes{% endif %}",
			validate: func(t *testing.T, tmpl *Template) {
				// Find the IfNode in the root statements
				var ifNode *IfNode
				for _, stmt := range tmpl.root {
					if node, ok := stmt.(*IfNode); ok {
						ifNode = node
						break
					}
				}
				assert.NotNil(t, ifNode, "IfNode should exist in root")

				// Validate structure using reflection
				expectedBranches := 1
				expectedElseBody := 0
				assert.Equal(t, expectedBranches, len(ifNode.Branches))
				assert.Equal(t, expectedElseBody, len(ifNode.ElseBody))

				// Check position info
				line, col := ifNode.Position()
				assert.Greater(t, line, 0)
				assert.Greater(t, col, 0)
			},
		},
		{
			name:     "for node structure",
			template: "{% for item in items %}x{% endfor %}",
			validate: func(t *testing.T, tmpl *Template) {
				// Find the ForNode in the root statements
				var forNode *ForNode
				for _, stmt := range tmpl.root {
					if node, ok := stmt.(*ForNode); ok {
						forNode = node
						break
					}
				}
				assert.NotNil(t, forNode, "ForNode should exist in root")

				// Validate structure
				expectedLoopVars := []string{"item"}
				assert.True(t, reflect.DeepEqual(expectedLoopVars, forNode.LoopVars))
				assert.NotNil(t, forNode.Collection)
				assert.GreaterOrEqual(t, len(forNode.Body), 1)

				// Check position info
				line, col := forNode.Position()
				assert.Greater(t, line, 0)
				assert.Greater(t, col, 0)
			},
		},
		{
			name:     "break node structure",
			template: "{% for item in items %}{% break %}{% endfor %}",
			validate: func(t *testing.T, tmpl *Template) {
				forNode := tmpl.root[0].(*ForNode)
				breakNode, ok := forNode.Body[0].(*BreakNode)
				assert.True(t, ok)

				// Check position info
				line, col := breakNode.Position()
				assert.Equal(t, 1, line)
				assert.Greater(t, col, 0)

				// Check String() method
				assert.Equal(t, "Break", breakNode.String())
			},
		},
		{
			name:     "continue node structure",
			template: "{% for item in items %}{% continue %}{% endfor %}",
			validate: func(t *testing.T, tmpl *Template) {
				forNode := tmpl.root[0].(*ForNode)
				continueNode, ok := forNode.Body[0].(*ContinueNode)
				assert.True(t, ok)

				// Check position info
				line, col := continueNode.Position()
				assert.Equal(t, 1, line)
				assert.Greater(t, col, 0)

				// Check String() method
				assert.Equal(t, "Continue", continueNode.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.template)
			assert.NoError(t, err)
			tt.validate(t, tmpl)
		})
	}
}

// Moved from integration_test.go
func TestIntegration_BasicTemplateRendering(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "plain text",
			template: "Hello, world!",
			data:     map[string]interface{}{},
			expected: "Hello, world!",
		},
		{
			name:     "simple variable",
			template: "Hello, {{ name }}!",
			data:     map[string]interface{}{"name": "Alice"},
			expected: "Hello, Alice!",
		},
		{
			name:     "multiple variables",
			template: "{{ greeting }}, {{ name }}!",
			data:     map[string]interface{}{"greeting": "Hello", "name": "Bob"},
			expected: "Hello, Bob!",
		},
		{
			name:     "variable with number",
			template: "Age: {{ age }}",
			data:     map[string]interface{}{"age": 30},
			expected: "Age: 30",
		},
		{
			name:     "nested property",
			template: "{{ user.name }}",
			data: map[string]interface{}{
				"user": map[string]interface{}{"name": "Charlie"},
			},
			expected: "Charlie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Moved from integration_test.go
func TestIntegration_IfElseConditions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "if true",
			template: "{% if show %}visible{% endif %}",
			data:     map[string]interface{}{"show": true},
			expected: "visible",
		},
		{
			name:     "if false",
			template: "{% if show %}visible{% endif %}",
			data:     map[string]interface{}{"show": false},
			expected: "",
		},
		{
			name:     "if-else true branch",
			template: "{% if flag %}yes{% else %}no{% endif %}",
			data:     map[string]interface{}{"flag": true},
			expected: "yes",
		},
		{
			name:     "if-else false branch",
			template: "{% if flag %}yes{% else %}no{% endif %}",
			data:     map[string]interface{}{"flag": false},
			expected: "no",
		},
		{
			name:     "if-elif-else first",
			template: "{% if x %}first{% elif y %}second{% else %}third{% endif %}",
			data:     map[string]interface{}{"x": true, "y": false},
			expected: "first",
		},
		{
			name:     "if-elif-else second",
			template: "{% if x %}first{% elif y %}second{% else %}third{% endif %}",
			data:     map[string]interface{}{"x": false, "y": true},
			expected: "second",
		},
		{
			name:     "if-elif-else third",
			template: "{% if x %}first{% elif y %}second{% else %}third{% endif %}",
			data:     map[string]interface{}{"x": false, "y": false},
			expected: "third",
		},
		{
			name:     "nested if",
			template: "{% if outer %}{% if inner %}both{% endif %}{% endif %}",
			data:     map[string]interface{}{"outer": true, "inner": true},
			expected: "both",
		},
		{
			name:     "if with comparison",
			template: "{% if count > 5 %}many{% else %}few{% endif %}",
			data:     map[string]interface{}{"count": 10},
			expected: "many",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Moved from integration_test.go
func TestIntegration_ForLoops(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "simple for loop",
			template: "{% for item in items %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3}},
			expected: "123",
		},
		{
			name:     "for loop with separator",
			template: "{% for item in items %}{{ item }},{% endfor %}",
			data:     map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "a,b,c,",
		},
		{
			name:     "for loop with index",
			template: "{% for i, item in items %}{{ i }}:{{ item }};{% endfor %}",
			data:     map[string]interface{}{"items": []string{"x", "y"}},
			expected: "0:x;1:y;",
		},
		{
			name:     "empty for loop",
			template: "{% for item in items %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{}},
			expected: "",
		},
		{
			name:     "nested for loops",
			template: "{% for i in outer %}{% for j in inner %}{{ i }}{{ j }}{% endfor %}{% endfor %}",
			data:     map[string]interface{}{"outer": []int{1, 2}, "inner": []string{"a", "b"}},
			expected: "1a1b2a2b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Moved from integration_test.go
func TestIntegration_BreakContinue(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "break in loop",
			template: "{% for item in items %}{% if item > 2 %}{% break %}{% endif %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3, 4}},
			expected: "12",
		},
		{
			name:     "continue in loop",
			template: "{% for item in items %}{% if item > 1 and item < 4 %}{% continue %}{% endif %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3, 4}},
			expected: "14",
		},
		{
			name:     "break immediately",
			template: "{% for item in items %}{% break %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3}},
			expected: "",
		},
		{
			name:     "continue all iterations",
			template: "{% for item in items %}{% continue %}{{ item }}{% endfor %}",
			data:     map[string]interface{}{"items": []int{1, 2, 3}},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Moved from integration_test.go
func TestIntegration_Filters(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "upper filter",
			template: "{{ name|upper }}",
			data:     map[string]interface{}{"name": "alice"},
			expected: "ALICE",
		},
		{
			name:     "lower filter",
			template: "{{ name|lower }}",
			data:     map[string]interface{}{"name": "ALICE"},
			expected: "alice",
		},
		{
			name:     "capitalize filter",
			template: "{{ name|capitalize }}",
			data:     map[string]interface{}{"name": "alice"},
			expected: "Alice",
		},
		{
			name:     "length filter",
			template: "{{ items|length }}",
			data:     map[string]interface{}{"items": []int{1, 2, 3, 4, 5}},
			expected: "5",
		},
		{
			name:     "default filter with value",
			template: `{{ name|default:"Anonymous" }}`,
			data:     map[string]interface{}{"name": "Alice"},
			expected: "Alice",
		},
		{
			name:     "default filter without value",
			template: `{{ name|default:"Anonymous" }}`,
			data:     map[string]interface{}{},
			expected: "Anonymous",
		},
		{
			name:     "chained filters",
			template: "{{ name|lower|capitalize }}",
			data:     map[string]interface{}{"name": "ALICE"},
			expected: "Alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Moved from integration_test.go
func TestIntegration_ComplexStructures(t *testing.T) {
	type User struct {
		Name    string
		Age     int
		Email   string
		Active  bool
		Profile struct {
			Bio     string
			Website string
		}
	}

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "struct field access",
			template: "{{ user.Name }}",
			data: map[string]interface{}{
				"user": User{Name: "Alice", Age: 30},
			},
			expected: "Alice",
		},
		{
			name:     "nested struct field",
			template: "{{ user.Profile.Bio }}",
			data: map[string]interface{}{
				"user": User{
					Name: "Alice",
					Profile: struct {
						Bio     string
						Website string
					}{Bio: "Software Engineer", Website: "https://example.com"},
				},
			},
			expected: "Software Engineer",
		},
		{
			name:     "struct in loop",
			template: "{% for user in users %}{{ user.Name }},{% endfor %}",
			data: map[string]interface{}{
				"users": []User{
					{Name: "Alice"},
					{Name: "Bob"},
					{Name: "Charlie"},
				},
			},
			expected: "Alice,Bob,Charlie,",
		},
		{
			name:     "struct with if condition",
			template: "{% if user.Active %}{{ user.Name }} is active{% endif %}",
			data: map[string]interface{}{
				"user": User{Name: "Alice", Active: true},
			},
			expected: "Alice is active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Moved from integration_test.go
func TestIntegration_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		data        map[string]interface{}
		expectError bool
		expected    string
	}{
		{
			name:        "empty template",
			template:    "",
			data:        map[string]interface{}{},
			expectError: false,
			expected:    "",
		},
		{
			name:        "whitespace only",
			template:    "   \n\t  ",
			data:        map[string]interface{}{},
			expectError: false,
			expected:    "   \n\t  ",
		},
		{
			name:        "unclosed variable tag",
			template:    "{{ name",
			data:        map[string]interface{}{"name": "Alice"},
			expectError: true,
		},
		{
			name:        "unclosed block tag",
			template:    "{% if x",
			data:        map[string]interface{}{"x": true},
			expectError: true,
		},
		{
			name:        "missing endif",
			template:    "{% if x %}yes",
			data:        map[string]interface{}{"x": true},
			expectError: true,
		},
		{
			name:        "missing endfor",
			template:    "{% for item in items %}{{ item }}",
			data:        map[string]interface{}{"items": []int{1, 2}},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Moved from integration_test.go
func TestIntegration_ComplexTemplates(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "user profile template",
			template: `
Name: {{ user.name }}
Age: {{ user.age }}
{% if user.active %}Status: Active{% else %}Status: Inactive{% endif %}
`,
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name":   "Alice",
					"age":    30,
					"active": true,
				},
			},
			expected: `
Name: Alice
Age: 30
Status: Active
`,
		},
		{
			name: "list rendering with conditions",
			template: `{% for item in items %}{% if item > 5 %}{{ item }} is big
{% else %}{{ item }} is small
{% endif %}{% endfor %}`,
			data: map[string]interface{}{
				"items": []int{3, 7, 4, 9},
			},
			expected: `3 is small
7 is big
4 is small
9 is big
`,
		},
		{
			name: "nested loops with break",
			template: `{% for i in rows %}Row {{ i }}:{% for j in cols %}{% if j > 2 %}{% break %}{% endif %} {{ j }}{% endfor %}
{% endfor %}`,
			data: map[string]interface{}{
				"rows": []int{1, 2},
				"cols": []int{1, 2, 3, 4},
			},
			expected: `Row 1: 1 2
Row 2: 1 2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Moved from integration_test.go
func TestIntegration_ArithmeticAndComparison(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "addition",
			template: "{{ a + b }}",
			data:     map[string]interface{}{"a": 5, "b": 3},
			expected: "8",
		},
		{
			name:     "subtraction",
			template: "{{ a - b }}",
			data:     map[string]interface{}{"a": 10, "b": 3},
			expected: "7",
		},
		{
			name:     "multiplication",
			template: "{{ a * b }}",
			data:     map[string]interface{}{"a": 4, "b": 3},
			expected: "12",
		},
		{
			name:     "division",
			template: "{{ a / b }}",
			data:     map[string]interface{}{"a": 10, "b": 2},
			expected: "5",
		},
		{
			name:     "modulo",
			template: "{{ a % b }}",
			data:     map[string]interface{}{"a": 10, "b": 3},
			expected: "1",
		},
		{
			name:     "comparison greater than",
			template: "{% if a > b %}yes{% else %}no{% endif %}",
			data:     map[string]interface{}{"a": 10, "b": 5},
			expected: "yes",
		},
		{
			name:     "comparison less than",
			template: "{% if a < b %}yes{% else %}no{% endif %}",
			data:     map[string]interface{}{"a": 3, "b": 5},
			expected: "yes",
		},
		{
			name:     "logical and",
			template: "{% if a and b %}yes{% else %}no{% endif %}",
			data:     map[string]interface{}{"a": true, "b": true},
			expected: "yes",
		},
		{
			name:     "logical or",
			template: "{% if a or b %}yes{% else %}no{% endif %}",
			data:     map[string]interface{}{"a": false, "b": true},
			expected: "yes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// saveTagRegistry saves and returns the current tag registry for later restoration.
func saveTagRegistry() map[string]TagParser {
	saved := tagRegistry
	tagRegistry = make(map[string]TagParser)
	// Copy built-in tags so tests start from a known state.
	for k, v := range saved {
		tagRegistry[k] = v
	}
	return saved
}

func restoreTagRegistry(saved map[string]TagParser) {
	tagRegistry = saved
}

func TestRegisterTagExternalStatement(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	// Register a "set" tag demonstrating that external packages
	// can now implement Statement directly.
	UnregisterTag("set")
	err := RegisterTag("set", func(doc *Parser, start *Token, arguments *Parser) (Statement, error) {
		varToken, err := arguments.ExpectIdentifier()
		if err != nil {
			return nil, arguments.Error("expected variable name after 'set'")
		}

		if arguments.Match(TokenSymbol, "=") == nil {
			return nil, arguments.Error("expected '=' after variable name")
		}

		expr, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}

		if arguments.Remaining() > 0 {
			return nil, arguments.Error("unexpected tokens after expression")
		}

		return &testSetNode{
			varName:    varToken.Value,
			expression: expr,
			line:       start.Line,
			col:        start.Col,
		}, nil
	})
	assert.NoError(t, err)

	result, err := Render(`{% set greeting = "Hello" %}{{ greeting }}, World!`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", result)
}

// testSetNode is a Statement implementation used in TestRegisterTagLowLevel.
type testSetNode struct {
	varName    string
	expression Expression
	line       int
	col        int
}

func (n *testSetNode) Position() (int, int) { return n.line, n.col }
func (n *testSetNode) String() string       { return fmt.Sprintf("Set(%s)", n.varName) }

func (n *testSetNode) Execute(ctx *ExecutionContext, _ io.Writer) error {
	val, err := n.expression.Evaluate(ctx)
	if err != nil {
		return err
	}
	ctx.Set(n.varName, val.Interface())
	return nil
}

func TestListTags(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	tags := ListTags()
	// Built-in tags should be present.
	assert.Contains(t, tags, "if")
	assert.Contains(t, tags, "for")
	assert.Contains(t, tags, "break")
	assert.Contains(t, tags, "continue")

	// Should be sorted.
	for i := 1; i < len(tags); i++ {
		assert.LessOrEqual(t, tags[i-1], tags[i])
	}
}

func TestHasTag(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	assert.True(t, HasTag("if"))
	assert.True(t, HasTag("for"))
	assert.False(t, HasTag("nonexistent"))
}

func TestUnregisterTag(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	assert.True(t, HasTag("if"))
	UnregisterTag("if")
	assert.False(t, HasTag("if"))

	// Unregistering a non-existent tag is a no-op.
	UnregisterTag("nonexistent")
}

func TestDuplicateRegistration(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	// Registering a tag with the same name as a built-in should fail.
	err := RegisterTag("if", parseIfTag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrTagAlreadyRegistered.Error())

	err = RegisterTag("for", parseForTag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrTagAlreadyRegistered.Error())
}

func TestTagRegistrySaveRestore(t *testing.T) {
	// Verify that save/restore works correctly for test isolation.
	saved := saveTagRegistry()

	// Add a custom tag.
	err := RegisterTag("mytesttag", func(doc *Parser, start *Token, arguments *Parser) (Statement, error) {
		return NewTextNode("", start.Line, start.Col), nil
	})
	assert.NoError(t, err)
	assert.True(t, HasTag("mytesttag"))

	// Restore.
	restoreTagRegistry(saved)

	// Custom tag should be gone.
	assert.False(t, HasTag("mytesttag"))
	// Built-in tags should still be present.
	assert.True(t, HasTag("if"))
}
