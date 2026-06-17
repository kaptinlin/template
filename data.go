package template

import (
	"encoding"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/go-json-experiment/json"
)

// Data stores template variables as a string-keyed map.
// Values can be of any type. Dot-notation (e.g., "user.name") is
// supported for nested access.
type Data map[string]any

// DataBuilder provides a fluent API for building a [Data] with
// error collection.
type DataBuilder struct {
	context Data
	errors  []error
}

// renderContext holds the execution state for template rendering,
// separating render input data (Data) from local bindings (Locals).
type renderContext struct {
	Data   Data // render input data
	Locals Data // local bindings (e.g., loop counters)

	// engine is the owning Engine for loader-backed renders; nil when a
	// template parsed from a source string renders without a loader.
	engine *Engine

	// autoescape controls whether {{ expr }} output is HTML-escaped.
	// True only for HTML-format engine renders.
	autoescape bool

	// includeDepth tracks the current {% include %} nesting depth to
	// defend against runaway recursion.
	includeDepth int

	// currentLeaf is the "most-child" template in the current extends
	// chain, used by blockNode.Execute to walk up through overrides.
	currentLeaf *Template
}

// NewData creates and returns a new empty [Data].
func NewData() Data {
	return make(Data)
}

// NewDataBuilder creates a new [DataBuilder] for fluent [Data]
// construction.
//
//	ctx, err := NewDataBuilder().
//	    KeyValue("name", "John").
//	    Struct(user).
//	    Build()
func NewDataBuilder() *DataBuilder {
	return &DataBuilder{
		context: make(Data),
	}
}

// Set inserts a value into the Data with the specified key.
// Dot-notation (e.g., "user.address.city") creates nested map
// structures. Top-level keys preserve original data types; nested keys
// use map[string]any. Empty keys are silently ignored.
func (c Data) Set(key string, value any) {
	if key == "" {
		return
	}

	parts := splitDotPath(key)
	if len(parts) == 1 {
		c[key] = value
		return
	}

	// Build intermediate map[string]any nodes for nested keys.
	current := c
	last := len(parts) - 1
	for _, part := range parts[:last] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[part] = next
		}
		current = next
	}
	current[parts[last]] = value
}

// Get retrieves a value from the Data by key. Dot-separated keys
// (e.g., "user.profile.name") navigate nested structures. Array indices
// are supported (e.g., "items.0").
//
// Get returns [ErrContextKeyNotFound], [ErrContextIndexOutOfRange], or
// [ErrContextInvalidKeyType] on failure.
func (c Data) Get(key string) (any, error) {
	if key == "" {
		return map[string]any(c), nil
	}

	parts := splitDotPath(key)
	if slices.Contains(parts, "") {
		return nil, fmt.Errorf("%w: empty path component in '%s'", ErrContextInvalidKeyType, key)
	}

	current := any(c)
	for _, part := range parts {
		next, err := dataPathValue(current, part)
		if err != nil {
			return nil, wrapDataGetError(err, key)
		}
		current = next
	}
	return current, nil
}

func dataPathValue(current any, part string) (any, error) {
	rv := newValue(current).resolved()
	if !rv.IsValid() {
		return nil, ErrContextInvalidKeyType
	}

	switch rv.Kind() {
	case reflect.Map:
		key, ok := dataPathMapKey(part, rv.Type().Key())
		if !ok {
			return nil, ErrContextKeyNotFound
		}
		value := rv.MapIndex(key)
		if !value.IsValid() {
			return nil, ErrContextKeyNotFound
		}
		return value.Interface(), nil
	case reflect.Struct:
		field, found := findStructField(rv, part)
		if !found {
			return nil, ErrContextKeyNotFound
		}
		return field.Interface(), nil
	case reflect.Slice, reflect.Array:
		index, err := dataPathIndex(part)
		if err != nil || index < 0 || index >= rv.Len() {
			return nil, ErrContextIndexOutOfRange
		}
		return rv.Index(index).Interface(), nil
	case reflect.String:
		index, err := dataPathIndex(part)
		if err != nil {
			return nil, ErrContextKeyNotFound
		}
		runes := []rune(rv.String())
		if index < 0 || index >= len(runes) {
			return nil, ErrContextIndexOutOfRange
		}
		return string(runes[index]), nil
	default:
		return nil, ErrContextKeyNotFound
	}
}

func dataPathMapKey(part string, target reflect.Type) (reflect.Value, bool) {
	if key, ok := mapKeyValue(part, target); ok {
		return key, true
	}

	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(part, 10, target.Bits())
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(i).Convert(target), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		i, err := strconv.ParseUint(part, 10, target.Bits())
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(i).Convert(target), true
	default:
		return reflect.Value{}, false
	}
}

func dataPathIndex(part string) (int, error) {
	i, err := strconv.ParseInt(part, 10, 0)
	if err != nil || !int64FitsInInt(i) {
		return 0, ErrContextIndexOutOfRange
	}
	return int(i), nil
}

func wrapDataGetError(err error, key string) error {
	return fmt.Errorf("%w: '%s'", err, key)
}

// splitDotPath splits a dot-notation string into path components.
func splitDotPath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}

// KeyValue sets a key-value pair and returns the builder for chaining.
//
//	builder := NewDataBuilder().
//	    KeyValue("name", "John").
//	    KeyValue("age", 30)
func (cb *DataBuilder) KeyValue(key string, value any) *DataBuilder {
	cb.context.Set(key, value)
	return cb
}

// Struct expands struct fields into the Data using JSON
// serialization. Fields are flattened to top-level keys based on their
// json tags. Nested structs are preserved as nested maps accessible via
// dot notation. If serialization fails, the error is collected and
// returned by [DataBuilder.Build].
func (cb *DataBuilder) Struct(v any) *DataBuilder {
	if temp, ok := dataFromStructFast(v); ok {
		maps.Copy(cb.context, temp)
		return cb
	}

	jsonData, err := json.Marshal(v)
	if err != nil {
		cb.errors = append(cb.errors, fmt.Errorf("marshal struct: %w", err))
		return cb
	}

	var temp map[string]any
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		cb.errors = append(cb.errors, fmt.Errorf("unmarshal struct: %w", err))
		return cb
	}

	maps.Copy(cb.context, temp)
	return cb
}

var (
	jsonMarshalerType = reflect.TypeFor[stdjson.Marshaler]()
	textMarshalerType = reflect.TypeFor[encoding.TextMarshaler]()
)

func dataFromStructFast(v any) (Data, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil, false
	}

	rv = indirectValue(rv)
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return nil, false
	}
	if hasCustomJSONSemantics(rv.Type()) {
		return nil, false
	}

	out, ok := structToMapValue(rv)
	if !ok {
		return nil, false
	}
	return Data(out), true
}

func structToMapValue(rv reflect.Value) (map[string]any, bool) {
	rt := rv.Type()
	out := make(map[string]any, rt.NumField())

	for i := range rt.NumField() {
		field := rt.Field(i)
		if field.Anonymous {
			return nil, false
		}
		if !field.IsExported() {
			continue
		}

		name, opts := parseJSONTag(field.Tag.Get("json"))
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}

		value, ok := valueToDataValue(rv.Field(i))
		if !ok {
			return nil, false
		}
		if opts.omitEmpty && isEmptyJSONValue(rv.Field(i)) {
			continue
		}
		out[name] = value
	}

	return out, true
}

func valueToDataValue(rv reflect.Value) (any, bool) {
	rv = indirectValue(rv)
	if !rv.IsValid() {
		return nil, true
	}
	if hasCustomJSONSemantics(rv.Type()) {
		return nil, false
	}

	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Interface(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Interface(), true
	case reflect.Float32, reflect.Float64:
		return rv.Interface(), true
	case reflect.String:
		return rv.String(), true
	case reflect.Slice, reflect.Array:
		out := make([]any, rv.Len())
		for i := range rv.Len() {
			value, ok := valueToDataValue(rv.Index(i))
			if !ok {
				return nil, false
			}
			out[i] = value
		}
		return out, true
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return nil, false
		}
		out := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			value, ok := valueToDataValue(iter.Value())
			if !ok {
				return nil, false
			}
			out[iter.Key().String()] = value
		}
		return out, true
	case reflect.Struct:
		return structToMapValue(rv)
	case reflect.Interface, reflect.Pointer:
		return valueToDataValue(rv)
	default:
		return nil, false
	}
}

func indirectValue(rv reflect.Value) reflect.Value {
	for rv.IsValid() && (rv.Kind() == reflect.Interface || rv.Kind() == reflect.Pointer) {
		if rv.IsNil() {
			return reflect.Value{}
		}
		rv = rv.Elem()
	}
	return rv
}

func hasCustomJSONSemantics(rt reflect.Type) bool {
	if rt == nil {
		return false
	}
	if rt.Implements(jsonMarshalerType) || rt.Implements(textMarshalerType) {
		return true
	}
	if rt.Kind() != reflect.Pointer {
		ptr := reflect.PointerTo(rt)
		return ptr.Implements(jsonMarshalerType) || ptr.Implements(textMarshalerType)
	}
	return false
}

type jsonTagOptions struct {
	omitEmpty bool
}

func parseJSONTag(tag string) (string, jsonTagOptions) {
	if tag == "" {
		return "", jsonTagOptions{}
	}

	name, rest, _ := strings.Cut(tag, ",")
	opts := jsonTagOptions{}
	for opt := range strings.SplitSeq(rest, ",") {
		if opt == "omitempty" {
			opts.omitEmpty = true
		}
	}
	return name, opts
}

func isEmptyJSONValue(rv reflect.Value) bool {
	rv = indirectValue(rv)
	if !rv.IsValid() {
		return true
	}

	switch rv.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return rv.IsNil()
	default:
		return false
	}
}

// Build returns the constructed Data and any collected errors.
// Errors from [DataBuilder.KeyValue] or [DataBuilder.Struct]
// operations are joined into a single error.
func (cb *DataBuilder) Build() (Data, error) {
	if len(cb.errors) > 0 {
		return cb.context, errors.Join(cb.errors...)
	}
	return cb.context, nil
}

// newRenderContext creates a new [renderContext] from user data.
func newRenderContext(data Data) *renderContext {
	return &renderContext{
		Data:   data,
		Locals: NewData(),
	}
}

// Get retrieves a variable, checking Locals first, then Data.
func (ec *renderContext) Get(name string) (any, bool) {
	if val, err := ec.Locals.Get(name); err == nil {
		return val, true
	}
	root, _, _ := strings.Cut(name, ".")
	if _, ok := ec.Locals[root]; ok {
		return nil, false
	}
	if val, err := ec.Data.Get(name); err == nil {
		return val, true
	}
	return nil, false
}

// Set stores a variable in the local bindings.
func (ec *renderContext) Set(name string, value any) {
	ec.Locals.Set(name, value)
}

// newChildContext creates a child [renderContext] that shares the
// parent's Data, copies the Locals for isolated scope, and preserves
// runtime rendering state.
func newChildContext(parent *renderContext) *renderContext {
	return &renderContext{
		Data:         parent.Data,
		Locals:       maps.Clone(parent.Locals),
		engine:       parent.engine,
		autoescape:   parent.autoescape,
		includeDepth: parent.includeDepth,
		currentLeaf:  parent.currentLeaf,
	}
}

// newIsolatedChildContext creates a child [renderContext] with fresh
// Data/Locals while preserving runtime rendering state.
func newIsolatedChildContext(parent *renderContext) *renderContext {
	return &renderContext{
		Data:         nil,
		Locals:       NewData(),
		engine:       parent.engine,
		autoescape:   parent.autoescape,
		includeDepth: parent.includeDepth,
		currentLeaf:  parent.currentLeaf,
	}
}
