// Package main demonstrates redacting secrets from rendered output.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/kaptinlin/template"
)

const (
	redacted    = "[REDACTED]"
	secretToken = "sk_live_123456"
)

var (
	errExpectedRenderError = errors.New("expected missing filter render error")
	errSecretLeaked        = errors.New("render error leaked the secret token")
)

type redactingWriter struct {
	inner   io.Writer
	secrets []string
	replace string
}

func (w *redactingWriter) Write(p []byte) (int, error) {
	out := p
	for _, secret := range w.secrets {
		if secret == "" {
			continue
		}
		out = bytes.ReplaceAll(out, []byte(secret), []byte(w.replace))
	}

	n, err := w.inner.Write(out)
	if err != nil {
		return 0, err
	}
	if n != len(out) {
		return 0, io.ErrShortWrite
	}
	return len(p), nil
}

func main() {
	runMain(os.Stdout, log.Fatal)
}

func runMain(out io.Writer, fatal func(...any)) {
	if err := run(out); err != nil {
		fatal(err)
	}
}

func run(out io.Writer) error {
	engine := template.New(
		template.WithLoader(template.NewMemoryLoader(map[string]string{
			"audit.txt":    "audit user={{ user }} token={{ token }}\n",
			"filtered.txt": "filtered user={{ user }} token={{ token|redact }}\n",
			"broken.txt":   "broken token={{ token|missing_filter }}\n",
		})),
		template.WithFormat(template.FormatText),
		template.WithFilter("redact", func(_ any, _ ...any) (any, error) {
			return redacted, nil
		}),
	)

	data := template.Data{
		"user":  "alice",
		"token": secretToken,
	}

	if _, err := fmt.Fprintln(out, "writer redaction:"); err != nil {
		return err
	}
	if err := engine.RenderTo("audit.txt", &redactingWriter{
		inner:   out,
		secrets: []string{secretToken},
		replace: redacted,
	}, data); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "filter redaction:"); err != nil {
		return err
	}
	if err := engine.RenderTo("filtered.txt", out, data); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "safe render error:"); err != nil {
		return err
	}
	err := engine.RenderTo("broken.txt", io.Discard, data)
	if err == nil {
		return errExpectedRenderError
	}
	if strings.Contains(err.Error(), secretToken) {
		return errSecretLeaked
	}

	var renderErr *template.RenderError
	if errors.As(err, &renderErr) {
		_, err = fmt.Fprintf(
			out,
			"%s:%d:%d: %v\n",
			renderErr.Template,
			renderErr.Line,
			renderErr.Col,
			renderErr.Cause,
		)
		return err
	}

	_, err = fmt.Fprintln(out, err)
	return err
}
