package template

import (
	"bytes"
	"errors"
	"testing"
)

func TestTextNode(t *testing.T) {
	n := NewTextNode("hello", 1, 5)
	line, col := n.Position()
	if line != 1 || col != 5 {
		t.Errorf("Position() = (%d, %d), want (1, 5)", line, col)
	}
	if got := n.String(); got != `Text("hello")` {
		t.Errorf("String() = %q, want %q", got, `Text("hello")`)
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "hello" {
		t.Errorf("Execute() wrote %q, want %q", got, "hello")
	}
}

func TestOutputNode(t *testing.T) {
	n := NewOutputNode(NewLiteralNode("world", 1, 1), 2, 3)
	line, col := n.Position()
	if line != 2 || col != 3 {
		t.Errorf("Position() = (%d, %d), want (2, 3)", line, col)
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
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
		n := NewLiteralNode(tt.value, 1, 1)
		ctx := NewExecutionContext(map[string]any{})
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
	n := NewVariableNode("name", 3, 7)
	if got := n.String(); got != "Var(name)" {
		t.Errorf("String() = %q, want %q", got, "Var(name)")
	}

	t.Run("defined", func(t *testing.T) {
		ctx := NewExecutionContext(map[string]any{"name": "Alice"})
		val, err := n.Evaluate(ctx)
		if err != nil {
			t.Fatalf("Evaluate() error: %v", err)
		}
		if got := val.String(); got != "Alice" {
			t.Errorf("Evaluate() = %q, want %q", got, "Alice")
		}
	})

	t.Run("undefined", func(t *testing.T) {
		ctx := NewExecutionContext(map[string]any{})
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
			n := NewBinaryOpNode(tt.op,
				NewLiteralNode(tt.left, 1, 1),
				NewLiteralNode(tt.right, 1, 1), 1, 1)
			ctx := NewExecutionContext(map[string]any{})
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
	n := NewBinaryOpNode("+",
		NewLiteralNode("hello ", 1, 1),
		NewLiteralNode("world", 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if got := val.String(); got != "hello world" {
		t.Errorf("Evaluate() = %q, want %q", got, "hello world")
	}
}

func TestBinaryOpNodeDivisionByZero(t *testing.T) {
	n := NewBinaryOpNode("/",
		NewLiteralNode(10.0, 1, 1),
		NewLiteralNode(0.0, 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrDivisionByZero) {
		t.Errorf("error = %v, want %v", err, ErrDivisionByZero)
	}
}

func TestBinaryOpNodeModuloByZero(t *testing.T) {
	n := NewBinaryOpNode("%",
		NewLiteralNode(10, 1, 1),
		NewLiteralNode(0, 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrModuloByZero) {
		t.Errorf("error = %v, want %v", err, ErrModuloByZero)
	}
}

func TestBinaryOpNodeUnsupported(t *testing.T) {
	n := NewBinaryOpNode("^^",
		NewLiteralNode(1.0, 1, 1),
		NewLiteralNode(2.0, 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrUnsupportedOperator) {
		t.Errorf("error = %v, want %v", err, ErrUnsupportedOperator)
	}
}

func TestUnaryOpNode(t *testing.T) {
	t.Run("not", func(t *testing.T) {
		n := NewUnaryOpNode("not", NewLiteralNode(true, 1, 1), 1, 1)
		ctx := NewExecutionContext(map[string]any{})
		val, err := n.Evaluate(ctx)
		if err != nil {
			t.Fatalf("Evaluate() error: %v", err)
		}
		if val.IsTrue() {
			t.Error("Evaluate() = true, want false")
		}
	})

	t.Run("negate", func(t *testing.T) {
		n := NewUnaryOpNode("-", NewLiteralNode(5.0, 1, 1), 1, 1)
		ctx := NewExecutionContext(map[string]any{})
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
		n := NewUnaryOpNode("~", NewLiteralNode(1.0, 1, 1), 1, 1)
		ctx := NewExecutionContext(map[string]any{})
		_, err := n.Evaluate(ctx)
		if !errors.Is(err, ErrUnsupportedUnaryOp) {
			t.Errorf("error = %v, want %v", err, ErrUnsupportedUnaryOp)
		}
	})
}

func TestIfNode(t *testing.T) {
	t.Run("true branch", func(t *testing.T) {
		n := &IfNode{
			Branches: []IfBranch{{
				Condition: NewLiteralNode(true, 1, 1),
				Body:      []Node{NewTextNode("yes", 1, 1)},
			}},
			ElseBody: []Node{NewTextNode("no", 1, 1)},
			Line:     1, Col: 1,
		}
		var buf bytes.Buffer
		ctx := NewExecutionContext(map[string]any{})
		if err := n.Execute(ctx, &buf); err != nil {
			t.Fatalf("Execute() error: %v", err)
		}
		if got := buf.String(); got != "yes" {
			t.Errorf("Execute() = %q, want %q", got, "yes")
		}
	})

	t.Run("else branch", func(t *testing.T) {
		n := &IfNode{
			Branches: []IfBranch{{
				Condition: NewLiteralNode(false, 1, 1),
				Body:      []Node{NewTextNode("yes", 1, 1)},
			}},
			ElseBody: []Node{NewTextNode("no", 1, 1)},
			Line:     1, Col: 1,
		}
		var buf bytes.Buffer
		ctx := NewExecutionContext(map[string]any{})
		if err := n.Execute(ctx, &buf); err != nil {
			t.Fatalf("Execute() error: %v", err)
		}
		if got := buf.String(); got != "no" {
			t.Errorf("Execute() = %q, want %q", got, "no")
		}
	})

	t.Run("no match no else", func(t *testing.T) {
		n := &IfNode{
			Branches: []IfBranch{{
				Condition: NewLiteralNode(false, 1, 1),
				Body:      []Node{NewTextNode("yes", 1, 1)},
			}},
			Line: 1, Col: 1,
		}
		var buf bytes.Buffer
		ctx := NewExecutionContext(map[string]any{})
		if err := n.Execute(ctx, &buf); err != nil {
			t.Fatalf("Execute() error: %v", err)
		}
		if got := buf.String(); got != "" {
			t.Errorf("Execute() = %q, want empty", got)
		}
	})
}

func TestForNode(t *testing.T) {
	n := &ForNode{
		Vars:       []string{"item"},
		Collection: NewLiteralNode([]any{"a", "b", "c"}, 1, 1),
		Body: []Node{
			NewOutputNode(NewVariableNode("item", 1, 1), 1, 1),
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "abc" {
		t.Errorf("Execute() = %q, want %q", got, "abc")
	}
}

func TestBreakNode(t *testing.T) {
	n := &BreakNode{Line: 1, Col: 1}
	line, col := n.Position()
	if line != 1 || col != 1 {
		t.Errorf("Position() = (%d, %d), want (1, 1)", line, col)
	}
	if got := n.String(); got != "Break" {
		t.Errorf("String() = %q, want %q", got, "Break")
	}
	err := n.Execute(nil, nil)
	var breakErr *BreakError
	if !errors.As(err, &breakErr) {
		t.Errorf("Execute() error type = %T, want *BreakError", err)
	}
}

func TestContinueNode(t *testing.T) {
	n := &ContinueNode{Line: 2, Col: 3}
	line, col := n.Position()
	if line != 2 || col != 3 {
		t.Errorf("Position() = (%d, %d), want (2, 3)", line, col)
	}
	if got := n.String(); got != "Continue" {
		t.Errorf("String() = %q, want %q", got, "Continue")
	}
	err := n.Execute(nil, nil)
	var contErr *ContinueError
	if !errors.As(err, &contErr) {
		t.Errorf("Execute() error type = %T, want *ContinueError", err)
	}
}

func TestPropertyAccessNode(t *testing.T) {
	type User struct{ Name string }
	n := NewPropertyAccessNode(
		NewVariableNode("user", 1, 1), "Name", 1, 1)
	ctx := NewExecutionContext(map[string]any{
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
	n := NewSubscriptNode(
		NewVariableNode("items", 1, 1),
		NewLiteralNode(1, 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{
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
	n := NewFilterNode(
		NewLiteralNode("hello", 1, 1),
		"upper", nil, 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	val, err := n.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if got := val.String(); got != "HELLO" {
		t.Errorf("Evaluate() = %q, want %q", got, "HELLO")
	}
}

func TestFilterNodeNotFound(t *testing.T) {
	n := NewFilterNode(
		NewLiteralNode("hello", 1, 1),
		"nonexistent_xyz", nil, 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrFilterNotFound) {
		t.Errorf("error = %v, want %v", err, ErrFilterNotFound)
	}
}

func TestForNodeKeyValue(t *testing.T) {
	n := &ForNode{
		Vars:       []string{"k", "v"},
		Collection: NewLiteralNode(map[string]any{"x": 1}, 1, 1),
		Body: []Node{
			NewOutputNode(NewVariableNode("k", 1, 1), 1, 1),
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("ForNode.Execute() error: %v", err)
	}
	if got := buf.String(); got != "x" {
		t.Errorf("ForNode.Execute() = %q, want %q", got, "x")
	}
}

func TestForNodeLoopContext(t *testing.T) {
	n := &ForNode{
		Vars:       []string{"item"},
		Collection: NewLiteralNode([]any{"a", "b"}, 1, 1),
		Body:       []Node{NewTextNode(".", 1, 1)},
		Line:       1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("ForNode.Execute() error: %v", err)
	}
	// After loop, loop context should still be accessible
	if got := buf.String(); got != ".." {
		t.Errorf("ForNode.Execute() = %q, want %q", got, "..")
	}
}

func TestForNodeString(t *testing.T) {
	n := &ForNode{
		Vars:       []string{"item"},
		Collection: NewVariableNode("items", 1, 1),
		Line:       1, Col: 1,
	}
	want := "For([item] in Var(items))"
	if got := n.String(); got != want {
		t.Errorf("ForNode.String() = %q, want %q", got, want)
	}
}

func TestOutputNodeString(t *testing.T) {
	n := NewOutputNode(NewVariableNode("x", 1, 1), 1, 1)
	want := "Output(Var(x))"
	if got := n.String(); got != want {
		t.Errorf("OutputNode.String() = %q, want %q", got, want)
	}
}

func TestFilterNodeString(t *testing.T) {
	t.Run("without args", func(t *testing.T) {
		n := NewFilterNode(NewVariableNode("x", 1, 1), "upper", nil, 1, 1)
		want := "Filter(Var(x)|upper)"
		if got := n.String(); got != want {
			t.Errorf("FilterNode.String() = %q, want %q", got, want)
		}
	})

	t.Run("with args", func(t *testing.T) {
		args := []Expression{NewLiteralNode(10.0, 1, 1)}
		n := NewFilterNode(NewVariableNode("x", 1, 1), "add", args, 1, 1)
		if got := n.String(); got == "" {
			t.Error("FilterNode.String() returned empty string")
		}
	})
}

func TestUnaryOpNodePlus(t *testing.T) {
	n := NewUnaryOpNode("+", NewLiteralNode(5.0, 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
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
	n := NewBinaryOpNode("%",
		NewLiteralNode(10, 1, 1),
		NewLiteralNode(3, 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
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
	n := NewSubscriptNode(
		NewVariableNode("dict", 1, 1),
		NewLiteralNode("key", 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{
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
	n := &IfNode{
		Branches: []IfBranch{
			{
				Condition: NewLiteralNode(false, 1, 1),
				Body:      []Node{NewTextNode("first", 1, 1)},
			},
			{
				Condition: NewLiteralNode(true, 1, 1),
				Body:      []Node{NewTextNode("second", 1, 1)},
			},
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "second" {
		t.Errorf("Execute() = %q, want %q", got, "second")
	}
}

func TestExecuteBody(t *testing.T) {
	body := []Node{
		NewTextNode("a", 1, 1),
		NewTextNode("b", 1, 2),
		NewTextNode("c", 1, 3),
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := executeBody(body, ctx, &buf); err != nil {
		t.Fatalf("executeBody() error: %v", err)
	}
	if got := buf.String(); got != "abc" {
		t.Errorf("executeBody() wrote %q, want %q", got, "abc")
	}
}

func TestExecuteBodySkipsNonStatements(t *testing.T) {
	// LiteralNode implements Expression but not Statement.
	body := []Node{
		NewTextNode("a", 1, 1),
		NewLiteralNode("skip", 1, 1),
		NewTextNode("b", 1, 2),
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := executeBody(body, ctx, &buf); err != nil {
		t.Fatalf("executeBody() error: %v", err)
	}
	if got := buf.String(); got != "ab" {
		t.Errorf("executeBody() wrote %q, want %q", got, "ab")
	}
}

func TestBreakErrorMessage(t *testing.T) {
	err := &BreakError{}
	want := ErrBreakOutsideLoop.Error()
	if got := err.Error(); got != want {
		t.Errorf("BreakError.Error() = %q, want %q", got, want)
	}
}

func TestContinueErrorMessage(t *testing.T) {
	err := &ContinueError{}
	want := ErrContinueOutsideLoop.Error()
	if got := err.Error(); got != want {
		t.Errorf("ContinueError.Error() = %q, want %q", got, want)
	}
}

func TestBinaryOpNodeSubtractTypeError(t *testing.T) {
	n := NewBinaryOpNode("-",
		NewLiteralNode("a", 1, 1),
		NewLiteralNode(1.0, 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotSubtractTypes) {
		t.Errorf("error = %v, want %v", err, ErrCannotSubtractTypes)
	}
}

func TestBinaryOpNodeMultiplyTypeError(t *testing.T) {
	n := NewBinaryOpNode("*",
		NewLiteralNode("a", 1, 1),
		NewLiteralNode(1.0, 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotMultiplyTypes) {
		t.Errorf("error = %v, want %v", err, ErrCannotMultiplyTypes)
	}
}

func TestBinaryOpNodeDivideTypeError(t *testing.T) {
	n := NewBinaryOpNode("/",
		NewLiteralNode("a", 1, 1),
		NewLiteralNode(1.0, 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotDivideTypes) {
		t.Errorf("error = %v, want %v", err, ErrCannotDivideTypes)
	}
}

func TestBinaryOpNodeModuloTypeError(t *testing.T) {
	n := NewBinaryOpNode("%",
		NewLiteralNode("a", 1, 1),
		NewLiteralNode(1, 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotModuloTypes) {
		t.Errorf("error = %v, want %v", err, ErrCannotModuloTypes)
	}
}

func TestUnaryOpNodeNegateTypeError(t *testing.T) {
	n := NewUnaryOpNode("-", NewLiteralNode("text", 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotNegate) {
		t.Errorf("error = %v, want %v", err, ErrCannotNegate)
	}
}

func TestUnaryOpNodePlusTypeError(t *testing.T) {
	n := NewUnaryOpNode("+", NewLiteralNode("text", 1, 1), 1, 1)
	ctx := NewExecutionContext(map[string]any{})
	_, err := n.Evaluate(ctx)
	if !errors.Is(err, ErrCannotApplyUnaryPlus) {
		t.Errorf("error = %v, want %v", err, ErrCannotApplyUnaryPlus)
	}
}

func TestForNodeMapSingleVar(t *testing.T) {
	n := &ForNode{
		Vars:       []string{"k"},
		Collection: NewLiteralNode(map[string]any{"x": 1}, 1, 1),
		Body: []Node{
			NewOutputNode(NewVariableNode("k", 1, 1), 1, 1),
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "x" {
		t.Errorf("Execute() = %q, want %q", got, "x")
	}
}

func TestForNodeBreak(t *testing.T) {
	n := &ForNode{
		Vars:       []string{"item"},
		Collection: NewLiteralNode([]any{"a", "b", "c"}, 1, 1),
		Body: []Node{
			NewOutputNode(NewVariableNode("item", 1, 1), 1, 1),
			&BreakNode{Line: 1, Col: 1},
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "a" {
		t.Errorf("Execute() = %q, want %q", got, "a")
	}
}

func TestForNodeContinue(t *testing.T) {
	n := &ForNode{
		Vars:       []string{"item"},
		Collection: NewLiteralNode([]any{"a", "b", "c"}, 1, 1),
		Body: []Node{
			&ContinueNode{Line: 1, Col: 1},
			NewOutputNode(NewVariableNode("item", 1, 1), 1, 1),
		},
		Line: 1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if got := buf.String(); got != "" {
		t.Errorf("Execute() = %q, want empty", got)
	}
}

func TestPropertyAccessNodeString(t *testing.T) {
	n := NewPropertyAccessNode(
		NewVariableNode("user", 1, 1), "Name", 1, 1)
	want := "PropAccess(Var(user).Name)"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestSubscriptNodeString(t *testing.T) {
	n := NewSubscriptNode(
		NewVariableNode("items", 1, 1),
		NewLiteralNode(0, 1, 1), 1, 1)
	want := "Subscript(Var(items)[Literal(0)])"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestBinaryOpNodeString(t *testing.T) {
	n := NewBinaryOpNode("+",
		NewVariableNode("a", 1, 1),
		NewVariableNode("b", 1, 1), 1, 1)
	want := "BinOp(Var(a) + Var(b))"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestUnaryOpNodeString(t *testing.T) {
	n := NewUnaryOpNode("not",
		NewVariableNode("x", 1, 1), 1, 1)
	want := "UnaryOp(not Var(x))"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestLiteralNodeString(t *testing.T) {
	n := NewLiteralNode(42, 1, 1)
	want := "Literal(42)"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestIfNodeString(t *testing.T) {
	n := &IfNode{
		Branches: []IfBranch{
			{Condition: NewLiteralNode(true, 1, 1)},
			{Condition: NewLiteralNode(false, 1, 1)},
		},
		Line: 1, Col: 1,
	}
	want := "If(2 branches)"
	if got := n.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestForNodeLoopContextFields(t *testing.T) {
	var captured *LoopContext
	// We'll use a custom approach: iterate and capture loop context.
	n := &ForNode{
		Vars:       []string{"item"},
		Collection: NewLiteralNode([]any{"a", "b", "c"}, 1, 1),
		Body:       []Node{NewTextNode(".", 1, 1)},
		Line:       1, Col: 1,
	}
	var buf bytes.Buffer
	ctx := NewExecutionContext(map[string]any{})
	if err := n.Execute(ctx, &buf); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	// After the loop, check the last loop context set.
	if v, ok := ctx.Get("loop"); ok {
		captured = v.(*LoopContext)
	}
	if captured == nil {
		t.Fatal("loop context not found after execution")
	}
	if captured.Length != 3 {
		t.Errorf("Length = %d, want 3", captured.Length)
	}
	if captured.Index != 2 {
		t.Errorf("Index = %d, want 2", captured.Index)
	}
	if captured.Counter != 3 {
		t.Errorf("Counter = %d, want 3", captured.Counter)
	}
	if !captured.Last {
		t.Error("Last = false, want true")
	}
}
