package template

import (
	"fmt"

	"github.com/kaptinlin/filter"
)

// registerMathFilters registers all math-related filters.
func registerMathFilters() {
	RegisterFilter("abs", absFilter)
	RegisterFilter("atLeast", atLeastFilter)
	RegisterFilter("atMost", atMostFilter)
	RegisterFilter("round", roundFilter)
	RegisterFilter("floor", floorFilter)
	RegisterFilter("ceil", ceilFilter)
	RegisterFilter("plus", plusFilter)
	RegisterFilter("minus", minusFilter)
	RegisterFilter("times", timesFilter)
	RegisterFilter("divide", divideFilter)
	RegisterFilter("modulo", moduloFilter)
}

// absFilter calculates the absolute value of a number.
func absFilter(value any, _ ...string) (any, error) {
	return filter.Abs(value)
}

// atLeastFilter ensures the number is at least as large as the minimum value provided.
func atLeastFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: atLeast filter requires one argument", ErrInsufficientArgs)
	}
	return filter.AtLeast(value, args[0])
}

// atMostFilter ensures the number is no larger than the maximum value provided.
func atMostFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: atMost filter requires one argument", ErrInsufficientArgs)
	}
	return filter.AtMost(value, args[0])
}

// roundFilter rounds the input to the specified number of decimal places.
func roundFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: round filter requires one argument for precision", ErrInsufficientArgs)
	}
	return filter.Round(value, args[0])
}

// floorFilter rounds the input down to the nearest whole number.
func floorFilter(value any, _ ...string) (any, error) {
	return filter.Floor(value)
}

// ceilFilter rounds the input up to the nearest whole number.
func ceilFilter(value any, _ ...string) (any, error) {
	return filter.Ceil(value)
}

// plusFilter adds two numbers.
func plusFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: plus filter requires one argument", ErrInsufficientArgs)
	}
	return filter.Plus(value, args[0])
}

// minusFilter subtracts the second value from the first.
func minusFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: minus filter requires one argument", ErrInsufficientArgs)
	}
	return filter.Minus(value, args[0])
}

// timesFilter multiplies the first value by the second.
func timesFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: times filter requires one argument", ErrInsufficientArgs)
	}
	return filter.Times(value, args[0])
}

// divideFilter divides the first value by the second.
func divideFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: divide filter requires one argument", ErrInsufficientArgs)
	}
	return filter.Divide(value, args[0])
}

// moduloFilter returns the remainder of the division of the first value by the second.
func moduloFilter(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: modulo filter requires one argument", ErrInsufficientArgs)
	}
	return filter.Modulo(value, args[0])
}
