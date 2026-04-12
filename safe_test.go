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
// inside engines with FeatureLayout enabled.
func TestSafeFilter_NotRegisteredGlobally(t *testing.T) {
	t.Parallel()

	if _, ok := defaultRegistry.Filter("safe"); ok {
		t.Error("safe filter should not be in the global registry")
	}
}

// The safe filter is available in engines with FeatureLayout enabled.
func TestSafeFilter_AvailableInFeatureLayout(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `{{ x | safe }}`,
		})),
		WithFormat(FormatText),
		WithLayout(),
	)
	got, err := engine.Render("a.txt", Data{"x": "<b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "<b>" {
		t.Errorf("got %q", got)
	}
}

// The built-in escape filter returns a plain string.
func TestEscapeFilter_GlobalReturnsString(t *testing.T) {
	t.Parallel()

	fn, ok := defaultRegistry.Filter("escape")
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

// Inside FormatHTML the escape filter is overridden with a SafeString-
// returning variant so the auto-escape pipeline does not double-escape.
func TestEscapeFilter_FormatHTMLReturnsSafeString(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `{{ x | escape }}`,
		})),
		WithFormat(FormatHTML),
	)
	got, err := engine.Render("a.html", Data{"x": "<b>"})
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

// In FormatText, escape is overridden? No — only FormatHTML needs the safe
// variant. FormatText's escape uses the global (plain string) filter via
// parent fallback.
func TestEscapeFilter_FormatTextUsesGlobal(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `{{ x | escape }}`,
		})),
		WithFormat(FormatText),
	)
	got, err := engine.Render("a.txt", Data{"x": "<b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// FormatText does not auto-escape, so escape's string return is
	// written as-is.
	if got != "&lt;b&gt;" {
		t.Errorf("got %q", got)
	}
}

// Built-in filter lookup for "safe" returns ErrFilterNotFound via a
// render that exercises the FilterNode path.
func TestCoreEngine_NoSafeFilter(t *testing.T) {
	t.Parallel()

	_, err := renderSourceTemplate(`{{ x | safe }}`, Data{"x": "<b>"})
	if !errors.Is(err, ErrFilterNotFound) {
		t.Errorf("err = %v, want ErrFilterNotFound", err)
	}
}
