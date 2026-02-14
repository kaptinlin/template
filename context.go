package template

import (
	"errors"
	"fmt"
	"maps"
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
		if next, ok := current[part].(map[string]any); ok {
			current = next
		} else {
			next = make(map[string]any)
			current[part] = next
			current = next
		}
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
func NewExecutionContext(data map[string]any) *ExecutionContext {
	return &ExecutionContext{
		Public:  Context(data),
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
// parent's Public context but copies the Private context for isolated
// scope.
func NewChildContext(parent *ExecutionContext) *ExecutionContext {
	return &ExecutionContext{
		Public:  parent.Public,
		Private: maps.Clone(parent.Private),
	}
}
