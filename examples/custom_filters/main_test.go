package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
)

var errWrite = errors.New("write failed")

func TestMainWritesRepeatedWordToStdout(t *testing.T) {
	got := captureStdout(t, main)
	if got != "hahaha\n" {
		t.Fatalf("main() output = %q, want %q", got, "hahaha\n")
	}
}

func TestRunPrintsRepeatedWord(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := run(&buf); err != nil {
		t.Fatalf("run() err = %v", err)
	}
	if got := buf.String(); got != "hahaha\n" {
		t.Fatalf("run() output = %q, want %q", got, "hahaha\n")
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
