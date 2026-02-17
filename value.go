package template

import (
	"cmp"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-json-experiment/json"
)

// Value wraps a Go value for template execution, providing
// type checking, conversion, and comparison operations.
type Value struct {
	val any
}

// NewValue creates a Value wrapping v.
func NewValue(v any) *Value {
	return &Value{val: v}
}

// Interface returns the underlying Go value.
func (v *Value) Interface() any {
	return v.val
}

// IsNil reports whether the value is nil.
func (v *Value) IsNil() bool {
	return !v.resolved().IsValid()
}

// resolved dereferences pointers and interfaces to get the underlying value.
func (v *Value) resolved() reflect.Value {
	if v.val == nil {
		return reflect.Value{}
	}
	rv := reflect.ValueOf(v.val)
	for rv.IsValid() && (rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface) {
		rv = rv.Elem()
	}
	return rv
}

// IsTrue reports whether the value is truthy in a template context.
// False values: nil, false, 0, "", empty slice/map/array.
func (v *Value) IsTrue() bool {
	rv := v.resolved()
	if !rv.IsValid() {
		return false
	}
	switch rv.Kind() { //nolint:exhaustive
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
	default:
		return true
	}
}

// formatFloat renders a float as a string.
// Whole-number floats omit the fractional part (e.g. "3");
// other values use the shortest decimal representation.
func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return strconv.FormatFloat(f, 'f', 0, 64)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// String returns the string representation of the value.
func (v *Value) String() string {
	rv := v.resolved()
	if !rv.IsValid() {
		return ""
	}

	// Handle special types before kind-based switch.
	if t, ok := rv.Interface().(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}
	if s, ok := rv.Interface().(fmt.Stringer); ok {
		return s.String()
	}

	switch rv.Kind() { //nolint:exhaustive
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
		return formatSlice(rv)
	default:
		b, err := json.Marshal(rv.Interface(), json.Deterministic(true))
		if err != nil {
			return fmt.Sprint(rv.Interface())
		}
		return string(b)
	}
}

// formatSlice formats a slice or array as [item1,item2,item3].
func formatSlice(rv reflect.Value) string {
	n := rv.Len()
	if n == 0 {
		return "[]"
	}

	var b strings.Builder
	b.Grow(n * 20)
	b.WriteByte('[')
	for i := range n {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(formatSliceItem(rv.Index(i)))
	}
	b.WriteByte(']')
	return b.String()
}

// formatSliceItem formats a single slice/array element.
// Primitive types are rendered directly; pointers and complex types
// use JSON serialization to preserve quoting and null semantics.
func formatSliceItem(rv reflect.Value) string {
	if !rv.IsValid() {
		return "null"
	}
	switch rv.Kind() { //nolint:exhaustive
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
		return formatSlice(rv)
	default:
		data, err := json.Marshal(rv.Interface(), json.Deterministic(true))
		if err != nil {
			return fmt.Sprint(rv.Interface())
		}
		return string(data)
	}
}

// Int returns the value as int64, converting if possible.
func (v *Value) Int() (int64, error) {
	rv := v.resolved()
	if !rv.IsValid() {
		return 0, ErrCannotConvertNilToInt
	}
	switch rv.Kind() { //nolint:exhaustive
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
	default:
		return 0, fmt.Errorf("%w: %T", ErrCannotConvertToInt, v.val)
	}
}

// Float returns the value as float64, converting if possible.
func (v *Value) Float() (float64, error) {
	rv := v.resolved()
	if !rv.IsValid() {
		return 0, ErrCannotConvertNilToFloat
	}
	switch rv.Kind() { //nolint:exhaustive
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(rv.Uint()), nil
	default:
		return 0, fmt.Errorf("%w: %T", ErrCannotConvertToFloat, v.val)
	}
}

// Bool reports whether the value is truthy.
func (v *Value) Bool() bool {
	return v.IsTrue()
}

// Len returns the length of the value (string, slice, map, or array).
func (v *Value) Len() (int, error) {
	rv := v.resolved()
	if !rv.IsValid() {
		return 0, nil
	}
	switch rv.Kind() { //nolint:exhaustive
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len(), nil
	default:
		return 0, fmt.Errorf("%w: %T", ErrTypeHasNoLength, v.val)
	}
}

// Index returns the element at index i (for slices, arrays, strings).
func (v *Value) Index(i int) (*Value, error) {
	rv := v.resolved()
	if !rv.IsValid() {
		return nil, ErrCannotIndexNil
	}
	switch rv.Kind() { //nolint:exhaustive
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
	default:
		return nil, fmt.Errorf("%w: %T", ErrTypeNotIndexable, v.val)
	}
}

// Key returns the map value for the given key.
func (v *Value) Key(key any) (*Value, error) {
	rv := v.resolved()
	if !rv.IsValid() {
		return nil, ErrCannotGetKeyFromNil
	}
	if rv.Kind() != reflect.Map {
		return nil, fmt.Errorf("%w: %T", ErrTypeNotMap, v.val)
	}
	result := rv.MapIndex(reflect.ValueOf(key))
	if !result.IsValid() {
		return NewValue(nil), nil
	}
	return NewValue(result.Interface()), nil
}

// Field returns the value of a struct field or map key by name.
// For structs, it searches by JSON tag first, then by exported field name.
func (v *Value) Field(name string) (*Value, error) {
	rv := v.resolved()
	if !rv.IsValid() {
		return nil, ErrCannotGetFieldFromNil
	}
	switch rv.Kind() { //nolint:exhaustive
	case reflect.Struct:
		field, found := findStructField(rv, name)
		if !found {
			return nil, fmt.Errorf("%w: %q", ErrStructHasNoField, name)
		}
		return NewValue(field.Interface()), nil
	case reflect.Map:
		result := rv.MapIndex(reflect.ValueOf(name))
		if !result.IsValid() {
			return NewValue(nil), nil
		}
		return NewValue(result.Interface()), nil
	default:
		return nil, fmt.Errorf("%w: %T %q", ErrTypeHasNoField, v.val, name)
	}
}

// findStructField finds a struct field by JSON tag or exported name.
// Search order: (1) matching JSON tag, (2) matching exported field name.
func findStructField(rv reflect.Value, name string) (reflect.Value, bool) {
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	typ := rv.Type()
	n := rv.NumField()

	// First pass: match by JSON tag.
	for i := range n {
		tag := typ.Field(i).Tag.Get("json")
		if tag == "" {
			continue
		}
		tagName, _, _ := strings.Cut(tag, ",")
		if tagName == "-" {
			continue
		}
		if tagName == name {
			return rv.Field(i), true
		}
	}

	// Second pass: match by exported field name.
	for i := range n {
		if typ.Field(i).Name == name {
			return rv.Field(i), true
		}
	}

	return reflect.Value{}, false
}

// Iterate calls fn for each element in a collection (slice, array, map, string).
// fn receives the iteration index, total count, key, and value.
// Returning false from fn stops iteration early.
func (v *Value) Iterate(fn func(idx, count int, key, val *Value) bool) error {
	rv := v.resolved()
	if !rv.IsValid() {
		return nil
	}
	switch rv.Kind() { //nolint:exhaustive
	case reflect.Slice, reflect.Array:
		count := rv.Len()
		for i := range count {
			if !fn(i, count, NewValue(i), NewValue(rv.Index(i).Interface())) {
				break
			}
		}
		return nil
	case reflect.Map:
		keys := rv.MapKeys()
		slices.SortFunc(keys, func(a, b reflect.Value) int {
			va, vb := NewValue(a.Interface()), NewValue(b.Interface())
			if fa, err := va.Float(); err == nil {
				if fb, err := vb.Float(); err == nil {
					return cmp.Compare(fa, fb)
				}
			}
			return cmp.Compare(va.String(), vb.String())
		})
		count := len(keys)
		for i, k := range keys {
			if !fn(i, count, NewValue(k.Interface()), NewValue(rv.MapIndex(k).Interface())) {
				break
			}
		}
		return nil
	case reflect.String:
		rs := []rune(rv.String())
		count := len(rs)
		for i, r := range rs {
			if !fn(i, count, NewValue(i), NewValue(string(r))) {
				break
			}
		}
		return nil
	default:
		return fmt.Errorf("%w: %T", ErrTypeNotIterable, v.val)
	}
}

// Compare compares v with other.
// It returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v *Value) Compare(other *Value) (int, error) {
	if v.IsNil() && other.IsNil() {
		return 0, nil
	}
	if v.IsNil() {
		return -1, nil
	}
	if other.IsNil() {
		return 1, nil
	}

	// Try numeric comparison first.
	vf, vErr := v.Float()
	of, oErr := other.Float()
	if vErr == nil && oErr == nil {
		return cmp.Compare(vf, of), nil
	}

	// Fall back to string comparison.
	return cmp.Compare(v.String(), other.String()), nil
}

// Equals reports whether v and other represent the same value.
func (v *Value) Equals(other *Value) bool {
	if v.IsNil() && other.IsNil() {
		return true
	}
	if v.IsNil() || other.IsNil() {
		return false
	}

	// Try numeric comparison first to handle int vs float64.
	vf, vErr := v.Float()
	of, oErr := other.Float()
	if vErr == nil && oErr == nil {
		return vf == of
	}

	// Handle string comparison including alias types.
	vRv := v.resolved()
	oRv := other.resolved()
	if vRv.Kind() == reflect.String && oRv.Kind() == reflect.String {
		return vRv.String() == oRv.String()
	}

	return reflect.DeepEqual(v.val, other.val)
}
