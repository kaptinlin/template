package template

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// eof is a convenience helper that returns an EOF token at the given position.
func eof(line, col int) *token {
	return &token{Type: tokenEOF, value: "", Line: line, Col: col}
}

// tok is a convenience helper that returns a token with the given fields.
func tok(typ tokenType, value string, line, col int) *token {
	return &token{Type: typ, value: value, Line: line, Col: col}
}

// compareTokens is a test helper that compares token slices using cmp.Diff.
func compareTokens(t *testing.T, label string, got, want []*token) {
	t.Helper()
	if diff := cmp.Diff(want, got, cmpopts.EquateComparable(token{})); diff != "" {
		t.Errorf("%s mismatch (-want +got):\n%s", label, diff)
	}
}

func TestLexerBasicTokenization(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*token
	}{
		{
			name:  "empty input",
			input: "",
			want:  []*token{eof(1, 1)},
		},
		{
			name:  "plain text only",
			input: "Hello World",
			want: []*token{
				tok(tokenText, "Hello World", 1, 1),
				eof(1, 12),
			},
		},
		{
			name:  "plain text with newline",
			input: "Hello\nWorld",
			want: []*token{
				tok(tokenText, "Hello\nWorld", 1, 1),
				eof(2, 6),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
		want  []*token
	}{
		{
			name:  "simple variable",
			input: "{{ name }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "name", 1, 4),
				tok(tokenVarEnd, "}}", 1, 9),
				eof(1, 11),
			},
		},
		{
			name:  "variable with property access",
			input: "{{ user.name }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "user", 1, 4),
				tok(tokenSymbol, ".", 1, 8),
				tok(tokenIdentifier, "name", 1, 9),
				tok(tokenVarEnd, "}}", 1, 14),
				eof(1, 16),
			},
		},
		{
			name:  "variable with filter",
			input: "{{ name | upper }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "name", 1, 4),
				tok(tokenSymbol, "|", 1, 9),
				tok(tokenIdentifier, "upper", 1, 11),
				tok(tokenVarEnd, "}}", 1, 17),
				eof(1, 19),
			},
		},
		{
			name:  "variable with filter and arguments",
			input: "{{ name | truncate:10 }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "name", 1, 4),
				tok(tokenSymbol, "|", 1, 9),
				tok(tokenIdentifier, "truncate", 1, 11),
				tok(tokenSymbol, ":", 1, 19),
				tok(tokenNumber, "10", 1, 20),
				tok(tokenVarEnd, "}}", 1, 23),
				eof(1, 25),
			},
		},
		{
			name:  "variable surrounded by text",
			input: "Hello {{ name }}, welcome!",
			want: []*token{
				tok(tokenText, "Hello ", 1, 1),
				tok(tokenVarBegin, "{{", 1, 7),
				tok(tokenIdentifier, "name", 1, 10),
				tok(tokenVarEnd, "}}", 1, 15),
				tok(tokenText, ", welcome!", 1, 17),
				eof(1, 27),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
		want  []*token
	}{
		{
			name:  "simple if tag",
			input: "{% if x %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "if", 1, 4),
				tok(tokenIdentifier, "x", 1, 7),
				tok(tokenTagEnd, "%}", 1, 9),
				eof(1, 11),
			},
		},
		{
			name:  "if tag with comparison",
			input: "{% if x > 5 %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "if", 1, 4),
				tok(tokenIdentifier, "x", 1, 7),
				tok(tokenSymbol, ">", 1, 9),
				tok(tokenNumber, "5", 1, 11),
				tok(tokenTagEnd, "%}", 1, 13),
				eof(1, 15),
			},
		},
		{
			name:  "for loop tag",
			input: "{% for item in items %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "for", 1, 4),
				tok(tokenIdentifier, "item", 1, 8),
				tok(tokenIdentifier, "in", 1, 13),
				tok(tokenIdentifier, "items", 1, 16),
				tok(tokenTagEnd, "%}", 1, 22),
				eof(1, 24),
			},
		},
		{
			name:  "endif tag",
			input: "{% endif %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "endif", 1, 4),
				tok(tokenTagEnd, "%}", 1, 10),
				eof(1, 12),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
		want  []*token
	}{
		{
			name:  "simple comment",
			input: "{# This is a comment #}",
			want:  []*token{eof(1, 24)},
		},
		{
			name:  "comment with text around",
			input: "Hello {# comment #} World",
			want: []*token{
				tok(tokenText, "Hello ", 1, 1),
				tok(tokenText, " World", 1, 20),
				eof(1, 26),
			},
		},
		{
			name:  "multiple comments",
			input: "{# comment1 #}{# comment2 #}",
			want:  []*token{eof(1, 29)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
		want  []*token
	}{
		{
			name:  "double quoted string",
			input: `{{ "hello" }}`,
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenString, "hello", 1, 4),
				tok(tokenVarEnd, "}}", 1, 12),
				eof(1, 14),
			},
		},
		{
			name:  "single quoted string",
			input: `{{ 'hello' }}`,
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenString, "hello", 1, 4),
				tok(tokenVarEnd, "}}", 1, 12),
				eof(1, 14),
			},
		},
		{
			name:  "string with escaped quotes",
			input: `{{ "hello \"world\"" }}`,
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenString, `hello "world"`, 1, 4),
				tok(tokenVarEnd, "}}", 1, 22),
				eof(1, 24),
			},
		},
		{
			name:  "string with escaped backslash",
			input: `{{ "path\\to\\file" }}`,
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenString, `path\to\file`, 1, 4),
				tok(tokenVarEnd, "}}", 1, 21),
				eof(1, 23),
			},
		},
		{
			name:  "string with newline escape",
			input: `{{ "line1\nline2" }}`,
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenString, "line1\nline2", 1, 4),
				tok(tokenVarEnd, "}}", 1, 19),
				eof(1, 21),
			},
		},
		{
			name:  "string with tab escape",
			input: `{{ "col1\tcol2" }}`,
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenString, "col1\tcol2", 1, 4),
				tok(tokenVarEnd, "}}", 1, 17),
				eof(1, 19),
			},
		},
		{
			name:  "empty string",
			input: `{{ "" }}`,
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenString, "", 1, 4),
				tok(tokenVarEnd, "}}", 1, 7),
				eof(1, 9),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
		want  []*token
	}{
		{
			name:  "integer",
			input: "{{ 42 }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenNumber, "42", 1, 4),
				tok(tokenVarEnd, "}}", 1, 7),
				eof(1, 9),
			},
		},
		{
			name:  "float",
			input: "{{ 3.14 }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenNumber, "3.14", 1, 4),
				tok(tokenVarEnd, "}}", 1, 9),
				eof(1, 11),
			},
		},
		{
			name:  "zero",
			input: "{{ 0 }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenNumber, "0", 1, 4),
				tok(tokenVarEnd, "}}", 1, 6),
				eof(1, 8),
			},
		},
		{
			name:  "decimal starting with zero",
			input: "{{ 0.5 }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenNumber, "0.5", 1, 4),
				tok(tokenVarEnd, "}}", 1, 8),
				eof(1, 10),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
		want  []*token
	}{
		{
			name:  "comparison operators",
			input: "{% if a == b %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "if", 1, 4),
				tok(tokenIdentifier, "a", 1, 7),
				tok(tokenSymbol, "==", 1, 9),
				tok(tokenIdentifier, "b", 1, 12),
				tok(tokenTagEnd, "%}", 1, 14),
				eof(1, 16),
			},
		},
		{
			name:  "not equal operator",
			input: "{% if a != b %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "if", 1, 4),
				tok(tokenIdentifier, "a", 1, 7),
				tok(tokenSymbol, "!=", 1, 9),
				tok(tokenIdentifier, "b", 1, 12),
				tok(tokenTagEnd, "%}", 1, 14),
				eof(1, 16),
			},
		},
		{
			name:  "less than and greater than",
			input: "{% if a < b and c > d %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "if", 1, 4),
				tok(tokenIdentifier, "a", 1, 7),
				tok(tokenSymbol, "<", 1, 9),
				tok(tokenIdentifier, "b", 1, 11),
				tok(tokenIdentifier, "and", 1, 13),
				tok(tokenIdentifier, "c", 1, 17),
				tok(tokenSymbol, ">", 1, 19),
				tok(tokenIdentifier, "d", 1, 21),
				tok(tokenTagEnd, "%}", 1, 23),
				eof(1, 25),
			},
		},
		{
			name:  "arithmetic operators",
			input: "{{ a + b - c * d / e }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "a", 1, 4),
				tok(tokenSymbol, "+", 1, 6),
				tok(tokenIdentifier, "b", 1, 8),
				tok(tokenSymbol, "-", 1, 10),
				tok(tokenIdentifier, "c", 1, 12),
				tok(tokenSymbol, "*", 1, 14),
				tok(tokenIdentifier, "d", 1, 16),
				tok(tokenSymbol, "/", 1, 18),
				tok(tokenIdentifier, "e", 1, 20),
				tok(tokenVarEnd, "}}", 1, 22),
				eof(1, 24),
			},
		},
		{
			name:  "subscript and property access",
			input: "{{ user.items[0] }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "user", 1, 4),
				tok(tokenSymbol, ".", 1, 8),
				tok(tokenIdentifier, "items", 1, 9),
				tok(tokenSymbol, "[", 1, 14),
				tok(tokenNumber, "0", 1, 15),
				tok(tokenSymbol, "]", 1, 16),
				tok(tokenVarEnd, "}}", 1, 18),
				eof(1, 20),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
		want  []*token
	}{
		{
			name:  "if-else template",
			input: "{% if user %}Hello {{ user.name }}{% else %}Guest{% endif %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "if", 1, 4),
				tok(tokenIdentifier, "user", 1, 7),
				tok(tokenTagEnd, "%}", 1, 12),
				tok(tokenText, "Hello ", 1, 14),
				tok(tokenVarBegin, "{{", 1, 20),
				tok(tokenIdentifier, "user", 1, 23),
				tok(tokenSymbol, ".", 1, 27),
				tok(tokenIdentifier, "name", 1, 28),
				tok(tokenVarEnd, "}}", 1, 33),
				tok(tokenTagBegin, "{%", 1, 35),
				tok(tokenIdentifier, "else", 1, 38),
				tok(tokenTagEnd, "%}", 1, 43),
				tok(tokenText, "Guest", 1, 45),
				tok(tokenTagBegin, "{%", 1, 50),
				tok(tokenIdentifier, "endif", 1, 53),
				tok(tokenTagEnd, "%}", 1, 59),
				eof(1, 61),
			},
		},
		{
			name:  "for loop template",
			input: "{% for item in items %}{{ item }}{% endfor %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "for", 1, 4),
				tok(tokenIdentifier, "item", 1, 8),
				tok(tokenIdentifier, "in", 1, 13),
				tok(tokenIdentifier, "items", 1, 16),
				tok(tokenTagEnd, "%}", 1, 22),
				tok(tokenVarBegin, "{{", 1, 24),
				tok(tokenIdentifier, "item", 1, 27),
				tok(tokenVarEnd, "}}", 1, 32),
				tok(tokenTagBegin, "{%", 1, 34),
				tok(tokenIdentifier, "endfor", 1, 37),
				tok(tokenTagEnd, "%}", 1, 44),
				eof(1, 46),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
			got, err := newLexer(tt.input).Tokenize()
			if err == nil {
				t.Fatalf("Tokenize(%q) = %v, want error containing %q", tt.input, got, tt.wantMsg)
			}
			if got != nil {
				t.Errorf("Tokenize(%q) returned non-nil tokens on error", tt.input)
			}

			var lexErr *lexerError
			if !errors.As(err, &lexErr) {
				t.Fatalf("Tokenize(%q) error type = %T, want *lexerError", tt.input, err)
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
		want  []*token
	}{
		{
			name:  "extra whitespace in variable tag",
			input: "{{   name   }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "name", 1, 6),
				tok(tokenVarEnd, "}}", 1, 13),
				eof(1, 15),
			},
		},
		{
			name:  "tabs and spaces",
			input: "{{\t\tname\t\t}}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "name", 1, 5),
				tok(tokenVarEnd, "}}", 1, 11),
				eof(1, 13),
			},
		},
		{
			name:  "newlines in block tag",
			input: "{%\nif\nx\n%}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "if", 2, 1),
				tok(tokenIdentifier, "x", 3, 1),
				tok(tokenTagEnd, "%}", 4, 1),
				eof(4, 3),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
		want  []*token
	}{
		{
			name:  "boolean true",
			input: "{{ true }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "true", 1, 4),
				tok(tokenVarEnd, "}}", 1, 9),
				eof(1, 11),
			},
		},
		{
			name:  "boolean false",
			input: "{{ false }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "false", 1, 4),
				tok(tokenVarEnd, "}}", 1, 10),
				eof(1, 12),
			},
		},
		{
			name:  "logical operators",
			input: "{% if a and b or not c %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "if", 1, 4),
				tok(tokenIdentifier, "a", 1, 7),
				tok(tokenIdentifier, "and", 1, 9),
				tok(tokenIdentifier, "b", 1, 13),
				tok(tokenIdentifier, "or", 1, 15),
				tok(tokenIdentifier, "not", 1, 18),
				tok(tokenIdentifier, "c", 1, 22),
				tok(tokenTagEnd, "%}", 1, 24),
				eof(1, 26),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
			_, err := newLexer(tt.input).Tokenize()
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
	got, err := newLexer(input).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize returned unexpected error: %v", err)
	}

	want := []*token{
		tok(tokenText, "line1\n", 1, 1),
		tok(tokenTagBegin, "{%", 2, 1),
		tok(tokenIdentifier, "if", 2, 4),
		tok(tokenIdentifier, "x", 2, 7),
		tok(tokenTagEnd, "%}", 2, 9),
		tok(tokenText, "\nHello\n", 2, 11),
		tok(tokenTagBegin, "{%", 4, 1),
		tok(tokenIdentifier, "endif", 4, 4),
		tok(tokenTagEnd, "%}", 4, 10),
		tok(tokenText, "\nline5", 4, 12),
		eof(5, 6),
	}
	compareTokens(t, "multiline template", got, want)
}

func TestLexerScanBlockTagError(t *testing.T) {
	// Cover the error return from scanInsideTag inside scanBlockTag.
	_, err := newLexer("{% @ %}").Tokenize()
	if err == nil {
		t.Fatal("expected error for invalid character in block tag")
	}
	var lexErr *lexerError
	if !errors.As(err, &lexErr) {
		t.Fatalf("error type = %T, want *lexerError", err)
	}
	if !strings.Contains(lexErr.Message, "unexpected character") {
		t.Errorf("error message = %q, want 'unexpected character'", lexErr.Message)
	}
}

func TestLexerScanStringEscapes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*token
	}{
		{
			name:  "carriage return escape",
			input: `{{ "a\rb" }}`,
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenString, "a\rb", 1, 4),
				tok(tokenVarEnd, "}}", 1, 11),
				eof(1, 13),
			},
		},
		{
			name:  "escaped single quote in double quoted string",
			input: `{{ "it\'s" }}`,
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenString, "it's", 1, 4),
				tok(tokenVarEnd, "}}", 1, 12),
				eof(1, 14),
			},
		},
		{
			name:  "escaped single quote in single quoted string",
			input: `{{ 'it\'s' }}`,
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenString, "it's", 1, 4),
				tok(tokenVarEnd, "}}", 1, 12),
				eof(1, 14),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
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
	_, err := newLexer("{{ x   ").Tokenize()
	if err == nil {
		t.Fatal("expected error for unclosed variable tag with trailing whitespace")
	}
	var lexErr *lexerError
	if !errors.As(err, &lexErr) {
		t.Fatalf("error type = %T, want *lexerError", err)
	}
	if !strings.Contains(lexErr.Message, "unclosed variable tag") {
		t.Errorf("error message = %q, want 'unclosed variable tag'", lexErr.Message)
	}
}

func TestLexerEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*token
	}{
		{
			name:  "lone open brace is text",
			input: "{ not a tag }",
			want: []*token{
				tok(tokenText, "{ not a tag }", 1, 1),
				eof(1, 14),
			},
		},
		{
			name:  "brace at end of input",
			input: "text{",
			want: []*token{
				tok(tokenText, "text{", 1, 1),
				eof(1, 6),
			},
		},
		{
			name:  "identifier with underscore",
			input: "{{ _private }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "_private", 1, 4),
				tok(tokenVarEnd, "}}", 1, 13),
				eof(1, 15),
			},
		},
		{
			name:  "identifier with digits",
			input: "{{ item2 }}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "item2", 1, 4),
				tok(tokenVarEnd, "}}", 1, 10),
				eof(1, 12),
			},
		},
		{
			name:  "no whitespace in variable tag",
			input: "{{x}}",
			want: []*token{
				tok(tokenVarBegin, "{{", 1, 1),
				tok(tokenIdentifier, "x", 1, 3),
				tok(tokenVarEnd, "}}", 1, 4),
				eof(1, 6),
			},
		},
		{
			name:  "set tag with assignment",
			input: "{% set x = 1 %}",
			want: []*token{
				tok(tokenTagBegin, "{%", 1, 1),
				tok(tokenIdentifier, "set", 1, 4),
				tok(tokenIdentifier, "x", 1, 8),
				tok(tokenSymbol, "=", 1, 10),
				tok(tokenNumber, "1", 1, 12),
				tok(tokenTagEnd, "%}", 1, 14),
				eof(1, 16),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLexer(tt.input).Tokenize()
			if err != nil {
				t.Fatalf("Tokenize(%q) returned unexpected error: %v", tt.input, err)
			}
			compareTokens(t, "Tokenize("+tt.input+")", got, tt.want)
		})
	}
}
