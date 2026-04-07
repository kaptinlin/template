package template

import (
	"bytes"
	"errors"
	"testing"
)

// Phase B cycle 1: NewTextSet constructs and renders a basic template.
func TestNewTextSet_RenderBasicInterpolation(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"hello.txt": "Hello {{ name }}",
	}))

	var buf bytes.Buffer
	if err := set.Render("hello.txt", Context{"name": "world"}, &buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	if got := buf.String(); got != "Hello world" {
		t.Errorf("Render() = %q, want %q", got, "Hello world")
	}
}

// Phase B cycle 2: Set.Get propagates loader not-found errors.
func TestSet_Get_PropagatesNotFound(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{}))
	_, err := set.Get("missing.txt")
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("err = %v, want ErrTemplateNotFound", err)
	}
}

// Phase B cycle 3: Set.Get propagates parse errors.
func TestSet_Get_PropagatesParseError(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"bad.txt": "{% if %}unterminated",
	}))
	_, err := set.Get("bad.txt")
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

// Phase B cycle 4: RenderString convenience.
func TestSet_RenderString(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"t.txt": "a={{ a }} b={{ b }}",
	}))
	got, err := set.RenderString("t.txt", Context{"a": 1, "b": 2})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "a=1 b=2" {
		t.Errorf("got %q", got)
	}
}

// Phase B cycle 5: Set.Get caches compiled templates.
func TestSet_Get_Caches(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": "hello",
	}))
	first, err := set.Get("a.txt")
	if err != nil {
		t.Fatalf("first Get err = %v", err)
	}
	second, err := set.Get("a.txt")
	if err != nil {
		t.Fatalf("second Get err = %v", err)
	}
	if first != second {
		t.Errorf("cache miss: first=%p second=%p", first, second)
	}
}

// Phase B cycle 6: Set.Reset evicts cache.
func TestSet_Reset_EvictsCache(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": "hello",
	}))
	first, err := set.Get("a.txt")
	if err != nil {
		t.Fatalf("first Get err = %v", err)
	}
	set.Reset()
	second, err := set.Get("a.txt")
	if err != nil {
		t.Fatalf("second Get err = %v", err)
	}
	if first == second {
		t.Errorf("Reset() did not evict cache")
	}
}
