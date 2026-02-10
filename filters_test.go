package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// dummyFilter is a no-op filter used in registry tests.
func dummyFilter(value interface{}, _ ...string) (interface{}, error) {
	return value, nil
}

// --- Registry type tests ---

func TestRegistryRegisterAndFilter(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(r *Registry)
		registerName string
		registerFunc FilterFunc
		queryName    string
		expectFound  bool
	}{
		{
			name:         "register and retrieve new filter",
			setup:        func(_ *Registry) {},
			registerName: "testfilter",
			registerFunc: dummyFilter,
			queryName:    "testfilter",
			expectFound:  true,
		},
		{
			name: "overwrite existing filter",
			setup: func(r *Registry) {
				r.Register("existing", dummyFilter)
			},
			registerName: "existing",
			registerFunc: dummyFilter,
			queryName:    "existing",
			expectFound:  true,
		},
		{
			name:         "query non-existing filter returns nil",
			setup:        func(_ *Registry) {},
			registerName: "registered",
			registerFunc: dummyFilter,
			queryName:    "nonexistent",
			expectFound:  false,
		},
		{
			name: "register multiple filters independently",
			setup: func(r *Registry) {
				r.Register("alpha", dummyFilter)
			},
			registerName: "bravo",
			registerFunc: dummyFilter,
			queryName:    "alpha",
			expectFound:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.setup(r)
			r.Register(tt.registerName, tt.registerFunc)

			fn, ok := r.Filter(tt.queryName)
			assert.Equal(t, tt.expectFound, ok)
			if tt.expectFound {
				assert.NotNil(t, fn)
			}
		})
	}
}

func TestRegistryRegisterNilPanics(t *testing.T) {
	r := NewRegistry()
	assert.Panics(t, func() {
		r.Register("nilfilter", nil)
	})
}

func TestRegistryList(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(r *Registry)
		expected []string
	}{
		{
			name:     "empty registry returns empty list",
			setup:    func(_ *Registry) {},
			expected: []string{},
		},
		{
			name: "single filter",
			setup: func(r *Registry) {
				r.Register("alpha", dummyFilter)
			},
			expected: []string{"alpha"},
		},
		{
			name: "multiple filters returned in sorted order",
			setup: func(r *Registry) {
				r.Register("charlie", dummyFilter)
				r.Register("alpha", dummyFilter)
				r.Register("bravo", dummyFilter)
			},
			expected: []string{"alpha", "bravo", "charlie"},
		},
		{
			name: "numeric-prefixed names sort lexicographically",
			setup: func(r *Registry) {
				r.Register("2nd", dummyFilter)
				r.Register("10th", dummyFilter)
				r.Register("1st", dummyFilter)
			},
			expected: []string{"10th", "1st", "2nd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.setup(r)
			assert.Equal(t, tt.expected, r.List())
		})
	}
}

func TestRegistryHas(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(r *Registry)
		filter   string
		expected bool
	}{
		{
			name: "existing filter returns true",
			setup: func(r *Registry) {
				r.Register("myfilter", dummyFilter)
			},
			filter:   "myfilter",
			expected: true,
		},
		{
			name:     "non-existing filter returns false",
			setup:    func(_ *Registry) {},
			filter:   "nonexistent",
			expected: false,
		},
		{
			name:     "empty name returns false",
			setup:    func(_ *Registry) {},
			filter:   "",
			expected: false,
		},
		{
			name: "case sensitive lookup",
			setup: func(r *Registry) {
				r.Register("Upper", dummyFilter)
			},
			filter:   "upper",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.setup(r)
			assert.Equal(t, tt.expected, r.Has(tt.filter))
		})
	}
}

func TestRegistryUnregister(t *testing.T) {
	tests := []struct {
		name               string
		setup              func(r *Registry)
		unregisterName     string
		expectAfterRemoval map[string]bool
	}{
		{
			name: "unregister existing filter removes it",
			setup: func(r *Registry) {
				r.Register("removeme", dummyFilter)
				r.Register("keepme", dummyFilter)
			},
			unregisterName: "removeme",
			expectAfterRemoval: map[string]bool{
				"removeme": false,
				"keepme":   true,
			},
		},
		{
			name: "unregister non-existing filter is no-op",
			setup: func(r *Registry) {
				r.Register("keep", dummyFilter)
			},
			unregisterName: "nonexistent",
			expectAfterRemoval: map[string]bool{
				"keep":        true,
				"nonexistent": false,
			},
		},
		{
			name:               "unregister from empty registry is no-op",
			setup:              func(_ *Registry) {},
			unregisterName:     "anything",
			expectAfterRemoval: map[string]bool{"anything": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.setup(r)
			r.Unregister(tt.unregisterName)

			for name, expectExists := range tt.expectAfterRemoval {
				_, exists := r.Filter(name)
				assert.Equal(t, expectExists, exists,
					"filter %q existence mismatch", name)
			}
		})
	}
}

// --- Package-level convenience function tests ---

func TestRegisterFilterAndGetFilter(t *testing.T) {
	saved := defaultRegistry
	defer func() { defaultRegistry = saved }()
	defaultRegistry = NewRegistry()

	RegisterFilter("testfilter", dummyFilter)

	fn, ok := GetFilter("testfilter")
	assert.True(t, ok)
	assert.NotNil(t, fn)

	_, ok = GetFilter("nonexistent")
	assert.False(t, ok)
}

func TestRegisterFilterNilPanics(t *testing.T) {
	saved := defaultRegistry
	defer func() { defaultRegistry = saved }()
	defaultRegistry = NewRegistry()

	assert.Panics(t, func() {
		RegisterFilter("nilfilter", nil)
	})
}

func TestListFilters(t *testing.T) {
	saved := defaultRegistry
	defer func() { defaultRegistry = saved }()
	defaultRegistry = NewRegistry()

	RegisterFilter("charlie", dummyFilter)
	RegisterFilter("alpha", dummyFilter)
	RegisterFilter("bravo", dummyFilter)

	assert.Equal(t, []string{"alpha", "bravo", "charlie"}, ListFilters())
}

func TestHasFilter(t *testing.T) {
	saved := defaultRegistry
	defer func() { defaultRegistry = saved }()
	defaultRegistry = NewRegistry()

	RegisterFilter("myfilter", dummyFilter)

	assert.True(t, HasFilter("myfilter"))
	assert.False(t, HasFilter("nonexistent"))
}

func TestUnregisterFilter(t *testing.T) {
	saved := defaultRegistry
	defer func() { defaultRegistry = saved }()
	defaultRegistry = NewRegistry()

	RegisterFilter("removeme", dummyFilter)
	RegisterFilter("keepme", dummyFilter)
	UnregisterFilter("removeme")

	assert.False(t, HasFilter("removeme"))
	assert.True(t, HasFilter("keepme"))
}

func TestBuiltinFiltersRegistered(t *testing.T) {
	expectedFilters := []string{
		"upper", "lower", "capitalize", "length",
		"default", "trim", "join", "first", "last",
		"reverse", "abs", "round", "floor", "ceil",
		"plus", "minus", "times", "divide", "modulo",
		"date", "json", "number", "bytes",
		"unique", "shuffle", "size",
		"max", "min", "sum", "average", "extract",
	}

	for _, name := range expectedFilters {
		t.Run(name, func(t *testing.T) {
			_, ok := GetFilter(name)
			assert.True(t, ok, "built-in filter %q not registered", name)
		})
	}
}
