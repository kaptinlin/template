package template

import (
	"fmt"
	"slices"
	"sync"
)

// FilterFunc represents the signature of functions that can be applied as filters.
type FilterFunc func(value interface{}, args ...string) (interface{}, error)

// Registry is a concurrency-safe collection of named filter functions.
// Use [NewRegistry] to create an instance, or use the package-level
// functions that operate on the default registry.
type Registry struct {
	mu      sync.RWMutex
	filters map[string]FilterFunc
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
		panic(fmt.Sprintf("template: Register filter %q with nil function", name))
	}

	r.mu.Lock()
	r.filters[name] = fn
	r.mu.Unlock()
}

// Filter returns the filter registered under name and a boolean
// indicating whether it was found.
func (r *Registry) Filter(name string) (FilterFunc, bool) {
	r.mu.RLock()
	fn, ok := r.filters[name]
	r.mu.RUnlock()
	return fn, ok
}

// Has reports whether a filter with the given name is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	_, ok := r.filters[name]
	r.mu.RUnlock()
	return ok
}

// List returns the names of all registered filters in sorted order.
func (r *Registry) List() []string {
	r.mu.RLock()
	names := make([]string, 0, len(r.filters))
	for name := range r.filters {
		names = append(names, name)
	}
	r.mu.RUnlock()

	slices.Sort(names)
	return names
}

// Unregister removes the filter registered under name.
// It is a no-op if no such filter exists.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	delete(r.filters, name)
	r.mu.Unlock()
}

// defaultRegistry is the package-level registry used by the convenience
// functions below and by the built-in filter registrations in init.
var defaultRegistry = NewRegistry()

// RegisterFilter registers a filter in the default registry.
// It panics if fn is nil.
func RegisterFilter(name string, fn FilterFunc) {
	defaultRegistry.Register(name, fn)
}

// GetFilter retrieves a filter from the default registry.
func GetFilter(name string) (FilterFunc, bool) {
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
