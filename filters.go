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
				// Attempt to get the value of the variable from the context.
				val, err := ctx.Get(arg.Value().(string))
				if err != nil {
					// If the variable is not found in the context, return an error.
					return value, fmt.Errorf("%w: variable '%s' not found in context", ErrContextKeyNotFound, arg.Value().(string))
				}
				// Convert the variable's value to a string since filter functions expect string arguments.
				args[i] = fmt.Sprint(val)
			default:
				return value, fmt.Errorf("unknown argument type for filter '%s'", f.Name)
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

func (a StringArg) Value() interface{} {
	return a.val
}

func (a StringArg) Type() string {
	return "string"
}

// NumberArg holds a number argument.
type NumberArg struct {
	val float64
}

func (a NumberArg) Value() interface{} {
	return a.val
}

func (a NumberArg) Type() string {
	return "number"
}

// VariableArg holds a variable argument.
type VariableArg struct {
	name string
}

func (a VariableArg) Value() interface{} {
	return a.name
}

func (a VariableArg) Type() string {
	return "variable"
}
