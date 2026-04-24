package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/kaptinlin/template"
)

var errWrite = errors.New("write failed")

func TestMainWritesRenderedListToStdout(t *testing.T) {
	got := captureStdout(t, main)
	if got != "Hello, ALICE!\nItems:\n  0: foo\n  1: bar\n  2: baz\n" {
		t.Fatalf("main() output = %q, want %q", got, "Hello, ALICE!\nItems:\n  0: foo\n  1: bar\n  2: baz\n")
	}
}

func TestRunPrintsRenderedList(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := run(&buf); err != nil {
		t.Fatalf("run() err = %v", err)
	}
	want := "Hello, ALICE!\nItems:\n  0: foo\n  1: bar\n  2: baz\n"
	if got := buf.String(); got != want {
		t.Fatalf("run() output = %q, want %q", got, want)
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

func TestRenderReturnsParseError(t *testing.T) {
	t.Parallel()

	if err := render(&bytes.Buffer{}, `{% if %}`, nil); err == nil {
		t.Fatal("render() err = nil, want parse error")
	}
}

func TestRenderReturnsExecutionError(t *testing.T) {
	t.Parallel()

	if err := render(&bytes.Buffer{}, `{{ name|missing }}`, template.Data{"name": "alice"}); !errors.Is(err, template.ErrFilterNotFound) {
		t.Fatalf("render() err = %v, want ErrFilterNotFound", err)
	}
}
