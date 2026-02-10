package template

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValue_IsNil(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{
			name:     "nil value",
			value:    nil,
			expected: true,
		},
		{
			name:     "nil pointer",
			value:    (*int)(nil),
			expected: true,
		},
		{
			name:     "nil slice",
			value:    ([]int)(nil),
			expected: false,
		},
		{
			name:     "nil map",
			value:    (map[string]int)(nil),
			expected: false,
		},
		{
			name:     "nil interface",
			value:    (interface{})(nil),
			expected: true,
		},
		{
			name:     "zero int",
			value:    0,
			expected: false,
		},
		{
			name:     "non-zero int",
			value:    42,
			expected: false,
		},
		{
			name:     "pointer to zero int",
			value:    func() *int { i := 0; return &i }(),
			expected: false,
		},
		{
			name:     "pointer to non-zero int",
			value:    func() *int { i := 42; return &i }(),
			expected: false,
		},
		{
			name:     "empty string",
			value:    "",
			expected: false,
		},
		{
			name:     "non-empty string",
			value:    "hello",
			expected: false,
		},
		{
			name:     "empty slice",
			value:    []int{},
			expected: false,
		},
		{
			name:     "non-empty slice",
			value:    []int{1, 2, 3},
			expected: false,
		},
		{
			name:     "empty map",
			value:    map[string]int{},
			expected: false,
		},
		{
			name:     "non-empty map",
			value:    map[string]int{"a": 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result := v.IsNil()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValue_IsTrue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		// Nil values
		{
			name:     "nil",
			value:    nil,
			expected: false,
		},
		{
			name:     "nil pointer",
			value:    (*int)(nil),
			expected: false,
		},
		// Boolean values
		{
			name:     "bool true",
			value:    true,
			expected: true,
		},
		{
			name:     "bool false",
			value:    false,
			expected: false,
		},
		{
			name:     "pointer to bool true",
			value:    func() *bool { b := true; return &b }(),
			expected: true,
		},
		{
			name:     "pointer to bool false",
			value:    func() *bool { b := false; return &b }(),
			expected: false,
		},
		// Integer values
		{
			name:     "int zero",
			value:    0,
			expected: false,
		},
		{
			name:     "int positive",
			value:    42,
			expected: true,
		},
		{
			name:     "int negative",
			value:    -42,
			expected: true,
		},
		{
			name:     "pointer to int zero",
			value:    func() *int { i := 0; return &i }(),
			expected: false,
		},
		{
			name:     "pointer to int positive",
			value:    func() *int { i := 42; return &i }(),
			expected: true,
		},
		{
			name:     "int8 zero",
			value:    int8(0),
			expected: false,
		},
		{
			name:     "int8 non-zero",
			value:    int8(42),
			expected: true,
		},
		{
			name:     "int64 zero",
			value:    int64(0),
			expected: false,
		},
		{
			name:     "int64 non-zero",
			value:    int64(42),
			expected: true,
		},
		// Unsigned integer values
		{
			name:     "uint zero",
			value:    uint(0),
			expected: false,
		},
		{
			name:     "uint non-zero",
			value:    uint(42),
			expected: true,
		},
		{
			name:     "uint64 zero",
			value:    uint64(0),
			expected: false,
		},
		{
			name:     "uint64 non-zero",
			value:    uint64(42),
			expected: true,
		},
		// Float values
		{
			name:     "float32 zero",
			value:    float32(0),
			expected: false,
		},
		{
			name:     "float32 non-zero",
			value:    float32(3.14),
			expected: true,
		},
		{
			name:     "float64 zero",
			value:    float64(0),
			expected: false,
		},
		{
			name:     "float64 non-zero",
			value:    float64(3.14),
			expected: true,
		},
		{
			name:     "pointer to float64 zero",
			value:    func() *float64 { f := 0.0; return &f }(),
			expected: false,
		},
		{
			name:     "pointer to float64 non-zero",
			value:    func() *float64 { f := 3.14; return &f }(),
			expected: true,
		},
		// String values
		{
			name:     "empty string",
			value:    "",
			expected: false,
		},
		{
			name:     "non-empty string",
			value:    "hello",
			expected: true,
		},
		{
			name:     "pointer to empty string",
			value:    func() *string { s := ""; return &s }(),
			expected: false,
		},
		{
			name:     "pointer to non-empty string",
			value:    func() *string { s := "hello"; return &s }(),
			expected: true,
		},
		// Slice values
		{
			name:     "nil slice",
			value:    ([]int)(nil),
			expected: false,
		},
		{
			name:     "empty slice",
			value:    []int{},
			expected: false,
		},
		{
			name:     "non-empty slice",
			value:    []int{1, 2, 3},
			expected: true,
		},
		// Map values
		{
			name:     "nil map",
			value:    (map[string]int)(nil),
			expected: false,
		},
		{
			name:     "empty map",
			value:    map[string]int{},
			expected: false,
		},
		{
			name:     "non-empty map",
			value:    map[string]int{"a": 1},
			expected: true,
		},
		// Array values
		{
			name:     "empty array",
			value:    [0]int{},
			expected: false,
		},
		{
			name:     "non-empty array",
			value:    [3]int{1, 2, 3},
			expected: true,
		},
		// Struct values
		{
			name:     "empty struct",
			value:    struct{}{},
			expected: true,
		},
		{
			name:     "non-empty struct",
			value:    struct{ Name string }{Name: "test"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result := v.IsTrue()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValue_String(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "nil",
			value:    nil,
			expected: "",
		},
		{
			name:     "nil pointer",
			value:    (*int)(nil),
			expected: "",
		},
		{
			name:     "int",
			value:    42,
			expected: "42",
		},
		{
			name:     "pointer to int",
			value:    func() *int { i := 42; return &i }(),
			expected: "42",
		},
		{
			name:     "negative int",
			value:    -42,
			expected: "-42",
		},
		{
			name:     "float",
			value:    3.14,
			expected: "3.14",
		},
		{
			name:     "pointer to float",
			value:    func() *float64 { f := 3.14; return &f }(),
			expected: "3.14",
		},
		{
			name:     "string",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "pointer to string",
			value:    func() *string { s := "hello"; return &s }(),
			expected: "hello",
		},
		{
			name:     "empty string",
			value:    "",
			expected: "",
		},
		{
			name:     "bool true",
			value:    true,
			expected: "true",
		},
		{
			name:     "bool false",
			value:    false,
			expected: "false",
		},
		{
			name:     "pointer to bool",
			value:    func() *bool { b := true; return &b }(),
			expected: "true",
		},
		{
			name:     "slice",
			value:    []int{1, 2, 3},
			expected: "[1,2,3]",
		},
		{
			name:     "map",
			value:    map[string]int{"a": 1},
			expected: "{\"a\":1}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result := v.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValue_Int(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expected    int64
		expectError bool
	}{
		{
			name:        "nil",
			value:       nil,
			expected:    0,
			expectError: true,
		},
		{
			name:        "nil pointer",
			value:       (*int)(nil),
			expected:    0,
			expectError: true,
		},
		{
			name:        "int",
			value:       42,
			expected:    42,
			expectError: false,
		},
		{
			name:        "pointer to int",
			value:       func() *int { i := 42; return &i }(),
			expected:    42,
			expectError: false,
		},
		{
			name:        "int8",
			value:       int8(42),
			expected:    42,
			expectError: false,
		},
		{
			name:        "int16",
			value:       int16(42),
			expected:    42,
			expectError: false,
		},
		{
			name:        "int32",
			value:       int32(42),
			expected:    42,
			expectError: false,
		},
		{
			name:        "int64",
			value:       int64(42),
			expected:    42,
			expectError: false,
		},
		{
			name:        "uint",
			value:       uint(42),
			expected:    42,
			expectError: false,
		},
		{
			name:        "uint8",
			value:       uint8(42),
			expected:    42,
			expectError: false,
		},
		{
			name:        "uint16",
			value:       uint16(42),
			expected:    42,
			expectError: false,
		},
		{
			name:        "uint32",
			value:       uint32(42),
			expected:    42,
			expectError: false,
		},
		{
			name:        "uint64",
			value:       uint64(42),
			expected:    42,
			expectError: false,
		},
		{
			name:        "float32",
			value:       float32(42.7),
			expected:    42,
			expectError: false,
		},
		{
			name:        "float64",
			value:       float64(42.7),
			expected:    42,
			expectError: false,
		},
		{
			name:        "pointer to float64",
			value:       func() *float64 { f := 42.7; return &f }(),
			expected:    42,
			expectError: false,
		},
		{
			name:        "bool true",
			value:       true,
			expected:    1,
			expectError: false,
		},
		{
			name:        "bool false",
			value:       false,
			expected:    0,
			expectError: false,
		},
		{
			name:        "pointer to bool true",
			value:       func() *bool { b := true; return &b }(),
			expected:    1,
			expectError: false,
		},
		{
			name:        "string",
			value:       "hello",
			expected:    0,
			expectError: true,
		},
		{
			name:        "negative int",
			value:       -42,
			expected:    -42,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result, err := v.Int()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValue_Float(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expected    float64
		expectError bool
	}{
		{
			name:        "nil",
			value:       nil,
			expected:    0,
			expectError: true,
		},
		{
			name:        "nil pointer",
			value:       (*float64)(nil),
			expected:    0,
			expectError: true,
		},
		{
			name:        "float32",
			value:       float32(3.14),
			expected:    float64(float32(3.14)),
			expectError: false,
		},
		{
			name:        "float64",
			value:       float64(3.14),
			expected:    3.14,
			expectError: false,
		},
		{
			name:        "pointer to float64",
			value:       func() *float64 { f := 3.14; return &f }(),
			expected:    3.14,
			expectError: false,
		},
		{
			name:        "int",
			value:       42,
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "pointer to int",
			value:       func() *int { i := 42; return &i }(),
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "int8",
			value:       int8(42),
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "int16",
			value:       int16(42),
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "int32",
			value:       int32(42),
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "int64",
			value:       int64(42),
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "uint",
			value:       uint(42),
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "uint8",
			value:       uint8(42),
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "uint16",
			value:       uint16(42),
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "uint32",
			value:       uint32(42),
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "uint64",
			value:       uint64(42),
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "string",
			value:       "hello",
			expected:    0,
			expectError: true,
		},
		{
			name:        "negative float",
			value:       -3.14,
			expected:    -3.14,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result, err := v.Float()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValue_Bool(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{
			name:     "nil",
			value:    nil,
			expected: false,
		},
		{
			name:     "bool true",
			value:    true,
			expected: true,
		},
		{
			name:     "bool false",
			value:    false,
			expected: false,
		},
		{
			name:     "pointer to bool true",
			value:    func() *bool { b := true; return &b }(),
			expected: true,
		},
		{
			name:     "pointer to bool false",
			value:    func() *bool { b := false; return &b }(),
			expected: false,
		},
		{
			name:     "int zero",
			value:    0,
			expected: false,
		},
		{
			name:     "int non-zero",
			value:    42,
			expected: true,
		},
		{
			name:     "empty string",
			value:    "",
			expected: false,
		},
		{
			name:     "non-empty string",
			value:    "hello",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result := v.Bool()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValue_Len(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expected    int
		expectError bool
	}{
		{
			name:        "nil",
			value:       nil,
			expected:    0,
			expectError: false,
		},
		{
			name:        "nil slice",
			value:       ([]int)(nil),
			expected:    0,
			expectError: false,
		},
		{
			name:        "empty slice",
			value:       []int{},
			expected:    0,
			expectError: false,
		},
		{
			name:        "non-empty slice",
			value:       []int{1, 2, 3},
			expected:    3,
			expectError: false,
		},
		{
			name:        "pointer to slice",
			value:       func() *[]int { s := []int{1, 2, 3}; return &s }(),
			expected:    3,
			expectError: false,
		},
		{
			name:        "empty string",
			value:       "",
			expected:    0,
			expectError: false,
		},
		{
			name:        "non-empty string",
			value:       "hello",
			expected:    5,
			expectError: false,
		},
		{
			name:        "pointer to string",
			value:       func() *string { s := "hello"; return &s }(),
			expected:    5,
			expectError: false,
		},
		{
			name:        "empty map",
			value:       map[string]int{},
			expected:    0,
			expectError: false,
		},
		{
			name:        "non-empty map",
			value:       map[string]int{"a": 1, "b": 2},
			expected:    2,
			expectError: false,
		},
		{
			name:        "pointer to map",
			value:       func() *map[string]int { m := map[string]int{"a": 1}; return &m }(),
			expected:    1,
			expectError: false,
		},
		{
			name:        "array",
			value:       [3]int{1, 2, 3},
			expected:    3,
			expectError: false,
		},
		{
			name:        "int",
			value:       42,
			expected:    0,
			expectError: true,
		},
		{
			name:        "bool",
			value:       true,
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result, err := v.Len()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValue_Index(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		index       int
		expected    interface{}
		expectError bool
	}{
		{
			name:        "nil",
			value:       nil,
			index:       0,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "slice - valid index",
			value:       []int{10, 20, 30},
			index:       1,
			expected:    20,
			expectError: false,
		},
		{
			name:        "slice - first index",
			value:       []int{10, 20, 30},
			index:       0,
			expected:    10,
			expectError: false,
		},
		{
			name:        "slice - last index",
			value:       []int{10, 20, 30},
			index:       2,
			expected:    30,
			expectError: false,
		},
		{
			name:        "slice - negative index",
			value:       []int{10, 20, 30},
			index:       -1,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "slice - out of range index",
			value:       []int{10, 20, 30},
			index:       3,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "pointer to slice",
			value:       func() *[]int { s := []int{10, 20, 30}; return &s }(),
			index:       1,
			expected:    20,
			expectError: false,
		},
		{
			name:        "array - valid index",
			value:       [3]int{10, 20, 30},
			index:       1,
			expected:    20,
			expectError: false,
		},
		{
			name:        "string - valid index",
			value:       "hello",
			index:       1,
			expected:    "e",
			expectError: false,
		},
		{
			name:        "string - first index",
			value:       "hello",
			index:       0,
			expected:    "h",
			expectError: false,
		},
		{
			name:        "string - last index",
			value:       "hello",
			index:       4,
			expected:    "o",
			expectError: false,
		},
		{
			name:        "string - out of range",
			value:       "hello",
			index:       5,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "pointer to string",
			value:       func() *string { s := "hello"; return &s }(),
			index:       1,
			expected:    "e",
			expectError: false,
		},
		{
			name:        "map - not indexable",
			value:       map[string]int{"a": 1},
			index:       0,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "int - not indexable",
			value:       42,
			index:       0,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result, err := v.Index(tt.index)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result.Interface())
			}
		})
	}
}

func TestValue_Key(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		key         interface{}
		expected    interface{}
		expectError bool
	}{
		{
			name:        "nil",
			value:       nil,
			key:         "key",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "map with string key - found",
			value:       map[string]int{"a": 1, "b": 2},
			key:         "a",
			expected:    1,
			expectError: false,
		},
		{
			name:        "map with string key - not found",
			value:       map[string]int{"a": 1, "b": 2},
			key:         "c",
			expected:    nil,
			expectError: false,
		},
		{
			name:        "pointer to map",
			value:       func() *map[string]int { m := map[string]int{"a": 1}; return &m }(),
			key:         "a",
			expected:    1,
			expectError: false,
		},
		{
			name:        "map with int key - found",
			value:       map[int]string{1: "one", 2: "two"},
			key:         1,
			expected:    "one",
			expectError: false,
		},
		{
			name:        "map with int key - not found",
			value:       map[int]string{1: "one", 2: "two"},
			key:         3,
			expected:    nil,
			expectError: false,
		},
		{
			name:        "slice - not a map",
			value:       []int{1, 2, 3},
			key:         0,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "string - not a map",
			value:       "hello",
			key:         0,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "int - not a map",
			value:       42,
			key:         0,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result, err := v.Key(tt.key)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result.Interface())
			}
		})
	}
}

func TestValue_Field(t *testing.T) {
	type TestStruct struct {
		Name  string
		Age   int
		Email string
	}

	tests := []struct {
		name        string
		value       interface{}
		field       string
		expected    interface{}
		expectError bool
	}{
		{
			name:        "nil",
			value:       nil,
			field:       "Name",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "struct - field exists",
			value:       TestStruct{Name: "Alice", Age: 30, Email: "alice@example.com"},
			field:       "Name",
			expected:    "Alice",
			expectError: false,
		},
		{
			name:        "struct - field not exists",
			value:       TestStruct{Name: "Alice", Age: 30, Email: "alice@example.com"},
			field:       "Phone",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "pointer to struct",
			value:       &TestStruct{Name: "Bob", Age: 25, Email: "bob@example.com"},
			field:       "Age",
			expected:    25,
			expectError: false,
		},
		{
			name:        "map - field as string key exists",
			value:       map[string]interface{}{"Name": "Charlie", "Age": 35},
			field:       "Name",
			expected:    "Charlie",
			expectError: false,
		},
		{
			name:        "map - field as string key not exists",
			value:       map[string]interface{}{"Name": "Charlie", "Age": 35},
			field:       "Email",
			expected:    nil,
			expectError: false,
		},
		{
			name:        "pointer to map",
			value:       func() *map[string]string { m := map[string]string{"Name": "David"}; return &m }(),
			field:       "Name",
			expected:    "David",
			expectError: false,
		},
		{
			name:        "slice - not a struct or map",
			value:       []int{1, 2, 3},
			field:       "Name",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "int - not a struct or map",
			value:       42,
			field:       "Name",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result, err := v.Field(tt.field)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result.Interface())
			}
		})
	}
}

func TestValue_Iterate(t *testing.T) {
	tests := []struct {
		name          string
		value         interface{}
		expectedKeys  []interface{}
		expectedVals  []interface{}
		expectedCount int
		expectError   bool
	}{
		{
			name:          "nil",
			value:         nil,
			expectedKeys:  []interface{}{},
			expectedVals:  []interface{}{},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "empty slice",
			value:         []int{},
			expectedKeys:  []interface{}{},
			expectedVals:  []interface{}{},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "slice with elements",
			value:         []int{10, 20, 30},
			expectedKeys:  []interface{}{0, 1, 2},
			expectedVals:  []interface{}{10, 20, 30},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "pointer to slice",
			value:         func() *[]int { s := []int{10, 20}; return &s }(),
			expectedKeys:  []interface{}{0, 1},
			expectedVals:  []interface{}{10, 20},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "array",
			value:         [3]string{"a", "b", "c"},
			expectedKeys:  []interface{}{0, 1, 2},
			expectedVals:  []interface{}{"a", "b", "c"},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "map",
			value:         map[string]int{"a": 1, "b": 2},
			expectedKeys:  []interface{}{"a", "b"},
			expectedVals:  []interface{}{1, 2},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "pointer to map",
			value:         func() *map[string]int { m := map[string]int{"x": 10}; return &m }(),
			expectedKeys:  []interface{}{"x"},
			expectedVals:  []interface{}{10},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "int - not iterable",
			value:         42,
			expectedKeys:  nil,
			expectedVals:  nil,
			expectedCount: 0,
			expectError:   true,
		},
		{
			name:          "string - iterable",
			value:         "hello",
			expectedKeys:  []interface{}{0, 1, 2, 3, 4},
			expectedVals:  []interface{}{"h", "e", "l", "l", "o"},
			expectedCount: 5,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			var keys []interface{}
			var vals []interface{}
			var iterCount int

			err := v.Iterate(func(_ int, count int, key, value *Value) bool {
				iterCount = count
				keys = append(keys, key.Interface())
				vals = append(vals, value.Interface())
				return true
			})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, iterCount)
				// For maps, order is not guaranteed, so we check if all elements are present
				if tt.expectedCount > 0 {
					assert.ElementsMatch(t, tt.expectedKeys, keys)
					assert.ElementsMatch(t, tt.expectedVals, vals)
				} else {
					// For empty collections, just check length
					assert.Equal(t, len(tt.expectedKeys), len(keys))
					assert.Equal(t, len(tt.expectedVals), len(vals))
				}
			}
		})
	}
}

func TestValue_Iterate_EarlyExit(t *testing.T) {
	v := NewValue([]int{1, 2, 3, 4, 5})
	var collected []int

	err := v.Iterate(func(idx, _ int, _ *Value, value *Value) bool {
		collected = append(collected, value.Interface().(int))
		// Stop after collecting 3 elements
		return idx < 2
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, collected)
}

func TestValue_Compare(t *testing.T) {
	tests := []struct {
		name     string
		value1   interface{}
		value2   interface{}
		expected int
	}{
		{
			name:     "both nil",
			value1:   nil,
			value2:   nil,
			expected: 0,
		},
		{
			name:     "first nil",
			value1:   nil,
			value2:   42,
			expected: -1,
		},
		{
			name:     "second nil",
			value1:   42,
			value2:   nil,
			expected: 1,
		},
		{
			name:     "int equal",
			value1:   42,
			value2:   42,
			expected: 0,
		},
		{
			name:     "int less than",
			value1:   10,
			value2:   20,
			expected: -1,
		},
		{
			name:     "int greater than",
			value1:   20,
			value2:   10,
			expected: 1,
		},
		{
			name:     "pointer to int less than",
			value1:   func() *int { i := 10; return &i }(),
			value2:   func() *int { i := 20; return &i }(),
			expected: -1,
		},
		{
			name:     "float equal",
			value1:   3.14,
			value2:   3.14,
			expected: 0,
		},
		{
			name:     "float less than",
			value1:   3.14,
			value2:   6.28,
			expected: -1,
		},
		{
			name:     "float greater than",
			value1:   6.28,
			value2:   3.14,
			expected: 1,
		},
		{
			name:     "pointer to float less than",
			value1:   func() *float64 { f := 3.14; return &f }(),
			value2:   func() *float64 { f := 6.28; return &f }(),
			expected: -1,
		},
		{
			name:     "mixed int and float equal",
			value1:   42,
			value2:   42.0,
			expected: 0,
		},
		{
			name:     "mixed int and float less than",
			value1:   10,
			value2:   20.5,
			expected: -1,
		},
		{
			name:     "string equal",
			value1:   "hello",
			value2:   "hello",
			expected: 0,
		},
		{
			name:     "string less than",
			value1:   "abc",
			value2:   "xyz",
			expected: -1,
		},
		{
			name:     "string greater than",
			value1:   "xyz",
			value2:   "abc",
			expected: 1,
		},
		{
			name:     "pointer to string less than",
			value1:   func() *string { s := "abc"; return &s }(),
			value2:   func() *string { s := "xyz"; return &s }(),
			expected: -1,
		},
		{
			name:     "negative numbers",
			value1:   -10,
			value2:   -5,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1 := NewValue(tt.value1)
			v2 := NewValue(tt.value2)
			result, err := v1.Compare(v2)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValue_Equals(t *testing.T) {
	tests := []struct {
		name     string
		value1   interface{}
		value2   interface{}
		expected bool
	}{
		{
			name:     "both nil",
			value1:   nil,
			value2:   nil,
			expected: true,
		},
		{
			name:     "first nil",
			value1:   nil,
			value2:   42,
			expected: false,
		},
		{
			name:     "second nil",
			value1:   42,
			value2:   nil,
			expected: false,
		},
		{
			name:     "int equal",
			value1:   42,
			value2:   42,
			expected: true,
		},
		{
			name:     "int not equal",
			value1:   42,
			value2:   43,
			expected: false,
		},
		{
			name:     "pointer to int equal",
			value1:   func() *int { i := 42; return &i }(),
			value2:   func() *int { i := 42; return &i }(),
			expected: true,
		},
		{
			name:     "pointer to int not equal",
			value1:   func() *int { i := 42; return &i }(),
			value2:   func() *int { i := 43; return &i }(),
			expected: false,
		},
		{
			name:     "float equal",
			value1:   3.14,
			value2:   3.14,
			expected: true,
		},
		{
			name:     "float not equal",
			value1:   3.14,
			value2:   6.28,
			expected: false,
		},
		{
			name:     "pointer to float equal",
			value1:   func() *float64 { f := 3.14; return &f }(),
			value2:   func() *float64 { f := 3.14; return &f }(),
			expected: true,
		},
		{
			name:     "string equal",
			value1:   "hello",
			value2:   "hello",
			expected: true,
		},
		{
			name:     "string not equal",
			value1:   "hello",
			value2:   "world",
			expected: false,
		},
		{
			name:     "pointer to string equal",
			value1:   func() *string { s := "hello"; return &s }(),
			value2:   func() *string { s := "hello"; return &s }(),
			expected: true,
		},
		{
			name:     "bool equal",
			value1:   true,
			value2:   true,
			expected: true,
		},
		{
			name:     "bool not equal",
			value1:   true,
			value2:   false,
			expected: false,
		},
		{
			name:     "pointer to bool equal",
			value1:   func() *bool { b := true; return &b }(),
			value2:   func() *bool { b := true; return &b }(),
			expected: true,
		},
		{
			name:     "slice equal",
			value1:   []int{1, 2, 3},
			value2:   []int{1, 2, 3},
			expected: true,
		},
		{
			name:     "slice not equal",
			value1:   []int{1, 2, 3},
			value2:   []int{1, 2, 4},
			expected: false,
		},
		{
			name:     "map equal",
			value1:   map[string]int{"a": 1, "b": 2},
			value2:   map[string]int{"a": 1, "b": 2},
			expected: true,
		},
		{
			name:     "map not equal",
			value1:   map[string]int{"a": 1, "b": 2},
			value2:   map[string]int{"a": 1, "b": 3},
			expected: false,
		},
		{
			name:     "different types",
			value1:   42,
			value2:   "42",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1 := NewValue(tt.value1)
			v2 := NewValue(tt.value2)
			result := v1.Equals(v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValue_Interface(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{
			name:     "nil",
			value:    nil,
			expected: nil,
		},
		{
			name:     "int",
			value:    42,
			expected: 42,
		},
		{
			name:     "pointer to int",
			value:    func() *int { i := 42; return &i }(),
			expected: func() *int { i := 42; return &i }(),
		},
		{
			name:     "string",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "bool",
			value:    true,
			expected: true,
		},
		{
			name:     "slice",
			value:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "map",
			value:    map[string]int{"a": 1},
			expected: map[string]int{"a": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			result := v.Interface()
			// Use DeepEqual for complex types
			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestValue_NewValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "nil",
			input: nil,
		},
		{
			name:  "int",
			input: 42,
		},
		{
			name:  "string",
			input: "hello",
		},
		{
			name:  "bool",
			input: true,
		},
		{
			name:  "slice",
			input: []int{1, 2, 3},
		},
		{
			name:  "map",
			input: map[string]int{"a": 1},
		},
		{
			name:  "pointer",
			input: func() *int { i := 42; return &i }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.input)
			assert.NotNil(t, v)
			// Check that Interface() returns the original value
			if !reflect.DeepEqual(tt.input, v.Interface()) {
				t.Errorf("NewValue did not preserve input value")
			}
		})
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{name: "whole number", value: 42.0, expected: "42"},
		{name: "zero", value: 0.0, expected: "0"},
		{name: "negative whole", value: -10.0, expected: "-10"},
		{name: "decimal", value: 3.14, expected: "3.14"},
		{name: "negative decimal", value: -2.5, expected: "-2.5"},
		{name: "small decimal", value: 0.001, expected: "0.001"},
		{name: "large whole", value: 1000000.0, expected: "1000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			assert.Equal(t, tt.expected, v.String())
		})
	}
}

func TestValue_String_Uint(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{name: "uint", value: uint(100), expected: "100"},
		{name: "uint8", value: uint8(255), expected: "255"},
		{name: "uint16", value: uint16(1000), expected: "1000"},
		{name: "uint32", value: uint32(70000), expected: "70000"},
		{name: "uint64", value: uint64(99999), expected: "99999"},
		{name: "uint zero", value: uint(0), expected: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.value)
			assert.Equal(t, tt.expected, v.String())
		})
	}
}

func TestValue_Field_JSONTag(t *testing.T) {
	type Tagged struct {
		FullName string `json:"name"`
		Age      int    `json:"age,omitempty"`
		Hidden   string `json:"-"`
		NoTag    string
	}

	val := Tagged{
		FullName: "Alice",
		Age:      30,
		Hidden:   "secret",
		NoTag:    "visible",
	}

	tests := []struct {
		name        string
		field       string
		expected    interface{}
		expectError bool
	}{
		{
			name:     "json tag name",
			field:    "name",
			expected: "Alice",
		},
		{
			name:     "json tag with omitempty",
			field:    "age",
			expected: 30,
		},
		{
			name:        "hidden field via json tag dash",
			field:       "-",
			expectError: true,
		},
		{
			name:     "field by exported name",
			field:    "NoTag",
			expected: "visible",
		},
		{
			name:     "field by exported name FullName",
			field:    "FullName",
			expected: "Alice",
		},
		{
			name:        "nonexistent field",
			field:       "missing",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(val)
			result, err := v.Field(tt.field)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result.Interface())
			}
		})
	}
}

func TestValue_String_NestedSlice(t *testing.T) {
	v := NewValue([][]int{{1, 2}, {3, 4}})
	assert.Equal(t, "[[1,2],[3,4]]", v.String())
}

func TestValue_String_EmptySlice(t *testing.T) {
	v := NewValue([]int{})
	assert.Equal(t, "[]", v.String())
}

func TestValue_String_BoolSlice(t *testing.T) {
	v := NewValue([]bool{true, false, true})
	assert.Equal(t, "[true,false,true]", v.String())
}

func TestValue_Equals_CrossNumericTypes(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "int and float64 equal",
			a:        42,
			b:        float64(42),
			expected: true,
		},
		{
			name:     "int and float64 not equal",
			a:        42,
			b:        float64(42.5),
			expected: false,
		},
		{
			name:     "int64 and uint equal",
			a:        int64(100),
			b:        uint(100),
			expected: true,
		},
		{
			name:     "int32 and float32 equal",
			a:        int32(7),
			b:        float32(7),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va := NewValue(tt.a)
			vb := NewValue(tt.b)
			assert.Equal(t, tt.expected, va.Equals(vb))
		})
	}
}
