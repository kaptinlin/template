package template

import (
	"errors"
	"testing"
)

// SafeString is a string-compatible alias; constructing and converting
// to string is a direct cast.
func TestSafeString_StringCompatible(t *testing.T) {
	t.Parallel()

	s := SafeString("hello")
	if string(s) != "hello" {
		t.Errorf("string(SafeString) = %q, want %q", string(s), "hello")
	}
}

// The safe filter is NOT registered globally. It only becomes available
// inside templates loaded via NewHTMLSet or NewTextSet.
func TestSafeFilter_NotRegisteredGlobally(t *testing.T) {
	t.Parallel()

	if _, ok := Filter("safe"); ok {
		t.Error("safe filter should not be in the global registry")
	}
}

// The safe filter is available in Set-loaded templates.
func TestSafeFilter_AvailableInSetScope(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `{{ x | safe }}`,
	}))
	got, err := set.RenderString("a.txt", Context{"x": "<b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "<b>" {
		t.Errorf("got %q", got)
	}
}

// The global escape filter returns plain string (unchanged from the
// pre-layout era), so existing callers that type-assert the result as
// string continue to work.
func TestEscapeFilter_GlobalReturnsString(t *testing.T) {
	t.Parallel()

	fn, ok := Filter("escape")
	if !ok {
		t.Fatal("escape filter missing from global registry")
	}
	result, err := fn("<b>")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if _, ok := result.(string); !ok {
		t.Errorf("escape returned %T, want string", result)
	}
	if _, ok := result.(SafeString); ok {
		t.Error("escape should NOT return SafeString in global scope")
	}
}

// Inside HTMLSet the escape filter is overridden with a SafeString-
// returning variant so the auto-escape pipeline does not double-escape.
func TestEscapeFilter_HTMLSetReturnsSafeString(t *testing.T) {
	t.Parallel()

	set := NewHTMLSet(NewMemoryLoader(map[string]string{
		"a.html": `{{ x | escape }}`,
	}))
	got, err := set.RenderString("a.html", Context{"x": "<b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// If escape returned a plain string, the auto-escape output path
	// would double-escape and produce &amp;lt;b&amp;gt;. SafeString
	// prevents that.
	if got != "&lt;b&gt;" {
		t.Errorf("got %q, want &lt;b&gt;", got)
	}
}

// In TextSet, escape is overridden? No — only HTMLSet needs the safe
// variant. TextSet's escape uses the global (plain string) filter via
// parent fallback.
func TestEscapeFilter_TextSetUsesGlobal(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `{{ x | escape }}`,
	}))
	got, err := set.RenderString("a.txt", Context{"x": "<b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// TextSet does not auto-escape, so escape's string return is
	// written as-is.
	if got != "&lt;b&gt;" {
		t.Errorf("got %q", got)
	}
}

// Global filter lookup for "safe" returns ErrFilterNotFound via a
// render that exercises the FilterNode path.
func TestCompileCompat_NoSafeFilter(t *testing.T) {
	t.Parallel()

	_, err := Render(`{{ x | safe }}`, Context{"x": "<b>"})
	if !errors.Is(err, ErrFilterNotFound) {
		t.Errorf("err = %v, want ErrFilterNotFound", err)
	}
}
