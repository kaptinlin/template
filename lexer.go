package template

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer performs lexical analysis on template input.
type Lexer struct {
	input  string   // Input template string
	pos    int      // Current position in input
	line   int      // Current line number (1-based)
	col    int      // Current column number (1-based)
	tokens []*Token // Collected tokens
}

// NewLexer creates a new lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		col:    1,
		tokens: make([]*Token, 0, 100),
	}
}

// Tokenize performs lexical analysis and returns all tokens.
func (l *Lexer) Tokenize() ([]*Token, error) {
	for l.pos < len(l.input) {
		// Check for comment {# #}
		if l.peek("{#") {
			if err := l.scanComment(); err != nil {
				return nil, err
			}
			continue
		}

		// Check for variable tag {{ }}
		if l.peek("{{") {
			if err := l.scanVarTag(); err != nil {
				return nil, err
			}
			continue
		}

		// Check for block tag {% %}
		if l.peek("{%") {
			if err := l.scanBlockTag(); err != nil {
				return nil, err
			}
			continue
		}

		// Otherwise, scan plain text
		if err := l.scanText(); err != nil {
			return nil, err
		}
	}

	// Add EOF token
	l.emit(TokenEOF, "")

	return l.tokens, nil
}

// scanText scans plain text until {{ or {% or {# is encountered.
func (l *Lexer) scanText() error {
	start := l.pos
	startLine, startCol := l.line, l.col

	for l.pos < len(l.input) {
		// Stop if we encounter a tag start or comment
		if l.peek("{{") || l.peek("{%") || l.peek("{#") {
			break
		}

		// Track line and column
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}

		l.pos++
	}

	// Only emit if we scanned some text
	if l.pos > start {
		text := l.input[start:l.pos]
		l.emitAt(TokenText, text, startLine, startCol)
	}

	return nil
}

// scanComment scans a comment {# ... #}.
// Comments are ignored and do not produce tokens.
func (l *Lexer) scanComment() error {
	startLine, startCol := l.line, l.col

	l.advance(2) // Skip {#

	for !l.peek("#}") {
		if l.pos >= len(l.input) {
			return l.errorAt(startLine, startCol, "unclosed comment, expected '#}'")
		}

		// Newlines are not allowed in comments (Django spec)
		if l.input[l.pos] == '\n' {
			return l.errorAt(l.line, l.col, "newline not permitted in comment")
		}

		l.pos++
		l.col++
	}

	l.advance(2) // Skip #}

	// Comments are ignored, no token is emitted
	return nil
}

// scanVarTag scans a variable tag {{ ... }}.
func (l *Lexer) scanVarTag() error {
	startLine, startCol := l.line, l.col

	// Emit {{
	l.emit(TokenVarBegin, "{{")
	l.advance(2)

	// Scan contents until }}
	for !l.peek("}}") {
		if l.pos >= len(l.input) {
			return l.errorAt(startLine, startCol, "unclosed variable tag, expected '}}'")
		}

		l.skipWhitespace()

		if l.peek("}}") {
			break
		}

		// Scan token inside the tag
		if err := l.scanInsideTag(); err != nil {
			return err
		}
	}

	// Emit }}
	l.emit(TokenVarEnd, "}}")
	l.advance(2)

	return nil
}

// scanBlockTag scans a block tag {% ... %}.
func (l *Lexer) scanBlockTag() error {
	startLine, startCol := l.line, l.col

	// Emit {%
	l.emit(TokenTagBegin, "{%")
	l.advance(2)

	// Scan contents until %}
	for !l.peek("%}") {
		if l.pos >= len(l.input) {
			return l.errorAt(startLine, startCol, "unclosed block tag, expected '%}'")
		}

		l.skipWhitespace()

		if l.peek("%}") {
			break
		}

		// Scan token inside the tag
		if err := l.scanInsideTag(); err != nil {
			return err
		}
	}

	// Emit %}
	l.emit(TokenTagEnd, "%}")
	l.advance(2)

	return nil
}

// scanInsideTag scans a single token inside {{ }} or {% %}.
func (l *Lexer) scanInsideTag() error {
	// Skip whitespace first
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return nil
	}

	ch := l.input[l.pos]

	// String literal
	if ch == '"' || ch == '\'' {
		return l.scanString()
	}

	// Number literal
	if unicode.IsDigit(rune(ch)) {
		return l.scanNumber()
	}

	// Identifier or keyword
	if unicode.IsLetter(rune(ch)) || ch == '_' {
		return l.scanIdentifier()
	}

	// Symbol (operators and punctuation)
	return l.scanSymbol()
}

// scanIdentifier scans an identifier (variable name, keyword, etc).
func (l *Lexer) scanIdentifier() error {
	start := l.pos
	startLine, startCol := l.line, l.col

	// First character is letter or underscore (already checked)
	l.pos++
	l.col++

	// Continue with letters, digits, or underscores
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' {
			l.pos++
			l.col++
		} else {
			break
		}
	}

	value := l.input[start:l.pos]
	l.emitAt(TokenIdentifier, value, startLine, startCol)

	return nil
}

// scanNumber scans a number literal.
func (l *Lexer) scanNumber() error {
	start := l.pos
	startLine, startCol := l.line, l.col

	// Check if this number is part of a property access chain (e.g., "items.0.0")
	// If the character before this number is '.', then this is an array index,
	// and we should NOT treat subsequent '.' as a decimal point.
	isPropAccess := start > 0 && l.input[start-1] == '.'

	// Scan digits
	for l.pos < len(l.input) && unicode.IsDigit(rune(l.input[l.pos])) {
		l.pos++
		l.col++
	}

	// Check for decimal point followed by a digit
	// Don't consume '.' if:
	// 1. It's not followed by a digit (e.g., "0.name" should be "0" and ".name")
	// 2. This number is part of a property access chain (e.g., ".0.0" should be ".0" and ".0")
	if !isPropAccess && l.pos < len(l.input) && l.input[l.pos] == '.' {
		// Peek ahead to check if there's a digit after '.'
		if l.pos+1 < len(l.input) && unicode.IsDigit(rune(l.input[l.pos+1])) {
			l.pos++ // consume '.'
			l.col++

			// Scan fractional part
			for l.pos < len(l.input) && unicode.IsDigit(rune(l.input[l.pos])) {
				l.pos++
				l.col++
			}
		}
		// If '.' is not followed by a digit, don't consume it - leave it for the next token
	}

	value := l.input[start:l.pos]
	l.emitAt(TokenNumber, value, startLine, startCol)

	return nil
}

// scanString scans a string literal ("..." or '...').
func (l *Lexer) scanString() error {
	startLine, startCol := l.line, l.col
	quote := l.input[l.pos] // " or '

	l.pos++ // Skip opening quote
	l.col++

	var value strings.Builder
	escaped := false

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		// Handle escape sequences
		if escaped {
			// Process escape sequence
			switch ch {
			case '"':
				value.WriteByte('"')
			case '\'':
				value.WriteByte('\'')
			case '\\':
				value.WriteByte('\\')
			case 'n':
				value.WriteByte('\n')
			case 't':
				value.WriteByte('\t')
			case 'r':
				value.WriteByte('\r')
			default:
				return l.errorAt(l.line, l.col-1, fmt.Sprintf("unknown escape sequence: \\%c", ch))
			}
			escaped = false
			l.pos++
			l.col++
			continue
		}

		if ch == '\\' {
			escaped = true
			l.pos++
			l.col++
			continue
		}

		// End of string
		if ch == quote {
			l.emitAt(TokenString, value.String(), startLine, startCol)

			l.pos++ // Skip closing quote
			l.col++
			return nil
		}

		// Newlines are not allowed in strings (Django spec)
		if ch == '\n' {
			return l.errorAt(startLine, startCol, "newline in string is not allowed")
		}

		value.WriteByte(ch)
		l.col++
		l.pos++
	}

	return l.errorAt(startLine, startCol, fmt.Sprintf("unclosed string, expected %c", quote))
}

// scanSymbol scans an operator or punctuation symbol.
func (l *Lexer) scanSymbol() error {
	// Try to match 2-character symbols first
	if l.pos+1 < len(l.input) {
		twoChar := l.input[l.pos : l.pos+2]
		if IsSymbol(twoChar) {
			l.emit(TokenSymbol, twoChar)
			l.pos += 2
			l.col += 2
			return nil
		}
	}

	// Try 1-character symbol
	oneChar := string(l.input[l.pos])
	if IsSymbol(oneChar) {
		l.emit(TokenSymbol, oneChar)
		l.pos++
		l.col++
		return nil
	}

	return l.errorAt(l.line, l.col, fmt.Sprintf("unexpected character: %c", l.input[l.pos]))
}

// emit creates and appends a token to the token list.
func (l *Lexer) emit(typ TokenType, value string) {
	token := &Token{
		Type:  typ,
		Value: value,
		Line:  l.line,
		Col:   l.col,
	}
	l.tokens = append(l.tokens, token)
}

// emitAt creates and appends a token with explicit line/col position.
func (l *Lexer) emitAt(typ TokenType, value string, line, col int) {
	token := &Token{
		Type:  typ,
		Value: value,
		Line:  line,
		Col:   col,
	}
	l.tokens = append(l.tokens, token)
}

// peek checks if the input starts with the given string at the current position.
func (l *Lexer) peek(s string) bool {
	if l.pos+len(s) > len(l.input) {
		return false
	}
	return l.input[l.pos:l.pos+len(s)] == s
}

// advance moves the position forward by n characters.
func (l *Lexer) advance(n int) {
	for i := 0; i < n && l.pos < len(l.input); i++ {
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
}

// skipWhitespace skips whitespace characters.
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			if ch == '\n' {
				l.line++
				l.col = 1
			} else {
				l.col++
			}
			l.pos++
		} else {
			break
		}
	}
}

// errorAt creates a lexer error at the specified position.
func (l *Lexer) errorAt(line, col int, msg string) error {
	return &LexerError{
		Message: msg,
		Line:    line,
		Col:     col,
	}
}

// LexerError represents a lexical analysis error.
type LexerError struct {
	Message string
	Line    int
	Col     int
}

// Error implements the error interface.
func (e *LexerError) Error() string {
	return fmt.Sprintf("lexer error at line %d, col %d: %s", e.Line, e.Col, e.Message)
}
