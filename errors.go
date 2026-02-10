package template

import "errors"

// Context navigation errors.
var (
	ErrContextKeyNotFound     = errors.New("key not found in context")
	ErrContextInvalidKeyType  = errors.New("invalid key type for navigation")
	ErrContextIndexOutOfRange = errors.New("index out of range in context")
)

// Filter errors.
var (
	ErrFilterNotFound               = errors.New("filter not found")
	ErrFilterExecutionFailed        = errors.New("filter execution failed")
	ErrFilterInputInvalid           = errors.New("filter input is invalid")
	ErrFilterArgsInvalid            = errors.New("filter arguments are invalid")
	ErrFilterInputEmpty             = errors.New("filter input is empty")
	ErrFilterInputNotSlice          = errors.New("filter input is not a slice")
	ErrFilterInputNotNumeric        = errors.New("filter input is not numeric")
	ErrFilterInputInvalidTimeFormat = errors.New("filter input has an invalid time format")
	ErrFilterInputUnsupportedType   = errors.New("filter input is of an unsupported type")
	ErrInsufficientArgs             = errors.New("insufficient arguments provided")
	ErrInvalidFilterName            = errors.New("invalid filter name")
	ErrUnknownFilterArgumentType    = errors.New("unknown argument type")
	ErrExpectedFilterName           = errors.New("expected filter name after '|'")
)

// Lexer errors.
var (
	ErrUnexpectedCharacter = errors.New("unexpected character")
	ErrUnterminatedString  = errors.New("unterminated string literal")
)

// Parser errors.
var (
	ErrInvalidNumber   = errors.New("invalid number")
	ErrExpectedRParen  = errors.New("expected ')'")
	ErrUnexpectedToken = errors.New("unexpected token")
	ErrUnknownNodeType = errors.New("unknown node type")
	ErrIntegerOverflow = errors.New("unsigned integer value exceeds maximum int64 value")
)

// Type and operator errors.
var (
	ErrUnsupportedType     = errors.New("unsupported type")
	ErrUnsupportedOperator = errors.New("unsupported operator")
	ErrUnsupportedUnaryOp  = errors.New("unsupported unary operator")
)

// Variable and property access errors.
var (
	ErrUndefinedVariable     = errors.New("undefined variable")
	ErrUndefinedProperty     = errors.New("undefined property")
	ErrNonStructProperty     = errors.New("cannot access property of non-struct value")
	ErrCannotAccessProperty  = errors.New("cannot access property")
	ErrNonObjectProperty     = errors.New("cannot access property of non-object")
	ErrInvalidVariableAccess = errors.New("invalid variable access")
)

// Arithmetic errors.
var (
	ErrCannotAddTypes       = errors.New("cannot add values of these types")
	ErrCannotSubtractTypes  = errors.New("cannot subtract values of these types")
	ErrCannotMultiplyTypes  = errors.New("cannot multiply values of these types")
	ErrCannotDivideTypes    = errors.New("cannot divide values of these types")
	ErrCannotModuloTypes    = errors.New("cannot modulo values of these types")
	ErrDivisionByZero       = errors.New("division by zero")
	ErrModuloByZero         = errors.New("modulo by zero")
	ErrCannotConvertToBool  = errors.New("cannot convert type to boolean")
	ErrCannotNegate         = errors.New("cannot negate value")
	ErrCannotApplyUnaryPlus = errors.New("cannot apply unary plus")
	ErrCannotCompareTypes   = errors.New("cannot compare values of these types")
)

// Indexing and field access errors.
var (
	ErrInvalidIndexType      = errors.New("invalid index type")
	ErrInvalidArrayIndex     = errors.New("invalid array index")
	ErrIndexOutOfRange       = errors.New("index out of range")
	ErrCannotIndexNil        = errors.New("cannot index nil")
	ErrTypeNotIndexable      = errors.New("type is not indexable")
	ErrCannotGetKeyFromNil   = errors.New("cannot get key from nil")
	ErrTypeNotMap            = errors.New("type is not a map")
	ErrCannotGetFieldFromNil = errors.New("cannot get field from nil")
	ErrStructHasNoField      = errors.New("struct has no field")
	ErrTypeHasNoField        = errors.New("type has no field")
	ErrUnsupportedArrayType  = errors.New("unsupported array type")
)

// Collection and iteration errors.
var (
	ErrUnsupportedCollectionType = errors.New("unsupported collection type for for loop")
	ErrTypeNotIterable           = errors.New("type is not iterable")
	ErrTypeHasNoLength           = errors.New("type has no length")
)

// Type conversion errors.
var (
	ErrCannotConvertNilToInt   = errors.New("cannot convert nil to int")
	ErrCannotConvertToInt      = errors.New("cannot convert value to int")
	ErrCannotConvertNilToFloat = errors.New("cannot convert nil to float")
	ErrCannotConvertToFloat    = errors.New("cannot convert value to float")
)

// Control flow errors.
var (
	ErrBreakOutsideLoop    = errors.New("break statement outside of loop")
	ErrContinueOutsideLoop = errors.New("continue statement outside of loop")
)

// Tag registration and parsing errors.
var (
	ErrTagAlreadyRegistered = errors.New("tag already registered")

	// If/elif/else tag errors.
	ErrMultipleElseStatements         = errors.New("multiple 'else' statements found in if block, use 'elif' for additional conditions")
	ErrUnexpectedTokensAfterCondition = errors.New("unexpected tokens after condition")
	ErrElseNoArgs                     = errors.New("else does not take arguments")
	ErrEndifNoArgs                    = errors.New("endif does not take arguments")
	ErrElifAfterElse                  = errors.New("elif cannot appear after else")

	// For tag errors.
	ErrExpectedVariable                = errors.New("expected variable name")
	ErrExpectedSecondVariable          = errors.New("expected second variable name after comma")
	ErrExpectedInKeyword               = errors.New("expected 'in' keyword")
	ErrUnexpectedTokensAfterCollection = errors.New("unexpected tokens after collection")
	ErrEndforNoArgs                    = errors.New("endfor does not take arguments")

	// Break/continue tag errors.
	ErrBreakNoArgs    = errors.New("break does not take arguments")
	ErrContinueNoArgs = errors.New("continue does not take arguments")
)
