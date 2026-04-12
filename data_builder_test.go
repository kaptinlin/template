package template

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNewDataBuilder(t *testing.T) {
	got, err := NewDataBuilder().Build()
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}
	if want := (Data{}); !cmp.Equal(got, want) {
		t.Errorf("Build() = %v, want %v", got, want)
	}
}

func TestDataBuilderKeyValue(t *testing.T) {
	tests := []struct {
		name   string
		keys   []string
		values []any
		want   Data
	}{
		{
			name:   "single string key-value",
			keys:   []string{"name"},
			values: []any{"John"},
			want:   Data{"name": "John"},
		},
		{
			name:   "multiple key-values of different types",
			keys:   []string{"name", "age", "active"},
			values: []any{"John", 30, true},
			want:   Data{"name": "John", "age": 30, "active": true},
		},
		{
			name:   "nested key via dot notation",
			keys:   []string{"user.name"},
			values: []any{"John"},
			want: Data{
				"user": map[string]any{
					"name": "John",
				},
			},
		},
		{
			name:   "deeply nested key",
			keys:   []string{"a.b.c"},
			values: []any{"deep"},
			want: Data{
				"a": map[string]any{
					"b": map[string]any{
						"c": "deep",
					},
				},
			},
		},
		{
			name:   "overwrite same key keeps last value",
			keys:   []string{"key", "key"},
			values: []any{"first", "second"},
			want:   Data{"key": "second"},
		},
		{
			name:   "nil value stored correctly",
			keys:   []string{"empty"},
			values: []any{nil},
			want:   Data{"empty": nil},
		},
		{
			name:   "slice value stored correctly",
			keys:   []string{"items"},
			values: []any{[]int{1, 2, 3}},
			want:   Data{"items": []int{1, 2, 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewDataBuilder()
			for i, key := range tt.keys {
				b.KeyValue(key, tt.values[i])
			}
			got, err := b.Build()
			if err != nil {
				t.Fatalf("Build() unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Build() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDataBuilderStruct(t *testing.T) {
	type SimpleUser struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Active bool   `json:"active"`
	}

	type NestedProfile struct {
		Name    string `json:"name"`
		Profile struct {
			Bio     string `json:"bio"`
			Website string `json:"website"`
		} `json:"profile"`
	}

	type OmitemptyUser struct {
		Name  string `json:"name"`
		Email string `json:"email,omitempty"`
	}

	tests := []struct {
		name  string
		input any
		want  Data
	}{
		{
			name:  "simple struct with string and bool fields",
			input: SimpleUser{Name: "John", Email: "john@test.com", Active: true},
			want: Data{
				"name":   "John",
				"email":  "john@test.com",
				"active": true,
			},
		},
		{
			name:  "struct with zero values",
			input: SimpleUser{},
			want: Data{
				"name":   "",
				"email":  "",
				"active": false,
			},
		},
		{
			name: "nested struct flattened to maps",
			input: NestedProfile{
				Name: "Jane",
				Profile: struct {
					Bio     string `json:"bio"`
					Website string `json:"website"`
				}{Bio: "Engineer", Website: "https://example.com"},
			},
			want: Data{
				"name": "Jane",
				"profile": map[string]any{
					"bio":     "Engineer",
					"website": "https://example.com",
				},
			},
		},
		{
			name:  "omitempty matches json semantics",
			input: OmitemptyUser{Name: "Jane"},
			want: Data{
				"name": "Jane",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDataBuilder().Struct(tt.input).Build()
			if err != nil {
				t.Fatalf("Build() unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Build() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDataBuilderStructPreservesCustomJSONSemantics(t *testing.T) {
	type Event struct {
		At time.Time `json:"at"`
	}

	got, err := NewDataBuilder().Struct(Event{
		At: time.Date(2026, 4, 12, 10, 30, 0, 0, time.UTC),
	}).Build()
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}

	if _, ok := got["at"].(string); !ok {
		t.Fatalf("at = %T, want string", got["at"])
	}
}

func TestDataBuilderStructWithMarshalError(t *testing.T) {
	got, err := NewDataBuilder().Struct(make(chan int)).Build()
	if err == nil {
		t.Error("Build() = _, nil, want error for chan type")
	}
	if got == nil {
		t.Error("Build() returned nil context, want non-nil even on error")
	}
}

func TestDataBuilderChaining(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name  string
		build func() *DataBuilder
		want  Data
	}{
		{
			name: "KeyValue then Struct merges both",
			build: func() *DataBuilder {
				return NewDataBuilder().
					KeyValue("extra", "value").
					Struct(User{Name: "Alice"})
			},
			want: Data{
				"extra": "value",
				"name":  "Alice",
			},
		},
		{
			name: "Struct then KeyValue overrides struct field",
			build: func() *DataBuilder {
				return NewDataBuilder().
					Struct(User{Name: "Alice"}).
					KeyValue("name", "Bob")
			},
			want: Data{"name": "Bob"},
		},
		{
			name: "multiple KeyValue calls",
			build: func() *DataBuilder {
				return NewDataBuilder().
					KeyValue("a", 1).
					KeyValue("b", 2).
					KeyValue("c", 3)
			},
			want: Data{"a": 1, "b": 2, "c": 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.build().Build()
			if err != nil {
				t.Fatalf("Build() unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Build() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDataBuilderBuildErrorCollection(t *testing.T) {
	tests := []struct {
		name      string
		build     func() *DataBuilder
		wantError bool
	}{
		{
			name: "no errors returns nil error",
			build: func() *DataBuilder {
				return NewDataBuilder().KeyValue("key", "value")
			},
		},
		{
			name: "single Struct error collected",
			build: func() *DataBuilder {
				return NewDataBuilder().Struct(make(chan int))
			},
			wantError: true,
		},
		{
			name: "multiple Struct errors collected",
			build: func() *DataBuilder {
				return NewDataBuilder().
					Struct(make(chan int)).
					Struct(make(chan string))
			},
			wantError: true,
		},
		{
			name: "mixed valid and invalid still reports error",
			build: func() *DataBuilder {
				type Valid struct {
					X string `json:"x"`
				}
				return NewDataBuilder().
					KeyValue("key", "value").
					Struct(Valid{X: "ok"}).
					Struct(make(chan int))
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.build().Build()
			if tt.wantError && err == nil {
				t.Error("Build() error = nil, want error")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Build() unexpected error: %v", err)
			}
			if got == nil {
				t.Error("Build() returned nil context, want non-nil")
			}
		})
	}
}

func TestNewChildContext(t *testing.T) {
	tests := []struct {
		name              string
		data              map[string]any
		parentPrivate     map[string]any
		childPrivate      map[string]any
		wantChildGet      map[string]any
		wantParentPrivate map[string]any
	}{
		{
			name:              "child inherits render data",
			data:              map[string]any{"name": "Alice", "age": 30},
			parentPrivate:     map[string]any{},
			childPrivate:      map[string]any{},
			wantChildGet:      map[string]any{"name": "Alice", "age": 30},
			wantParentPrivate: map[string]any{},
		},
		{
			name:              "child inherits parent locals",
			data:              map[string]any{},
			parentPrivate:     map[string]any{"counter": 1, "flag": true},
			childPrivate:      map[string]any{},
			wantChildGet:      map[string]any{"counter": 1, "flag": true},
			wantParentPrivate: map[string]any{"counter": 1, "flag": true},
		},
		{
			name:              "child private modification does not affect parent",
			data:              map[string]any{},
			parentPrivate:     map[string]any{"x": 10},
			childPrivate:      map[string]any{"x": 99, "y": 20},
			wantChildGet:      map[string]any{"x": 99, "y": 20},
			wantParentPrivate: map[string]any{"x": 10},
		},
		{
			name:              "child private does not leak to parent",
			data:              map[string]any{},
			parentPrivate:     map[string]any{},
			childPrivate:      map[string]any{"new_var": "child_only"},
			wantChildGet:      map[string]any{"new_var": "child_only"},
			wantParentPrivate: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := NewRenderContext(tt.data)
			for k, v := range tt.parentPrivate {
				parent.Set(k, v)
			}
			child := NewChildContext(parent)
			for k, v := range tt.childPrivate {
				child.Set(k, v)
			}

			for key, want := range tt.wantChildGet {
				got, ok := child.Get(key)
				if !ok {
					t.Errorf("child.Get(%q) = _, false, want true", key)
				}
				if got != want {
					t.Errorf("child.Get(%q) = %v, want %v", key, got, want)
				}
			}
			if !reflect.DeepEqual(map[string]any(parent.Locals), tt.wantParentPrivate) {
				t.Errorf("parent.Locals = %v, want %v",
					parent.Locals, tt.wantParentPrivate)
			}
		})
	}
}

func TestNewChildContextSharesData(t *testing.T) {
	parent := NewRenderContext(map[string]any{"shared": "original"})
	child := NewChildContext(parent)
	child.Data.Set("shared", "modified")

	got, ok := parent.Get("shared")
	if !ok {
		t.Fatal("parent.Get(\"shared\") = _, false, want true")
	}
	if got != "modified" {
		t.Errorf("parent.Get(\"shared\") = %v, want %q", got, "modified")
	}
}

func TestNewChildContextPreservesRuntimeState(t *testing.T) {
	parent := NewRenderContext(map[string]any{"shared": "value"})
	parent.Set("private", "secret")
	parent.engine = New(WithFormat(FormatHTML))
	parent.autoescape = true
	parent.includeDepth = 3
	parent.currentLeaf = &Template{name: "leaf.txt"}

	child := NewChildContext(parent)

	if child.engine != parent.engine {
		t.Fatal("child.engine was not preserved")
	}
	if child.autoescape != parent.autoescape {
		t.Fatalf("child.autoescape = %v, want %v", child.autoescape, parent.autoescape)
	}
	if child.includeDepth != parent.includeDepth {
		t.Fatalf("child.includeDepth = %d, want %d", child.includeDepth, parent.includeDepth)
	}
	if child.currentLeaf != parent.currentLeaf {
		t.Fatal("child.currentLeaf was not preserved")
	}
}

func TestNewIsolatedChildContextPreservesRuntimeState(t *testing.T) {
	parent := NewRenderContext(map[string]any{"shared": "value"})
	parent.Set("private", "secret")
	parent.engine = New(WithFormat(FormatHTML))
	parent.autoescape = true
	parent.includeDepth = 2
	parent.currentLeaf = &Template{name: "leaf.txt"}

	child := NewIsolatedChildContext(parent)

	if child.Data != nil {
		t.Fatalf("child.Data = %v, want nil", child.Data)
	}
	if len(child.Locals) != 0 {
		t.Fatalf("len(child.Locals) = %d, want 0", len(child.Locals))
	}
	if child.engine != parent.engine {
		t.Fatal("child.engine was not preserved")
	}
	if child.autoescape != parent.autoescape {
		t.Fatalf("child.autoescape = %v, want %v", child.autoescape, parent.autoescape)
	}
	if child.includeDepth != parent.includeDepth {
		t.Fatalf("child.includeDepth = %d, want %d", child.includeDepth, parent.includeDepth)
	}
	if child.currentLeaf != parent.currentLeaf {
		t.Fatal("child.currentLeaf was not preserved")
	}
}

func TestRenderContextGetPriority(t *testing.T) {
	tests := []struct {
		name      string
		public    map[string]any
		private   map[string]any
		key       string
		wantVal   any
		wantFound bool
	}{
		{
			name:      "locals take precedence over data",
			public:    map[string]any{"key": "public_value"},
			private:   map[string]any{"key": "private_value"},
			key:       "key",
			wantVal:   "private_value",
			wantFound: true,
		},
		{
			name:      "falls back to data when not in locals",
			public:    map[string]any{"public_only": "found"},
			private:   map[string]any{},
			key:       "public_only",
			wantVal:   "found",
			wantFound: true,
		},
		{
			name:      "not found in either returns false",
			public:    map[string]any{"a": 1},
			private:   map[string]any{"b": 2},
			key:       "missing",
			wantVal:   nil,
			wantFound: false,
		},
		{
			name:      "private only key found",
			public:    map[string]any{},
			private:   map[string]any{"private_only": 42},
			key:       "private_only",
			wantVal:   42,
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ec := NewRenderContext(tt.public)
			for k, v := range tt.private {
				ec.Set(k, v)
			}
			got, ok := ec.Get(tt.key)
			if ok != tt.wantFound {
				t.Errorf("Get(%q) found = %v, want %v", tt.key, ok, tt.wantFound)
			}
			if got != tt.wantVal {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.wantVal)
			}
		})
	}
}
