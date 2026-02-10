package template

import (
	"errors"
	"fmt"
	"io"
	"reflect"
)

// Node is the interface that all AST nodes must implement.
// Each node represents a part of the template syntax tree.
type Node interface {
	// Position returns the line and column where this node starts.
	Position() (line, col int)

	// String returns a string representation of the node for debugging.
	String() string
}

// =============================================================================
// Statement Nodes - represent template statements (if, for, output, etc.)
// =============================================================================

// Statement is the interface for all statement nodes.
// Statements are executed and produce output or side effects.
type Statement interface {
	Node
	Execute(ctx *ExecutionContext, writer io.Writer) error
}

// TextNode represents plain text outside of template tags.
type TextNode struct {
	Text string
	Line int
	Col  int
}

// Position implements Node.
func (n *TextNode) Position() (int, int) { return n.Line, n.Col }
func (n *TextNode) String() string       { return fmt.Sprintf("Text(%q)", n.Text) }

// Execute writes the raw text to the output.
func (n *TextNode) Execute(_ *ExecutionContext, w io.Writer) error {
	_, err := w.Write([]byte(n.Text))
	return err
}

// OutputNode represents a variable output {{ ... }}.
// Example: {{ name }}, {{ user.email|upper }}
type OutputNode struct {
	Expr Expression // The expression to evaluate and output
	Line int
	Col  int
}

// Position implements Node.
func (n *OutputNode) Position() (int, int) { return n.Line, n.Col }
func (n *OutputNode) String() string       { return fmt.Sprintf("Output(%s)", n.Expr) }

// Execute evaluates the expression and writes its string value.
func (n *OutputNode) Execute(ctx *ExecutionContext, w io.Writer) error {
	val, err := n.Expr.Evaluate(ctx)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(val.String()))
	return err
}

// IfNode represents an if-elif-else conditional block.
// Example: {% if x > 0 %} ... {% elif x < 0 %} ... {% else %} ... {% endif %}
type IfNode struct {
	Branches []IfBranch // List of if/elif branches
	ElseBody []Node     // Optional else body
	Line     int
	Col      int
}

// IfBranch represents a single if or elif branch.
type IfBranch struct {
	Condition Expression // The condition expression
	Body      []Node     // Statements to execute if condition is true
}

// Position implements Node.
func (n *IfNode) Position() (int, int) { return n.Line, n.Col }
func (n *IfNode) String() string       { return fmt.Sprintf("If(%d branches)", len(n.Branches)) }

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

// LoopContext represents the loop information available in templates.
// Provides loop metadata compatible with both Django and Jinja2 styles.
type LoopContext struct {
	Index      int          // Current zero-based index.
	Counter    int          // Current one-based counter (Index + 1).
	Revindex   int          // Reverse index (length-1 to 0).
	Revcounter int          // Remaining iterations including current (length to 1).
	First      bool         // True on first iteration.
	Last       bool         // True on last iteration.
	Length     int          // Total collection length.
	Parent     *LoopContext // Parent loop context for nested loops.
}

// ForNode represents a for loop.
// Example: {% for item in items %} ... {% endfor %}
// Example: {% for key, value in dict %} ... {% endfor %}
type ForNode struct {
	Vars       []string   // Variable names (e.g., ["item"] or ["key", "value"])
	Collection Expression // Expression that evaluates to an iterable
	Body       []Node     // Statements inside the loop
	Line       int
	Col        int
}

// Position implements Node.
func (n *ForNode) Position() (int, int) { return n.Line, n.Col }
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
	bindToKey := false
	rv := col.resolved()
	if rv.IsValid() && rv.Kind() == reflect.Map {
		bindToKey = true
	}

	var parent *LoopContext
	if v, ok := ctx.Get("loop"); ok {
		if lc, ok := v.(*LoopContext); ok {
			parent = lc
		}
	}

	var execErr error

	iterErr := col.Iterate(func(idx, count int, key, value *Value) bool {
		if len(n.Vars) == 1 {
			if bindToKey {
				ctx.Set(n.Vars[0], key.Interface())
			} else {
				ctx.Set(n.Vars[0], value.Interface())
			}
		} else if len(n.Vars) == 2 {
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
				var brk *BreakError
				if errors.As(err, &brk) {
					return false
				}
				var cont *ContinueError
				if errors.As(err, &cont) {
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

// BreakNode represents a {% break %} statement.
type BreakNode struct {
	Line int
	Col  int
}

// Position implements Node.
func (n *BreakNode) Position() (int, int) { return n.Line, n.Col }
func (n *BreakNode) String() string       { return "Break" }

// Execute signals loop termination via BreakError.
func (n *BreakNode) Execute(_ *ExecutionContext, _ io.Writer) error {
	return &BreakError{}
}

// ContinueNode represents a {% continue %} statement.
type ContinueNode struct {
	Line int
	Col  int
}

// Position implements Node.
func (n *ContinueNode) Position() (int, int) { return n.Line, n.Col }
func (n *ContinueNode) String() string       { return "Continue" }

// Execute signals loop continuation via ContinueError.
func (n *ContinueNode) Execute(_ *ExecutionContext, _ io.Writer) error {
	return &ContinueError{}
}

// =============================================================================
// Expression Nodes - represent expressions that evaluate to values
// =============================================================================

// Expression is the interface for all expression nodes.
// Expressions are evaluated to produce values.
type Expression interface {
	Node
	Evaluate(ctx *ExecutionContext) (*Value, error)
}

// LiteralNode represents a literal value (string, number, boolean).
// Examples: "hello", 42, 3.14, true, false
type LiteralNode struct {
	Value any // The literal value (string, float64, bool)
	Line  int
	Col   int
}

// Position implements Node.
func (n *LiteralNode) Position() (int, int) { return n.Line, n.Col }
func (n *LiteralNode) String() string       { return fmt.Sprintf("Literal(%v)", n.Value) }

// Evaluate returns the literal value as-is.
func (n *LiteralNode) Evaluate(_ *ExecutionContext) (*Value, error) {
	return NewValue(n.Value), nil
}

// VariableNode represents a variable reference.
// Example: name, user, items
type VariableNode struct {
	Name string
	Line int
	Col  int
}

// Position implements Node.
func (n *VariableNode) Position() (int, int) { return n.Line, n.Col }
func (n *VariableNode) String() string       { return fmt.Sprintf("Var(%s)", n.Name) }

// Evaluate resolves the variable in the current execution context.
func (n *VariableNode) Evaluate(ctx *ExecutionContext) (*Value, error) {
	val, ok := ctx.Get(n.Name)
	if !ok {
		return NewValue(nil), nil // Return nil for undefined variables
	}
	return NewValue(val), nil
}

// BinaryOpNode represents a binary operation.
// Examples: a + b, x == 10, count > 0
type BinaryOpNode struct {
	Operator string     // Operator: +, -, *, /, ==, !=, <, >, <=, >=, and, or
	Left     Expression // Left operand
	Right    Expression // Right operand
	Line     int
	Col      int
}

// Position implements Node.
func (n *BinaryOpNode) Position() (int, int) { return n.Line, n.Col }
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

	case "<":
		cmp, err := left.Compare(right)
		if err != nil {
			return nil, err
		}
		return NewValue(cmp < 0), nil

	case ">":
		cmp, err := left.Compare(right)
		if err != nil {
			return nil, err
		}
		return NewValue(cmp > 0), nil

	case "<=":
		cmp, err := left.Compare(right)
		if err != nil {
			return nil, err
		}
		return NewValue(cmp <= 0), nil

	case ">=":
		cmp, err := left.Compare(right)
		if err != nil {
			return nil, err
		}
		return NewValue(cmp >= 0), nil

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

	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedOperator, n.Operator)
	}
}

// UnaryOpNode represents a unary operation.
// Examples: not flag, -value
type UnaryOpNode struct {
	Operator string     // Operator: not, -, +
	Operand  Expression // The operand
	Line     int
	Col      int
}

// Position implements Node.
func (n *UnaryOpNode) Position() (int, int) { return n.Line, n.Col }
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

// PropertyAccessNode represents property/attribute access.
// Example: user.name, item.price
type PropertyAccessNode struct {
	Object   Expression // The object to access
	Property string     // The property name
	Line     int
	Col      int
}

// Position implements Node.
func (n *PropertyAccessNode) Position() (int, int) { return n.Line, n.Col }
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

// SubscriptNode represents subscript/index access.
// Example: items[0], dict["key"]
type SubscriptNode struct {
	Object Expression // The object to subscript
	Index  Expression // The index/key expression
	Line   int
	Col    int
}

// Position implements Node.
func (n *SubscriptNode) Position() (int, int) { return n.Line, n.Col }
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

	if idx, err := index.Int(); err == nil { // if NO error
		return object.Index(int(idx))
	}
	return object.Key(index.Interface())
}

// FilterNode represents a filter application.
// Examples: name|upper, price|add:10, text|slice:0:10
type FilterNode struct {
	Expr Expression   // The expression to filter
	Name string       // The filter name
	Args []Expression // Filter arguments (can be empty)
	Line int
	Col  int
}

// Position implements Node.
func (n *FilterNode) Position() (int, int) { return n.Line, n.Col }
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

	fn, ok := GetFilter(n.Name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrFilterNotFound, n.Name)
	}

	args := make([]string, 0, len(n.Args))
	for _, arg := range n.Args {
		v, err := arg.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		args = append(args, v.String())
	}

	result, err := fn(val.Interface(), args...)
	if err != nil {
		return nil, fmt.Errorf("filter %s: %w", n.Name, err)
	}
	return NewValue(result), nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// executeBody executes a list of nodes as statements, returning the first error.
func executeBody(body []Node, ctx *ExecutionContext, w io.Writer) error {
	for _, node := range body {
		if s, ok := node.(Statement); ok {
			if err := s.Execute(ctx, w); err != nil {
				return err
			}
		}
	}
	return nil
}

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

// =============================================================================
// Loop Control Errors
// =============================================================================

// BreakError signals loop termination.
type BreakError struct{}

func (e *BreakError) Error() string { return ErrBreakOutsideLoop.Error() }

// ContinueError signals loop continuation.
type ContinueError struct{}

func (e *ContinueError) Error() string { return ErrContinueOutsideLoop.Error() }
