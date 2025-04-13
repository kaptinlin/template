package template

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ExpressionNode is the interface for all expression nodes
type ExpressionNode interface {
	Evaluate(ctx Context) (*Value, error)
}

// Grammar represents the parser structure
type Grammar struct {
	tokens  []Token // Token stream from lexer
	pos     int     // Current token position
	current Token   // Current token
}

// NewGrammar creates a new parser
func NewGrammar(tokens []Token) *Grammar {
	g := &Grammar{
		tokens: tokens,
	}
	if len(tokens) > 0 {
		g.current = tokens[0]
	}
	return g
}

// Structure definitions for various expression nodes
type BinaryExpressionNode struct {
	Left     ExpressionNode
	Right    ExpressionNode
	Operator string
}

type UnaryExpressionNode struct {
	Operator string
	Right    ExpressionNode
}

type NumberLiteralNode struct {
	Value float64
}

type StringLiteralNode struct {
	Value string
}

type BooleanLiteralNode struct {
	Value bool
}

type VariableNode struct {
	Name string
}

// Value defines the value type
type Value struct {
	Type  ValueType
	Int   int64
	Float float64
	Str   string
	Bool  bool
	Slice interface{}
	Map   map[string]interface{}
}

type ValueType int

const (
	TypeInt ValueType = iota
	TypeFloat
	TypeString
	TypeBool
	TypeSlice
	TypeMap
)

// FilterExpressionNode defines filter expression node
type FilterExpressionNode struct {
	Expression ExpressionNode // Expression to be filtered
	Filter     string         // Filter name and parameters
}

// Parse starts parsing
func (g *Grammar) Parse() (ExpressionNode, error) {
	return g.parseExpression()
}

// parseExpression is the entry point for expression parsing
func (g *Grammar) parseExpression() (ExpressionNode, error) {
	return g.parseLogicalOr()
}

// parseLogicalOr parses logical OR expressions
func (g *Grammar) parseLogicalOr() (ExpressionNode, error) {
	left, err := g.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for g.current.Typ == TokenOperator && g.current.Val == "||" {
		operator := g.current.Val
		g.advance() // Consume || operator

		right, err := g.parseLogicalAnd()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpressionNode{
			Left:     left,
			Right:    right,
			Operator: operator,
		}
	}

	return left, nil
}

// parseLogicalAnd parses logical AND expressions
func (g *Grammar) parseLogicalAnd() (ExpressionNode, error) {
	left, err := g.parseComparison()
	if err != nil {
		return nil, err
	}

	for g.current.Typ == TokenOperator && g.current.Val == "&&" {
		operator := g.current.Val
		g.advance()

		right, err := g.parseComparison()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpressionNode{
			Left:     left,
			Right:    right,
			Operator: operator,
		}
	}

	return left, nil
}

// parseComparison parses comparison expressions
func (g *Grammar) parseComparison() (ExpressionNode, error) {
	left, err := g.parseAdditive()
	if err != nil {
		return nil, err
	}

	for g.current.Typ == TokenOperator && isComparisonOperator(g.current.Val) {
		operator := g.current.Val
		g.advance()

		right, err := g.parseAdditive()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpressionNode{
			Left:     left,
			Right:    right,
			Operator: operator,
		}
	}

	return left, nil
}

// parseAdditive parses addition and subtraction expressions
func (g *Grammar) parseAdditive() (ExpressionNode, error) {
	left, err := g.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for g.current.Typ == TokenArithOp && (g.current.Val == "+" || g.current.Val == "-") {
		operator := g.current.Val
		g.advance()

		right, err := g.parseMultiplicative()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpressionNode{
			Left:     left,
			Right:    right,
			Operator: operator,
		}
	}

	return left, nil
}

// parseMultiplicative parses multiplication, division and modulo expressions
func (g *Grammar) parseMultiplicative() (ExpressionNode, error) {
	left, err := g.parseUnary()
	if err != nil {
		return nil, err
	}

	for g.current.Typ == TokenArithOp && (g.current.Val == "*" || g.current.Val == "/" || g.current.Val == "%") {
		operator := g.current.Val
		g.advance()

		right, err := g.parseUnary()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpressionNode{
			Left:     left,
			Right:    right,
			Operator: operator,
		}
	}

	return left, nil
}

// parseUnary parses unary expressions
func (g *Grammar) parseUnary() (ExpressionNode, error) {
	if (g.current.Typ == TokenNot) || (g.current.Typ == TokenArithOp && g.current.Val == "-") {
		operator := g.current.Val
		g.advance()

		right, err := g.parseUnary()
		if err != nil {
			return nil, err
		}

		return &UnaryExpressionNode{
			Operator: operator,
			Right:    right,
		}, nil
	}

	return g.parsePrimary()
}

// parsePrimary parses primary expressions
func (g *Grammar) parsePrimary() (ExpressionNode, error) {
	expr, err := g.parseBasicPrimary()
	if err != nil {
		return nil, err
	}

	// Check for filters
	for g.current.Typ == TokenPipe {
		g.advance() // Consume pipe symbol

		// Expect filter name as next token
		if g.current.Typ != TokenFilter {
			return nil, fmt.Errorf("%w: got %v", ErrExpectedFilterName, g.current)
		}

		filterName := g.current.Val
		g.advance() // Consume filter name

		expr = &FilterExpressionNode{
			Expression: expr,
			Filter:     filterName,
		}
	}

	return expr, nil
}

// parseBasicPrimary parses basic primary expressions
func (g *Grammar) parseBasicPrimary() (ExpressionNode, error) {
	switch g.current.Typ {
	case TokenNumber:
		value, err := strconv.ParseFloat(g.current.Val, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidNumber, g.current.Val)
		}
		g.advance()
		return &NumberLiteralNode{Value: value}, nil

	case TokenString:
		value := g.current.Val
		g.advance()
		return &StringLiteralNode{Value: value}, nil

	case TokenBool:
		value := g.current.Val == "true"
		g.advance()
		return &BooleanLiteralNode{Value: value}, nil

	case TokenIdentifier:
		name := g.current.Val
		g.advance()
		return &VariableNode{Name: name}, nil

	case TokenLParen:
		g.advance() // Consume left parenthesis
		expr, err := g.parseExpression()
		if err != nil {
			return nil, err
		}

		if g.current.Typ != TokenRParen {
			return nil, fmt.Errorf("%w: got %v", ErrExpectedRParen, g.current.Val)
		}
		g.advance() // Consume right parenthesis
		return expr, nil

	case TokenOperator:
		return nil, fmt.Errorf("%w: operator %v", ErrUnexpectedToken, g.current.Val)
	case TokenArithOp:
		return nil, fmt.Errorf("%w: arithmetic operator %v", ErrUnexpectedToken, g.current.Val)
	case TokenNot:
		return nil, fmt.Errorf("%w: not operator %v", ErrUnexpectedToken, g.current.Val)
	case TokenRParen:
		return nil, fmt.Errorf("%w: right parenthesis %v", ErrUnexpectedToken, g.current.Val)
	case TokenPipe:
		return nil, fmt.Errorf("%w: pipe operator %v", ErrUnexpectedToken, g.current.Val)
	case TokenFilter:
		return nil, fmt.Errorf("%w: filter %v", ErrUnexpectedToken, g.current.Val)
	case TokenEOF:
		return nil, fmt.Errorf("%w: end of file", ErrUnexpectedToken)
	case TokenDot:
		return nil, fmt.Errorf("%w: unexpected dot operator", ErrUnexpectedToken)

	default:
		return nil, fmt.Errorf("%w: %v", ErrUnexpectedToken, g.current)
	}
}

// advance moves to the next token
func (g *Grammar) advance() {
	g.pos++
	if g.pos < len(g.tokens) {
		g.current = g.tokens[g.pos]
	}
}

// isComparisonOperator checks if the operator is a comparison operator
func isComparisonOperator(op string) bool {
	switch op {
	case "==", "!=", "<", ">", "<=", ">=":
		return true
	default:
		return false
	}
}

// Evaluate method implementation for FilterExpressionNode
func (n *FilterExpressionNode) Evaluate(ctx Context) (*Value, error) {
	// First evaluate the base expression
	val, err := n.Expression.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	filters := parseFilters(n.Filter)
	// Apply filter directly
	result, err := ApplyFilters(val.toInterface(), filters, ctx)
	if err != nil {
		return nil, err
	}

	return NewValue(result)
}

// Helper methods for Value
func (v *Value) toInterface() interface{} {
	switch v.Type {
	case TypeInt:
		return v.Int
	case TypeFloat:
		return v.Float
	case TypeString:
		return v.Str
	case TypeBool:
		return v.Bool
	case TypeSlice:
		return v.Slice
	case TypeMap:
		return v.Map
	default:
		return nil
	}
}

// NewValue creates a new Value
func NewValue(v interface{}) (*Value, error) {
	switch val := v.(type) {
	case int:
		return &Value{Type: TypeInt, Int: int64(val)}, nil
	case int64:
		return &Value{Type: TypeInt, Int: val}, nil
	case float64:
		return &Value{Type: TypeFloat, Float: val}, nil
	case string:
		return &Value{Type: TypeString, Str: val}, nil
	case bool:
		return &Value{Type: TypeBool, Bool: val}, nil
	case []interface{}, []string, []int, []float64, []bool:
		return &Value{Type: TypeSlice, Slice: val}, nil
	case map[string]interface{}:
		return &Value{Type: TypeMap, Map: val}, nil
	default:
		return nil, fmt.Errorf("%w: %T", ErrUnsupportedType, v)
	}
}

// Evaluate method implementation for BinaryExpressionNode
func (n *BinaryExpressionNode) Evaluate(ctx Context) (*Value, error) {
	left, err := n.Left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	right, err := n.Right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	// Execute operation based on operator
	switch n.Operator {
	case "+":
		return left.Add(right)
	case "-":
		return left.Subtract(right)
	case "*":
		return left.Multiply(right)
	case "/":
		return left.Divide(right)
	case "&&":
		return left.And(right)
	case "||":
		return left.Or(right)
	case "==":
		return left.Equal(right)
	case "!=":
		return left.NotEqual(right)
	case "<":
		return left.LessThan(right)
	case ">":
		return left.GreaterThan(right)
	case "<=":
		return left.LessEqual(right)
	case ">=":
		return left.GreaterEqual(right)
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedOperator, n.Operator)
}

// Evaluate method implementation for UnaryExpressionNode
func (n *UnaryExpressionNode) Evaluate(ctx Context) (*Value, error) {
	right, err := n.Right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if n.Operator == "!" {
		rightBool, err := right.toBool()
		if err != nil {
			return nil, err
		}
		return NewValue(!rightBool)
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedUnaryOp, n.Operator)
}

// Evaluate method implementation for NumberLiteralNode
func (n *NumberLiteralNode) Evaluate(ctx Context) (*Value, error) {
	return NewValue(n.Value)
}

// Evaluate method implementation for StringLiteralNode
func (n *StringLiteralNode) Evaluate(ctx Context) (*Value, error) {
	return NewValue(n.Value)
}

// Evaluate method implementation for BooleanLiteralNode
func (n *BooleanLiteralNode) Evaluate(ctx Context) (*Value, error) {
	return NewValue(n.Value)
}

// Evaluate method implementation for VariableNode
func (n *VariableNode) Evaluate(ctx Context) (*Value, error) {
	parts := strings.Split(n.Name, ".")
	var current interface{}
	var ok bool

	// Get root variable
	if current, ok = ctx[parts[0]]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrUndefinedVariable, parts[0])
	}

	// Traverse property path
	for _, part := range parts[1:] {
		switch v := current.(type) {
		case []interface{}, []string, []int, []float64, []bool:
			current = part
		case map[string]interface{}:
			if current, ok = v[part]; !ok {
				return nil, fmt.Errorf("%w: %s in %s", ErrUndefinedProperty, part, n.Name)
			}
		case struct{}, *struct{}:
			// Use reflection to get struct field
			val := reflect.ValueOf(current)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			if val.Kind() != reflect.Struct {
				return nil, fmt.Errorf("%w: %s", ErrNonStructProperty, part)
			}
			field := val.FieldByName(part)
			if !field.IsValid() {
				return nil, fmt.Errorf("%w: %s in struct %s", ErrUndefinedProperty, part, n.Name)
			}
			current = field.Interface()
		default:
			return nil, fmt.Errorf("%w: %T", ErrCannotAccessProperty, current)
		}
	}

	return NewValue(current)
}

// Add implements addition operation
func (v *Value) Add(right *Value) (*Value, error) {
	switch v.Type {
	case TypeInt:
		switch right.Type {
		case TypeInt:
			return NewValue(v.Int + right.Int)
		case TypeFloat:
			return NewValue(float64(v.Int) + right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		}
	case TypeFloat:
		switch right.Type {
		case TypeInt:
			return NewValue(v.Float + float64(right.Int))
		case TypeFloat:
			return NewValue(v.Float + right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		}
	case TypeString:
		switch right.Type {
		case TypeString:
			return NewValue(v.Str + right.Str)
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		}
	case TypeBool:
		switch right.Type {
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		}
	case TypeSlice:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
	case TypeMap:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
	}
	return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
}

// Subtract implements subtraction operation
func (v *Value) Subtract(right *Value) (*Value, error) {
	switch v.Type {
	case TypeInt:
		switch right.Type {
		case TypeInt:
			return NewValue(v.Int - right.Int)
		case TypeFloat:
			return NewValue(float64(v.Int) - right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		}
	case TypeFloat:
		switch right.Type {
		case TypeInt:
			return NewValue(v.Float - float64(right.Int))
		case TypeFloat:
			return NewValue(v.Float - right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		}
	case TypeString:
		switch right.Type {
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		}
	case TypeBool:
		switch right.Type {
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
		}
	case TypeSlice:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
	case TypeMap:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
	}
	return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
}

// Multiply implements multiplication operation
func (v *Value) Multiply(right *Value) (*Value, error) {
	switch v.Type {
	case TypeInt:
		switch right.Type {
		case TypeInt:
			return NewValue(v.Int * right.Int)
		case TypeFloat:
			return NewValue(float64(v.Int) * right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		}
	case TypeFloat:
		switch right.Type {
		case TypeInt:
			return NewValue(v.Float * float64(right.Int))
		case TypeFloat:
			return NewValue(v.Float * right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		}
	case TypeString:
		switch right.Type {
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		}
	case TypeBool:
		switch right.Type {
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
		}
	case TypeSlice:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
	case TypeMap:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
	}
	return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
}

// Divide implements division operation
func (v *Value) Divide(right *Value) (*Value, error) {
	switch v.Type {
	case TypeInt:
		switch right.Type {
		case TypeInt:
			if right.Int == 0 {
				return nil, ErrDivisionByZero
			}
			return NewValue(float64(v.Int) / float64(right.Int))
		case TypeFloat:
			if right.Float == 0 {
				return nil, ErrDivisionByZero
			}
			return NewValue(float64(v.Int) / right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		}
	case TypeFloat:
		switch right.Type {
		case TypeInt:
			if right.Int == 0 {
				return nil, ErrDivisionByZero
			}
			return NewValue(v.Float / float64(right.Int))
		case TypeFloat:
			if right.Float == 0 {
				return nil, ErrDivisionByZero
			}
			return NewValue(v.Float / right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		}
	case TypeString:
		switch right.Type {
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		}
	case TypeBool:
		switch right.Type {
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
		}
	case TypeSlice:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
	case TypeMap:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
	}
	return nil, fmt.Errorf("%w: %v and %v", ErrCannotDivideTypes, v.Type, right.Type)
}

// And implements logical AND operation
func (v *Value) And(right *Value) (*Value, error) {
	leftBool, err := v.toBool()
	if err != nil {
		return nil, err
	}
	rightBool, err := right.toBool()
	if err != nil {
		return nil, err
	}
	return NewValue(leftBool && rightBool)
}

// Or implements logical OR operation
func (v *Value) Or(right *Value) (*Value, error) {
	leftBool, err := v.toBool()
	if err != nil {
		return nil, err
	}
	rightBool, err := right.toBool()
	if err != nil {
		return nil, err
	}
	return NewValue(leftBool || rightBool)
}

// toBool converts value to boolean
func (v *Value) toBool() (bool, error) {
	switch v.Type {
	case TypeBool:
		return v.Bool, nil
	case TypeString:
		return v.Str != "", nil
	case TypeInt:
		return v.Int != 0, nil
	case TypeFloat:
		return v.Float != 0, nil
	case TypeSlice:
		val := reflect.ValueOf(v.Slice)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			return val.Len() > 0, nil
		}
		return false, nil

	case TypeMap:
		return len(v.Map) > 0, nil
	default:
		return false, fmt.Errorf("%w: %v", ErrCannotConvertToBool, v.Type)
	}
}

// Equal implements equality comparison
func (v *Value) Equal(right *Value) (*Value, error) {
	switch v.Type {
	case TypeInt:
		switch right.Type {
		case TypeInt:
			return NewValue(v.Int == right.Int)
		case TypeFloat:
			return NewValue(float64(v.Int) == right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		}
	case TypeFloat:
		switch right.Type {
		case TypeInt:
			return NewValue(v.Float == float64(right.Int))
		case TypeFloat:
			return NewValue(v.Float == right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		}
	case TypeString:
		switch right.Type {
		case TypeString:
			return NewValue(v.Str == right.Str)
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		}
	case TypeBool:
		switch right.Type {
		case TypeBool:
			return NewValue(v.Bool == right.Bool)
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		}
	case TypeSlice:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
	case TypeMap:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
	}
	return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
}

// NotEqual implements inequality comparison
func (v *Value) NotEqual(right *Value) (*Value, error) {
	eq, err := v.Equal(right)
	if err != nil {
		return nil, err
	}
	return NewValue(!eq.Bool)
}

// LessThan implements less than comparison
func (v *Value) LessThan(right *Value) (*Value, error) {
	switch v.Type {
	case TypeInt:
		switch right.Type {
		case TypeInt:
			return NewValue(v.Int < right.Int)
		case TypeFloat:
			return NewValue(float64(v.Int) < right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		}
	case TypeFloat:
		switch right.Type {
		case TypeInt:
			return NewValue(v.Float < float64(right.Int))
		case TypeFloat:
			return NewValue(v.Float < right.Float)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		}
	case TypeString:
		switch right.Type {
		case TypeString:
			return NewValue(v.Str < right.Str)
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		}
	case TypeBool:
		switch right.Type {
		case TypeBool:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeInt:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeFloat:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeString:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeSlice:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		case TypeMap:
			return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
		}
	case TypeSlice:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
	case TypeMap:
		return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
	}
	return nil, fmt.Errorf("%w: %v and %v", ErrCannotCompareTypes, v.Type, right.Type)
}

// GreaterThan implements greater than comparison
func (v *Value) GreaterThan(right *Value) (*Value, error) {
	return right.LessThan(v)
}

// LessEqual implements less than or equal comparison
func (v *Value) LessEqual(right *Value) (*Value, error) {
	gt, err := v.GreaterThan(right)
	if err != nil {
		return nil, err
	}
	return NewValue(!gt.Bool)
}

// GreaterEqual implements greater than or equal comparison
func (v *Value) GreaterEqual(right *Value) (*Value, error) {
	lt, err := v.LessThan(right)
	if err != nil {
		return nil, err
	}
	return NewValue(!lt.Bool)
}
