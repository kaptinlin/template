package template

import (
	"testing"
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
			if err != nil {
				t.Fatalf("Lexer.Lex() error = %v", err)
			}

			grammar := NewGrammar(tokens)
			ast, err := grammar.Parse()
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			result, err := ast.Evaluate(*tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

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

				if got != tt.expected {
					t.Errorf("Evaluate() got = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}
