package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer_Lex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "Complex conditional expression",
			input: "user.age >= 18 && (is_admin || !is_guest) | upper",
			expected: []Token{
				{Typ: TokenIdentifier, Val: "user.age"},
				{Typ: TokenOperator, Val: ">="},
				{Typ: TokenNumber, Val: "18"},
				{Typ: TokenOperator, Val: "&&"},
				{Typ: TokenLParen, Val: "("},
				{Typ: TokenIdentifier, Val: "is_admin"},
				{Typ: TokenOperator, Val: "||"},
				{Typ: TokenNot, Val: "!"},
				{Typ: TokenIdentifier, Val: "is_guest"},
				{Typ: TokenRParen, Val: ")"},
				{Typ: TokenPipe, Val: "|"},
				{Typ: TokenFilter, Val: "upper"},
				{Typ: TokenEOF, Val: "EOF"},
			},
		},
		{
			name:  "Simple boolean expression",
			input: "is_active == true",
			expected: []Token{
				{Typ: TokenIdentifier, Val: "is_active"},
				{Typ: TokenOperator, Val: "=="},
				{Typ: TokenBool, Val: "true"},
				{Typ: TokenEOF, Val: "EOF"},
			},
		},
		{
			name:  "Numeric comparison expression",
			input: "count <= 10 && count > 0",
			expected: []Token{
				{Typ: TokenIdentifier, Val: "count"},
				{Typ: TokenOperator, Val: "<="},
				{Typ: TokenNumber, Val: "10"},
				{Typ: TokenOperator, Val: "&&"},
				{Typ: TokenIdentifier, Val: "count"},
				{Typ: TokenOperator, Val: ">"},
				{Typ: TokenNumber, Val: "0"},
				{Typ: TokenEOF, Val: "EOF"},
			},
		},
		{
			name:  "String concatenation expression",
			input: `user.name + " is " + user.age + " years old" 3.3 * user.age`,
			expected: []Token{
				{Typ: TokenIdentifier, Val: "user.name"},
				{Typ: TokenArithOp, Val: "+"},
				{Typ: TokenString, Val: " is "},
				{Typ: TokenArithOp, Val: "+"},
				{Typ: TokenIdentifier, Val: "user.age"},
				{Typ: TokenArithOp, Val: "+"},
				{Typ: TokenString, Val: " years old"},
				{Typ: TokenNumber, Val: "3.3"},
				{Typ: TokenArithOp, Val: "*"},
				{Typ: TokenIdentifier, Val: "user.age"},
				{Typ: TokenEOF, Val: "EOF"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := &Lexer{
				input: tt.input,
			}
			got, err := lexer.Lex()
			require.NoError(t, err)

			if !assert.Equal(t, tt.expected, got) {
				t.Logf("\nActual tokens:")
				for i, token := range got {
					t.Logf("%d: {Type: %v, Val: %q}", i, token.Typ, token.Val)
				}
				t.Logf("\nExpected tokens:")
				for i, token := range tt.expected {
					t.Logf("%d: {Type: %v, Val: %q}", i, token.Typ, token.Val)
				}
			}
		})
	}
}
