package template

import (
	"fmt"
	"regexp"
)

// FilterFunc represents the signature of functions that can be applied as filters.
type FilterFunc func(interface{}, ...string) (interface{}, error)

var filters = make(map[string]FilterFunc)

// Global variable for validating filter names
var validFilterNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

// RegisterFilter adds a filter to the global registry with name validation.
func RegisterFilter(name string, fn FilterFunc) error {
	if !validFilterNameRegex.MatchString(name) {
		return fmt.Errorf("%w: '%s'", ErrInvalidFilterName, name)
	}
	filters[name] = fn
	return nil
}

// ApplyFilters executes a series of filters on a value within a context, supporting variable arguments.
func ApplyFilters(value interface{}, fs []Filter, ctx Context) (interface{}, error) {
	var err error
	for _, f := range fs {
		fn, exists := filters[f.Name]
		if !exists {
			return value, fmt.Errorf("%w: filter '%s' not found", ErrFilterNotFound, f.Name)
		}

		// Prepare arguments by checking their types and extracting values for VariableArg.
		args := make([]string, len(f.Args))
		for i, arg := range f.Args {
			switch arg := arg.(type) {
			case StringArg:
				args[i] = arg.Value().(string)
			case NumberArg:
				args[i] = fmt.Sprint(arg.Value())
			case VariableArg:
				val, err := ctx.Get(arg.Value().(string))
				if err != nil {
					return value, fmt.Errorf("%w: variable '%s' not found in context", ErrContextKeyNotFound, arg.Value().(string))
				}
				args[i] = fmt.Sprint(val)
			default:
				return value, fmt.Errorf("%w for filter '%s'", ErrUnknownFilterArgumentType, f.Name)
			}
		}

		// Apply each filter with the prepared arguments.
		value, err = fn(value, args...)
		if err != nil {
			return value, fmt.Errorf("error applying '%s' filter: %w", f.Name, err)
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
	Value() interface{}
	Type() string
}

// StringArg holds a string argument.
type StringArg struct {
	val string
}

// Value returns the string argument value.
func (a StringArg) Value() interface{} {
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
func (a NumberArg) Value() interface{} {
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
func (a VariableArg) Value() interface{} {
	return a.name
}

// Type returns the argument type as "variable".
func (a VariableArg) Type() string {
	return "variable"
}
