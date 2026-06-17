package template

import (
	"strconv"
	"strings"
)

// tokenType represents the type of a token.
type tokenType int

const (
	// tokenError indicates a lexical error.
	tokenError tokenType = iota

	// tokenEOF indicates the end of input.
	tokenEOF

	// tokenText represents plain text outside of template tags.
	tokenText

	// tokenVarBegin represents the {{ variable tag opener.
	tokenVarBegin

	// tokenVarEnd represents the }} variable tag closer.
	tokenVarEnd

	// tokenTagBegin represents the {% block tag opener.
	tokenTagBegin

	// tokenTagEnd represents the %} block tag closer.
	tokenTagEnd

	// tokenIdentifier represents an identifier such as a variable name or keyword.
	tokenIdentifier

	// tokenString represents a quoted string literal.
	tokenString

	// tokenNumber represents an integer or floating-point literal.
	tokenNumber

	// tokenSymbol represents an operator or punctuation character.
	tokenSymbol
)

// tokenTypeNames maps token types to their string representations.
var tokenTypeNames = [...]string{
	tokenError:      "ERROR",
	tokenEOF:        "EOF",
	tokenText:       "TEXT",
	tokenVarBegin:   "VAR_BEGIN",
	tokenVarEnd:     "VAR_END",
	tokenTagBegin:   "TAG_BEGIN",
	tokenTagEnd:     "TAG_END",
	tokenIdentifier: "IDENTIFIER",
	tokenString:     "STRING",
	tokenNumber:     "NUMBER",
	tokenSymbol:     "SYMBOL",
}

// token represents a single lexical token.
type token struct {
	Type  tokenType
	value string
	Line  int // 1-based
	Col   int // 1-based
}

// String returns a string representation of the token type.
func (t tokenType) String() string {
	if int(t) >= 0 && int(t) < len(tokenTypeNames) {
		return tokenTypeNames[t]
	}
	return "UNKNOWN(" + strconv.Itoa(int(t)) + ")"
}

// String returns a human-readable representation of the token.
func (t *token) String() string {
	var b strings.Builder
	b.Grow(48) // pre-size for typical output

	name := t.Type.String()
	b.WriteString(name)

	if t.Type != tokenEOF {
		b.WriteByte('(')
		v := t.value
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

// isKeyword reports whether ident is a reserved keyword.
func isKeyword(ident string) bool {
	switch ident {
	case "in", "and", "or", "not", "true", "false",
		"if", "elif", "else", "endif",
		"for", "endfor", "break", "continue":
		return true
	}
	return false
}

// isSymbol reports whether s is a valid operator or punctuation symbol.
func isSymbol(s string) bool {
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
