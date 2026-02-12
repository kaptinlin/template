package template

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// eof is a convenience helper that returns an EOF token at the given position.
func eof(line, col int) *Token {
	return &Token{Type: TokenEOF, Value: "", Line: line, Col: col}
}

// tok is a convenience helper that returns a token with the given fields.
func tok(typ TokenType, value string, line, col int) *Token {
	return &Token{Type: typ, Value: value, Line: line, Col: col}
}

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
			want:  []*Token{eof(1, 1)},
		},
		{
			name:  "plain text only",
			input: "Hello World",
			want: []*Token{
				tok(TokenText, "Hello World", 1, 1),
				eof(1, 12),
			},
		},
		{
			name:  "plain text with newline",
			input: "Hello\nWorld",
			want: []*Token{
				tok(TokenText, "Hello\nWorld", 1, 1),
				eof(2, 6),
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
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "name", 1, 4),
				tok(TokenVarEnd, "}}", 1, 9),
				eof(1, 11),
			},
		},
		{
			name:  "variable with property access",
			input: "{{ user.name }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "user", 1, 4),
				tok(TokenSymbol, ".", 1, 8),
				tok(TokenIdentifier, "name", 1, 9),
				tok(TokenVarEnd, "}}", 1, 14),
				eof(1, 16),
			},
		},
		{
			name:  "variable with filter",
			input: "{{ name | upper }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "name", 1, 4),
				tok(TokenSymbol, "|", 1, 9),
				tok(TokenIdentifier, "upper", 1, 11),
				tok(TokenVarEnd, "}}", 1, 17),
				eof(1, 19),
			},
		},
		{
			name:  "variable with filter and arguments",
			input: "{{ name | truncate:10 }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "name", 1, 4),
				tok(TokenSymbol, "|", 1, 9),
				tok(TokenIdentifier, "truncate", 1, 11),
				tok(TokenSymbol, ":", 1, 19),
				tok(TokenNumber, "10", 1, 20),
				tok(TokenVarEnd, "}}", 1, 23),
				eof(1, 25),
			},
		},
		{
			name:  "variable surrounded by text",
			input: "Hello {{ name }}, welcome!",
			want: []*Token{
				tok(TokenText, "Hello ", 1, 1),
				tok(TokenVarBegin, "{{", 1, 7),
				tok(TokenIdentifier, "name", 1, 10),
				tok(TokenVarEnd, "}}", 1, 15),
				tok(TokenText, ", welcome!", 1, 17),
				eof(1, 27),
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
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "if", 1, 4),
				tok(TokenIdentifier, "x", 1, 7),
				tok(TokenTagEnd, "%}", 1, 9),
				eof(1, 11),
			},
		},
		{
			name:  "if tag with comparison",
			input: "{% if x > 5 %}",
			want: []*Token{
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "if", 1, 4),
				tok(TokenIdentifier, "x", 1, 7),
				tok(TokenSymbol, ">", 1, 9),
				tok(TokenNumber, "5", 1, 11),
				tok(TokenTagEnd, "%}", 1, 13),
				eof(1, 15),
			},
		},
		{
			name:  "for loop tag",
			input: "{% for item in items %}",
			want: []*Token{
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "for", 1, 4),
				tok(TokenIdentifier, "item", 1, 8),
				tok(TokenIdentifier, "in", 1, 13),
				tok(TokenIdentifier, "items", 1, 16),
				tok(TokenTagEnd, "%}", 1, 22),
				eof(1, 24),
			},
		},
		{
			name:  "endif tag",
			input: "{% endif %}",
			want: []*Token{
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "endif", 1, 4),
				tok(TokenTagEnd, "%}", 1, 10),
				eof(1, 12),
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
			want:  []*Token{eof(1, 24)},
		},
		{
			name:  "comment with text around",
			input: "Hello {# comment #} World",
			want: []*Token{
				tok(TokenText, "Hello ", 1, 1),
				tok(TokenText, " World", 1, 20),
				eof(1, 26),
			},
		},
		{
			name:  "multiple comments",
			input: "{# comment1 #}{# comment2 #}",
			want:  []*Token{eof(1, 29)},
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
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenString, "hello", 1, 4),
				tok(TokenVarEnd, "}}", 1, 12),
				eof(1, 14),
			},
		},
		{
			name:  "single quoted string",
			input: `{{ 'hello' }}`,
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenString, "hello", 1, 4),
				tok(TokenVarEnd, "}}", 1, 12),
				eof(1, 14),
			},
		},
		{
			name:  "string with escaped quotes",
			input: `{{ "hello \"world\"" }}`,
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenString, `hello "world"`, 1, 4),
				tok(TokenVarEnd, "}}", 1, 22),
				eof(1, 24),
			},
		},
		{
			name:  "string with escaped backslash",
			input: `{{ "path\\to\\file" }}`,
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenString, `path\to\file`, 1, 4),
				tok(TokenVarEnd, "}}", 1, 21),
				eof(1, 23),
			},
		},
		{
			name:  "string with newline escape",
			input: `{{ "line1\nline2" }}`,
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenString, "line1\nline2", 1, 4),
				tok(TokenVarEnd, "}}", 1, 19),
				eof(1, 21),
			},
		},
		{
			name:  "string with tab escape",
			input: `{{ "col1\tcol2" }}`,
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenString, "col1\tcol2", 1, 4),
				tok(TokenVarEnd, "}}", 1, 17),
				eof(1, 19),
			},
		},
		{
			name:  "empty string",
			input: `{{ "" }}`,
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenString, "", 1, 4),
				tok(TokenVarEnd, "}}", 1, 7),
				eof(1, 9),
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
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenNumber, "42", 1, 4),
				tok(TokenVarEnd, "}}", 1, 7),
				eof(1, 9),
			},
		},
		{
			name:  "float",
			input: "{{ 3.14 }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenNumber, "3.14", 1, 4),
				tok(TokenVarEnd, "}}", 1, 9),
				eof(1, 11),
			},
		},
		{
			name:  "zero",
			input: "{{ 0 }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenNumber, "0", 1, 4),
				tok(TokenVarEnd, "}}", 1, 6),
				eof(1, 8),
			},
		},
		{
			name:  "decimal starting with zero",
			input: "{{ 0.5 }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenNumber, "0.5", 1, 4),
				tok(TokenVarEnd, "}}", 1, 8),
				eof(1, 10),
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
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "if", 1, 4),
				tok(TokenIdentifier, "a", 1, 7),
				tok(TokenSymbol, "==", 1, 9),
				tok(TokenIdentifier, "b", 1, 12),
				tok(TokenTagEnd, "%}", 1, 14),
				eof(1, 16),
			},
		},
		{
			name:  "not equal operator",
			input: "{% if a != b %}",
			want: []*Token{
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "if", 1, 4),
				tok(TokenIdentifier, "a", 1, 7),
				tok(TokenSymbol, "!=", 1, 9),
				tok(TokenIdentifier, "b", 1, 12),
				tok(TokenTagEnd, "%}", 1, 14),
				eof(1, 16),
			},
		},
		{
			name:  "less than and greater than",
			input: "{% if a < b and c > d %}",
			want: []*Token{
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "if", 1, 4),
				tok(TokenIdentifier, "a", 1, 7),
				tok(TokenSymbol, "<", 1, 9),
				tok(TokenIdentifier, "b", 1, 11),
				tok(TokenIdentifier, "and", 1, 13),
				tok(TokenIdentifier, "c", 1, 17),
				tok(TokenSymbol, ">", 1, 19),
				tok(TokenIdentifier, "d", 1, 21),
				tok(TokenTagEnd, "%}", 1, 23),
				eof(1, 25),
			},
		},
		{
			name:  "arithmetic operators",
			input: "{{ a + b - c * d / e }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "a", 1, 4),
				tok(TokenSymbol, "+", 1, 6),
				tok(TokenIdentifier, "b", 1, 8),
				tok(TokenSymbol, "-", 1, 10),
				tok(TokenIdentifier, "c", 1, 12),
				tok(TokenSymbol, "*", 1, 14),
				tok(TokenIdentifier, "d", 1, 16),
				tok(TokenSymbol, "/", 1, 18),
				tok(TokenIdentifier, "e", 1, 20),
				tok(TokenVarEnd, "}}", 1, 22),
				eof(1, 24),
			},
		},
		{
			name:  "subscript and property access",
			input: "{{ user.items[0] }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "user", 1, 4),
				tok(TokenSymbol, ".", 1, 8),
				tok(TokenIdentifier, "items", 1, 9),
				tok(TokenSymbol, "[", 1, 14),
				tok(TokenNumber, "0", 1, 15),
				tok(TokenSymbol, "]", 1, 16),
				tok(TokenVarEnd, "}}", 1, 18),
				eof(1, 20),
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
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "if", 1, 4),
				tok(TokenIdentifier, "user", 1, 7),
				tok(TokenTagEnd, "%}", 1, 12),
				tok(TokenText, "Hello ", 1, 14),
				tok(TokenVarBegin, "{{", 1, 20),
				tok(TokenIdentifier, "user", 1, 23),
				tok(TokenSymbol, ".", 1, 27),
				tok(TokenIdentifier, "name", 1, 28),
				tok(TokenVarEnd, "}}", 1, 33),
				tok(TokenTagBegin, "{%", 1, 35),
				tok(TokenIdentifier, "else", 1, 38),
				tok(TokenTagEnd, "%}", 1, 43),
				tok(TokenText, "Guest", 1, 45),
				tok(TokenTagBegin, "{%", 1, 50),
				tok(TokenIdentifier, "endif", 1, 53),
				tok(TokenTagEnd, "%}", 1, 59),
				eof(1, 61),
			},
		},
		{
			name:  "for loop template",
			input: "{% for item in items %}{{ item }}{% endfor %}",
			want: []*Token{
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "for", 1, 4),
				tok(TokenIdentifier, "item", 1, 8),
				tok(TokenIdentifier, "in", 1, 13),
				tok(TokenIdentifier, "items", 1, 16),
				tok(TokenTagEnd, "%}", 1, 22),
				tok(TokenVarBegin, "{{", 1, 24),
				tok(TokenIdentifier, "item", 1, 27),
				tok(TokenVarEnd, "}}", 1, 32),
				tok(TokenTagBegin, "{%", 1, 34),
				tok(TokenIdentifier, "endfor", 1, 37),
				tok(TokenTagEnd, "%}", 1, 44),
				eof(1, 46),
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
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "name", 1, 6),
				tok(TokenVarEnd, "}}", 1, 13),
				eof(1, 15),
			},
		},
		{
			name:  "tabs and spaces",
			input: "{{\t\tname\t\t}}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "name", 1, 5),
				tok(TokenVarEnd, "}}", 1, 11),
				eof(1, 13),
			},
		},
		{
			name:  "newlines in block tag",
			input: "{%\nif\nx\n%}",
			want: []*Token{
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "if", 2, 1),
				tok(TokenIdentifier, "x", 3, 1),
				tok(TokenTagEnd, "%}", 4, 1),
				eof(4, 3),
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
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "true", 1, 4),
				tok(TokenVarEnd, "}}", 1, 9),
				eof(1, 11),
			},
		},
		{
			name:  "boolean false",
			input: "{{ false }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "false", 1, 4),
				tok(TokenVarEnd, "}}", 1, 10),
				eof(1, 12),
			},
		},
		{
			name:  "logical operators",
			input: "{% if a and b or not c %}",
			want: []*Token{
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "if", 1, 4),
				tok(TokenIdentifier, "a", 1, 7),
				tok(TokenIdentifier, "and", 1, 9),
				tok(TokenIdentifier, "b", 1, 13),
				tok(TokenIdentifier, "or", 1, 15),
				tok(TokenIdentifier, "not", 1, 18),
				tok(TokenIdentifier, "c", 1, 22),
				tok(TokenTagEnd, "%}", 1, 24),
				eof(1, 26),
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
		tok(TokenText, "line1\n", 1, 1),
		tok(TokenTagBegin, "{%", 2, 1),
		tok(TokenIdentifier, "if", 2, 4),
		tok(TokenIdentifier, "x", 2, 7),
		tok(TokenTagEnd, "%}", 2, 9),
		tok(TokenText, "\nHello\n", 2, 11),
		tok(TokenTagBegin, "{%", 4, 1),
		tok(TokenIdentifier, "endif", 4, 4),
		tok(TokenTagEnd, "%}", 4, 10),
		tok(TokenText, "\nline5", 4, 12),
		eof(5, 6),
	}
	compareTokens(t, "multiline template", got, want)
}

func TestLexerScanBlockTagError(t *testing.T) {
	// Cover the error return from scanInsideTag inside scanBlockTag.
	_, err := NewLexer("{% @ %}").Tokenize()
	if err == nil {
		t.Fatal("expected error for invalid character in block tag")
	}
	var lexErr *LexerError
	if !errors.As(err, &lexErr) {
		t.Fatalf("error type = %T, want *LexerError", err)
	}
	if !strings.Contains(lexErr.Message, "unexpected character") {
		t.Errorf("error message = %q, want 'unexpected character'", lexErr.Message)
	}
}

func TestLexerScanStringEscapes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*Token
	}{
		{
			name:  "carriage return escape",
			input: `{{ "a\rb" }}`,
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenString, "a\rb", 1, 4),
				tok(TokenVarEnd, "}}", 1, 11),
				eof(1, 13),
			},
		},
		{
			name:  "escaped single quote in double quoted string",
			input: `{{ "it\'s" }}`,
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenString, "it's", 1, 4),
				tok(TokenVarEnd, "}}", 1, 12),
				eof(1, 14),
			},
		},
		{
			name:  "escaped single quote in single quoted string",
			input: `{{ 'it\'s' }}`,
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenString, "it's", 1, 4),
				tok(TokenVarEnd, "}}", 1, 12),
				eof(1, 14),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, tt.name, got, tt.want)
		})
	}
}

func TestLexerScanInsideTagEOF(t *testing.T) {
	// Cover the pos >= len early return in scanInsideTag.
	// After "x" is scanned, skipWhitespace advances past trailing spaces
	// to EOF. scanInsideTag is called but pos >= len, so it returns nil.
	// Then the outer loop detects EOF and returns the "unclosed" error.
	_, err := NewLexer("{{ x   ").Tokenize()
	if err == nil {
		t.Fatal("expected error for unclosed variable tag with trailing whitespace")
	}
	var lexErr *LexerError
	if !errors.As(err, &lexErr) {
		t.Fatalf("error type = %T, want *LexerError", err)
	}
	if !strings.Contains(lexErr.Message, "unclosed variable tag") {
		t.Errorf("error message = %q, want 'unclosed variable tag'", lexErr.Message)
	}
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
				tok(TokenText, "{ not a tag }", 1, 1),
				eof(1, 14),
			},
		},
		{
			name:  "brace at end of input",
			input: "text{",
			want: []*Token{
				tok(TokenText, "text{", 1, 1),
				eof(1, 6),
			},
		},
		{
			name:  "identifier with underscore",
			input: "{{ _private }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "_private", 1, 4),
				tok(TokenVarEnd, "}}", 1, 13),
				eof(1, 15),
			},
		},
		{
			name:  "identifier with digits",
			input: "{{ item2 }}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "item2", 1, 4),
				tok(TokenVarEnd, "}}", 1, 10),
				eof(1, 12),
			},
		},
		{
			name:  "no whitespace in variable tag",
			input: "{{x}}",
			want: []*Token{
				tok(TokenVarBegin, "{{", 1, 1),
				tok(TokenIdentifier, "x", 1, 3),
				tok(TokenVarEnd, "}}", 1, 4),
				eof(1, 6),
			},
		},
		{
			name:  "set tag with assignment",
			input: "{% set x = 1 %}",
			want: []*Token{
				tok(TokenTagBegin, "{%", 1, 1),
				tok(TokenIdentifier, "set", 1, 4),
				tok(TokenIdentifier, "x", 1, 8),
				tok(TokenSymbol, "=", 1, 10),
				tok(TokenNumber, "1", 1, 12),
				tok(TokenTagEnd, "%}", 1, 14),
				eof(1, 16),
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
