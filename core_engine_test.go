package template

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

// Phase N cycle 1: a layout-capable HTML engine sets autoescape=true.
func TestHTMLFeatureEngine_ConstructsWithAutoescape(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"x.html": `{{ x }}`,
		})),
		WithFormat(FormatHTML),
		WithLayout(),
	)
	got, _ := engine.Render("x.html", Data{"x": "<b>"})
	if got != "&lt;b&gt;" {
		t.Errorf("got %q", got)
	}
}

// Phase N cycle 2: Reset() evicts cache so reloaded templates pick up changes.
// Uses a mutable loader so we can simulate file edits.
func TestEngine_Reset_ReloadsAfterMutation(t *testing.T) {
	t.Parallel()

	loader := &mutableLoader{files: map[string]string{"a.txt": "v1"}}
	engine := New(
		WithLoader(loader),
		WithFormat(FormatText),
		WithLayout(),
	)

	got1, _ := engine.Render("a.txt", nil)
	if got1 != "v1" {
		t.Fatalf("first render = %q", got1)
	}

	loader.set("a.txt", "v2")
	got2, _ := engine.Render("a.txt", nil)
	if got2 != "v1" {
		t.Errorf("without reset expected cached v1, got %q", got2)
	}

	engine.Reset()
	got3, _ := engine.Render("a.txt", nil)
	if got3 != "v2" {
		t.Errorf("after reset got %q, want v2", got3)
	}
}

// Phase N cycle 3: concurrent Render of a cached template is safe.
// Run with go test -race to validate.
func TestEngine_Concurrent_Render(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `{{ x }}-{{ y }}`,
		})),
		WithFormat(FormatHTML),
		WithLayout(),
	)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := engine.Render("a.html", Data{"x": i, "y": "z"})
			if err != nil {
				t.Errorf("err = %v", err)
			}
		}(i)
	}
	wg.Wait()
}

// Phase N cycle 4: WithDefaults values are visible; render ctx overrides defaults.
func TestEngine_WithDefaults_CtxOverridesGlobals(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `{{ g }}-{{ p }}`,
		})),
		WithFormat(FormatText),
		WithLayout(),
		WithDefaults(Data{"g": "global", "p": "default"}),
	)

	got, err := engine.Render("a.txt", Data{"p": "override"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "global-override" {
		t.Errorf("got %q", got)
	}
}

// Phase N cycle 5: source-string parsing supports basic interpolation.
func TestCoreEngine_BasicInterpolation(t *testing.T) {
	t.Parallel()

	engine := New()
	tpl, err := engine.ParseString(`Hello {{ name }}!`)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got, err := tpl.Render(Data{"name": "World"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "Hello World!" {
		t.Errorf("got %q", got)
	}
}

// Phase N cycle 6: source-string rendering does NOT auto-escape.
func TestCoreEngine_NoAutoEscape(t *testing.T) {
	t.Parallel()

	engine := New()
	tpl, err := engine.ParseString(`{{ x }}`)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got, err := tpl.Render(Data{"x": "<b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "<b>" {
		t.Errorf("got %q, core engine must not auto-escape by default", got)
	}
}

// Source-string parsing does NOT see layout tags unless FeatureLayout is enabled.
func TestCoreEngine_LayoutTagsAreUnknown(t *testing.T) {
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
			_, err := New().ParseString(c.src)
			if err == nil {
				t.Fatal("expected parse error")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want contains %q", err, c.want)
			}
		})
	}
}

// Source-string rendering has no "safe" filter by default.
func TestCoreEngine_NoSafeFilterViaRender(t *testing.T) {
	t.Parallel()

	tpl, err := New().ParseString(`{{ x | safe }}`)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	_, err = tpl.Render(Data{"x": "hi"})
	if !errors.Is(err, ErrFilterNotFound) {
		t.Errorf("err = %v, want ErrFilterNotFound", err)
	}
}

// The built-in escape filter returns plain string (not SafeString).
// This is verified by the built-in escape lookup in safe_test.go.
// Here we additionally prove the result is written to output as-is
// with no double-escaping surprise.
func TestCoreEngine_EscapeFilterRendersPlain(t *testing.T) {
	t.Parallel()

	tpl, err := New().ParseString(`{{ x | escape }}`)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got, err := tpl.Render(Data{"x": "<b>"})
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
