package template

import (
	"testing"
)

// Phase G cycle 1: FormatHTML auto-escapes {{ expr }} output.
func TestFormatHTML_VariableOutputIsEscaped(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `<p>{{ title }}</p>`,
		})),
		WithFormat(FormatHTML),
	)
	got, err := engine.Render("a.html", Data{"title": "<script>alert(1)</script>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := `<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Phase G cycle 2: FormatText does NOT escape output.
func TestFormatText_VariableOutputIsNotEscaped(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `{{ title }}`,
		})),
		WithFormat(FormatText),
	)
	got, err := engine.Render("a.txt", Data{"title": "<script>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "<script>" {
		t.Errorf("got %q, want <script>", got)
	}
}

// Phase G cycle 3: "| safe" skips auto-escape in an HTML engine.
func TestFormatHTML_SafeFilter_SkipsEscape(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `{{ content | safe }}`,
		})),
		WithFormat(FormatHTML),
		WithLayout(),
	)
	got, err := engine.Render("a.html", Data{"content": "<b>bold</b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "<b>bold</b>" {
		t.Errorf("got %q", got)
	}
}

// Phase G cycle 4: safe status is downgraded by non-safe-aware filters.
// {{ x | safe | upper }} should escape because upper is not safe-aware.
func TestFormatHTML_SafeThenNonSafeFilter_Downgrades(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `{{ x | safe | upper }}`,
		})),
		WithFormat(FormatHTML),
		WithLayout(),
	)
	got, err := engine.Render("a.html", Data{"x": "<b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "&lt;B&gt;" {
		t.Errorf("got %q, want &lt;B&gt;", got)
	}
}

// Phase G cycle 5: safe at the terminal position keeps the value safe.
func TestFormatHTML_UpperThenSafe_IsSafe(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `{{ x | upper | safe }}`,
		})),
		WithFormat(FormatHTML),
		WithLayout(),
	)
	got, err := engine.Render("a.html", Data{"x": "<b>"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "<B>" {
		t.Errorf("got %q", got)
	}
}

// Phase G cycle 6: escape filter output is idempotent in a chain.
func TestFormatHTML_EscapeIdempotent(t *testing.T) {
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
	// escape filter returns SafeString, so the output-level auto-escape
	// does not double-escape.
	if got != "&lt;b&gt;" {
		t.Errorf("got %q", got)
	}
}

func TestFormatHTML_EscapeOnceIsSafe(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `{{ x | escape_once }}`,
		})),
		WithFormat(FormatHTML),
	)
	got, err := engine.Render("a.html", Data{"x": "&lt;b&gt; & <i>"})
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	want := "&lt;b&gt; &amp; &lt;i&gt;"
	if got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

// Phase G cycle 7: quotes in attribute context are escaped.
func TestFormatHTML_AttributeQuoteEscaped(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `<a href="{{ url }}">link</a>`,
		})),
		WithFormat(FormatHTML),
	)
	got, err := engine.Render("a.html", Data{"url": `" onclick="x`})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// Quotes must be escaped to prevent attribute injection.
	want := `<a href="&#34; onclick=&#34;x">link</a>`
	if got != want {
		// Accept either &#34; or &quot; encoding.
		altWant := `<a href="&quot; onclick=&quot;x">link</a>`
		if got != altWant {
			t.Errorf("got %q, want %q or %q", got, want, altWant)
		}
	}
}
