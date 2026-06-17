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

func withTag(name string, parser tagParser) EngineOption {
	return func(e *Engine) {
		e.tags.Replace(name, parser)
	}
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

func TestWithFeatures_EnablesLayoutRegistryLayer(t *testing.T) {
	t.Parallel()

	engine := New(WithFeatures(FeatureLayout))
	if !engine.HasFeature(FeatureLayout) {
		t.Fatal("HasFeature(FeatureLayout) = false, want true")
	}
	if !engine.HasFilter("safe") {
		t.Fatal("HasFilter(safe) = false, want true")
	}

	core := New()
	if core.HasFilter("safe") {
		t.Fatal("core HasFilter(safe) = true, want false")
	}
}

func TestEngineFilterQueries_IncludeLocalAndBuiltInLayers(t *testing.T) {
	t.Parallel()

	engine := New(
		WithFilter("bang", func(value any, _ ...any) (any, error) {
			return toString(value) + "!", nil
		}),
	)
	if !engine.HasFilter("bang") {
		t.Fatal("HasFilter(bang) = false, want true")
	}
	if !engine.HasFilter("escape") {
		t.Fatal("HasFilter(escape) = false, want built-in parent filter")
	}
}

func TestEngineInternalTagOption_AddsEngineLocalTag(t *testing.T) {
	t.Parallel()

	engine := New(withTag("set", func(_ *parser, start *token, args *parser) (statement, error) {
		v, err := args.ExpectIdentifier()
		if err != nil {
			return nil, args.Error("expected variable name after 'set'")
		}
		if args.Match(tokenSymbol, "=") == nil {
			return nil, args.Error("expected '=' after variable name")
		}
		expr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		return &testSetNode{
			varName: v.value,
			expr:    expr,
			line:    start.Line,
			col:     start.Col,
		}, nil
	}))

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
	if err := engine.RegisterFilter("repeat", func(value any, args ...any) (any, error) {
		return toString(value) + toString(value), nil
	}); err != nil {
		t.Fatalf("RegisterFilter(repeat) err = %v", err)
	}

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

func TestEngineRegisterFilter_NilFilterReturnsError(t *testing.T) {
	t.Parallel()

	engine := New()
	if err := engine.RegisterFilter("bad", nil); !errors.Is(err, errNilFilterFunction) {
		t.Fatalf("RegisterFilter(bad, nil) err = %v, want errNilFilterFunction", err)
	}
	if err := engine.ReplaceFilter("bad", nil); !errors.Is(err, errNilFilterFunction) {
		t.Fatalf("ReplaceFilter(bad, nil) err = %v, want errNilFilterFunction", err)
	}
}

func TestEngineReplaceFilter_OverridesEngineLocalFilter(t *testing.T) {
	t.Parallel()

	engine := New()
	if err := engine.RegisterFilter("mark", func(value any, _ ...any) (any, error) {
		return "old:" + toString(value), nil
	}); err != nil {
		t.Fatalf("RegisterFilter(mark) err = %v", err)
	}
	if err := engine.ReplaceFilter("mark", func(value any, _ ...any) (any, error) {
		return "new:" + toString(value), nil
	}); err != nil {
		t.Fatalf("ReplaceFilter(mark) err = %v", err)
	}

	tpl, err := engine.ParseString(`{{ word | mark }}`)
	if err != nil {
		t.Fatalf("ParseString() err = %v", err)
	}
	got, err := tpl.Render(Data{"word": "ok"})
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	if got != "new:ok" {
		t.Errorf("Render() = %q, want %q", got, "new:ok")
	}
}

func TestWithFilter_NilFilterReturnsCompileError(t *testing.T) {
	t.Parallel()

	engine := New(WithFilter("bad", nil))
	_, err := engine.ParseString(`{{ word | bad }}`)
	if !errors.Is(err, errNilFilterFunction) {
		t.Fatalf("ParseString() err = %v, want errNilFilterFunction", err)
	}
}

func TestWithFilter_AddsEngineLocalFilterAtConstruction(t *testing.T) {
	t.Parallel()

	engine := New(WithFilter("mark", func(value any, _ ...any) (any, error) {
		return "mark:" + toString(value), nil
	}))

	tpl, err := engine.ParseString(`{{ word | mark }}`)
	if err != nil {
		t.Fatalf("ParseString() err = %v", err)
	}
	got, err := tpl.Render(Data{"word": "ok"})
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	if got != "mark:ok" {
		t.Errorf("Render() = %q, want %q", got, "mark:ok")
	}
}

func TestEngineCachedTemplateKeepsCompiledFilterAfterRejectedMutation(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"page.txt": `{{ word | mark }}`,
		})),
	)
	if err := engine.RegisterFilter("mark", func(value any, _ ...any) (any, error) {
		return "old:" + toString(value), nil
	}); err != nil {
		t.Fatalf("RegisterFilter(mark) err = %v", err)
	}

	got, err := engine.Render("page.txt", Data{"word": "ok"})
	if err != nil {
		t.Fatalf("Render(page.txt) err = %v", err)
	}
	if got != "old:ok" {
		t.Fatalf("Render(page.txt) = %q, want %q", got, "old:ok")
	}

	err = engine.ReplaceFilter("mark", func(value any, _ ...any) (any, error) {
		return "new:" + toString(value), nil
	})
	if !errors.Is(err, ErrEngineCompiled) {
		t.Fatalf("ReplaceFilter(mark) err = %v, want ErrEngineCompiled", err)
	}

	got, err = engine.Render("page.txt", Data{"word": "ok"})
	if err != nil {
		t.Fatalf("Render(page.txt) after mutation err = %v", err)
	}
	if got != "old:ok" {
		t.Errorf("Render(page.txt) after mutation = %q, want %q", got, "old:ok")
	}
}

func TestEngineCompiledFiltersStayEngineLocal(t *testing.T) {
	t.Parallel()

	left := New(WithFilter("mark", func(value any, _ ...any) (any, error) {
		return "left:" + toString(value), nil
	}))
	right := New(WithFilter("mark", func(value any, _ ...any) (any, error) {
		return "right:" + toString(value), nil
	}))

	leftTpl, err := left.ParseString(`{{ word | mark }}`)
	if err != nil {
		t.Fatalf("left ParseString() err = %v", err)
	}
	rightTpl, err := right.ParseString(`{{ word | mark }}`)
	if err != nil {
		t.Fatalf("right ParseString() err = %v", err)
	}

	err = left.ReplaceFilter("mark", func(value any, _ ...any) (any, error) {
		return "changed:" + toString(value), nil
	})
	if !errors.Is(err, ErrEngineCompiled) {
		t.Fatalf("left ReplaceFilter(mark) err = %v, want ErrEngineCompiled", err)
	}

	got, err := leftTpl.Render(Data{"word": "ok"})
	if err != nil {
		t.Fatalf("left Render() err = %v", err)
	}
	if got != "left:ok" {
		t.Errorf("left Render() = %q, want %q", got, "left:ok")
	}

	got, err = rightTpl.Render(Data{"word": "ok"})
	if err != nil {
		t.Fatalf("right Render() err = %v", err)
	}
	if got != "right:ok" {
		t.Errorf("right Render() = %q, want %q", got, "right:ok")
	}
}

func TestEngineRegisterFilterAfterCompileErrors(t *testing.T) {
	t.Parallel()

	engine := New()
	tpl, err := engine.ParseString(`{{ word }}`)
	if err != nil {
		t.Fatalf("ParseString() err = %v", err)
	}

	err = engine.RegisterFilter("late", func(value any, _ ...any) (any, error) {
		return value, nil
	})
	if !errors.Is(err, ErrEngineCompiled) {
		t.Fatalf("RegisterFilter(late) err = %v, want ErrEngineCompiled", err)
	}

	got, err := tpl.Render(Data{"word": "ok"})
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	if got != "ok" {
		t.Errorf("Render() = %q, want %q", got, "ok")
	}
}

func TestEngineReplaceFilterAfterLoadErrors(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"page.txt": `{{ word | mark }}`,
		})),
		WithFilter("mark", func(value any, _ ...any) (any, error) {
			return "old:" + toString(value), nil
		}),
	)

	got, err := engine.Render("page.txt", Data{"word": "ok"})
	if err != nil {
		t.Fatalf("Render(page.txt) err = %v", err)
	}
	if got != "old:ok" {
		t.Fatalf("Render(page.txt) = %q, want %q", got, "old:ok")
	}

	err = engine.ReplaceFilter("mark", func(value any, _ ...any) (any, error) {
		return "new:" + toString(value), nil
	})
	if !errors.Is(err, ErrEngineCompiled) {
		t.Fatalf("ReplaceFilter(mark) err = %v, want ErrEngineCompiled", err)
	}

	got, err = engine.Render("page.txt", Data{"word": "ok"})
	if err != nil {
		t.Fatalf("Render(page.txt) after ReplaceFilter err = %v", err)
	}
	if got != "old:ok" {
		t.Errorf("Render(page.txt) after ReplaceFilter = %q, want %q", got, "old:ok")
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

func TestEngineRenderTo_ReturnsWriterError(t *testing.T) {
	t.Parallel()

	engine := New(WithLoader(NewMemoryLoader(map[string]string{
		"a.txt": "hello",
	})))

	err := engine.RenderTo("a.txt", &errWriter{err: errMockWrite}, nil)
	if !errors.Is(err, errMockWrite) {
		t.Errorf("RenderTo() err = %v, want %v", err, errMockWrite)
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
	if _, ok := errors.AsType[*ParseError](err); !ok {
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

	engine := New(
		WithFormat(FormatHTML),
		WithFilter("bang", func(value any, _ ...any) (any, error) {
			return toString(value) + "!", nil
		}),
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
	if err := engine.RegisterFilter("alpha", func(value any, _ ...any) (any, error) {
		return "alpha:" + toString(value), nil
	}); err != nil {
		t.Fatalf("RegisterFilter(alpha) err = %v", err)
	}

	clone := engine.Clone()
	if err := clone.RegisterFilter("beta", func(value any, _ ...any) (any, error) {
		return "beta:" + toString(value), nil
	}); err != nil {
		t.Fatalf("clone RegisterFilter(beta) err = %v", err)
	}

	if engine.HasFilter("beta") {
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

func TestEngineCloneAllowsRegistrationAfterOriginalCompiled(t *testing.T) {
	t.Parallel()

	engine := New()
	if _, err := engine.ParseString(`{{ word }}`); err != nil {
		t.Fatalf("ParseString() err = %v", err)
	}
	if err := engine.RegisterFilter("late", func(value any, _ ...any) (any, error) {
		return value, nil
	}); !errors.Is(err, ErrEngineCompiled) {
		t.Fatalf("RegisterFilter(late) err = %v, want ErrEngineCompiled", err)
	}

	clone := engine.Clone()
	if err := clone.RegisterFilter("late", func(value any, _ ...any) (any, error) {
		return "clone:" + toString(value), nil
	}); err != nil {
		t.Fatalf("clone RegisterFilter(late) err = %v", err)
	}

	tpl, err := clone.ParseString(`{{ word | late }}`)
	if err != nil {
		t.Fatalf("clone ParseString() err = %v", err)
	}
	got, err := tpl.Render(Data{"word": "ok"})
	if err != nil {
		t.Fatalf("clone Render() err = %v", err)
	}
	if got != "clone:ok" {
		t.Errorf("clone Render() = %q, want %q", got, "clone:ok")
	}
}

func TestEngineResetDoesNotReopenRegistries(t *testing.T) {
	t.Parallel()

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"page.txt": `ok`,
		})),
	)
	if _, err := engine.Load("page.txt"); err != nil {
		t.Fatalf("Load(page.txt) err = %v", err)
	}

	engine.Reset()

	err := engine.RegisterFilter("late", func(value any, _ ...any) (any, error) {
		return value, nil
	})
	if !errors.Is(err, ErrEngineCompiled) {
		t.Fatalf("RegisterFilter(late) after Reset err = %v, want ErrEngineCompiled", err)
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

	parseErr, ok := errors.AsType[*ParseError](err)
	if !ok {
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

	probeTag := func(_ *parser, start *token, args *parser) (statement, error) {
		parseCount.Add(1)
		if args.Remaining() > 0 {
			return nil, args.Error("probe does not take arguments")
		}
		return newTextNode("ok", start.Line, start.Col), nil
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
		withTag("probe", probeTag),
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

	gateTag := func(_ *parser, start *token, args *parser) (statement, error) {
		if args.Remaining() > 0 {
			return nil, args.Error("gate does not take arguments")
		}
		startOnce.Do(func() { close(started) })
		<-release
		return newTextNode("ok", start.Line, start.Col), nil
	}

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `{% gate %}`,
		})),
		withTag("gate", gateTag),
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

	probeTag := func(_ *parser, start *token, args *parser) (statement, error) {
		parseCount.Add(1)
		if args.Remaining() > 0 {
			return nil, args.Error("probe does not take arguments")
		}
		return newTextNode("ok", start.Line, start.Col), nil
	}

	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `{% probe %}`,
		})),
		withTag("probe", probeTag),
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

	probeTag := func(_ *parser, start *token, args *parser) (statement, error) {
		parseCount.Add(1)
		if args.Remaining() > 0 {
			return nil, args.Error("probe does not take arguments")
		}
		return newTextNode("ok", start.Line, start.Col), nil
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
		withTag("probe", probeTag),
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
