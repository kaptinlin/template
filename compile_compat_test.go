package template

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

// Phase N cycle 1: NewHTMLSet constructor sets autoescape=true.
func TestNewHTMLSet_ConstructsWithAutoescape(t *testing.T) {
	t.Parallel()

	set := NewHTMLSet(NewMemoryLoader(map[string]string{
		"x.html": `{{ x }}`,
	}))
	got, _ := set.RenderString("x.html", Context{"x": "<b>"})
	if got != "&lt;b&gt;" {
		t.Errorf("got %q", got)
	}
}

// Phase N cycle 2: Reset() evicts cache so reloaded templates pick up changes.
// Uses a mutable loader so we can simulate file edits.
func TestSet_Reset_ReloadsAfterMutation(t *testing.T) {
	t.Parallel()

	loader := &mutableLoader{files: map[string]string{"a.txt": "v1"}}
	set := NewTextSet(loader)

	got1, _ := set.RenderString("a.txt", nil)
	if got1 != "v1" {
		t.Fatalf("first render = %q", got1)
	}

	loader.set("a.txt", "v2")
	got2, _ := set.RenderString("a.txt", nil)
	if got2 != "v1" {
		t.Errorf("without reset expected cached v1, got %q", got2)
	}

	set.Reset()
	got3, _ := set.RenderString("a.txt", nil)
	if got3 != "v2" {
		t.Errorf("after reset got %q, want v2", got3)
	}
}

// Phase N cycle 3: concurrent Render of a cached template is safe.
// Run with go test -race to validate.
func TestSet_Concurrent_Render(t *testing.T) {
	t.Parallel()

	set := NewHTMLSet(NewMemoryLoader(map[string]string{
		"a.html": `{{ x }}-{{ y }}`,
	}))

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := set.RenderString("a.html", Context{"x": i, "y": "z"})
			if err != nil {
				t.Errorf("err = %v", err)
			}
		}(i)
	}
	wg.Wait()
}

// Phase N cycle 4: WithGlobals values are visible; render ctx overrides globals.
func TestSet_WithGlobals_CtxOverridesGlobals(t *testing.T) {
	t.Parallel()

	set := NewTextSet(
		NewMemoryLoader(map[string]string{
			"a.txt": `{{ g }}-{{ p }}`,
		}),
		WithGlobals(Context{"g": "global", "p": "default"}),
	)

	got, err := set.RenderString("a.txt", Context{"p": "override"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "global-override" {
		t.Errorf("got %q", got)
	}
}

// Phase N cycle 5: Compile/Render basic interpolation behavior unchanged.
func TestCompileCompat_BasicInterpolation(t *testing.T) {
	t.Parallel()

	tpl, err := Compile(`Hello {{ name }}!`)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got, err := tpl.Render(Context{"name": "World"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "Hello World!" {
		t.Errorf("got %q", got)
	}
}

// Phase N cycle 6: Compile-path does NOT auto-escape (it's a text template).
func TestCompileCompat_NoAutoEscape(t *testing.T) {
	t.Parallel()

	got, err := Render(`{{ x }}`, Context{"x": "<b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "<b>" {
		t.Errorf("got %q, Compile path must not auto-escape", got)
	}
}

// Compile-path templates do NOT see the layout tags (include/extends/
// block) or raw — those are Set-scoped.
func TestCompileCompat_LayoutTagsAreUnknown(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		src  string
		want string
	}{
		{"include", `{% include "x" %}`, "unknown tag: include"},
		{"extends", `{% extends "x" %}`, "unknown tag: extends"},
		{"block", `{% block x %}hi{% endblock %}`, "unknown tag: block"},
		{"raw", `{% raw %}hi{% endraw %}`, "unknown tag: raw"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := Compile(c.src)
			if err == nil {
				t.Fatal("expected parse error")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want contains %q", err, c.want)
			}
		})
	}
}

// Compile-path has no "safe" filter.
func TestCompileCompat_NoSafeFilterViaRender(t *testing.T) {
	t.Parallel()

	_, err := Render(`{{ x | safe }}`, Context{"x": "hi"})
	if !errors.Is(err, ErrFilterNotFound) {
		t.Errorf("err = %v, want ErrFilterNotFound", err)
	}
}

// Compile-path escape filter returns plain string (not SafeString).
// This is verified by the global Filter("escape") in safe_test.go.
// Here we additionally prove the result is written to output as-is
// with no double-escaping surprise.
func TestCompileCompat_EscapeFilterRendersPlain(t *testing.T) {
	t.Parallel()

	got, err := Render(`{{ x | escape }}`, Context{"x": "<b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "&lt;b&gt;" {
		t.Errorf("got %q, want &lt;b&gt;", got)
	}
}

// mutableLoader is a test double whose contents can be mutated.
type mutableLoader struct {
	mu    sync.RWMutex
	files map[string]string
}

func (l *mutableLoader) set(name, src string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.files[name] = src
}

func (l *mutableLoader) Open(name string) (string, string, error) {
	if err := ValidateName(name); err != nil {
		return "", "", err
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	src, ok := l.files[name]
	if !ok {
		return "", "", ErrTemplateNotFound
	}
	return src, name, nil
}
