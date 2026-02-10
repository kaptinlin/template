package template

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-json-experiment/json"
)

// Value wraps any Go value for template execution.
// It provides methods for type checking, conversion, and operations.
type Value struct {
	val interface{}
}

// NewValue creates a new Value wrapper.
func NewValue(v interface{}) *Value {
	return &Value{val: v}
}

// Interface returns the underlying Go value.
func (v *Value) Interface() interface{} {
	return v.val
}

// IsNil checks if the value is nil.
func (v *Value) IsNil() bool {
	return !v.getResolvedValue().IsValid()
}

// getResolvedValue recursively dereferences pointers and interfaces to get the underlying value.
// This method is called by all Value methods to handle pointer types correctly.
func (v *Value) getResolvedValue() reflect.Value {
	if v.val == nil {
		return reflect.Value{}
	}

	rv := reflect.ValueOf(v.val)
	// Unwrap pointers and interfaces to get to the underlying value
	for rv.IsValid() && (rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface) {
		rv = rv.Elem()
	}
	return rv
}

// IsTrue checks if the value is considered true in template context.
// False values: nil, false, 0, "", empty slice/map
func (v *Value) IsTrue() bool {
	rv := v.getResolvedValue()
	if !rv.IsValid() {
		return false
	}

	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0

	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0

	case reflect.String:
		return rv.String() != ""

	case reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() > 0

	case reflect.Invalid:
		return false

	case reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Ptr, reflect.Struct, reflect.UnsafePointer:
		// Other non-nil values are considered truthy.
		return true

	default:
		// Other types (struct, etc.) are considered true if not nil
		return true
	}
}

// formatFloat renders a float as a string.
// Whole-number floats are rendered without a fractional part (e.g. "3");
// other values use the shortest decimal representation.
func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return strconv.FormatFloat(f, 'f', 0, 64)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// String returns the string representation of the value.
func (v *Value) String() string {
	rv := v.getResolvedValue()
	if !rv.IsValid() {
		return ""
	}

	// Handle special types before kind-based switch
	// Check for time.Time first
	if t, ok := rv.Interface().(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}

	// Check for fmt.Stringer interface
	if s, ok := rv.Interface().(fmt.Stringer); ok {
		return s.String()
	}

	// Handle string conversion based on type
	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return formatFloat(rv.Float())
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	case reflect.Slice, reflect.Array:
		return v.formatSlice(rv)
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Map, reflect.Ptr, reflect.Struct, reflect.UnsafePointer:
		// For complex types (map, struct, etc.), use JSON serialization
		// This ensures consistent output format matching the old implementation
		jsonBytes, err := json.Marshal(rv.Interface(), json.Deterministic(true))
		if err != nil {
			// Fallback to fmt.Sprint if JSON marshaling fails
			return fmt.Sprint(rv.Interface())
		}
		return string(jsonBytes)
	}

	return ""
}

// formatSlice formats an array/slice as [item1,item2,item3] with comma separation.
func (v *Value) formatSlice(rv reflect.Value) string {
	length := rv.Len()
	if length == 0 {
		return "[]"
	}

	var builder strings.Builder
	builder.Grow(length * 20) // Estimate ~20 chars per item
	builder.WriteByte('[')

	for i := 0; i < length; i++ {
		if i > 0 {
			builder.WriteByte(',')
		}
		item := rv.Index(i).Interface()

		// Format the item based on its type
		itemRv := reflect.ValueOf(item)
		if !itemRv.IsValid() {
			builder.WriteString("null")
			continue
		}

		var str string
		switch itemRv.Kind() {
		case reflect.String:
			str = itemRv.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			str = strconv.FormatInt(itemRv.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			str = strconv.FormatUint(itemRv.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			str = formatFloat(itemRv.Float())
		case reflect.Bool:
			str = strconv.FormatBool(itemRv.Bool())
		case reflect.Slice, reflect.Array:
			// Recursively format nested arrays
			itemValue := NewValue(item)
			str = itemValue.formatSlice(itemRv)
		case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
			reflect.Interface, reflect.Map, reflect.Ptr, reflect.Struct, reflect.UnsafePointer:
			// For complex types (map, struct, etc.), use JSON serialization
			jsonBytes, err := json.Marshal(item, json.Deterministic(true))
			if err != nil {
				str = fmt.Sprint(item)
			} else {
				str = string(jsonBytes)
			}
		}
		builder.WriteString(str)
	}

	builder.WriteByte(']')
	return builder.String()
}

// Int returns the integer value, with conversion if possible.
func (v *Value) Int() (int64, error) {
	rv := v.getResolvedValue()
	if !rv.IsValid() {
		return 0, ErrCannotConvertNilToInt
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u := rv.Uint()
		if u > math.MaxInt64 {
			return 0, ErrIntegerOverflow
		}
		return int64(u), nil

	case reflect.Float32, reflect.Float64:
		return int64(rv.Float()), nil

	case reflect.Bool:
		if rv.Bool() {
			return 1, nil
		}
		return 0, nil

	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.String, reflect.Struct,
		reflect.UnsafePointer:
		return 0, fmt.Errorf("%w: %T", ErrCannotConvertToInt, v.val)
	}

	return 0, fmt.Errorf("%w: %T", ErrCannotConvertToInt, v.val)
}

// Float returns the float value, with conversion if possible.
func (v *Value) Float() (float64, error) {
	rv := v.getResolvedValue()
	if !rv.IsValid() {
		return 0, ErrCannotConvertNilToFloat
	}

	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(rv.Uint()), nil

	case reflect.Invalid, reflect.Bool, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.String,
		reflect.Struct, reflect.UnsafePointer:
		return 0, fmt.Errorf("%w: %T", ErrCannotConvertToFloat, v.val)
	}

	return 0, fmt.Errorf("%w: %T", ErrCannotConvertToFloat, v.val)
}

// Bool returns the boolean value.
func (v *Value) Bool() bool {
	return v.IsTrue()
}

// Len returns the length of the value (for strings, slices, maps, arrays).
func (v *Value) Len() (int, error) {
	rv := v.getResolvedValue()
	if !rv.IsValid() {
		return 0, nil
	}

	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len(), nil

	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Ptr, reflect.Struct, reflect.UnsafePointer:
		return 0, fmt.Errorf("%w: %T", ErrTypeHasNoLength, v.val)
	}

	return 0, fmt.Errorf("%w: %T", ErrTypeHasNoLength, v.val)
}

// Index returns the value at the given index (for slices, arrays, strings).
func (v *Value) Index(i int) (*Value, error) {
	rv := v.getResolvedValue()
	if !rv.IsValid() {
		return nil, ErrCannotIndexNil
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if i < 0 || i >= rv.Len() {
			return nil, fmt.Errorf("%w: %d", ErrIndexOutOfRange, i)
		}
		return NewValue(rv.Index(i).Interface()), nil

	case reflect.String:
		s := rv.String()
		if i < 0 || i >= len(s) {
			return nil, fmt.Errorf("%w: %d", ErrIndexOutOfRange, i)
		}
		return NewValue(string(s[i])), nil

	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Map, reflect.Ptr, reflect.Struct, reflect.UnsafePointer:
		return nil, fmt.Errorf("%w: %T", ErrTypeNotIndexable, v.val)
	}

	return nil, fmt.Errorf("%w: %T", ErrTypeNotIndexable, v.val)
}

// Key returns the value for the given key (for maps).
func (v *Value) Key(key interface{}) (*Value, error) {
	rv := v.getResolvedValue()
	if !rv.IsValid() {
		return nil, ErrCannotGetKeyFromNil
	}

	if rv.Kind() != reflect.Map {
		return nil, fmt.Errorf("%w: %T", ErrTypeNotMap, v.val)
	}

	keyVal := reflect.ValueOf(key)
	result := rv.MapIndex(keyVal)

	if !result.IsValid() {
		return NewValue(nil), nil
	}

	return NewValue(result.Interface()), nil
}

// Field returns the value of a struct field or map key by name.
func (v *Value) Field(name string) (*Value, error) {
	rv := v.getResolvedValue()
	if !rv.IsValid() {
		return nil, ErrCannotGetFieldFromNil
	}

	switch rv.Kind() {
	case reflect.Struct:
		// Try to find field by JSON tag first, then by field name
		field, found := findStructField(rv, name)
		if !found {
			return nil, fmt.Errorf("%w: %q", ErrStructHasNoField, name)
		}
		return NewValue(field.Interface()), nil

	case reflect.Map:
		// For map, treat field name as string key
		keyVal := reflect.ValueOf(name)
		result := rv.MapIndex(keyVal)
		if !result.IsValid() {
			return NewValue(nil), nil
		}
		return NewValue(result.Interface()), nil

	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Ptr, reflect.Slice, reflect.String, reflect.UnsafePointer:
		return nil, fmt.Errorf("%w: %T %q", ErrTypeHasNoField, v.val, name)
	}

	return nil, fmt.Errorf("%w: %T %q", ErrTypeHasNoField, v.val, name)
}

// findStructField finds a struct field by JSON tag or field name.
// It searches in the following order:
//  1. Field with matching JSON tag (supports json:"name" or json:"name,omitempty")
//  2. Field with matching exported name (case-sensitive)
//
// This allows templates to use lowercase names (matching JSON tags) while
// maintaining compatibility with direct field access.
func findStructField(rv reflect.Value, name string) (reflect.Value, bool) {
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	typ := rv.Type()
	numFields := rv.NumField()

	// First pass: try to find by JSON tag
	for i := 0; i < numFields; i++ {
		field := typ.Field(i)

		// Check JSON tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			// Extract tag name (before comma if any)
			// e.g., "name,omitempty" -> "name"
			tagName := jsonTag
			for j, ch := range jsonTag {
				if ch == ',' {
					tagName = jsonTag[:j]
					break
				}
			}

			// Skip if tag is "-" (explicitly ignored)
			if tagName == "-" {
				continue
			}

			if tagName == name {
				return rv.Field(i), true
			}
		}
	}

	// Second pass: try to find by exact field name
	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		if field.Name == name {
			return rv.Field(i), true
		}
	}

	return reflect.Value{}, false
}

// Iterate iterates over a collection (slice, array, map).
// The callback receives the index/key and value for each element.
// Returns an error if the value is not iterable.
func (v *Value) Iterate(fn func(idx, count int, key, value *Value) bool) error {
	rv := v.getResolvedValue()
	if !rv.IsValid() {
		return nil // Empty iteration for nil
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		count := rv.Len()
		for i := 0; i < count; i++ {
			key := NewValue(i)
			val := NewValue(rv.Index(i).Interface())
			if !fn(i, count, key, val) {
				break
			}
		}
		return nil

	case reflect.Map:
		keys := sortedKeys(rv.MapKeys())
		// Sort keys to ensure consistent iteration order
		sort.Sort(keys)
		count := len(keys)
		for i, k := range keys {
			key := NewValue(k.Interface())
			val := NewValue(rv.MapIndex(k).Interface())
			if !fn(i, count, key, val) {
				break
			}
		}
		return nil

	case reflect.String:
		// Convert string to rune slice for proper Unicode character iteration
		rs := []rune(rv.String())
		count := len(rs)
		for i := 0; i < count; i++ {
			key := NewValue(i)
			val := NewValue(string(rs[i]))
			if !fn(i, count, key, val) {
				break
			}
		}
		return nil

	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Ptr, reflect.Struct, reflect.UnsafePointer:
		return fmt.Errorf("%w: %T", ErrTypeNotIterable, v.val)
	}

	return fmt.Errorf("%w: %T", ErrTypeNotIterable, v.val)
}

// Compare compares this value with another value.
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
func (v *Value) Compare(other *Value) (int, error) {
	// Handle nil cases
	if v.IsNil() && other.IsNil() {
		return 0, nil
	}
	if v.IsNil() {
		return -1, nil
	}
	if other.IsNil() {
		return 1, nil
	}

	// Try numeric comparison - Float() will use getResolvedValue() internally
	vf, vErr := v.Float()
	of, oErr := other.Float()
	if vErr == nil && oErr == nil {
		if vf < of {
			return -1, nil
		} else if vf > of {
			return 1, nil
		}
		return 0, nil
	}

	// Try string comparison
	vs := v.String()
	otherStr := other.String()
	if vs < otherStr {
		return -1, nil
	} else if vs > otherStr {
		return 1, nil
	}
	return 0, nil
}

// Equals checks if this value equals another value.
func (v *Value) Equals(other *Value) bool {
	if v.IsNil() && other.IsNil() {
		return true
	}
	if v.IsNil() || other.IsNil() {
		return false
	}

	// Try numeric comparison first - this handles int vs float64 comparisons
	vf, vErr := v.Float()
	of, oErr := other.Float()
	if vErr == nil && oErr == nil {
		// Both are numeric, compare as floats
		return vf == of
	}

	// Handle string comparison (including string alias types)
	// For example, type Department string should equal to "Engineering"
	vRv := v.getResolvedValue()
	oRv := other.getResolvedValue()

	if vRv.Kind() == reflect.String && oRv.Kind() == reflect.String {
		return vRv.String() == oRv.String()
	}

	// Fall back to reflect.DeepEqual for non-numeric types
	return reflect.DeepEqual(v.val, other.val)
}

// sortedKeys is a helper type for sorting map keys
type sortedKeys []reflect.Value

func (sk sortedKeys) Len() int {
	return len(sk)
}

func (sk sortedKeys) Less(i, j int) bool {
	vi := NewValue(sk[i].Interface())
	vj := NewValue(sk[j].Interface())

	// Try numeric comparison first
	if fi, erri := vi.Float(); erri == nil {
		if fj, errj := vj.Float(); errj == nil {
			return fi < fj
		}
	}

	// Fall back to string comparison
	return vi.String() < vj.String()
}

func (sk sortedKeys) Swap(i, j int) {
	sk[i], sk[j] = sk[j], sk[i]
}
