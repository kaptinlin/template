package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/kaptinlin/template"
)

var errWrite = errors.New("write failed")
var errEvaluate = errors.New("evaluate failed")

type literalExpr struct {
	value any
	err   error
}

func (e literalExpr) Position() (int, int) { return 0, 0 }
func (e literalExpr) String() string       { return fmt.Sprintf("Literal(%v)", e.value) }
func (e literalExpr) Evaluate(*template.RenderContext) (*template.Value, error) {
	if e.err != nil {
		return nil, e.err
	}
	return template.NewValue(e.value), nil
}

func TestMainWritesSetTagResultToStdout(t *testing.T) {
	got := captureStdout(t, main)
	if got != "Hello, World!\n" {
		t.Fatalf("main() output = %q, want %q", got, "Hello, World!\n")
	}
}

func TestRunPrintsSetTagResult(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := run(&buf); err != nil {
		t.Fatalf("run() err = %v", err)
	}
	if got := buf.String(); got != "Hello, World!\n" {
		t.Fatalf("run() output = %q, want %q", got, "Hello, World!\n")
	}
}

func TestSetNode(t *testing.T) {
	t.Parallel()

	node := &SetNode{VarName: "greeting", Expr: literalExpr{value: "hello"}, Line: 2, Col: 3}

	line, col := node.Position()
	if line != 2 || col != 3 {
		t.Fatalf("Position() = %d, %d; want 2, 3", line, col)
	}
	if got := node.String(); got != "Set(greeting)" {
		t.Fatalf("String() = %q, want %q", got, "Set(greeting)")
	}

	ctx := template.NewRenderContext(nil)
	if err := node.Execute(ctx, nil); err != nil {
		t.Fatalf("Execute() err = %v", err)
	}
	got, ok := ctx.Get("greeting")
	if !ok {
		t.Fatal("RenderContext.Get(greeting) ok = false, want true")
	}
	if got != "hello" {
		t.Fatalf("RenderContext.Get(greeting) = %q, want %q", got, "hello")
	}
}

func TestSetNodeExecuteReturnsExpressionError(t *testing.T) {
	t.Parallel()

	node := &SetNode{VarName: "greeting", Expr: literalExpr{err: errEvaluate}}

	if err := node.Execute(template.NewRenderContext(nil), nil); !errors.Is(err, errEvaluate) {
		t.Fatalf("Execute() err = %v, want %v", err, errEvaluate)
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errWrite }

func TestRunReturnsWriterError(t *testing.T) {
	t.Parallel()

	if err := run(errWriter{}); !errors.Is(err, errWrite) {
		t.Fatalf("run() err = %v, want %v", err, errWrite)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() err = %v", err)
	}

	old := os.Stdout
	os.Stdout = write
	t.Cleanup(func() { os.Stdout = old })

	fn()

	if err := write.Close(); err != nil {
		t.Fatalf("Close() err = %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, read); err != nil {
		t.Fatalf("io.Copy() err = %v", err)
	}
	if err := read.Close(); err != nil {
		t.Fatalf("Close() err = %v", err)
	}

	return buf.String()
}

func TestSetTagRejectsMalformedSyntax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
	}{
		{name: "missing variable", source: `{% set %}`},
		{name: "missing equals", source: `{% set greeting "Hello" %}`},
		{name: "invalid expression", source: `{% set greeting = %}`},
		{name: "extra tokens", source: `{% set greeting = "Hello" "World" %}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			engine := template.New()
			if err := registerSetTag(engine); err != nil {
				t.Fatalf("registerSetTag() err = %v", err)
			}
			if _, err := engine.ParseString(tc.source); err == nil {
				t.Fatal("ParseString() err = nil, want error")
			}
		})
	}
}

func TestRunMainCallsFatalOnError(t *testing.T) {
	t.Parallel()

	var got []any
	runMain(errWriter{}, func(args ...any) { got = args })
	if len(got) != 1 {
		t.Fatalf("fatal args = %v, want one error", got)
	}
	if err, ok := got[0].(error); !ok || !errors.Is(err, errWrite) {
		t.Fatalf("fatal args = %v, want %v", got, errWrite)
	}
}
