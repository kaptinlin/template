package template

import (
	"strconv"
	"strings"
)

// LexerError represents a lexical analysis error with position information.
type LexerError struct {
	Message string
	Line    int
	Col     int
}

// Error implements the error interface.
func (e *LexerError) Error() string {
	var b strings.Builder
	b.Grow(48)
	b.WriteString("lexer error at line ")
	b.WriteString(strconv.Itoa(e.Line))
	b.WriteString(", col ")
	b.WriteString(strconv.Itoa(e.Col))
	b.WriteString(": ")
	b.WriteString(e.Message)
	return b.String()
}

// Lexer performs lexical analysis on template input.
type Lexer struct {
	input  string
	pos    int
	line   int // 1-based
	col    int // 1-based
	tokens []*Token
	len    int // cached len(input)

	// allowRaw enables {% raw %}...{% endraw %} block scanning. It is
	// false by default so Compile(src) retains its original behavior
	// (raw is treated as an unknown tag). Set.Get enables it via
	// compileForSet so templates loaded via NewHTMLSet/NewTextSet get
	// the raw feature.
	allowRaw bool
}

// NewLexer creates a new Lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		line:   1,
		col:    1,
		tokens: make([]*Token, 0, 100),
		len:    len(input),
	}
}

// Tokenize performs lexical analysis and returns all tokens.
func (l *Lexer) Tokenize() ([]*Token, error) {
	for l.pos < l.len {
		// Check for tag openers: all start with '{'.
		if l.input[l.pos] == '{' && l.pos+1 < l.len {
			switch l.input[l.pos+1] {
			case '#':
				if err := l.scanComment(); err != nil {
					return nil, err
				}
				continue
			case '{':
				if err := l.scanVarTag(); err != nil {
					return nil, err
				}
				continue
			case '%':
				// Intercept {% raw %}...{% endraw %} at the lexer level
				// — but only when allowRaw is enabled. Compile(src)
				// leaves it off, so raw is treated as an unknown tag in
				// that path.
				if l.allowRaw {
					if n := l.matchRawOpen(); n > 0 {
						if err := l.scanRawBlock(n); err != nil {
							return nil, err
						}
						continue
					}
				}
				if err := l.scanBlockTag(); err != nil {
					return nil, err
				}
				continue
			}
		}

		if err := l.scanText(); err != nil {
			return nil, err
		}
	}

	l.emit(TokenEOF, "")
	return l.tokens, nil
}

// matchRawOpen reports the byte length of a {% raw %} opener at the
// current position, allowing for flexible whitespace. Returns 0 if the
// current position is not a raw opener.
func (l *Lexer) matchRawOpen() int {
	return l.matchBlockTagKeyword("raw")
}

// matchBlockTagKeyword returns the byte length (including "%}") of a
// block tag "{% kw %}" at the current position, or 0 on no match.
// It tolerates whitespace around the keyword.
func (l *Lexer) matchBlockTagKeyword(kw string) int {
	i := l.pos
	if i+2 > l.len || l.input[i] != '{' || l.input[i+1] != '%' {
		return 0
	}
	i += 2
	// Skip whitespace after {%.
	for i < l.len && isSpace(l.input[i]) {
		i++
	}
	// Match keyword.
	if i+len(kw) > l.len || l.input[i:i+len(kw)] != kw {
		return 0
	}
	j := i + len(kw)
	// Must be followed by whitespace or %}.
	if j >= l.len {
		return 0
	}
	if !isSpace(l.input[j]) && (j+1 >= l.len || l.input[j] != '%' || l.input[j+1] != '}') {
		return 0
	}
	// Skip whitespace after keyword.
	for j < l.len && isSpace(l.input[j]) {
		j++
	}
	// Expect %}.
	if j+1 >= l.len || l.input[j] != '%' || l.input[j+1] != '}' {
		return 0
	}
	return j + 2 - l.pos
}

// scanRawBlock consumes everything from the current {% raw %} opener
// (openerLen bytes long) up to a matching {% endraw %}, emitting a
// single TokenText for the body. Returns ErrUnclosedRaw if no endraw
// is found.
func (l *Lexer) scanRawBlock(openerLen int) error {
	startLine, startCol := l.line, l.col
	// Skip the {% raw %} opener, tracking line/col.
	l.advance(openerLen)

	bodyStart := l.pos
	bodyLine, bodyCol := l.line, l.col

	for l.pos < l.len {
		if n := l.matchBlockTagKeyword("endraw"); n > 0 {
			body := l.input[bodyStart:l.pos]
			if len(body) > 0 {
				l.emitAt(TokenText, body, bodyLine, bodyCol)
			}
			l.advance(n)
			return nil
		}
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
	return l.errorAtErr(startLine, startCol, ErrUnclosedRaw)
}

// errorAtErr wraps a sentinel error with position info as a LexerError.
func (l *Lexer) errorAtErr(line, col int, sentinel error) error {
	return &LexerError{Message: sentinel.Error(), Line: line, Col: col}
}

// isSpace reports whether b is ASCII whitespace.
func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// scanText scans plain text until a tag opener is encountered.
func (l *Lexer) scanText() error {
	start := l.pos
	startLine, startCol := l.line, l.col

	for l.pos < l.len {
		ch := l.input[l.pos]
		if ch == '{' && l.pos+1 < l.len {
			switch l.input[l.pos+1] {
			case '{', '%', '#':
				goto done
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

done:
	if l.pos > start {
		l.emitAt(TokenText, l.input[start:l.pos], startLine, startCol)
	}
	return nil
}

// scanComment scans and discards a comment {# ... #}.
func (l *Lexer) scanComment() error {
	startLine, startCol := l.line, l.col
	l.advance(2) // skip {#

	for !l.peek2('#', '}') {
		if l.pos >= l.len {
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

	for !l.peek2('}', '}') {
		if l.pos >= l.len {
			return l.errorAt(startLine, startCol, "unclosed variable tag, expected '}}'")
		}

		l.skipWhitespace()

		if l.peek2('}', '}') {
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

	for !l.peek2('%', '}') {
		if l.pos >= l.len {
			return l.errorAt(startLine, startCol, "unclosed block tag, expected '%}'")
		}

		l.skipWhitespace()

		if l.peek2('%', '}') {
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

	if l.pos >= l.len {
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

	for l.pos < l.len {
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

	for l.pos < l.len && isDigit(l.input[l.pos]) {
		l.pos++
		l.col++
	}

	// Consume decimal point only if not in a property access chain
	// and the '.' is followed by a digit.
	if !propAccess && l.pos < l.len && l.input[l.pos] == '.' {
		if l.pos+1 < l.len && isDigit(l.input[l.pos+1]) {
			l.pos++
			l.col++
			for l.pos < l.len && isDigit(l.input[l.pos]) {
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

	for l.pos < l.len {
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
	ch := l.input[l.pos]

	// Try two-character symbols first.
	if l.pos+1 < l.len {
		next := l.input[l.pos+1]
		if isTwoCharSymbol(ch, next) {
			two := l.input[l.pos : l.pos+2]
			l.emit(TokenSymbol, two)
			l.pos += 2
			l.col += 2
			return nil
		}
	}

	if isOneCharSymbol(ch) {
		l.emit(TokenSymbol, l.input[l.pos:l.pos+1])
		l.pos++
		l.col++
		return nil
	}

	return l.errorAt(l.line, l.col, "unexpected character: "+string(ch))
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

// peek2 reports whether the next two bytes match a and b.
func (l *Lexer) peek2(a, b byte) bool {
	return l.pos+2 <= l.len && l.input[l.pos] == a && l.input[l.pos+1] == b
}

// advance moves the position forward by n bytes for tag delimiters
// (e.g. {{ }}, {# #}, {% %}) that never span newlines.
func (l *Lexer) advance(n int) {
	l.pos += n
	l.col += n
}

// skipWhitespace advances past any whitespace characters.
func (l *Lexer) skipWhitespace() {
	for l.pos < l.len {
		switch l.input[l.pos] {
		case '\n':
			l.line++
			l.col = 1
		case ' ', '\t', '\r':
			l.col++
		default:
			return
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

// isLetter reports whether the byte is an ASCII letter.
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isDigit reports whether the byte is an ASCII digit.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// isOneCharSymbol reports whether ch is a valid single-character symbol.
func isOneCharSymbol(ch byte) bool {
	switch ch {
	case '+', '-', '*', '/', '%',
		'<', '>', '!', '=',
		'|', ':', ',', '.',
		'(', ')', '[', ']':
		return true
	}
	return false
}

// isTwoCharSymbol reports whether a, b form a valid two-character symbol.
func isTwoCharSymbol(a, b byte) bool {
	switch a {
	case '=', '!', '<', '>':
		return b == '='
	case '&':
		return b == '&'
	case '|':
		return b == '|'
	}
	return false
}
