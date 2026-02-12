package template

import (
	"reflect"
	"testing"
)

// Test helpers to reduce repetition in pointer construction.

func intPtr(v int) *int             { return &v }
func float64Ptr(v float64) *float64 { return &v }
func stringPtr(v string) *string    { return &v }
func boolPtr(v bool) *bool          { return &v }

func TestValue_IsNil(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil value", nil, true},
		{"nil pointer", (*int)(nil), true},
		{"nil slice", ([]int)(nil), false},
		{"nil map", (map[string]int)(nil), false},
		{"nil interface", (any)(nil), true},
		{"zero int", 0, false},
		{"non-zero int", 42, false},
		{"pointer to zero int", intPtr(0), false},
		{"pointer to non-zero int", intPtr(42), false},
		{"empty string", "", false},
		{"non-empty string", "hello", false},
		{"empty slice", []int{}, false},
		{"non-empty slice", []int{1, 2, 3}, false},
		{"empty map", map[string]int{}, false},
		{"non-empty map", map[string]int{"a": 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewValue(tt.value).IsNil(); got != tt.want {
				t.Errorf("NewValue(%v).IsNil() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValue_IsTrue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, false},
		{"nil pointer", (*int)(nil), false},
		{"bool true", true, true},
		{"bool false", false, false},
		{"pointer to bool true", boolPtr(true), true},
		{"pointer to bool false", boolPtr(false), false},
		{"int zero", 0, false},
		{"int positive", 42, true},
		{"int negative", -42, true},
		{"pointer to int zero", intPtr(0), false},
		{"pointer to int positive", intPtr(42), true},
		{"int8 zero", int8(0), false},
		{"int8 non-zero", int8(42), true},
		{"int64 zero", int64(0), false},
		{"int64 non-zero", int64(42), true},
		{"uint zero", uint(0), false},
		{"uint non-zero", uint(42), true},
		{"uint64 zero", uint64(0), false},
		{"uint64 non-zero", uint64(42), true},
		{"float32 zero", float32(0), false},
		{"float32 non-zero", float32(3.14), true},
		{"float64 zero", float64(0), false},
		{"float64 non-zero", float64(3.14), true},
		{"pointer to float64 zero", float64Ptr(0), false},
		{"pointer to float64 non-zero", float64Ptr(3.14), true},
		{"empty string", "", false},
		{"non-empty string", "hello", true},
		{"pointer to empty string", stringPtr(""), false},
		{"pointer to non-empty string", stringPtr("hello"), true},
		{"nil slice", ([]int)(nil), false},
		{"empty slice", []int{}, false},
		{"non-empty slice", []int{1, 2, 3}, true},
		{"nil map", (map[string]int)(nil), false},
		{"empty map", map[string]int{}, false},
		{"non-empty map", map[string]int{"a": 1}, true},
		{"empty array", [0]int{}, false},
		{"non-empty array", [3]int{1, 2, 3}, true},
		{"empty struct", struct{}{}, true},
		{"non-empty struct", struct{ Name string }{Name: "test"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewValue(tt.value).IsTrue(); got != tt.want {
				t.Errorf("NewValue(%v).IsTrue() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValue_String(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"nil", nil, ""},
		{"nil pointer", (*int)(nil), ""},
		{"int", 42, "42"},
		{"pointer to int", intPtr(42), "42"},
		{"negative int", -42, "-42"},
		{"float", 3.14, "3.14"},
		{"pointer to float", float64Ptr(3.14), "3.14"},
		{"string", "hello", "hello"},
		{"pointer to string", stringPtr("hello"), "hello"},
		{"empty string", "", ""},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"pointer to bool", boolPtr(true), "true"},
		{"slice", []int{1, 2, 3}, "[1,2,3]"},
		{"map", map[string]int{"a": 1}, `{"a":1}`},
		{"nested slice", [][]int{{1, 2}, {3, 4}}, "[[1,2],[3,4]]"},
		{"empty slice", []int{}, "[]"},
		{"bool slice", []bool{true, false, true}, "[true,false,true]"},

		// Unsigned integer types.
		{"uint", uint(100), "100"},
		{"uint8", uint8(255), "255"},
		{"uint16", uint16(1000), "1000"},
		{"uint32", uint32(70000), "70000"},
		{"uint64", uint64(99999), "99999"},
		{"uint zero", uint(0), "0"},

		// Float formatting.
		{"whole number float", 42.0, "42"},
		{"zero float", 0.0, "0"},
		{"negative whole float", -10.0, "-10"},
		{"decimal float", 3.14, "3.14"},
		{"negative decimal float", -2.5, "-2.5"},
		{"small decimal float", 0.001, "0.001"},
		{"large whole float", 1000000.0, "1000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewValue(tt.value).String(); got != tt.want {
				t.Errorf("NewValue(%v).String() = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestValue_Int(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		want      int64
		wantError bool
	}{
		{"nil", nil, 0, true},
		{"nil pointer", (*int)(nil), 0, true},
		{"int", 42, 42, false},
		{"pointer to int", intPtr(42), 42, false},
		{"int8", int8(42), 42, false},
		{"int16", int16(42), 42, false},
		{"int32", int32(42), 42, false},
		{"int64", int64(42), 42, false},
		{"uint", uint(42), 42, false},
		{"uint8", uint8(42), 42, false},
		{"uint16", uint16(42), 42, false},
		{"uint32", uint32(42), 42, false},
		{"uint64", uint64(42), 42, false},
		{"float32", float32(42.7), 42, false},
		{"float64", float64(42.7), 42, false},
		{"pointer to float64", float64Ptr(42.7), 42, false},
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		{"pointer to bool true", boolPtr(true), 1, false},
		{"string", "hello", 0, true},
		{"negative int", -42, -42, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewValue(tt.value).Int()
			if tt.wantError {
				if err == nil {
					t.Errorf("NewValue(%v).Int() error = nil, want error", tt.value)
				}
				return
			}
			if err != nil {
				t.Errorf("NewValue(%v).Int() unexpected error: %v", tt.value, err)
				return
			}
			if got != tt.want {
				t.Errorf("NewValue(%v).Int() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValue_Float(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		want      float64
		wantError bool
	}{
		{"nil", nil, 0, true},
		{"nil pointer", (*float64)(nil), 0, true},
		{"float32", float32(3.14), float64(float32(3.14)), false},
		{"float64", float64(3.14), 3.14, false},
		{"pointer to float64", float64Ptr(3.14), 3.14, false},
		{"int", 42, 42.0, false},
		{"pointer to int", intPtr(42), 42.0, false},
		{"int8", int8(42), 42.0, false},
		{"int16", int16(42), 42.0, false},
		{"int32", int32(42), 42.0, false},
		{"int64", int64(42), 42.0, false},
		{"uint", uint(42), 42.0, false},
		{"uint8", uint8(42), 42.0, false},
		{"uint16", uint16(42), 42.0, false},
		{"uint32", uint32(42), 42.0, false},
		{"uint64", uint64(42), 42.0, false},
		{"string", "hello", 0, true},
		{"negative float", -3.14, -3.14, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewValue(tt.value).Float()
			if tt.wantError {
				if err == nil {
					t.Errorf("NewValue(%v).Float() error = nil, want error", tt.value)
				}
				return
			}
			if err != nil {
				t.Errorf("NewValue(%v).Float() unexpected error: %v", tt.value, err)
				return
			}
			if got != tt.want {
				t.Errorf("NewValue(%v).Float() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValue_Bool(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, false},
		{"bool true", true, true},
		{"bool false", false, false},
		{"pointer to bool true", boolPtr(true), true},
		{"pointer to bool false", boolPtr(false), false},
		{"int zero", 0, false},
		{"int non-zero", 42, true},
		{"empty string", "", false},
		{"non-empty string", "hello", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewValue(tt.value).Bool(); got != tt.want {
				t.Errorf("NewValue(%v).Bool() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValue_Len(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		want      int
		wantError bool
	}{
		{"nil", nil, 0, false},
		{"nil slice", ([]int)(nil), 0, false},
		{"empty slice", []int{}, 0, false},
		{"non-empty slice", []int{1, 2, 3}, 3, false},
		{"pointer to slice", func() *[]int { s := []int{1, 2, 3}; return &s }(), 3, false},
		{"empty string", "", 0, false},
		{"non-empty string", "hello", 5, false},
		{"pointer to string", stringPtr("hello"), 5, false},
		{"empty map", map[string]int{}, 0, false},
		{"non-empty map", map[string]int{"a": 1, "b": 2}, 2, false},
		{"pointer to map", func() *map[string]int { m := map[string]int{"a": 1}; return &m }(), 1, false},
		{"array", [3]int{1, 2, 3}, 3, false},
		{"int", 42, 0, true},
		{"bool", true, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewValue(tt.value).Len()
			if tt.wantError {
				if err == nil {
					t.Errorf("NewValue(%v).Len() error = nil, want error", tt.value)
				}
				return
			}
			if err != nil {
				t.Errorf("NewValue(%v).Len() unexpected error: %v", tt.value, err)
				return
			}
			if got != tt.want {
				t.Errorf("NewValue(%v).Len() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValue_Index(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		index     int
		want      any
		wantError bool
	}{
		{"nil", nil, 0, nil, true},
		{"slice-valid", []int{10, 20, 30}, 1, 20, false},
		{"slice-first", []int{10, 20, 30}, 0, 10, false},
		{"slice-last", []int{10, 20, 30}, 2, 30, false},
		{"slice-negative", []int{10, 20, 30}, -1, nil, true},
		{"slice-out-of-range", []int{10, 20, 30}, 3, nil, true},
		{"pointer-to-slice", func() *[]int { s := []int{10, 20, 30}; return &s }(), 1, 20, false},
		{"array-valid", [3]int{10, 20, 30}, 1, 20, false},
		{"string-valid", "hello", 1, "e", false},
		{"string-first", "hello", 0, "h", false},
		{"string-last", "hello", 4, "o", false},
		{"string-out-of-range", "hello", 5, nil, true},
		{"pointer-to-string", stringPtr("hello"), 1, "e", false},
		{"map-not-indexable", map[string]int{"a": 1}, 0, nil, true},
		{"int-not-indexable", 42, 0, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewValue(tt.value).Index(tt.index)
			if tt.wantError {
				if err == nil {
					t.Errorf("NewValue(%v).Index(%d) error = nil, want error", tt.value, tt.index)
				}
				return
			}
			if err != nil {
				t.Errorf("NewValue(%v).Index(%d) unexpected error: %v", tt.value, tt.index, err)
				return
			}
			if got.Interface() != tt.want {
				t.Errorf("NewValue(%v).Index(%d) = %v, want %v", tt.value, tt.index, got.Interface(), tt.want)
			}
		})
	}
}

func TestValue_Key(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		key       any
		want      any
		wantError bool
	}{
		{"nil", nil, "key", nil, true},
		{"string-key-found", map[string]int{"a": 1, "b": 2}, "a", 1, false},
		{"string-key-not-found", map[string]int{"a": 1, "b": 2}, "c", nil, false},
		{"pointer-to-map", func() *map[string]int { m := map[string]int{"a": 1}; return &m }(), "a", 1, false},
		{"int-key-found", map[int]string{1: "one", 2: "two"}, 1, "one", false},
		{"int-key-not-found", map[int]string{1: "one", 2: "two"}, 3, nil, false},
		{"slice-not-map", []int{1, 2, 3}, 0, nil, true},
		{"string-not-map", "hello", 0, nil, true},
		{"int-not-map", 42, 0, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewValue(tt.value).Key(tt.key)
			if tt.wantError {
				if err == nil {
					t.Errorf("NewValue(%v).Key(%v) error = nil, want error", tt.value, tt.key)
				}
				return
			}
			if err != nil {
				t.Errorf("NewValue(%v).Key(%v) unexpected error: %v", tt.value, tt.key, err)
				return
			}
			if got.Interface() != tt.want {
				t.Errorf("NewValue(%v).Key(%v) = %v, want %v", tt.value, tt.key, got.Interface(), tt.want)
			}
		})
	}
}

func TestValue_Field(t *testing.T) {
	type testStruct struct {
		Name  string
		Age   int
		Email string
	}

	tests := []struct {
		name      string
		value     any
		field     string
		want      any
		wantError bool
	}{
		{"nil", nil, "Name", nil, true},
		{"struct-field-exists", testStruct{Name: "Alice", Age: 30, Email: "alice@example.com"}, "Name", "Alice", false},
		{"struct-field-not-exists", testStruct{Name: "Alice", Age: 30, Email: "alice@example.com"}, "Phone", nil, true},
		{"pointer-to-struct", &testStruct{Name: "Bob", Age: 25, Email: "bob@example.com"}, "Age", 25, false},
		{"map-key-exists", map[string]any{"Name": "Charlie", "Age": 35}, "Name", "Charlie", false},
		{"map-key-not-exists", map[string]any{"Name": "Charlie", "Age": 35}, "Email", nil, false},
		{"pointer-to-map", func() *map[string]string { m := map[string]string{"Name": "David"}; return &m }(), "Name", "David", false},
		{"slice-no-field", []int{1, 2, 3}, "Name", nil, true},
		{"int-no-field", 42, "Name", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewValue(tt.value).Field(tt.field)
			if tt.wantError {
				if err == nil {
					t.Errorf("NewValue(%v).Field(%q) error = nil, want error", tt.value, tt.field)
				}
				return
			}
			if err != nil {
				t.Errorf("NewValue(%v).Field(%q) unexpected error: %v", tt.value, tt.field, err)
				return
			}
			if got.Interface() != tt.want {
				t.Errorf("NewValue(%v).Field(%q) = %v, want %v", tt.value, tt.field, got.Interface(), tt.want)
			}
		})
	}
}

func TestValue_Field_JSONTag(t *testing.T) {
	type tagged struct {
		FullName string `json:"name"`
		Age      int    `json:"age,omitempty"`
		Hidden   string `json:"-"`
		NoTag    string
	}

	val := tagged{
		FullName: "Alice",
		Age:      30,
		Hidden:   "secret",
		NoTag:    "visible",
	}

	tests := []struct {
		name      string
		field     string
		want      any
		wantError bool
	}{
		{"json-tag-name", "name", "Alice", false},
		{"json-tag-with-omitempty", "age", 30, false},
		{"hidden-field-via-dash", "-", nil, true},
		{"field-by-exported-name", "NoTag", "visible", false},
		{"field-by-exported-name-FullName", "FullName", "Alice", false},
		{"nonexistent-field", "missing", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewValue(val).Field(tt.field)
			if tt.wantError {
				if err == nil {
					t.Errorf("Value.Field(%q) error = nil, want error", tt.field)
				}
				return
			}
			if err != nil {
				t.Errorf("Value.Field(%q) unexpected error: %v", tt.field, err)
				return
			}
			if got.Interface() != tt.want {
				t.Errorf("Value.Field(%q) = %v, want %v", tt.field, got.Interface(), tt.want)
			}
		})
	}
}

func TestValue_Iterate(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		wantKeys  []any
		wantVals  []any
		wantCount int
		wantError bool
	}{
		{"nil", nil, nil, nil, 0, false},
		{"empty slice", []int{}, nil, nil, 0, false},
		{
			"slice with elements",
			[]int{10, 20, 30},
			[]any{0, 1, 2},
			[]any{10, 20, 30},
			3, false,
		},
		{
			"pointer to slice",
			func() *[]int { s := []int{10, 20}; return &s }(),
			[]any{0, 1},
			[]any{10, 20},
			2, false,
		},
		{
			"array",
			[3]string{"a", "b", "c"},
			[]any{0, 1, 2},
			[]any{"a", "b", "c"},
			3, false,
		},
		{
			"map",
			map[string]int{"a": 1, "b": 2},
			[]any{"a", "b"},
			[]any{1, 2},
			2, false,
		},
		{
			"pointer to map",
			func() *map[string]int {
				m := map[string]int{"x": 10}
				return &m
			}(),
			[]any{"x"},
			[]any{10},
			1, false,
		},
		{"int-not-iterable", 42, nil, nil, 0, true},
		{
			"string-iterable",
			"hello",
			[]any{0, 1, 2, 3, 4},
			[]any{"h", "e", "l", "l", "o"},
			5, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			var keys, vals []any
			var count int

			err := v.Iterate(func(_ int, c int, key, val *Value) bool {
				count = c
				keys = append(keys, key.Interface())
				vals = append(vals, val.Interface())
				return true
			})

			if tt.wantError {
				if err == nil {
					t.Errorf("NewValue(%v).Iterate() error = nil, want error", tt.value)
				}
				return
			}
			if err != nil {
				t.Errorf("NewValue(%v).Iterate() unexpected error: %v", tt.value, err)
				return
			}
			if count != tt.wantCount {
				t.Errorf("NewValue(%v).Iterate() count = %d, want %d", tt.value, count, tt.wantCount)
			}
			if tt.wantCount > 0 {
				if !elementsMatch(keys, tt.wantKeys) {
					t.Errorf("NewValue(%v).Iterate() keys = %v, want %v", tt.value, keys, tt.wantKeys)
				}
				if !elementsMatch(vals, tt.wantVals) {
					t.Errorf("NewValue(%v).Iterate() vals = %v, want %v", tt.value, vals, tt.wantVals)
				}
			}
		})
	}
}

// elementsMatch reports whether two slices contain the same elements
// regardless of order.
func elementsMatch(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	used := make([]bool, len(b))
	for _, av := range a {
		found := false
		for j, bv := range b {
			if !used[j] && reflect.DeepEqual(av, bv) {
				used[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestValue_Iterate_EarlyExit(t *testing.T) {
	v := NewValue([]int{1, 2, 3, 4, 5})
	var got []int

	err := v.Iterate(func(idx, _ int, _ *Value, val *Value) bool {
		got = append(got, val.Interface().(int))
		return idx < 2
	})

	if err != nil {
		t.Fatalf("Iterate() unexpected error: %v", err)
	}
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Iterate() collected = %v, want %v", got, want)
	}
}

func TestValue_Compare(t *testing.T) {
	tests := []struct {
		name string
		a    any
		b    any
		want int
	}{
		{"both nil", nil, nil, 0},
		{"first nil", nil, 42, -1},
		{"second nil", 42, nil, 1},
		{"int equal", 42, 42, 0},
		{"int less", 10, 20, -1},
		{"int greater", 20, 10, 1},
		{"pointer-to-int less", intPtr(10), intPtr(20), -1},
		{"float equal", 3.14, 3.14, 0},
		{"float less", 3.14, 6.28, -1},
		{"float greater", 6.28, 3.14, 1},
		{"pointer-to-float less", float64Ptr(3.14), float64Ptr(6.28), -1},
		{"mixed int-float equal", 42, 42.0, 0},
		{"mixed int-float less", 10, 20.5, -1},
		{"string equal", "hello", "hello", 0},
		{"string less", "abc", "xyz", -1},
		{"string greater", "xyz", "abc", 1},
		{"pointer-to-string less", stringPtr("abc"), stringPtr("xyz"), -1},
		{"negative numbers", -10, -5, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewValue(tt.a).Compare(NewValue(tt.b))
			if err != nil {
				t.Fatalf("Compare(%v, %v) unexpected error: %v", tt.a, tt.b, err)
			}
			if got != tt.want {
				t.Errorf("Compare(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestValue_Equals(t *testing.T) {
	tests := []struct {
		name string
		a    any
		b    any
		want bool
	}{
		{"both nil", nil, nil, true},
		{"first nil", nil, 42, false},
		{"second nil", 42, nil, false},
		{"int equal", 42, 42, true},
		{"int not equal", 42, 43, false},
		{"pointer-to-int equal", intPtr(42), intPtr(42), true},
		{"pointer-to-int not equal", intPtr(42), intPtr(43), false},
		{"float equal", 3.14, 3.14, true},
		{"float not equal", 3.14, 6.28, false},
		{"pointer-to-float equal", float64Ptr(3.14), float64Ptr(3.14), true},
		{"string equal", "hello", "hello", true},
		{"string not equal", "hello", "world", false},
		{"pointer-to-string equal", stringPtr("hello"), stringPtr("hello"), true},
		{"bool equal", true, true, true},
		{"bool not equal", true, false, false},
		{"pointer-to-bool equal", boolPtr(true), boolPtr(true), true},
		{"slice equal", []int{1, 2, 3}, []int{1, 2, 3}, true},
		{"slice not equal", []int{1, 2, 3}, []int{1, 2, 4}, false},
		{"map equal", map[string]int{"a": 1, "b": 2}, map[string]int{"a": 1, "b": 2}, true},
		{"map not equal", map[string]int{"a": 1, "b": 2}, map[string]int{"a": 1, "b": 3}, false},
		{"different types", 42, "42", false},

		// Cross-numeric type comparisons.
		{"int and float64 equal", 42, float64(42), true},
		{"int and float64 not equal", 42, float64(42.5), false},
		{"int64 and uint equal", int64(100), uint(100), true},
		{"int32 and float32 equal", int32(7), float32(7), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewValue(tt.a).Equals(NewValue(tt.b)); got != tt.want {
				t.Errorf("Equals(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestValue_Interface(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  any
	}{
		{"nil", nil, nil},
		{"int", 42, 42},
		{"string", "hello", "hello"},
		{"bool", true, true},
		{"slice", []int{1, 2, 3}, []int{1, 2, 3}},
		{"map", map[string]int{"a": 1}, map[string]int{"a": 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValue(tt.value).Interface()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewValue(%v).Interface() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestNewValue(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"nil", nil},
		{"int", 42},
		{"string", "hello"},
		{"bool", true},
		{"slice", []int{1, 2, 3}},
		{"map", map[string]int{"a": 1}},
		{"pointer", intPtr(42)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.input)
			if v == nil {
				t.Fatal("NewValue() returned nil")
			}
			if !reflect.DeepEqual(v.Interface(), tt.input) {
				t.Errorf("NewValue(%v).Interface() = %v, want %v", tt.input, v.Interface(), tt.input)
			}
		})
	}
}

// =============================================================================
// Edge Case Tests for Coverage
// =============================================================================

func TestValueIsTrueEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"uint zero", uint(0), false},
		{"uint nonzero", uint(1), true},
		{"empty array", [0]int{}, false},
		{"nonempty array", [2]int{1, 2}, true},
		{"struct", struct{ X int }{1}, true},
		{"func", func() {}, true},
		{"channel", make(chan int), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.input)
			if got := v.IsTrue(); got != tt.expected {
				t.Errorf("IsTrue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValueFormatSliceItemEdgeCases(t *testing.T) {
	// Nested slices.
	v := NewValue([][]int{{1, 2}, {3, 4}})
	got := v.String()
	if got != "[[1,2],[3,4]]" {
		t.Errorf("String() = %q, want %q", got, "[[1,2],[3,4]]")
	}

	// Slice of maps.
	v2 := NewValue([]map[string]int{{"a": 1}})
	s := v2.String()
	if s == "" {
		t.Error("String() returned empty for slice of maps")
	}

	// Slice of bools.
	v3 := NewValue([]bool{true, false})
	if got := v3.String(); got != "[true,false]" {
		t.Errorf("String() = %q, want %q", got, "[true,false]")
	}

	// Slice of uints.
	v4 := NewValue([]uint{10, 20})
	if got := v4.String(); got != "[10,20]" {
		t.Errorf("String() = %q, want %q", got, "[10,20]")
	}

	// Slice of floats.
	v5 := NewValue([]float64{1.5, 2.5})
	if got := v5.String(); got != "[1.5,2.5]" {
		t.Errorf("String() = %q, want %q", got, "[1.5,2.5]")
	}
}

func TestValueLessEdgeCases(t *testing.T) {
	// String keys.
	keys := sortedKeys{reflect.ValueOf("banana"), reflect.ValueOf("apple")}
	if !keys.Less(1, 0) {
		t.Error("Less('apple', 'banana') should be true")
	}
	if keys.Less(0, 1) {
		t.Error("Less('banana', 'apple') should be false")
	}

	// Numeric keys.
	numKeys := sortedKeys{reflect.ValueOf(3), reflect.ValueOf(1)}
	if !numKeys.Less(1, 0) {
		t.Error("Less(1, 3) should be true")
	}

	// Mixed: one numeric, one not — falls back to string.
	mixedKeys := sortedKeys{reflect.ValueOf("abc"), reflect.ValueOf(1)}
	_ = mixedKeys.Less(0, 1) // cover the branch
}

func TestValueStringUint(t *testing.T) {
	v := NewValue(uint(42))
	if got := v.String(); got != "42" {
		t.Errorf("String() = %q, want %q", got, "42")
	}
}

type stringerVal string

func (s stringerVal) String() string { return "stringer:" + string(s) }

func TestValueStringStringer(t *testing.T) {
	v := NewValue(stringerVal("hello"))
	if got := v.String(); got != "stringer:hello" {
		t.Errorf("String() = %q, want %q", got, "stringer:hello")
	}
}

func TestValueStringJSONFallback(t *testing.T) {
	v := NewValue(map[string]int{"x": 1})
	got := v.String()
	if got == "" {
		t.Error("String() returned empty for map")
	}
}
