package template

import (
	"fmt"
	"maps"
	"slices"
	"sync"
)

// FilterFunc represents the signature of functions that can be applied as filters.
type FilterFunc func(value any, args ...any) (any, error)

// Registry is a concurrency-safe collection of named filter functions.
// Use [NewRegistry] to create an instance, or use the package-level
// functions that operate on the default registry.
//
// Registry supports an optional parent: when a lookup misses in this
// registry, the parent is consulted. This allows an Engine to layer its
// own private filters (like safe and the HTML-aware escape) over the
// global registry without copying entries.
type Registry struct {
	mu      sync.RWMutex
	filters map[string]FilterFunc
	parent  *Registry
}

// NewRegistry creates an empty filter registry.
func NewRegistry() *Registry {
	return &Registry{filters: make(map[string]FilterFunc)}
}

func (r *Registry) validate(name string, fn FilterFunc) {
	if fn == nil {
		panic(fmt.Sprintf("template: nil filter function for %q", name))
	}
}

// Register adds or replaces a filter function under the given name.
//
// Register preserves the original Registry behavior for compatibility.
// For new code, prefer [Registry.Replace] when overwrite semantics are
// intentional, or introduce a higher-level guard before calling Register.
func (r *Registry) Register(name string, fn FilterFunc) {
	r.Replace(name, fn)
}

// Replace stores fn under name, overwriting any direct existing entry.
//
// Replace panics if fn is nil.
func (r *Registry) Replace(name string, fn FilterFunc) {
	r.validate(name, fn)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.filters[name] = fn
}

// MustRegister is an explicit panic-on-failure registration helper.
//
// For Registry this currently matches [Registry.Replace], and exists to make
// bootstrap code read consistently with [TagRegistry.MustRegister].
func (r *Registry) MustRegister(name string, fn FilterFunc) {
	r.Replace(name, fn)
}

// Filter returns the filter registered under name and a boolean
// indicating whether it was found. If not found locally and a parent
// registry is set, the parent is consulted.
func (r *Registry) Filter(name string) (FilterFunc, bool) {
	r.mu.RLock()
	fn, ok := r.filters[name]
	r.mu.RUnlock()
	if ok {
		return fn, true
	}
	if r.parent != nil {
		return r.parent.Filter(name)
	}
	return nil, false
}

// Has reports whether a filter with the given name is registered
// (including in ancestor registries).
func (r *Registry) Has(name string) bool {
	_, ok := r.Filter(name)
	return ok
}

// List returns the names of all registered filters in sorted order.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.filters) == 0 {
		return nil
	}
	return slices.Sorted(maps.Keys(r.filters))
}

// Unregister removes the filter registered under name.
// It is a no-op if no such filter exists.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.filters, name)
}

// Clone returns a shallow copy of the registry and its direct entries.
// The parent registry reference is preserved.
func (r *Registry) Clone() *Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cloned := NewRegistry()
	maps.Copy(cloned.filters, r.filters)
	cloned.parent = r.parent
	return cloned
}

// defaultRegistry is the package-level built-in filter registry used as the
// parent layer for engine-local registries.
var defaultRegistry = NewRegistry()

// init registers all built-in filters when the package is imported.
// The register*Filters functions are implemented in filter_*.go files.
func init() {
	registerStringFilters()
	registerMathFilters()
	registerArrayFilters()
	registerMapFilters()
	registerDateFilters()
	registerFormatFilters()
	registerNumberFilters()
}
