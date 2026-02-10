package template

import (
	"strconv"
	"strings"
)

// Lexer performs lexical analysis on template input.
type Lexer struct {
	input  string
	pos    int
	line   int // 1-based
	col    int // 1-based
	tokens []*Token
}

// NewLexer creates a new Lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		line:   1,
		col:    1,
		tokens: make([]*Token, 0, 100),
	}
}

// Tokenize performs lexical analysis and returns all tokens.
func (l *Lexer) Tokenize() ([]*Token, error) {
	for l.pos < len(l.input) {
		if l.peek("{#") {
			if err := l.scanComment(); err != nil {
				return nil, err
			}
			continue
		}

		if l.peek("{{") {
			if err := l.scanVarTag(); err != nil {
				return nil, err
			}
			continue
		}

		if l.peek("{%") {
			if err := l.scanBlockTag(); err != nil {
				return nil, err
			}
			continue
		}

		if err := l.scanText(); err != nil {
			return nil, err
		}
	}

	l.emit(TokenEOF, "")
	return l.tokens, nil
}

// scanText scans plain text until a tag opener is encountered.
func (l *Lexer) scanText() error {
	start := l.pos
	startLine, startCol := l.line, l.col

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if ch == '{' && l.pos+1 < len(l.input) {
			next := l.input[l.pos+1]
			if next == '{' || next == '%' || next == '#' {
				break
			}
		}

		if ch == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}

		l.pos++
	}

	if l.pos > start {
		l.emitAt(TokenText, l.input[start:l.pos], startLine, startCol)
	}

	return nil
}

// scanComment scans and discards a comment {# ... #}.
func (l *Lexer) scanComment() error {
	startLine, startCol := l.line, l.col
	l.advance(2) // skip {#

	for !l.peek("#}") {
		if l.pos >= len(l.input) {
			return l.errorAt(startLine, startCol, "unclosed comment, expected '#}'")
		}
		if l.input[l.pos] == '\n' {
			return l.errorAt(l.line, l.col, "newline not permitted in comment")
		}
		l.pos++
		l.col++
	}

	l.advance(2) // skip #}
	return nil
}

// scanVarTag scans a variable tag {{ ... }}.
func (l *Lexer) scanVarTag() error {
	startLine, startCol := l.line, l.col

	l.emit(TokenVarBegin, "{{")
	l.advance(2)

	for !l.peek("}}") {
		if l.pos >= len(l.input) {
			return l.errorAt(startLine, startCol, "unclosed variable tag, expected '}}'")
		}

		l.skipWhitespace()

		if l.peek("}}") {
			break
		}

		if err := l.scanInsideTag(); err != nil {
			return err
		}
	}

	l.emit(TokenVarEnd, "}}")
	l.advance(2)

	return nil
}

// scanBlockTag scans a block tag {% ... %}.
func (l *Lexer) scanBlockTag() error {
	startLine, startCol := l.line, l.col

	l.emit(TokenTagBegin, "{%")
	l.advance(2)

	for !l.peek("%}") {
		if l.pos >= len(l.input) {
			return l.errorAt(startLine, startCol, "unclosed block tag, expected '%}'")
		}

		l.skipWhitespace()

		if l.peek("%}") {
			break
		}

		if err := l.scanInsideTag(); err != nil {
			return err
		}
	}

	l.emit(TokenTagEnd, "%}")
	l.advance(2)

	return nil
}

// scanInsideTag scans a single token inside {{ }} or {% %}.
func (l *Lexer) scanInsideTag() error {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return nil
	}

	ch := l.input[l.pos]

	switch {
	case ch == '"' || ch == '\'':
		return l.scanString()
	case isDigit(ch):
		return l.scanNumber()
	case isLetter(ch) || ch == '_':
		return l.scanIdentifier()
	default:
		return l.scanSymbol()
	}
}

// scanIdentifier scans an identifier or keyword.
func (l *Lexer) scanIdentifier() error {
	start := l.pos
	startLine, startCol := l.line, l.col

	l.pos++
	l.col++

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			break
		}
		l.pos++
		l.col++
	}

	l.emitAt(TokenIdentifier, l.input[start:l.pos], startLine, startCol)
	return nil
}

// scanNumber scans a number literal (integer or float).
func (l *Lexer) scanNumber() error {
	start := l.pos
	startLine, startCol := l.line, l.col

	// If preceded by '.', treat as array index — don't consume decimal point.
	propAccess := start > 0 && l.input[start-1] == '.'

	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.pos++
		l.col++
	}

	// Consume decimal point only if not in a property access chain
	// and the '.' is followed by a digit.
	if !propAccess && l.pos < len(l.input) && l.input[l.pos] == '.' {
		if l.pos+1 < len(l.input) && isDigit(l.input[l.pos+1]) {
			l.pos++
			l.col++
			for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
				l.pos++
				l.col++
			}
		}
	}

	l.emitAt(TokenNumber, l.input[start:l.pos], startLine, startCol)
	return nil
}

// scanString scans a quoted string literal ("..." or '...').
func (l *Lexer) scanString() error {
	startLine, startCol := l.line, l.col
	quote := l.input[l.pos]

	l.pos++
	l.col++

	var buf strings.Builder
	buf.Grow(16)
	escaped := false

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if escaped {
			switch ch {
			case '"':
				buf.WriteByte('"')
			case '\'':
				buf.WriteByte('\'')
			case '\\':
				buf.WriteByte('\\')
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case 'r':
				buf.WriteByte('\r')
			default:
				return l.errorAt(l.line, l.col-1, "unknown escape sequence: \\"+string(ch))
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

		if ch == quote {
			l.emitAt(TokenString, buf.String(), startLine, startCol)
			l.pos++
			l.col++
			return nil
		}

		if ch == '\n' {
			return l.errorAt(startLine, startCol, "newline in string is not allowed")
		}

		buf.WriteByte(ch)
		l.col++
		l.pos++
	}

	return l.errorAt(startLine, startCol, "unclosed string, expected "+string(quote))
}

// scanSymbol scans an operator or punctuation symbol.
func (l *Lexer) scanSymbol() error {
	// Try two-character symbols first.
	if l.pos+1 < len(l.input) {
		two := l.input[l.pos : l.pos+2]
		if IsSymbol(two) {
			l.emit(TokenSymbol, two)
			l.pos += 2
			l.col += 2
			return nil
		}
	}

	one := l.input[l.pos : l.pos+1]
	if IsSymbol(one) {
		l.emit(TokenSymbol, one)
		l.pos++
		l.col++
		return nil
	}

	return l.errorAt(l.line, l.col, "unexpected character: "+one)
}

// emit appends a token at the current position.
func (l *Lexer) emit(typ TokenType, value string) {
	l.tokens = append(l.tokens, &Token{
		Type:  typ,
		Value: value,
		Line:  l.line,
		Col:   l.col,
	})
}

// emitAt appends a token with an explicit position.
func (l *Lexer) emitAt(typ TokenType, value string, line, col int) {
	l.tokens = append(l.tokens, &Token{
		Type:  typ,
		Value: value,
		Line:  line,
		Col:   col,
	})
}

// peek reports whether the input at the current position starts with s.
func (l *Lexer) peek(s string) bool {
	return l.pos+len(s) <= len(l.input) && l.input[l.pos:l.pos+len(s)] == s
}

// advance moves the position forward by n characters, tracking lines.
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

// skipWhitespace advances past any whitespace characters.
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			break
		}
		if ch == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
}

// errorAt returns a LexerError at the given position.
func (l *Lexer) errorAt(line, col int, msg string) error {
	return &LexerError{
		Message: msg,
		Line:    line,
		Col:     col,
	}
}

// LexerError represents a lexical analysis error with position information.
type LexerError struct {
	Message string
	Line    int
	Col     int
}

// Error implements the error interface.
func (e *LexerError) Error() string {
	return "lexer error at line " + strconv.Itoa(e.Line) + ", col " + strconv.Itoa(e.Col) + ": " + e.Message
}

// isLetter reports whether the byte is an ASCII letter.
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isDigit reports whether the byte is an ASCII digit.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
