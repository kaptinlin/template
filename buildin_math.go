package template

import (
	"fmt"

	"github.com/kaptinlin/filter"
)

func init() {
	// Register all math filters
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
func absFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Abs(value)
}

// atLeastFilter ensures the number is at least as large as the minimum value provided.
func atLeastFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: atLeast filter requires one argument", ErrInsufficientArgs)
	}
	min := args[0]
	return filter.AtLeast(value, min)
}

// atMostFilter ensures the number is no larger than the maximum value provided.
func atMostFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: atMost filter requires one argument", ErrInsufficientArgs)
	}
	max := args[0]
	return filter.AtMost(value, max)
}

// roundFilter rounds the input to the specified number of decimal places.
func roundFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: round filter requires one argument for precision", ErrInsufficientArgs)
	}
	precision := args[0]
	return filter.Round(value, precision)
}

// floorFilter rounds the input down to the nearest whole number.
func floorFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Floor(value)
}

// ceilFilter rounds the input up to the nearest whole number.
func ceilFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Ceil(value)
}

// plusFilter adds two numbers.
func plusFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: plus filter requires one argument", ErrInsufficientArgs)
	}
	addend := args[0]
	return filter.Plus(value, addend)
}

// minusFilter subtracts the second value from the first.
func minusFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: minus filter requires one argument", ErrInsufficientArgs)
	}
	subtrahend := args[0]
	return filter.Minus(value, subtrahend)
}

// timesFilter multiplies the first value by the second.
func timesFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: times filter requires one argument", ErrInsufficientArgs)
	}
	multiplier := args[0]
	return filter.Times(value, multiplier)
}

// divideFilter divides the first value by the second.
func divideFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: divide filter requires one argument", ErrInsufficientArgs)
	}
	divisor := args[0]
	return filter.Divide(value, divisor)
}

// moduloFilter returns the remainder of the division of the first value by the second.
func moduloFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: modulo filter requires one argument", ErrInsufficientArgs)
	}
	modulus := args[0]
	return filter.Modulo(value, modulus)
}
