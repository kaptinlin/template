package template

import (
	"errors"
	"strings"
	"testing"
)

// rawEngine is a tiny helper: build a text engine with layout enabled and
// one template named "a.txt".
func rawEngine(src string) *Engine {
	return New(
		WithLoader(NewMemoryLoader(map[string]string{"a.txt": src})),
		WithFormat(FormatText),
		WithLayout(),
	)
}

// Raw block preserves {{ var }} literally in a layout-enabled engine.
func TestRaw_PreservesVariableSyntax(t *testing.T) {
	t.Parallel()

	engine := rawEngine(`{% raw %}{{ foo }}{% endraw %}`)
	got, err := engine.Render("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "{{ foo }}" {
		t.Errorf("got %q, want %q", got, "{{ foo }}")
	}
}

// Raw block preserves {% tag %} literally.
func TestRaw_PreservesTagSyntax(t *testing.T) {
	t.Parallel()

	engine := rawEngine(`{% raw %}{% for x in items %}{% endraw %}`)
	got, err := engine.Render("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "{% for x in items %}" {
		t.Errorf("got %q", got)
	}
}

// Raw supports multiline content.
func TestRaw_Multiline(t *testing.T) {
	t.Parallel()

	src := "{% raw %}line1\nline2\n{{ x }}\nline3{% endraw %}"
	engine := rawEngine(src)
	got, err := engine.Render("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := "line1\nline2\n{{ x }}\nline3"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Unclosed raw block errors.
func TestRaw_UnclosedErrors(t *testing.T) {
	t.Parallel()

	engine := rawEngine(`{% raw %}never closes`)
	_, err := engine.Load("a.txt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnclosedRaw) && !strings.Contains(err.Error(), "unclosed raw") {
		t.Errorf("err = %v, want ErrUnclosedRaw", err)
	}
}

// Raw in an HTML engine: content is emitted as literal text, not HTML-escaped.
func TestRaw_WorksInFormatHTML_NotEscaped(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `{% raw %}<b>{{ x }}</b>{% endraw %}`,
		})),
		WithFormat(FormatHTML),
		WithLayout(),
	)
	got, err := engine.Render("a.html", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "<b>{{ x }}</b>" {
		t.Errorf("got %q", got)
	}
}

// Core parsing does NOT know {% raw %} — it's a FeatureLayout-only feature.
func TestRaw_CoreEngine_UnknownTag(t *testing.T) {
	t.Parallel()

	_, err := parseSourceTemplate(`{% raw %}{{ x }}{% endraw %}`)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown tag: raw") {
		t.Errorf("err = %v, want 'unknown tag: raw'", err)
	}
}
