package template

import (
	"errors"
	"fmt"
	"regexp"
)

// ErrFilterNotFound indicates that the requested filter was not found in the global registry.
var ErrFilterNotFound = errors.New("filter not found")

// ErrFilterInputInvalid indicates an issue with the filter input value being of an unexpected type or format.
var ErrFilterInputInvalid = errors.New("filter input is invalid")

// ErrFilterArgsInvalid indicates an issue with the filter arguments, such as wrong type, format, or number of arguments.
var ErrFilterArgsInvalid = errors.New("filter arguments are invalid")

// ErrFilterInputEmpty indicates that the input value is empty or nil.
var ErrFilterInputEmpty = errors.New("filter input is empty")

// ErrInsufficientArgs indicates that the filter was called with insufficient arguments.
var ErrInsufficientArgs = errors.New("insufficient arguments provided")

// FilterFunc represents the signature of functions that can be applied as filters.
type FilterFunc func(interface{}, ...string) (interface{}, error)

var filters = make(map[string]FilterFunc)

// Global variable for validating filter names
var validFilterNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

// RegisterFilter adds a filter to the global registry with name validation.
func RegisterFilter(name string, fn FilterFunc) error {
	if !validFilterNameRegex.MatchString(name) {
		return fmt.Errorf("invalid filter name '%s'", name)
	}
	filters[name] = fn
	return nil
}

// ApplyFilters executes a series of filters on a value within a context.
func ApplyFilters(value interface{}, fs []Filter, ctx Context) (interface{}, error) {
	var err error
	for _, f := range fs {
		fn, exists := filters[f.Name]
		if !exists {
			return value, fmt.Errorf("%w: filter '%s' not found", ErrFilterNotFound, f.Name)
		}
		// Apply each filter without converting the value to a string immediately.
		value, err = fn(value, f.Args...)
		if err != nil {
			return value, fmt.Errorf("error applying '%s': %w", f.Name, err)
		}
	}
	return value, nil
}

// Filter defines a transformation to apply to a template variable.
type Filter struct {
	Name string
	Args []string
}
