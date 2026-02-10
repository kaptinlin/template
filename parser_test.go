package template

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Group 1: Basic Text Node Parsing
// =============================================================================

func TestParserTextNode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Statement
	}{
		{
			name:  "Plain text",
			input: "Hello, World!",
			expected: []Statement{
				&TextNode{Text: "Hello, World!", Line: 1, Col: 1},
			},
		},
		{
			name:  "Multiple lines of text",
			input: "Line 1\nLine 2\nLine 3",
			expected: []Statement{
				&TextNode{Text: "Line 1\nLine 2\nLine 3", Line: 1, Col: 1},
			},
		},
		{
			name:  "Text with special characters",
			input: "Hello <html> & \"quotes\"",
			expected: []Statement{
				&TextNode{Text: "Hello <html> & \"quotes\"", Line: 1, Col: 1},
			},
		},
		{
			name:     "Empty template",
			input:    "",
			expected: nil, // Parse returns nil for empty input
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			nodes, err := parser.Parse()
			require.NoError(t, err)

			assert.Equal(t, tt.expected, nodes)
		})
	}
}

// =============================================================================
// Test Group 2: Output Node (Variable) Parsing
// =============================================================================

func TestParserOutputNode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Statement
	}{
		{
			name:  "Simple variable",
			input: "{{ name }}",
			expected: []Statement{
				&OutputNode{
					Expression: &VariableNode{Name: "name", Line: 1, Col: 4},
					Line:       1,
					Col:        1,
				},
			},
		},
		{
			name:  "Variable with number",
			input: "{{ 42 }}",
			expected: []Statement{
				&OutputNode{
					Expression: &LiteralNode{Value: 42.0, Line: 1, Col: 4},
					Line:       1,
					Col:        1,
				},
			},
		},
		{
			name:  "Variable with string",
			input: `{{ "hello" }}`,
			expected: []Statement{
				&OutputNode{
					Expression: &LiteralNode{Value: "hello", Line: 1, Col: 4},
					Line:       1,
					Col:        1,
				},
			},
		},
		{
			name:  "Variable with boolean",
			input: "{{ true }}",
			expected: []Statement{
				&OutputNode{
					Expression: &LiteralNode{Value: true, Line: 1, Col: 4},
					Line:       1,
					Col:        1,
				},
			},
		},
		{
			name:  "Variable with property access",
			input: "{{ user.name }}",
			expected: []Statement{
				&OutputNode{
					Expression: &PropertyAccessNode{
						Object:   &VariableNode{Name: "user", Line: 1, Col: 4},
						Property: "name",
						Line:     1,
						Col:      8,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "Variable with subscript",
			input: "{{ items[0] }}",
			expected: []Statement{
				&OutputNode{
					Expression: &SubscriptNode{
						Object: &VariableNode{Name: "items", Line: 1, Col: 4},
						Index:  &LiteralNode{Value: 0.0, Line: 1, Col: 10},
						Line:   1,
						Col:    9,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "Variable with filter",
			input: "{{ name|upper }}",
			expected: []Statement{
				&OutputNode{
					Expression: &FilterNode{
						Expression: &VariableNode{Name: "name", Line: 1, Col: 4},
						FilterName: "upper",
						Args:       nil, // Args is nil, not empty slice
						Line:       1,
						Col:        8,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "Variable with filter and argument",
			input: "{{ price|add:10 }}",
			expected: []Statement{
				&OutputNode{
					Expression: &FilterNode{
						Expression: &VariableNode{Name: "price", Line: 1, Col: 4},
						FilterName: "add",
						Args: []Expression{
							&LiteralNode{Value: 10.0, Line: 1, Col: 14},
						},
						Line: 1,
						Col:  9, // Col of "|" is 9, not 10
					},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "Variable with binary operation",
			input: "{{ a + b }}",
			expected: []Statement{
				&OutputNode{
					Expression: &BinaryOpNode{
						Operator: "+",
						Left:     &VariableNode{Name: "a", Line: 1, Col: 4},
						Right:    &VariableNode{Name: "b", Line: 1, Col: 8},
						Line:     1,
						Col:      6,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "Variable with comparison",
			input: "{{ x > 10 }}",
			expected: []Statement{
				&OutputNode{
					Expression: &BinaryOpNode{
						Operator: ">",
						Left:     &VariableNode{Name: "x", Line: 1, Col: 4},
						Right:    &LiteralNode{Value: 10.0, Line: 1, Col: 8},
						Line:     1,
						Col:      6,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			nodes, err := parser.Parse()
			require.NoError(t, err)

			assert.Equal(t, tt.expected, nodes)
		})
	}
}

// =============================================================================
// Test Group 3: If Tag Parsing
// =============================================================================

func TestParserIfTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Statement
	}{
		{
			name:  "Simple if",
			input: "{% if x %}yes{% endif %}",
			expected: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&TextNode{Text: "yes", Line: 1, Col: 11},
							},
						},
					},
					ElseBody: nil,
					Line:     1,
					Col:      4, // Col of "if" token, not "{%"
				},
			},
		},
		{
			name:  "If with else",
			input: "{% if x %}yes{% else %}no{% endif %}",
			expected: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&TextNode{Text: "yes", Line: 1, Col: 11},
							},
						},
					},
					ElseBody: []Node{
						&TextNode{Text: "no", Line: 1, Col: 24},
					},
					Line: 1,
					Col:  4, // Col of "if" token
				},
			},
		},
		{
			name:  "If with elif",
			input: "{% if x %}a{% elif y %}b{% endif %}",
			expected: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&TextNode{Text: "a", Line: 1, Col: 11},
							},
						},
						{
							Condition: &VariableNode{Name: "y", Line: 1, Col: 20},
							Body: []Node{
								&TextNode{Text: "b", Line: 1, Col: 24},
							},
						},
					},
					ElseBody: nil,
					Line:     1,
					Col:      4, // Col of "if" token
				},
			},
		},
		{
			name:  "If with elif and else",
			input: "{% if x %}a{% elif y %}b{% else %}c{% endif %}",
			expected: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&TextNode{Text: "a", Line: 1, Col: 11},
							},
						},
						{
							Condition: &VariableNode{Name: "y", Line: 1, Col: 20},
							Body: []Node{
								&TextNode{Text: "b", Line: 1, Col: 24},
							},
						},
					},
					ElseBody: []Node{
						&TextNode{Text: "c", Line: 1, Col: 35},
					},
					Line: 1,
					Col:  4, // Col of "if" token
				},
			},
		},
		{
			name:  "If with multiple elif",
			input: "{% if x %}a{% elif y %}b{% elif z %}c{% endif %}",
			expected: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&TextNode{Text: "a", Line: 1, Col: 11},
							},
						},
						{
							Condition: &VariableNode{Name: "y", Line: 1, Col: 20},
							Body: []Node{
								&TextNode{Text: "b", Line: 1, Col: 24},
							},
						},
						{
							Condition: &VariableNode{Name: "z", Line: 1, Col: 33},
							Body: []Node{
								&TextNode{Text: "c", Line: 1, Col: 37},
							},
						},
					},
					ElseBody: nil,
					Line:     1,
					Col:      4, // Col of "if" token
				},
			},
		},
		{
			name:  "If with condition expression",
			input: "{% if x > 0 %}positive{% endif %}",
			expected: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &BinaryOpNode{
								Operator: ">",
								Left:     &VariableNode{Name: "x", Line: 1, Col: 7},
								Right:    &LiteralNode{Value: 0.0, Line: 1, Col: 11},
								Line:     1,
								Col:      9,
							},
							Body: []Node{
								&TextNode{Text: "positive", Line: 1, Col: 15},
							},
						},
					},
					ElseBody: nil,
					Line:     1,
					Col:      4, // Col of "if" token
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			nodes, err := parser.Parse()
			require.NoError(t, err)

			assert.Equal(t, tt.expected, nodes)
		})
	}
}

// =============================================================================
// Test Group 4: For Tag Parsing
// =============================================================================

func TestParserForTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Statement
	}{
		{
			name:  "Simple for loop",
			input: "{% for item in items %}{{ item }}{% endfor %}",
			expected: []Statement{
				&ForNode{
					LoopVars:   []string{"item"},
					Collection: &VariableNode{Name: "items", Line: 1, Col: 16},
					Body: []Node{
						&OutputNode{
							Expression: &VariableNode{Name: "item", Line: 1, Col: 27},
							Line:       1,
							Col:        24,
						},
					},
					Line: 1,
					Col:  4, // Col of "for" token
				},
			},
		},
		{
			name:  "For loop with two variables",
			input: "{% for key, value in dict %}{{ key }}: {{ value }}{% endfor %}",
			expected: []Statement{
				&ForNode{
					LoopVars:   []string{"key", "value"},
					Collection: &VariableNode{Name: "dict", Line: 1, Col: 22},
					Body: []Node{
						&OutputNode{
							Expression: &VariableNode{Name: "key", Line: 1, Col: 32},
							Line:       1,
							Col:        29,
						},
						&TextNode{Text: ": ", Line: 1, Col: 38},
						&OutputNode{
							Expression: &VariableNode{Name: "value", Line: 1, Col: 43},
							Line:       1,
							Col:        40,
						},
					},
					Line: 1,
					Col:  4, // Col of "for" token
				},
			},
		},
		{
			name:  "For loop with text",
			input: "{% for i in nums %}Item: {{ i }}{% endfor %}",
			expected: []Statement{
				&ForNode{
					LoopVars:   []string{"i"},
					Collection: &VariableNode{Name: "nums", Line: 1, Col: 13},
					Body: []Node{
						&TextNode{Text: "Item: ", Line: 1, Col: 20},
						&OutputNode{
							Expression: &VariableNode{Name: "i", Line: 1, Col: 29},
							Line:       1,
							Col:        26,
						},
					},
					Line: 1,
					Col:  4, // Col of "for" token
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			nodes, err := parser.Parse()
			require.NoError(t, err)

			assert.Equal(t, tt.expected, nodes)
		})
	}
}

// =============================================================================
// Test Group 5: Break and Continue Tag Parsing
// =============================================================================

func TestParserBreakContinueTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Statement
	}{
		{
			name:  "Break in for loop",
			input: "{% for i in nums %}{% break %}{% endfor %}",
			expected: []Statement{
				&ForNode{
					LoopVars:   []string{"i"},
					Collection: &VariableNode{Name: "nums", Line: 1, Col: 13},
					Body: []Node{
						&BreakNode{Line: 1, Col: 23}, // Col of "break" token
					},
					Line: 1,
					Col:  4, // Col of "for" token
				},
			},
		},
		{
			name:  "Continue in for loop",
			input: "{% for i in nums %}{% continue %}{% endfor %}",
			expected: []Statement{
				&ForNode{
					LoopVars:   []string{"i"},
					Collection: &VariableNode{Name: "nums", Line: 1, Col: 13},
					Body: []Node{
						&ContinueNode{Line: 1, Col: 23}, // Col of "continue" token
					},
					Line: 1,
					Col:  4, // Col of "for" token
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			nodes, err := parser.Parse()
			require.NoError(t, err)

			assert.Equal(t, tt.expected, nodes)
		})
	}
}

// =============================================================================
// Test Group 6: Mixed Content Parsing
// =============================================================================

func TestParserMixedContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Statement
	}{
		{
			name:  "Text and variable",
			input: "Hello {{ name }}!",
			expected: []Statement{
				&TextNode{Text: "Hello ", Line: 1, Col: 1},
				&OutputNode{
					Expression: &VariableNode{Name: "name", Line: 1, Col: 10},
					Line:       1,
					Col:        7,
				},
				&TextNode{Text: "!", Line: 1, Col: 17},
			},
		},
		{
			name:  "Multiple variables with text",
			input: "{{ x }} and {{ y }}",
			expected: []Statement{
				&OutputNode{
					Expression: &VariableNode{Name: "x", Line: 1, Col: 4},
					Line:       1,
					Col:        1,
				},
				&TextNode{Text: " and ", Line: 1, Col: 8},
				&OutputNode{
					Expression: &VariableNode{Name: "y", Line: 1, Col: 16},
					Line:       1,
					Col:        13,
				},
			},
		},
		{
			name:  "If with mixed content",
			input: "Start{% if x %}{{ x }}{% endif %}End",
			expected: []Statement{
				&TextNode{Text: "Start", Line: 1, Col: 1},
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 12},
							Body: []Node{
								&OutputNode{
									Expression: &VariableNode{Name: "x", Line: 1, Col: 19},
									Line:       1,
									Col:        16,
								},
							},
						},
					},
					ElseBody: nil,
					Line:     1,
					Col:      9, // Col of "if" token
				},
				&TextNode{Text: "End", Line: 1, Col: 34},
			},
		},
		{
			name:  "For with mixed content",
			input: "Start{% for i in nums %}{{ i }}{% endfor %}End",
			expected: []Statement{
				&TextNode{Text: "Start", Line: 1, Col: 1},
				&ForNode{
					LoopVars:   []string{"i"},
					Collection: &VariableNode{Name: "nums", Line: 1, Col: 18},
					Body: []Node{
						&OutputNode{
							Expression: &VariableNode{Name: "i", Line: 1, Col: 28},
							Line:       1,
							Col:        25,
						},
					},
					Line: 1,
					Col:  9, // Col of "for" token
				},
				&TextNode{Text: "End", Line: 1, Col: 44},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			nodes, err := parser.Parse()
			require.NoError(t, err)

			assert.Equal(t, tt.expected, nodes)
		})
	}
}

// =============================================================================
// Test Group 7: Nested Structures
// =============================================================================

func TestParserNestedStructures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Statement
	}{
		{
			name:  "Nested if",
			input: "{% if x %}{% if y %}yes{% endif %}{% endif %}",
			expected: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&IfNode{
									Branches: []IfBranch{
										{
											Condition: &VariableNode{Name: "y", Line: 1, Col: 17},
											Body: []Node{
												&TextNode{Text: "yes", Line: 1, Col: 21},
											},
										},
									},
									ElseBody: nil,
									Line:     1,
									Col:      14, // Col of inner "if" token
								},
							},
						},
					},
					ElseBody: nil,
					Line:     1,
					Col:      4, // Col of outer "if" token
				},
			},
		},
		{
			name:  "Nested for",
			input: "{% for i in nums %}{% for j in items %}{{ j }}{% endfor %}{% endfor %}",
			expected: []Statement{
				&ForNode{
					LoopVars:   []string{"i"},
					Collection: &VariableNode{Name: "nums", Line: 1, Col: 13},
					Body: []Node{
						&ForNode{
							LoopVars:   []string{"j"},
							Collection: &VariableNode{Name: "items", Line: 1, Col: 32},
							Body: []Node{
								&OutputNode{
									Expression: &VariableNode{Name: "j", Line: 1, Col: 43},
									Line:       1,
									Col:        40,
								},
							},
							Line: 1,
							Col:  23, // Col of inner "for" token
						},
					},
					Line: 1,
					Col:  4, // Col of outer "for" token
				},
			},
		},
		{
			name:  "For inside if",
			input: "{% if x %}{% for i in nums %}{{ i }}{% endfor %}{% endif %}",
			expected: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&ForNode{
									LoopVars:   []string{"i"},
									Collection: &VariableNode{Name: "nums", Line: 1, Col: 23},
									Body: []Node{
										&OutputNode{
											Expression: &VariableNode{Name: "i", Line: 1, Col: 33},
											Line:       1,
											Col:        30,
										},
									},
									Line: 1,
									Col:  14, // Col of "for" token
								},
							},
						},
					},
					ElseBody: nil,
					Line:     1,
					Col:      4, // Col of "if" token
				},
			},
		},
		{
			name:  "If inside for",
			input: "{% for i in nums %}{% if i %}{{ i }}{% endif %}{% endfor %}",
			expected: []Statement{
				&ForNode{
					LoopVars:   []string{"i"},
					Collection: &VariableNode{Name: "nums", Line: 1, Col: 13},
					Body: []Node{
						&IfNode{
							Branches: []IfBranch{
								{
									Condition: &VariableNode{Name: "i", Line: 1, Col: 26},
									Body: []Node{
										&OutputNode{
											Expression: &VariableNode{Name: "i", Line: 1, Col: 33},
											Line:       1,
											Col:        30,
										},
									},
								},
							},
							ElseBody: nil,
							Line:     1,
							Col:      23, // Col of "if" token
						},
					},
					Line: 1,
					Col:  4, // Col of "for" token
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			nodes, err := parser.Parse()
			require.NoError(t, err)

			assert.Equal(t, tt.expected, nodes)
		})
	}
}

// =============================================================================
// Test Group 8: Error Cases
// =============================================================================

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Note: Unclosed tags are caught by the lexer, not the parser
		{
			name:  "Empty variable tag",
			input: "{{ }}",
		},
		{
			name:  "Unknown tag",
			input: "{% unknown %}",
		},
		{
			name:  "Missing endif",
			input: "{% if x %}yes",
		},
		{
			name:  "Missing endfor",
			input: "{% for i in nums %}{{ i }}",
		},
		{
			name:  "Invalid for syntax - missing in",
			input: "{% for i nums %}{{ i }}{% endfor %}",
		},
		{
			name:  "Invalid for syntax - no variable",
			input: "{% for %}{% endfor %}",
		},
		{
			name:  "Break with arguments",
			input: "{% for i in nums %}{% break x %}{% endfor %}",
		},
		{
			name:  "Continue with arguments",
			input: "{% for i in nums %}{% continue y %}{% endfor %}",
		},
		{
			name:  "Endif with arguments",
			input: "{% if x %}yes{% endif x %}",
		},
		{
			name:  "Endfor with arguments",
			input: "{% for i in nums %}{{ i }}{% endfor i %}",
		},
		{
			name:  "Else with arguments",
			input: "{% if x %}yes{% else x %}no{% endif %}",
		},
		{
			name:  "Extra tokens after if condition",
			input: "{% if x y %}yes{% endif %}",
		},
		{
			name:  "Extra tokens after for collection",
			input: "{% for i in nums extra %}{{ i }}{% endfor %}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			_, err = parser.Parse()
			assert.Error(t, err, "Expected error for input: %s", tt.input)
		})
	}
}

// =============================================================================
// Test Group 9: Parser Helper Methods
// =============================================================================

func TestParserHelperMethods(t *testing.T) {
	t.Run("Current and Advance", func(t *testing.T) {
		tokens := []*Token{
			{Type: TokenText, Value: "hello", Line: 1, Col: 1},
			{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 6},
			{Type: TokenIdentifier, Value: "x", Line: 1, Col: 8},
			{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 9},
			{Type: TokenEOF, Value: "", Line: 1, Col: 11},
		}

		parser := NewParser(tokens)

		// Test Current
		assert.Equal(t, tokens[0], parser.Current())

		// Test Advance
		parser.Advance()
		assert.Equal(t, tokens[1], parser.Current())

		parser.Advance()
		assert.Equal(t, tokens[2], parser.Current())

		// Test Remaining
		assert.Equal(t, 3, parser.Remaining())
	})

	t.Run("Match", func(t *testing.T) {
		tokens := []*Token{
			{Type: TokenIdentifier, Value: "for", Line: 1, Col: 1},
			{Type: TokenIdentifier, Value: "i", Line: 1, Col: 5},
			{Type: TokenEOF, Value: "", Line: 1, Col: 6},
		}

		parser := NewParser(tokens)

		// Match success
		tok := parser.Match(TokenIdentifier, "for")
		assert.NotNil(t, tok)
		assert.Equal(t, "for", tok.Value)

		// Match failure (wrong value)
		tok = parser.Match(TokenIdentifier, "while")
		assert.Nil(t, tok)

		// Current should be "i"
		assert.Equal(t, "i", parser.Current().Value)
	})

	t.Run("ExpectIdentifier", func(t *testing.T) {
		tokens := []*Token{
			{Type: TokenIdentifier, Value: "name", Line: 1, Col: 1},
			{Type: TokenNumber, Value: "42", Line: 1, Col: 6},
			{Type: TokenEOF, Value: "", Line: 1, Col: 8},
		}

		parser := NewParser(tokens)

		// Expect success
		tok, err := parser.ExpectIdentifier()
		assert.NoError(t, err)
		assert.Equal(t, "name", tok.Value)

		// Expect failure (wrong type)
		_, err = parser.ExpectIdentifier()
		assert.Error(t, err)
	})

	t.Run("Error methods", func(t *testing.T) {
		tokens := []*Token{
			{Type: TokenText, Value: "hello", Line: 5, Col: 10},
			{Type: TokenEOF, Value: "", Line: 5, Col: 15},
		}

		parser := NewParser(tokens)

		// Test Error
		err := parser.Error("test error")
		assert.Error(t, err)
		var parseErr *ParseError
		ok := errors.As(err, &parseErr)
		assert.True(t, ok)
		assert.Equal(t, 5, parseErr.Line)
		assert.Equal(t, 10, parseErr.Col)
		assert.Equal(t, "test error", parseErr.Message)

		// Test Errorf
		err = parser.Errorf("error with %s", "format")
		assert.Error(t, err)
		ok = errors.As(err, &parseErr)
		assert.True(t, ok)
		assert.Equal(t, "error with format", parseErr.Message)
	})
}

// =============================================================================
// Test Group 10: Edge Cases
// =============================================================================

func TestParserEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Statement
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: nil, // Empty input returns nil, not []Statement{}
		},
		{
			name:  "Only whitespace",
			input: "   \n\t  ",
			expected: []Statement{
				&TextNode{Text: "   \n\t  ", Line: 1, Col: 1},
			},
		},
		{
			name:  "Adjacent tags",
			input: "{{ x }}{{ y }}",
			expected: []Statement{
				&OutputNode{
					Expression: &VariableNode{Name: "x", Line: 1, Col: 4},
					Line:       1,
					Col:        1,
				},
				&OutputNode{
					Expression: &VariableNode{Name: "y", Line: 1, Col: 11},
					Line:       1,
					Col:        8,
				},
			},
		},
		{
			name:  "Empty if body",
			input: "{% if x %}{% endif %}",
			expected: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body:      nil, // Empty body is nil, not []Node{}
						},
					},
					ElseBody: nil,
					Line:     1,
					Col:      4, // Col of "if" token
				},
			},
		},
		{
			name:  "Empty for body",
			input: "{% for i in nums %}{% endfor %}",
			expected: []Statement{
				&ForNode{
					LoopVars:   []string{"i"},
					Collection: &VariableNode{Name: "nums", Line: 1, Col: 13},
					Body:       nil, // Empty body is nil, not []Node{}
					Line:       1,
					Col:        4, // Col of "for" token
				},
			},
		},
		{
			name:  "Empty else body",
			input: "{% if x %}yes{% else %}{% endif %}",
			expected: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&TextNode{Text: "yes", Line: 1, Col: 11},
							},
						},
					},
					ElseBody: nil, // Empty else body is nil, not []Node{}
					Line:     1,
					Col:      4, // Col of "if" token
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			nodes, err := parser.Parse()
			require.NoError(t, err)

			assert.Equal(t, tt.expected, nodes)
		})
	}
}

// =============================================================================
// Test Group 11: Complex Real-World Templates
// =============================================================================

func TestParserComplexTemplates(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, nodes []Statement)
	}{
		{
			name: "Complete template with mixed content",
			input: `<html>
<body>
  <h1>{{ title }}</h1>
  {% if user %}
    <p>Welcome, {{ user.name }}!</p>
  {% else %}
    <p>Please log in.</p>
  {% endif %}
</body>
</html>`,
			check: func(t *testing.T, nodes []Statement) {
				// Structure: text, output, text, if, text
				// Actually it's: text (including newlines and spaces), output, text, if, text
				assert.Equal(t, 5, len(nodes))
			},
		},
		{
			name: "List rendering with for loop",
			input: `<ul>
{% for item in items %}
  <li>{{ item.name }}: {{ item.price }}</li>
{% endfor %}
</ul>`,
			check: func(t *testing.T, nodes []Statement) {
				// Structure: text, for, text (newlines create separate text nodes)
				assert.Equal(t, 3, len(nodes))

				// Check for loop
				forNode, ok := nodes[1].(*ForNode)
				assert.True(t, ok)
				assert.Equal(t, []string{"item"}, forNode.LoopVars)
				// Body has: text (newline+spaces), text ("<li>"), output, text ("</li>"), text (newline)
				assert.Equal(t, 5, len(forNode.Body))
			},
		},
		{
			name: "Nested loops and conditionals",
			input: `{% for category in categories %}
  <h2>{{ category.name }}</h2>
  {% for product in category.products %}
    {% if product.inStock %}
      <p>{{ product.name }} - ${{ product.price }}</p>
    {% endif %}
  {% endfor %}
{% endfor %}`,
			check: func(t *testing.T, nodes []Statement) {
				assert.Equal(t, 1, len(nodes))

				// Outer for loop
				outerFor, ok := nodes[0].(*ForNode)
				assert.True(t, ok)
				assert.Equal(t, []string{"category"}, outerFor.LoopVars)

				// Should have text, output, text, for, text
				assert.Equal(t, 5, len(outerFor.Body))

				// Inner for loop
				innerFor, ok := outerFor.Body[3].(*ForNode)
				assert.True(t, ok)
				assert.Equal(t, []string{"product"}, innerFor.LoopVars)

				// Inner for should have text, if, text
				assert.Equal(t, 3, len(innerFor.Body))

				// If inside inner for
				ifNode, ok := innerFor.Body[1].(*IfNode)
				assert.True(t, ok)
				assert.Equal(t, 1, len(ifNode.Branches))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			nodes, err := parser.Parse()
			require.NoError(t, err)

			tt.check(t, nodes)
		})
	}
}

// =============================================================================
// Test Group 12: ParseExpression Method
// =============================================================================

func TestParserParseExpression(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []*Token
		expected Expression
	}{
		{
			name: "Simple variable",
			tokens: []*Token{
				{Type: TokenIdentifier, Value: "x", Line: 1, Col: 1},
				{Type: TokenEOF, Value: "", Line: 1, Col: 2},
			},
			expected: &VariableNode{Name: "x", Line: 1, Col: 1},
		},
		{
			name: "Binary expression",
			tokens: []*Token{
				{Type: TokenIdentifier, Value: "a", Line: 1, Col: 1},
				{Type: TokenSymbol, Value: "+", Line: 1, Col: 3},
				{Type: TokenIdentifier, Value: "b", Line: 1, Col: 5},
				{Type: TokenEOF, Value: "", Line: 1, Col: 6},
			},
			expected: &BinaryOpNode{
				Operator: "+",
				Left:     &VariableNode{Name: "a", Line: 1, Col: 1},
				Right:    &VariableNode{Name: "b", Line: 1, Col: 5},
				Line:     1,
				Col:      3,
			},
		},
		{
			name: "Expression with filter",
			tokens: []*Token{
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 1},
				{Type: TokenSymbol, Value: "|", Line: 1, Col: 5},
				{Type: TokenIdentifier, Value: "upper", Line: 1, Col: 6},
				{Type: TokenEOF, Value: "", Line: 1, Col: 11},
			},
			expected: &FilterNode{
				Expression: &VariableNode{Name: "name", Line: 1, Col: 1},
				FilterName: "upper",
				Args:       nil, // Args is nil, not empty slice
				Line:       1,
				Col:        5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.tokens)
			expr, err := parser.ParseExpression()
			require.NoError(t, err)

			// Use reflect.DeepEqual for complex structure comparison
			if !reflect.DeepEqual(tt.expected, expr) {
				t.Errorf("Expression mismatch.\nExpected: %#v\nGot: %#v", tt.expected, expr)
			}
		})
	}
}
