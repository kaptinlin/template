package template

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// noopFilter is a no-op filter used in registry tests.
func noopFilter(value any, _ ...any) (any, error) {
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

func TestRegistryReplaceOverwrites(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register("answer", func(_ any, _ ...any) (any, error) {
		return "old", nil
	})
	r.Replace("answer", func(_ any, _ ...any) (any, error) {
		return "new", nil
	})

	fn, ok := r.Filter("answer")
	if !ok {
		t.Fatal("Filter(\"answer\") = _, false, want true")
	}
	got, err := fn(nil)
	if err != nil {
		t.Fatalf("fn() err = %v", err)
	}
	if got != "new" {
		t.Fatalf("fn() = %v, want %q", got, "new")
	}
}

func TestRegistryMustRegisterNilPanics(t *testing.T) {
	t.Parallel()

	r := NewRegistry()

	defer func() {
		if recover() == nil {
			t.Error("MustRegister(nil) did not panic, want panic")
		}
	}()

	r.MustRegister("nilfilter", nil)
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

func TestRegistryCloneCopiesDirectFiltersAndParent(t *testing.T) {
	t.Parallel()

	parent := NewRegistry()
	parent.Register("parent", noopFilter)

	r := NewRegistry()
	r.parent = parent
	r.Register("local", noopFilter)

	clone := r.Clone()
	r.Unregister("local")

	if _, ok := clone.Filter("local"); !ok {
		t.Error("clone.Filter(\"local\") = _, false, want true")
	}
	if _, ok := clone.Filter("parent"); !ok {
		t.Error("clone.Filter(\"parent\") = _, false, want true")
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

// --- Built-in registry tests ---

func TestBuiltinRegistryRegisterAndFilter(t *testing.T) {
	r := NewRegistry()
	r.Register("testfilter", noopFilter)

	fn, ok := r.Filter("testfilter")
	if !ok {
		t.Error("Filter(\"testfilter\") = _, false, want true")
	}
	if fn == nil {
		t.Error("Filter(\"testfilter\") returned nil, want non-nil")
	}

	_, ok = r.Filter("nonexistent")
	if ok {
		t.Error("Filter(\"nonexistent\") = _, true, want false")
	}
}

func TestBuiltinRegistryRegisterNilPanics(t *testing.T) {
	r := NewRegistry()
	defer func() {
		if recover() == nil {
			t.Error("Register(nil) did not panic, want panic")
		}
	}()

	r.Register("nilfilter", nil)
}

func TestBuiltinRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register("charlie", noopFilter)
	r.Register("alpha", noopFilter)
	r.Register("bravo", noopFilter)

	got := r.List()
	want := []string{"alpha", "bravo", "charlie"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("List() mismatch (-want +got):\n%s", diff)
	}
}

func TestBuiltinRegistryHas(t *testing.T) {
	r := NewRegistry()
	r.Register("myfilter", noopFilter)

	if got := r.Has("myfilter"); !got {
		t.Errorf("Has(\"myfilter\") = %v, want true", got)
	}
	if got := r.Has("nonexistent"); got {
		t.Errorf("Has(\"nonexistent\") = %v, want false", got)
	}
}

func TestBuiltinRegistryUnregister(t *testing.T) {
	r := NewRegistry()
	r.Register("removeme", noopFilter)
	r.Register("keepme", noopFilter)
	r.Unregister("removeme")

	if got := r.Has("removeme"); got {
		t.Errorf("Has(\"removeme\") = %v, want false", got)
	}
	if got := r.Has("keepme"); !got {
		t.Errorf("Has(\"keepme\") = %v, want true", got)
	}
}

func TestBuiltinFiltersRegistered(t *testing.T) {
	t.Parallel()

	builtins := []string{
		// String filters
		"default", "strip", "lstrip", "rstrip", "split",
		"replace", "replace_first", "replace_last",
		"remove", "remove_first", "remove_last",
		"append", "prepend", "length",
		"upcase", "downcase", "capitalize",
		"escape", "escape_once", "h",
		"strip_html", "strip_newlines",
		"url_encode", "url_decode",
		"base64_encode", "base64_decode",
		"truncate", "truncatewords", "truncate_words",
		"slice",
		"trim", "trim_left", "trim_right",
		"upper", "lower",
		"titleize", "camelize", "pascalize", "dasherize",
		"slugify", "pluralize", "ordinalize",
		// Array filters
		"uniq", "join", "first", "last",
		"reverse", "size",
		"sort", "sort_natural", "compact", "concat",
		"where", "reject", "find", "find_index", "has",
		"map", "sum",
		"unique",
		"random", "shuffle", "max", "min", "average",
		// Math filters
		"abs", "at_least", "at_most", "round", "floor", "ceil",
		"plus", "minus", "times", "divided_by", "modulo",
		"divide",
		// Date filters
		"date", "month", "month_full", "year", "day",
		"week", "weekday", "time_ago", "timeago",
		// Number filters
		"number", "bytes",
		// Format filters
		"json",
		// Map filters
		"extract",
	}

	for _, name := range builtins {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, ok := defaultRegistry.Filter(name); !ok {
				t.Errorf("Filter(%q) = _, false, want true", name)
			}
		})
	}
}

func TestFilterRegistryConcurrentAccess(_ *testing.T) {
	// Concurrent register and query to verify thread safety.
	const goroutines = 10
	done := make(chan struct{})
	r := NewRegistry()

	for i := range goroutines {
		go func(id int) {
			defer func() { done <- struct{}{} }()
			name := fmt.Sprintf("concurrent_filter_%d", id)
			r.Register(name, func(value any, _ ...any) (any, error) {
				return value, nil
			})
			_, _ = r.Filter(name)
			_ = r.List()
			_ = r.Has(name)
		}(i)
	}

	for range goroutines {
		<-done
	}
}
