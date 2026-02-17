package template

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// noopFilter is a no-op filter used in registry tests.
func noopFilter(value any, _ ...string) (any, error) {
	return value, nil
}

// --- Registry type tests ---

func TestRegistryRegisterAndFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(r *Registry)
		register  string
		query     string
		wantFound bool
	}{
		{
			name:      "register and retrieve",
			setup:     func(_ *Registry) {},
			register:  "testfilter",
			query:     "testfilter",
			wantFound: true,
		},
		{
			name: "overwrite existing",
			setup: func(r *Registry) {
				r.Register("existing", noopFilter)
			},
			register:  "existing",
			query:     "existing",
			wantFound: true,
		},
		{
			name:      "query missing filter",
			setup:     func(_ *Registry) {},
			register:  "registered",
			query:     "nonexistent",
			wantFound: false,
		},
		{
			name: "multiple filters independent",
			setup: func(r *Registry) {
				r.Register("alpha", noopFilter)
			},
			register:  "bravo",
			query:     "alpha",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NewRegistry()
			tt.setup(r)
			r.Register(tt.register, noopFilter)

			fn, got := r.Filter(tt.query)
			if got != tt.wantFound {
				t.Errorf("Filter(%q) found = %v, want %v", tt.query, got, tt.wantFound)
			}
			if tt.wantFound && fn == nil {
				t.Errorf("Filter(%q) returned nil function, want non-nil", tt.query)
			}
		})
	}
}

func TestRegistryRegisterNilPanics(t *testing.T) {
	t.Parallel()

	r := NewRegistry()

	defer func() {
		if recover() == nil {
			t.Error("Register(nil) did not panic, want panic")
		}
	}()

	r.Register("nilfilter", nil)
}

func TestRegistryList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(r *Registry)
		want  []string
	}{
		{
			name:  "empty registry",
			setup: func(_ *Registry) {},
			want:  nil,
		},
		{
			name: "single filter",
			setup: func(r *Registry) {
				r.Register("alpha", noopFilter)
			},
			want: []string{"alpha"},
		},
		{
			name: "sorted order",
			setup: func(r *Registry) {
				r.Register("charlie", noopFilter)
				r.Register("alpha", noopFilter)
				r.Register("bravo", noopFilter)
			},
			want: []string{"alpha", "bravo", "charlie"},
		},
		{
			name: "lexicographic numeric sort",
			setup: func(r *Registry) {
				r.Register("2nd", noopFilter)
				r.Register("10th", noopFilter)
				r.Register("1st", noopFilter)
			},
			want: []string{"10th", "1st", "2nd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NewRegistry()
			tt.setup(r)

			got := r.List()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("List() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRegistryHas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setup  func(r *Registry)
		filter string
		want   bool
	}{
		{
			name: "existing filter",
			setup: func(r *Registry) {
				r.Register("myfilter", noopFilter)
			},
			filter: "myfilter",
			want:   true,
		},
		{
			name:   "missing filter",
			setup:  func(_ *Registry) {},
			filter: "nonexistent",
			want:   false,
		},
		{
			name:   "empty name",
			setup:  func(_ *Registry) {},
			filter: "",
			want:   false,
		},
		{
			name: "case sensitive",
			setup: func(r *Registry) {
				r.Register("Upper", noopFilter)
			},
			filter: "upper",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NewRegistry()
			tt.setup(r)

			if got := r.Has(tt.filter); got != tt.want {
				t.Errorf("Has(%q) = %v, want %v", tt.filter, got, tt.want)
			}
		})
	}
}

func TestRegistryUnregister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(r *Registry)
		unregister string
		wantAfter  map[string]bool
	}{
		{
			name: "removes existing filter",
			setup: func(r *Registry) {
				r.Register("removeme", noopFilter)
				r.Register("keepme", noopFilter)
			},
			unregister: "removeme",
			wantAfter: map[string]bool{
				"removeme": false,
				"keepme":   true,
			},
		},
		{
			name: "no-op for missing filter",
			setup: func(r *Registry) {
				r.Register("keep", noopFilter)
			},
			unregister: "nonexistent",
			wantAfter: map[string]bool{
				"keep":        true,
				"nonexistent": false,
			},
		},
		{
			name:       "no-op on empty registry",
			setup:      func(_ *Registry) {},
			unregister: "anything",
			wantAfter:  map[string]bool{"anything": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NewRegistry()
			tt.setup(r)
			r.Unregister(tt.unregister)

			for name, want := range tt.wantAfter {
				if _, got := r.Filter(name); got != want {
					t.Errorf("after Unregister(%q): Has(%q) = %v, want %v",
						tt.unregister, name, got, want)
				}
			}
		})
	}
}

// --- Package-level convenience function tests ---

// withFreshRegistry replaces the default registry for the duration of
// the test and restores it on cleanup.
func withFreshRegistry(t *testing.T) {
	t.Helper()
	saved := defaultRegistry
	defaultRegistry = NewRegistry()
	t.Cleanup(func() { defaultRegistry = saved })
}

func TestRegisterFilterAndFilter(t *testing.T) {
	withFreshRegistry(t)

	RegisterFilter("testfilter", noopFilter)

	fn, ok := Filter("testfilter")
	if !ok {
		t.Error("Filter(\"testfilter\") = _, false, want true")
	}
	if fn == nil {
		t.Error("Filter(\"testfilter\") returned nil, want non-nil")
	}

	_, ok = Filter("nonexistent")
	if ok {
		t.Error("Filter(\"nonexistent\") = _, true, want false")
	}
}

func TestRegisterFilterNilPanics(t *testing.T) {
	withFreshRegistry(t)

	defer func() {
		if recover() == nil {
			t.Error("RegisterFilter(nil) did not panic, want panic")
		}
	}()

	RegisterFilter("nilfilter", nil)
}

func TestListFilters(t *testing.T) {
	withFreshRegistry(t)

	RegisterFilter("charlie", noopFilter)
	RegisterFilter("alpha", noopFilter)
	RegisterFilter("bravo", noopFilter)

	got := ListFilters()
	want := []string{"alpha", "bravo", "charlie"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ListFilters() mismatch (-want +got):\n%s", diff)
	}
}

func TestHasFilter(t *testing.T) {
	withFreshRegistry(t)

	RegisterFilter("myfilter", noopFilter)

	if got := HasFilter("myfilter"); !got {
		t.Errorf("HasFilter(\"myfilter\") = %v, want true", got)
	}
	if got := HasFilter("nonexistent"); got {
		t.Errorf("HasFilter(\"nonexistent\") = %v, want false", got)
	}
}

func TestUnregisterFilter(t *testing.T) {
	withFreshRegistry(t)

	RegisterFilter("removeme", noopFilter)
	RegisterFilter("keepme", noopFilter)
	UnregisterFilter("removeme")

	if got := HasFilter("removeme"); got {
		t.Errorf("HasFilter(\"removeme\") = %v, want false", got)
	}
	if got := HasFilter("keepme"); !got {
		t.Errorf("HasFilter(\"keepme\") = %v, want true", got)
	}
}

func TestBuiltinFiltersRegistered(t *testing.T) {
	t.Parallel()

	builtins := []string{
		"upper", "lower", "capitalize", "length",
		"default", "trim", "join", "first", "last",
		"reverse", "abs", "round", "floor", "ceil",
		"plus", "minus", "times", "divide", "modulo",
		"date", "json", "number", "bytes",
		"unique", "shuffle", "size",
		"max", "min", "sum", "average", "extract",
	}

	for _, name := range builtins {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, ok := Filter(name); !ok {
				t.Errorf("Filter(%q) = _, false, want true", name)
			}
		})
	}
}

func TestFilterRegistryConcurrentAccess(_ *testing.T) {
	// Concurrent register and query to verify thread safety.
	const goroutines = 10
	done := make(chan struct{})

	for i := range goroutines {
		go func(id int) {
			defer func() { done <- struct{}{} }()
			name := fmt.Sprintf("concurrent_filter_%d", id)
			RegisterFilter(name, func(value any, _ ...string) (any, error) {
				return value, nil
			})
			_, _ = Filter(name)
			_ = ListFilters()
			_ = HasFilter(name)
		}(i)
	}

	for range goroutines {
		<-done
	}
}
