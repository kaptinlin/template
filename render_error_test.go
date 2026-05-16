package template

import (
	"errors"
	"strings"
	"testing"
)

func TestRenderError_FilterNotFound_PreservesSentinelAndPosition(t *testing.T) {
	t.Parallel()

	source := "hello {{ name | bogus }}"
	_, err := renderSourceTemplate(source, Data{"name": "world"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrFilterNotFound) {
		t.Fatalf("errors.Is ErrFilterNotFound = false, want true; err=%v", err)
	}

	var re *RenderError
	if !errors.As(err, &re) {
		t.Fatalf("errors.As *RenderError = false; err=%v", err)
	}
	if re.Line != 1 {
		t.Errorf("RenderError.Line = %d, want 1", re.Line)
	}
	if re.Col == 0 {
		t.Error("RenderError.Col = 0, want > 0")
	}
	if re.Template != "" {
		t.Errorf("RenderError.Template = %q, want empty for ParseString template", re.Template)
	}
}

func TestRenderError_MultiLineLineNumber(t *testing.T) {
	t.Parallel()

	source := "first line\nsecond line\n{{ name | bogus }}"
	_, err := renderSourceTemplate(source, Data{"name": "world"})
	if err == nil {
		t.Fatal("expected error")
	}

	var re *RenderError
	if !errors.As(err, &re) {
		t.Fatalf("errors.As *RenderError = false; err=%v", err)
	}
	if re.Line != 3 {
		t.Errorf("RenderError.Line = %d, want 3", re.Line)
	}
}

func TestRenderError_ErrorStringIncludesPosition(t *testing.T) {
	t.Parallel()

	source := "{{ x | bogus }}"
	_, err := renderSourceTemplate(source, Data{"x": 1})
	if err == nil {
		t.Fatal("expected error")
	}
	got := err.Error()
	if !strings.Contains(got, "1:") {
		t.Errorf("error string %q missing line marker", got)
	}
	if !strings.Contains(got, "filter not found") {
		t.Errorf("error string %q missing cause text", got)
	}
}

func TestRenderError_TemplateNameFromLoader(t *testing.T) {
	t.Parallel()

	loader := NewMemoryLoader(map[string]string{
		"page.txt": "line1\n{{ x | bogus }}\n",
	})
	engine := New(WithLoader(loader))

	_, err := engine.Render("page.txt", Data{"x": 1})
	if err == nil {
		t.Fatal("expected error")
	}

	var re *RenderError
	if !errors.As(err, &re) {
		t.Fatalf("errors.As *RenderError = false; err=%v", err)
	}
	if re.Template != "page.txt" {
		t.Errorf("RenderError.Template = %q, want %q", re.Template, "page.txt")
	}
	if re.Line != 2 {
		t.Errorf("RenderError.Line = %d, want 2", re.Line)
	}
}

func TestRenderError_IncludePreservesChildTemplate(t *testing.T) {
	t.Parallel()

	loader := NewMemoryLoader(map[string]string{
		"parent.txt": `{% include "child.txt" %}`,
		"child.txt":  "{{ x | bogus }}",
	})
	engine := New(WithLoader(loader), WithLayout())

	_, err := engine.Render("parent.txt", Data{"x": 1})
	if err == nil {
		t.Fatal("expected error")
	}

	var re *RenderError
	if !errors.As(err, &re) {
		t.Fatalf("errors.As *RenderError = false; err=%v", err)
	}
	if re.Template != "child.txt" {
		t.Errorf("RenderError.Template = %q, want %q (child name should win, deepest position)", re.Template, "child.txt")
	}
	if !errors.Is(err, ErrFilterNotFound) {
		t.Errorf("errors.Is ErrFilterNotFound = false; err=%v", err)
	}
}

func TestRenderError_DivisionByZero(t *testing.T) {
	t.Parallel()

	source := "value: {{ 10 / 0 }}"
	_, err := renderSourceTemplate(source, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrDivisionByZero) {
		t.Errorf("errors.Is ErrDivisionByZero = false; err=%v", err)
	}

	var re *RenderError
	if !errors.As(err, &re) {
		t.Fatalf("errors.As *RenderError = false")
	}
	if re.Line != 1 {
		t.Errorf("RenderError.Line = %d, want 1", re.Line)
	}
}

func TestRenderError_ForLoopBodyError(t *testing.T) {
	t.Parallel()

	source := "{% for x in items %}{{ x | bogus }}{% endfor %}"
	_, err := renderSourceTemplate(source, Data{"items": []int{1, 2}})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrFilterNotFound) {
		t.Errorf("errors.Is ErrFilterNotFound = false; err=%v", err)
	}

	var re *RenderError
	if !errors.As(err, &re) {
		t.Fatalf("errors.As *RenderError = false")
	}
	// The deepest wrap is the failing OutputNode inside the loop body.
	if re.Line != 1 {
		t.Errorf("RenderError.Line = %d, want 1", re.Line)
	}
}
