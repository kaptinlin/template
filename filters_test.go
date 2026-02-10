package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// dummyFilter is a no-op filter used in registry tests.
func dummyFilter(value interface{}, args ...string) (interface{}, error) {
	return value, nil
}

func TestRegisterFilterAndGetFilter(t *testing.T) {
	originalRegistry := filterRegistry
	defer func() { filterRegistry = originalRegistry }()

	tests := []struct {
		name         string
		setup        func()
		registerName string
		registerFunc FilterFunc
		queryName    string
		expectFound  bool
	}{
		{
			name: "register and retrieve new filter",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
			},
			registerName: "testfilter",
			registerFunc: dummyFilter,
			queryName:    "testfilter",
			expectFound:  true,
		},
		{
			name: "overwrite existing filter",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
				filterRegistry["existing"] = dummyFilter
			},
			registerName: "existing",
			registerFunc: dummyFilter,
			queryName:    "existing",
			expectFound:  true,
		},
		{
			name: "query non-existing filter returns nil",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
			},
			registerName: "registered",
			registerFunc: dummyFilter,
			queryName:    "nonexistent",
			expectFound:  false,
		},
		{
			name: "register multiple filters independently",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
				filterRegistry["alpha"] = dummyFilter
			},
			registerName: "bravo",
			registerFunc: dummyFilter,
			queryName:    "alpha",
			expectFound:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			RegisterFilter(tt.registerName, tt.registerFunc)

			fn, ok := GetFilter(tt.queryName)
			assert.Equal(t, tt.expectFound, ok)
			if tt.expectFound {
				assert.NotNil(t, fn)
			}
		})
	}
}

func TestRegisterFilterNilPanics(t *testing.T) {
	originalRegistry := filterRegistry
	defer func() { filterRegistry = originalRegistry }()
	filterRegistry = make(map[string]FilterFunc)

	assert.Panics(t, func() {
		RegisterFilter("nilfilter", nil)
	})
}

func TestListFilters(t *testing.T) {
	originalRegistry := filterRegistry
	defer func() { filterRegistry = originalRegistry }()

	tests := []struct {
		name     string
		setup    func()
		expected []string
	}{
		{
			name: "empty registry returns empty list",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
			},
			expected: []string{},
		},
		{
			name: "single filter",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
				filterRegistry["alpha"] = dummyFilter
			},
			expected: []string{"alpha"},
		},
		{
			name: "multiple filters returned in sorted order",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
				filterRegistry["charlie"] = dummyFilter
				filterRegistry["alpha"] = dummyFilter
				filterRegistry["bravo"] = dummyFilter
			},
			expected: []string{"alpha", "bravo", "charlie"},
		},
		{
			name: "numeric-prefixed names sort lexicographically",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
				filterRegistry["2nd"] = dummyFilter
				filterRegistry["10th"] = dummyFilter
				filterRegistry["1st"] = dummyFilter
			},
			expected: []string{"10th", "1st", "2nd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := ListFilters()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasFilter(t *testing.T) {
	originalRegistry := filterRegistry
	defer func() { filterRegistry = originalRegistry }()

	tests := []struct {
		name     string
		setup    func()
		filter   string
		expected bool
	}{
		{
			name: "existing filter returns true",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
				filterRegistry["myfilter"] = dummyFilter
			},
			filter:   "myfilter",
			expected: true,
		},
		{
			name: "non-existing filter returns false",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
			},
			filter:   "nonexistent",
			expected: false,
		},
		{
			name: "empty name returns false",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
			},
			filter:   "",
			expected: false,
		},
		{
			name: "case sensitive lookup",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
				filterRegistry["Upper"] = dummyFilter
			},
			filter:   "upper",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := HasFilter(tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnregisterFilter(t *testing.T) {
	originalRegistry := filterRegistry
	defer func() { filterRegistry = originalRegistry }()

	tests := []struct {
		name               string
		setup              func()
		unregisterName     string
		expectAfterRemoval map[string]bool
	}{
		{
			name: "unregister existing filter removes it",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
				filterRegistry["removeme"] = dummyFilter
				filterRegistry["keepme"] = dummyFilter
			},
			unregisterName: "removeme",
			expectAfterRemoval: map[string]bool{
				"removeme": false,
				"keepme":   true,
			},
		},
		{
			name: "unregister non-existing filter is no-op",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
				filterRegistry["keep"] = dummyFilter
			},
			unregisterName: "nonexistent",
			expectAfterRemoval: map[string]bool{
				"keep":        true,
				"nonexistent": false,
			},
		},
		{
			name: "unregister from empty registry is no-op",
			setup: func() {
				filterRegistry = make(map[string]FilterFunc)
			},
			unregisterName:     "anything",
			expectAfterRemoval: map[string]bool{"anything": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			UnregisterFilter(tt.unregisterName)

			for name, expectExists := range tt.expectAfterRemoval {
				_, exists := GetFilter(name)
				assert.Equal(t, expectExists, exists, "filter %q existence mismatch", name)
			}
		})
	}
}

func TestBuiltinFiltersRegistered(t *testing.T) {
	expectedFilters := []struct {
		name   string
		exists bool
	}{
		{name: "upper", exists: true},
		{name: "lower", exists: true},
		{name: "capitalize", exists: true},
		{name: "length", exists: true},
		{name: "default", exists: true},
		{name: "trim", exists: true},
		{name: "join", exists: true},
		{name: "first", exists: true},
		{name: "last", exists: true},
		{name: "reverse", exists: true},
		{name: "abs", exists: true},
		{name: "round", exists: true},
		{name: "floor", exists: true},
		{name: "ceil", exists: true},
		{name: "plus", exists: true},
		{name: "minus", exists: true},
		{name: "times", exists: true},
		{name: "divide", exists: true},
		{name: "modulo", exists: true},
		{name: "date", exists: true},
		{name: "json", exists: true},
		{name: "number", exists: true},
		{name: "bytes", exists: true},
		{name: "unique", exists: true},
		{name: "shuffle", exists: true},
		{name: "size", exists: true},
		{name: "max", exists: true},
		{name: "min", exists: true},
		{name: "sum", exists: true},
		{name: "average", exists: true},
		{name: "extract", exists: true},
	}

	for _, tt := range expectedFilters {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := GetFilter(tt.name)
			assert.Equal(t, tt.exists, ok)
		})
	}
}
