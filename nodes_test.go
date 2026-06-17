package template

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
)

// Test sentinel errors for mock types.
var (
	errMockEval       = errors.New("eval failed")
	errMockWrite      = errors.New("write failed")
	errMockObjectEval = errors.New("object eval failed")
	errMockIndexEval  = errors.New("index eval failed")
	errMockLeftEval   = errors.New("left eval failed")
	errMockRightEval  = errors.New("right eval failed")
	errMockOperand    = errors.New("operand eval failed")
	errMockCollection = errors.New("collection eval failed")
	errMockBody       = errors.New("body error")
	errLoopMissing    = errors.New("loop context missing")
	errLoopBadType    = errors.New("loop context has unexpected type")
)

// errExpr is a mock expression that always returns an error.
type errExpr struct{ err error }

func (e *errExpr) Position() (int, int)                      { return 1, 1 }
func (e *errExpr) String() string                            { return "errExpr" }
func (e *errExpr) Evaluate(_ *renderContext) (*value, error) { return nil, e.err }

// errWriter is a mock writer that always returns an error.
type errWriter struct{ err error }

func (w *errWriter) Write(_ []byte) (int, error) { return 0, w.err }

type captureLoopStmt struct {
	last *loopContext
}

func (s *captureLoopStmt) Position() (int, int) { return 1, 1 }
func (s *captureLoopStmt) String() string       { return "captureLoopStmt" }
func (s *captureLoopStmt) Execute(ctx *renderContext, _ io.Writer) error {
	v, ok := ctx.Get("loop")
	if !ok {
		return errLoopMissing
	}
	lc, ok := v.(*loopContext)
	if !ok {
		return errLoopBadType
	}
	copy := *lc
	s.last = &copy
	return nil
}

func TestTextNode(t *testing.T) {
	n := newTextNode("hello", 1, 5)
	line, col := n.Position()
	if line != 1 || col != 5 {
		t.Errorf("Position() = (%d, %d), want (1, 5)", line, col)
	}
	if got := n.String(); got != `Text("hello")` {
		t.Errorf("String() = %q, want %q", got, `Text("hello")`)
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "hello" {
		t.Errorf("Execute() wrote %q, want %q", got, "hello")
	}
}

func TestOutputNode(t *testing.T) {
	n := newOutputNode(newLiteralNode("world", 1, 1), 2, 3)
	line, col := n.Position()
	if line != 2 || col != 3 {
		t.Errorf("Position() = (%d, %d), want (2, 3)", line, col)
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "world" {
		t.Errorf("Execute() wrote %q, want %q", got, "world")
	}
}

func TestLiteralNode(t *testing.T) {
	tests := []struct {
		value any
		want  string
	}{
		{value: "hello", want: "hello"},
		{value: 42.0, want: "42"},
		{value: true, want: "true"},
		{value: nil, want: ""},
	}
	for _, tt := range tests {
		n := newLiteralNode(tt.value, 1, 1)
		ctx := newRenderContext(map[string]any{})
		val, err := n.Evaluate(ctx)
		if err != nil {
			t.Fatalf("Evaluate(%v) error: %v", tt.value, err)
		}
		if got := val.String(); got != tt.want {
			t.Errorf("Evaluate(%v) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestVariableNode(t *testing.T) {
	n := newVariableNode("name", 3, 7)
	if got := n.String(); got != "Var(name)" {
		t.Errorf("String() = %q, want %q", got, "Var(name)")
	}

	t.Run("defined", func(t *testing.T) {
		ctx := newRenderContext(map[string]any{"name": "Alice"})
		val, err := n.Evaluate(ctx)
		if err != nil {
			t.Fatalf("Evaluate() error: %v", err)
		}
		if got := val.String(); got != "Alice" {
			t.Errorf("Evaluate() = %q, want %q", got, "Alice")
		}
	})

	t.Run("undefined", func(t *testing.T) {
		ctx := newRenderContext(map[string]any{})
		val, err := n.Evaluate(ctx)
		if err != nil {
			t.Fatalf("Evaluate() error: %v", err)
		}
		if !val.IsNil() {
			t.Errorf("Evaluate() = %v, want nil", val.Interface())
		}
	})
}

func TestBinaryOpNode(t *testing.T) {
	tests := []struct {
		op        string
		left      any
		right     any
		wantFloat float64
		wantBool  bool
		isBool    bool
	}{
		{op: "+", left: 3.0, right: 4.0, wantFloat: 7},
		{op: "-", left: 10.0, right: 3.0, wantFloat: 7},
		{op: "*", left: 3.0, right: 4.0, wantFloat: 12},
		{op: "/", left: 10.0, right: 2.0, wantFloat: 5},
		{op: "==", left: 5.0, right: 5.0, wantBool: true, isBool: true},
		{op: "!=", left: 5.0, right: 3.0, wantBool: true, isBool: true},
		{op: "<", left: 3.0, right: 5.0, wantBool: true, isBool: true},
		{op: ">", left: 5.0, right: 3.0, wantBool: true, isBool: true},
		{op: "<=", left: 3.0, right: 3.0, wantBool: true, isBool: true},
		{op: ">=", left: 5.0, right: 3.0, wantBool: true, isBool: true},
		{op: "and", left: true, right: true, wantBool: true, isBool: true},
		{op: "and", left: false, right: true, wantBool: false, isBool: true},
		{op: "or", left: false, right: true, wantBool: true, isBool: true},
		{op: "or", left: false, right: false, wantBool: false, isBool: true},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			n := newBinaryOpNode(tt.op,
				newLiteralNode(tt.left, 1, 1),
				newLiteralNode(tt.right, 1, 1), 1, 1)
			ctx := newRenderContext(map[string]any{})
			val, err := n.Evaluate(ctx)
			if err != nil {
				t.Fatalf("Evaluate() error: %v", err)
			}
			if tt.isBool {
				if got := val.IsTrue(); got != tt.wantBool {
					t.Errorf("Evaluate() = %v, want %v", got, tt.wantBool)
				}
			} else {
				got, err := val.Float()
				if err != nil {
					t.Fatalf("Float() error: %v", err)
				}
				if got != tt.wantFloat {
					t.Errorf("Evaluate() = %v, want %v", got, tt.wantFloat)
				}
			}
		})
	}
}

func TestBinaryOpNodeStringConcat(t *testing.T) {
	n := newBinaryOpNode("+",
		newLiteralNode("hello ", 1, 1),
		newLiteralNode("world", 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if got := val.String(); got != "hello world" {
		t.Errorf("Evaluate() = %q, want %q", got, "hello world")
	}
}

func TestBinaryOpNodeDivisionByZero(t *testing.T) {
	n := newBinaryOpNode("/",
		newLiteralNode(10.0, 1, 1),
		newLiteralNode(0.0, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrDivisionByZero) {
		t.Errorf("error = %v, want %v", err, ErrDivisionByZero)
	}
}

func TestBinaryOpNodeModuloByZero(t *testing.T) {
	n := newBinaryOpNode("%",
		newLiteralNode(10, 1, 1),
		newLiteralNode(0, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrModuloByZero) {
		t.Errorf("error = %v, want %v", err, ErrModuloByZero)
	}
}

func TestBinaryOpNodeUnsupported(t *testing.T) {
	n := newBinaryOpNode("^^",
		newLiteralNode(1.0, 1, 1),
		newLiteralNode(2.0, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrUnsupportedOperator) {
		t.Errorf("error = %v, want %v", err, ErrUnsupportedOperator)
	}
}

func TestUnaryOpNode(t *testing.T) {
	t.Run("not", func(t *testing.T) {
		n := newUnaryOpNode("not", newLiteralNode(true, 1, 1), 1, 1)
		ctx := newRenderContext(map[string]any{})
		val, err := n.Evaluate(ctx)
		if err != nil {
			t.Fatalf("Evaluate() error: %v", err)
		}
		if val.IsTrue() {
			t.Error("Evaluate() = true, want false")
		}
	})

	t.Run("negate", func(t *testing.T) {
		n := newUnaryOpNode("-", newLiteralNode(5.0, 1, 1), 1, 1)
		ctx := newRenderContext(map[string]any{})
		val, err := n.Evaluate(ctx)
		if err != nil {
			t.Fatalf("Evaluate() error: %v", err)
		}
		f, _ := val.Float()
		if f != -5.0 {
			t.Errorf("Evaluate() = %v, want -5", f)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		n := newUnaryOpNode("~", newLiteralNode(1.0, 1, 1), 1, 1)
		ctx := newRenderContext(map[string]any{})
		_, err := n.Evaluate(ctx)
		if !errors.Is(err, ErrUnsupportedUnaryOp) {
			t.Errorf("error = %v, want %v", err, ErrUnsupportedUnaryOp)
		}
	})
}

func TestIfNode(t *testing.T) {
	t.Run("true branch", func(t *testing.T) {
		n := &ifNode{
			Branches: []ifBranch{{
				Condition: newLiteralNode(true, 1, 1),
				Body:      []node{newTextNode("yes", 1, 1)},
			}},
			ElseBody: []node{newTextNode("no", 1, 1)},
			Line:     1, Col: 1,
		}
		var buf bytes.Buffer
		ctx := newRenderContext(map[string]any{})
		if err := n.Execute(ctx, &buf); err != nil {
			t.Fatalf("Execute() error: %v", err)
		}
		if got := buf.String(); got != "yes" {
			t.Errorf("Execute() = %q, want %q", got, "yes")
		}
	})

	t.Run("else branch", func(t *testing.T) {
		n := &ifNode{
			Branches: []ifBranch{{
				Condition: newLiteralNode(false, 1, 1),
				Body:      []node{newTextNode("yes", 1, 1)},
			}},
			ElseBody: []node{newTextNode("no", 1, 1)},
			Line:     1, Col: 1,
		}
		var buf bytes.Buffer
		ctx := newRenderContext(map[string]any{})
		if err := n.Execute(ctx, &buf); err != nil {
			t.Fatalf("Execute() error: %v", err)
		}
		if got := buf.String(); got != "no" {
			t.Errorf("Execute() = %q, want %q", got, "no")
		}
	})

	t.Run("no match no else", func(t *testing.T) {
		n := &ifNode{
			Branches: []ifBranch{{
				Condition: newLiteralNode(false, 1, 1),
				Body:      []node{newTextNode("yes", 1, 1)},
			}},
			Line: 1, Col: 1,
		}
		var buf bytes.Buffer
		ctx := newRenderContext(map[string]any{})
		if err := n.Execute(ctx, &buf); err != nil {
			t.Fatalf("Execute() error: %v", err)
		}
		if got := buf.String(); got != "" {
			t.Errorf("Execute() = %q, want empty", got)
		}
	})
}

func TestForNode(t *testing.T) {
	n := &forNode{
		Vars:       []string{"item"},
		Collection: newLiteralNode([]any{"a", "b", "c"}, 1, 1),
		Body: []node{
			newOutputNode(newVariableNode("item", 1, 1), 1, 1),
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "abc" {
		t.Errorf("Execute() = %q, want %q", got, "abc")
	}
}

func TestBreakNode(t *testing.T) {
	n := &breakNode{Line: 1, Col: 1}
	line, col := n.Position()
	if line != 1 || col != 1 {
		t.Errorf("Position() = (%d, %d), want (1, 1)", line, col)
	}
	if got := n.String(); got != "Break" {
		t.Errorf("String() = %q, want %q", got, "Break")
	}
	err := n.Execute(nil, nil)
	var breakErr *breakError
	if !errors.As(err, &breakErr) {
		t.Errorf("Execute() error type = %T, want *breakError", err)
	}
}

func TestContinueNode(t *testing.T) {
	n := &continueNode{Line: 2, Col: 3}
	line, col := n.Position()
	if line != 2 || col != 3 {
		t.Errorf("Position() = (%d, %d), want (2, 3)", line, col)
	}
	if got := n.String(); got != "Continue" {
		t.Errorf("String() = %q, want %q", got, "Continue")
	}
	err := n.Execute(nil, nil)
	var contErr *continueError
	if !errors.As(err, &contErr) {
		t.Errorf("Execute() error type = %T, want *continueError", err)
	}
}

func TestPropertyAccessNode(t *testing.T) {
	type User struct{ Name string }
	n := newPropertyAccessNode(
		newVariableNode("user", 1, 1), "Name", 1, 1)
	ctx := newRenderContext(map[string]any{
		"user": User{Name: "Alice"},
	})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if got := val.String(); got != "Alice" {
		t.Errorf("Evaluate() = %q, want %q", got, "Alice")
	}
}

func TestSubscriptNode(t *testing.T) {
	n := newSubscriptNode(
		newVariableNode("items", 1, 1),
		newLiteralNode(1, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{
		"items": []string{"a", "b", "c"},
	})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if got := val.String(); got != "b" {
		t.Errorf("Evaluate() = %q, want %q", got, "b")
	}
}

func TestFilterNode(t *testing.T) {
	n := newFilterNode(
		newLiteralNode("hello", 1, 1),
		"upper", nil, 1, 1)
	ctx := newRenderContext(map[string]any{})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if got := val.String(); got != "HELLO" {
		t.Errorf("Evaluate() = %q, want %q", got, "HELLO")
	}
}

func TestFilterNodeNotFound(t *testing.T) {
	n := newFilterNode(
		newLiteralNode("hello", 1, 1),
		"nonexistent_xyz", nil, 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrFilterNotFound) {
		t.Errorf("error = %v, want %v", err, ErrFilterNotFound)
	}
}

func TestForNodeKeyValue(t *testing.T) {
	n := &forNode{
		Vars:       []string{"k", "v"},
		Collection: newLiteralNode(map[string]any{"x": 1}, 1, 1),
		Body: []node{
			newOutputNode(newVariableNode("k", 1, 1), 1, 1),
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("forNode.Execute() error: %v", err)
	}
	if got := buf.String(); got != "x" {
		t.Errorf("forNode.Execute() = %q, want %q", got, "x")
	}
}

func TestForNodeLoopContext(t *testing.T) {
	n := &forNode{
		Vars:       []string{"item"},
		Collection: newLiteralNode([]any{"a", "b"}, 1, 1),
		Body:       []node{newTextNode(".", 1, 1)},
		Line:       1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("forNode.Execute() error: %v", err)
	}
	// After loop, loop context should still be accessible
	if got := buf.String(); got != ".." {
		t.Errorf("forNode.Execute() = %q, want %q", got, "..")
	}
}

func TestForNodeString(t *testing.T) {
	n := &forNode{
		Vars:       []string{"item"},
		Collection: newVariableNode("items", 1, 1),
		Line:       1, Col: 1,
	}
	want := "For([item] in Var(items))"
	if got := n.String(); got != want {
		t.Errorf("forNode.String() = %q, want %q", got, want)
	}
}

func TestOutputNodeString(t *testing.T) {
	n := newOutputNode(newVariableNode("x", 1, 1), 1, 1)
	want := "Output(Var(x))"
	if got := n.String(); got != want {
		t.Errorf("outputNode.String() = %q, want %q", got, want)
	}
}

func TestFilterNodeString(t *testing.T) {
	t.Run("without args", func(t *testing.T) {
		n := newFilterNode(newVariableNode("x", 1, 1), "upper", nil, 1, 1)
		want := "Filter(Var(x)|upper)"
		if got := n.String(); got != want {
			t.Errorf("filterNode.String() = %q, want %q", got, want)
		}
	})

	t.Run("with args", func(t *testing.T) {
		args := []expression{newLiteralNode(10.0, 1, 1)}
		n := newFilterNode(newVariableNode("x", 1, 1), "add", args, 1, 1)
		if got := n.String(); got == "" {
			t.Error("filterNode.String() returned empty string")
		}
	})
}

func TestUnaryOpNodePlus(t *testing.T) {
	n := newUnaryOpNode("+", newLiteralNode(5.0, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	f, _ := val.Float()
	if f != 5.0 {
		t.Errorf("Evaluate() = %v, want 5", f)
	}
}

func TestBinaryOpNodeModulo(t *testing.T) {
	n := newBinaryOpNode("%",
		newLiteralNode(10, 1, 1),
		newLiteralNode(3, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	i, _ := val.Int()
	if i != 1 {
		t.Errorf("Evaluate() = %v, want 1", i)
	}
}

func TestSubscriptNodeStringKey(t *testing.T) {
	n := newSubscriptNode(
		newVariableNode("dict", 1, 1),
		newLiteralNode("key", 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{
		"dict": map[string]any{"key": "val"},
	})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if got := val.String(); got != "val" {
		t.Errorf("Evaluate() = %q, want %q", got, "val")
	}
}

func TestIfNodeElif(t *testing.T) {
	n := &ifNode{
		Branches: []ifBranch{
			{
				Condition: newLiteralNode(false, 1, 1),
				Body:      []node{newTextNode("first", 1, 1)},
			},
			{
				Condition: newLiteralNode(true, 1, 1),
				Body:      []node{newTextNode("second", 1, 1)},
			},
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "second" {
		t.Errorf("Execute() = %q, want %q", got, "second")
	}
}

func TestExecuteBody(t *testing.T) {
	body := []node{
		newTextNode("a", 1, 1),
		newTextNode("b", 1, 2),
		newTextNode("c", 1, 3),
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := executeBody(body, ctx, &buf); err != nil {
		t.Fatalf("executeBody() error: %v", err)
	}
	if got := buf.String(); got != "abc" {
		t.Errorf("executeBody() wrote %q, want %q", got, "abc")
	}
}

func TestExecuteBodySkipsNonStatements(t *testing.T) {
	// literalNode implements expression but not statement.
	body := []node{
		newTextNode("a", 1, 1),
		newLiteralNode("skip", 1, 1),
		newTextNode("b", 1, 2),
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := executeBody(body, ctx, &buf); err != nil {
		t.Fatalf("executeBody() error: %v", err)
	}
	if got := buf.String(); got != "ab" {
		t.Errorf("executeBody() wrote %q, want %q", got, "ab")
	}
}

func TestBreakErrorMessage(t *testing.T) {
	err := &breakError{}
	want := ErrBreakOutsideLoop.Error()
	if got := err.Error(); got != want {
		t.Errorf("breakError.Error() = %q, want %q", got, want)
	}
}

func TestContinueErrorMessage(t *testing.T) {
	err := &continueError{}
	want := ErrContinueOutsideLoop.Error()
	if got := err.Error(); got != want {
		t.Errorf("continueError.Error() = %q, want %q", got, want)
	}
}

func TestBinaryOpNodeSubtractTypeError(t *testing.T) {
	n := newBinaryOpNode("-",
		newLiteralNode("a", 1, 1),
		newLiteralNode(1.0, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotSubtractTypes) {
		t.Errorf("error = %v, want %v", err, ErrCannotSubtractTypes)
	}
}

func TestBinaryOpNodeMultiplyTypeError(t *testing.T) {
	n := newBinaryOpNode("*",
		newLiteralNode("a", 1, 1),
		newLiteralNode(1.0, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotMultiplyTypes) {
		t.Errorf("error = %v, want %v", err, ErrCannotMultiplyTypes)
	}
}

func TestBinaryOpNodeDivideTypeError(t *testing.T) {
	n := newBinaryOpNode("/",
		newLiteralNode("a", 1, 1),
		newLiteralNode(1.0, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotDivideTypes) {
		t.Errorf("error = %v, want %v", err, ErrCannotDivideTypes)
	}
}

func TestBinaryOpNodeModuloTypeError(t *testing.T) {
	n := newBinaryOpNode("%",
		newLiteralNode("a", 1, 1),
		newLiteralNode(1, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotModuloTypes) {
		t.Errorf("error = %v, want %v", err, ErrCannotModuloTypes)
	}
}

func TestUnaryOpNodeNegateTypeError(t *testing.T) {
	n := newUnaryOpNode("-", newLiteralNode("text", 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotNegate) {
		t.Errorf("error = %v, want %v", err, ErrCannotNegate)
	}
}

func TestUnaryOpNodePlusTypeError(t *testing.T) {
	n := newUnaryOpNode("+", newLiteralNode("text", 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotApplyUnaryPlus) {
		t.Errorf("error = %v, want %v", err, ErrCannotApplyUnaryPlus)
	}
}

func TestForNodeMapSingleVar(t *testing.T) {
	n := &forNode{
		Vars:       []string{"k"},
		Collection: newLiteralNode(map[string]any{"x": 1}, 1, 1),
		Body: []node{
			newOutputNode(newVariableNode("k", 1, 1), 1, 1),
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "x" {
		t.Errorf("Execute() = %q, want %q", got, "x")
	}
}

func TestForNodeBreak(t *testing.T) {
	n := &forNode{
		Vars:       []string{"item"},
		Collection: newLiteralNode([]any{"a", "b", "c"}, 1, 1),
		Body: []node{
			newOutputNode(newVariableNode("item", 1, 1), 1, 1),
			&breakNode{Line: 1, Col: 1},
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "a" {
		t.Errorf("Execute() = %q, want %q", got, "a")
	}
}

func TestForNodeContinue(t *testing.T) {
	n := &forNode{
		Vars:       []string{"item"},
		Collection: newLiteralNode([]any{"a", "b", "c"}, 1, 1),
		Body: []node{
			&continueNode{Line: 1, Col: 1},
			newOutputNode(newVariableNode("item", 1, 1), 1, 1),
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "" {
		t.Errorf("Execute() = %q, want empty", got)
	}
}

func TestPropertyAccessNodeString(t *testing.T) {
	n := newPropertyAccessNode(
		newVariableNode("user", 1, 1), "Name", 1, 1)
	want := "PropAccess(Var(user).Name)"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestSubscriptNodeString(t *testing.T) {
	n := newSubscriptNode(
		newVariableNode("items", 1, 1),
		newLiteralNode(0, 1, 1), 1, 1)
	want := "Subscript(Var(items)[Literal(0)])"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestBinaryOpNodeString(t *testing.T) {
	n := newBinaryOpNode("+",
		newVariableNode("a", 1, 1),
		newVariableNode("b", 1, 1), 1, 1)
	want := "BinOp(Var(a) + Var(b))"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestUnaryOpNodeString(t *testing.T) {
	n := newUnaryOpNode("not",
		newVariableNode("x", 1, 1), 1, 1)
	want := "UnaryOp(not Var(x))"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestLiteralNodeString(t *testing.T) {
	n := newLiteralNode(42, 1, 1)
	want := "Literal(42)"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestIfNodeString(t *testing.T) {
	n := &ifNode{
		Branches: []ifBranch{
			{Condition: newLiteralNode(true, 1, 1)},
			{Condition: newLiteralNode(false, 1, 1)},
		},
		Line: 1, Col: 1,
	}
	want := "If(2 branches)"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestForNodeLoopContextFields(t *testing.T) {
	capture := &captureLoopStmt{}
	n := &forNode{
		Vars:       []string{"item"},
		Collection: newLiteralNode([]any{"a", "b", "c"}, 1, 1),
		Body:       []node{capture},
		Line:       1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if capture.last == nil {
		t.Fatal("loop context was not captured during execution")
	}
	if _, ok := ctx.Get("loop"); ok {
		t.Fatal("loop context leaked after execution")
	}
	if capture.last.Length != 3 {
		t.Errorf("Length = %d, want 3", capture.last.Length)
	}
	if capture.last.Index != 2 {
		t.Errorf("Index = %d, want 2", capture.last.Index)
	}
	if capture.last.Counter != 3 {
		t.Errorf("Counter = %d, want 3", capture.last.Counter)
	}
	if !capture.last.Last {
		t.Error("Last = false, want true")
	}
}

// =============================================================================
// Edge Case Tests for Coverage
// =============================================================================

func TestOutputNodeEvaluateError(t *testing.T) {
	mockErr := errMockEval
	n := newOutputNode(&errExpr{err: mockErr}, 1, 1)
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	err := n.Execute(ctx, &buf)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "eval failed" {
		t.Errorf("error = %q, want %q", err.Error(), "eval failed")
	}
}

func TestOutputNodeWriteError(t *testing.T) {
	n := newOutputNode(newLiteralNode("hello", 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	writeErr := errMockWrite
	err := n.Execute(ctx, &errWriter{err: writeErr})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "write failed" {
		t.Errorf("error = %q, want %q", err.Error(), "write failed")
	}
}

func TestPropertyAccessNodeEvaluateError(t *testing.T) {
	mockErr := errMockObjectEval
	n := newPropertyAccessNode(&errExpr{err: mockErr}, "name", 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPropertyAccessNodeFieldNotFound(t *testing.T) {
	// Access a non-existent field on a primitive value.
	n := newPropertyAccessNode(newLiteralNode(42, 1, 1), "foo", 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if err == nil {
		t.Fatal("expected error for field access on non-struct")
	}
}

func TestSubscriptNodeObjectError(t *testing.T) {
	mockErr := errMockObjectEval
	n := newSubscriptNode(&errExpr{err: mockErr}, newLiteralNode(0, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSubscriptNodeIndexError(t *testing.T) {
	mockErr := errMockIndexEval
	n := newSubscriptNode(newLiteralNode([]int{1, 2, 3}, 1, 1), &errExpr{err: mockErr}, 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSubscriptNodeStringKeyFallback(t *testing.T) {
	// Map subscripts use map-key semantics, so string keys do not pass
	// through integer-index conversion.
	m := map[string]any{"name": "Alice"}
	n := newSubscriptNode(
		newLiteralNode(m, 1, 1),
		newLiteralNode("name", 1, 1),
		1, 1,
	)
	ctx := newRenderContext(map[string]any{})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if got := val.String(); got != "Alice" {
		t.Errorf("value = %q, want %q", got, "Alice")
	}
}

func TestBinaryOpNodeLeftError(t *testing.T) {
	mockErr := errMockLeftEval
	n := newBinaryOpNode("+", &errExpr{err: mockErr}, newLiteralNode(1, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBinaryOpNodeRightError(t *testing.T) {
	mockErr := errMockRightEval
	n := newBinaryOpNode("+", newLiteralNode(1, 1, 1), &errExpr{err: mockErr}, 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBinaryOpNodeStringConcatFallback(t *testing.T) {
	// "+" with string left, numeric right — falls back to string concatenation.
	n := newBinaryOpNode("+", newLiteralNode("count: ", 1, 1), newLiteralNode(42, 1, 1), 1, 1)
	ctx := newRenderContext(map[string]any{})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if got := val.String(); got != "count: 42" {
		t.Errorf("value = %q, want %q", got, "count: 42")
	}
}

func TestBinaryOpNodeComparisonSuccess(t *testing.T) {
	// Comparison operators with various types to cover all branches.
	tests := []struct {
		op       string
		left     any
		right    any
		expected bool
	}{
		{"<", 1, 2, true},
		{"<", 2, 1, false},
		{">", 2, 1, true},
		{">", 1, 2, false},
		{"<=", 1, 1, true},
		{"<=", 2, 1, false},
		{">=", 1, 1, true},
		{">=", 1, 2, false},
		// String comparison fallback.
		{"<", "apple", "banana", true},
		{">", "banana", "apple", true},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%v_%s_%v", tt.left, tt.op, tt.right)
		t.Run(name, func(t *testing.T) {
			n := newBinaryOpNode(tt.op,
				newLiteralNode(tt.left, 1, 1),
				newLiteralNode(tt.right, 1, 1),
				1, 1)
			ctx := newRenderContext(map[string]any{})
			val, err := n.Evaluate(ctx)
			if err != nil {
				t.Fatalf("Evaluate() error: %v", err)
			}
			if got := val.IsTrue(); got != tt.expected {
				t.Errorf("%v %s %v = %v, want %v", tt.left, tt.op, tt.right, got, tt.expected)
			}
		})
	}
}

func TestUnaryOpNodeOperandError(t *testing.T) {
	mockErr := errMockOperand
	n := newUnaryOpNode("not", &errExpr{err: mockErr}, 1, 1)
	ctx := newRenderContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestForNodeCollectionError(t *testing.T) {
	mockErr := errMockCollection
	n := &forNode{
		Vars:       []string{"item"},
		Collection: &errExpr{err: mockErr},
		Body:       nil,
		Line:       1,
		Col:        1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	err := n.Execute(ctx, &buf)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestForNodeBodyError(t *testing.T) {
	// Body statement returns a non-break/continue error.
	n := &forNode{
		Vars:       []string{"item"},
		Collection: newLiteralNode([]int{1, 2}, 1, 1),
		Body: []node{
			// Use an outputNode with an errExpr to trigger error in body.
			newOutputNode(&errExpr{err: errMockBody}, 1, 1),
		},
		Line: 1,
		Col:  1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	err := n.Execute(ctx, &buf)
	if err == nil {
		t.Fatal("expected error from body execution")
	}
}

func TestForNodeNestedLoop(t *testing.T) {
	// Nested for loops to verify parent loop context restoration.
	tmpl := `{% for i in outer %}{% for j in inner %}{{ i }}-{{ j }} {% endfor %}{% endfor %}`
	compiled, err := parseSourceTemplate(tmpl)
	if err != nil {
		t.Fatalf("parseSourceTemplate() error: %v", err)
	}
	result, err := compiled.Render(map[string]any{
		"outer": []int{1, 2},
		"inner": []string{"a", "b"},
	})
	if err != nil {
		t.Fatalf("renderSourceTemplate() error: %v", err)
	}
	expected := "1-a 1-b 2-a 2-b "
	if result != expected {
		t.Errorf("result = %q, want %q", result, expected)
	}
}

func TestForNodeEmptyCollection(t *testing.T) {
	n := &forNode{
		Vars:       []string{"item"},
		Collection: newLiteralNode([]int{}, 1, 1),
		Body:       []node{newTextNode("x", 1, 1)},
		Line:       1,
		Col:        1,
	}
	var buf bytes.Buffer
	ctx := newRenderContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if buf.String() != "" {
		t.Errorf("output = %q, want empty", buf.String())
	}
}
