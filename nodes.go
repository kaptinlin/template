package template

import (
	"errors"
	"fmt"
	"io"
	"reflect"
)

// Interfaces

// Node is the interface that all AST nodes must implement.
// Each node represents a part of the template syntax tree.
type Node interface {
	// Position returns the line and column where this node starts.
	Position() (line, col int)

	// String returns a string representation of the node for debugging.
	String() string
}

// Statement is the interface for all statement nodes.
// Statements are executed and produce output or side effects.
type Statement interface {
	Node
	Execute(ctx *ExecutionContext, writer io.Writer) error
}

// Expression is the interface for all expression nodes.
// Expressions are evaluated to produce values.
type Expression interface {
	Node
	Evaluate(ctx *ExecutionContext) (*Value, error)
}

// Statement Node Types

// TextNode represents plain text outside of template tags.
type TextNode struct {
	Text string
	Line int
	Col  int
}

// OutputNode represents a variable output {{ ... }}.
type OutputNode struct {
	Expr Expression
	Line int
	Col  int
}

// IfNode represents an if-elif-else conditional block.
type IfNode struct {
	Branches []IfBranch
	ElseBody []Node
	Line     int
	Col      int
}

// IfBranch represents a single if or elif branch.
type IfBranch struct {
	Condition Expression
	Body      []Node
}

// LoopContext represents loop metadata for templates.
type LoopContext struct {
	Index      int
	Counter    int
	Revindex   int
	Revcounter int
	First      bool
	Last       bool
	Length     int
	Parent     *LoopContext
}

// ForNode represents a for loop.
type ForNode struct {
	Vars       []string
	Collection Expression
	Body       []Node
	Line       int
	Col        int
}

// BreakNode represents a {% break %} statement.
type BreakNode struct {
	Line int
	Col  int
}

// ContinueNode represents a {% continue %} statement.
type ContinueNode struct {
	Line int
	Col  int
}

// Expression Node Types

// LiteralNode represents a literal value (string, number, boolean).
type LiteralNode struct {
	Value any
	Line  int
	Col   int
}

// VariableNode represents a variable reference.
type VariableNode struct {
	Name string
	Line int
	Col  int
}

// BinaryOpNode represents a binary operation.
type BinaryOpNode struct {
	Operator string
	Left     Expression
	Right    Expression
	Line     int
	Col      int
}

// UnaryOpNode represents a unary operation.
type UnaryOpNode struct {
	Operator string
	Operand  Expression
	Line     int
	Col      int
}

// PropertyAccessNode represents property/attribute access.
type PropertyAccessNode struct {
	Object   Expression
	Property string
	Line     int
	Col      int
}

// SubscriptNode represents subscript/index access.
type SubscriptNode struct {
	Object Expression
	Index  Expression
	Line   int
	Col    int
}

// FilterNode represents a filter application.
type FilterNode struct {
	Expr Expression
	Name string
	Args []Expression
	Line int
	Col  int
}

// Loop Control Errors

// BreakError signals loop termination.
type BreakError struct{}

// Error implements the error interface.
func (e *BreakError) Error() string { return ErrBreakOutsideLoop.Error() }

// ContinueError signals loop continuation.
type ContinueError struct{}

// Error implements the error interface.
func (e *ContinueError) Error() string { return ErrContinueOutsideLoop.Error() }

// Constructors

// NewTextNode returns a new TextNode.
func NewTextNode(text string, line, col int) *TextNode {
	return &TextNode{Text: text, Line: line, Col: col}
}

// NewOutputNode returns a new OutputNode.
func NewOutputNode(expr Expression, line, col int) *OutputNode {
	return &OutputNode{Expr: expr, Line: line, Col: col}
}

// NewLiteralNode returns a new LiteralNode.
func NewLiteralNode(value any, line, col int) *LiteralNode {
	return &LiteralNode{Value: value, Line: line, Col: col}
}

// NewVariableNode returns a new VariableNode.
func NewVariableNode(name string, line, col int) *VariableNode {
	return &VariableNode{Name: name, Line: line, Col: col}
}

// NewFilterNode returns a new FilterNode.
func NewFilterNode(expr Expression, name string, args []Expression, line, col int) *FilterNode {
	return &FilterNode{Expr: expr, Name: name, Args: args, Line: line, Col: col}
}

// NewPropertyAccessNode returns a new PropertyAccessNode.
func NewPropertyAccessNode(object Expression, property string, line, col int) *PropertyAccessNode {
	return &PropertyAccessNode{Object: object, Property: property, Line: line, Col: col}
}

// NewSubscriptNode returns a new SubscriptNode.
func NewSubscriptNode(object, index Expression, line, col int) *SubscriptNode {
	return &SubscriptNode{Object: object, Index: index, Line: line, Col: col}
}

// NewBinaryOpNode returns a new BinaryOpNode.
func NewBinaryOpNode(operator string, left, right Expression, line, col int) *BinaryOpNode {
	return &BinaryOpNode{Operator: operator, Left: left, Right: right, Line: line, Col: col}
}

// NewUnaryOpNode returns a new UnaryOpNode.
func NewUnaryOpNode(operator string, operand Expression, line, col int) *UnaryOpNode {
	return &UnaryOpNode{Operator: operator, Operand: operand, Line: line, Col: col}
}

// TextNode Methods

// Position returns the position of the TextNode.
func (n *TextNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the TextNode.
func (n *TextNode) String() string { return fmt.Sprintf("Text(%q)", n.Text) }

// Execute writes the raw text to the output.
func (n *TextNode) Execute(_ *ExecutionContext, w io.Writer) error {
	_, err := io.WriteString(w, n.Text)
	return err
}

// OutputNode Methods

// Position returns the position of the OutputNode.
func (n *OutputNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the OutputNode.
func (n *OutputNode) String() string { return fmt.Sprintf("Output(%s)", n.Expr) }

// Execute evaluates the expression and writes its string value.
func (n *OutputNode) Execute(ctx *ExecutionContext, w io.Writer) error {
	val, err := n.Expr.Evaluate(ctx)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, val.String())
	return err
}

// IfNode Methods

// Position returns the position of the IfNode.
func (n *IfNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the IfNode.
func (n *IfNode) String() string { return fmt.Sprintf("If(%d branches)", len(n.Branches)) }

// Execute runs the first truthy branch, or the else block if no branch matches.
func (n *IfNode) Execute(ctx *ExecutionContext, w io.Writer) error {
	for _, branch := range n.Branches {
		val, err := branch.Condition.Evaluate(ctx)
		if err != nil {
			return err
		}
		if val.IsTrue() {
			return executeBody(branch.Body, ctx, w)
		}
	}
	if n.ElseBody != nil {
		return executeBody(n.ElseBody, ctx, w)
	}
	return nil
}

// ForNode Methods

// Position returns the position of the ForNode.
func (n *ForNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the ForNode.
func (n *ForNode) String() string {
	return fmt.Sprintf("For(%v in %s)", n.Vars, n.Collection)
}

// Execute evaluates the iterable and executes the loop body for each element.
func (n *ForNode) Execute(ctx *ExecutionContext, w io.Writer) error {
	col, err := n.Collection.Evaluate(ctx)
	if err != nil {
		return err
	}

	// Django behavior: {% for item in map %} binds to the map key.
	rv := col.resolved()
	bindToKey := rv.IsValid() && rv.Kind() == reflect.Map

	var parent *LoopContext
	if v, ok := ctx.Get("loop"); ok {
		if lc, ok := v.(*LoopContext); ok {
			parent = lc
		}
	}

	var execErr error

	iterErr := col.Iterate(func(idx, count int, key, value *Value) bool {
		switch len(n.Vars) {
		case 1:
			if bindToKey {
				ctx.Set(n.Vars[0], key.Interface())
			} else {
				ctx.Set(n.Vars[0], value.Interface())
			}
		case 2:
			ctx.Set(n.Vars[0], key.Interface())
			ctx.Set(n.Vars[1], value.Interface())
		}

		ctx.Set("loop", &LoopContext{
			Index:      idx,
			Counter:    idx + 1,
			Revindex:   count - 1 - idx,
			Revcounter: count - idx,
			First:      idx == 0,
			Last:       idx == count-1,
			Length:     count,
			Parent:     parent,
		})

		for _, stmt := range n.Body {
			s, ok := stmt.(Statement)
			if !ok {
				continue
			}
			if err := s.Execute(ctx, w); err != nil {
				if _, ok := errors.AsType[*BreakError](err); ok {
					return false
				}
				if _, ok := errors.AsType[*ContinueError](err); ok {
					return true
				}
				execErr = err
				return false
			}
		}
		return true
	})

	// Restore parent loop context.
	if parent != nil {
		ctx.Set("loop", parent)
	}

	if execErr != nil {
		return execErr
	}
	return iterErr
}

// BreakNode Methods

// Position returns the position of the BreakNode.
func (n *BreakNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the BreakNode.
func (n *BreakNode) String() string { return "Break" }

// Execute signals loop termination via BreakError.
func (n *BreakNode) Execute(_ *ExecutionContext, _ io.Writer) error {
	return &BreakError{}
}

// ContinueNode Methods

// Position returns the position of the ContinueNode.
func (n *ContinueNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the ContinueNode.
func (n *ContinueNode) String() string { return "Continue" }

// Execute signals loop continuation via ContinueError.
func (n *ContinueNode) Execute(_ *ExecutionContext, _ io.Writer) error {
	return &ContinueError{}
}

// LiteralNode Methods

// Position returns the position of the LiteralNode.
func (n *LiteralNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the LiteralNode.
func (n *LiteralNode) String() string { return fmt.Sprintf("Literal(%v)", n.Value) }

// Evaluate returns the literal value wrapped in a Value.
func (n *LiteralNode) Evaluate(_ *ExecutionContext) (*Value, error) {
	return NewValue(n.Value), nil
}

// VariableNode Methods

// Position returns the position of the VariableNode.
func (n *VariableNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the VariableNode.
func (n *VariableNode) String() string { return fmt.Sprintf("Var(%s)", n.Name) }

// Evaluate resolves the variable in the current execution context.
func (n *VariableNode) Evaluate(ctx *ExecutionContext) (*Value, error) {
	val, ok := ctx.Get(n.Name)
	if !ok {
		return NewValue(nil), nil
	}
	return NewValue(val), nil
}

// BinaryOpNode Methods

// Position returns the position of the BinaryOpNode.
func (n *BinaryOpNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the BinaryOpNode.
func (n *BinaryOpNode) String() string {
	return fmt.Sprintf("BinOp(%s %s %s)", n.Left, n.Operator, n.Right)
}

// Evaluate computes the binary operation result.
func (n *BinaryOpNode) Evaluate(ctx *ExecutionContext) (*Value, error) {
	left, err := n.Left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	right, err := n.Right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	switch n.Operator {
	case "+":
		lf, lerr := left.Float()
		rf, rerr := right.Float()
		if lerr == nil && rerr == nil {
			return NewValue(lf + rf), nil
		}
		return NewValue(left.String() + right.String()), nil

	case "-":
		lf, err := left.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotSubtractTypes, err)
		}
		rf, err := right.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotSubtractTypes, err)
		}
		return NewValue(lf - rf), nil

	case "*":
		lf, err := left.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotMultiplyTypes, err)
		}
		rf, err := right.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotMultiplyTypes, err)
		}
		return NewValue(lf * rf), nil

	case "/":
		lf, err := left.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotDivideTypes, err)
		}
		rf, err := right.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotDivideTypes, err)
		}
		if rf == 0 {
			return nil, ErrDivisionByZero
		}
		return NewValue(lf / rf), nil

	case "%":
		li, err := left.Int()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotModuloTypes, err)
		}
		ri, err := right.Int()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotModuloTypes, err)
		}
		if ri == 0 {
			return nil, ErrModuloByZero
		}
		return NewValue(li % ri), nil

	case "==":
		return NewValue(left.Equals(right)), nil

	case "!=":
		return NewValue(!left.Equals(right)), nil

	case "<", ">", "<=", ">=":
		cmp, err := left.Compare(right)
		if err != nil {
			return nil, err
		}
		switch n.Operator {
		case "<":
			return NewValue(cmp < 0), nil
		case ">":
			return NewValue(cmp > 0), nil
		case "<=":
			return NewValue(cmp <= 0), nil
		case ">=":
			return NewValue(cmp >= 0), nil
		}

	case "and":
		if !left.IsTrue() {
			return NewValue(false), nil
		}
		return NewValue(right.IsTrue()), nil

	case "or":
		if left.IsTrue() {
			return NewValue(true), nil
		}
		return NewValue(right.IsTrue()), nil
	}

	return nil, fmt.Errorf("%w: %q", ErrUnsupportedOperator, n.Operator)
}

// UnaryOpNode Methods

// Position returns the position of the UnaryOpNode.
func (n *UnaryOpNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the UnaryOpNode.
func (n *UnaryOpNode) String() string {
	return fmt.Sprintf("UnaryOp(%s %s)", n.Operator, n.Operand)
}

// Evaluate computes the unary operation result.
func (n *UnaryOpNode) Evaluate(ctx *ExecutionContext) (*Value, error) {
	operand, err := n.Operand.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	switch n.Operator {
	case "not":
		return NewValue(!operand.IsTrue()), nil

	case "-":
		f, err := operand.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotNegate, err)
		}
		return NewValue(-f), nil

	case "+":
		f, err := operand.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotApplyUnaryPlus, err)
		}
		return NewValue(f), nil

	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedUnaryOp, n.Operator)
	}
}

// PropertyAccessNode Methods

// Position returns the position of the PropertyAccessNode.
func (n *PropertyAccessNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the PropertyAccessNode.
func (n *PropertyAccessNode) String() string {
	return fmt.Sprintf("PropAccess(%s.%s)", n.Object, n.Property)
}

// Evaluate returns the property value from the evaluated object.
func (n *PropertyAccessNode) Evaluate(ctx *ExecutionContext) (*Value, error) {
	object, err := n.Object.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	return object.Field(n.Property)
}

// SubscriptNode Methods

// Position returns the position of the SubscriptNode.
func (n *SubscriptNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the SubscriptNode.
func (n *SubscriptNode) String() string {
	return fmt.Sprintf("Subscript(%s[%s])", n.Object, n.Index)
}

// Evaluate returns the indexed or keyed value.
func (n *SubscriptNode) Evaluate(ctx *ExecutionContext) (*Value, error) {
	object, err := n.Object.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	index, err := n.Index.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if idx, err := index.Int(); err == nil {
		return object.Index(int(idx))
	}
	return object.Key(index.Interface())
}

// FilterNode Methods

// Position returns the position of the FilterNode.
func (n *FilterNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the FilterNode.
func (n *FilterNode) String() string {
	if len(n.Args) > 0 {
		return fmt.Sprintf("Filter(%s|%s:%v)", n.Expr, n.Name, n.Args)
	}
	return fmt.Sprintf("Filter(%s|%s)", n.Expr, n.Name)
}

// Evaluate applies the named filter to the expression value.
func (n *FilterNode) Evaluate(ctx *ExecutionContext) (*Value, error) {
	val, err := n.Expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	fn, ok := Filter(n.Name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrFilterNotFound, n.Name)
	}

	args := make([]string, len(n.Args))
	for i, arg := range n.Args {
		v, err := arg.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		args[i] = v.String()
	}

	result, err := fn(val.Interface(), args...)
	if err != nil {
		return nil, fmt.Errorf("filter %s: %w", n.Name, err)
	}
	return NewValue(result), nil
}

// Helper Functions

// executeBody executes a list of nodes as statements, returning the first error.
func executeBody(body []Node, ctx *ExecutionContext, w io.Writer) error {
	for _, node := range body {
		s, ok := node.(Statement)
		if !ok {
			continue
		}
		if err := s.Execute(ctx, w); err != nil {
			return err
		}
	}
	return nil
}
