package template

import (
	"fmt"

	"github.com/kaptinlin/filter"
)

// registerArrayFilters registers all array-related filters.
func registerArrayFilters() {
	// Liquid-standard primary names
	RegisterFilter("uniq", uniqueFilter)
	RegisterFilter("join", joinFilter)
	RegisterFilter("first", firstFilter)
	RegisterFilter("last", lastFilter)
	RegisterFilter("reverse", reverseFilter)
	RegisterFilter("size", sizeFilter)
	RegisterFilter("sort", sortFilter)
	RegisterFilter("sort_natural", sortNaturalFilter)
	RegisterFilter("compact", compactFilter)
	RegisterFilter("concat", concatFilter)
	RegisterFilter("where", whereFilter)
	RegisterFilter("reject", rejectFilter)
	RegisterFilter("find", findFilter)
	RegisterFilter("find_index", findIndexFilter)
	RegisterFilter("has", hasFilter)
	RegisterFilter("map", mapFilter)
	RegisterFilter("sum", sumFilter)

	// Extension filters (non-Liquid)
	RegisterFilter("random", randomFilter)
	RegisterFilter("shuffle", shuffleFilter)
	RegisterFilter("max", maxFilter)
	RegisterFilter("min", minFilter)
	RegisterFilter("average", averageFilter)

	// Aliases
	RegisterFilter("unique", uniqueFilter)
}

// uniqueFilter removes duplicate elements from a slice.
// An optional property argument deduplicates objects by a specific key.
func uniqueFilter(value any, args ...any) (any, error) {
	if len(args) >= 1 {
		// Deduplicate by property: extract key values and keep first occurrence.
		key := toString(args[0])
		values, err := filter.Map(value, key)
		if err != nil {
			return filter.Unique(value)
		}
		slice, err := toSlice(value)
		if err != nil {
			return nil, err
		}
		seen := make(map[any]bool, len(slice))
		result := make([]any, 0, len(slice))
		for i, item := range slice {
			k := values[i]
			if !seen[k] {
				seen[k] = true
				result = append(result, item)
			}
		}
		return result, nil
	}
	return filter.Unique(value)
}

// joinFilter joins the elements of a slice into a single string.
// The separator defaults to " " (space) if not provided.
func joinFilter(value any, args ...any) (any, error) {
	sep := " "
	if len(args) >= 1 {
		sep = toString(args[0])
	}
	return filter.Join(value, sep)
}

// firstFilter returns the first element of a slice.
func firstFilter(value any, _ ...any) (any, error) {
	return filter.First(value)
}

// lastFilter returns the last element of a slice.
func lastFilter(value any, _ ...any) (any, error) {
	return filter.Last(value)
}

// randomFilter selects a random element from a slice.
func randomFilter(value any, _ ...any) (any, error) {
	return filter.Random(value)
}

// reverseFilter reverses the order of elements in a slice.
func reverseFilter(value any, _ ...any) (any, error) {
	return filter.Reverse(value)
}

// shuffleFilter randomly rearranges the elements within the slice.
func shuffleFilter(value any, _ ...any) (any, error) {
	return filter.Shuffle(value)
}

// sizeFilter returns the size (length) of a string, slice, array, or map.
func sizeFilter(value any, _ ...any) (any, error) {
	return lengthFilter(value)
}

// maxFilter finds and returns the maximum value from a slice of numbers.
func maxFilter(value any, _ ...any) (any, error) {
	return filter.Max(value)
}

// minFilter finds and returns the minimum value from a slice of numbers.
func minFilter(value any, _ ...any) (any, error) {
	return filter.Min(value)
}

// sumFilter calculates the sum of all elements in a numerical slice.
// An optional property argument sums a specific field from objects.
func sumFilter(value any, args ...any) (any, error) {
	if len(args) >= 1 {
		key := toString(args[0])
		values, err := filter.Map(value, key)
		if err != nil {
			return float64(0), err
		}
		var sum float64
		for _, v := range values {
			f, err := toFloat64(v)
			if err != nil {
				continue
			}
			sum += f
		}
		return sum, nil
	}
	return filter.Sum(value)
}

// averageFilter computes the average value of a numerical slice.
func averageFilter(value any, _ ...any) (any, error) {
	return filter.Average(value)
}

// mapFilter extracts a slice of values for a specified key from each map in the input slice.
func mapFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: map filter requires a key argument", ErrInsufficientArgs)
	}
	return filter.Map(value, toString(args[0]))
}

// sortFilter sorts the elements of a slice in ascending order.
// An optional key argument sorts by a specific property for slices of maps.
func sortFilter(value any, args ...any) (any, error) {
	if len(args) >= 1 {
		return filter.Sort(value, toString(args[0]))
	}
	return filter.Sort(value)
}

// sortNaturalFilter sorts the elements case-insensitively.
// An optional key argument sorts by a specific property for slices of maps.
func sortNaturalFilter(value any, args ...any) (any, error) {
	if len(args) >= 1 {
		return filter.SortNatural(value, toString(args[0]))
	}
	return filter.SortNatural(value)
}

// compactFilter removes nil values from a slice.
// An optional key argument removes items where the key's value is nil.
func compactFilter(value any, args ...any) (any, error) {
	if len(args) >= 1 {
		return filter.Compact(value, toString(args[0]))
	}
	return filter.Compact(value)
}

// concatFilter combines two arrays into one.
func concatFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: concat filter requires an array argument", ErrInsufficientArgs)
	}
	return filter.Concat(value, args[0])
}

// whereFilter selects items where a key matches a value.
// If no value is given, selects items where the key is truthy.
func whereFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: where filter requires a key argument", ErrInsufficientArgs)
	}
	if len(args) >= 2 {
		return filter.Where(value, toString(args[0]), args[1])
	}
	return filter.Where(value, toString(args[0]))
}

// rejectFilter selects items where a key does not match a value.
// If no value is given, rejects items where the key is truthy.
func rejectFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: reject filter requires a key argument", ErrInsufficientArgs)
	}
	if len(args) >= 2 {
		return filter.Reject(value, toString(args[0]), args[1])
	}
	return filter.Reject(value, toString(args[0]))
}

// findFilter returns the first item where a key matches a value.
func findFilter(value any, args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: find filter requires key and value arguments", ErrInsufficientArgs)
	}
	return filter.Find(value, toString(args[0]), args[1])
}

// findIndexFilter returns the index of the first item where a key matches a value.
func findIndexFilter(value any, args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: find_index filter requires key and value arguments", ErrInsufficientArgs)
	}
	return filter.FindIndex(value, toString(args[0]), args[1])
}

// hasFilter checks whether any item in a slice has a key matching a value.
// If no value is given, checks whether any item has a truthy key.
func hasFilter(value any, args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: has filter requires a key argument", ErrInsufficientArgs)
	}
	if len(args) >= 2 {
		return filter.Has(value, toString(args[0]), args[1])
	}
	return filter.Has(value, toString(args[0]))
}
