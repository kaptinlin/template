package template

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		name      string
		tokenType TokenType
		expected  string
	}{
		{name: "TokenError", tokenType: TokenError, expected: "ERROR"},
		{name: "TokenEOF", tokenType: TokenEOF, expected: "EOF"},
		{name: "TokenText", tokenType: TokenText, expected: "TEXT"},
		{name: "TokenVarBegin", tokenType: TokenVarBegin, expected: "VAR_BEGIN"},
		{name: "TokenVarEnd", tokenType: TokenVarEnd, expected: "VAR_END"},
		{name: "TokenTagBegin", tokenType: TokenTagBegin, expected: "TAG_BEGIN"},
		{name: "TokenTagEnd", tokenType: TokenTagEnd, expected: "TAG_END"},
		{name: "TokenIdentifier", tokenType: TokenIdentifier, expected: "IDENTIFIER"},
		{name: "TokenString", tokenType: TokenString, expected: "STRING"},
		{name: "TokenNumber", tokenType: TokenNumber, expected: "NUMBER"},
		{name: "TokenSymbol", tokenType: TokenSymbol, expected: "SYMBOL"},
		{name: "unknown type 999", tokenType: TokenType(999), expected: fmt.Sprintf("UNKNOWN(%d)", 999)},
		{name: "unknown type -1", tokenType: TokenType(-1), expected: fmt.Sprintf("UNKNOWN(%d)", -1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tokenType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenString(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected string
	}{
		{
			name:     "EOF token",
			token:    Token{Type: TokenEOF, Line: 1, Col: 1},
			expected: "EOF at line 1, col 1",
		},
		{
			name:     "EOF at different position",
			token:    Token{Type: TokenEOF, Line: 10, Col: 25},
			expected: "EOF at line 10, col 25",
		},
		{
			name:     "short identifier",
			token:    Token{Type: TokenIdentifier, Value: "name", Line: 1, Col: 5},
			expected: `IDENTIFIER("name") at line 1, col 5`,
		},
		{
			name:     "empty value",
			token:    Token{Type: TokenText, Value: "", Line: 1, Col: 1},
			expected: `TEXT("") at line 1, col 1`,
		},
		{
			name:     "exactly 20 characters not truncated",
			token:    Token{Type: TokenText, Value: "12345678901234567890", Line: 2, Col: 1},
			expected: `TEXT("12345678901234567890") at line 2, col 1`,
		},
		{
			name:     "21 characters truncated",
			token:    Token{Type: TokenText, Value: "123456789012345678901", Line: 3, Col: 1},
			expected: `TEXT("12345678901234567890"...) at line 3, col 1`,
		},
		{
			name:     "long text truncated",
			token:    Token{Type: TokenText, Value: "This is a very long text that exceeds twenty characters", Line: 1, Col: 1},
			expected: `TEXT("This is a very long "...) at line 1, col 1`,
		},
		{
			name:     "string literal token",
			token:    Token{Type: TokenString, Value: "hello", Line: 1, Col: 10},
			expected: `STRING("hello") at line 1, col 10`,
		},
		{
			name:     "number token",
			token:    Token{Type: TokenNumber, Value: "42", Line: 5, Col: 3},
			expected: `NUMBER("42") at line 5, col 3`,
		},
		{
			name:     "symbol token",
			token:    Token{Type: TokenSymbol, Value: "==", Line: 1, Col: 8},
			expected: `SYMBOL("==") at line 1, col 8`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsKeyword(t *testing.T) {
	tests := []struct {
		name     string
		ident    string
		expected bool
	}{
		// All valid keywords
		{name: "in", ident: "in", expected: true},
		{name: "and", ident: "and", expected: true},
		{name: "or", ident: "or", expected: true},
		{name: "not", ident: "not", expected: true},
		{name: "true", ident: "true", expected: true},
		{name: "false", ident: "false", expected: true},
		{name: "if", ident: "if", expected: true},
		{name: "elif", ident: "elif", expected: true},
		{name: "else", ident: "else", expected: true},
		{name: "endif", ident: "endif", expected: true},
		{name: "for", ident: "for", expected: true},
		{name: "endfor", ident: "endfor", expected: true},
		{name: "break", ident: "break", expected: true},
		{name: "continue", ident: "continue", expected: true},
		// Non-keywords
		{name: "regular identifier name", ident: "name", expected: false},
		{name: "regular identifier render", ident: "render", expected: false},
		{name: "empty string", ident: "", expected: false},
		{name: "uppercase IF is not keyword", ident: "IF", expected: false},
		{name: "uppercase TRUE is not keyword", ident: "TRUE", expected: false},
		{name: "mixed case Not is not keyword", ident: "Not", expected: false},
		{name: "partial keyword end", ident: "end", expected: false},
		{name: "partial keyword el", ident: "el", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKeyword(tt.ident)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSymbol(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Comparison operators
		{name: "equal", input: "==", expected: true},
		{name: "not equal", input: "!=", expected: true},
		{name: "less than", input: "<", expected: true},
		{name: "greater than", input: ">", expected: true},
		{name: "less or equal", input: "<=", expected: true},
		{name: "greater or equal", input: ">=", expected: true},
		// Arithmetic operators
		{name: "plus", input: "+", expected: true},
		{name: "minus", input: "-", expected: true},
		{name: "multiply", input: "*", expected: true},
		{name: "divide", input: "/", expected: true},
		{name: "modulo", input: "%", expected: true},
		// Logical operators
		{name: "logical and", input: "&&", expected: true},
		{name: "logical or", input: "||", expected: true},
		{name: "logical not", input: "!", expected: true},
		// Other symbols
		{name: "pipe", input: "|", expected: true},
		{name: "colon", input: ":", expected: true},
		{name: "comma", input: ",", expected: true},
		{name: "dot", input: ".", expected: true},
		{name: "left paren", input: "(", expected: true},
		{name: "right paren", input: ")", expected: true},
		{name: "left bracket", input: "[", expected: true},
		{name: "right bracket", input: "]", expected: true},
		{name: "assignment", input: "=", expected: true},
		// Non-symbols
		{name: "word is not symbol", input: "abc", expected: false},
		{name: "empty string is not symbol", input: "", expected: false},
		{name: "triple equals is not symbol", input: "===", expected: false},
		{name: "power operator is not symbol", input: "**", expected: false},
		{name: "arrow is not symbol", input: "->", expected: false},
		{name: "left brace is not symbol", input: "{", expected: false},
		{name: "right brace is not symbol", input: "}", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSymbol(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
