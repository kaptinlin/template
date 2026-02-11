package template

import "errors"

// ErrContextKeyNotFound indicates a key was not found in the execution context.
// ErrContextInvalidKeyType indicates an invalid key type during context navigation.
// ErrContextIndexOutOfRange indicates an index out of range during context navigation.
var (
	ErrContextKeyNotFound     = errors.New("key not found in context")
	ErrContextInvalidKeyType  = errors.New("invalid key type for navigation")
	ErrContextIndexOutOfRange = errors.New("index out of range in context")
)

// ErrFilterNotFound indicates a referenced filter does not exist.
// ErrFilterExecutionFailed indicates a filter failed during execution.
// ErrFilterInputInvalid indicates the filter received invalid input.
// ErrFilterArgsInvalid indicates the filter received invalid arguments.
// ErrFilterInputEmpty indicates the filter received empty input.
// ErrFilterInputNotSlice indicates the filter expected a slice input.
// ErrFilterInputNotNumeric indicates the filter expected numeric input.
// ErrFilterInputInvalidTimeFormat indicates the filter received an invalid time format.
// ErrFilterInputUnsupportedType indicates the filter received an unsupported input type.
// ErrInsufficientArgs indicates insufficient arguments were provided to a filter.
// ErrInvalidFilterName indicates an invalid filter name was used.
// ErrUnknownFilterArgumentType indicates an unknown argument type was passed to a filter.
// ErrExpectedFilterName indicates a filter name was expected after the pipe operator.
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

// ErrUnexpectedCharacter indicates the lexer encountered an unexpected character.
// ErrUnterminatedString indicates a string literal was not properly closed.
var (
	ErrUnexpectedCharacter = errors.New("unexpected character")
	ErrUnterminatedString  = errors.New("unterminated string literal")
)

// ErrInvalidNumber indicates the parser encountered an invalid numeric literal.
// ErrExpectedRParen indicates a closing parenthesis was expected but not found.
// ErrUnexpectedToken indicates the parser encountered an unexpected token.
// ErrUnknownNodeType indicates an unknown AST node type was encountered.
// ErrIntegerOverflow indicates an unsigned integer value exceeds the maximum int64 value.
var (
	ErrInvalidNumber   = errors.New("invalid number")
	ErrExpectedRParen  = errors.New("expected ')'")
	ErrUnexpectedToken = errors.New("unexpected token")
	ErrUnknownNodeType = errors.New("unknown node type")
	ErrIntegerOverflow = errors.New("unsigned integer value exceeds maximum int64 value")
)

// ErrUnsupportedType indicates an unsupported type was encountered.
// ErrUnsupportedOperator indicates an unsupported operator was used.
// ErrUnsupportedUnaryOp indicates an unsupported unary operator was used.
var (
	ErrUnsupportedType     = errors.New("unsupported type")
	ErrUnsupportedOperator = errors.New("unsupported operator")
	ErrUnsupportedUnaryOp  = errors.New("unsupported unary operator")
)

// ErrUndefinedVariable indicates a referenced variable is not defined.
// ErrUndefinedProperty indicates a referenced property is not defined.
// ErrNonStructProperty indicates a property access was attempted on a non-struct value.
// ErrCannotAccessProperty indicates a property cannot be accessed.
// ErrNonObjectProperty indicates a property access was attempted on a non-object value.
// ErrInvalidVariableAccess indicates an invalid variable access pattern.
var (
	ErrUndefinedVariable     = errors.New("undefined variable")
	ErrUndefinedProperty     = errors.New("undefined property")
	ErrNonStructProperty     = errors.New("cannot access property of non-struct value")
	ErrCannotAccessProperty  = errors.New("cannot access property")
	ErrNonObjectProperty     = errors.New("cannot access property of non-object")
	ErrInvalidVariableAccess = errors.New("invalid variable access")
)

// ErrCannotAddTypes indicates addition is not supported for the given types.
// ErrCannotSubtractTypes indicates subtraction is not supported for the given types.
// ErrCannotMultiplyTypes indicates multiplication is not supported for the given types.
// ErrCannotDivideTypes indicates division is not supported for the given types.
// ErrCannotModuloTypes indicates modulo is not supported for the given types.
// ErrDivisionByZero indicates a division by zero was attempted.
// ErrModuloByZero indicates a modulo by zero was attempted.
// ErrCannotConvertToBool indicates a value cannot be converted to boolean.
// ErrCannotNegate indicates a value cannot be negated.
// ErrCannotApplyUnaryPlus indicates unary plus cannot be applied to the value.
// ErrCannotCompareTypes indicates comparison is not supported for the given types.
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

// ErrInvalidIndexType indicates an invalid index type was used.
// ErrInvalidArrayIndex indicates an invalid array index was used.
// ErrIndexOutOfRange indicates an index is out of range.
// ErrCannotIndexNil indicates an indexing operation was attempted on nil.
// ErrTypeNotIndexable indicates the type does not support indexing.
// ErrCannotGetKeyFromNil indicates a key lookup was attempted on nil.
// ErrTypeNotMap indicates the type is not a map.
// ErrCannotGetFieldFromNil indicates a field access was attempted on nil.
// ErrStructHasNoField indicates the struct does not have the requested field.
// ErrTypeHasNoField indicates the type does not have the requested field.
// ErrUnsupportedArrayType indicates an unsupported array type was encountered.
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

// ErrUnsupportedCollectionType indicates the collection type is not supported in a for loop.
// ErrTypeNotIterable indicates the type does not support iteration.
// ErrTypeHasNoLength indicates the type does not support the length operation.
var (
	ErrUnsupportedCollectionType = errors.New("unsupported collection type for for loop")
	ErrTypeNotIterable           = errors.New("type is not iterable")
	ErrTypeHasNoLength           = errors.New("type has no length")
)

// ErrCannotConvertNilToInt indicates nil cannot be converted to int.
// ErrCannotConvertToInt indicates the value cannot be converted to int.
// ErrCannotConvertNilToFloat indicates nil cannot be converted to float.
// ErrCannotConvertToFloat indicates the value cannot be converted to float.
var (
	ErrCannotConvertNilToInt   = errors.New("cannot convert nil to int")
	ErrCannotConvertToInt      = errors.New("cannot convert value to int")
	ErrCannotConvertNilToFloat = errors.New("cannot convert nil to float")
	ErrCannotConvertToFloat    = errors.New("cannot convert value to float")
)

// ErrBreakOutsideLoop indicates a break statement was used outside of a loop.
// ErrContinueOutsideLoop indicates a continue statement was used outside of a loop.
var (
	ErrBreakOutsideLoop    = errors.New("break statement outside of loop")
	ErrContinueOutsideLoop = errors.New("continue statement outside of loop")
)

// ErrTagAlreadyRegistered indicates a tag with the same name is already registered.
//
// ErrMultipleElseStatements indicates multiple else clauses were found in an if block.
// ErrUnexpectedTokensAfterCondition indicates unexpected tokens after a condition expression.
// ErrElseNoArgs indicates the else tag received unexpected arguments.
// ErrEndifNoArgs indicates the endif tag received unexpected arguments.
// ErrElifAfterElse indicates an elif clause appeared after an else clause.
//
// ErrExpectedVariable indicates a variable name was expected but not found.
// ErrExpectedSecondVariable indicates a second variable name was expected after a comma.
// ErrExpectedInKeyword indicates the "in" keyword was expected but not found.
// ErrUnexpectedTokensAfterCollection indicates unexpected tokens after a collection expression.
// ErrEndforNoArgs indicates the endfor tag received unexpected arguments.
//
// ErrBreakNoArgs indicates the break tag received unexpected arguments.
// ErrContinueNoArgs indicates the continue tag received unexpected arguments.
var (
	ErrTagAlreadyRegistered = errors.New("tag already registered")

	ErrMultipleElseStatements         = errors.New("multiple 'else' statements found in if block, use 'elif' for additional conditions")
	ErrUnexpectedTokensAfterCondition = errors.New("unexpected tokens after condition")
	ErrElseNoArgs                     = errors.New("else does not take arguments")
	ErrEndifNoArgs                    = errors.New("endif does not take arguments")
	ErrElifAfterElse                  = errors.New("elif cannot appear after else")

	ErrExpectedVariable                = errors.New("expected variable name")
	ErrExpectedSecondVariable          = errors.New("expected second variable name after comma")
	ErrExpectedInKeyword               = errors.New("expected 'in' keyword")
	ErrUnexpectedTokensAfterCollection = errors.New("unexpected tokens after collection")
	ErrEndforNoArgs                    = errors.New("endfor does not take arguments")

	ErrBreakNoArgs    = errors.New("break does not take arguments")
	ErrContinueNoArgs = errors.New("continue does not take arguments")
)
