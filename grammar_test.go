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
