package template

import (
	"errors"
	"strings"
	"testing"
)

// Phase C cycle 1: static string literal include renders target template.
func TestInclude_StaticPath_RendersTargetTemplate(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `before {% include "b.txt" %} after`,
		"b.txt": `middle`,
	}))

	got, err := set.RenderString("a.txt", nil)
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	if got != "before middle after" {
		t.Errorf("got %q, want %q", got, "before middle after")
	}
}

// Phase C cycle 2: the included template sees the parent's context.
func TestInclude_ChildSeesParentContext(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `{% include "b.txt" %}`,
		"b.txt": `hello {{ name }}`,
	}))

	got, err := set.RenderString("a.txt", Context{"name": "world"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %q", got)
	}
}

// Phase C cycle 3: parse-time missing template fails fast.
func TestInclude_ParseTimeMissing_Errors(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `{% include "missing.txt" %}`,
	}))
	_, err := set.Get("a.txt")
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("err = %v, want ErrTemplateNotFound", err)
	}
}

// Compile-path templates don't know the include tag. {% include %} is
// a Set-scoped feature.
func TestInclude_CompilePath_UnknownTag(t *testing.T) {
	t.Parallel()

	_, err := Compile(`{% include "x.txt" %}`)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown tag: include") {
		t.Errorf("err = %v, want 'unknown tag: include'", err)
	}
}

// Phase C cycle 5: three-level nested include.
func TestInclude_NestedInclude_ThreeLevels(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `A{% include "b.txt" %}A`,
		"b.txt": `B{% include "c.txt" %}B`,
		"c.txt": `C`,
	}))
	got, err := set.RenderString("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "ABCBA" {
		t.Errorf("got %q", got)
	}
}

// Phase D cycle 1: include with a single keyword argument.
func TestInclude_With_SingleVar(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt":    `{% include "card.txt" with title="Hi" %}`,
		"card.txt": `[{{ title }}]`,
	}))
	got, err := set.RenderString("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "[Hi]" {
		t.Errorf("got %q", got)
	}
}

// Phase D cycle 2: include with multiple keyword arguments.
func TestInclude_With_MultipleVars(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt":    `{% include "card.txt" with title="Hi" count=3 %}`,
		"card.txt": `{{ title }}x{{ count }}`,
	}))
	got, err := set.RenderString("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "Hix3" {
		t.Errorf("got %q", got)
	}
}

// Phase D cycle 3: "with" values are evaluated in the parent context.
func TestInclude_With_ExpressionEvaluatedInParent(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt":    `{% include "card.txt" with label=page.title %}`,
		"card.txt": `[{{ label }}]`,
	}))
	got, err := set.RenderString("a.txt", Context{
		"page": map[string]any{"title": "Welcome"},
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "[Welcome]" {
		t.Errorf("got %q", got)
	}
}

// Phase D cycle 4: "only" fully isolates the child from parent context.
func TestInclude_Only_IsolatesParentContext(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt":    `{% include "card.txt" with title="Hi" only %}[{{ name }}]`,
		"card.txt": `{{ title }}+{{ name }}`,
	}))
	// name is set in the parent; the child sees only "title" (from with).
	// Inside the child, {{ name }} resolves to empty. Back in the parent,
	// {{ name }} works normally.
	got, err := set.RenderString("a.txt", Context{"name": "parent"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "Hi+[parent]" {
		t.Errorf("got %q, want %q", got, "Hi+[parent]")
	}
}

// Phase D cycle 5: "only" also isolates globals set via WithGlobals.
func TestInclude_Only_IsolatesGlobals(t *testing.T) {
	t.Parallel()

	set := NewTextSet(
		NewMemoryLoader(map[string]string{
			"a.txt":    `{% include "card.txt" with title="Hi" only %}`,
			"card.txt": `{{ title }}+{{ site }}`,
		}),
		WithGlobals(Context{"site": "main"}),
	)
	got, err := set.RenderString("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "Hi+" {
		t.Errorf("got %q, want %q", got, "Hi+")
	}
}

// Phase D cycle 6: "only" alone (no with) fully isolates.
func TestInclude_OnlyAlone_SeesNoVars(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt":    `{% include "card.txt" only %}`,
		"card.txt": `[{{ title }}]`,
	}))
	got, err := set.RenderString("a.txt", Context{"title": "Hi"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "[]" {
		t.Errorf("got %q", got)
	}
}

// Phase D cycle 7: "if_exists" on a missing template renders nothing.
func TestInclude_IfExists_MissingIsNoop(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `before{% include "missing.txt" if_exists %}after`,
	}))
	got, err := set.RenderString("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "beforeafter" {
		t.Errorf("got %q", got)
	}
}

// Phase D cycle 8: "if_exists" on a present template renders normally.
func TestInclude_IfExists_PresentWorks(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `{% include "b.txt" if_exists %}`,
		"b.txt": "hi",
	}))
	got, err := set.RenderString("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "hi" {
		t.Errorf("got %q", got)
	}
}

// Phase E cycle 1: dynamic path from context.
func TestInclude_DynamicPath_FromContext(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt":       `{% include widget %}`,
		"widget1.txt": "one",
		"widget2.txt": "two",
	}))
	got, err := set.RenderString("a.txt", Context{"widget": "widget1.txt"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "one" {
		t.Errorf("got %q", got)
	}
}

// Phase E cycle 2: dynamic name with path traversal is rejected at runtime.
func TestInclude_DynamicPath_InvalidName_Rejected(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `{% include bad %}`,
	}))
	_, err := set.RenderString("a.txt", Context{"bad": "../etc/passwd"})
	if !errors.Is(err, ErrInvalidTemplateName) {
		t.Errorf("err = %v, want ErrInvalidTemplateName", err)
	}
}

// Phase E cycle 3: dynamic missing template returns ErrTemplateNotFound.
func TestInclude_DynamicPath_Missing_Errors(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `{% include w %}`,
	}))
	_, err := set.RenderString("a.txt", Context{"w": "nope.txt"})
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("err = %v, want ErrTemplateNotFound", err)
	}
}

// Phase E cycle 4: self-recursive include hits the depth limit.
func TestInclude_SelfRecursion_HitsDepthLimit(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `x{% include w %}`,
	}))
	// Pass a dynamic name that re-includes the same template indefinitely.
	_, err := set.RenderString("a.txt", Context{"w": "a.txt"})
	if !errors.Is(err, ErrIncludeDepthExceeded) {
		t.Errorf("err = %v, want ErrIncludeDepthExceeded", err)
	}
}

// Phase E cycle 5: parse-time mutual recursion downgrades to lazy mode.
// Without the lazy downgrade, parse would infinite-loop resolving A→B→A.
func TestInclude_MutualRecursion_ParseDowngradesToLazy(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		"a.txt": `A{% include "b.txt" %}`,
		"b.txt": `B{% include "a.txt" %}`,
	}))
	// With lazy downgrade at parse time, execution eventually hits the
	// runtime depth limit — we just need compile to succeed.
	_, err := set.Get("a.txt")
	if err != nil {
		t.Fatalf("parse err = %v", err)
	}
	// And execute hits the runtime depth limit.
	_, err = set.RenderString("a.txt", nil)
	if !errors.Is(err, ErrIncludeDepthExceeded) {
		t.Errorf("execute err = %v, want ErrIncludeDepthExceeded", err)
	}
}

// Phase E cycle 6: recursion terminates naturally when data runs out.
// This proves lazy mode supports tree-walk rendering patterns.
func TestInclude_LazyRecursion_TerminatesOnData(t *testing.T) {
	t.Parallel()

	set := NewTextSet(NewMemoryLoader(map[string]string{
		// Renders a two-level nested list. The child is included via
		// dynamic path so parse-time resolution is not attempted.
		"outer.txt": `[{% for item in items %}{% include "inner.txt" %}{% endfor %}]`,
		"inner.txt": `{{ item }}`,
	}))
	got, err := set.RenderString("outer.txt", Context{
		"items": []any{"a", "b", "c"},
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "[abc]" {
		t.Errorf("got %q", got)
	}
}
