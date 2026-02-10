package template

import (
	"fmt"
	"sort"
	"sync"
)

// FilterFunc represents the signature of functions that can be applied as filters.
// It takes a value and optional string arguments, returning the transformed value or an error.
type FilterFunc func(value interface{}, args ...string) (interface{}, error)

// Global filter registry
var (
	filterRegistry = make(map[string]FilterFunc)
	filterMu       sync.RWMutex
)

// RegisterFilter registers a filter function with the given name.
// It is safe to call from multiple goroutines.
// If a filter with the same name already exists, it will be overwritten.
func RegisterFilter(name string, fn FilterFunc) {
	filterMu.Lock()
	defer filterMu.Unlock()

	if fn == nil {
		panic(fmt.Sprintf("parse: filter function for '%s' is nil", name))
	}

	filterRegistry[name] = fn
}

// GetFilter retrieves a registered filter by name.
// Returns the filter function and true if found, nil and false otherwise.
// It is safe to call from multiple goroutines.
func GetFilter(name string) (FilterFunc, bool) {
	filterMu.RLock()
	defer filterMu.RUnlock()

	fn, ok := filterRegistry[name]
	return fn, ok
}

// ListFilters returns a sorted list of all registered filter names.
// It is safe to call from multiple goroutines.
func ListFilters() []string {
	filterMu.RLock()
	defer filterMu.RUnlock()

	names := make([]string, 0, len(filterRegistry))
	for name := range filterRegistry {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// HasFilter checks if a filter with the given name is registered.
// It is safe to call from multiple goroutines.
func HasFilter(name string) bool {
	filterMu.RLock()
	defer filterMu.RUnlock()

	_, ok := filterRegistry[name]
	return ok
}

// UnregisterFilter removes a filter from the registry.
// It is safe to call from multiple goroutines.
func UnregisterFilter(name string) {
	filterMu.Lock()
	defer filterMu.Unlock()

	delete(filterRegistry, name)
}

// ========================================
// Built-in Filter Registration
// ========================================

// init automatically registers all built-in filters when the package is imported.
// The register*Filters() functions are implemented in filter_*.go files.
func init() {
	registerStringFilters()
	registerMathFilters()
	registerArrayFilters()
	registerMapFilters()
	registerDateFilters()
	registerFormatFilters()
	registerNumberFilters()
}
