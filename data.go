package template

import (
	"encoding"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/jsonpointer"
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

// RenderContext holds the execution state for template rendering,
// separating render input data (Data) from local bindings (Locals).
type RenderContext struct {
	Data   Data // render input data
	Locals Data // local bindings (e.g., loop counters)

	// engine is the owning Engine when rendering via Engine.Render; nil for
	// standalone Template.Execute calls.
	engine *Engine

	// autoescape controls whether {{ expr }} output is HTML-escaped.
	// True only for HTML-format engine renders.
	autoescape bool

	// includeDepth tracks the current {% include %} nesting depth to
	// defend against runaway recursion.
	includeDepth int

	// currentLeaf is the "most-child" template in the current extends
	// chain, used by BlockNode.Execute to walk up through overrides.
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

	value, err := jsonpointer.Get(c, parts...)
	if err != nil {
		return nil, classifyGetError(err, key)
	}
	return value, nil
}

// classifyGetError maps jsonpointer errors to data-level sentinel
// errors.
func classifyGetError(err error, key string) error {
	switch {
	case errors.Is(err, jsonpointer.ErrNotFound),
		errors.Is(err, jsonpointer.ErrKeyNotFound),
		errors.Is(err, jsonpointer.ErrFieldNotFound):
		return fmt.Errorf("%w: '%s'", ErrContextKeyNotFound, key)
	case errors.Is(err, jsonpointer.ErrIndexOutOfBounds),
		errors.Is(err, jsonpointer.ErrInvalidIndex):
		return fmt.Errorf("%w: '%s'", ErrContextIndexOutOfRange, key)
	case errors.Is(err, jsonpointer.ErrInvalidPath),
		errors.Is(err, jsonpointer.ErrInvalidPathStep):
		return fmt.Errorf("%w: '%s'", ErrContextInvalidKeyType, key)
	default:
		return fmt.Errorf("accessing '%s': %w", key, err)
	}
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

	switch rv.Kind() { //nolint:exhaustive
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

	switch rv.Kind() { //nolint:exhaustive
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

// NewRenderContext creates a new [RenderContext] from user data.
func NewRenderContext(data Data) *RenderContext {
	return &RenderContext{
		Data:   data,
		Locals: NewData(),
	}
}

// Get retrieves a variable, checking Locals first, then Data.
func (ec *RenderContext) Get(name string) (any, bool) {
	if val, err := ec.Locals.Get(name); err == nil {
		return val, true
	}
	if val, err := ec.Data.Get(name); err == nil {
		return val, true
	}
	return nil, false
}

// Set stores a variable in the local bindings.
func (ec *RenderContext) Set(name string, value any) {
	ec.Locals.Set(name, value)
}

// NewChildContext creates a child [RenderContext] that shares the
// parent's Data, copies the Locals for isolated scope, and preserves
// runtime rendering state.
func NewChildContext(parent *RenderContext) *RenderContext {
	return &RenderContext{
		Data:         parent.Data,
		Locals:       maps.Clone(parent.Locals),
		engine:       parent.engine,
		autoescape:   parent.autoescape,
		includeDepth: parent.includeDepth,
		currentLeaf:  parent.currentLeaf,
	}
}

// NewIsolatedChildContext creates a child [RenderContext] with fresh
// Data/Locals while preserving runtime rendering state.
func NewIsolatedChildContext(parent *RenderContext) *RenderContext {
	return &RenderContext{
		Data:         nil,
		Locals:       NewData(),
		engine:       parent.engine,
		autoescape:   parent.autoescape,
		includeDepth: parent.includeDepth,
		currentLeaf:  parent.currentLeaf,
	}
}
