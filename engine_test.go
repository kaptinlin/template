package template

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type loaderResult struct {
	source   string
	resolved string
}

type aliasLoader struct {
	mu    sync.Mutex
	files map[string]loaderResult
	opens map[string]int
}

func (l *aliasLoader) Open(name string) (string, string, error) {
	if err := ValidateName(name); err != nil {
		return "", "", err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.opens[name]++

	result, ok := l.files[name]
	if !ok {
		return "", "", ErrTemplateNotFound
	}
	return result.source, result.resolved, nil
}

func TestEngineParseString_DefaultCoreOnly(t *testing.T) {
	t.Parallel()

	engine := New()

	got, err := engine.ParseString(`Hello, {{ name }}!`)
	if err != nil {
		t.Fatalf("ParseString() err = %v", err)
	}

	out, err := got.Render(Data{"name": "world"})
	if err != nil {
		t.Fatalf("renderSourceTemplate() err = %v", err)
	}
	if out != "Hello, world!" {
		t.Errorf("renderSourceTemplate() = %q, want %q", out, "Hello, world!")
	}
}

func TestWithLayout_EnablesFeatureLayout(t *testing.T) {
	t.Parallel()

	engine := New(WithLayout())
	if !engine.HasFeature(FeatureLayout) {
		t.Fatal("HasFeature(FeatureLayout) = false, want true")
	}
}

func TestEngineRegisterTag_AddsEngineLocalTag(t *testing.T) {
	t.Parallel()

	engine := New()
	err := engine.RegisterTag("set", func(_ *Parser, start *Token, args *Parser) (Statement, error) {
		v, err := args.ExpectIdentifier()
		if err != nil {
			return nil, args.Error("expected variable name after 'set'")
		}
		if args.Match(TokenSymbol, "=") == nil {
			return nil, args.Error("expected '=' after variable name")
		}
		expr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		return &testSetNode{
			varName: v.Value,
			expr:    expr,
			line:    start.Line,
			col:     start.Col,
		}, nil
	})
	if err != nil {
		t.Fatalf("RegisterTag() err = %v", err)
	}

	tpl, err := engine.ParseString(`{% set greeting = "Hello" %}{{ greeting }}`)
	if err != nil {
		t.Fatalf("ParseString() err = %v", err)
	}
	got, err := tpl.Render(nil)
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	if got != "Hello" {
		t.Errorf("Render() = %q, want %q", got, "Hello")
	}
}

func TestEngineRegisterFilter_AddsEngineLocalFilter(t *testing.T) {
	t.Parallel()

	engine := New()
	engine.RegisterFilter("repeat", func(value any, args ...any) (any, error) {
		return toString(value) + toString(value), nil
	})

	tpl, err := engine.ParseString(`{{ word | repeat }}`)
	if err != nil {
		t.Fatalf("ParseString() err = %v", err)
	}
	got, err := tpl.Render(Data{"word": "ha"})
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	if got != "haha" {
		t.Errorf("Render() = %q, want %q", got, "haha")
	}
}

func TestTemplateRenderTo_UsesDataPath(t *testing.T) {
	t.Parallel()

	tpl, err := New().ParseString(`Hello, {{ name }}!`)
	if err != nil {
		t.Fatalf("ParseString() err = %v", err)
	}

	var buf bytes.Buffer
	if err := tpl.RenderTo(&buf, Data{"name": "world"}); err != nil {
		t.Fatalf("RenderTo() err = %v", err)
	}
	if got := buf.String(); got != "Hello, world!" {
		t.Errorf("RenderTo() = %q, want %q", got, "Hello, world!")
	}
}

func TestEngineLoad_LayoutRequiresFeature(t *testing.T) {
	t.Parallel()

	loader := NewMemoryLoader(map[string]string{
		"page.txt":    `{% include "partial.txt" %}`,
		"partial.txt": "hello",
	})

	engine := New(WithLoader(loader))
	_, err := engine.Load("page.txt")
	if err == nil {
		t.Fatal("Load() err = nil, want error")
	}
	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("Load() err = %T, want *ParseError", err)
	}
	if !strings.Contains(err.Error(), "page.txt:") {
		t.Fatalf("Load() err = %q, want template name prefix", err.Error())
	}
}

func TestEngineLoad_LayoutFeatureEnabled(t *testing.T) {
	t.Parallel()

	loader := NewMemoryLoader(map[string]string{
		"page.txt":    `{% extends "base.txt" %}{% block content %}hello{% endblock %}`,
		"base.txt":    `[{% block content %}{% endblock %}]`,
		"widget.txt":  `{% include "partial.txt" %}`,
		"partial.txt": `ok`,
	})

	engine := New(
		WithLoader(loader),
		WithLayout(),
	)

	page, err := engine.Load("page.txt")
	if err != nil {
		t.Fatalf("Load(page.txt) err = %v", err)
	}
	got, err := page.Render(nil)
	if err != nil {
		t.Fatalf("renderSourceTemplate() err = %v", err)
	}
	if got != "[hello]" {
		t.Errorf("renderSourceTemplate() = %q, want %q", got, "[hello]")
	}

	widget, err := engine.Load("widget.txt")
	if err != nil {
		t.Fatalf("Load(widget.txt) err = %v", err)
	}
	got, err = widget.Render(nil)
	if err != nil {
		t.Fatalf("renderSourceTemplate() err = %v", err)
	}
	if got != "ok" {
		t.Errorf("renderSourceTemplate() = %q, want %q", got, "ok")
	}
}

func TestEngineRenderString_HTMLFormatAutoescapes(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `{{ title }}`,
		})),
		WithFormat(FormatHTML),
	)

	got, err := engine.Render("a.html", Data{"title": "<b>"})
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	if got != "&lt;b&gt;" {
		t.Errorf("Render() = %q, want %q", got, "&lt;b&gt;")
	}
}

func TestEngineTemplateRender_InheritsEngineSemantics(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	registry.Register("bang", func(value any, _ ...any) (any, error) {
		return toString(value) + "!", nil
	})

	engine := New(
		WithFormat(FormatHTML),
		WithFilters(registry),
	)

	tpl, err := engine.ParseString(`{{ title | bang }}`)
	if err != nil {
		t.Fatalf("ParseString() err = %v", err)
	}

	got, err := tpl.Render(Data{"title": "<b>"})
	if err != nil {
		t.Fatalf("renderSourceTemplate() err = %v", err)
	}
	if got != "&lt;b&gt;!" {
		t.Errorf("renderSourceTemplate() = %q, want %q", got, "&lt;b&gt;!")
	}
}

func TestEngineClone_IsolatesLocalRegistries(t *testing.T) {
	t.Parallel()

	engine := New()
	engine.RegisterFilter("alpha", func(value any, _ ...any) (any, error) {
		return "alpha:" + toString(value), nil
	})

	clone := engine.Clone()
	clone.RegisterFilter("beta", func(value any, _ ...any) (any, error) {
		return "beta:" + toString(value), nil
	})

	if engine.Filters().Has("beta") {
		t.Fatal("original engine unexpectedly sees clone-local filter")
	}

	tpl, err := clone.ParseString(`{{ x | beta }}`)
	if err != nil {
		t.Fatalf("ParseString() err = %v", err)
	}
	got, err := tpl.Render(Data{"x": "ok"})
	if err != nil {
		t.Fatalf("renderSourceTemplate() err = %v", err)
	}
	if got != "beta:ok" {
		t.Errorf("renderSourceTemplate() = %q, want %q", got, "beta:ok")
	}
}

func TestEngineLoad_NotFoundIncludesTemplateName(t *testing.T) {
	t.Parallel()

	engine := New(WithLoader(NewMemoryLoader(map[string]string{})))

	_, err := engine.Load("missing.txt")
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("Load(missing.txt) err = %v, want ErrTemplateNotFound", err)
	}
	if !strings.Contains(err.Error(), "missing.txt") {
		t.Fatalf("Load(missing.txt) err = %q, want template name", err.Error())
	}
}

func TestEngineLoad_EmptyTagExpressionIncludesPosition(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"bad.txt": `{% if %}x{% endif %}`,
		})),
	)

	_, err := engine.Load("bad.txt")
	if err == nil {
		t.Fatal("Load(bad.txt) err = nil, want error")
	}

	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("Load(bad.txt) err = %T, want *ParseError", err)
	}
	if parseErr.Line == 0 || parseErr.Col == 0 {
		t.Fatalf("ParseError position = %d:%d, want non-zero", parseErr.Line, parseErr.Col)
	}
	if !strings.Contains(err.Error(), "bad.txt:") {
		t.Fatalf("Load(bad.txt) err = %q, want template name prefix", err.Error())
	}
}

func TestEngineLoad_CachesByResolvedName(t *testing.T) {
	t.Parallel()

	var parseCount atomic.Int32

	registry := NewTagRegistry()
	if err := registry.Register("probe", func(_ *Parser, start *Token, args *Parser) (Statement, error) {
		parseCount.Add(1)
		if args.Remaining() > 0 {
			return nil, args.Error("probe does not take arguments")
		}
		return NewTextNode("ok", start.Line, start.Col), nil
	}); err != nil {
		t.Fatalf("Register(probe) err = %v", err)
	}

	loader := &aliasLoader{
		files: map[string]loaderResult{
			"a.txt":     {source: `{% probe %}`, resolved: "shared.txt"},
			"alias.txt": {source: `{% probe %}`, resolved: "shared.txt"},
		},
		opens: make(map[string]int),
	}

	engine := New(
		WithLoader(loader),
		WithTags(registry),
	)

	first, err := engine.Load("a.txt")
	if err != nil {
		t.Fatalf("Load(a.txt) err = %v", err)
	}
	second, err := engine.Load("alias.txt")
	if err != nil {
		t.Fatalf("Load(alias.txt) err = %v", err)
	}

	if first != second {
		t.Fatal("Load() returned distinct template pointers for the same resolved name")
	}
	if got := parseCount.Load(); got != 1 {
		t.Fatalf("parse count = %d, want 1", got)
	}
	if len(engine.cache) != 1 {
		t.Fatalf("len(engine.cache) = %d, want 1", len(engine.cache))
	}
	if _, ok := engine.cache["shared.txt"]; !ok {
		t.Fatal("engine cache missing resolved key shared.txt")
	}
}

func TestEngineLoad_WaitsForInFlightCompile(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	release := make(chan struct{})
	var startOnce sync.Once

	registry := NewTagRegistry()
	if err := registry.Register("gate", func(_ *Parser, start *Token, args *Parser) (Statement, error) {
		if args.Remaining() > 0 {
			return nil, args.Error("gate does not take arguments")
		}
		startOnce.Do(func() { close(started) })
		<-release
		return NewTextNode("ok", start.Line, start.Col), nil
	}); err != nil {
		t.Fatalf("Register(gate) err = %v", err)
	}

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `{% gate %}`,
		})),
		WithTags(registry),
	)

	firstDone := make(chan struct{})
	var first *Template
	var firstErr error
	go func() {
		defer close(firstDone)
		first, firstErr = engine.Load("a.txt")
	}()

	<-started

	secondDone := make(chan struct{})
	var second *Template
	var secondErr error
	go func() {
		defer close(secondDone)
		second, secondErr = engine.Load("a.txt")
	}()

	close(release)
	<-firstDone
	<-secondDone

	if firstErr != nil {
		t.Fatalf("first Load(a.txt) err = %v", firstErr)
	}
	if secondErr != nil {
		t.Fatalf("second Load(a.txt) err = %v", secondErr)
	}
	if first != second {
		t.Fatal("Load() returned distinct template pointers for in-flight requests")
	}
}

func TestEngineLoad_ConcurrentRequestsCompileOnce(t *testing.T) {
	t.Parallel()

	var parseCount atomic.Int32

	registry := NewTagRegistry()
	if err := registry.Register("probe", func(_ *Parser, start *Token, args *Parser) (Statement, error) {
		parseCount.Add(1)
		if args.Remaining() > 0 {
			return nil, args.Error("probe does not take arguments")
		}
		return NewTextNode("ok", start.Line, start.Col), nil
	}); err != nil {
		t.Fatalf("Register(probe) err = %v", err)
	}

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `{% probe %}`,
		})),
		WithTags(registry),
	)

	const workers = 16

	results := make(chan *Template, workers)
	errs := make(chan error, workers)

	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			tpl, err := engine.Load("a.txt")
			if err != nil {
				errs <- err
				return
			}
			results <- tpl
		}()
	}
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		t.Fatalf("Load(a.txt) err = %v", err)
	}

	var first *Template
	for tpl := range results {
		if first == nil {
			first = tpl
			continue
		}
		if tpl != first {
			t.Fatal("Load() returned distinct template pointers for concurrent requests")
		}
	}
	if got := parseCount.Load(); got != 1 {
		t.Fatalf("parse count = %d, want 1", got)
	}
}

func TestEngineRenderString_ConcurrentAliasNamesShareResolvedTemplate(t *testing.T) {
	t.Parallel()

	var parseCount atomic.Int32

	registry := NewTagRegistry()
	if err := registry.Register("probe", func(_ *Parser, start *Token, args *Parser) (Statement, error) {
		parseCount.Add(1)
		if args.Remaining() > 0 {
			return nil, args.Error("probe does not take arguments")
		}
		return NewTextNode("ok", start.Line, start.Col), nil
	}); err != nil {
		t.Fatalf("Register(probe) err = %v", err)
	}

	loader := &aliasLoader{
		files: map[string]loaderResult{
			"a.txt":     {source: `{% probe %}`, resolved: "shared.txt"},
			"alias.txt": {source: `{% probe %}`, resolved: "shared.txt"},
		},
		opens: make(map[string]int),
	}

	engine := New(
		WithLoader(loader),
		WithTags(registry),
	)

	const workers = 16

	errs := make(chan error, workers)
	results := make(chan string, workers)

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := range workers {
		go func(i int) {
			defer wg.Done()

			name := "a.txt"
			if i%2 == 1 {
				name = "alias.txt"
			}

			got, err := engine.Render(name, nil)
			if err != nil {
				errs <- err
				return
			}
			results <- got
		}(i)
	}
	wg.Wait()
	close(errs)
	close(results)

	for err := range errs {
		t.Fatalf("Render() err = %v", err)
	}
	for got := range results {
		if got != "ok" {
			t.Fatalf("Render() = %q, want %q", got, "ok")
		}
	}
	if got := parseCount.Load(); got != 1 {
		t.Fatalf("parse count = %d, want 1", got)
	}
}
