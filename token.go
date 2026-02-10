package template

import "fmt"

// TokenType represents the type of a token.
type TokenType int

const (
	// TokenError indicates a lexical error.
	TokenError TokenType = iota

	// TokenEOF indicates the end of input.
	TokenEOF

	// TokenText represents plain text outside of template tags.
	TokenText

	// TokenVarBegin represents {{ (variable tag start).
	TokenVarBegin

	// TokenVarEnd represents }} (variable tag end).
	TokenVarEnd

	// TokenTagBegin represents {% (block tag start).
	TokenTagBegin

	// TokenTagEnd represents %} (block tag end).
	TokenTagEnd

	// TokenIdentifier represents an identifier (variable name, keyword, etc).
	TokenIdentifier

	// TokenString represents a string literal ("..." or '...').
	TokenString

	// TokenNumber represents a number literal.
	TokenNumber

	// TokenSymbol represents operators and punctuation (+, -, *, /, ==, !=, |, :, ,, ., etc).
	TokenSymbol
)

// Token represents a single lexical token.
type Token struct {
	Type  TokenType // Token type
	Value string    // Token text content
	Line  int       // Line number (1-based)
	Col   int       // Column number (1-based)
}

// String returns a string representation of the token type.
func (t TokenType) String() string {
	switch t {
	case TokenError:
		return "ERROR"
	case TokenEOF:
		return "EOF"
	case TokenText:
		return "TEXT"
	case TokenVarBegin:
		return "VAR_BEGIN"
	case TokenVarEnd:
		return "VAR_END"
	case TokenTagBegin:
		return "TAG_BEGIN"
	case TokenTagEnd:
		return "TAG_END"
	case TokenIdentifier:
		return "IDENTIFIER"
	case TokenString:
		return "STRING"
	case TokenNumber:
		return "NUMBER"
	case TokenSymbol:
		return "SYMBOL"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// String returns a string representation of the token.
func (t *Token) String() string {
	if t.Type == TokenEOF {
		return fmt.Sprintf("%s at line %d, col %d", t.Type, t.Line, t.Col)
	}
	if len(t.Value) > 20 {
		return fmt.Sprintf("%s(%q...) at line %d, col %d", t.Type, t.Value[:20], t.Line, t.Col)
	}
	return fmt.Sprintf("%s(%q) at line %d, col %d", t.Type, t.Value, t.Line, t.Col)
}

// IsKeyword checks if an identifier is a reserved keyword.
func IsKeyword(ident string) bool {
	keywords := map[string]bool{
		"in":       true,
		"and":      true,
		"or":       true,
		"not":      true,
		"true":     true,
		"false":    true,
		"if":       true,
		"elif":     true,
		"else":     true,
		"endif":    true,
		"for":      true,
		"endfor":   true,
		"break":    true,
		"continue": true,
	}
	return keywords[ident]
}

// IsSymbol checks if a string is a valid symbol.
func IsSymbol(s string) bool {
	symbols := map[string]bool{
		// Comparison operators
		"==": true,
		"!=": true,
		"<":  true,
		">":  true,
		"<=": true,
		">=": true,

		// Arithmetic operators
		"+": true,
		"-": true,
		"*": true,
		"/": true,
		"%": true,

		// Logical operators
		"&&": true,
		"||": true,
		"!":  true,

		// Other symbols
		"|": true, // Filter pipe
		":": true, // Filter argument separator
		",": true, // Argument separator
		".": true, // Property access
		"(": true, // Parenthesis
		")": true,
		"[": true, // Bracket
		"]": true,
		"=": true, // Assignment (for {% set %})
	}
	return symbols[s]
}
