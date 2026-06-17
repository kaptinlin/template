package template

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/kaptinlin/filter"
)

// Interfaces

// node is the interface that all AST nodes must implement.
// Each node represents a part of the template syntax tree.
type node interface {
	// Position returns the line and column where this node starts.
	Position() (line, col int)

	// String returns a string representation of the node for debugging.
	String() string
}

// statement is the interface for all statement nodes.
// Statements are executed and produce output or side effects.
type statement interface {
	node
	Execute(ctx *renderContext, writer io.Writer) error
}

// expression is the interface for all expression nodes.
// Expressions are evaluated to produce values.
type expression interface {
	node
	Evaluate(ctx *renderContext) (*value, error)
}

// statement node Types

// textNode represents plain text outside of template tags.
type textNode struct {
	Text string
	Line int
	Col  int
}

// outputNode represents a variable output {{ ... }}.
type outputNode struct {
	Expr expression
	Line int
	Col  int
}

// ifNode represents an if-elif-else conditional block.
type ifNode struct {
	Branches []ifBranch
	ElseBody []node
	Line     int
	Col      int
}

// ifBranch represents a single if or elif branch.
type ifBranch struct {
	Condition expression
	Body      []node
}

// loopContext represents loop metadata for templates.
type loopContext struct {
	Index      int
	Counter    int
	Revindex   int
	Revcounter int
	First      bool
	Last       bool
	Length     int
	Parent     *loopContext
}

// forNode represents a for loop.
type forNode struct {
	Vars       []string
	Collection expression
	Body       []node
	Line       int
	Col        int
}

// breakNode represents a {% break %} statement.
type breakNode struct {
	Line int
	Col  int
}

// continueNode represents a {% continue %} statement.
type continueNode struct {
	Line int
	Col  int
}

// expression node Types

// literalNode represents a literal value (string, number, boolean).
type literalNode struct {
	Value any
	Line  int
	Col   int
}

// variableNode represents a variable reference.
type variableNode struct {
	Name string
	Line int
	Col  int
}

// binaryOpNode represents a binary operation.
type binaryOpNode struct {
	Operator string
	Left     expression
	Right    expression
	Line     int
	Col      int
}

// unaryOpNode represents a unary operation.
type unaryOpNode struct {
	Operator string
	Operand  expression
	Line     int
	Col      int
}

// propertyAccessNode represents property/attribute access.
type propertyAccessNode struct {
	Object   expression
	Property string
	Line     int
	Col      int
}

// subscriptNode represents subscript/index access.
type subscriptNode struct {
	Object expression
	Index  expression
	Line   int
	Col    int
}

// filterNode represents a filter application.
type filterNode struct {
	Expr expression
	Name string
	Args []expression
	Line int
	Col  int
}

// Loop Control Errors

// breakError signals loop termination.
type breakError struct{}

// Error implements the error interface.
func (e *breakError) Error() string { return ErrBreakOutsideLoop.Error() }

// continueError signals loop continuation.
type continueError struct{}

// Error implements the error interface.
func (e *continueError) Error() string { return ErrContinueOutsideLoop.Error() }

// Constructors

// newTextNode returns a new textNode.
func newTextNode(text string, line, col int) *textNode {
	return &textNode{Text: text, Line: line, Col: col}
}

// newOutputNode returns a new outputNode.
func newOutputNode(expr expression, line, col int) *outputNode {
	return &outputNode{Expr: expr, Line: line, Col: col}
}

// newLiteralNode returns a new literalNode.
func newLiteralNode(value any, line, col int) *literalNode {
	return &literalNode{Value: value, Line: line, Col: col}
}

// newVariableNode returns a new variableNode.
func newVariableNode(name string, line, col int) *variableNode {
	return &variableNode{Name: name, Line: line, Col: col}
}

// newFilterNode returns a new filterNode.
func newFilterNode(expr expression, name string, args []expression, line, col int) *filterNode {
	return &filterNode{Expr: expr, Name: name, Args: args, Line: line, Col: col}
}

// newPropertyAccessNode returns a new propertyAccessNode.
func newPropertyAccessNode(object expression, property string, line, col int) *propertyAccessNode {
	return &propertyAccessNode{Object: object, Property: property, Line: line, Col: col}
}

// newSubscriptNode returns a new subscriptNode.
func newSubscriptNode(object, index expression, line, col int) *subscriptNode {
	return &subscriptNode{Object: object, Index: index, Line: line, Col: col}
}

// newBinaryOpNode returns a new binaryOpNode.
func newBinaryOpNode(operator string, left, right expression, line, col int) *binaryOpNode {
	return &binaryOpNode{Operator: operator, Left: left, Right: right, Line: line, Col: col}
}

// newUnaryOpNode returns a new unaryOpNode.
func newUnaryOpNode(operator string, operand expression, line, col int) *unaryOpNode {
	return &unaryOpNode{Operator: operator, Operand: operand, Line: line, Col: col}
}

// textNode Methods

// Position returns the position of the textNode.
func (n *textNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the textNode.
func (n *textNode) String() string { return fmt.Sprintf("Text(%q)", n.Text) }

// Execute writes the raw text to the output.
func (n *textNode) Execute(_ *renderContext, w io.Writer) error {
	_, err := io.WriteString(w, n.Text)
	return err
}

// outputNode Methods

// Position returns the position of the outputNode.
func (n *outputNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the outputNode.
func (n *outputNode) String() string { return fmt.Sprintf("Output(%s)", n.Expr) }

// Execute evaluates the expression and writes its string value.
//
// When the render context has autoescape enabled (FormatHTML engine rendering),
// the output is HTML-escaped UNLESS the underlying Go value is a
// [SafeHTML]. In non-autoescape rendering, SafeHTML is treated as a plain
// string.
func (n *outputNode) Execute(ctx *renderContext, w io.Writer) error {
	val, err := n.Expr.Evaluate(ctx)
	if err != nil {
		return err
	}
	// Detect SafeHTML at the raw-value level: the filter pipeline
	// downgrades to plain string when any non-safe-aware filter runs,
	// which is exactly the behavior we want.
	if s, ok := val.Interface().(SafeHTML); ok {
		_, err = io.WriteString(w, string(s))
		return err
	}
	out := val.String()
	if ctx.autoescape {
		out = filter.Escape(out)
	}
	_, err = io.WriteString(w, out)
	return err
}

// ifNode Methods

// Position returns the position of the ifNode.
func (n *ifNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the ifNode.
func (n *ifNode) String() string { return fmt.Sprintf("If(%d branches)", len(n.Branches)) }

// Execute runs the first truthy branch, or the else block if no branch matches.
func (n *ifNode) Execute(ctx *renderContext, w io.Writer) error {
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

// forNode Methods

// Position returns the position of the forNode.
func (n *forNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the forNode.
func (n *forNode) String() string {
	return fmt.Sprintf("For(%v in %s)", n.Vars, n.Collection)
}

// Execute evaluates the iterable and executes the loop body for each element.
func (n *forNode) Execute(ctx *renderContext, w io.Writer) error {
	col, err := n.Collection.Evaluate(ctx)
	if err != nil {
		return err
	}

	// Django behavior: {% for item in map %} binds to the map key.
	rv := col.resolved()
	bindToKey := rv.IsValid() && rv.Kind() == reflect.Map

	prevLoop, hadLoop := ctx.Locals["loop"]
	var parent *loopContext
	if lc, ok := prevLoop.(*loopContext); ok {
		parent = lc
	}

	prevBindings := make(map[string]any, len(n.Vars))
	hadBindings := make(map[string]bool, len(n.Vars))
	for _, name := range n.Vars {
		prevBindings[name], hadBindings[name] = ctx.Locals[name]
	}

	defer func() {
		if hadLoop {
			ctx.Locals["loop"] = prevLoop
		} else {
			delete(ctx.Locals, "loop")
		}
		for _, name := range n.Vars {
			if hadBindings[name] {
				ctx.Locals[name] = prevBindings[name]
			} else {
				delete(ctx.Locals, name)
			}
		}
	}()

	var execErr error

	iterErr := col.Iterate(func(idx, count int, key, value *value) bool {
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

		ctx.Set("loop", &loopContext{
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
			s, ok := stmt.(statement)
			if !ok {
				continue
			}
			if err := s.Execute(ctx, w); err != nil {
				if _, ok := errors.AsType[*breakError](err); ok {
					return false
				}
				if _, ok := errors.AsType[*continueError](err); ok {
					return true
				}
				execErr = wrapRender(s, err)
				return false
			}
		}
		return true
	})

	if execErr != nil {
		return execErr
	}
	return iterErr
}

// breakNode Methods

// Position returns the position of the breakNode.
func (n *breakNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the breakNode.
func (n *breakNode) String() string { return "Break" }

// Execute signals loop termination via breakError.
func (n *breakNode) Execute(_ *renderContext, _ io.Writer) error {
	return &breakError{}
}

// continueNode Methods

// Position returns the position of the continueNode.
func (n *continueNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the continueNode.
func (n *continueNode) String() string { return "Continue" }

// Execute signals loop continuation via continueError.
func (n *continueNode) Execute(_ *renderContext, _ io.Writer) error {
	return &continueError{}
}

// literalNode Methods

// Position returns the position of the literalNode.
func (n *literalNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the literalNode.
func (n *literalNode) String() string { return fmt.Sprintf("Literal(%v)", n.Value) }

// Evaluate returns the literal value wrapped in a value.
func (n *literalNode) Evaluate(_ *renderContext) (*value, error) {
	return newValue(n.Value), nil
}

// variableNode Methods

// Position returns the position of the variableNode.
func (n *variableNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the variableNode.
func (n *variableNode) String() string { return fmt.Sprintf("Var(%s)", n.Name) }

// Evaluate resolves the variable in the current render context.
func (n *variableNode) Evaluate(ctx *renderContext) (*value, error) {
	val, ok := ctx.Get(n.Name)
	if !ok {
		return newValue(nil), nil
	}
	return newValue(val), nil
}

// binaryOpNode Methods

// Position returns the position of the binaryOpNode.
func (n *binaryOpNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the binaryOpNode.
func (n *binaryOpNode) String() string {
	return fmt.Sprintf("BinOp(%s %s %s)", n.Left, n.Operator, n.Right)
}

// Evaluate computes the binary operation result.
func (n *binaryOpNode) Evaluate(ctx *renderContext) (*value, error) {
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
			return newValue(lf + rf), nil
		}
		return newValue(left.String() + right.String()), nil

	case "-":
		lf, err := left.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotSubtractTypes, err)
		}
		rf, err := right.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotSubtractTypes, err)
		}
		return newValue(lf - rf), nil

	case "*":
		lf, err := left.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotMultiplyTypes, err)
		}
		rf, err := right.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotMultiplyTypes, err)
		}
		return newValue(lf * rf), nil

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
		return newValue(lf / rf), nil

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
		return newValue(li % ri), nil

	case "==":
		return newValue(left.Equals(right)), nil

	case "!=":
		return newValue(!left.Equals(right)), nil

	case "<", ">", "<=", ">=":
		cmp, err := left.Compare(right)
		if err != nil {
			return nil, err
		}
		switch n.Operator {
		case "<":
			return newValue(cmp < 0), nil
		case ">":
			return newValue(cmp > 0), nil
		case "<=":
			return newValue(cmp <= 0), nil
		case ">=":
			return newValue(cmp >= 0), nil
		}

	case "and":
		if !left.IsTrue() {
			return newValue(false), nil
		}
		return newValue(right.IsTrue()), nil

	case "or":
		if left.IsTrue() {
			return newValue(true), nil
		}
		return newValue(right.IsTrue()), nil
	}

	return nil, fmt.Errorf("%w: %q", ErrUnsupportedOperator, n.Operator)
}

// unaryOpNode Methods

// Position returns the position of the unaryOpNode.
func (n *unaryOpNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the unaryOpNode.
func (n *unaryOpNode) String() string {
	return fmt.Sprintf("UnaryOp(%s %s)", n.Operator, n.Operand)
}

// Evaluate computes the unary operation result.
func (n *unaryOpNode) Evaluate(ctx *renderContext) (*value, error) {
	operand, err := n.Operand.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	switch n.Operator {
	case "not":
		return newValue(!operand.IsTrue()), nil

	case "-":
		f, err := operand.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotNegate, err)
		}
		return newValue(-f), nil

	case "+":
		f, err := operand.Float()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotApplyUnaryPlus, err)
		}
		return newValue(f), nil

	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedUnaryOp, n.Operator)
	}
}

// propertyAccessNode Methods

// Position returns the position of the propertyAccessNode.
func (n *propertyAccessNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the propertyAccessNode.
func (n *propertyAccessNode) String() string {
	return fmt.Sprintf("PropAccess(%s.%s)", n.Object, n.Property)
}

// Evaluate returns the property value from the evaluated object.
func (n *propertyAccessNode) Evaluate(ctx *renderContext) (*value, error) {
	object, err := n.Object.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	return object.Field(n.Property)
}

// subscriptNode Methods

// Position returns the position of the subscriptNode.
func (n *subscriptNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the subscriptNode.
func (n *subscriptNode) String() string {
	return fmt.Sprintf("Subscript(%s[%s])", n.Object, n.Index)
}

// Evaluate returns the indexed or keyed value.
func (n *subscriptNode) Evaluate(ctx *renderContext) (*value, error) {
	object, err := n.Object.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	index, err := n.Index.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	//exhaustive:ignore - only collection kinds need specialized subscript handling.
	switch object.resolved().Kind() {
	case reflect.Map:
		return object.Key(index.Interface())
	case reflect.Slice, reflect.Array, reflect.String:
		idx, err := index.Int()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidIndexType, err)
		}
		if !int64FitsInInt(idx) {
			return nil, fmt.Errorf("%w: %d", ErrIndexOutOfRange, idx)
		}
		return object.Index(int(idx))
	}
	if idx, err := index.Int(); err == nil {
		return object.Index(int(idx))
	}
	return object.Key(index.Interface())
}

// filterNode Methods

// Position returns the position of the filterNode.
func (n *filterNode) Position() (int, int) { return n.Line, n.Col }

// String returns a debug representation of the filterNode.
func (n *filterNode) String() string {
	if len(n.Args) > 0 {
		return fmt.Sprintf("Filter(%s|%s:%v)", n.Expr, n.Name, n.Args)
	}
	return fmt.Sprintf("Filter(%s|%s)", n.Expr, n.Name)
}

// Evaluate applies the named filter to the expression value.
//
// Filter lookup consults the per-engine filter registry first, falling
// back to the built-in registry. This gives each engine access to its own
// feature-gated filters and format-specific overrides.
func (n *filterNode) Evaluate(ctx *renderContext) (*value, error) {
	return n.evaluate(ctx, nil)
}

type boundFilterNode struct {
	*filterNode
	fn FilterFunc
}

func (n *boundFilterNode) Evaluate(ctx *renderContext) (*value, error) {
	return n.evaluate(ctx, n.fn)
}

func (n *filterNode) evaluate(ctx *renderContext, filterFn FilterFunc) (*value, error) {
	val, err := n.Expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	found := filterFn != nil
	if !found {
		if ctx.engine != nil && ctx.engine.filters != nil {
			filterFn, found = ctx.engine.filters.Filter(n.Name)
		} else {
			filterFn, found = defaultRegistry.Filter(n.Name)
		}
	}
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrFilterNotFound, n.Name)
	}

	args := make([]any, len(n.Args))
	for i, arg := range n.Args {
		v, err := arg.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		args[i] = v.Interface()
	}

	result, err := filterFn(val.Interface(), args...)
	if err != nil {
		return nil, fmt.Errorf("filter %s: %w", n.Name, err)
	}
	return newValue(result), nil
}

// Helper Functions

// executeBody executes a list of nodes as statements, returning the first error.
func executeBody(body []node, ctx *renderContext, w io.Writer) error {
	for _, node := range body {
		s, ok := node.(statement)
		if !ok {
			continue
		}
		if err := s.Execute(ctx, w); err != nil {
			return wrapRender(s, err)
		}
	}
	return nil
}
