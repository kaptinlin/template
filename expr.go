package template

import (
	"fmt"
	"strconv"
)

// ExprParser parses expressions from a token stream.
// It handles operator precedence, filters, property access, etc.
type ExprParser struct {
	tokens []*Token
	pos    int
}

// NewExprParser creates a new expression parser.
func NewExprParser(tokens []*Token) *ExprParser {
	return &ExprParser{
		tokens: tokens,
		pos:    0,
	}
}

// ParseExpression parses a complete expression.
// This is the entry point for expression parsing.
func (p *ExprParser) ParseExpression() (Expression, error) {
	return p.parseOrExpression()
}

// Operator precedence (lowest to highest):
// 1. or
// 2. and
// 3. comparison (==, !=, <, >, <=, >=)
// 4. addition/subtraction (+, -)
// 5. multiplication/division (*, /, %)
// 6. unary (not, -, +)
// 7. postfix (filter |, property ., subscript [])
// 8. primary (literals, variables, parentheses)

// parseOrExpression parses "or" expressions (lowest precedence).
// Example: a or b or c, a || b || c
func (p *ExprParser) parseOrExpression() (Expression, error) {
	left, err := p.parseAndExpression()
	if err != nil {
		return nil, err
	}

	for p.current() != nil {
		tok := p.current()
		// Support both "or" keyword and "||" symbol
		isOr := (tok.Type == TokenIdentifier && tok.Value == "or") ||
			(tok.Type == TokenSymbol && tok.Value == "||")

		if !isOr {
			break
		}

		opToken := tok
		p.advance() // consume "or" or "||"

		right, err := p.parseAndExpression()
		if err != nil {
			return nil, err
		}

		left = NewBinaryOpNode("or", left, right, opToken.Line, opToken.Col)
	}

	return left, nil
}

// parseAndExpression parses "and" expressions.
// Example: a and b and c, a && b && c
func (p *ExprParser) parseAndExpression() (Expression, error) {
	left, err := p.parseComparisonExpression()
	if err != nil {
		return nil, err
	}

	for p.current() != nil {
		tok := p.current()
		// Support both "and" keyword and "&&" symbol
		isAnd := (tok.Type == TokenIdentifier && tok.Value == "and") ||
			(tok.Type == TokenSymbol && tok.Value == "&&")

		if !isAnd {
			break
		}

		opToken := tok
		p.advance() // consume "and" or "&&"

		right, err := p.parseComparisonExpression()
		if err != nil {
			return nil, err
		}

		left = NewBinaryOpNode("and", left, right, opToken.Line, opToken.Col)
	}

	return left, nil
}

// parseComparisonExpression parses comparison expressions.
// Example: a == b, x > 10, count <= 100
func (p *ExprParser) parseComparisonExpression() (Expression, error) {
	left, err := p.parseAdditionExpression()
	if err != nil {
		return nil, err
	}

	// Comparison operators
	for p.current() != nil && p.current().Type == TokenSymbol {
		op := p.current().Value
		if op == "==" || op == "!=" || op == "<" || op == ">" || op == "<=" || op == ">=" {
			opToken := p.current()
			p.advance() // consume operator

			right, err := p.parseAdditionExpression()
			if err != nil {
				return nil, err
			}

			left = NewBinaryOpNode(op, left, right, opToken.Line, opToken.Col)
		} else {
			break
		}
	}

	return left, nil
}

// parseAdditionExpression parses addition and subtraction.
// Example: a + b - c
func (p *ExprParser) parseAdditionExpression() (Expression, error) {
	left, err := p.parseMultiplicationExpression()
	if err != nil {
		return nil, err
	}

	for p.current() != nil && p.current().Type == TokenSymbol {
		op := p.current().Value
		if op == "+" || op == "-" {
			opToken := p.current()
			p.advance() // consume operator

			right, err := p.parseMultiplicationExpression()
			if err != nil {
				return nil, err
			}

			left = NewBinaryOpNode(op, left, right, opToken.Line, opToken.Col)
		} else {
			break
		}
	}

	return left, nil
}

// parseMultiplicationExpression parses multiplication, division, and modulo.
// Example: a * b / c % d
func (p *ExprParser) parseMultiplicationExpression() (Expression, error) {
	left, err := p.parseUnaryExpression()
	if err != nil {
		return nil, err
	}

	for p.current() != nil && p.current().Type == TokenSymbol {
		op := p.current().Value
		if op == "*" || op == "/" || op == "%" {
			opToken := p.current()
			p.advance() // consume operator

			right, err := p.parseUnaryExpression()
			if err != nil {
				return nil, err
			}

			left = NewBinaryOpNode(op, left, right, opToken.Line, opToken.Col)
		} else {
			break
		}
	}

	return left, nil
}

// parseUnaryExpression parses unary expressions.
// Example: not flag, !flag, -value, +value
func (p *ExprParser) parseUnaryExpression() (Expression, error) {
	tok := p.current()
	if tok == nil {
		return nil, p.error("unexpected end of expression")
	}

	// Check for unary operators
	// Support both "not" keyword and "!" symbol
	if (tok.Type == TokenIdentifier && tok.Value == "not") ||
		(tok.Type == TokenSymbol && tok.Value == "!") {
		p.advance() // consume "not" or "!"
		operand, err := p.parseUnaryExpression()
		if err != nil {
			return nil, err
		}
		return NewUnaryOpNode("not", operand, tok.Line, tok.Col), nil
	}

	if tok.Type == TokenSymbol && (tok.Value == "-" || tok.Value == "+") {
		p.advance() // consume operator
		operand, err := p.parseUnaryExpression()
		if err != nil {
			return nil, err
		}
		return NewUnaryOpNode(tok.Value, operand, tok.Line, tok.Col), nil
	}

	// Not a unary operator, parse postfix expression
	return p.parsePostfixExpression()
}

// parsePostfixExpression parses postfix expressions.
// Handles: property access (.), subscript ([]), and filters (|)
// Example: user.name[0]|upper
func (p *ExprParser) parsePostfixExpression() (Expression, error) {
	expr, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.current()
		if tok == nil {
			break
		}

		switch {
		case tok.Type == TokenSymbol && tok.Value == ".":
			// Property access: user.name
			expr, err = p.parsePropertyAccess(expr)
			if err != nil {
				return nil, err
			}

		case tok.Type == TokenSymbol && tok.Value == "[":
			// Subscript: items[0]
			expr, err = p.parseSubscript(expr)
			if err != nil {
				return nil, err
			}

		case tok.Type == TokenSymbol && tok.Value == "|":
			// Filter: name|upper
			expr, err = p.parseFilter(expr)
			if err != nil {
				return nil, err
			}

		default:
			// No more postfix operators
			return expr, nil
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
		return nil, p.error("expected property name or numeric index after '.'")
	}

	// Check if it's a numeric index (e.g., .0, .1)
	// This supports the syntax: user.items.0.name
	if tok.Type == TokenNumber {
		// Convert to subscript access: user.items[0].name
		indexStr := tok.Value
		p.advance() // consume number

		// Parse as integer for array index
		// strconv.ParseInt(s, base, bitSize) - base 10, int type
		indexValue, err := strconv.ParseInt(indexStr, 10, 0)
		if err != nil {
			// If it's not a valid integer, try as float (though unusual for array index)
			floatValue, err := strconv.ParseFloat(indexStr, 64)
			if err != nil {
				return nil, p.errorAt(tok, fmt.Sprintf("invalid numeric index: %s", indexStr))
			}
			// Use float as index (will be converted to int during evaluation)
			indexExpr := NewLiteralNode(floatValue, dotToken.Line, dotToken.Col)
			return NewSubscriptNode(object, indexExpr, dotToken.Line, dotToken.Col), nil
		}

		// Create a literal node for the integer index
		indexExpr := NewLiteralNode(int(indexValue), dotToken.Line, dotToken.Col)
		return NewSubscriptNode(object, indexExpr, dotToken.Line, dotToken.Col), nil
	}

	// Regular property access
	if tok.Type != TokenIdentifier {
		return nil, p.error("expected property name after '.'")
	}

	property := tok.Value
	p.advance() // consume property name

	return NewPropertyAccessNode(object, property, dotToken.Line, dotToken.Col), nil
}

// parseSubscript parses subscript access: object[index]
func (p *ExprParser) parseSubscript(object Expression) (Expression, error) {
	bracketToken := p.current()
	p.advance() // consume "["

	// Parse the index expression
	index, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	// Expect closing bracket
	tok := p.current()
	if tok == nil || tok.Type != TokenSymbol || tok.Value != "]" {
		return nil, p.error("expected ']' after subscript index")
	}
	p.advance() // consume "]"

	return NewSubscriptNode(object, index, bracketToken.Line, bracketToken.Col), nil
}

// parseFilter parses filter application: expression|filter or expression|filter:arg1:arg2
func (p *ExprParser) parseFilter(expr Expression) (Expression, error) {
	pipeToken := p.current()
	p.advance() // consume "|"

	// Parse filter name
	tok := p.current()
	if tok == nil || tok.Type != TokenIdentifier {
		return nil, p.error("expected filter name after '|'")
	}

	filterName := tok.Value
	p.advance() // consume filter name

	// Parse filter arguments (if any)
	var args []Expression

	// Check for filter arguments (must start with ":")
	if p.current() != nil && p.current().Type == TokenSymbol && p.current().Value == ":" {
		p.advance() // consume ":"

		// Parse first argument
		arg, err := p.parseFilterArgument()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Parse additional arguments (separated by ",")
		for p.current() != nil && p.current().Type == TokenSymbol && p.current().Value == "," {
			p.advance() // consume ","

			arg, err := p.parseFilterArgument()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	return NewFilterNode(expr, filterName, args, pipeToken.Line, pipeToken.Col), nil
}

// parseFilterArgument parses a single filter argument.
// Filter arguments are simpler - just literals or variables, no complex expressions.
func (p *ExprParser) parseFilterArgument() (Expression, error) {
	tok := p.current()
	if tok == nil {
		return nil, p.error("expected filter argument")
	}

	switch tok.Type {
	case TokenString:
		p.advance()
		return NewLiteralNode(tok.Value, tok.Line, tok.Col), nil

	case TokenNumber:
		p.advance()
		num, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, p.errorAt(tok, fmt.Sprintf("invalid number: %s", tok.Value))
		}
		return NewLiteralNode(num, tok.Line, tok.Col), nil

	case TokenIdentifier:
		// Could be a variable or a boolean literal
		if tok.Value == "true" {
			p.advance()
			return NewLiteralNode(true, tok.Line, tok.Col), nil
		}
		if tok.Value == "false" {
			p.advance()
			return NewLiteralNode(false, tok.Line, tok.Col), nil
		}
		// It's a variable reference
		p.advance()
		return NewVariableNode(tok.Value, tok.Line, tok.Col), nil

	case TokenError, TokenEOF, TokenText, TokenVarBegin, TokenVarEnd, TokenTagBegin, TokenTagEnd, TokenSymbol:
		return nil, p.errorAt(tok, "expected literal or variable as filter argument")
	}

	return nil, p.errorAt(tok, "expected literal or variable as filter argument")
}

// parsePrimaryExpression parses primary expressions.
// These are the building blocks: literals, variables, and parenthesized expressions.
func (p *ExprParser) parsePrimaryExpression() (Expression, error) {
	tok := p.current()
	if tok == nil {
		return nil, p.error("unexpected end of expression")
	}

	switch tok.Type {
	case TokenString:
		// String literal
		p.advance()
		return NewLiteralNode(tok.Value, tok.Line, tok.Col), nil

	case TokenNumber:
		// Number literal
		p.advance()
		num, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, p.errorAt(tok, fmt.Sprintf("invalid number: %s", tok.Value))
		}
		return NewLiteralNode(num, tok.Line, tok.Col), nil

	case TokenIdentifier:
		// Could be a variable, keyword (true/false), or in
		if tok.Value == "true" {
			p.advance()
			return NewLiteralNode(true, tok.Line, tok.Col), nil
		}
		if tok.Value == "false" {
			p.advance()
			return NewLiteralNode(false, tok.Line, tok.Col), nil
		}

		// Variable reference
		p.advance()
		return NewVariableNode(tok.Value, tok.Line, tok.Col), nil

	case TokenSymbol:
		if tok.Value == "(" {
			// Parenthesized expression
			p.advance() // consume "("

			expr, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}

			// Expect closing parenthesis
			closeTok := p.current()
			if closeTok == nil || closeTok.Type != TokenSymbol || closeTok.Value != ")" {
				return nil, p.error("expected ')' after expression")
			}
			p.advance() // consume ")"

			return expr, nil
		}

		return nil, p.errorAt(tok, fmt.Sprintf("unexpected token: %s", tok.Value))

	case TokenError, TokenEOF, TokenText, TokenVarBegin, TokenVarEnd, TokenTagBegin, TokenTagEnd:
		return nil, p.errorAt(tok, fmt.Sprintf("unexpected token: %s", tok.Value))
	}

	return nil, p.errorAt(tok, fmt.Sprintf("unexpected token: %s", tok.Value))
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

// error creates a parser error with the current position.
func (p *ExprParser) error(msg string) error {
	tok := p.current()
	if tok != nil {
		return &ParseError{
			Message: msg,
			Line:    tok.Line,
			Col:     tok.Col,
		}
	}
	return &ParseError{
		Message: msg,
		Line:    0,
		Col:     0,
	}
}

// errorAt creates a parser error at the given token's position.
func (p *ExprParser) errorAt(tok *Token, msg string) error {
	return &ParseError{
		Message: msg,
		Line:    tok.Line,
		Col:     tok.Col,
	}
}

// ParseError represents a parsing error.
type ParseError struct {
	Message string
	Line    int
	Col     int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, col %d: %s", e.Line, e.Col, e.Message)
}
