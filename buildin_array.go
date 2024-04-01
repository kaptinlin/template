package template

import (
	"fmt"

	"github.com/kaptinlin/filter"
)

func init() {
	// Register all array filters with their corresponding functions
	RegisterFilter("unique", uniqueFilter)
	RegisterFilter("join", joinFilter)
	RegisterFilter("first", firstFilter)
	RegisterFilter("last", lastFilter)
	RegisterFilter("random", randomFilter)
	RegisterFilter("reverse", reverseFilter)
	RegisterFilter("shuffle", shuffleFilter)
	RegisterFilter("size", sizeFilter)
	RegisterFilter("max", maxFilter)
	RegisterFilter("min", minFilter)
	RegisterFilter("sum", sumFilter)
	RegisterFilter("average", averageFilter)
	RegisterFilter("map", mapFilter)
}

// uniqueFilter removes duplicate elements from a slice.
func uniqueFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Unique(value)
}

// joinFilter joins the elements of a slice into a single string with a specified separator.
func joinFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: join filter requires a separator argument", ErrInsufficientArgs)
	}
	separator := args[0]
	return filter.Join(value, separator)
}

// firstFilter returns the first element of a slice.
func firstFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.First(value)
}

// lastFilter returns the last element of a slice.
func lastFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Last(value)
}

// randomFilter selects a random element from a slice.
func randomFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Random(value)
}

// reverseFilter reverses the order of elements in a slice.
func reverseFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Reverse(value)
}

// shuffleFilter randomly rearranges the elements within the slice.
func shuffleFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Shuffle(value)
}

// sizeFilter returns the size (length) of a slice.
func sizeFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Size(value)
}

// maxFilter finds and returns the maximum value from a slice of numbers.
func maxFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Max(value)
}

// minFilter finds and returns the minimum value from a slice of numbers.
func minFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Min(value)
}

// sumFilter calculates the sum of all elements in a numerical slice.
func sumFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Sum(value)
}

// averageFilter computes the average value of a numerical slice.
func averageFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Average(value)
}

// mapFilter extracts a slice of values for a specified key from each map in the input slice.
func mapFilter(value interface{}, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: map filter requires a key argument", ErrInsufficientArgs)
	}
	key := args[0]
	return filter.Map(value, key)
}
