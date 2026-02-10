package template

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLexerBasicTokenization tests basic tokenization functionality
func TestLexerBasicTokenization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "Empty input",
			input: "",
			expected: []*Token{
				{Type: TokenEOF, Value: "", Line: 1, Col: 1},
			},
		},
		{
			name:  "Plain text only",
			input: "Hello World",
			expected: []*Token{
				{Type: TokenText, Value: "Hello World", Line: 1, Col: 1},
				{Type: TokenEOF, Value: "", Line: 1, Col: 12},
			},
		},
		{
			name:  "Plain text with newline",
			input: "Hello\nWorld",
			expected: []*Token{
				{Type: TokenText, Value: "Hello\nWorld", Line: 1, Col: 1},
				{Type: TokenEOF, Value: "", Line: 2, Col: 6},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}

// TestLexerVariableTags tests variable tag tokenization {{ }}
func TestLexerVariableTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "Simple variable",
			input: "{{ name }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 9},
				{Type: TokenEOF, Value: "", Line: 1, Col: 11},
			},
		},
		{
			name:  "Variable with property access",
			input: "{{ user.name }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "user", Line: 1, Col: 4},
				{Type: TokenSymbol, Value: ".", Line: 1, Col: 8},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 9},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 14},
				{Type: TokenEOF, Value: "", Line: 1, Col: 16},
			},
		},
		{
			name:  "Variable with filter",
			input: "{{ name | upper }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 4},
				{Type: TokenSymbol, Value: "|", Line: 1, Col: 9},
				{Type: TokenIdentifier, Value: "upper", Line: 1, Col: 11},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 17},
				{Type: TokenEOF, Value: "", Line: 1, Col: 19},
			},
		},
		{
			name:  "Variable with filter and arguments",
			input: "{{ name | truncate:10 }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 4},
				{Type: TokenSymbol, Value: "|", Line: 1, Col: 9},
				{Type: TokenIdentifier, Value: "truncate", Line: 1, Col: 11},
				{Type: TokenSymbol, Value: ":", Line: 1, Col: 19},
				{Type: TokenNumber, Value: "10", Line: 1, Col: 20},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 23},
				{Type: TokenEOF, Value: "", Line: 1, Col: 25},
			},
		},
		{
			name:  "Variable surrounded by text",
			input: "Hello {{ name }}, welcome!",
			expected: []*Token{
				{Type: TokenText, Value: "Hello ", Line: 1, Col: 1},
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 7},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 10},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 15},
				{Type: TokenText, Value: ", welcome!", Line: 1, Col: 17},
				{Type: TokenEOF, Value: "", Line: 1, Col: 27},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}

// TestLexerBlockTags tests block tag tokenization {% %}
func TestLexerBlockTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "Simple if tag",
			input: "{% if x %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "if", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "x", Line: 1, Col: 7},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 9},
				{Type: TokenEOF, Value: "", Line: 1, Col: 11},
			},
		},
		{
			name:  "If tag with comparison",
			input: "{% if x > 5 %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "if", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "x", Line: 1, Col: 7},
				{Type: TokenSymbol, Value: ">", Line: 1, Col: 9},
				{Type: TokenNumber, Value: "5", Line: 1, Col: 11},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 13},
				{Type: TokenEOF, Value: "", Line: 1, Col: 15},
			},
		},
		{
			name:  "For loop tag",
			input: "{% for item in items %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "for", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "item", Line: 1, Col: 8},
				{Type: TokenIdentifier, Value: "in", Line: 1, Col: 13},
				{Type: TokenIdentifier, Value: "items", Line: 1, Col: 16},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 22},
				{Type: TokenEOF, Value: "", Line: 1, Col: 24},
			},
		},
		{
			name:  "Endif tag",
			input: "{% endif %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "endif", Line: 1, Col: 4},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 10},
				{Type: TokenEOF, Value: "", Line: 1, Col: 12},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}

// TestLexerComments tests comment handling
func TestLexerComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "Simple comment",
			input: "{# This is a comment #}",
			expected: []*Token{
				{Type: TokenEOF, Value: "", Line: 1, Col: 24},
			},
		},
		{
			name:  "Comment with text around",
			input: "Hello {# comment #} World",
			expected: []*Token{
				{Type: TokenText, Value: "Hello ", Line: 1, Col: 1},
				{Type: TokenText, Value: " World", Line: 1, Col: 20},
				{Type: TokenEOF, Value: "", Line: 1, Col: 26},
			},
		},
		{
			name:  "Multiple comments",
			input: "{# comment1 #}{# comment2 #}",
			expected: []*Token{
				{Type: TokenEOF, Value: "", Line: 1, Col: 29},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}

// TestLexerStrings tests string literal tokenization
func TestLexerStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "Double quoted string",
			input: `{{ "hello" }}`,
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: "hello", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 12},
				{Type: TokenEOF, Value: "", Line: 1, Col: 14},
			},
		},
		{
			name:  "Single quoted string",
			input: `{{ 'hello' }}`,
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: "hello", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 12},
				{Type: TokenEOF, Value: "", Line: 1, Col: 14},
			},
		},
		{
			name:  "String with escaped quotes",
			input: `{{ "hello \"world\"" }}`,
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: `hello "world"`, Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 22},
				{Type: TokenEOF, Value: "", Line: 1, Col: 24},
			},
		},
		{
			name:  "String with escaped backslash",
			input: `{{ "path\\to\\file" }}`,
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: `path\to\file`, Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 21},
				{Type: TokenEOF, Value: "", Line: 1, Col: 23},
			},
		},
		{
			name:  "String with newline escape",
			input: `{{ "line1\nline2" }}`,
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: "line1\nline2", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 19},
				{Type: TokenEOF, Value: "", Line: 1, Col: 21},
			},
		},
		{
			name:  "String with tab escape",
			input: `{{ "col1\tcol2" }}`,
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: "col1\tcol2", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 17},
				{Type: TokenEOF, Value: "", Line: 1, Col: 19},
			},
		},
		{
			name:  "Empty string",
			input: `{{ "" }}`,
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: "", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 7},
				{Type: TokenEOF, Value: "", Line: 1, Col: 9},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}

// TestLexerNumbers tests number literal tokenization
func TestLexerNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "Integer",
			input: "{{ 42 }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenNumber, Value: "42", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 7},
				{Type: TokenEOF, Value: "", Line: 1, Col: 9},
			},
		},
		{
			name:  "Float",
			input: "{{ 3.14 }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenNumber, Value: "3.14", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 9},
				{Type: TokenEOF, Value: "", Line: 1, Col: 11},
			},
		},
		{
			name:  "Zero",
			input: "{{ 0 }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenNumber, Value: "0", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 6},
				{Type: TokenEOF, Value: "", Line: 1, Col: 8},
			},
		},
		{
			name:  "Decimal starting with zero",
			input: "{{ 0.5 }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenNumber, Value: "0.5", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 8},
				{Type: TokenEOF, Value: "", Line: 1, Col: 10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}

// TestLexerOperators tests operator tokenization
func TestLexerOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "Comparison operators",
			input: "{% if a == b %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "if", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "a", Line: 1, Col: 7},
				{Type: TokenSymbol, Value: "==", Line: 1, Col: 9},
				{Type: TokenIdentifier, Value: "b", Line: 1, Col: 12},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 14},
				{Type: TokenEOF, Value: "", Line: 1, Col: 16},
			},
		},
		{
			name:  "Not equal operator",
			input: "{% if a != b %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "if", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "a", Line: 1, Col: 7},
				{Type: TokenSymbol, Value: "!=", Line: 1, Col: 9},
				{Type: TokenIdentifier, Value: "b", Line: 1, Col: 12},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 14},
				{Type: TokenEOF, Value: "", Line: 1, Col: 16},
			},
		},
		{
			name:  "Less than and greater than",
			input: "{% if a < b and c > d %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "if", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "a", Line: 1, Col: 7},
				{Type: TokenSymbol, Value: "<", Line: 1, Col: 9},
				{Type: TokenIdentifier, Value: "b", Line: 1, Col: 11},
				{Type: TokenIdentifier, Value: "and", Line: 1, Col: 13},
				{Type: TokenIdentifier, Value: "c", Line: 1, Col: 17},
				{Type: TokenSymbol, Value: ">", Line: 1, Col: 19},
				{Type: TokenIdentifier, Value: "d", Line: 1, Col: 21},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 23},
				{Type: TokenEOF, Value: "", Line: 1, Col: 25},
			},
		},
		{
			name:  "Arithmetic operators",
			input: "{{ a + b - c * d / e }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "a", Line: 1, Col: 4},
				{Type: TokenSymbol, Value: "+", Line: 1, Col: 6},
				{Type: TokenIdentifier, Value: "b", Line: 1, Col: 8},
				{Type: TokenSymbol, Value: "-", Line: 1, Col: 10},
				{Type: TokenIdentifier, Value: "c", Line: 1, Col: 12},
				{Type: TokenSymbol, Value: "*", Line: 1, Col: 14},
				{Type: TokenIdentifier, Value: "d", Line: 1, Col: 16},
				{Type: TokenSymbol, Value: "/", Line: 1, Col: 18},
				{Type: TokenIdentifier, Value: "e", Line: 1, Col: 20},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 22},
				{Type: TokenEOF, Value: "", Line: 1, Col: 24},
			},
		},
		{
			name:  "Subscript and property access",
			input: "{{ user.items[0] }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "user", Line: 1, Col: 4},
				{Type: TokenSymbol, Value: ".", Line: 1, Col: 8},
				{Type: TokenIdentifier, Value: "items", Line: 1, Col: 9},
				{Type: TokenSymbol, Value: "[", Line: 1, Col: 14},
				{Type: TokenNumber, Value: "0", Line: 1, Col: 15},
				{Type: TokenSymbol, Value: "]", Line: 1, Col: 16},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 18},
				{Type: TokenEOF, Value: "", Line: 1, Col: 20},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}

// TestLexerComplexTemplates tests complex real-world templates
func TestLexerComplexTemplates(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "If-else template",
			input: "{% if user %}Hello {{ user.name }}{% else %}Guest{% endif %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "if", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "user", Line: 1, Col: 7},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 12},
				{Type: TokenText, Value: "Hello ", Line: 1, Col: 14},
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 20},
				{Type: TokenIdentifier, Value: "user", Line: 1, Col: 23},
				{Type: TokenSymbol, Value: ".", Line: 1, Col: 27},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 28},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 33},
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 35},
				{Type: TokenIdentifier, Value: "else", Line: 1, Col: 38},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 43},
				{Type: TokenText, Value: "Guest", Line: 1, Col: 45},
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 50},
				{Type: TokenIdentifier, Value: "endif", Line: 1, Col: 53},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 59},
				{Type: TokenEOF, Value: "", Line: 1, Col: 61},
			},
		},
		{
			name:  "For loop template",
			input: "{% for item in items %}{{ item }}{% endfor %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "for", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "item", Line: 1, Col: 8},
				{Type: TokenIdentifier, Value: "in", Line: 1, Col: 13},
				{Type: TokenIdentifier, Value: "items", Line: 1, Col: 16},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 22},
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 24},
				{Type: TokenIdentifier, Value: "item", Line: 1, Col: 27},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 32},
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 34},
				{Type: TokenIdentifier, Value: "endfor", Line: 1, Col: 37},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 44},
				{Type: TokenEOF, Value: "", Line: 1, Col: 46},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			// Use reflect.DeepEqual for complex comparison
			if !reflect.DeepEqual(tt.expected, tokens) {
				t.Errorf("Token mismatch\nExpected: %+v\nGot:      %+v", tt.expected, tokens)
			}
		})
	}
}

// TestLexerErrors tests error conditions
func TestLexerErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Unclosed variable tag",
			input:       "{{ name",
			expectError: true,
			errorMsg:    "unclosed variable tag",
		},
		{
			name:        "Unclosed block tag",
			input:       "{% if x",
			expectError: true,
			errorMsg:    "unclosed block tag",
		},
		{
			name:        "Unclosed comment",
			input:       "{# comment",
			expectError: true,
			errorMsg:    "unclosed comment",
		},
		{
			name:        "Unclosed string",
			input:       `{{ "hello`,
			expectError: true,
			errorMsg:    "unclosed string",
		},
		{
			name:        "Newline in comment",
			input:       "{# line1\nline2 #}",
			expectError: true,
			errorMsg:    "newline not permitted in comment",
		},
		{
			name:        "Newline in string",
			input:       "{{ \"line1\nline2\" }}",
			expectError: true,
			errorMsg:    "newline in string is not allowed",
		},
		{
			name:        "Invalid escape sequence",
			input:       `{{ "hello\x" }}`,
			expectError: true,
			errorMsg:    "unknown escape sequence",
		},
		{
			name:        "Unexpected character",
			input:       "{{ @ }}",
			expectError: true,
			errorMsg:    "unexpected character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, tokens)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokens)
			}
		})
	}
}

// TestLexerWhitespace tests whitespace handling
func TestLexerWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "Extra whitespace in variable tag",
			input: "{{   name   }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 6},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 13},
				{Type: TokenEOF, Value: "", Line: 1, Col: 15},
			},
		},
		{
			name:  "Tabs and spaces",
			input: "{{\t\tname\t\t}}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 5},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 11},
				{Type: TokenEOF, Value: "", Line: 1, Col: 13},
			},
		},
		{
			name:  "Newlines in block tag",
			input: "{%\nif\nx\n%}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "if", Line: 2, Col: 1},
				{Type: TokenIdentifier, Value: "x", Line: 3, Col: 1},
				{Type: TokenTagEnd, Value: "%}", Line: 4, Col: 1},
				{Type: TokenEOF, Value: "", Line: 4, Col: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}

// TestLexerKeywords tests keyword tokenization
func TestLexerKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "Boolean true",
			input: "{{ true }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "true", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 9},
				{Type: TokenEOF, Value: "", Line: 1, Col: 11},
			},
		},
		{
			name:  "Boolean false",
			input: "{{ false }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "false", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 10},
				{Type: TokenEOF, Value: "", Line: 1, Col: 12},
			},
		},
		{
			name:  "Logical operators",
			input: "{% if a and b or not c %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "if", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "a", Line: 1, Col: 7},
				{Type: TokenIdentifier, Value: "and", Line: 1, Col: 9},
				{Type: TokenIdentifier, Value: "b", Line: 1, Col: 13},
				{Type: TokenIdentifier, Value: "or", Line: 1, Col: 15},
				{Type: TokenIdentifier, Value: "not", Line: 1, Col: 18},
				{Type: TokenIdentifier, Value: "c", Line: 1, Col: 22},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 24},
				{Type: TokenEOF, Value: "", Line: 1, Col: 26},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}

// TestLexerEdgeCases tests edge cases and boundary conditions
func TestLexerEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Token
	}{
		{
			name:  "Empty variable tag",
			input: "{{ }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 4},
				{Type: TokenEOF, Value: "", Line: 1, Col: 6},
			},
		},
		{
			name:  "Empty block tag",
			input: "{% %}",
			expected: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 4},
				{Type: TokenEOF, Value: "", Line: 1, Col: 6},
			},
		},
		{
			name:  "Adjacent tags",
			input: "{{a}}{{b}}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "a", Line: 1, Col: 3},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 4},
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 6},
				{Type: TokenIdentifier, Value: "b", Line: 1, Col: 8},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 9},
				{Type: TokenEOF, Value: "", Line: 1, Col: 11},
			},
		},
		{
			name:  "Identifier with underscore",
			input: "{{ _private_var }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "_private_var", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 17},
				{Type: TokenEOF, Value: "", Line: 1, Col: 19},
			},
		},
		{
			name:  "Identifier with numbers",
			input: "{{ var123 }}",
			expected: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "var123", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 11},
				{Type: TokenEOF, Value: "", Line: 1, Col: 13},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}
