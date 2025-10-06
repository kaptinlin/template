// Package template provides a simple and efficient template engine for Go.
package template

import (
	"fmt"
)

// TokenType represents the type of a token in the template language.
type TokenType int

const (
	// TokenIdentifier represents a variable identifier (e.g., user.age)
	TokenIdentifier TokenType = iota
	// TokenBool represents a boolean value (e.g., true, false)
	TokenBool
	// TokenNumber represents a number (e.g., 18)
	TokenNumber
	// TokenString represents a string constant
	TokenString
	// TokenOperator represents operators (e.g., ==, !=, <, >, &&, ||)
	TokenOperator
	// TokenArithOp represents arithmetic operators
	TokenArithOp
	// TokenNot represents the not operator (!)
	TokenNot
	// TokenLParen represents a left parenthesis (()
	TokenLParen
	// TokenRParen represents a right parenthesis ())
	TokenRParen
	// TokenPipe represents the pipe operator (|)
	TokenPipe
	// TokenFilter represents a filter (including name and args, e.g., upper, truncate:30)
	TokenFilter
	// TokenEOF represents the end of input marker
	TokenEOF
	// TokenDot represents the dot operator (.)
	TokenDot
)

// Token is a token in the template language.
// It represents a type and a value.
type Token struct {
	Typ TokenType // Token type
	Val string    // Token value
}

// Lexer tokenizes template expressions into a sequence of tokens.
type Lexer struct {
	input  string  // Input expression
	pos    int     // Current scanning position
	start  int     // Start position of current token
	tokens []Token // Generated tokens
}

// Lex tokenizes the input string into a list of tokens.
func (l *Lexer) Lex() ([]Token, error) {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		switch {
		case ch == '.':
			l.pos++
			l.emit(TokenDot)
			l.start = l.pos
		case isSpace(ch):
			l.ignore() // Ignore whitespace
		case isDigit(ch):
			l.lexNumber() // Parse number
		case ch == '"' || ch == '\'':
			if err := l.lexString(); err != nil { // Parse string
				return nil, err
			}
		case isLetter(ch) || ch == '_':
			l.lexIdentifierOrKeyword() // Parse identifier or keyword
		case ch == '(':
			l.pos++
			l.emit(TokenLParen) // Left parenthesis
			l.start = l.pos
		case ch == ')':
			l.pos++
			l.emit(TokenRParen) // Right parenthesis
			l.start = l.pos
		case isArithOperator(ch):
			l.lexArithOperator() // Parse arithmetic operator
		case isOperatorChar(ch, l.pos, l.input):
			l.lexOperator() // Parse operator
		case ch == '!':
			l.lexNot() // Parse not operator
		case ch == '|':
			l.lexPipeOrFilter() // Parse pipe or filter
		default:
			return nil, fmt.Errorf("%w: %c", ErrUnexpectedCharacter, ch)
		}
	}
	l.emit(TokenEOF) // Add end marker
	return l.tokens, nil
}

func (l *Lexer) lexNumber() {
	for l.pos < len(l.input) && (isDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
		l.pos++
	}
	l.emit(TokenNumber)
}

func (l *Lexer) lexIdentifierOrKeyword() {
	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos]) || l.input[l.pos] == '.' || l.input[l.pos] == '_') {
		l.pos++
	}
	val := l.input[l.start:l.pos]

	// Check for boolean literals (case-insensitive)
	if val == "true" || val == "True" || val == "false" || val == "False" {
		l.emit(TokenBool)
		return
	}

	// Check for text operators
	if val == "and" || val == "or" || val == "not" || val == "in" {
		// Check for multi-word operators like "not in"
		if val == "not" {
			// Look ahead for "in"
			savedPos := l.pos
			// Skip whitespace
			for l.pos < len(l.input) && isSpace(l.input[l.pos]) {
				l.pos++
			}
			// Check if next word is "in"
			nextStart := l.pos
			for l.pos < len(l.input) && isLetter(l.input[l.pos]) {
				l.pos++
			}
			if l.pos > nextStart && l.input[nextStart:l.pos] == "in" {
				// Emit "not in" as single operator
				l.tokens = append(l.tokens, Token{Typ: TokenOperator, Val: "not in"})
				l.start = l.pos
				return
			}
			// Restore position if not "not in"
			l.pos = savedPos
		}
		l.emit(TokenOperator)
		return
	}

	// Otherwise it's an identifier (including null/Null/none/None literals)
	l.emit(TokenIdentifier)
}

func (l *Lexer) lexNot() {
	l.pos++
	l.emit(TokenNot)
}

func (l *Lexer) lexPipeOrFilter() {
	l.pos++ // Skip '|'
	l.emit(TokenPipe)

	// Skip whitespace
	for l.pos < len(l.input) && isSpace(l.input[l.pos]) {
		l.pos++
	}
	l.start = l.pos

	if l.pos < len(l.input) && isLetter(l.input[l.pos]) {
		// Parse filter name and arguments (if any)
		for l.pos < len(l.input) {
			ch := l.input[l.pos]
			// Allow letters, digits, colons, commas, quotes and spaces as part of filter
			if isLetter(ch) || isDigit(ch) || ch == ':' || ch == ',' ||
				ch == '"' || ch == '\'' || ch == ' ' || ch == '_' {
				l.pos++
				continue
			}
			// Stop at other characters
			break
		}

		// Remove trailing spaces
		end := l.pos
		for end > l.start && isSpace(l.input[end-1]) {
			end--
		}

		// Emit filter token with complete filter expression
		l.tokens = append(l.tokens, Token{
			Typ: TokenFilter,
			Val: l.input[l.start:end],
		})
		l.start = l.pos
	}
}

func (l *Lexer) lexOperator() {
	// Support multi-character operators (e.g., ==, !=, &&, ||)
	if l.pos+1 < len(l.input) {
		doubleOp := l.input[l.pos : l.pos+2]
		if isOperator(doubleOp) {
			l.pos += 2
			l.emit(TokenOperator)
			return
		}
	}
	// Single character operator
	l.pos++
	l.emit(TokenOperator)
}

func (l *Lexer) lexString() error {
	quote := l.input[l.pos] // Store quote type
	l.pos++                 // Skip opening quote
	l.start = l.pos         // Start recording after quote

	for l.pos < len(l.input) {
		if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) {
			l.pos += 2 // Skip escape character
			continue
		}
		if l.input[l.pos] == quote {
			val := l.input[l.start:l.pos]
			l.pos++ // Skip closing quote
			l.tokens = append(l.tokens, Token{Typ: TokenString, Val: val})
			l.start = l.pos
			return nil
		}
		l.pos++
	}
	return ErrUnterminatedString
}

func (l *Lexer) lexArithOperator() {
	l.pos++
	l.emit(TokenArithOp)
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isOperatorChar(ch byte, pos int, input string) bool {
	if ch == '|' && pos+1 < len(input) && input[pos+1] != '|' {
		return false
	}
	if ch == '!' && pos+1 < len(input) && input[pos+1] != '=' {
		return false
	}
	return ch == '=' || ch == '!' || ch == '<' || ch == '>' || ch == '&' || ch == '|'
}

func isOperator(op string) bool {
	operators := []string{
		"==", "!=", "<=", ">=", "<", ">",
		"&&", "||", // C-style (backward compatibility)
		"and", "or", "not", // Django/Liquid style
		"in", "not in", // Membership operators
	}
	for _, o := range operators {
		if op == o {
			return true
		}
	}
	return false
}

func isArithOperator(ch byte) bool {
	return ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '%'
}

func (l *Lexer) emit(typ TokenType) {
	if typ != TokenEOF {
		val := l.input[l.start:l.pos]
		l.tokens = append(l.tokens, Token{Typ: typ, Val: val})
	} else {
		l.tokens = append(l.tokens, Token{Typ: typ, Val: "EOF"})
	}
	l.start = l.pos
}

func (l *Lexer) ignore() {
	l.pos++
	l.start = l.pos
}
