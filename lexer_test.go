package template

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// compareTokens is a test helper that compares token slices using cmp.Diff.
func compareTokens(t *testing.T, label string, got, want []*Token) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("%s mismatch (-want +got):\n%s", label, diff)
	}
}

func TestLexerBasicTokenization(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "empty input",
			input: "",
			want: []*Token{
				{Type: TokenEOF, Value: "", Line: 1, Col: 1},
			},
		},
		{
			name:  "plain text only",
			input: "Hello World",
			want: []*Token{
				{Type: TokenText, Value: "Hello World", Line: 1, Col: 1},
				{Type: TokenEOF, Value: "", Line: 1, Col: 12},
			},
		},
		{
			name:  "plain text with newline",
			input: "Hello\nWorld",
			want: []*Token{
				{Type: TokenText, Value: "Hello\nWorld", Line: 1, Col: 1},
				{Type: TokenEOF, Value: "", Line: 2, Col: 6},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize("+tt.input+")", got, tt.want)
		})
	}
}

func TestLexerVariableTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "simple variable",
			input: "{{ name }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 9},
				{Type: TokenEOF, Value: "", Line: 1, Col: 11},
			},
		},
		{
			name:  "variable with property access",
			input: "{{ user.name }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "user", Line: 1, Col: 4},
				{Type: TokenSymbol, Value: ".", Line: 1, Col: 8},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 9},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 14},
				{Type: TokenEOF, Value: "", Line: 1, Col: 16},
			},
		},
		{
			name:  "variable with filter",
			input: "{{ name | upper }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 4},
				{Type: TokenSymbol, Value: "|", Line: 1, Col: 9},
				{Type: TokenIdentifier, Value: "upper", Line: 1, Col: 11},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 17},
				{Type: TokenEOF, Value: "", Line: 1, Col: 19},
			},
		},
		{
			name:  "variable with filter and arguments",
			input: "{{ name | truncate:10 }}",
			want: []*Token{
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
			name:  "variable surrounded by text",
			input: "Hello {{ name }}, welcome!",
			want: []*Token{
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
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize("+tt.input+")", got, tt.want)
		})
	}
}

func TestLexerBlockTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "simple if tag",
			input: "{% if x %}",
			want: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "if", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "x", Line: 1, Col: 7},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 9},
				{Type: TokenEOF, Value: "", Line: 1, Col: 11},
			},
		},
		{
			name:  "if tag with comparison",
			input: "{% if x > 5 %}",
			want: []*Token{
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
			name:  "for loop tag",
			input: "{% for item in items %}",
			want: []*Token{
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
			name:  "endif tag",
			input: "{% endif %}",
			want: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "endif", Line: 1, Col: 4},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 10},
				{Type: TokenEOF, Value: "", Line: 1, Col: 12},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize("+tt.input+")", got, tt.want)
		})
	}
}

func TestLexerComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "simple comment",
			input: "{# This is a comment #}",
			want: []*Token{
				{Type: TokenEOF, Value: "", Line: 1, Col: 24},
			},
		},
		{
			name:  "comment with text around",
			input: "Hello {# comment #} World",
			want: []*Token{
				{Type: TokenText, Value: "Hello ", Line: 1, Col: 1},
				{Type: TokenText, Value: " World", Line: 1, Col: 20},
				{Type: TokenEOF, Value: "", Line: 1, Col: 26},
			},
		},
		{
			name:  "multiple comments",
			input: "{# comment1 #}{# comment2 #}",
			want: []*Token{
				{Type: TokenEOF, Value: "", Line: 1, Col: 29},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize("+tt.input+")", got, tt.want)
		})
	}
}

func TestLexerStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "double quoted string",
			input: `{{ "hello" }}`,
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: "hello", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 12},
				{Type: TokenEOF, Value: "", Line: 1, Col: 14},
			},
		},
		{
			name:  "single quoted string",
			input: `{{ 'hello' }}`,
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: "hello", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 12},
				{Type: TokenEOF, Value: "", Line: 1, Col: 14},
			},
		},
		{
			name:  "string with escaped quotes",
			input: `{{ "hello \"world\"" }}`,
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: `hello "world"`, Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 22},
				{Type: TokenEOF, Value: "", Line: 1, Col: 24},
			},
		},
		{
			name:  "string with escaped backslash",
			input: `{{ "path\\to\\file" }}`,
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: `path\to\file`, Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 21},
				{Type: TokenEOF, Value: "", Line: 1, Col: 23},
			},
		},
		{
			name:  "string with newline escape",
			input: `{{ "line1\nline2" }}`,
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: "line1\nline2", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 19},
				{Type: TokenEOF, Value: "", Line: 1, Col: 21},
			},
		},
		{
			name:  "string with tab escape",
			input: `{{ "col1\tcol2" }}`,
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: "col1\tcol2", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 17},
				{Type: TokenEOF, Value: "", Line: 1, Col: 19},
			},
		},
		{
			name:  "empty string",
			input: `{{ "" }}`,
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenString, Value: "", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 7},
				{Type: TokenEOF, Value: "", Line: 1, Col: 9},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize("+tt.input+")", got, tt.want)
		})
	}
}

func TestLexerNumbers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "integer",
			input: "{{ 42 }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenNumber, Value: "42", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 7},
				{Type: TokenEOF, Value: "", Line: 1, Col: 9},
			},
		},
		{
			name:  "float",
			input: "{{ 3.14 }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenNumber, Value: "3.14", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 9},
				{Type: TokenEOF, Value: "", Line: 1, Col: 11},
			},
		},
		{
			name:  "zero",
			input: "{{ 0 }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenNumber, Value: "0", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 6},
				{Type: TokenEOF, Value: "", Line: 1, Col: 8},
			},
		},
		{
			name:  "decimal starting with zero",
			input: "{{ 0.5 }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenNumber, Value: "0.5", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 8},
				{Type: TokenEOF, Value: "", Line: 1, Col: 10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize("+tt.input+")", got, tt.want)
		})
	}
}

func TestLexerOperators(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "comparison operators",
			input: "{% if a == b %}",
			want: []*Token{
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
			name:  "not equal operator",
			input: "{% if a != b %}",
			want: []*Token{
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
			name:  "less than and greater than",
			input: "{% if a < b and c > d %}",
			want: []*Token{
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
			name:  "arithmetic operators",
			input: "{{ a + b - c * d / e }}",
			want: []*Token{
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
			name:  "subscript and property access",
			input: "{{ user.items[0] }}",
			want: []*Token{
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
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize("+tt.input+")", got, tt.want)
		})
	}
}

func TestLexerComplexTemplates(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "if-else template",
			input: "{% if user %}Hello {{ user.name }}{% else %}Guest{% endif %}",
			want: []*Token{
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
			name:  "for loop template",
			input: "{% for item in items %}{{ item }}{% endfor %}",
			want: []*Token{
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
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize(complex)", got, tt.want)
		})
	}
}

func TestLexerErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantMsg string
	}{
		{
			name:    "unclosed variable tag",
			input:   "{{ name",
			wantMsg: "unclosed variable tag",
		},
		{
			name:    "unclosed block tag",
			input:   "{% if x",
			wantMsg: "unclosed block tag",
		},
		{
			name:    "unclosed comment",
			input:   "{# comment",
			wantMsg: "unclosed comment",
		},
		{
			name:    "unclosed string",
			input:   `{{ "hello`,
			wantMsg: "unclosed string",
		},
		{
			name:    "newline in comment",
			input:   "{# line1\nline2 #}",
			wantMsg: "newline not permitted in comment",
		},
		{
			name:    "newline in string",
			input:   "{{ \"line1\nline2\" }}",
			wantMsg: "newline in string is not allowed",
		},
		{
			name:    "invalid escape sequence",
			input:   `{{ "hello\x" }}`,
			wantMsg: "unknown escape sequence",
		},
		{
			name:    "unexpected character",
			input:   "{{ @ }}",
			wantMsg: "unexpected character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLexer(tt.input).Tokenize()
			if err == nil {
				t.Fatalf("Tokenize(%q) = %v, want error containing %q", tt.input, got, tt.wantMsg)
			}
			if got != nil {
				t.Errorf("Tokenize(%q) returned non-nil tokens on error", tt.input)
			}

			var lexErr *LexerError
			if !errors.As(err, &lexErr) {
				t.Fatalf("Tokenize(%q) error type = %T, want *LexerError", tt.input, err)
			}
			if !strings.Contains(lexErr.Message, tt.wantMsg) {
				t.Errorf("Tokenize(%q) error message = %q, want substring %q", tt.input, lexErr.Message, tt.wantMsg)
			}
			if lexErr.Line <= 0 {
				t.Errorf("Tokenize(%q) error line = %d, want positive", tt.input, lexErr.Line)
			}
			if lexErr.Col <= 0 {
				t.Errorf("Tokenize(%q) error col = %d, want positive", tt.input, lexErr.Col)
			}
		})
	}
}

func TestLexerWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "extra whitespace in variable tag",
			input: "{{   name   }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 6},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 13},
				{Type: TokenEOF, Value: "", Line: 1, Col: 15},
			},
		},
		{
			name:  "tabs and spaces",
			input: "{{\t\tname\t\t}}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "name", Line: 1, Col: 5},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 11},
				{Type: TokenEOF, Value: "", Line: 1, Col: 13},
			},
		},
		{
			name:  "newlines in block tag",
			input: "{%\nif\nx\n%}",
			want: []*Token{
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
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize", got, tt.want)
		})
	}
}

func TestLexerKeywords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "boolean true",
			input: "{{ true }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "true", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 9},
				{Type: TokenEOF, Value: "", Line: 1, Col: 11},
			},
		},
		{
			name:  "boolean false",
			input: "{{ false }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "false", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 10},
				{Type: TokenEOF, Value: "", Line: 1, Col: 12},
			},
		},
		{
			name:  "logical operators",
			input: "{% if a and b or not c %}",
			want: []*Token{
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
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize("+tt.input+")", got, tt.want)
		})
	}
}

func TestIsOneCharSymbol(t *testing.T) {
	valid := []byte{'+', '-', '*', '/', '%', '<', '>', '!', '=', '|', ':', ',', '.', '(', ')', '[', ']'}
	for _, ch := range valid {
		if !isOneCharSymbol(ch) {
			t.Errorf("isOneCharSymbol(%q) = false, want true", ch)
		}
	}

	invalid := []byte{'{', '}', '@', '#', '$', '^', '~', 'a', '0', ' '}
	for _, ch := range invalid {
		if isOneCharSymbol(ch) {
			t.Errorf("isOneCharSymbol(%q) = true, want false", ch)
		}
	}
}

func TestIsTwoCharSymbol(t *testing.T) {
	tests := []struct {
		a, b byte
		want bool
	}{
		{'=', '=', true},
		{'!', '=', true},
		{'<', '=', true},
		{'>', '=', true},
		{'&', '&', true},
		{'|', '|', true},
		{'+', '+', false},
		{'-', '>', false},
		{'*', '*', false},
		{'=', '!', false},
		{'&', '|', false},
	}

	for _, tt := range tests {
		got := isTwoCharSymbol(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("isTwoCharSymbol(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestLexerErrorFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "error includes line and col",
			input:   "{{ @ }}",
			wantErr: "lexer error at line 1, col 4: unexpected character: @",
		},
		{
			name:    "unclosed string error position",
			input:   `{{ "hello`,
			wantErr: "lexer error at line 1, col 4: unclosed string, expected \"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLexer(tt.input).Tokenize()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Errorf("error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestLexerMultilineTemplate(t *testing.T) {
	input := "line1\n{% if x %}\nHello\n{% endif %}\nline5"
	got, err := NewLexer(input).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize returned unexpected error: %v", err)
	}

	want := []*Token{
		{Type: TokenText, Value: "line1\n", Line: 1, Col: 1},
		{Type: TokenTagBegin, Value: "{%", Line: 2, Col: 1},
		{Type: TokenIdentifier, Value: "if", Line: 2, Col: 4},
		{Type: TokenIdentifier, Value: "x", Line: 2, Col: 7},
		{Type: TokenTagEnd, Value: "%}", Line: 2, Col: 9},
		{Type: TokenText, Value: "\nHello\n", Line: 2, Col: 11},
		{Type: TokenTagBegin, Value: "{%", Line: 4, Col: 1},
		{Type: TokenIdentifier, Value: "endif", Line: 4, Col: 4},
		{Type: TokenTagEnd, Value: "%}", Line: 4, Col: 10},
		{Type: TokenText, Value: "\nline5", Line: 4, Col: 12},
		{Type: TokenEOF, Value: "", Line: 5, Col: 6},
	}
	compareTokens(t, "multiline template", got, want)
}

func TestLexerEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "lone open brace is text",
			input: "{ not a tag }",
			want: []*Token{
				{Type: TokenText, Value: "{ not a tag }", Line: 1, Col: 1},
				{Type: TokenEOF, Value: "", Line: 1, Col: 14},
			},
		},
		{
			name:  "brace at end of input",
			input: "text{",
			want: []*Token{
				{Type: TokenText, Value: "text{", Line: 1, Col: 1},
				{Type: TokenEOF, Value: "", Line: 1, Col: 6},
			},
		},
		{
			name:  "identifier with underscore",
			input: "{{ _private }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "_private", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 13},
				{Type: TokenEOF, Value: "", Line: 1, Col: 15},
			},
		},
		{
			name:  "identifier with digits",
			input: "{{ item2 }}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "item2", Line: 1, Col: 4},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 10},
				{Type: TokenEOF, Value: "", Line: 1, Col: 12},
			},
		},
		{
			name:  "no whitespace in variable tag",
			input: "{{x}}",
			want: []*Token{
				{Type: TokenVarBegin, Value: "{{", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "x", Line: 1, Col: 3},
				{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 4},
				{Type: TokenEOF, Value: "", Line: 1, Col: 6},
			},
		},
		{
			name:  "set tag with assignment",
			input: "{% set x = 1 %}",
			want: []*Token{
				{Type: TokenTagBegin, Value: "{%", Line: 1, Col: 1},
				{Type: TokenIdentifier, Value: "set", Line: 1, Col: 4},
				{Type: TokenIdentifier, Value: "x", Line: 1, Col: 8},
				{Type: TokenSymbol, Value: "=", Line: 1, Col: 10},
				{Type: TokenNumber, Value: "1", Line: 1, Col: 12},
				{Type: TokenTagEnd, Value: "%}", Line: 1, Col: 14},
				{Type: TokenEOF, Value: "", Line: 1, Col: 16},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize("+tt.input+")", got, tt.want)
		})
	}
}
