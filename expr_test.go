package template

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// mustTokenize creates tokens from a string expression for testing.
// It wraps the expression in {{ }} and extracts the inner tokens.
func mustTokenize(t *testing.T, expr string) []*Token {
	t.Helper()
	source := "{{ " + expr + " }}"
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Failed to tokenize %q: %v", expr, err)
	}

	var exprTokens []*Token
	inExpr := false
	for _, tok := range tokens {
		if tok.Type == TokenVarBegin {
			inExpr = true
			continue
		}
		if tok.Type == TokenVarEnd {
			break
		}
		if inExpr {
			exprTokens = append(exprTokens, tok)
		}
	}
	return exprTokens
}

func TestNewExprParser(t *testing.T) {
	tests := []struct {
		name   string
		tokens []*Token
	}{
		{
			name:   "empty tokens",
			tokens: []*Token{},
		},
		{
			name: "single token",
			tokens: []*Token{
				{Type: TokenNumber, Value: "42", Line: 1, Col: 1},
			},
		},
		{
			name: "multiple tokens",
			tokens: []*Token{
				{Type: TokenIdentifier, Value: "x", Line: 1, Col: 1},
				{Type: TokenSymbol, Value: "+", Line: 1, Col: 3},
				{Type: TokenNumber, Value: "10", Line: 1, Col: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExprParser(tt.tokens)
			if parser == nil {
				t.Fatalf("NewExprParser(%v) = nil, want non-nil", tt.tokens)
			}
			if got, want := parser.tokens, tt.tokens; !reflect.DeepEqual(got, want) {
				t.Errorf("NewExprParser(%v).tokens = %v, want %v", tt.tokens, got, want)
			}
			if got, want := parser.pos, 0; got != want {
				t.Errorf("NewExprParser(%v).pos = %v, want %v", tt.tokens, got, want)
			}
		})
	}
}

func TestParseLiterals(t *testing.T) {
	tests := []struct {
		name          string
		expr          string
		expectedType  string
		expectedValue interface{}
	}{
		{
			name:          "string literal",
			expr:          `"hello"`,
			expectedType:  "LiteralNode",
			expectedValue: "hello",
		},
		{
			name:          "integer literal",
			expr:          "42",
			expectedType:  "LiteralNode",
			expectedValue: 42.0,
		},
		{
			name:          "float literal",
			expr:          "3.14",
			expectedType:  "LiteralNode",
			expectedValue: 3.14,
		},
		{
			name:          "boolean true",
			expr:          "true",
			expectedType:  "LiteralNode",
			expectedValue: true,
		},
		{
			name:          "boolean false",
			expr:          "false",
			expectedType:  "LiteralNode",
			expectedValue: false,
		},
		{
			name:          "negative number",
			expr:          "-10",
			expectedType:  "UnaryOpNode",
			expectedValue: nil, // Check operator instead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}
			if result == nil {
				t.Fatalf("ParseExpression(%q) = nil, want non-nil", tt.expr)
			}

			switch tt.expectedType {
			case "LiteralNode":
				litNode, ok := result.(*LiteralNode)
				if !ok {
					t.Errorf("ParseExpression(%q) type = %T, want *LiteralNode", tt.expr, result)
					return
				}
				if got, want := litNode.Value, tt.expectedValue; got != want {
					t.Errorf("ParseExpression(%q).Value = %v, want %v", tt.expr, got, want)
				}
			case "UnaryOpNode":
				unaryNode, ok := result.(*UnaryOpNode)
				if !ok {
					t.Errorf("ParseExpression(%q) type = %T, want *UnaryOpNode", tt.expr, result)
					return
				}
				if got, want := unaryNode.Operator, "-"; got != want {
					t.Errorf("ParseExpression(%q).Operator = %v, want %v", tt.expr, got, want)
				}
			}
		})
	}
}

func TestParseVariables(t *testing.T) {
	tests := []struct {
		name         string
		expr         string
		expectedName string
	}{
		{
			name:         "simple variable",
			expr:         "x",
			expectedName: "x",
		},
		{
			name:         "variable with underscore",
			expr:         "user_name",
			expectedName: "user_name",
		},
		{
			name:         "variable with number",
			expr:         "var123",
			expectedName: "var123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}

			varNode, ok := result.(*VariableNode)
			if !ok {
				t.Errorf("ParseExpression(%q) type = %T, want *VariableNode", tt.expr, result)
				return
			}
			if got, want := varNode.Name, tt.expectedName; got != want {
				t.Errorf("ParseExpression(%q).Name = %v, want %v", tt.expr, got, want)
			}
		})
	}
}

func TestParseUnaryOperators(t *testing.T) {
	tests := []struct {
		name             string
		expr             string
		expectedOperator string
	}{
		{
			name:             "not operator",
			expr:             "not x",
			expectedOperator: "not",
		},
		{
			name:             "negative operator",
			expr:             "-10",
			expectedOperator: "-",
		},
		{
			name:             "positive operator",
			expr:             "+10",
			expectedOperator: "+",
		},
		{
			name:             "nested not",
			expr:             "not not x",
			expectedOperator: "not",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}

			unaryNode, ok := result.(*UnaryOpNode)
			if !ok {
				t.Errorf("ParseExpression(%q) type = %T, want *UnaryOpNode", tt.expr, result)
				return
			}
			if got, want := unaryNode.Operator, tt.expectedOperator; got != want {
				t.Errorf("ParseExpression(%q).Operator = %v, want %v", tt.expr, got, want)
			}
		})
	}
}

func TestParseBinaryOperators(t *testing.T) {
	tests := []struct {
		name             string
		expr             string
		expectedOperator string
	}{
		// Arithmetic operators
		{
			name:             "addition",
			expr:             "a + b",
			expectedOperator: "+",
		},
		{
			name:             "subtraction",
			expr:             "a - b",
			expectedOperator: "-",
		},
		{
			name:             "multiplication",
			expr:             "a * b",
			expectedOperator: "*",
		},
		{
			name:             "division",
			expr:             "a / b",
			expectedOperator: "/",
		},
		{
			name:             "modulo",
			expr:             "a % b",
			expectedOperator: "%",
		},
		// Comparison operators
		{
			name:             "equal",
			expr:             "a == b",
			expectedOperator: "==",
		},
		{
			name:             "not equal",
			expr:             "a != b",
			expectedOperator: "!=",
		},
		{
			name:             "less than",
			expr:             "a < b",
			expectedOperator: "<",
		},
		{
			name:             "greater than",
			expr:             "a > b",
			expectedOperator: ">",
		},
		{
			name:             "less than or equal",
			expr:             "a <= b",
			expectedOperator: "<=",
		},
		{
			name:             "greater than or equal",
			expr:             "a >= b",
			expectedOperator: ">=",
		},
		// Logical operators
		{
			name:             "logical and",
			expr:             "a and b",
			expectedOperator: "and",
		},
		{
			name:             "logical or",
			expr:             "a or b",
			expectedOperator: "or",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}

			binNode, ok := result.(*BinaryOpNode)
			if !ok {
				t.Errorf("ParseExpression(%q) type = %T, want *BinaryOpNode", tt.expr, result)
				return
			}
			if got, want := binNode.Operator, tt.expectedOperator; got != want {
				t.Errorf("ParseExpression(%q).Operator = %v, want %v", tt.expr, got, want)
			}
		})
	}
}

func TestParseOperatorPrecedence(t *testing.T) {
	tests := []struct {
		name           string
		expr           string
		validateResult func(*testing.T, Expression)
	}{
		{
			name: "multiplication before addition",
			expr: "a + b * c",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: a + (b * c)
				binNode := result.(*BinaryOpNode)
				if got, want := binNode.Operator, "+"; got != want {
					t.Errorf("Operator = %v, want %v", got, want)
				}

				// Right side should be multiplication
				rightBin, ok := binNode.Right.(*BinaryOpNode)
				if !ok {
					t.Errorf("Right type = %T, want *BinaryOpNode", binNode.Right)
					return
				}
				if got, want := rightBin.Operator, "*"; got != want {
					t.Errorf("Right.Operator = %v, want %v", got, want)
				}
			},
		},
		{
			name: "addition before comparison",
			expr: "a + b > c",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: (a + b) > c
				binNode := result.(*BinaryOpNode)
				if got, want := binNode.Operator, ">"; got != want {
					t.Errorf("Operator = %v, want %v", got, want)
				}

				// Left side should be addition
				leftBin, ok := binNode.Left.(*BinaryOpNode)
				if !ok {
					t.Errorf("Left type = %T, want *BinaryOpNode", binNode.Left)
					return
				}
				if got, want := leftBin.Operator, "+"; got != want {
					t.Errorf("Left.Operator = %v, want %v", got, want)
				}
			},
		},
		{
			name: "comparison before and",
			expr: "a > b and c < d",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: (a > b) and (c < d)
				binNode := result.(*BinaryOpNode)
				if got, want := binNode.Operator, "and"; got != want {
					t.Errorf("Operator = %v, want %v", got, want)
				}

				// Both sides should be comparisons
				leftBin, ok := binNode.Left.(*BinaryOpNode)
				if !ok {
					t.Errorf("Left type = %T, want *BinaryOpNode", binNode.Left)
					return
				}
				if got, want := leftBin.Operator, ">"; got != want {
					t.Errorf("Left.Operator = %v, want %v", got, want)
				}

				rightBin, ok := binNode.Right.(*BinaryOpNode)
				if !ok {
					t.Errorf("Right type = %T, want *BinaryOpNode", binNode.Right)
					return
				}
				if got, want := rightBin.Operator, "<"; got != want {
					t.Errorf("Right.Operator = %v, want %v", got, want)
				}
			},
		},
		{
			name: "and before or",
			expr: "a or b and c",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: a or (b and c)
				binNode := result.(*BinaryOpNode)
				if got, want := binNode.Operator, "or"; got != want {
					t.Errorf("Operator = %v, want %v", got, want)
				}

				// Right side should be and
				rightBin, ok := binNode.Right.(*BinaryOpNode)
				if !ok {
					t.Errorf("Right type = %T, want *BinaryOpNode", binNode.Right)
					return
				}
				if got, want := rightBin.Operator, "and"; got != want {
					t.Errorf("Right.Operator = %v, want %v", got, want)
				}
			},
		},
		{
			name: "complex precedence",
			expr: "a + b * c > d and e or f",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: ((a + (b * c)) > d and e) or f
				binNode := result.(*BinaryOpNode)
				if got, want := binNode.Operator, "or"; got != want {
					t.Errorf("Operator = %v, want %v", got, want)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}
			tt.validateResult(t, result)
		})
	}
}

func TestParsePropertyAccess(t *testing.T) {
	tests := []struct {
		name             string
		expr             string
		expectedProperty string
	}{
		{
			name:             "simple property",
			expr:             "user.name",
			expectedProperty: "name",
		},
		{
			name:             "chained property",
			expr:             "user.profile.email",
			expectedProperty: "email",
		},
		{
			name:             "property with underscore",
			expr:             "user.first_name",
			expectedProperty: "first_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}

			// For chained properties like "user.profile.email",
			// the result is PropertyAccessNode with property="email"
			propNode, ok := result.(*PropertyAccessNode)
			if !ok {
				t.Errorf("ParseExpression(%q) type = %T, want *PropertyAccessNode", tt.expr, result)
				return
			}
			if got, want := propNode.Property, tt.expectedProperty; got != want {
				t.Errorf("ParseExpression(%q).Property = %v, want %v", tt.expr, got, want)
			}
		})
	}
}

func TestParseSubscript(t *testing.T) {
	tests := []struct {
		name           string
		expr           string
		validateResult func(*testing.T, Expression)
	}{
		{
			name: "numeric subscript",
			expr: "items[0]",
			validateResult: func(t *testing.T, result Expression) {
				subNode, ok := result.(*SubscriptNode)
				if !ok {
					t.Errorf("type = %T, want *SubscriptNode", result)
					return
				}

				// Check index is a literal number
				litNode, ok := subNode.Index.(*LiteralNode)
				if !ok {
					t.Errorf("Index type = %T, want *LiteralNode", subNode.Index)
					return
				}
				if got, want := litNode.Value, 0.0; got != want {
					t.Errorf("Index.Value = %v, want %v", got, want)
				}
			},
		},
		{
			name: "string subscript",
			expr: `dict["key"]`,
			validateResult: func(t *testing.T, result Expression) {
				subNode, ok := result.(*SubscriptNode)
				if !ok {
					t.Errorf("type = %T, want *SubscriptNode", result)
					return
				}

				// Check index is a literal string
				litNode, ok := subNode.Index.(*LiteralNode)
				if !ok {
					t.Errorf("Index type = %T, want *LiteralNode", subNode.Index)
					return
				}
				if got, want := litNode.Value, "key"; got != want {
					t.Errorf("Index.Value = %v, want %v", got, want)
				}
			},
		},
		{
			name: "variable subscript",
			expr: "items[i]",
			validateResult: func(t *testing.T, result Expression) {
				subNode, ok := result.(*SubscriptNode)
				if !ok {
					t.Errorf("type = %T, want *SubscriptNode", result)
					return
				}

				// Check index is a variable
				varNode, ok := subNode.Index.(*VariableNode)
				if !ok {
					t.Errorf("Index type = %T, want *VariableNode", subNode.Index)
					return
				}
				if got, want := varNode.Name, "i"; got != want {
					t.Errorf("Index.Name = %v, want %v", got, want)
				}
			},
		},
		{
			name: "chained subscript",
			expr: "matrix[0][1]",
			validateResult: func(t *testing.T, result Expression) {
				// Outer subscript
				outerSub, ok := result.(*SubscriptNode)
				if !ok {
					t.Errorf("type = %T, want *SubscriptNode", result)
					return
				}

				// Inner subscript
				innerSub, ok := outerSub.Object.(*SubscriptNode)
				if !ok {
					t.Errorf("Object type = %T, want *SubscriptNode", outerSub.Object)
					return
				}
				if innerSub == nil {
					t.Errorf("Object = nil, want non-nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}
			tt.validateResult(t, result)
		})
	}
}

func TestParseFilter(t *testing.T) {
	tests := []struct {
		name           string
		expr           string
		expectedFilter string
		expectedArgs   int
	}{
		{
			name:           "filter without args",
			expr:           "name|upper",
			expectedFilter: "upper",
			expectedArgs:   0,
		},
		{
			name:           "filter with one arg",
			expr:           "price|add:10",
			expectedFilter: "add",
			expectedArgs:   1,
		},
		{
			name:           "filter with multiple args",
			expr:           "text|slice:0,10",
			expectedFilter: "slice",
			expectedArgs:   2,
		},
		{
			name:           "chained filters",
			expr:           "name|lower|capitalize",
			expectedFilter: "capitalize",
			expectedArgs:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}

			// Find the outermost filter
			filterNode, ok := result.(*FilterNode)
			if !ok {
				t.Errorf("ParseExpression(%q) type = %T, want *FilterNode", tt.expr, result)
				return
			}
			if got, want := filterNode.Name, tt.expectedFilter; got != want {
				t.Errorf("ParseExpression(%q).Name = %v, want %v", tt.expr, got, want)
			}
			if got, want := len(filterNode.Args), tt.expectedArgs; got != want {
				t.Errorf("ParseExpression(%q) args count = %v, want %v", tt.expr, got, want)
			}
		})
	}
}

func TestParseParentheses(t *testing.T) {
	tests := []struct {
		name           string
		expr           string
		validateResult func(*testing.T, Expression)
	}{
		{
			name: "simple parentheses",
			expr: "(x)",
			validateResult: func(t *testing.T, result Expression) {
				varNode, ok := result.(*VariableNode)
				if !ok {
					t.Errorf("type = %T, want *VariableNode", result)
					return
				}
				if got, want := varNode.Name, "x"; got != want {
					t.Errorf("Name = %v, want %v", got, want)
				}
			},
		},
		{
			name: "parentheses change precedence",
			expr: "(a + b) * c",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: (a + b) * c
				binNode := result.(*BinaryOpNode)
				if got, want := binNode.Operator, "*"; got != want {
					t.Errorf("Operator = %v, want %v", got, want)
				}

				// Left side should be addition
				leftBin, ok := binNode.Left.(*BinaryOpNode)
				if !ok {
					t.Errorf("Left type = %T, want *BinaryOpNode", binNode.Left)
					return
				}
				if got, want := leftBin.Operator, "+"; got != want {
					t.Errorf("Left.Operator = %v, want %v", got, want)
				}
			},
		},
		{
			name: "nested parentheses",
			expr: "((x + y) * z)",
			validateResult: func(t *testing.T, result Expression) {
				binNode, ok := result.(*BinaryOpNode)
				if !ok {
					t.Errorf("type = %T, want *BinaryOpNode", result)
					return
				}
				if got, want := binNode.Operator, "*"; got != want {
					t.Errorf("Operator = %v, want %v", got, want)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}
			tt.validateResult(t, result)
		})
	}
}

func TestParseComplexExpressions(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{
			name: "mixed operators and property access",
			expr: "user.age > 18 and user.active",
		},
		{
			name: "subscript with arithmetic",
			expr: "items[i + 1]",
		},
		{
			name: "filter with property access",
			expr: "user.name|upper",
		},
		{
			name: "complex nested expression",
			expr: "(a + b) * c > d and not e or f.g[0]|lower",
		},
		{
			name: "multiple property accesses",
			expr: "user.profile.settings.theme",
		},
		{
			name: "arithmetic with multiple operations",
			expr: "a + b - c * d / e % f",
		},
		{
			name: "logical operations",
			expr: "a and b or c and d or e",
		},
		{
			name: "comparison chain",
			expr: "a < b and b < c and c < d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}
			if result == nil {
				t.Fatalf("ParseExpression(%q) = nil, want non-nil", tt.expr)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name          string
		expr          string
		errorContains string
	}{
		{
			name:          "missing closing parenthesis",
			expr:          "(x + y",
			errorContains: "expected ')'",
		},
		{
			name:          "missing closing bracket",
			expr:          "items[0",
			errorContains: "expected ']'",
		},
		{
			name:          "property without name",
			expr:          "user.",
			errorContains: "expected property name",
		},
		{
			name:          "filter without name",
			expr:          "name|",
			errorContains: "expected filter name",
		},
		{
			name:          "unexpected end",
			expr:          "",
			errorContains: "unexpected end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			_, err := parser.ParseExpression()
			if err == nil {
				t.Errorf("ParseExpression(%q) error = nil, want error containing %q", tt.expr, tt.errorContains)
				return
			}
			if !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("ParseExpression(%q) error = %q, want error containing %q", tt.expr, err.Error(), tt.errorContains)
			}
		})
	}
}

func TestParseFilterArguments(t *testing.T) {
	tests := []struct {
		name         string
		expr         string
		expectedArgs []interface{}
	}{
		{
			name:         "string argument",
			expr:         `name|default:"Anonymous"`,
			expectedArgs: []interface{}{"Anonymous"},
		},
		{
			name:         "number argument",
			expr:         "value|add:10",
			expectedArgs: []interface{}{10.0},
		},
		{
			name:         "boolean argument true",
			expr:         "flag|default:true",
			expectedArgs: []interface{}{true},
		},
		{
			name:         "boolean argument false",
			expr:         "flag|default:false",
			expectedArgs: []interface{}{false},
		},
		{
			name:         "variable argument",
			expr:         "text|slice:start,end",
			expectedArgs: []interface{}{"start", "end"}, // Variable names as identifiers
		},
		{
			name:         "multiple mixed arguments",
			expr:         `item|format:"Name: %s",name,true`,
			expectedArgs: []interface{}{"Name: %s", "name", true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}

			filterNode, ok := result.(*FilterNode)
			if !ok {
				t.Errorf("ParseExpression(%q) type = %T, want *FilterNode", tt.expr, result)
				return
			}
			if got, want := len(filterNode.Args), len(tt.expectedArgs); got != want {
				t.Errorf("ParseExpression(%q) args count = %v, want %v", tt.expr, got, want)
			}

			for i, expectedArg := range tt.expectedArgs {
				arg := filterNode.Args[i]
				switch expected := expectedArg.(type) {
				case string:
					// Could be either LiteralNode or VariableNode
					if litNode, ok := arg.(*LiteralNode); ok {
						if got, want := litNode.Value, expected; got != want {
							t.Errorf("Args[%d].Value = %v, want %v", i, got, want)
						}
					} else if varNode, ok := arg.(*VariableNode); ok {
						if got, want := varNode.Name, expected; got != want {
							t.Errorf("Args[%d].Name = %v, want %v", i, got, want)
						}
					} else {
						t.Errorf("Args[%d] type = %T, want *LiteralNode or *VariableNode", i, arg)
					}
				case float64:
					litNode, ok := arg.(*LiteralNode)
					if !ok {
						t.Errorf("Args[%d] type = %T, want *LiteralNode", i, arg)
						continue
					}
					if got, want := litNode.Value, expected; got != want {
						t.Errorf("Args[%d].Value = %v, want %v", i, got, want)
					}
				case bool:
					litNode, ok := arg.(*LiteralNode)
					if !ok {
						t.Errorf("Args[%d] type = %T, want *LiteralNode", i, arg)
						continue
					}
					if got, want := litNode.Value, expected; got != want {
						t.Errorf("Args[%d].Value = %v, want %v", i, got, want)
					}
				}
			}
		})
	}
}

func TestExprParserHelperMethods(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "current returns current token",
			test: func(t *testing.T) {
				tokens := []*Token{
					{Type: TokenNumber, Value: "1", Line: 1, Col: 1},
					{Type: TokenNumber, Value: "2", Line: 1, Col: 2},
				}
				parser := NewExprParser(tokens)
				if got, want := parser.current(), tokens[0]; got != want {
					t.Errorf("current() = %v, want %v", got, want)
				}
			},
		},
		{
			name: "current returns nil at end",
			test: func(t *testing.T) {
				tokens := []*Token{}
				parser := NewExprParser(tokens)
				if got := parser.current(); got != nil {
					t.Errorf("current() = %v, want nil", got)
				}
			},
		},
		{
			name: "advance moves position",
			test: func(t *testing.T) {
				tokens := []*Token{
					{Type: TokenNumber, Value: "1", Line: 1, Col: 1},
					{Type: TokenNumber, Value: "2", Line: 1, Col: 2},
				}
				parser := NewExprParser(tokens)
				parser.advance()
				if got, want := parser.current(), tokens[1]; got != want {
					t.Errorf("current() after advance() = %v, want %v", got, want)
				}
				if got, want := parser.pos, 1; got != want {
					t.Errorf("pos after advance() = %v, want %v", got, want)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		expected string
	}{
		{
			name: "error with position",
			err: &ParseError{
				Message: "unexpected token",
				Line:    5,
				Col:     10,
			},
			expected: "parse error at line 5, col 10: unexpected token",
		},
		{
			name: "error without position",
			err: &ParseError{
				Message: "syntax error",
				Line:    0,
				Col:     0,
			},
			expected: "parse error at line 0, col 0: syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if got, want := result, tt.expected; got != want {
				t.Errorf("Error() = %v, want %v", got, want)
			}
		})
	}
}

func TestExprParserErrorMethods(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "parseErr creates error with current position",
			test: func(t *testing.T) {
				tokens := []*Token{
					{Type: TokenNumber, Value: "1", Line: 5, Col: 10},
				}
				parser := NewExprParser(tokens)
				err := parser.parseErr("test error")
				var parseErr *ParseError
				if !errors.As(err, &parseErr) {
					t.Errorf("parseErr() error type = %T, want *ParseError", err)
					return
				}
				if got, want := parseErr.Message, "test error"; got != want {
					t.Errorf("parseErr().Message = %v, want %v", got, want)
				}
				if got, want := parseErr.Line, 5; got != want {
					t.Errorf("parseErr().Line = %v, want %v", got, want)
				}
				if got, want := parseErr.Col, 10; got != want {
					t.Errorf("parseErr().Col = %v, want %v", got, want)
				}
			},
		},
		{
			name: "parseErr at end returns error with zero position",
			test: func(t *testing.T) {
				tokens := []*Token{}
				parser := NewExprParser(tokens)
				err := parser.parseErr("test error")
				var parseErr *ParseError
				if !errors.As(err, &parseErr) {
					t.Errorf("parseErr() error type = %T, want *ParseError", err)
					return
				}
				if got, want := parseErr.Message, "test error"; got != want {
					t.Errorf("parseErr().Message = %v, want %v", got, want)
				}
				if got, want := parseErr.Line, 0; got != want {
					t.Errorf("parseErr().Line = %v, want %v", got, want)
				}
				if got, want := parseErr.Col, 0; got != want {
					t.Errorf("parseErr().Col = %v, want %v", got, want)
				}
			},
		},
		{
			name: "errAtTok creates error at token position",
			test: func(t *testing.T) {
				token := &Token{Type: TokenNumber, Value: "1", Line: 3, Col: 7}
				parser := NewExprParser([]*Token{})
				err := parser.errAtTok(token, "custom error")
				var parseErr *ParseError
				if !errors.As(err, &parseErr) {
					t.Errorf("errAtTok() error type = %T, want *ParseError", err)
					return
				}
				if got, want := parseErr.Message, "custom error"; got != want {
					t.Errorf("errAtTok().Message = %v, want %v", got, want)
				}
				if got, want := parseErr.Line, 3; got != want {
					t.Errorf("errAtTok().Line = %v, want %v", got, want)
				}
				if got, want := parseErr.Col, 7; got != want {
					t.Errorf("errAtTok().Col = %v, want %v", got, want)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestExpressionNodeTypes(t *testing.T) {
	tests := []struct {
		name           string
		expr           string
		expectedType   string
		validateStruct func(*testing.T, Expression)
	}{
		{
			name:         "LiteralNode structure",
			expr:         "42",
			expectedType: "*template.LiteralNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*LiteralNode)
				if got, want := node.Value, 42.0; got != want {
					t.Errorf("Value = %v, want %v", got, want)
				}
				line, col := node.Position()
				if line <= 0 {
					t.Errorf("Position() line = %v, want > 0", line)
				}
				if col <= 0 {
					t.Errorf("Position() col = %v, want > 0", col)
				}
			},
		},
		{
			name:         "VariableNode structure",
			expr:         "username",
			expectedType: "*template.VariableNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*VariableNode)
				if got, want := node.Name, "username"; got != want {
					t.Errorf("Name = %v, want %v", got, want)
				}
				line, col := node.Position()
				if line <= 0 {
					t.Errorf("Position() line = %v, want > 0", line)
				}
				if col <= 0 {
					t.Errorf("Position() col = %v, want > 0", col)
				}
			},
		},
		{
			name:         "BinaryOpNode structure",
			expr:         "a + b",
			expectedType: "*template.BinaryOpNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*BinaryOpNode)
				if got, want := node.Operator, "+"; got != want {
					t.Errorf("Operator = %v, want %v", got, want)
				}
				if node.Left == nil {
					t.Errorf("Left = nil, want non-nil")
				}
				if node.Right == nil {
					t.Errorf("Right = nil, want non-nil")
				}
				line, col := node.Position()
				if line <= 0 {
					t.Errorf("Position() line = %v, want > 0", line)
				}
				if col <= 0 {
					t.Errorf("Position() col = %v, want > 0", col)
				}
			},
		},
		{
			name:         "UnaryOpNode structure",
			expr:         "not x",
			expectedType: "*template.UnaryOpNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*UnaryOpNode)
				if got, want := node.Operator, "not"; got != want {
					t.Errorf("Operator = %v, want %v", got, want)
				}
				if node.Operand == nil {
					t.Errorf("Operand = nil, want non-nil")
				}
				line, col := node.Position()
				if line <= 0 {
					t.Errorf("Position() line = %v, want > 0", line)
				}
				if col <= 0 {
					t.Errorf("Position() col = %v, want > 0", col)
				}
			},
		},
		{
			name:         "PropertyAccessNode structure",
			expr:         "user.name",
			expectedType: "*template.PropertyAccessNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*PropertyAccessNode)
				if got, want := node.Property, "name"; got != want {
					t.Errorf("Property = %v, want %v", got, want)
				}
				if node.Object == nil {
					t.Errorf("Object = nil, want non-nil")
				}
				line, col := node.Position()
				if line <= 0 {
					t.Errorf("Position() line = %v, want > 0", line)
				}
				if col <= 0 {
					t.Errorf("Position() col = %v, want > 0", col)
				}
			},
		},
		{
			name:         "SubscriptNode structure",
			expr:         "items[0]",
			expectedType: "*template.SubscriptNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*SubscriptNode)
				if node.Object == nil {
					t.Errorf("Object = nil, want non-nil")
				}
				if node.Index == nil {
					t.Errorf("Index = nil, want non-nil")
				}
				line, col := node.Position()
				if line <= 0 {
					t.Errorf("Position() line = %v, want > 0", line)
				}
				if col <= 0 {
					t.Errorf("Position() col = %v, want > 0", col)
				}
			},
		},
		{
			name:         "FilterNode structure",
			expr:         "name|upper",
			expectedType: "*template.FilterNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*FilterNode)
				if got, want := node.Name, "upper"; got != want {
					t.Errorf("Name = %v, want %v", got, want)
				}
				if node.Expr == nil {
					t.Errorf("Expression = nil, want non-nil")
				}
				line, col := node.Position()
				if line <= 0 {
					t.Errorf("Position() line = %v, want > 0", line)
				}
				if col <= 0 {
					t.Errorf("Position() col = %v, want > 0", col)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
			}

			// Check type using reflection
			typeName := reflect.TypeOf(result).String()
			if got, want := typeName, tt.expectedType; got != want {
				t.Errorf("ParseExpression(%q) type = %v, want %v", tt.expr, got, want)
			}

			// Validate structure
			if tt.validateStruct != nil {
				tt.validateStruct(t, result)
			}
		})
	}
}

func TestParseExpressionEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		expr          string
		expectError   bool
		errorContains string
	}{
		{
			name:        "expression with whitespace",
			expr:        "  x  +  y  ",
			expectError: false,
		},
		{
			name:          "empty expression",
			expr:          "",
			expectError:   true,
			errorContains: "unexpected end",
		},
		{
			name:        "deeply nested parentheses",
			expr:        "((((x))))",
			expectError: false,
		},
		{
			name:        "long property chain",
			expr:        "a.b.c.d.e.f.g.h",
			expectError: false,
		},
		{
			name:        "multiple chained filters",
			expr:        "text|lower|trim|capitalize",
			expectError: false,
		},
		{
			name:          "invalid number",
			expr:          "123abc",
			expectError:   false, // Will be treated as identifier
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseExpression(%q) error = nil, want error", tt.expr)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ParseExpression(%q) error = %q, want error containing %q", tt.expr, err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
					return
				}
				if result == nil {
					t.Errorf("ParseExpression(%q) = nil, want non-nil", tt.expr)
				}
			}
		})
	}
}

// TestLogicalOperatorSymbols tests support for &&, ||, and ! operators
func TestLogicalOperatorSymbols(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		context  map[string]interface{}
		expected interface{}
		wantErr  bool
	}{
		// || operator tests
		{
			name:     "|| with both false",
			expr:     "false || false",
			context:  map[string]interface{}{},
			expected: false,
		},
		{
			name:     "|| with first true",
			expr:     "true || false",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "|| with second true",
			expr:     "false || true",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "|| with both true",
			expr:     "true || true",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "|| with variables",
			expr:     "a || b",
			context:  map[string]interface{}{"a": false, "b": true},
			expected: true,
		},

		// && operator tests
		{
			name:     "&& with both false",
			expr:     "false && false",
			context:  map[string]interface{}{},
			expected: false,
		},
		{
			name:     "&& with first false",
			expr:     "false && true",
			context:  map[string]interface{}{},
			expected: false,
		},
		{
			name:     "&& with second false",
			expr:     "true && false",
			context:  map[string]interface{}{},
			expected: false,
		},
		{
			name:     "&& with both true",
			expr:     "true && true",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "&& with variables",
			expr:     "a && b",
			context:  map[string]interface{}{"a": true, "b": true},
			expected: true,
		},

		// ! operator tests
		{
			name:     "! with true",
			expr:     "!true",
			context:  map[string]interface{}{},
			expected: false,
		},
		{
			name:     "! with false",
			expr:     "!false",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "! with variable",
			expr:     "!flag",
			context:  map[string]interface{}{"flag": false},
			expected: true,
		},
		{
			name:     "! with number 0",
			expr:     "!0",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "! with number non-zero",
			expr:     "!5",
			context:  map[string]interface{}{},
			expected: false,
		},

		// Complex expressions
		{
			name:     "Mixed && and ||",
			expr:     "true && false || true",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "! with &&",
			expr:     "!false && true",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "! with ||",
			expr:     "!true || false",
			context:  map[string]interface{}{},
			expected: false,
		},
		{
			name:     "Nested !",
			expr:     "!!true",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "Comparison with &&",
			expr:     "score >= 80 && grade == 'A'",
			context:  map[string]interface{}{"score": 85, "grade": "A"},
			expected: true,
		},
		{
			name:     "Comparison with ||",
			expr:     "score < 50 || grade == 'F'",
			context:  map[string]interface{}{"score": 40, "grade": "D"},
			expected: true,
		},
		{
			name:     "! with comparison",
			expr:     "!(score < 60)",
			context:  map[string]interface{}{"score": 85},
			expected: true,
		},

		// Parentheses with new operators
		{
			name:     "Parentheses with &&",
			expr:     "(true && true) && true",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "Parentheses with ||",
			expr:     "(false || true) || false",
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name:     "Parentheses with !",
			expr:     "!(true && false)",
			context:  map[string]interface{}{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := mustTokenize(t, tt.expr)
			parser := NewExprParser(tokens)
			expr, err := parser.ParseExpression()

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseExpression(%q) error = nil, want error", tt.expr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseExpression(%q) error = %v, want nil", tt.expr, err)
				return
			}
			if expr == nil {
				t.Fatalf("ParseExpression(%q) = nil, want non-nil", tt.expr)
			}

			// Evaluate the expression
			ctx := NewExecutionContext(tt.context)
			result, err := expr.Evaluate(ctx)
			if err != nil {
				t.Errorf("Evaluate(%q) error = %v, want nil", tt.expr, err)
				return
			}
			if result == nil {
				t.Fatalf("Evaluate(%q) = nil, want non-nil", tt.expr)
			}

			// Check the result
			if b, ok := tt.expected.(bool); ok {
				if got, want := result.IsTrue(), b; got != want {
					t.Errorf("Evaluate(%q).IsTrue() = %v, want %v", tt.expr, got, want)
				}
			} else {
				if got, want := result.Interface(), tt.expected; got != want {
					t.Errorf("Evaluate(%q).Interface() = %v, want %v", tt.expr, got, want)
				}
			}
		})
	}
}

// TestKeywordAndSymbolEquivalence tests that keywords and symbols are equivalent
func TestKeywordAndSymbolEquivalence(t *testing.T) {
	tests := []struct {
		name        string
		exprKeyword string
		exprSymbol  string
		context     map[string]interface{}
	}{
		{
			name:        "and vs &&",
			exprKeyword: "true and false",
			exprSymbol:  "true && false",
			context:     map[string]interface{}{},
		},
		{
			name:        "or vs ||",
			exprKeyword: "true or false",
			exprSymbol:  "true || false",
			context:     map[string]interface{}{},
		},
		{
			name:        "not vs !",
			exprKeyword: "not true",
			exprSymbol:  "!true",
			context:     map[string]interface{}{},
		},
		{
			name:        "Complex: and/or vs &&/||",
			exprKeyword: "a and b or c",
			exprSymbol:  "a && b || c",
			context:     map[string]interface{}{"a": true, "b": false, "c": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse and evaluate keyword version
			tokensKeyword := mustTokenize(t, tt.exprKeyword)
			parserKeyword := NewExprParser(tokensKeyword)
			exprKeyword, err := parserKeyword.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.exprKeyword, err)
			}

			ctx := NewExecutionContext(tt.context)
			resultKeyword, err := exprKeyword.Evaluate(ctx)
			if err != nil {
				t.Fatalf("Evaluate(%q) error = %v, want nil", tt.exprKeyword, err)
			}

			// Parse and evaluate symbol version
			tokensSymbol := mustTokenize(t, tt.exprSymbol)
			parserSymbol := NewExprParser(tokensSymbol)
			exprSymbol, err := parserSymbol.ParseExpression()
			if err != nil {
				t.Fatalf("ParseExpression(%q) error = %v, want nil", tt.exprSymbol, err)
			}

			resultSymbol, err := exprSymbol.Evaluate(ctx)
			if err != nil {
				t.Fatalf("Evaluate(%q) error = %v, want nil", tt.exprSymbol, err)
			}

			// Results should be identical
			if got, want := resultSymbol.IsTrue(), resultKeyword.IsTrue(); got != want {
				t.Errorf("Keyword %q and symbol %q should produce same result: got %v, want %v",
					tt.exprKeyword, tt.exprSymbol, got, want)
			}
		})
	}
}
