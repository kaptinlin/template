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

// Context stores template variables as a string-keyed map.
// Values can be of any type. Dot-notation (e.g., "user.name") is
// supported for nested access.
type Context map[string]any

// ContextBuilder provides a fluent API for building a [Context] with
// error collection.
type ContextBuilder struct {
	context Context
	errors  []error
}

// ExecutionContext holds the execution state for template rendering,
// separating user-provided variables (Public) from internal variables
// (Private).
type ExecutionContext struct {
	Public  Context // user-provided variables
	Private Context // internal variables (e.g., loop counters)

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

// NewContext creates and returns a new empty [Context].
func NewContext() Context {
	return make(Context)
}

// NewContextBuilder creates a new [ContextBuilder] for fluent [Context]
// construction.
//
//	ctx, err := NewContextBuilder().
//	    KeyValue("name", "John").
//	    Struct(user).
//	    Build()
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		context: make(Context),
	}
}

// Set inserts a value into the Context with the specified key.
// Dot-notation (e.g., "user.address.city") creates nested map
// structures. Top-level keys preserve original data types; nested keys
// use map[string]any. Empty keys are silently ignored.
func (c Context) Set(key string, value any) {
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

// Get retrieves a value from the Context by key. Dot-separated keys
// (e.g., "user.profile.name") navigate nested structures. Array indices
// are supported (e.g., "items.0").
//
// Get returns [ErrContextKeyNotFound], [ErrContextIndexOutOfRange], or
// [ErrContextInvalidKeyType] on failure.
func (c Context) Get(key string) (any, error) {
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

// classifyGetError maps jsonpointer errors to context-level sentinel
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
//	builder := NewContextBuilder().
//	    KeyValue("name", "John").
//	    KeyValue("age", 30)
func (cb *ContextBuilder) KeyValue(key string, value any) *ContextBuilder {
	cb.context.Set(key, value)
	return cb
}

// Struct expands struct fields into the Context using JSON
// serialization. Fields are flattened to top-level keys based on their
// json tags. Nested structs are preserved as nested maps accessible via
// dot notation. If serialization fails, the error is collected and
// returned by [ContextBuilder.Build].
func (cb *ContextBuilder) Struct(v any) *ContextBuilder {
	if temp, ok := contextFromStructFast(v); ok {
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
	jsonMarshalerType = reflect.TypeOf((*stdjson.Marshaler)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

func contextFromStructFast(v any) (Context, bool) {
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
	return Context(out), true
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

		value, ok := valueToContextValue(rv.Field(i))
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

func valueToContextValue(rv reflect.Value) (any, bool) {
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
			value, ok := valueToContextValue(rv.Index(i))
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
			value, ok := valueToContextValue(iter.Value())
			if !ok {
				return nil, false
			}
			out[iter.Key().String()] = value
		}
		return out, true
	case reflect.Struct:
		return structToMapValue(rv)
	case reflect.Interface, reflect.Pointer:
		return valueToContextValue(rv)
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
	parts := strings.Split(tag, ",")
	opts := jsonTagOptions{}
	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			opts.omitEmpty = true
		}
	}
	return parts[0], opts
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

// Build returns the constructed Context and any collected errors.
// Errors from [ContextBuilder.KeyValue] or [ContextBuilder.Struct]
// operations are joined into a single error.
func (cb *ContextBuilder) Build() (Context, error) {
	if len(cb.errors) > 0 {
		return cb.context, errors.Join(cb.errors...)
	}
	return cb.context, nil
}

// NewExecutionContext creates a new [ExecutionContext] from user data.
func NewExecutionContext(data Context) *ExecutionContext {
	return &ExecutionContext{
		Public:  data,
		Private: NewContext(),
	}
}

// Get retrieves a variable, checking Private first, then Public.
func (ec *ExecutionContext) Get(name string) (any, bool) {
	if val, err := ec.Private.Get(name); err == nil {
		return val, true
	}
	if val, err := ec.Public.Get(name); err == nil {
		return val, true
	}
	return nil, false
}

// Set stores a variable in the private context.
func (ec *ExecutionContext) Set(name string, value any) {
	ec.Private.Set(name, value)
}

// NewChildContext creates a child [ExecutionContext] that shares the
// parent's Public context, copies the Private context for isolated
// scope, and preserves runtime rendering state.
func NewChildContext(parent *ExecutionContext) *ExecutionContext {
	return &ExecutionContext{
		Public:       parent.Public,
		Private:      maps.Clone(parent.Private),
		engine:       parent.engine,
		autoescape:   parent.autoescape,
		includeDepth: parent.includeDepth,
		currentLeaf:  parent.currentLeaf,
	}
}

// NewIsolatedChildContext creates a child [ExecutionContext] with a fresh
// public/private scope while preserving runtime rendering state.
func NewIsolatedChildContext(parent *ExecutionContext) *ExecutionContext {
	return &ExecutionContext{
		Public:       nil,
		Private:      NewContext(),
		engine:       parent.engine,
		autoescape:   parent.autoescape,
		includeDepth: parent.includeDepth,
		currentLeaf:  parent.currentLeaf,
	}
}
