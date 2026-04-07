package template

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"sync"
)

// Set is the entry point for the multi-file template system. A Set holds
// a Loader, a template cache, and rendering mode (HTML vs text).
//
// Use NewHTMLSet for HTML output (auto-escapes {{ expr }}) or NewTextSet
// for non-HTML output (plain text, code generation, config files).
//
// Sets are safe for concurrent use: compiled templates are cached and
// treated as read-only after compilation, so multiple goroutines can
// Render concurrently against the same Set.
type Set struct {
	loader     Loader
	autoescape bool
	globals    Context

	// tags is the per-Set tag registry, layered on top of the global
	// registry. Layout tags (include/extends/block) live here so
	// Compile(src) does not see them.
	tags *TagRegistry

	// filters is the per-Set filter registry, layered on top of the
	// global registry. The safe filter and (in HTML mode) HTML-aware
	// escape variants live here so Compile(src) does not see them.
	filters *Registry

	mu      sync.RWMutex
	cache   map[string]*Template
	parsing map[string]bool // resolved names currently mid-compile
}

// NewTextSet constructs a Set for plain-text output. {{ expr }} is rendered
// without any escaping. Use this for code generation, config files, YAML,
// TOML, Taskfile, plain-text emails, etc.
func NewTextSet(loader Loader, opts ...SetOption) *Set {
	return newSet(loader, false, opts...)
}

// NewHTMLSet constructs a Set for HTML output. {{ expr }} is automatically
// HTML-escaped; values wrapped in SafeString (e.g. via the safe filter)
// are emitted verbatim. Use this for HTML pages, HTML emails, and any
// other HTML-context output.
func NewHTMLSet(loader Loader, opts ...SetOption) *Set {
	return newSet(loader, true, opts...)
}

// SetOption configures a Set at construction time.
type SetOption func(*Set)

// WithGlobals injects variables available to every render call. Render-time
// ctx keys override globals on conflict.
func WithGlobals(g Context) SetOption {
	return func(s *Set) {
		s.globals = g
	}
}

func newSet(loader Loader, autoescape bool, opts ...SetOption) *Set {
	s := &Set{
		loader:     loader,
		autoescape: autoescape,
		cache:      make(map[string]*Template),
		parsing:    make(map[string]bool),
		tags:       NewTagRegistry(),
		filters:    NewRegistry(),
	}
	// Layer the per-Set tag registry on top of the global one so Sets
	// see if/for/break/continue plus the layout tags, while Compile(src)
	// sees only the global entries.
	s.tags.parent = defaultTagRegistry
	for _, bt := range layoutTags {
		if err := s.tags.Register(bt.name, bt.parser); err != nil {
			panic(fmt.Sprintf("template: failed to register layout tag %q: %v", bt.name, err))
		}
	}
	// Layer the per-Set filter registry on top of the global one. The
	// safe filter is always available in Set-loaded templates. In
	// HTML mode, we override escape/escape_once/h with SafeString-
	// returning variants so the auto-escape output path doesn't
	// double-escape their results.
	s.filters.parent = defaultRegistry
	s.filters.Register("safe", safeFilter)
	if autoescape {
		s.filters.Register("escape", escapeFilterSafe)
		s.filters.Register("escape_once", escapeOnceFilterSafe)
		s.filters.Register("h", escapeFilterSafe)
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Get loads and compiles the named template, caching the result. Subsequent
// calls with the same name return the same *Template instance until Reset
// is called.
func (s *Set) Get(name string) (*Template, error) {
	s.mu.RLock()
	if tpl, ok := s.cache[name]; ok {
		s.mu.RUnlock()
		return tpl, nil
	}
	s.mu.RUnlock()

	src, resolved, err := s.loader.Open(name)
	if err != nil {
		return nil, err
	}

	// Mark this template as mid-compile so nested {% include %} of the
	// same name can detect the cycle and downgrade to lazy mode instead
	// of infinite-recursing at parse time.
	s.markParsing(resolved, true)
	defer s.markParsing(resolved, false)

	tpl, err := compileForSet(src, s)
	if err != nil {
		return nil, err
	}
	tpl.name = resolved
	tpl.set = s

	s.mu.Lock()
	defer s.mu.Unlock()
	if cached, ok := s.cache[name]; ok {
		return cached, nil
	}
	s.cache[name] = tpl
	return tpl, nil
}

// isParsing reports whether a template with the given resolved name is
// currently mid-compile. Used by {% include %} to detect parse-time
// circular references.
func (s *Set) isParsing(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.parsing[name]
}

// markParsing toggles the parsing flag for a resolved name.
func (s *Set) markParsing(name string, v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v {
		s.parsing[name] = true
	} else {
		delete(s.parsing, name)
	}
}

// Render loads the named template and writes its output to w.
// ctx is merged with any globals configured via WithGlobals; ctx keys take
// precedence over globals.
func (s *Set) Render(name string, ctx Context, w io.Writer) error {
	tpl, err := s.Get(name)
	if err != nil {
		return err
	}
	merged := s.mergeContext(ctx)
	ec := NewExecutionContext(merged)
	ec.set = s
	ec.autoescape = s.autoescape
	return tpl.Execute(ec, w)
}

// RenderString is a convenience wrapper around Render returning a string.
func (s *Set) RenderString(name string, ctx Context) (string, error) {
	var buf bytes.Buffer
	if err := s.Render(name, ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Reset clears the template cache. Intended for dev-server hot-reload
// workflows: call Reset after template files change on disk to force
// recompilation on the next Render.
func (s *Set) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = make(map[string]*Template)
}

func (s *Set) mergeContext(ctx Context) Context {
	if s.globals == nil {
		return ctx
	}
	merged := make(Context, len(s.globals)+len(ctx))
	maps.Copy(merged, s.globals)
	maps.Copy(merged, ctx)
	return merged
}
