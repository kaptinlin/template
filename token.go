package template

import (
	"strconv"
	"strings"
)

// TokenType represents the type of a token.
type TokenType int

const (
	// TokenError indicates a lexical error.
	TokenError TokenType = iota

	// TokenEOF indicates the end of input.
	TokenEOF

	// TokenText represents plain text outside of template tags.
	TokenText

	// TokenVarBegin represents the {{ variable tag opener.
	TokenVarBegin

	// TokenVarEnd represents the }} variable tag closer.
	TokenVarEnd

	// TokenTagBegin represents the {% block tag opener.
	TokenTagBegin

	// TokenTagEnd represents the %} block tag closer.
	TokenTagEnd

	// TokenIdentifier represents an identifier such as a variable name or keyword.
	TokenIdentifier

	// TokenString represents a quoted string literal.
	TokenString

	// TokenNumber represents an integer or floating-point literal.
	TokenNumber

	// TokenSymbol represents an operator or punctuation character.
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
	var b strings.Builder
	b.Grow(48) // pre-size for typical output

	name := t.Type.String()
	b.WriteString(name)

	if t.Type != TokenEOF {
		b.WriteByte('(')
		v := t.Value
		if len(v) > 20 {
			b.WriteString(strconv.Quote(v[:20]))
			b.WriteString("...")
		} else {
			b.WriteString(strconv.Quote(v))
		}
		b.WriteByte(')')
	}

	b.WriteString(" at line ")
	b.WriteString(strconv.Itoa(t.Line))
	b.WriteString(", col ")
	b.WriteString(strconv.Itoa(t.Col))
	return b.String()
}

// IsKeyword reports whether ident is a reserved keyword.
func IsKeyword(ident string) bool {
	switch ident {
	case "in", "and", "or", "not", "true", "false",
		"if", "elif", "else", "endif",
		"for", "endfor", "break", "continue":
		return true
	}
	return false
}

// IsSymbol reports whether s is a valid operator or punctuation symbol.
func IsSymbol(s string) bool {
	switch s {
	case "==", "!=", "<", ">", "<=", ">=",
		"+", "-", "*", "/", "%",
		"&&", "||", "!",
		"|", ":", ",", ".",
		"(", ")", "[", "]", "=":
		return true
	}
	return false
}
