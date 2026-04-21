package template

import (
	"testing"
)

// Phase K cycle 1: single-level super renders parent + child content.
func TestBlockSuper_OneLevel(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"base.txt":  `{% block x %}PARENT{% endblock %}`,
			"child.txt": `{% extends "base.txt" %}{% block x %}before-{{ block.super }}-after{% endblock %}`,
		})),
		WithFormat(FormatText),
		WithLayout(),
	)
	got, err := engine.Render("child.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "before-PARENT-after" {
		t.Errorf("got %q", got)
	}
}

// Phase K cycle 2: three-level super chain accumulates content.
func TestBlockSuper_ThreeLevelChain(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt":      `{% block x %}A{% endblock %}`,
			"middle.txt": `{% extends "a.txt" %}{% block x %}M({{ block.super }}){% endblock %}`,
			"leaf.txt":   `{% extends "middle.txt" %}{% block x %}L[{{ block.super }}]{% endblock %}`,
		})),
		WithFormat(FormatText),
		WithLayout(),
	)
	got, err := engine.Render("leaf.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "L[M(A)]" {
		t.Errorf("got %q, want L[M(A)]", got)
	}
}

// Phase K cycle 3: block.super in a block with no parent renders empty.
func TestBlockSuper_NoParent_Empty(t *testing.T) {
	t.Parallel()

	// Must be loaded via an engine with FeatureLayout because {% block %}
	// is a layout feature.
	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `[{% block x %}({{ block.super }}){% endblock %}]`,
		})),
		WithFormat(FormatText),
		WithLayout(),
	)
	got, err := engine.Render("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "[()]" {
		t.Errorf("got %q, want [()]", got)
	}
}

// Phase K cycle 4: block.super result is SafeString and not re-escaped
// in FormatHTML rendering.
func TestBlockSuper_FormatHTML_NotReEscaped(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"base.html":  `{% block body %}<em>parent</em>{% endblock %}`,
			"child.html": `{% extends "base.html" %}{% block body %}{{ block.super }}-child{% endblock %}`,
		})),
		WithFormat(FormatHTML),
		WithLayout(),
	)
	got, err := engine.Render("child.html", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// The parent's rendered output contains literal <em>...</em>. When
	// interpolated via block.super, it must NOT be escaped again.
	if got != "<em>parent</em>-child" {
		t.Errorf("got %q", got)
	}
}
