package template

import (
	"cmp"
	"fmt"
	"math"
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

// BinaryExpressionNode represents a binary expression in the AST.
type BinaryExpressionNode struct {
	Left     ExpressionNode
	Right    ExpressionNode
	Operator string
}

// UnaryExpressionNode represents a unary expression in the AST.
type UnaryExpressionNode struct {
	Operator string
	Right    ExpressionNode
}

// NumberLiteralNode represents a number literal in the AST.
type NumberLiteralNode struct {
	Value float64
}

// StringLiteralNode represents a string literal in the AST.
type StringLiteralNode struct {
	Value string
}

// BooleanLiteralNode represents a boolean literal in the AST.
type BooleanLiteralNode struct {
	Value bool
}

// NilLiteralNode represents a nil/null/none literal in the AST.
type NilLiteralNode struct{}

// VariableNode represents a variable reference in the AST.
type VariableNode struct {
	Name string
}

// Value defines the value type
type Value struct {
	Type   ValueType
	Int    int64
	Float  float64
	Str    string
	Bool   bool
	Slice  interface{}
	Map    interface{}
	Struct interface{}
}

// ValueType represents the type of a value.
type ValueType int

const (
	// TypeInt represents an integer value type
	TypeInt ValueType = iota
	// TypeFloat represents a floating-point value type
	TypeFloat
	// TypeString represents a string value type
	TypeString
	// TypeBool represents a boolean value type
	TypeBool
	// TypeSlice represents a slice value type
	TypeSlice
	// TypeMap represents a map value type
	TypeMap
	// TypeNil represents a nil value type
	TypeNil
	// TypeStruct represents a struct value type
	TypeStruct
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

// parseLogicalOr parses logical OR expressions (supports both || and or)
func (g *Grammar) parseLogicalOr() (ExpressionNode, error) {
	left, err := g.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for g.current.Typ == TokenOperator && (g.current.Val == "||" || g.current.Val == "or") {
		operator := g.current.Val
		g.advance() // Consume || or or operator

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

// parseLogicalAnd parses logical AND expressions (supports both && and and)
func (g *Grammar) parseLogicalAnd() (ExpressionNode, error) {
	left, err := g.parseNot()
	if err != nil {
		return nil, err
	}

	for g.current.Typ == TokenOperator && (g.current.Val == "&&" || g.current.Val == "and") {
		operator := g.current.Val
		g.advance()

		right, err := g.parseNot()
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

// parseNot parses the NOT operator
func (g *Grammar) parseNot() (ExpressionNode, error) {
	if g.current.Typ == TokenOperator && g.current.Val == "not" {
		g.advance()
		right, err := g.parseNot() // Right-associative for `not not x`
		if err != nil {
			return nil, err
		}
		return &UnaryExpressionNode{
			Operator: "not",
			Right:    right,
		}, nil
	}
	return g.parseIn()
}

// parseIn parses membership operators (in, not in)
func (g *Grammar) parseIn() (ExpressionNode, error) {
	left, err := g.parseComparison()
	if err != nil {
		return nil, err
	}

	if g.current.Typ == TokenOperator &&
		(g.current.Val == "in" || g.current.Val == "not in") {
		operator := g.current.Val
		g.advance()
		right, err := g.parseComparison()
		if err != nil {
			return nil, err
		}
		return &BinaryExpressionNode{
			Left:     left,
			Right:    right,
			Operator: operator,
		}, nil
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
		value := g.current.Val == "true" || g.current.Val == "True"
		g.advance()
		return &BooleanLiteralNode{Value: value}, nil

	case TokenIdentifier:
		name := g.current.Val
		// Check for null/Null/none/None literals
		if name == "null" || name == "Null" || name == "none" || name == "None" {
			g.advance()
			return &NilLiteralNode{}, nil
		}
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
	case TypeNil:
		return nil
	case TypeStruct:
		return v.Struct
	default:
		return nil
	}
}

// NewValue creates a new Value from an interface{}.
func NewValue(v interface{}) (*Value, error) {
	// Handle nil interface{} case
	if v == nil {
		return &Value{Type: TypeNil}, nil
	}

	rv := reflect.ValueOf(v)

	// Handle pointers by dereferencing until we get a non-pointer value
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return &Value{Type: TypeNil}, nil
		}
		rv = rv.Elem()
	}

	// Now rv contains the final dereferenced value
	kind := rv.Kind()

	// Use reflect to check the underlying type to handle alias types
	//nolint:exhaustive // Only handle types supported by the template engine
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Value{Type: TypeInt, Int: rv.Int()}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal := rv.Uint()
		// Check for overflow before converting uint64 to int64
		if uintVal > math.MaxInt64 {
			return nil, fmt.Errorf("%w: %d", ErrIntegerOverflow, uintVal)
		}
		return &Value{Type: TypeInt, Int: int64(uintVal)}, nil
	case reflect.Float32, reflect.Float64:
		return &Value{Type: TypeFloat, Float: rv.Float()}, nil
	case reflect.String:
		return &Value{Type: TypeString, Str: rv.String()}, nil
	case reflect.Bool:
		return &Value{Type: TypeBool, Bool: rv.Bool()}, nil
	case reflect.Slice, reflect.Array:
		return &Value{Type: TypeSlice, Slice: rv.Interface()}, nil
	case reflect.Map:
		if rv.Type().Key().Kind() == reflect.String {
			if m, ok := rv.Interface().(map[string]interface{}); ok {
				return &Value{Type: TypeMap, Map: m}, nil
			}

			result := make(map[string]interface{})
			for _, k := range rv.MapKeys() {
				result[k.String()] = rv.MapIndex(k).Interface()
			}
			return &Value{Type: TypeMap, Map: result}, nil
		}
		return nil, fmt.Errorf("%w: map with non-string key type %T", ErrUnsupportedType, rv.Interface())
	case reflect.Struct:
		return &Value{Type: TypeStruct, Struct: rv.Interface()}, nil
	default:
		return nil, fmt.Errorf("%w: %T", ErrUnsupportedType, rv.Interface())
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
	case "&&", "and":
		// Short-circuit AND: if left is falsy, return false without evaluating right
		return NewValue(isTruthy(left) && isTruthy(right))
	case "||", "or":
		// Short-circuit OR: if left is truthy, return true without re-evaluating right
		return NewValue(isTruthy(left) || isTruthy(right))
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
	case "in":
		return left.In(right)
	case "not in":
		result, err := left.In(right)
		if err != nil {
			return nil, err
		}
		return NewValue(!result.Bool)
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedOperator, n.Operator)
}

// Evaluate method implementation for UnaryExpressionNode
func (n *UnaryExpressionNode) Evaluate(ctx Context) (*Value, error) {
	right, err := n.Right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if n.Operator == "!" || n.Operator == "not" {
		return NewValue(!isTruthy(right))
	}

	if n.Operator == "-" {
		// Numeric negation
		switch right.Type {
		case TypeInt:
			return NewValue(-right.Int)
		case TypeFloat:
			return NewValue(-right.Float)
		case TypeString, TypeBool, TypeSlice, TypeMap, TypeNil, TypeStruct:
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedUnaryOp, n.Operator)
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedUnaryOp, n.Operator)
}

// Evaluate method implementation for NumberLiteralNode
func (n *NumberLiteralNode) Evaluate(_ Context) (*Value, error) {
	return NewValue(n.Value)
}

// Evaluate method implementation for StringLiteralNode
func (n *StringLiteralNode) Evaluate(_ Context) (*Value, error) {
	return NewValue(n.Value)
}

// Evaluate method implementation for BooleanLiteralNode
func (n *BooleanLiteralNode) Evaluate(_ Context) (*Value, error) {
	return NewValue(n.Value)
}

// Evaluate method implementation for NilLiteralNode
func (n *NilLiteralNode) Evaluate(_ Context) (*Value, error) {
	return &Value{Type: TypeNil}, nil
}

// Evaluate method implementation for VariableNode
func (n *VariableNode) Evaluate(ctx Context) (*Value, error) {
	value, err := resolveVariable(n.Name, ctx)
	if err != nil {
		// Instead of returning an error, return the original variable placeholder.
		return &Value{Type: TypeString, Str: n.Name}, err
	}

	return NewValue(value)
}

// isNumeric checks if the value is a numeric type (int or float)
func (v *Value) isNumeric() bool {
	return v.Type == TypeInt || v.Type == TypeFloat
}

// toFloat converts the value to float64 for numeric operations
func (v *Value) toFloat() float64 {
	if v.Type == TypeInt {
		return float64(v.Int)
	}
	return v.Float
}

// addNumeric performs addition on two numeric values
func (v *Value) addNumeric(right *Value) (*Value, error) {
	// Return int if both are integers, otherwise float
	if v.Type == TypeInt && right.Type == TypeInt {
		return NewValue(v.Int + right.Int)
	}
	return NewValue(v.toFloat() + right.toFloat())
}

// Add implements addition operation
func (v *Value) Add(right *Value) (*Value, error) {
	// Handle numeric types (int + int, int + float, float + int, float + float)
	if v.isNumeric() && right.isNumeric() {
		return v.addNumeric(right)
	}

	// Handle string concatenation
	if v.Type == TypeString && right.Type == TypeString {
		return NewValue(v.Str + right.Str)
	}

	return nil, fmt.Errorf("%w: %v and %v", ErrCannotAddTypes, v.Type, right.Type)
}

// subtractNumeric performs subtraction on two numeric values
func (v *Value) subtractNumeric(right *Value) (*Value, error) {
	// Return int if both are integers, otherwise float
	if v.Type == TypeInt && right.Type == TypeInt {
		return NewValue(v.Int - right.Int)
	}
	return NewValue(v.toFloat() - right.toFloat())
}

// Subtract implements subtraction operation
func (v *Value) Subtract(right *Value) (*Value, error) {
	// Only numeric types support subtraction
	if v.isNumeric() && right.isNumeric() {
		return v.subtractNumeric(right)
	}

	return nil, fmt.Errorf("%w: %v and %v", ErrCannotSubtractTypes, v.Type, right.Type)
}

// multiplyNumeric performs multiplication on two numeric values
func (v *Value) multiplyNumeric(right *Value) (*Value, error) {
	// Return int if both are integers, otherwise float
	if v.Type == TypeInt && right.Type == TypeInt {
		return NewValue(v.Int * right.Int)
	}
	return NewValue(v.toFloat() * right.toFloat())
}

// Multiply implements multiplication operation
func (v *Value) Multiply(right *Value) (*Value, error) {
	// Only numeric types support multiplication
	if v.isNumeric() && right.isNumeric() {
		return v.multiplyNumeric(right)
	}

	return nil, fmt.Errorf("%w: %v and %v", ErrCannotMultiplyTypes, v.Type, right.Type)
}

// divideNumeric performs division on two numeric values
func (v *Value) divideNumeric(right *Value) (*Value, error) {
	// Check for division by zero
	rightFloat := right.toFloat()
	if rightFloat == 0 {
		return nil, ErrDivisionByZero
	}

	// Division always returns float for consistency
	return NewValue(v.toFloat() / rightFloat)
}

// Divide implements division operation
func (v *Value) Divide(right *Value) (*Value, error) {
	// Only numeric types support division
	if v.isNumeric() && right.isNumeric() {
		return v.divideNumeric(right)
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
		val := reflect.ValueOf(v.Map)
		if val.Kind() == reflect.Map {
			return val.Len() > 0, nil
		}
		return false, nil
	case TypeNil:
		return false, nil
	case TypeStruct:
		if v.Struct == nil {
			return false, nil
		}

		val := reflect.ValueOf(v.Struct)

		isZero := true
		for i := 0; i < val.NumField(); i++ {
			if !val.Field(i).IsZero() {
				isZero = false
				break
			}
		}
		return !isZero, nil
	default:
		return false, fmt.Errorf("%w: %v", ErrCannotConvertToBool, v.Type)
	}
}

// compareNumeric compares two numeric values.
// Returns:
// - 0 if v == right
// - -1 if v < right
// - 1 if v > right
// - true if comparison is valid (both are numeric)
func (v *Value) compareNumeric(right *Value) (int, bool) {
	if v.Type == TypeInt && right.Type == TypeInt {
		return cmp.Compare(v.Int, right.Int), true
	}

	var vf, rf float64

	//nolint:exhaustive // Only numeric types are relevant here
	switch v.Type {
	case TypeInt:
		vf = float64(v.Int)
	case TypeFloat:
		vf = v.Float
	default:
		return 0, false
	}

	//nolint:exhaustive // Only numeric types are relevant here
	switch right.Type {
	case TypeInt:
		rf = float64(right.Int)
	case TypeFloat:
		rf = right.Float
	default:
		return 0, false
	}

	return cmp.Compare(vf, rf), true
}

// Equal implements equality comparison
func (v *Value) Equal(right *Value) (*Value, error) {
	if v.isNumeric() && right.isNumeric() {
		cmp, _ := v.compareNumeric(right)
		return NewValue(cmp == 0)
	}

	if v.Type == TypeString && right.Type == TypeString {
		return NewValue(v.Str == right.Str)
	}

	if v.Type == TypeBool && right.Type == TypeBool {
		return NewValue(v.Bool == right.Bool)
	}

	if v.Type == TypeNil {
		return NewValue(right.Type == TypeNil)
	}

	if right.Type == TypeNil {
		return NewValue(false)
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
	if v.isNumeric() && right.isNumeric() {
		cmp, _ := v.compareNumeric(right)
		return NewValue(cmp < 0)
	}

	if v.Type == TypeString && right.Type == TypeString {
		return NewValue(v.Str < right.Str)
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

// isTruthy determines if a value is truthy according to Django/template semantics
func isTruthy(v *Value) bool {
	if v == nil {
		return false
	}
	switch v.Type {
	case TypeBool:
		return v.Bool
	case TypeInt:
		return v.Int != 0
	case TypeFloat:
		return v.Float != 0.0
	case TypeString:
		return v.Str != ""
	case TypeSlice:
		if v.Slice == nil {
			return false
		}
		rv := reflect.ValueOf(v.Slice)
		return rv.Len() > 0
	case TypeMap:
		if v.Map == nil {
			return false
		}
		rv := reflect.ValueOf(v.Map)
		return rv.Len() > 0
	case TypeNil:
		return false
	case TypeStruct:
		// Struct types are truthy if not nil
		return true
	}
	return false // unreachable, but required for exhaustive switch
}

// In implements membership test operator
func (v *Value) In(haystack *Value) (*Value, error) {
	// String in string (substring check)
	if v.Type == TypeString && haystack.Type == TypeString {
		return NewValue(strings.Contains(haystack.Str, v.Str))
	}

	// Item in slice
	if haystack.Type == TypeSlice {
		rv := reflect.ValueOf(haystack.Slice)
		for i := 0; i < rv.Len(); i++ {
			item := rv.Index(i).Interface()
			itemVal, err := NewValue(item)
			if err != nil {
				continue
			}
			eq, err := v.Equal(itemVal)
			if err == nil && eq.Bool {
				return NewValue(true)
			}
		}
		return NewValue(false)
	}

	// Key in map
	if haystack.Type == TypeMap {
		rv := reflect.ValueOf(haystack.Map)
		// Convert v to appropriate key type and check existence
		for _, key := range rv.MapKeys() {
			keyVal, err := NewValue(key.Interface())
			if err != nil {
				continue
			}
			eq, err := v.Equal(keyVal)
			if err == nil && eq.Bool {
				return NewValue(true)
			}
		}
		return NewValue(false)
	}

	return nil, fmt.Errorf("%w: 'in' operator for %v in %v", ErrUnsupportedOperator, v.Type, haystack.Type)
}
