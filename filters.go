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
// registry, the parent is consulted. This allows a Set to layer its
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

// Register adds a filter function under the given name.
// If a filter with the same name already exists, it is overwritten.
//
// Register panics if fn is nil.
func (r *Registry) Register(name string, fn FilterFunc) {
	if fn == nil {
		panic(fmt.Sprintf("template: nil filter function for %q", name))
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.filters[name] = fn
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

// defaultRegistry is the package-level registry used by the convenience
// functions below and by the built-in filter registrations in init.
var defaultRegistry = NewRegistry()

// RegisterFilter registers a filter in the default registry.
// It panics if fn is nil.
func RegisterFilter(name string, fn FilterFunc) {
	defaultRegistry.Register(name, fn)
}

// Filter retrieves a filter from the default registry.
func Filter(name string) (FilterFunc, bool) {
	return defaultRegistry.Filter(name)
}

// ListFilters returns a sorted list of all filter names in the
// default registry.
func ListFilters() []string {
	return defaultRegistry.List()
}

// HasFilter reports whether the default registry contains a filter
// with the given name.
func HasFilter(name string) bool {
	return defaultRegistry.Has(name)
}

// UnregisterFilter removes a filter from the default registry.
func UnregisterFilter(name string) {
	defaultRegistry.Unregister(name)
}

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
