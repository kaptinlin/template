package template

import (
	"fmt"
	"strconv"
)

// comparisonOps is the set of comparison operators for fast lookup.
var comparisonOps = map[string]bool{
	"==": true, "!=": true,
	"<": true, ">": true,
	"<=": true, ">=": true,
}

// ExprParser parses expressions from a token stream.
// It handles operator precedence, filters, property access, etc.
type ExprParser struct {
	tokens []*Token
	pos    int
}

// NewExprParser creates a new expression parser.
func NewExprParser(tokens []*Token) *ExprParser {
	return &ExprParser{tokens: tokens}
}

// ParseExpression parses a complete expression.
// This is the entry point for expression parsing.
func (p *ExprParser) ParseExpression() (Expression, error) {
	return p.parseOr()
}

// Operator precedence (lowest to highest):
//  1. or
//  2. and
//  3. comparison (==, !=, <, >, <=, >=)
//  4. addition/subtraction (+, -)
//  5. multiplication/division (*, /, %)
//  6. unary (not, -, +)
//  7. postfix (filter |, property ., subscript [])
//  8. primary (literals, variables, parentheses)

// parseOr parses "or" expressions (lowest precedence).
// Example: a or b or c, a || b || c
func (p *ExprParser) parseOr() (Expression, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for tok := p.current(); tok != nil; tok = p.current() {
		if !p.isOr(tok) {
			break
		}
		p.advance()

		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}

		left = NewBinaryOpNode("or", left, right, tok.Line, tok.Col)
	}

	return left, nil
}

// parseAnd parses "and" expressions.
// Example: a and b and c, a && b && c
func (p *ExprParser) parseAnd() (Expression, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for tok := p.current(); tok != nil; tok = p.current() {
		if !p.isAnd(tok) {
			break
		}
		p.advance()

		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}

		left = NewBinaryOpNode("and", left, right, tok.Line, tok.Col)
	}

	return left, nil
}

// parseComparison parses comparison expressions.
// Example: a == b, x > 10, count <= 100
func (p *ExprParser) parseComparison() (Expression, error) {
	left, err := p.parseAddition()
	if err != nil {
		return nil, err
	}

	for tok := p.current(); tok != nil && tok.Type == TokenSymbol; tok = p.current() {
		if !comparisonOps[tok.Value] {
			break
		}
		op := tok.Value
		p.advance()

		right, err := p.parseAddition()
		if err != nil {
			return nil, err
		}

		left = NewBinaryOpNode(op, left, right, tok.Line, tok.Col)
	}

	return left, nil
}

// parseAddition parses addition and subtraction.
// Example: a + b - c
func (p *ExprParser) parseAddition() (Expression, error) {
	left, err := p.parseMultiplication()
	if err != nil {
		return nil, err
	}

	for tok := p.current(); tok != nil && tok.Type == TokenSymbol; tok = p.current() {
		op := tok.Value
		if op != "+" && op != "-" {
			break
		}
		p.advance()

		right, err := p.parseMultiplication()
		if err != nil {
			return nil, err
		}

		left = NewBinaryOpNode(op, left, right, tok.Line, tok.Col)
	}

	return left, nil
}

// parseMultiplication parses multiplication, division, and modulo.
// Example: a * b / c % d
func (p *ExprParser) parseMultiplication() (Expression, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for tok := p.current(); tok != nil && tok.Type == TokenSymbol; tok = p.current() {
		op := tok.Value
		if op != "*" && op != "/" && op != "%" {
			break
		}
		p.advance()

		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}

		left = NewBinaryOpNode(op, left, right, tok.Line, tok.Col)
	}

	return left, nil
}

// parseUnary parses unary expressions.
// Example: not flag, !flag, -value, +value
func (p *ExprParser) parseUnary() (Expression, error) {
	tok := p.current()
	if tok == nil {
		return nil, p.parseErr("unexpected end of expression")
	}

	// Support both "not" keyword and "!" symbol
	if (tok.Type == TokenIdentifier && tok.Value == "not") ||
		(tok.Type == TokenSymbol && tok.Value == "!") {
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return NewUnaryOpNode("not", operand, tok.Line, tok.Col), nil
	}

	if tok.Type == TokenSymbol && (tok.Value == "-" || tok.Value == "+") {
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return NewUnaryOpNode(tok.Value, operand, tok.Line, tok.Col), nil
	}

	return p.parsePostfix()
}

// parsePostfix parses postfix expressions.
// Handles: property access (.), subscript ([]), and filters (|)
// Example: user.name[0]|upper
func (p *ExprParser) parsePostfix() (Expression, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for tok := p.current(); tok != nil && tok.Type == TokenSymbol; tok = p.current() {
		switch tok.Value {
		case ".":
			expr, err = p.parsePropertyAccess(expr)
		case "[":
			expr, err = p.parseSubscript(expr)
		case "|":
			expr, err = p.parseFilter(expr)
		default:
			return expr, nil
		}
		if err != nil {
			return nil, err
		}
	}

	return expr, nil
}

// parsePropertyAccess parses property access: object.property or object.0 (numeric index)
func (p *ExprParser) parsePropertyAccess(object Expression) (Expression, error) {
	dotToken := p.current()
	p.advance() // consume "."

	tok := p.current()
	if tok == nil {
		return nil, p.parseErr("expected property name or numeric index after '.'")
	}

	// Numeric index (e.g., .0, .1) converts to subscript access.
	if tok.Type == TokenNumber {
		indexStr := tok.Value
		p.advance()

		indexValue, err := strconv.ParseInt(indexStr, 10, 0)
		if err != nil {
			// Try as float (unusual for array index).
			floatValue, fErr := strconv.ParseFloat(indexStr, 64)
			if fErr != nil {
				return nil, p.errAtTok(tok, "invalid numeric index: "+indexStr)
			}
			indexExpr := NewLiteralNode(floatValue, dotToken.Line, dotToken.Col)
			return NewSubscriptNode(object, indexExpr, dotToken.Line, dotToken.Col), nil
		}

		indexExpr := NewLiteralNode(int(indexValue), dotToken.Line, dotToken.Col)
		return NewSubscriptNode(object, indexExpr, dotToken.Line, dotToken.Col), nil
	}

	if tok.Type != TokenIdentifier {
		return nil, p.parseErr("expected property name after '.'")
	}

	property := tok.Value
	p.advance()

	return NewPropertyAccessNode(object, property, dotToken.Line, dotToken.Col), nil
}

// parseSubscript parses subscript access: object[index]
func (p *ExprParser) parseSubscript(object Expression) (Expression, error) {
	bracketToken := p.current()
	p.advance() // consume "["

	index, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	tok := p.current()
	if tok == nil || tok.Type != TokenSymbol || tok.Value != "]" {
		return nil, p.parseErr("expected ']' after subscript index")
	}
	p.advance() // consume "]"

	return NewSubscriptNode(object, index, bracketToken.Line, bracketToken.Col), nil
}

// parseFilter parses filter application: expression|filter or expression|filter:arg1,arg2
func (p *ExprParser) parseFilter(expr Expression) (Expression, error) {
	pipeToken := p.current()
	p.advance() // consume "|"

	tok := p.current()
	if tok == nil || tok.Type != TokenIdentifier {
		return nil, p.parseErr("expected filter name after '|'")
	}

	filterName := tok.Value
	p.advance()

	var args []Expression

	// Parse filter arguments starting with ":"
	if p.current() != nil && p.current().Type == TokenSymbol && p.current().Value == ":" {
		p.advance() // consume ":"

		arg, err := p.parseFilterArg()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Additional arguments separated by ","
		for p.current() != nil && p.current().Type == TokenSymbol && p.current().Value == "," {
			p.advance() // consume ","

			arg, err := p.parseFilterArg()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	return NewFilterNode(expr, filterName, args, pipeToken.Line, pipeToken.Col), nil
}

// parseFilterArg parses a single filter argument.
// Filter arguments are simpler — just literals or variables, no complex expressions.
func (p *ExprParser) parseFilterArg() (Expression, error) {
	tok := p.current()
	if tok == nil {
		return nil, p.parseErr("expected filter argument")
	}

	switch tok.Type {
	case TokenString:
		p.advance()
		return NewLiteralNode(tok.Value, tok.Line, tok.Col), nil

	case TokenNumber:
		p.advance()
		num, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, p.errAtTok(tok, "invalid number: "+tok.Value)
		}
		return NewLiteralNode(num, tok.Line, tok.Col), nil

	case TokenIdentifier:
		p.advance()
		switch tok.Value {
		case "true":
			return NewLiteralNode(true, tok.Line, tok.Col), nil
		case "false":
			return NewLiteralNode(false, tok.Line, tok.Col), nil
		default:
			return NewVariableNode(tok.Value, tok.Line, tok.Col), nil
		}

	case TokenError, TokenEOF, TokenText, TokenVarBegin, TokenVarEnd,
		TokenTagBegin, TokenTagEnd, TokenSymbol:
		return nil, p.errAtTok(tok, "expected literal or variable as filter argument")
	}

	return nil, p.errAtTok(tok, "expected literal or variable as filter argument")
}

// parsePrimary parses primary expressions.
// These are the building blocks: literals, variables, and parenthesized expressions.
func (p *ExprParser) parsePrimary() (Expression, error) {
	tok := p.current()
	if tok == nil {
		return nil, p.parseErr("unexpected end of expression")
	}

	switch tok.Type {
	case TokenString:
		p.advance()
		return NewLiteralNode(tok.Value, tok.Line, tok.Col), nil

	case TokenNumber:
		p.advance()
		num, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, p.errAtTok(tok, "invalid number: "+tok.Value)
		}
		return NewLiteralNode(num, tok.Line, tok.Col), nil

	case TokenIdentifier:
		p.advance()
		switch tok.Value {
		case "true":
			return NewLiteralNode(true, tok.Line, tok.Col), nil
		case "false":
			return NewLiteralNode(false, tok.Line, tok.Col), nil
		default:
			return NewVariableNode(tok.Value, tok.Line, tok.Col), nil
		}

	case TokenSymbol:
		if tok.Value == "(" {
			p.advance() // consume "("

			expr, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}

			closeTok := p.current()
			if closeTok == nil || closeTok.Type != TokenSymbol || closeTok.Value != ")" {
				return nil, p.parseErr("expected ')' after expression")
			}
			p.advance() // consume ")"

			return expr, nil
		}

		return nil, p.errAtTok(tok, "unexpected symbol: "+tok.Value)

	case TokenError, TokenEOF, TokenText, TokenVarBegin, TokenVarEnd,
		TokenTagBegin, TokenTagEnd:
		return nil, p.errAtTok(tok, "unexpected token: "+tok.Value)
	}

	return nil, p.errAtTok(tok, "unexpected token: "+tok.Value)
}

// Helper methods

// current returns the current token without consuming it.
func (p *ExprParser) current() *Token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	return p.tokens[p.pos]
}

// advance moves to the next token.
func (p *ExprParser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

// peek returns the token at the given offset without consuming it.
func (p *ExprParser) peek(offset int) *Token {
	pos := p.pos + offset
	if pos >= len(p.tokens) {
		return nil
	}
	return p.tokens[pos]
}

// isOr reports whether tok represents an "or" operator.
func (p *ExprParser) isOr(tok *Token) bool {
	return (tok.Type == TokenIdentifier && tok.Value == "or") ||
		(tok.Type == TokenSymbol && tok.Value == "||")
}

// isAnd reports whether tok represents an "and" operator.
func (p *ExprParser) isAnd(tok *Token) bool {
	return (tok.Type == TokenIdentifier && tok.Value == "and") ||
		(tok.Type == TokenSymbol && tok.Value == "&&")
}

// parseErr creates a parse error at the current token position.
func (p *ExprParser) parseErr(msg string) error {
	tok := p.current()
	if tok != nil {
		return &ParseError{
			Message: msg,
			Line:    tok.Line,
			Col:     tok.Col,
		}
	}
	return &ParseError{Message: msg}
}

// errAtTok creates a parse error at the given token's position.
func (p *ExprParser) errAtTok(tok *Token, msg string) error {
	return &ParseError{
		Message: msg,
		Line:    tok.Line,
		Col:     tok.Col,
	}
}

// ParseError represents a parsing error with source location.
type ParseError struct {
	Message string
	Line    int
	Col     int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, col %d: %s", e.Line, e.Col, e.Message)
}
