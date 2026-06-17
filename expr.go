package template

import (
	"fmt"
	"strconv"
)

// comparisonOps is the set of comparison operators for fast lookup.
var comparisonOps = map[string]struct{}{
	"==": {},
	"!=": {},
	"<":  {},
	">":  {},
	"<=": {},
	">=": {},
}

// exprParser parses expressions from a token stream.
// It handles operator precedence, filters, property access, etc.
type exprParser struct {
	tokens  []*token
	pos     int
	filters *registry
}

// ParseError represents a parsing error with source location.
type ParseError struct {
	Message string
	Line    int
	Col     int
}

// Error returns a human-readable error message with source location.
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, col %d: %s", e.Line, e.Col, e.Message)
}

// newExprParser creates a new expression parser.
func newExprParser(tokens []*token) *exprParser {
	return &exprParser{tokens: tokens}
}

// ParseExpression parses a complete expression.
// This is the entry point for expression parsing.
func (p *exprParser) ParseExpression() (expression, error) {
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
func (p *exprParser) parseOr() (expression, error) {
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

		left = newBinaryOpNode("or", left, right, tok.Line, tok.Col)
	}

	return left, nil
}

// parseAnd parses "and" expressions.
// Example: a and b and c, a && b && c
func (p *exprParser) parseAnd() (expression, error) {
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

		left = newBinaryOpNode("and", left, right, tok.Line, tok.Col)
	}

	return left, nil
}

// parseComparison parses comparison expressions.
// Example: a == b, x > 10, count <= 100
func (p *exprParser) parseComparison() (expression, error) {
	left, err := p.parseAddition()
	if err != nil {
		return nil, err
	}

	for tok := p.current(); tok != nil && tok.Type == tokenSymbol; tok = p.current() {
		if _, ok := comparisonOps[tok.value]; !ok {
			break
		}
		op := tok.value
		p.advance()

		right, err := p.parseAddition()
		if err != nil {
			return nil, err
		}

		left = newBinaryOpNode(op, left, right, tok.Line, tok.Col)
	}

	return left, nil
}

// parseAddition parses addition and subtraction.
// Example: a + b - c
func (p *exprParser) parseAddition() (expression, error) {
	left, err := p.parseMultiplication()
	if err != nil {
		return nil, err
	}

	for tok := p.current(); tok != nil && tok.Type == tokenSymbol; tok = p.current() {
		op := tok.value
		if op != "+" && op != "-" {
			break
		}
		p.advance()

		right, err := p.parseMultiplication()
		if err != nil {
			return nil, err
		}

		left = newBinaryOpNode(op, left, right, tok.Line, tok.Col)
	}

	return left, nil
}

// parseMultiplication parses multiplication, division, and modulo.
// Example: a * b / c % d
func (p *exprParser) parseMultiplication() (expression, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for tok := p.current(); tok != nil && tok.Type == tokenSymbol; tok = p.current() {
		op := tok.value
		if op != "*" && op != "/" && op != "%" {
			break
		}
		p.advance()

		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}

		left = newBinaryOpNode(op, left, right, tok.Line, tok.Col)
	}

	return left, nil
}

// parseUnary parses unary expressions.
// Example: not flag, !flag, -value, +value
func (p *exprParser) parseUnary() (expression, error) {
	tok := p.current()
	if tok == nil {
		return nil, p.parseErr("unexpected end of expression")
	}

	// Support both "not" keyword and "!" symbol
	if (tok.Type == tokenIdentifier && tok.value == "not") ||
		(tok.Type == tokenSymbol && tok.value == "!") {
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return newUnaryOpNode("not", operand, tok.Line, tok.Col), nil
	}

	if tok.Type == tokenSymbol && (tok.value == "-" || tok.value == "+") {
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return newUnaryOpNode(tok.value, operand, tok.Line, tok.Col), nil
	}

	return p.parsePostfix()
}

// parsePostfix parses postfix expressions.
// Handles: property access (.), subscript ([]), and filters (|)
// Example: user.name[0]|upper
func (p *exprParser) parsePostfix() (expression, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for tok := p.current(); tok != nil && tok.Type == tokenSymbol; tok = p.current() {
		switch tok.value {
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
func (p *exprParser) parsePropertyAccess(object expression) (expression, error) {
	dotToken := p.current()
	p.advance() // consume "."

	tok := p.current()
	if tok == nil {
		return nil, p.parseErr("expected property name or numeric index after '.'")
	}

	// Numeric index (e.g., .0, .1) converts to subscript access.
	if tok.Type == tokenNumber {
		indexStr := tok.value
		p.advance()

		indexValue, err := strconv.ParseInt(indexStr, 10, 0)
		if err != nil {
			// Try as float (unusual for array index).
			floatValue, fErr := strconv.ParseFloat(indexStr, 64)
			if fErr != nil {
				return nil, p.errAtTok(tok, "invalid numeric index: "+indexStr)
			}
			indexExpr := newLiteralNode(floatValue, dotToken.Line, dotToken.Col)
			return newSubscriptNode(object, indexExpr, dotToken.Line, dotToken.Col), nil
		}

		indexExpr := newLiteralNode(int(indexValue), dotToken.Line, dotToken.Col)
		return newSubscriptNode(object, indexExpr, dotToken.Line, dotToken.Col), nil
	}

	if tok.Type != tokenIdentifier {
		return nil, p.parseErr("expected property name after '.'")
	}

	property := tok.value
	p.advance()

	return newPropertyAccessNode(object, property, dotToken.Line, dotToken.Col), nil
}

// parseSubscript parses subscript access: object[index]
func (p *exprParser) parseSubscript(object expression) (expression, error) {
	bracketToken := p.current()
	p.advance() // consume "["

	index, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	tok := p.current()
	if tok == nil || tok.Type != tokenSymbol || tok.value != "]" {
		return nil, p.parseErr("expected ']' after subscript index")
	}
	p.advance() // consume "]"

	return newSubscriptNode(object, index, bracketToken.Line, bracketToken.Col), nil
}

// parseFilter parses filter application: expression|filter or expression|filter:arg1,arg2
func (p *exprParser) parseFilter(expr expression) (expression, error) {
	pipeToken := p.current()
	p.advance() // consume "|"

	tok := p.current()
	if tok == nil || tok.Type != tokenIdentifier {
		return nil, p.parseErr("expected filter name after '|'")
	}

	filterName := tok.value
	p.advance()

	var args []expression

	// Parse filter arguments starting with ":"
	if p.current() != nil && p.current().Type == tokenSymbol && p.current().value == ":" {
		p.advance() // consume ":"

		arg, err := p.parseFilterArg()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Additional arguments separated by ","
		for p.current() != nil && p.current().Type == tokenSymbol && p.current().value == "," {
			p.advance() // consume ","

			arg, err := p.parseFilterArg()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	node := newFilterNode(expr, filterName, args, pipeToken.Line, pipeToken.Col)
	if p.filters != nil {
		if fn, ok := p.filters.Filter(filterName); ok {
			return &boundFilterNode{filterNode: node, fn: fn}, nil
		}
	}
	return node, nil
}

// parseFilterArg parses a single filter argument.
// Filter arguments are simpler — just literals or variables, no complex expressions.
func (p *exprParser) parseFilterArg() (expression, error) {
	tok := p.current()
	if tok == nil {
		return nil, p.parseErr("expected filter argument")
	}

	switch tok.Type {
	case tokenString:
		p.advance()
		return newLiteralNode(tok.value, tok.Line, tok.Col), nil

	case tokenNumber:
		p.advance()
		num, err := strconv.ParseFloat(tok.value, 64)
		if err != nil {
			return nil, p.errAtTok(tok, "invalid number: "+tok.value)
		}
		return newLiteralNode(num, tok.Line, tok.Col), nil

	case tokenIdentifier:
		p.advance()
		if tok.value == "true" {
			return newLiteralNode(true, tok.Line, tok.Col), nil
		}
		if tok.value == "false" {
			return newLiteralNode(false, tok.Line, tok.Col), nil
		}
		return newVariableNode(tok.value, tok.Line, tok.Col), nil

	case tokenError, tokenEOF, tokenText, tokenVarBegin, tokenVarEnd,
		tokenTagBegin, tokenTagEnd, tokenSymbol:
		return nil, p.errAtTok(tok, "expected literal or variable as filter argument")
	}

	return nil, p.errAtTok(tok, "expected literal or variable as filter argument")
}

// parsePrimary parses primary expressions.
// These are the building blocks: literals, variables, and parenthesized expressions.
func (p *exprParser) parsePrimary() (expression, error) {
	tok := p.current()
	if tok == nil {
		return nil, p.parseErr("unexpected end of expression")
	}

	switch tok.Type {
	case tokenString:
		p.advance()
		return newLiteralNode(tok.value, tok.Line, tok.Col), nil

	case tokenNumber:
		p.advance()
		num, err := strconv.ParseFloat(tok.value, 64)
		if err != nil {
			return nil, p.errAtTok(tok, "invalid number: "+tok.value)
		}
		return newLiteralNode(num, tok.Line, tok.Col), nil

	case tokenIdentifier:
		p.advance()
		if tok.value == "true" {
			return newLiteralNode(true, tok.Line, tok.Col), nil
		}
		if tok.value == "false" {
			return newLiteralNode(false, tok.Line, tok.Col), nil
		}
		return newVariableNode(tok.value, tok.Line, tok.Col), nil

	case tokenSymbol:
		if tok.value == "(" {
			p.advance() // consume "("

			expr, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}

			closeTok := p.current()
			if closeTok == nil || closeTok.Type != tokenSymbol || closeTok.value != ")" {
				return nil, p.parseErr("expected ')' after expression")
			}
			p.advance() // consume ")"

			return expr, nil
		}

		return nil, p.errAtTok(tok, "unexpected symbol: "+tok.value)

	case tokenError, tokenEOF, tokenText, tokenVarBegin, tokenVarEnd,
		tokenTagBegin, tokenTagEnd:
		return nil, p.errAtTok(tok, "unexpected token: "+tok.value)
	}

	return nil, p.errAtTok(tok, "unexpected token: "+tok.value)
}

// Helper methods

// current returns the current token without consuming it.
func (p *exprParser) current() *token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	return p.tokens[p.pos]
}

// advance moves to the next token.
func (p *exprParser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

// isOr reports whether tok represents an "or" operator.
func (p *exprParser) isOr(tok *token) bool {
	return (tok.Type == tokenIdentifier && tok.value == "or") ||
		(tok.Type == tokenSymbol && tok.value == "||")
}

// isAnd reports whether tok represents an "and" operator.
func (p *exprParser) isAnd(tok *token) bool {
	return (tok.Type == tokenIdentifier && tok.value == "and") ||
		(tok.Type == tokenSymbol && tok.value == "&&")
}

// parseErr creates a parse error at the current token position.
func (p *exprParser) parseErr(msg string) error {
	if tok := p.current(); tok != nil {
		return &ParseError{
			Message: msg,
			Line:    tok.Line,
			Col:     tok.Col,
		}
	}
	if p.pos > 0 && p.pos-1 < len(p.tokens) {
		tok := p.tokens[p.pos-1]
		return &ParseError{
			Message: msg,
			Line:    tok.Line,
			Col:     tok.Col,
		}
	}
	return &ParseError{Message: msg}
}

// errAtTok creates a parse error at the given token's position.
func (p *exprParser) errAtTok(tok *token, msg string) error {
	return &ParseError{
		Message: msg,
		Line:    tok.Line,
		Col:     tok.Col,
	}
}
