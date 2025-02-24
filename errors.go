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

	// ErrExpectedFilterName is returned when a filter name is expected after the pipe symbol.
	ErrExpectedFilterName = errors.New("expected filter name after '|'")

	// ErrInvalidNumber is returned when a number cannot be parsed.
	ErrInvalidNumber = errors.New("invalid number")

	// ErrExpectedRParen is returned when a right parenthesis is expected.
	ErrExpectedRParen = errors.New("expected ')'")

	// ErrUnexpectedToken is returned when an unexpected token is encountered.
	ErrUnexpectedToken = errors.New("unexpected token")

	// ErrUnsupportedType is returned when an unsupported type is encountered.
	ErrUnsupportedType = errors.New("unsupported type")

	// ErrUnsupportedOperator is returned when an unsupported operator is used.
	ErrUnsupportedOperator = errors.New("unsupported operator")

	// ErrUnsupportedUnaryOp is returned when an unsupported unary operator is used.
	ErrUnsupportedUnaryOp = errors.New("unsupported unary operator")

	// ErrUndefinedVariable is returned when a variable is not found in the context.
	ErrUndefinedVariable = errors.New("undefined variable")

	// ErrUndefinedProperty is returned when a property is not found in an object.
	ErrUndefinedProperty = errors.New("undefined property")

	// ErrNonStructProperty is returned when attempting to access a property on a non-struct value.
	ErrNonStructProperty = errors.New("cannot access property of non-struct value")

	// ErrCannotAccessProperty is returned when property access fails.
	ErrCannotAccessProperty = errors.New("cannot access property")

	// ErrCannotAddTypes is returned when attempting to add incompatible types.
	ErrCannotAddTypes = errors.New("cannot add values of these types")

	// ErrCannotSubtractTypes is returned when attempting to subtract incompatible types.
	ErrCannotSubtractTypes = errors.New("cannot subtract values of these types")

	// ErrCannotMultiplyTypes is returned when attempting to multiply incompatible types.
	ErrCannotMultiplyTypes = errors.New("cannot multiply values of these types")

	// ErrDivisionByZero is returned when attempting to divide by zero.
	ErrDivisionByZero = errors.New("division by zero")

	// ErrCannotDivideTypes is returned when attempting to divide incompatible types.
	ErrCannotDivideTypes = errors.New("cannot divide values of these types")

	// ErrCannotConvertToBool is returned when a value cannot be converted to boolean.
	ErrCannotConvertToBool = errors.New("cannot convert type to boolean")

	// ErrCannotCompareTypes is returned when attempting to compare incompatible types.
	ErrCannotCompareTypes = errors.New("cannot compare values of these types")

	// ErrInvalidIndexType is returned when an invalid type is used as an array index.
	ErrInvalidIndexType = errors.New("invalid index type")

	// ErrInvalidArrayIndex is returned when an array index is invalid.
	ErrInvalidArrayIndex = errors.New("invalid array index")

	// ErrIndexOutOfRange is returned when an array index is out of bounds.
	ErrIndexOutOfRange = errors.New("index out of range")

	// ErrUnsupportedArrayType is returned when an unsupported array type is encountered.
	ErrUnsupportedArrayType = errors.New("unsupported array type")

	// ErrNonObjectProperty is returned when attempting to access a property on a non-object value.
	ErrNonObjectProperty = errors.New("cannot access property of non-object")

	// ErrInvalidVariableAccess is returned when variable access is invalid.
	ErrInvalidVariableAccess = errors.New("invalid variable access")

	// ErrUnsupportedCollectionType is returned when an unsupported collection type is used in a for loop.
	ErrUnsupportedCollectionType = errors.New("unsupported collection type for for loop")

	// ErrUnexpectedCharacter is returned when the lexer encounters an unexpected character.
	ErrUnexpectedCharacter = errors.New("unexpected character")

	// ErrUnterminatedString is returned when a string literal is not properly terminated.
	ErrUnterminatedString = errors.New("unterminated string literal")
)
