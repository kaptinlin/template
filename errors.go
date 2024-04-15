package template

import "errors"

var (
	// ErrContextKeyNotFound is returned when a key is not found in the context.
	ErrContextKeyNotFound = errors.New("key not found in context")

	// ErrContextInvalidKeyType is returned when an unexpected type is encountered while navigating the context.
	ErrContextInvalidKeyType = errors.New("invalid key type for navigation")

	// ErrContextIndexOutOfRange is returned when an index is out of range in the context.
	ErrContextIndexOutOfRange = errors.New("index out of range in context")

	// ErrFilterNotFound indicates that the requested filter was not found in the global registry.
	ErrFilterNotFound = errors.New("filter not found")

	// ErrFilterInputInvalid indicates an issue with the filter input value being of an unexpected type or format.
	ErrFilterInputInvalid = errors.New("filter input is invalid")

	// ErrFilterArgsInvalid indicates an issue with the filter arguments, such as wrong type, format, or number of arguments.
	ErrFilterArgsInvalid = errors.New("filter arguments are invalid")

	// ErrFilterInputEmpty indicates that the input value is empty or nil.
	ErrFilterInputEmpty = errors.New("filter input is empty")

	// ErrInsufficientArgs indicates that the filter was called with insufficient arguments.
	ErrInsufficientArgs = errors.New("insufficient arguments provided")

	// ErrFilterInputNotSlice indicates a filter expected a slice but received a different type.
	ErrFilterInputNotSlice = errors.New("filter input is not a slice")

	// ErrFilterInputNotNumeric indicates a filter expected a numeric value but received a different type.
	ErrFilterInputNotNumeric = errors.New("filter input is not numeric")

	// ErrFilterInputInvalidTimeFormat indicates a filter expected a valid time format but didn't receive it.
	ErrFilterInputInvalidTimeFormat = errors.New("filter input has an invalid time format")

	// ErrFilterInputUnsupportedType indicates the filter received a type it does not support.
	ErrFilterInputUnsupportedType = errors.New("filter input is of an unsupported type")

	// ErrInvalidFilterName is returned when a filter name does not meet the required criteria.
	ErrInvalidFilterName = errors.New("invalid filter name")

	// ErrUnknownFilterArgumentType is returned when a filter argument type is unknown.
	ErrUnknownFilterArgumentType = errors.New("unknown argument type")

	// ErrUnknownNodeType is returned when an unexpected node type is encountered.
	ErrUnknownNodeType = errors.New("unknown node type")
)
