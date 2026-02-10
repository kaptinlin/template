package template

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/jsonpointer"
)

// Context stores template variables in a map structure, used for passing and accessing variables during template execution.
// Keys are strings, and values can be of any type, supporting dot-notation (.) for nested access.
type Context map[string]any

// ContextBuilder provides a fluent API for building Context with error collection.
type ContextBuilder struct {
	context Context
	errors  []error
}

// NewContext creates and returns a new empty Context instance.
// Example usage: ctx := NewContext()
func NewContext() Context {
	return make(Context)
}

// NewContextBuilder creates a new ContextBuilder for fluent Context construction.
// Example usage:
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

// Set inserts a value into the Context with the specified key, supporting dot-notation (.) for nested keys.
// This method preserves original data types for top-level keys and creates minimal map structures
// only when needed for nested access. Relies on jsonpointer's powerful reading capabilities.
func (c Context) Set(key string, value any) {
	if key == "" {
		return // Silently ignore empty keys for backward compatibility
	}

	// Convert dot notation to path components compatible with jsonpointer
	parts := convertDotToPath(key)

	// For top-level keys, always store the original type directly
	// This preserves structs, slices, arrays, and other complex types
	if len(parts) == 1 {
		c[key] = value
		return
	}

	// For nested keys, create a simple nested map structure
	// jsonpointer will handle reading from complex types, but for setting
	// we need a predictable map structure
	current := c
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part: set the actual value
			current[part] = value
		} else {
			// Intermediate parts: ensure they are map types
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]any)
			} else if _, ok := current[part].(map[string]any); !ok {
				// If exists but not a map, replace with map for nested access
				current[part] = make(map[string]any)
			}
			current = current[part].(map[string]any)
		}
	}
}

// Get retrieves a value from the Context for the specified key, supporting nested key access.
// Uses jsonpointer.Get to handle complex key paths, such as array indices and nested properties.
// Parameters:
//   - key: The key to retrieve, can be a dot-separated nested key like "user.profile.name"
//
// Returns:
//   - any: The retrieved value
//   - error: Returns an appropriate error if the key doesn't exist or access fails
func (c Context) Get(key string) (any, error) {
	// Handle empty key
	if key == "" {
		return map[string]any(c), nil
	}

	// Convert dot notation to path components
	parts := convertDotToPath(key)

	// Validate each part (basic validation before calling jsonpointer)
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("%w: empty path component in key '%s'", ErrContextInvalidKeyType, key)
		}
	}

	// Use jsonpointer.Get with variadic arguments (more efficient than string concatenation)
	value, err := jsonpointer.Get(c, parts...)
	if err != nil {
		switch {
		case errors.Is(err, jsonpointer.ErrNotFound) || errors.Is(err, jsonpointer.ErrKeyNotFound) || errors.Is(err, jsonpointer.ErrFieldNotFound):
			return nil, fmt.Errorf("%w: key '%s' not found in context", ErrContextKeyNotFound, key)
		case errors.Is(err, jsonpointer.ErrIndexOutOfBounds) || errors.Is(err, jsonpointer.ErrInvalidIndex):
			return nil, fmt.Errorf("%w: index out of range for key '%s'", ErrContextIndexOutOfRange, key)
		case errors.Is(err, jsonpointer.ErrInvalidPath) || errors.Is(err, jsonpointer.ErrInvalidPathStep):
			return nil, fmt.Errorf("%w: invalid type for key '%s', cannot navigate", ErrContextInvalidKeyType, key)
		}

		return nil, fmt.Errorf("unknown error while accessing key '%s': %w", key, err)
	}
	return value, nil
}

// convertDotToPath converts dot notation to path components array
func convertDotToPath(dotNotation string) []string {
	if dotNotation == "" {
		return []string{}
	}
	return strings.Split(dotNotation, ".")
}

// KeyValue sets a key-value pair and returns the ContextBuilder to support method chaining.
// Example:
//
//	builder := NewContextBuilder().
//	    KeyValue("name", "John").
//	    KeyValue("age", 30)
func (cb *ContextBuilder) KeyValue(key string, value any) *ContextBuilder {
	cb.context.Set(key, value)
	return cb
}

// Struct expands struct fields into the Context using JSON serialization.
// The struct fields are flattened to top-level keys based on their json tags.
// Nested structs are preserved as nested maps and can be accessed using dot notation.
// If serialization fails, the error is collected and can be retrieved via Build().
//
// Example:
//
//	type User struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//	builder := NewContextBuilder().Struct(User{Name: "John", Age: 30})
//	// Results in: context["name"]="John", context["age"]=30
//	// Template access: {{ name }}, {{ age }}
func (cb *ContextBuilder) Struct(v any) *ContextBuilder {
	// Marshal struct to JSON
	jsonData, err := json.Marshal(v)
	if err != nil {
		cb.errors = append(cb.errors, fmt.Errorf("struct: marshal failed: %w", err))
		return cb
	}

	// Unmarshal to temporary map
	temp := make(map[string]any)
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		cb.errors = append(cb.errors, fmt.Errorf("struct: unmarshal failed: %w", err))
		return cb
	}

	// Explicitly merge into Context
	for k, v := range temp {
		cb.context[k] = v
	}

	return cb
}

// Build returns the constructed Context and any errors collected during the build process.
// If there were any errors during KeyValue or Struct operations, they are returned as a joined error.
// Example:
//
//	ctx, err := NewContextBuilder().
//	    KeyValue("name", "John").
//	    Struct(user).
//	    Build()
//	if err != nil {
//	    // handle errors
//	}
func (cb *ContextBuilder) Build() (Context, error) {
	if len(cb.errors) > 0 {
		return cb.context, errors.Join(cb.errors...)
	}
	return cb.context, nil
}

// ========================================
// ExecutionContext - for template execution
// ========================================

// ExecutionContext holds the execution state for template rendering.
// It separates user-provided variables (Public) from internal variables (Private).
type ExecutionContext struct {
	// Public contains user-provided variables
	Public Context

	// Private contains internal variables (e.g., loop counters, temporary values)
	Private Context
}

// NewExecutionContext creates a new execution context from user data.
// The data is stored in the Public context, and an empty Private context is created.
func NewExecutionContext(data map[string]interface{}) *ExecutionContext {
	return &ExecutionContext{
		Public:  Context(data),
		Private: NewContext(),
	}
}

// Get retrieves a variable from the execution context.
// It first checks Private (for loop variables, etc.), then checks Public.
func (ctx *ExecutionContext) Get(name string) (interface{}, bool) {
	// Check private first (for loop variables, etc.)
	if val, err := ctx.Private.Get(name); err == nil {
		return val, true
	}

	// Then check public
	if val, err := ctx.Public.Get(name); err == nil {
		return val, true
	}

	return nil, false
}

// Set sets a variable in the private context.
// This is used for internal variables like loop counters.
func (ctx *ExecutionContext) Set(name string, value interface{}) {
	ctx.Private.Set(name, value)
}

// NewChildContext creates a child execution context.
// It shares the same Public context but gets a copy of the Private context.
// This is useful for nested scopes (like loops or blocks).
func NewChildContext(parent *ExecutionContext) *ExecutionContext {
	// Create new private context and copy parent's private variables
	childPrivate := NewContext()
	for k, v := range parent.Private {
		childPrivate[k] = v
	}

	return &ExecutionContext{
		Public:  parent.Public, // Share public context
		Private: childPrivate,  // Copy private context
	}
}
