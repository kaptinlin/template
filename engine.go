package template

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"sync"
)

// Format controls output semantics during rendering.
type Format uint8

const (
	// FormatText renders output without automatic escaping.
	FormatText Format = iota
	// FormatHTML automatically HTML-escapes {{ expr }} output unless it is safe.
	FormatHTML
)

// Feature is an optional language capability that can be enabled per engine.
type Feature uint8

const (
	// FeatureLayout enables include, extends, block, raw, and safe-aware engine behavior.
	FeatureLayout Feature = 1 << iota
)

// Engine is the entry point for the loader-backed template system.
//
// An Engine holds a loader, compile cache, rendering format, feature flags,
// and per-engine tag/filter registries. Engines are safe for concurrent use:
// compiled templates are cached and treated as read-only after compilation.
type Engine struct {
	loader   Loader
	format   Format
	features Feature
	defaults Data

	// tags is the per-engine tag registry, layered on top of the global
	// registry. Optional language features register into this private layer.
	tags *tagRegistry

	// filters is the per-engine filter registry, layered on top of the
	// global registry. Engine-specific overrides live here.
	filters *registry

	mu        sync.RWMutex
	cache     map[string]*Template
	aliases   map[string]string
	loading   map[string]*loadCall
	parsing   map[string]bool // resolved names currently mid-compile
	configErr error
	compiled  bool
}

type loadCall struct {
	done chan struct{}
	tpl  *Template
	err  error
}

// EngineOption configures an Engine at construction time.
type EngineOption func(*Engine)

// New constructs an Engine with the provided options.
func New(opts ...EngineOption) *Engine {
	e := &Engine{
		format:  FormatText,
		tags:    newTagRegistry(),
		filters: newRegistry(),
		cache:   make(map[string]*Template),
		aliases: make(map[string]string),
		loading: make(map[string]*loadCall),
		parsing: make(map[string]bool),
	}
	e.tags.parent = defaultTagRegistry
	e.filters.parent = defaultRegistry
	for _, opt := range opts {
		opt(e)
	}
	e.rebuildRegistries()
	return e
}

// WithLoader configures the Engine loader used by Load and Render.
func WithLoader(loader Loader) EngineOption {
	return func(e *Engine) {
		e.loader = loader
	}
}

// WithFormat configures how output is rendered.
func WithFormat(format Format) EngineOption {
	return func(e *Engine) {
		e.format = format
	}
}

// WithFeatures enables optional language features.
func WithFeatures(features ...Feature) EngineOption {
	return func(e *Engine) {
		for _, feature := range features {
			e.features |= feature
		}
	}
}

// WithLayout enables layout features such as include, extends, block, raw,
// and safe-aware engine behavior.
func WithLayout() EngineOption {
	return WithFeatures(FeatureLayout)
}

// WithDefaults injects variables available to every render call. Render-time
// ctx keys override defaults on conflict.
func WithDefaults(g Data) EngineOption {
	return func(e *Engine) {
		e.defaults = g
	}
}

// WithFilter registers an engine-local filter at construction time.
func WithFilter(name string, fn FilterFunc) EngineOption {
	return func(e *Engine) {
		if fn == nil {
			e.addConfigError(fmt.Errorf("filter %q: %w", name, errNilFilterFunction))
			return
		}
		e.filters.Replace(name, fn)
	}
}

func (e *Engine) rebuildRegistries() {
	if e.tags == nil {
		e.tags = newTagRegistry()
	}
	e.tags.parent = defaultTagRegistry

	if e.filters == nil {
		e.filters = newRegistry()
	}
	e.filters.parent = defaultRegistry

	if e.features&FeatureLayout != 0 {
		for _, bt := range layoutTags {
			e.tags.Replace(bt.name, bt.parser)
		}
		e.filters.Replace("safe", safeFilter)
	}
	if e.format == FormatHTML {
		e.filters.Replace("escape", escapeFilterSafe)
		e.filters.Replace("escape_once", escapeOnceFilterSafe)
		e.filters.Replace("h", escapeFilterSafe)
	}
}

// HasFilter reports whether this engine can apply the named filter.
func (e *Engine) HasFilter(name string) bool {
	return e.filters.Has(name)
}

// RegisterFilter adds or replaces an engine-local filter for future template
// compilation. Prefer [WithFilter] for normal construction-time setup; templates
// that are already compiled keep the filter functions they resolved at compile
// time. Once this engine starts compiling templates, RegisterFilter returns
// [ErrEngineCompiled]; use [Engine.Clone] to derive a new configurable engine.
func (e *Engine) RegisterFilter(name string, fn FilterFunc) error {
	return e.mutateRegistry("register filter", name, func() error {
		if fn == nil {
			return fmt.Errorf("register filter %q: %w", name, errNilFilterFunction)
		}
		e.filters.Register(name, fn)
		return nil
	})
}

// ReplaceFilter overwrites an engine-local filter for future template
// compilation. Prefer [WithFilter] for normal construction-time setup. Once
// this engine starts compiling templates, ReplaceFilter returns
// [ErrEngineCompiled]; use [Engine.Clone] to derive a new configurable engine.
func (e *Engine) ReplaceFilter(name string, fn FilterFunc) error {
	return e.mutateRegistry("replace filter", name, func() error {
		if fn == nil {
			return fmt.Errorf("replace filter %q: %w", name, errNilFilterFunction)
		}
		e.filters.Replace(name, fn)
		return nil
	})
}

// Clone copies engine configuration into a fresh Engine with an empty cache.
func (e *Engine) Clone(opts ...EngineOption) *Engine {
	e.mu.RLock()
	defer e.mu.RUnlock()

	clone := &Engine{
		loader:    e.loader,
		format:    e.format,
		features:  e.features,
		defaults:  maps.Clone(e.defaults),
		tags:      e.tags.Clone(),
		filters:   e.filters.Clone(),
		cache:     make(map[string]*Template),
		aliases:   make(map[string]string),
		loading:   make(map[string]*loadCall),
		parsing:   make(map[string]bool),
		configErr: e.configErr,
	}
	for _, opt := range opts {
		opt(clone)
	}
	clone.rebuildRegistries()
	return clone
}

// HasFeature reports whether the engine has an optional feature enabled.
func (e *Engine) HasFeature(feature Feature) bool {
	return e.features&feature != 0
}

// ParseString compiles a template source string in the context of this engine.
func (e *Engine) ParseString(source string) (*Template, error) {
	if err := e.configError(); err != nil {
		return nil, err
	}
	e.markCompiled()
	tpl, err := compileNamedForEngine("", source, e)
	if err != nil {
		return nil, err
	}
	tpl.engine = e
	return tpl, nil
}

// Load resolves, compiles, and caches the named template.
func (e *Engine) Load(name string) (*Template, error) {
	if err := e.configError(); err != nil {
		return nil, err
	}
	if e.loader == nil {
		return nil, wrapTemplateSourceError(name, ErrTemplateNotFound)
	}

	if tpl, ok := e.cachedTemplate(name); ok {
		return tpl, nil
	}

	src, resolved, err := e.loader.Open(name)
	if err != nil {
		return nil, wrapTemplateSourceError(name, err)
	}
	return e.loadOpened(name, src, resolved)
}

func (e *Engine) loadOpened(name, src, resolved string) (*Template, error) {
	tpl, call, wait := e.beginLoad(name, resolved)
	if tpl != nil {
		return tpl, nil
	}
	if wait {
		<-call.done
		return call.tpl, call.err
	}

	e.markCompiled()
	e.markParsing(resolved, true)
	defer e.markParsing(resolved, false)

	tpl, err := compileNamedForEngine(resolved, src, e)
	if err != nil {
		e.finishLoad(resolved, call, nil, err)
		return nil, err
	}
	tpl.name = resolved
	tpl.engine = e

	e.finishLoad(resolved, call, tpl, nil)
	return tpl, nil
}

func (e *Engine) loadInclude(name string) (*Template, bool, error) {
	tpl, _, parsing, err := e.loadDependency(name)
	return tpl, parsing, err
}

func (e *Engine) loadExtends(name string) (*Template, error) {
	tpl, resolved, parsing, err := e.loadDependency(name)
	if parsing {
		return nil, fmt.Errorf("%w: %q", ErrCircularExtends, resolved)
	}
	return tpl, err
}

// loadDependency resolves a parse-time template dependency without waiting on
// the current parse stack. Static include downgrades to lazy mode when parsing;
// extends turns the same signal into a circular-chain error.
func (e *Engine) loadDependency(name string) (*Template, string, bool, error) {
	if e.loader == nil {
		return nil, "", false, wrapTemplateSourceError(name, ErrTemplateNotFound)
	}
	if tpl, ok := e.cachedTemplate(name); ok {
		return tpl, "", false, nil
	}

	src, resolved, err := e.loader.Open(name)
	if err != nil {
		return nil, "", false, wrapTemplateSourceError(name, err)
	}
	if e.isParsing(resolved) {
		return nil, resolved, true, nil
	}

	tpl, err := e.loadOpened(name, src, resolved)
	return tpl, resolved, false, err
}

func (e *Engine) cachedTemplate(name string) (*Template, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	resolved := name
	if cachedResolved, ok := e.aliases[name]; ok {
		resolved = cachedResolved
	}
	tpl, ok := e.cache[resolved]
	return tpl, ok
}

func (e *Engine) beginLoad(name, resolved string) (*Template, *loadCall, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.aliases[name] = resolved

	if tpl, ok := e.cache[resolved]; ok {
		return tpl, nil, false
	}
	if call, ok := e.loading[resolved]; ok {
		return nil, call, true
	}

	call := &loadCall{done: make(chan struct{})}
	e.loading[resolved] = call
	return nil, call, false
}

func (e *Engine) finishLoad(resolved string, call *loadCall, tpl *Template, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err == nil {
		e.cache[resolved] = tpl
	}
	delete(e.loading, resolved)
	call.tpl = tpl
	call.err = err
	close(call.done)
}

// isParsing reports whether a template with the given resolved name is
// currently mid-compile.
func (e *Engine) isParsing(name string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.parsing[name]
}

// markParsing toggles the parsing flag for a resolved name.
func (e *Engine) markParsing(name string, v bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if v {
		e.parsing[name] = true
	} else {
		delete(e.parsing, name)
	}
}

// Render loads the named template and returns its output as a string.
func (e *Engine) Render(name string, data Data) (string, error) {
	var buf bytes.Buffer
	if err := e.RenderTo(name, &buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderTo loads the named template and writes its output to w.
// data is merged with any defaults configured via WithDefaults; data keys take
// precedence over defaults.
func (e *Engine) RenderTo(name string, w io.Writer, data Data) error {
	tpl, err := e.Load(name)
	if err != nil {
		return err
	}
	merged := e.mergeContext(data)
	ec := newRenderContext(merged)
	ec.engine = e
	ec.autoescape = e.format == FormatHTML
	return tpl.execute(ec, w)
}

// Reset clears the template cache.
func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cache = make(map[string]*Template)
	e.aliases = make(map[string]string)
	e.loading = make(map[string]*loadCall)
}

func (e *Engine) mutateRegistry(action, name string, mutate func() error) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.compiled {
		return fmt.Errorf("%s %q: %w", action, name, ErrEngineCompiled)
	}
	return mutate()
}

func (e *Engine) addConfigError(err error) {
	e.configErr = errors.Join(e.configErr, err)
}

func (e *Engine) configError() error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.configErr
}

func (e *Engine) markCompiled() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.compiled = true
}

func (e *Engine) mergeContext(ctx Data) Data {
	if e.defaults == nil {
		return ctx
	}
	merged := make(Data, len(e.defaults)+len(ctx))
	maps.Copy(merged, e.defaults)
	maps.Copy(merged, ctx)
	return merged
}
