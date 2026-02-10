package template

import (
	"strconv"
	"testing"
)

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		name string
		typ  TokenType
		want string
	}{
		{name: "TokenError", typ: TokenError, want: "ERROR"},
		{name: "TokenEOF", typ: TokenEOF, want: "EOF"},
		{name: "TokenText", typ: TokenText, want: "TEXT"},
		{name: "TokenVarBegin", typ: TokenVarBegin, want: "VAR_BEGIN"},
		{name: "TokenVarEnd", typ: TokenVarEnd, want: "VAR_END"},
		{name: "TokenTagBegin", typ: TokenTagBegin, want: "TAG_BEGIN"},
		{name: "TokenTagEnd", typ: TokenTagEnd, want: "TAG_END"},
		{name: "TokenIdentifier", typ: TokenIdentifier, want: "IDENTIFIER"},
		{name: "TokenString", typ: TokenString, want: "STRING"},
		{name: "TokenNumber", typ: TokenNumber, want: "NUMBER"},
		{name: "TokenSymbol", typ: TokenSymbol, want: "SYMBOL"},
		{name: "unknown type 999", typ: TokenType(999), want: "UNKNOWN(" + strconv.Itoa(999) + ")"},
		{name: "unknown type -1", typ: TokenType(-1), want: "UNKNOWN(" + strconv.Itoa(-1) + ")"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.typ.String()
			if got != tt.want {
				t.Errorf("TokenType(%d).String() = %q, want %q", tt.typ, got, tt.want)
			}
		})
	}
}

func TestTokenString(t *testing.T) {
	tests := []struct {
		name  string
		token Token
		want  string
	}{
		{
			name:  "EOF token",
			token: Token{Type: TokenEOF, Line: 1, Col: 1},
			want:  "EOF at line 1, col 1",
		},
		{
			name:  "EOF at different position",
			token: Token{Type: TokenEOF, Line: 10, Col: 25},
			want:  "EOF at line 10, col 25",
		},
		{
			name:  "short identifier",
			token: Token{Type: TokenIdentifier, Value: "name", Line: 1, Col: 5},
			want:  `IDENTIFIER("name") at line 1, col 5`,
		},
		{
			name:  "empty value",
			token: Token{Type: TokenText, Value: "", Line: 1, Col: 1},
			want:  `TEXT("") at line 1, col 1`,
		},
		{
			name:  "exactly 20 characters not truncated",
			token: Token{Type: TokenText, Value: "12345678901234567890", Line: 2, Col: 1},
			want:  `TEXT("12345678901234567890") at line 2, col 1`,
		},
		{
			name:  "21 characters truncated",
			token: Token{Type: TokenText, Value: "123456789012345678901", Line: 3, Col: 1},
			want:  `TEXT("12345678901234567890"...) at line 3, col 1`,
		},
		{
			name:  "long text truncated",
			token: Token{Type: TokenText, Value: "This is a very long text that exceeds twenty characters", Line: 1, Col: 1},
			want:  `TEXT("This is a very long "...) at line 1, col 1`,
		},
		{
			name:  "string literal token",
			token: Token{Type: TokenString, Value: "hello", Line: 1, Col: 10},
			want:  `STRING("hello") at line 1, col 10`,
		},
		{
			name:  "number token",
			token: Token{Type: TokenNumber, Value: "42", Line: 5, Col: 3},
			want:  `NUMBER("42") at line 5, col 3`,
		},
		{
			name:  "symbol token",
			token: Token{Type: TokenSymbol, Value: "==", Line: 1, Col: 8},
			want:  `SYMBOL("==") at line 1, col 8`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.String()
			if got != tt.want {
				t.Errorf("Token.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	tests := []struct {
		name  string
		ident string
		want  bool
	}{
		// All valid keywords
		{name: "in", ident: "in", want: true},
		{name: "and", ident: "and", want: true},
		{name: "or", ident: "or", want: true},
		{name: "not", ident: "not", want: true},
		{name: "true", ident: "true", want: true},
		{name: "false", ident: "false", want: true},
		{name: "if", ident: "if", want: true},
		{name: "elif", ident: "elif", want: true},
		{name: "else", ident: "else", want: true},
		{name: "endif", ident: "endif", want: true},
		{name: "for", ident: "for", want: true},
		{name: "endfor", ident: "endfor", want: true},
		{name: "break", ident: "break", want: true},
		{name: "continue", ident: "continue", want: true},
		// Non-keywords
		{name: "regular identifier name", ident: "name", want: false},
		{name: "regular identifier render", ident: "render", want: false},
		{name: "empty string", ident: "", want: false},
		{name: "uppercase IF is not keyword", ident: "IF", want: false},
		{name: "uppercase TRUE is not keyword", ident: "TRUE", want: false},
		{name: "mixed case Not is not keyword", ident: "Not", want: false},
		{name: "partial keyword end", ident: "end", want: false},
		{name: "partial keyword el", ident: "el", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsKeyword(tt.ident)
			if got != tt.want {
				t.Errorf("IsKeyword(%q) = %v, want %v", tt.ident, got, tt.want)
			}
		})
	}
}

func TestIsSymbol(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// Comparison operators
		{name: "equal", input: "==", want: true},
		{name: "not equal", input: "!=", want: true},
		{name: "less than", input: "<", want: true},
		{name: "greater than", input: ">", want: true},
		{name: "less or equal", input: "<=", want: true},
		{name: "greater or equal", input: ">=", want: true},
		// Arithmetic operators
		{name: "plus", input: "+", want: true},
		{name: "minus", input: "-", want: true},
		{name: "multiply", input: "*", want: true},
		{name: "divide", input: "/", want: true},
		{name: "modulo", input: "%", want: true},
		// Logical operators
		{name: "logical and", input: "&&", want: true},
		{name: "logical or", input: "||", want: true},
		{name: "logical not", input: "!", want: true},
		// Other symbols
		{name: "pipe", input: "|", want: true},
		{name: "colon", input: ":", want: true},
		{name: "comma", input: ",", want: true},
		{name: "dot", input: ".", want: true},
		{name: "left paren", input: "(", want: true},
		{name: "right paren", input: ")", want: true},
		{name: "left bracket", input: "[", want: true},
		{name: "right bracket", input: "]", want: true},
		{name: "assignment", input: "=", want: true},
		// Non-symbols
		{name: "word is not symbol", input: "abc", want: false},
		{name: "empty string is not symbol", input: "", want: false},
		{name: "triple equals is not symbol", input: "===", want: false},
		{name: "power operator is not symbol", input: "**", want: false},
		{name: "arrow is not symbol", input: "->", want: false},
		{name: "left brace is not symbol", input: "{", want: false},
		{name: "right brace is not symbol", input: "}", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSymbol(tt.input)
			if got != tt.want {
				t.Errorf("IsSymbol(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
