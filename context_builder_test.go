package template

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewContextBuilder(t *testing.T) {
	got, err := NewContextBuilder().Build()
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}
	if want := (Context{}); !cmp.Equal(got, want) {
		t.Errorf("Build() = %v, want %v", got, want)
	}
}

func TestContextBuilderKeyValue(t *testing.T) {
	tests := []struct {
		name   string
		keys   []string
		values []any
		want   Context
	}{
		{
			name:   "single string key-value",
			keys:   []string{"name"},
			values: []any{"John"},
			want:   Context{"name": "John"},
		},
		{
			name:   "multiple key-values of different types",
			keys:   []string{"name", "age", "active"},
			values: []any{"John", 30, true},
			want:   Context{"name": "John", "age": 30, "active": true},
		},
		{
			name:   "nested key via dot notation",
			keys:   []string{"user.name"},
			values: []any{"John"},
			want: Context{
				"user": map[string]any{
					"name": "John",
				},
			},
		},
		{
			name:   "deeply nested key",
			keys:   []string{"a.b.c"},
			values: []any{"deep"},
			want: Context{
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
			want:   Context{"key": "second"},
		},
		{
			name:   "nil value stored correctly",
			keys:   []string{"empty"},
			values: []any{nil},
			want:   Context{"empty": nil},
		},
		{
			name:   "slice value stored correctly",
			keys:   []string{"items"},
			values: []any{[]int{1, 2, 3}},
			want:   Context{"items": []int{1, 2, 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewContextBuilder()
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

func TestContextBuilderStruct(t *testing.T) {
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

	tests := []struct {
		name  string
		input any
		want  Context
	}{
		{
			name:  "simple struct with string and bool fields",
			input: SimpleUser{Name: "John", Email: "john@test.com", Active: true},
			want: Context{
				"name":   "John",
				"email":  "john@test.com",
				"active": true,
			},
		},
		{
			name:  "struct with zero values",
			input: SimpleUser{},
			want: Context{
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
			want: Context{
				"name": "Jane",
				"profile": map[string]any{
					"bio":     "Engineer",
					"website": "https://example.com",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewContextBuilder().Struct(tt.input).Build()
			if err != nil {
				t.Fatalf("Build() unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Build() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestContextBuilderStructWithMarshalError(t *testing.T) {
	got, err := NewContextBuilder().Struct(make(chan int)).Build()
	if err == nil {
		t.Error("Build() = _, nil, want error for chan type")
	}
	if got == nil {
		t.Error("Build() returned nil context, want non-nil even on error")
	}
}

func TestContextBuilderChaining(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name  string
		build func() *ContextBuilder
		want  Context
	}{
		{
			name: "KeyValue then Struct merges both",
			build: func() *ContextBuilder {
				return NewContextBuilder().
					KeyValue("extra", "value").
					Struct(User{Name: "Alice"})
			},
			want: Context{
				"extra": "value",
				"name":  "Alice",
			},
		},
		{
			name: "Struct then KeyValue overrides struct field",
			build: func() *ContextBuilder {
				return NewContextBuilder().
					Struct(User{Name: "Alice"}).
					KeyValue("name", "Bob")
			},
			want: Context{"name": "Bob"},
		},
		{
			name: "multiple KeyValue calls",
			build: func() *ContextBuilder {
				return NewContextBuilder().
					KeyValue("a", 1).
					KeyValue("b", 2).
					KeyValue("c", 3)
			},
			want: Context{"a": 1, "b": 2, "c": 3},
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

func TestContextBuilderBuildErrorCollection(t *testing.T) {
	tests := []struct {
		name      string
		build     func() *ContextBuilder
		wantError bool
	}{
		{
			name: "no errors returns nil error",
			build: func() *ContextBuilder {
				return NewContextBuilder().KeyValue("key", "value")
			},
		},
		{
			name: "single Struct error collected",
			build: func() *ContextBuilder {
				return NewContextBuilder().Struct(make(chan int))
			},
			wantError: true,
		},
		{
			name: "multiple Struct errors collected",
			build: func() *ContextBuilder {
				return NewContextBuilder().
					Struct(make(chan int)).
					Struct(make(chan string))
			},
			wantError: true,
		},
		{
			name: "mixed valid and invalid still reports error",
			build: func() *ContextBuilder {
				type Valid struct {
					X string `json:"x"`
				}
				return NewContextBuilder().
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
		publicData        map[string]any
		parentPrivate     map[string]any
		childPrivate      map[string]any
		wantChildGet      map[string]any
		wantParentPrivate map[string]any
	}{
		{
			name:              "child inherits public variables",
			publicData:        map[string]any{"name": "Alice", "age": 30},
			parentPrivate:     map[string]any{},
			childPrivate:      map[string]any{},
			wantChildGet:      map[string]any{"name": "Alice", "age": 30},
			wantParentPrivate: map[string]any{},
		},
		{
			name:              "child inherits parent private variables",
			publicData:        map[string]any{},
			parentPrivate:     map[string]any{"counter": 1, "flag": true},
			childPrivate:      map[string]any{},
			wantChildGet:      map[string]any{"counter": 1, "flag": true},
			wantParentPrivate: map[string]any{"counter": 1, "flag": true},
		},
		{
			name:              "child private modification does not affect parent",
			publicData:        map[string]any{},
			parentPrivate:     map[string]any{"x": 10},
			childPrivate:      map[string]any{"x": 99, "y": 20},
			wantChildGet:      map[string]any{"x": 99, "y": 20},
			wantParentPrivate: map[string]any{"x": 10},
		},
		{
			name:              "child private does not leak to parent",
			publicData:        map[string]any{},
			parentPrivate:     map[string]any{},
			childPrivate:      map[string]any{"new_var": "child_only"},
			wantChildGet:      map[string]any{"new_var": "child_only"},
			wantParentPrivate: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := NewExecutionContext(tt.publicData)
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
			if !reflect.DeepEqual(map[string]any(parent.Private), tt.wantParentPrivate) {
				t.Errorf("parent.Private = %v, want %v",
					parent.Private, tt.wantParentPrivate)
			}
		})
	}
}

func TestNewChildContextSharesPublic(t *testing.T) {
	parent := NewExecutionContext(map[string]any{"shared": "original"})
	child := NewChildContext(parent)
	child.Public.Set("shared", "modified")

	got, ok := parent.Get("shared")
	if !ok {
		t.Fatal("parent.Get(\"shared\") = _, false, want true")
	}
	if got != "modified" {
		t.Errorf("parent.Get(\"shared\") = %v, want %q", got, "modified")
	}
}

func TestExecutionContextGetPriority(t *testing.T) {
	tests := []struct {
		name      string
		public    map[string]any
		private   map[string]any
		key       string
		wantVal   any
		wantFound bool
	}{
		{
			name:      "private takes precedence over public",
			public:    map[string]any{"key": "public_value"},
			private:   map[string]any{"key": "private_value"},
			key:       "key",
			wantVal:   "private_value",
			wantFound: true,
		},
		{
			name:      "falls back to public when not in private",
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
			ec := NewExecutionContext(tt.public)
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
