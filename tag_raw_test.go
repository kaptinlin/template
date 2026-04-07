package template

import (
	"errors"
	"strings"
	"testing"
)

// rawSet is a tiny helper: build a TextSet with one template named "a.txt".
func rawSet(src string) *Set {
	return NewTextSet(NewMemoryLoader(map[string]string{"a.txt": src}))
}

// Raw block preserves {{ var }} literally (Set-loaded template).
func TestRaw_PreservesVariableSyntax(t *testing.T) {
	t.Parallel()

	set := rawSet(`{% raw %}{{ foo }}{% endraw %}`)
	got, err := set.RenderString("a.txt", nil)
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

	set := rawSet(`{% raw %}{% for x in items %}{% endraw %}`)
	got, err := set.RenderString("a.txt", nil)
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
	set := rawSet(src)
	got, err := set.RenderString("a.txt", nil)
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

	set := rawSet(`{% raw %}never closes`)
	_, err := set.Get("a.txt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnclosedRaw) && !strings.Contains(err.Error(), "unclosed raw") {
		t.Errorf("err = %v, want ErrUnclosedRaw", err)
	}
}

// Raw in HTMLSet: content is emitted as literal text, not HTML-escaped.
func TestRaw_WorksInHTMLSet_NotEscaped(t *testing.T) {
	t.Parallel()

	set := NewHTMLSet(NewMemoryLoader(map[string]string{
		"a.html": `{% raw %}<b>{{ x }}</b>{% endraw %}`,
	}))
	got, err := set.RenderString("a.html", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "<b>{{ x }}</b>" {
		t.Errorf("got %q", got)
	}
}

// Compile-path templates do NOT know {% raw %} — it's a Set-only feature.
func TestRaw_CompilePath_UnknownTag(t *testing.T) {
	t.Parallel()

	_, err := Compile(`{% raw %}{{ x }}{% endraw %}`)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown tag: raw") {
		t.Errorf("err = %v, want 'unknown tag: raw'", err)
	}
}
