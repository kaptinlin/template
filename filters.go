package template

import (
	"fmt"
	"regexp"
	"sync"
)

// FilterFunc represents the signature of functions that can be applied as filters.
// Arguments are always passed as strings, even if the source in the template was a number literal or other type.
// Filter authors should parse string arguments into their expected types if necessary.
type FilterFunc func(any, ...string) (any, error)

var (
	filters   = make(map[string]FilterFunc)
	filtersMu sync.RWMutex
)

// validFilterNameRegex validates filter names.
var validFilterNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

// RegisterFilter adds a filter to the global registry with name validation.
func RegisterFilter(name string, fn FilterFunc) error {
	if !validFilterNameRegex.MatchString(name) {
		return fmt.Errorf("%w: '%s'", ErrInvalidFilterName, name)
	}
	filtersMu.Lock()
	filters[name] = fn
	filtersMu.Unlock()
	return nil
}

// mustRegisterFilters registers multiple filters and panics if any registration fails.
// This is intended for use in init() functions where registration failure indicates a programming error.
func mustRegisterFilters(filtersToRegister map[string]FilterFunc) {
	for name, fn := range filtersToRegister {
		if err := RegisterFilter(name, fn); err != nil {
			panic(fmt.Sprintf("failed to register filter %s: %v", name, err))
		}
	}
}

// ApplyFilters executes a series of filters on a value within a context, supporting variable arguments.
func ApplyFilters(value any, fs []Filter, ctx Context) (any, error) {
	var err error
	for _, f := range fs {
		filtersMu.RLock()
		fn, exists := filters[f.Name]
		filtersMu.RUnlock()
		if !exists {
			return value, fmt.Errorf("filter '%s' not found: %w", f.Name, ErrFilterNotFound)
		}

		// Prepare arguments by checking their types and extracting values for VariableArg.
		args := make([]string, len(f.Args))
		for i, arg := range f.Args {
			switch arg := arg.(type) {
			case StringArg:
				args[i] = arg.Value().(string)
			case NumberArg:
				// Use convertToString for consistent handling
				str, err := convertToString(arg.Value())
				if err != nil {
					return value, fmt.Errorf("could not convert filter argument %d to string: %w", i, err)
				}
				args[i] = str
			case VariableArg:
				val, err := ctx.Get(arg.Value().(string))
				if err != nil {
					return value, fmt.Errorf("variable '%s' not found in context: %w",
						arg.Value().(string), ErrContextKeyNotFound)
				}
				// Use convertToString for consistent handling
				str, err := convertToString(val)
				if err != nil {
					return value, fmt.Errorf("could not convert variable '%s' to string: %w", arg.Value().(string), err)
				}
				args[i] = str
			default:
				return value, fmt.Errorf("filter '%s': %w", f.Name, ErrUnknownFilterArgumentType)
			}
		}

		// Apply each filter with the prepared arguments.
		value, err = fn(value, args...)
		if err != nil {
			return value, fmt.Errorf("filter '%s' failed: %w", f.Name, err)
		}
	}
	return value, nil
}

// Filter defines a transformation to apply to a template variable.
type Filter struct {
	Name string
	Args []FilterArg
}

// FilterArg represents the interface for filter arguments.
type FilterArg interface {
	Value() any
	Type() string
}

// StringArg holds a string argument.
type StringArg struct {
	val string
}

// Value returns the string argument value.
func (a StringArg) Value() any {
	return a.val
}

// Type returns the argument type as "string".
func (a StringArg) Type() string {
	return "string"
}

// NumberArg holds a number argument.
type NumberArg struct {
	val float64
}

// Value returns the number argument value.
func (a NumberArg) Value() any {
	return a.val
}

// Type returns the argument type as "number".
func (a NumberArg) Type() string {
	return "number"
}

// VariableArg holds a variable argument.
type VariableArg struct {
	name string
}

// Value returns the variable name.
func (a VariableArg) Value() any {
	return a.name
}

// Type returns the argument type as "variable".
func (a VariableArg) Type() string {
	return "variable"
}
