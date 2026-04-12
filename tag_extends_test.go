package template

import (
	"errors"
	"strings"
	"testing"
)

func newLayoutTextExtendsEngine(loader Loader, opts ...EngineOption) *Engine {
	opts = append([]EngineOption{
		WithLoader(loader),
		WithFormat(FormatText),
		WithLayout(),
	}, opts...)
	return New(opts...)
}

// Phase I cycle 1: basic extends with single block override.
func TestExtends_BasicChild_OverridesBlock(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"base.txt":  `[{% block content %}default{% endblock %}]`,
		"child.txt": `{% extends "base.txt" %}{% block content %}custom{% endblock %}`,
	}))
	got, err := engine.Render("child.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "[custom]" {
		t.Errorf("got %q, want [custom]", got)
	}
}

// Phase I cycle 2: child block not defined falls back to parent default.
func TestExtends_UnoverriddenBlock_UsesParentDefault(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"base.txt":  `[{% block a %}A{% endblock %}][{% block b %}B{% endblock %}]`,
		"child.txt": `{% extends "base.txt" %}{% block a %}aa{% endblock %}`,
	}))
	got, err := engine.Render("child.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "[aa][B]" {
		t.Errorf("got %q", got)
	}
}

// Phase I cycle 3: extends must be the first tag.
func TestExtends_MustBeFirstTag(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"base.txt":  `{% block x %}{% endblock %}`,
		"child.txt": `junk {% extends "base.txt" %}`,
	}))
	_, err := engine.Load("child.txt")
	if !errors.Is(err, ErrExtendsNotFirst) && !strings.Contains(errString(err), "first tag") {
		t.Errorf("err = %v, want ErrExtendsNotFirst", err)
	}
}

// Phase I cycle 4: extends path must be a string literal.
func TestExtends_PathMustBeStringLiteral(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"child.txt": `{% extends parent_var %}`,
	}))
	_, err := engine.Load("child.txt")
	if !errors.Is(err, ErrExtendsPathNotLiteral) {
		t.Errorf("err = %v, want ErrExtendsPathNotLiteral", err)
	}
}

// Phase I cycle 5: child content outside blocks is ignored.
func TestExtends_NonBlockContentIgnored(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"base.txt":  `[{% block x %}A{% endblock %}]`,
		"child.txt": `{% extends "base.txt" %}This text is ignored.{% block x %}B{% endblock %}More ignored.`,
	}))
	got, err := engine.Render("child.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "[B]" {
		t.Errorf("got %q", got)
	}
}

// Phase I cycle 6: endblock can carry a matching name.
func TestBlock_EndblockNameMatches(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"a.txt": `[{% block x %}hi{% endblock x %}]`,
	}))
	got, err := engine.Render("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "[hi]" {
		t.Errorf("got %q", got)
	}
}

// Phase I cycle 7: endblock with wrong name errors.
func TestBlock_EndblockNameMismatch_Errors(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"a.txt": `{% block x %}hi{% endblock y %}`,
	}))
	_, err := engine.Load("a.txt")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errString(err), "endblock") {
		t.Errorf("err = %v", err)
	}
}

// Phase I cycle 8: duplicate block name in same template errors.
func TestBlock_DuplicateNameErrors(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"a.txt": `{% block x %}1{% endblock %}{% block x %}2{% endblock %}`,
	}))
	_, err := engine.Load("a.txt")
	if !errors.Is(err, ErrBlockRedefined) {
		t.Errorf("err = %v, want ErrBlockRedefined", err)
	}
}

// Phase I cycle 9: circular extends chain is detected.
func TestExtends_CircularChain_Errors(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"a.txt": `{% extends "b.txt" %}`,
		"b.txt": `{% extends "a.txt" %}`,
	}))
	_, err := engine.Load("a.txt")
	if !errors.Is(err, ErrCircularExtends) {
		t.Errorf("err = %v, want ErrCircularExtends", err)
	}
}

// Phase I cycle 10: extending a missing parent template errors.
func TestExtends_ParentMissing_Errors(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"child.txt": `{% extends "nope.txt" %}`,
	}))
	_, err := engine.Load("child.txt")
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("err = %v, want ErrTemplateNotFound", err)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// Phase J cycle 1: render uses root's body, not leaf's.
func TestExtends_RenderUsesRootBody(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"base.txt":  `start-{% block x %}default{% endblock %}-end`,
		"child.txt": `{% extends "base.txt" %}{% block x %}hi{% endblock %}`,
	}))
	got, err := engine.Render("child.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "start-hi-end" {
		t.Errorf("got %q", got)
	}
}

// Phase J cycle 2: three-level inheritance — deepest override wins.
func TestExtends_ThreeLevelChain(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"a.txt":      `{% block x %}A{% endblock %}`,
		"middle.txt": `{% extends "a.txt" %}{% block x %}M{% endblock %}`,
		"leaf.txt":   `{% extends "middle.txt" %}{% block x %}L{% endblock %}`,
	}))
	got, err := engine.Render("leaf.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "L" {
		t.Errorf("got %q, want L", got)
	}
}

// Phase J cycle 3: middle layer overrides, leaf does not — middle wins.
func TestExtends_ThreeLevel_LeafFallsThroughToMiddle(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"a.txt":      `{% block x %}A{% endblock %}`,
		"middle.txt": `{% extends "a.txt" %}{% block x %}M{% endblock %}`,
		"leaf.txt":   `{% extends "middle.txt" %}`,
	}))
	got, err := engine.Render("leaf.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "M" {
		t.Errorf("got %q, want M", got)
	}
}

// Phase J cycle 4: multiple independent blocks.
func TestExtends_MultipleBlocksOverriddenIndependently(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"base.txt": `[{% block a %}1{% endblock %}|{% block b %}2{% endblock %}]`,
		"c.txt":    `{% extends "base.txt" %}{% block a %}X{% endblock %}{% block b %}Y{% endblock %}`,
	}))
	got, err := engine.Render("c.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "[X|Y]" {
		t.Errorf("got %q", got)
	}
}

// Phase J cycle 5: blocks in included templates render inline and do
// NOT participate in the extends chain.
func TestInclude_BlockInsideIncluded_RendersInline(t *testing.T) {
	t.Parallel()

	engine := newLayoutTextExtendsEngine(NewMemoryLoader(map[string]string{
		"base.txt":    `[{% block content %}BASE{% endblock %}]`,
		"partial.txt": `{% block content %}PARTIAL{% endblock %}`,
		"page.txt":    `{% extends "base.txt" %}{% block content %}before-{% include "partial.txt" %}-after{% endblock %}`,
	}))
	got, err := engine.Render("page.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// Partial's block renders inline (PARTIAL), doesn't override page's
	// content block (which is what produces the surrounding text).
	if got != "[before-PARTIAL-after]" {
		t.Errorf("got %q", got)
	}
}
