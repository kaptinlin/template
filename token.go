package template

import "strconv"

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

// tokenTypeNames maps token types to their string representations.
var tokenTypeNames = [...]string{
	TokenError:      "ERROR",
	TokenEOF:        "EOF",
	TokenText:       "TEXT",
	TokenVarBegin:   "VAR_BEGIN",
	TokenVarEnd:     "VAR_END",
	TokenTagBegin:   "TAG_BEGIN",
	TokenTagEnd:     "TAG_END",
	TokenIdentifier: "IDENTIFIER",
	TokenString:     "STRING",
	TokenNumber:     "NUMBER",
	TokenSymbol:     "SYMBOL",
}

// Token represents a single lexical token.
type Token struct {
	Type  TokenType
	Value string
	Line  int // 1-based
	Col   int // 1-based
}

// String returns a string representation of the token type.
func (t TokenType) String() string {
	if int(t) >= 0 && int(t) < len(tokenTypeNames) {
		return tokenTypeNames[t]
	}
	return "UNKNOWN(" + strconv.Itoa(int(t)) + ")"
}

// String returns a human-readable representation of the token.
func (t *Token) String() string {
	pos := " at line " + strconv.Itoa(t.Line) + ", col " + strconv.Itoa(t.Col)
	name := t.Type.String()

	if t.Type == TokenEOF {
		return name + pos
	}

	v := t.Value
	if len(v) > 20 {
		return name + "(" + strconv.Quote(v[:20]) + "...)" + pos
	}
	return name + "(" + strconv.Quote(v) + ")" + pos
}

// keywords is the set of reserved keywords.
var keywords = map[string]bool{
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

// IsKeyword checks if an identifier is a reserved keyword.
func IsKeyword(ident string) bool {
	return keywords[ident]
}

// symbols is the set of valid operator and punctuation symbols.
var symbols = map[string]bool{
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

// IsSymbol checks if a string is a valid symbol.
func IsSymbol(s string) bool {
	return symbols[s]
}
