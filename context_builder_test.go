package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContextBuilder(t *testing.T) {
	builder := NewContextBuilder()
	ctx, err := builder.Build()

	assert.NoError(t, err)
	assert.Equal(t, Context{}, ctx)
}

func TestContextBuilderKeyValue(t *testing.T) {
	tests := []struct {
		name     string
		keys     []string
		values   []interface{}
		expected Context
	}{
		{
			name:     "single string key-value",
			keys:     []string{"name"},
			values:   []interface{}{"John"},
			expected: Context{"name": "John"},
		},
		{
			name:     "multiple key-values of different types",
			keys:     []string{"name", "age", "active"},
			values:   []interface{}{"John", 30, true},
			expected: Context{"name": "John", "age": 30, "active": true},
		},
		{
			name:   "nested key via dot notation",
			keys:   []string{"user.name"},
			values: []interface{}{"John"},
			expected: Context{
				"user": map[string]interface{}{
					"name": "John",
				},
			},
		},
		{
			name:   "deeply nested key",
			keys:   []string{"a.b.c"},
			values: []interface{}{"deep"},
			expected: Context{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "deep",
					},
				},
			},
		},
		{
			name:     "overwrite same key keeps last value",
			keys:     []string{"key", "key"},
			values:   []interface{}{"first", "second"},
			expected: Context{"key": "second"},
		},
		{
			name:     "nil value stored correctly",
			keys:     []string{"empty"},
			values:   []interface{}{nil},
			expected: Context{"empty": nil},
		},
		{
			name:     "slice value stored correctly",
			keys:     []string{"items"},
			values:   []interface{}{[]int{1, 2, 3}},
			expected: Context{"items": []int{1, 2, 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewContextBuilder()
			for i, key := range tt.keys {
				builder.KeyValue(key, tt.values[i])
			}
			ctx, err := builder.Build()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, ctx)
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
		name     string
		input    interface{}
		expected Context
	}{
		{
			name:  "simple struct with string and bool fields",
			input: SimpleUser{Name: "John", Email: "john@test.com", Active: true},
			expected: Context{
				"name":   "John",
				"email":  "john@test.com",
				"active": true,
			},
		},
		{
			name:  "struct with zero values",
			input: SimpleUser{},
			expected: Context{
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
			expected: Context{
				"name": "Jane",
				"profile": map[string]interface{}{
					"bio":     "Engineer",
					"website": "https://example.com",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewContextBuilder().Struct(tt.input)
			ctx, err := builder.Build()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, ctx)
		})
	}
}

func TestContextBuilderStructWithMarshalError(t *testing.T) {
	builder := NewContextBuilder().Struct(make(chan int))
	ctx, err := builder.Build()

	assert.Error(t, err)
	// Context is still returned (empty) even when errors occur
	assert.NotNil(t, ctx)
}

func TestContextBuilderChaining(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name     string
		build    func() *ContextBuilder
		expected Context
	}{
		{
			name: "KeyValue then Struct merges both",
			build: func() *ContextBuilder {
				return NewContextBuilder().
					KeyValue("extra", "value").
					Struct(User{Name: "Alice"})
			},
			expected: Context{
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
			expected: Context{
				"name": "Bob",
			},
		},
		{
			name: "multiple KeyValue calls",
			build: func() *ContextBuilder {
				return NewContextBuilder().
					KeyValue("a", 1).
					KeyValue("b", 2).
					KeyValue("c", 3)
			},
			expected: Context{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.build()
			ctx, err := builder.Build()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, ctx)
		})
	}
}

func TestContextBuilderBuildErrorCollection(t *testing.T) {
	tests := []struct {
		name        string
		build       func() *ContextBuilder
		expectError bool
	}{
		{
			name: "no errors returns nil error",
			build: func() *ContextBuilder {
				return NewContextBuilder().KeyValue("key", "value")
			},
			expectError: false,
		},
		{
			name: "single Struct error collected",
			build: func() *ContextBuilder {
				return NewContextBuilder().Struct(make(chan int))
			},
			expectError: true,
		},
		{
			name: "multiple Struct errors collected",
			build: func() *ContextBuilder {
				return NewContextBuilder().
					Struct(make(chan int)).
					Struct(make(chan string))
			},
			expectError: true,
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
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.build()
			ctx, err := builder.Build()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NotNil(t, ctx)
		})
	}
}

func TestNewChildContext(t *testing.T) {
	tests := []struct {
		name                   string
		publicData             map[string]interface{}
		parentPrivateSetup     map[string]interface{}
		childPrivateSetup      map[string]interface{}
		expectChildGet         map[string]interface{}
		expectParentPrivateGet map[string]interface{}
	}{
		{
			name:               "child inherits public variables",
			publicData:         map[string]interface{}{"name": "Alice", "age": 30},
			parentPrivateSetup: map[string]interface{}{},
			childPrivateSetup:  map[string]interface{}{},
			expectChildGet: map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
			expectParentPrivateGet: map[string]interface{}{},
		},
		{
			name:               "child inherits parent private variables",
			publicData:         map[string]interface{}{},
			parentPrivateSetup: map[string]interface{}{"counter": 1, "flag": true},
			childPrivateSetup:  map[string]interface{}{},
			expectChildGet: map[string]interface{}{
				"counter": 1,
				"flag":    true,
			},
			expectParentPrivateGet: map[string]interface{}{
				"counter": 1,
				"flag":    true,
			},
		},
		{
			name:               "child private modification does not affect parent",
			publicData:         map[string]interface{}{},
			parentPrivateSetup: map[string]interface{}{"x": 10},
			childPrivateSetup:  map[string]interface{}{"x": 99, "y": 20},
			expectChildGet: map[string]interface{}{
				"x": 99,
				"y": 20,
			},
			expectParentPrivateGet: map[string]interface{}{
				"x": 10,
			},
		},
		{
			name:               "child private does not leak to parent",
			publicData:         map[string]interface{}{},
			parentPrivateSetup: map[string]interface{}{},
			childPrivateSetup:  map[string]interface{}{"new_var": "child_only"},
			expectChildGet: map[string]interface{}{
				"new_var": "child_only",
			},
			expectParentPrivateGet: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := NewExecutionContext(tt.publicData)
			for k, v := range tt.parentPrivateSetup {
				parent.Set(k, v)
			}

			child := NewChildContext(parent)
			for k, v := range tt.childPrivateSetup {
				child.Set(k, v)
			}

			// Verify child can access expected variables
			for key, expectedVal := range tt.expectChildGet {
				val, ok := child.Get(key)
				assert.Equal(t, true, ok, "child should find key %q", key)
				assert.Equal(t, expectedVal, val, "child value mismatch for key %q", key)
			}

			// Verify parent private is unchanged
			assert.Equal(t, tt.expectParentPrivateGet, map[string]interface{}(parent.Private))
		})
	}
}

func TestNewChildContextSharesPublic(t *testing.T) {
	parent := NewExecutionContext(map[string]interface{}{"shared": "original"})
	child := NewChildContext(parent)

	// Modify public through child
	child.Public.Set("shared", "modified")

	// Parent should see the modification since Public is shared
	val, ok := parent.Get("shared")
	assert.Equal(t, true, ok)
	assert.Equal(t, "modified", val)
}

func TestExecutionContextGetPriority(t *testing.T) {
	tests := []struct {
		name          string
		publicData    map[string]interface{}
		privateData   map[string]interface{}
		queryKey      string
		expectedVal   interface{}
		expectedFound bool
	}{
		{
			name:          "private takes precedence over public",
			publicData:    map[string]interface{}{"key": "public_value"},
			privateData:   map[string]interface{}{"key": "private_value"},
			queryKey:      "key",
			expectedVal:   "private_value",
			expectedFound: true,
		},
		{
			name:          "falls back to public when not in private",
			publicData:    map[string]interface{}{"public_only": "found"},
			privateData:   map[string]interface{}{},
			queryKey:      "public_only",
			expectedVal:   "found",
			expectedFound: true,
		},
		{
			name:          "not found in either returns false",
			publicData:    map[string]interface{}{"a": 1},
			privateData:   map[string]interface{}{"b": 2},
			queryKey:      "missing",
			expectedVal:   nil,
			expectedFound: false,
		},
		{
			name:          "private only key found",
			publicData:    map[string]interface{}{},
			privateData:   map[string]interface{}{"private_only": 42},
			queryKey:      "private_only",
			expectedVal:   42,
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewExecutionContext(tt.publicData)
			for k, v := range tt.privateData {
				ctx.Set(k, v)
			}

			val, ok := ctx.Get(tt.queryKey)
			assert.Equal(t, tt.expectedFound, ok)
			assert.Equal(t, tt.expectedVal, val)
		})
	}
}
