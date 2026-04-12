package template

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// Phase B cycle 1: Engine constructs and renders a basic text template.
func TestEngine_RenderBasicInterpolation(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"hello.txt": "Hello {{ name }}",
		})),
		WithFormat(FormatText),
	)

	var buf bytes.Buffer
	if err := engine.RenderTo("hello.txt", &buf, Data{"name": "world"}); err != nil {
		t.Fatalf("renderSourceTemplate() err = %v", err)
	}
	if got := buf.String(); got != "Hello world" {
		t.Errorf("renderSourceTemplate() = %q, want %q", got, "Hello world")
	}
}

// Phase B cycle 2: Engine.Load propagates loader not-found errors.
func TestEngine_Load_PropagatesNotFound(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{})),
		WithFormat(FormatText),
	)
	_, err := engine.Load("missing.txt")
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("err = %v, want ErrTemplateNotFound", err)
	}
}

// Phase B cycle 3: Engine.Load propagates parse errors.
func TestEngine_Load_PropagatesParseError(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"bad.txt": "{% if %}unterminated",
		})),
		WithFormat(FormatText),
	)
	_, err := engine.Load("bad.txt")
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("err = %T, want *ParseError", err)
	}
	if !strings.Contains(err.Error(), "bad.txt:") {
		t.Fatalf("err = %q, want template name prefix", err.Error())
	}
}

// Phase B cycle 4: RenderString convenience.
func TestEngine_Render(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"t.txt": "a={{ a }} b={{ b }}",
		})),
		WithFormat(FormatText),
	)
	got, err := engine.Render("t.txt", Data{"a": 1, "b": 2})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "a=1 b=2" {
		t.Errorf("got %q", got)
	}
}

// Phase B cycle 5: Engine.Load caches compiled templates.
func TestEngine_Load_Caches(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": "hello",
		})),
		WithFormat(FormatText),
	)
	first, err := engine.Load("a.txt")
	if err != nil {
		t.Fatalf("first Load err = %v", err)
	}
	second, err := engine.Load("a.txt")
	if err != nil {
		t.Fatalf("second Load err = %v", err)
	}
	if first != second {
		t.Errorf("cache miss: first=%p second=%p", first, second)
	}
}

// Phase B cycle 6: Engine.Reset evicts cache.
func TestEngine_Reset_EvictsCache(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": "hello",
		})),
		WithFormat(FormatText),
	)
	first, err := engine.Load("a.txt")
	if err != nil {
		t.Fatalf("first Load err = %v", err)
	}
	engine.Reset()
	second, err := engine.Load("a.txt")
	if err != nil {
		t.Fatalf("second Load err = %v", err)
	}
	if first == second {
		t.Errorf("Reset() did not evict cache")
	}
}
