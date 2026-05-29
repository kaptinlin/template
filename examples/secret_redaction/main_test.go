package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

var errWrite = errors.New("write failed")

func TestMainWritesRedactedOutputToStdout(t *testing.T) {
	got := captureStdout(t, main)
	if !strings.Contains(got, "audit user=alice token="+redacted) {
		t.Fatalf("main() output missing writer redaction in:\n%s", got)
	}
	if strings.Contains(got, secretToken) {
		t.Fatalf("main() output leaked secret token in:\n%s", got)
	}
}

func TestRunRedactsOutputAndError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := run(&buf); err != nil {
		t.Fatalf("run() err = %v", err)
	}

	got := buf.String()
	checks := []string{
		"writer redaction:",
		"audit user=alice token=" + redacted,
		"filter redaction:",
		"filtered user=alice token=" + redacted,
		"safe render error:",
		"broken.txt:",
		"filter not found: missing_filter",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("run() output missing %q in:\n%s", check, got)
		}
	}
	if strings.Contains(got, secretToken) {
		t.Fatalf("run() output leaked secret token in:\n%s", got)
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errWrite }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }

type failOnMarker struct {
	marker string
	buf    bytes.Buffer
}

func (w *failOnMarker) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte(w.marker)) {
		return 0, errWrite
	}
	return w.buf.Write(p)
}

func TestRedactingWriterRedactsAndReportsInputLength(t *testing.T) {
	t.Parallel()

	input := []byte("token=" + secretToken)
	var buf bytes.Buffer
	w := &redactingWriter{
		inner:   &buf,
		secrets: []string{"", secretToken},
		replace: redacted,
	}

	n, err := w.Write(input)
	if err != nil {
		t.Fatalf("Write() err = %v", err)
	}
	if n != len(input) {
		t.Fatalf("Write() n = %d, want %d", n, len(input))
	}
	if got := buf.String(); got != "token="+redacted {
		t.Fatalf("Write() output = %q, want redacted token", got)
	}
}

func TestRedactingWriterReturnsInnerError(t *testing.T) {
	t.Parallel()

	w := &redactingWriter{inner: errWriter{}, secrets: []string{secretToken}, replace: redacted}
	if _, err := w.Write([]byte(secretToken)); !errors.Is(err, errWrite) {
		t.Fatalf("Write() err = %v, want %v", err, errWrite)
	}
}

func TestRedactingWriterRejectsShortWrite(t *testing.T) {
	t.Parallel()

	w := &redactingWriter{inner: shortWriter{}, secrets: []string{secretToken}, replace: redacted}
	if _, err := w.Write([]byte(secretToken)); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("Write() err = %v, want %v", err, io.ErrShortWrite)
	}
}

func TestRunReturnsErrorsFromOutputStages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		marker string
	}{
		{name: "filter heading", marker: "filter redaction:"},
		{name: "filtered render", marker: "filtered user="},
		{name: "error heading", marker: "safe render error:"},
		{name: "render diagnostic", marker: "broken.txt:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &failOnMarker{marker: tt.marker}
			if err := run(w); !errors.Is(err, errWrite) {
				t.Fatalf("run() err = %v, want %v", err, errWrite)
			}
		})
	}
}

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
