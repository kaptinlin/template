package template

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrammar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		context  *Context
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "Simple number",
			input:    "42",
			context:  &Context{},
			expected: float64(42),
		},
		{
			name:     "Simple string",
			input:    "'hello'",
			context:  &Context{},
			expected: "hello",
		},
		{
			name:  "Variable reference",
			input: "name",
			context: &Context{
				"name": "john",
			},
			expected: "john",
		},
		{
			name:     "Addition operation",
			input:    "1 + 2",
			context:  &Context{},
			expected: float64(3),
		},
		{
			name:     "Complex arithmetic operation",
			input:    "2 * (3 + 4)",
			context:  &Context{},
			expected: float64(14),
		},
		{
			name:     "Mixed arithmetic operation",
			input:    "1 + 2 * 3 - 4 / 2",
			context:  &Context{},
			expected: float64(5),
		},
		{
			name:     "Logical AND operation",
			input:    "1 < 2 && 3 > 1",
			context:  &Context{},
			expected: true,
		},
		{
			name:     "Logical OR operation",
			input:    "1 > 2 || 3 > 1",
			context:  &Context{},
			expected: true,
		},
		{
			name:     "Comparison operations",
			input:    "42 >= 42 && 42 <= 42 && 42 == 42 && 42 != 43",
			context:  &Context{},
			expected: true,
		},
		{
			name:  "Variable and filter combination",
			input: "name | upper | lower",
			context: &Context{
				"name": "John Doe",
			},
			expected: "john doe",
		},
		{
			name:  "Complex expression",
			input: "((name | upper) + ' is ') && ((age > 18 && age < 60 && 'adult') || (age <= 18 || age >= 60 && 'child'))",
			context: &Context{
				"name": "john",
				"age":  25,
			},
			expected: true,
		},
		{
			name:     "Logical NOT operator",
			input:    "!false",
			context:  &Context{},
			expected: true,
		},
		{
			name:    "Division by zero error",
			input:   "1 / 0",
			context: &Context{},
			wantErr: true,
		},
		{
			name:     "String filter",
			input:    "'hello' | upper",
			context:  &Context{},
			expected: "HELLO",
		},
		{
			name:    "Undefined variable",
			input:   "undefined_var",
			context: &Context{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := &Lexer{input: tt.input}
			tokens, err := lexer.Lex()
			require.NoError(t, err)

			grammar := NewGrammar(tokens)
			ast, err := grammar.Parse()
			if tt.wantErr && err != nil {
				// Error occurred during parsing as expected
				return
			}
			require.NoError(t, err)

			result, err := ast.Evaluate(*tt.context)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if err == nil {
				var got interface{}
				switch result.Type {
				case TypeInt:
					got = float64(result.Int)
				case TypeFloat:
					got = result.Float
				case TypeString:
					got = result.Str
				case TypeBool:
					got = result.Bool
				case TypeSlice:
					got = result.Slice
				case TypeMap:
					got = result.Map
				case TypeNil:
					got = nil
				case TypeStruct:
					got = result.Struct
				}

				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

// TestUnsignedIntegerOverflow tests that NewValue correctly handles
// unsigned integer values that would overflow when converted to int64
func TestUnsignedIntegerOverflow(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expectError bool
	}{
		{
			name:        "Valid uint64 within int64 range",
			value:       uint64(math.MaxInt64),
			expectError: false,
		},
		{
			name:        "Valid uint32 value",
			value:       uint32(1000),
			expectError: false,
		},
		{
			name:        "Valid uint16 value",
			value:       uint16(500),
			expectError: false,
		},
		{
			name:        "Valid uint8 value",
			value:       uint8(255),
			expectError: false,
		},
		{
			name:        "Overflow uint64 max value",
			value:       uint64(math.MaxUint64),
			expectError: true,
		},
		{
			name:        "Overflow uint64 near max",
			value:       uint64(math.MaxInt64 + 1),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewValue(tt.value)

			if tt.expectError {
				assert.Error(t, err, "Expected error for value %v", tt.value)
				assert.Nil(t, result, "Expected nil result on error")
			} else {
				require.NoError(t, err, "Unexpected error for value %v", tt.value)
				require.NotNil(t, result, "Expected valid result for value %v", tt.value)
				assert.Equal(t, TypeInt, result.Type, "Expected TypeInt for value %v", tt.value)
			}
		})
	}
}

// TestJSONTagFieldAccess tests that template fields can be accessed using JSON tag names
// in both variable output and conditional expressions. This ensures consistency between
// struct field lookup mechanisms across different template contexts.
func TestJSONTagFieldAccess(t *testing.T) {
	type Person struct {
		FullName  string `json:"name"`
		IsActive  bool   `json:"active"`
		IsEnabled bool   `json:"enabled,omitempty"`
	}

	type Container struct {
		User *Person `json:"user"`
	}

	tests := []struct {
		name     string
		data     *Container
		template string
		expected string
	}{
		{
			name:     "variable output with json tag",
			data:     &Container{User: &Person{FullName: "Alice", IsActive: true}},
			template: `{{ user.name }}`,
			expected: "Alice",
		},
		{
			name:     "boolean variable output",
			data:     &Container{User: &Person{IsActive: true}},
			template: `{{ user.active }}`,
			expected: "true",
		},
		{
			name:     "boolean in if condition - true",
			data:     &Container{User: &Person{IsEnabled: true}},
			template: `{% if user.enabled %}yes{% else %}no{% endif %}`,
			expected: "yes",
		},
		{
			name:     "boolean in if condition - false",
			data:     &Container{User: &Person{IsEnabled: false}},
			template: `{% if user.enabled %}yes{% else %}no{% endif %}`,
			expected: "no",
		},
		{
			name:     "multiple json tag fields",
			data:     &Container{User: &Person{FullName: "Bob", IsActive: true}},
			template: `{% if user.active %}{{ user.name }}{% endif %}`,
			expected: "Bob",
		},
		{
			name:     "and condition with boolean and string",
			data:     &Container{User: &Person{FullName: "Charlie", IsEnabled: true}},
			template: `{% if user.enabled and user.name != "" %}valid{% else %}invalid{% endif %}`,
			expected: "valid",
		},
		{
			name:     "and condition - first false",
			data:     &Container{User: &Person{FullName: "David", IsEnabled: false}},
			template: `{% if user.enabled and user.name != "" %}valid{% else %}invalid{% endif %}`,
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			ctx.Set("user", tt.data.User)

			result, err := Render(tt.template, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTextLogicalOperators tests text-based logical operators (and, or, not)
func TestTextLogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		template string
		ctx      map[string]interface{}
		want     string
	}{
		{"and-both-true", `{% if a and b %}y{% else %}n{% endif %}`,
			map[string]interface{}{"a": true, "b": true}, "y"},
		{"and-first-false", `{% if a and b %}y{% else %}n{% endif %}`,
			map[string]interface{}{"a": false, "b": true}, "n"},
		{"or-both-false", `{% if a or b %}y{% else %}n{% endif %}`,
			map[string]interface{}{"a": false, "b": false}, "n"},
		{"or-second-true", `{% if a or b %}y{% else %}n{% endif %}`,
			map[string]interface{}{"a": false, "b": true}, "y"},
		{"not-false", `{% if not a %}y{% else %}n{% endif %}`,
			map[string]interface{}{"a": false}, "y"},
		{"not-true", `{% if not a %}y{% else %}n{% endif %}`,
			map[string]interface{}{"a": true}, "n"},
		{"and-or-precedence", `{% if a or b and c %}y{% else %}n{% endif %}`,
			map[string]interface{}{"a": false, "b": true, "c": true}, "y"},
		{"not-and-precedence", `{% if not a and b %}y{% else %}n{% endif %}`,
			map[string]interface{}{"a": false, "b": true}, "y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			for k, v := range tt.ctx {
				ctx.Set(k, v)
			}
			result, err := Render(tt.template, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestInOperator tests membership operators (in, not in)
func TestInOperator(t *testing.T) {
	tests := []struct {
		name     string
		template string
		ctx      map[string]interface{}
		want     string
	}{
		{"string-in-string-found", `{% if "bc" in text %}y{% else %}n{% endif %}`,
			map[string]interface{}{"text": "abcdef"}, "y"},
		{"string-in-string-not-found", `{% if "xyz" in text %}y{% else %}n{% endif %}`,
			map[string]interface{}{"text": "abcdef"}, "n"},
		{"item-in-list", `{% if "hello" in items %}y{% else %}n{% endif %}`,
			map[string]interface{}{"items": []string{"hi", "hello", "hey"}}, "y"},
		{"item-not-in-list", `{% if "goodbye" in items %}y{% else %}n{% endif %}`,
			map[string]interface{}{"items": []string{"hi", "hello", "hey"}}, "n"},
		{"number-in-array", `{% if 2 in nums %}y{% else %}n{% endif %}`,
			map[string]interface{}{"nums": []int{1, 2, 3}}, "y"},
		{"not-in-operator", `{% if "xyz" not in text %}y{% else %}n{% endif %}`,
			map[string]interface{}{"text": "abcdef"}, "y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			for k, v := range tt.ctx {
				ctx.Set(k, v)
			}
			result, err := Render(tt.template, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestOperatorPrecedence tests operator precedence (Django-compatible)
func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    string
		ctx     map[string]interface{}
		want    string
		comment string
	}{
		{
			name:    "and-before-or",
			tmpl:    `{% if a or b and c %}y{% else %}n{% endif %}`,
			ctx:     map[string]interface{}{"a": false, "b": true, "c": true},
			want:    "y",
			comment: "a or (b and c) = false or true = true",
		},
		{
			name:    "not-before-and",
			tmpl:    `{% if not a and b %}y{% else %}n{% endif %}`,
			ctx:     map[string]interface{}{"a": false, "b": true},
			want:    "y",
			comment: "(not a) and b = true and true = true",
		},
		{
			name:    "comparison-before-and",
			tmpl:    `{% if a == b or c == d and e %}y{% else %}n{% endif %}`,
			ctx:     map[string]interface{}{"a": 1, "b": 2, "c": 3, "d": 3, "e": true},
			want:    "y",
			comment: "(a == b) or ((c == d) and e) = false or true = true",
		},
		{
			name:    "in-before-and",
			tmpl:    `{% if "x" in items and flag %}y{% else %}n{% endif %}`,
			ctx:     map[string]interface{}{"items": []string{"x", "y"}, "flag": true},
			want:    "y",
			comment: "(\"x\" in items) and flag = true and true = true",
		},
		{
			name:    "not-before-in",
			tmpl:    `{% if not flag and "x" in items %}y{% else %}n{% endif %}`,
			ctx:     map[string]interface{}{"flag": false, "items": []string{"x"}},
			want:    "y",
			comment: "(not flag) and (\"x\" in items) = true and true = true",
		},
		{
			name:    "complex-precedence",
			tmpl:    `{% if a or b and not c or d %}y{% else %}n{% endif %}`,
			ctx:     map[string]interface{}{"a": false, "b": true, "c": false, "d": false},
			want:    "y",
			comment: "a or (b and (not c)) or d = false or true or false = true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			for k, v := range tt.ctx {
				ctx.Set(k, v)
			}
			result, err := Render(tt.tmpl, ctx)
			require.NoError(t, err, "Comment: %s", tt.comment)
			assert.Equal(t, tt.want, result, "Comment: %s", tt.comment)
		})
	}
}

// TestDjangoPatterns tests real Django template patterns
func TestDjangoPatterns(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		ctx  map[string]interface{}
		want string
	}{
		{
			name: "auth-check",
			tmpl: `{% if user.authenticated and not user.banned %}ok{% endif %}`,
			ctx: map[string]interface{}{
				"user": map[string]interface{}{"authenticated": true, "banned": false},
			},
			want: "ok",
		},
		{
			name: "permission",
			tmpl: `{% if user.staff or user.admin %}admin{% else %}user{% endif %}`,
			ctx: map[string]interface{}{
				"user": map[string]interface{}{"staff": false, "admin": true},
			},
			want: "admin",
		},
		{
			name: "category-filter",
			tmpl: `{% if item.cat in allowed %}show{% endif %}`,
			ctx: map[string]interface{}{
				"item":    map[string]interface{}{"cat": "tech"},
				"allowed": []string{"tech", "news"},
			},
			want: "show",
		},
		{
			name: "null-check-with-and",
			tmpl: `{% if user != null and user.active %}active{% else %}inactive{% endif %}`,
			ctx: map[string]interface{}{
				"user": map[string]interface{}{"active": true},
			},
			want: "active",
		},
		{
			name: "null-check-is-null",
			tmpl: `{% if user == null %}guest{% else %}logged-in{% endif %}`,
			ctx: map[string]interface{}{
				"user": nil,
			},
			want: "guest",
		},
		{
			name: "empty-list-check",
			tmpl: `{% if items %}has-items{% else %}empty{% endif %}`,
			ctx: map[string]interface{}{
				"items": []string{},
			},
			want: "empty",
		},
		{
			name: "non-empty-list-check",
			tmpl: `{% if items %}has-items{% else %}empty{% endif %}`,
			ctx: map[string]interface{}{
				"items": []string{"a"},
			},
			want: "has-items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			for k, v := range tt.ctx {
				ctx.Set(k, v)
			}
			result, err := Render(tt.tmpl, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestLiterals tests boolean and null/none literals
func TestLiterals(t *testing.T) {
	tests := []struct {
		name     string
		template string
		ctx      map[string]interface{}
		want     string
	}{
		{"true-uppercase", `{% if val == True %}y{% else %}n{% endif %}`,
			map[string]interface{}{"val": true}, "y"},
		{"true-lowercase", `{% if val == true %}y{% else %}n{% endif %}`,
			map[string]interface{}{"val": true}, "y"},
		{"false-uppercase", `{% if val == False %}y{% else %}n{% endif %}`,
			map[string]interface{}{"val": false}, "y"},
		{"false-lowercase", `{% if val == false %}y{% else %}n{% endif %}`,
			map[string]interface{}{"val": false}, "y"},
		{"null-lowercase", `{% if val == null %}y{% else %}n{% endif %}`,
			map[string]interface{}{"val": nil}, "y"},
		{"null-capitalized", `{% if val == Null %}y{% else %}n{% endif %}`,
			map[string]interface{}{"val": nil}, "y"},
		{"none-lowercase", `{% if val == none %}y{% else %}n{% endif %}`,
			map[string]interface{}{"val": nil}, "y"},
		{"none-capitalized", `{% if val == None %}y{% else %}n{% endif %}`,
			map[string]interface{}{"val": nil}, "y"},
		{"not-null", `{% if val != null %}y{% else %}n{% endif %}`,
			map[string]interface{}{"val": "something"}, "y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			for k, v := range tt.ctx {
				ctx.Set(k, v)
			}
			result, err := Render(tt.template, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}
