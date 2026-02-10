package template

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to create tokens from a string expression
func tokenizeExpression(t *testing.T, expr string) []*Token {
	// Wrap expression in {{ }} to make it valid for lexer
	source := "{{ " + expr + " }}"
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	assert.NoError(t, err)

	// Extract tokens between {{ and }}
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
			assert.NotNil(t, parser)
			assert.Equal(t, tt.tokens, parser.tokens)
			assert.Equal(t, 0, parser.pos)
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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)
			assert.NotNil(t, result)

			switch tt.expectedType {
			case "LiteralNode":
				litNode, ok := result.(*LiteralNode)
				assert.True(t, ok, "expected LiteralNode")
				assert.Equal(t, tt.expectedValue, litNode.Value)
			case "UnaryOpNode":
				unaryNode, ok := result.(*UnaryOpNode)
				assert.True(t, ok, "expected UnaryOpNode")
				assert.Equal(t, "-", unaryNode.Operator)
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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)

			varNode, ok := result.(*VariableNode)
			assert.True(t, ok, "expected VariableNode")
			assert.Equal(t, tt.expectedName, varNode.Name)
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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)

			unaryNode, ok := result.(*UnaryOpNode)
			assert.True(t, ok, "expected UnaryOpNode")
			assert.Equal(t, tt.expectedOperator, unaryNode.Operator)
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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)

			binNode, ok := result.(*BinaryOpNode)
			assert.True(t, ok, "expected BinaryOpNode")
			assert.Equal(t, tt.expectedOperator, binNode.Operator)
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
				assert.Equal(t, "+", binNode.Operator)

				// Right side should be multiplication
				rightBin, ok := binNode.Right.(*BinaryOpNode)
				assert.True(t, ok)
				assert.Equal(t, "*", rightBin.Operator)
			},
		},
		{
			name: "addition before comparison",
			expr: "a + b > c",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: (a + b) > c
				binNode := result.(*BinaryOpNode)
				assert.Equal(t, ">", binNode.Operator)

				// Left side should be addition
				leftBin, ok := binNode.Left.(*BinaryOpNode)
				assert.True(t, ok)
				assert.Equal(t, "+", leftBin.Operator)
			},
		},
		{
			name: "comparison before and",
			expr: "a > b and c < d",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: (a > b) and (c < d)
				binNode := result.(*BinaryOpNode)
				assert.Equal(t, "and", binNode.Operator)

				// Both sides should be comparisons
				leftBin, ok := binNode.Left.(*BinaryOpNode)
				assert.True(t, ok)
				assert.Equal(t, ">", leftBin.Operator)

				rightBin, ok := binNode.Right.(*BinaryOpNode)
				assert.True(t, ok)
				assert.Equal(t, "<", rightBin.Operator)
			},
		},
		{
			name: "and before or",
			expr: "a or b and c",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: a or (b and c)
				binNode := result.(*BinaryOpNode)
				assert.Equal(t, "or", binNode.Operator)

				// Right side should be and
				rightBin, ok := binNode.Right.(*BinaryOpNode)
				assert.True(t, ok)
				assert.Equal(t, "and", rightBin.Operator)
			},
		},
		{
			name: "complex precedence",
			expr: "a + b * c > d and e or f",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: ((a + (b * c)) > d and e) or f
				binNode := result.(*BinaryOpNode)
				assert.Equal(t, "or", binNode.Operator)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)
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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)

			// For chained properties like "user.profile.email",
			// the result is PropertyAccessNode with property="email"
			propNode, ok := result.(*PropertyAccessNode)
			assert.True(t, ok, "expected PropertyAccessNode")
			assert.Equal(t, tt.expectedProperty, propNode.Property)
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
				assert.True(t, ok)

				// Check index is a literal number
				litNode, ok := subNode.Index.(*LiteralNode)
				assert.True(t, ok)
				assert.Equal(t, 0.0, litNode.Value)
			},
		},
		{
			name: "string subscript",
			expr: `dict["key"]`,
			validateResult: func(t *testing.T, result Expression) {
				subNode, ok := result.(*SubscriptNode)
				assert.True(t, ok)

				// Check index is a literal string
				litNode, ok := subNode.Index.(*LiteralNode)
				assert.True(t, ok)
				assert.Equal(t, "key", litNode.Value)
			},
		},
		{
			name: "variable subscript",
			expr: "items[i]",
			validateResult: func(t *testing.T, result Expression) {
				subNode, ok := result.(*SubscriptNode)
				assert.True(t, ok)

				// Check index is a variable
				varNode, ok := subNode.Index.(*VariableNode)
				assert.True(t, ok)
				assert.Equal(t, "i", varNode.Name)
			},
		},
		{
			name: "chained subscript",
			expr: "matrix[0][1]",
			validateResult: func(t *testing.T, result Expression) {
				// Outer subscript
				outerSub, ok := result.(*SubscriptNode)
				assert.True(t, ok)

				// Inner subscript
				innerSub, ok := outerSub.Object.(*SubscriptNode)
				assert.True(t, ok)
				assert.NotNil(t, innerSub)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)
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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)

			// Find the outermost filter
			filterNode, ok := result.(*FilterNode)
			assert.True(t, ok, "expected FilterNode")
			assert.Equal(t, tt.expectedFilter, filterNode.FilterName)
			assert.Equal(t, tt.expectedArgs, len(filterNode.Args))
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
				assert.True(t, ok)
				assert.Equal(t, "x", varNode.Name)
			},
		},
		{
			name: "parentheses change precedence",
			expr: "(a + b) * c",
			validateResult: func(t *testing.T, result Expression) {
				// Should parse as: (a + b) * c
				binNode := result.(*BinaryOpNode)
				assert.Equal(t, "*", binNode.Operator)

				// Left side should be addition
				leftBin, ok := binNode.Left.(*BinaryOpNode)
				assert.True(t, ok)
				assert.Equal(t, "+", leftBin.Operator)
			},
		},
		{
			name: "nested parentheses",
			expr: "((x + y) * z)",
			validateResult: func(t *testing.T, result Expression) {
				binNode, ok := result.(*BinaryOpNode)
				assert.True(t, ok)
				assert.Equal(t, "*", binNode.Operator)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)
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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)
			assert.NotNil(t, result)
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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			_, err := parser.ParseExpression()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)

			filterNode, ok := result.(*FilterNode)
			assert.True(t, ok)
			assert.Equal(t, len(tt.expectedArgs), len(filterNode.Args))

			for i, expectedArg := range tt.expectedArgs {
				arg := filterNode.Args[i]
				switch expected := expectedArg.(type) {
				case string:
					// Could be either LiteralNode or VariableNode
					if litNode, ok := arg.(*LiteralNode); ok {
						assert.Equal(t, expected, litNode.Value)
					} else if varNode, ok := arg.(*VariableNode); ok {
						assert.Equal(t, expected, varNode.Name)
					} else {
						t.Errorf("expected LiteralNode or VariableNode, got %T", arg)
					}
				case float64:
					litNode, ok := arg.(*LiteralNode)
					assert.True(t, ok)
					assert.Equal(t, expected, litNode.Value)
				case bool:
					litNode, ok := arg.(*LiteralNode)
					assert.True(t, ok)
					assert.Equal(t, expected, litNode.Value)
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
				assert.Equal(t, tokens[0], parser.current())
			},
		},
		{
			name: "current returns nil at end",
			test: func(t *testing.T) {
				tokens := []*Token{}
				parser := NewExprParser(tokens)
				assert.Nil(t, parser.current())
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
				assert.Equal(t, tokens[1], parser.current())
				assert.Equal(t, 1, parser.pos)
			},
		},
		{
			name: "peek returns token at offset",
			test: func(t *testing.T) {
				tokens := []*Token{
					{Type: TokenNumber, Value: "1", Line: 1, Col: 1},
					{Type: TokenNumber, Value: "2", Line: 1, Col: 2},
					{Type: TokenNumber, Value: "3", Line: 1, Col: 3},
				}
				parser := NewExprParser(tokens)
				assert.Equal(t, tokens[1], parser.peek(1))
				assert.Equal(t, tokens[2], parser.peek(2))
				assert.Nil(t, parser.peek(10))
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
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExprParserErrorMethods(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "error creates error with current position",
			test: func(t *testing.T) {
				tokens := []*Token{
					{Type: TokenNumber, Value: "1", Line: 5, Col: 10},
				}
				parser := NewExprParser(tokens)
				err := parser.error("test error")
				var parseErr *ParseError
				ok := errors.As(err, &parseErr)
				assert.True(t, ok)
				assert.Equal(t, "test error", parseErr.Message)
				assert.Equal(t, 5, parseErr.Line)
				assert.Equal(t, 10, parseErr.Col)
			},
		},
		{
			name: "error at end returns error with zero position",
			test: func(t *testing.T) {
				tokens := []*Token{}
				parser := NewExprParser(tokens)
				err := parser.error("test error")
				var parseErr *ParseError
				ok := errors.As(err, &parseErr)
				assert.True(t, ok)
				assert.Equal(t, "test error", parseErr.Message)
				assert.Equal(t, 0, parseErr.Line)
				assert.Equal(t, 0, parseErr.Col)
			},
		},
		{
			name: "errorAt creates error at token position",
			test: func(t *testing.T) {
				token := &Token{Type: TokenNumber, Value: "1", Line: 3, Col: 7}
				parser := NewExprParser([]*Token{})
				err := parser.errorAt(token, "custom error")
				var parseErr *ParseError
				ok := errors.As(err, &parseErr)
				assert.True(t, ok)
				assert.Equal(t, "custom error", parseErr.Message)
				assert.Equal(t, 3, parseErr.Line)
				assert.Equal(t, 7, parseErr.Col)
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
				assert.Equal(t, 42.0, node.Value)
				line, col := node.Position()
				assert.Greater(t, line, 0)
				assert.Greater(t, col, 0)
			},
		},
		{
			name:         "VariableNode structure",
			expr:         "username",
			expectedType: "*template.VariableNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*VariableNode)
				assert.Equal(t, "username", node.Name)
				line, col := node.Position()
				assert.Greater(t, line, 0)
				assert.Greater(t, col, 0)
			},
		},
		{
			name:         "BinaryOpNode structure",
			expr:         "a + b",
			expectedType: "*template.BinaryOpNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*BinaryOpNode)
				assert.Equal(t, "+", node.Operator)
				assert.NotNil(t, node.Left)
				assert.NotNil(t, node.Right)
				line, col := node.Position()
				assert.Greater(t, line, 0)
				assert.Greater(t, col, 0)
			},
		},
		{
			name:         "UnaryOpNode structure",
			expr:         "not x",
			expectedType: "*template.UnaryOpNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*UnaryOpNode)
				assert.Equal(t, "not", node.Operator)
				assert.NotNil(t, node.Operand)
				line, col := node.Position()
				assert.Greater(t, line, 0)
				assert.Greater(t, col, 0)
			},
		},
		{
			name:         "PropertyAccessNode structure",
			expr:         "user.name",
			expectedType: "*template.PropertyAccessNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*PropertyAccessNode)
				assert.Equal(t, "name", node.Property)
				assert.NotNil(t, node.Object)
				line, col := node.Position()
				assert.Greater(t, line, 0)
				assert.Greater(t, col, 0)
			},
		},
		{
			name:         "SubscriptNode structure",
			expr:         "items[0]",
			expectedType: "*template.SubscriptNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*SubscriptNode)
				assert.NotNil(t, node.Object)
				assert.NotNil(t, node.Index)
				line, col := node.Position()
				assert.Greater(t, line, 0)
				assert.Greater(t, col, 0)
			},
		},
		{
			name:         "FilterNode structure",
			expr:         "name|upper",
			expectedType: "*template.FilterNode",
			validateStruct: func(t *testing.T, expr Expression) {
				node := expr.(*FilterNode)
				assert.Equal(t, "upper", node.FilterName)
				assert.NotNil(t, node.Expression)
				line, col := node.Position()
				assert.Greater(t, line, 0)
				assert.Greater(t, col, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()
			assert.NoError(t, err)

			// Check type using reflection
			typeName := reflect.TypeOf(result).String()
			assert.Equal(t, tt.expectedType, typeName)

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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			result, err := parser.ParseExpression()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
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
			tokens := tokenizeExpression(t, tt.expr)
			parser := NewExprParser(tokens)
			expr, err := parser.ParseExpression()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, expr)

			// Evaluate the expression
			ctx := NewExecutionContext(tt.context)
			result, err := expr.Evaluate(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Check the result
			if b, ok := tt.expected.(bool); ok {
				assert.Equal(t, b, result.IsTrue())
			} else {
				assert.Equal(t, tt.expected, result.Interface())
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
			tokensKeyword := tokenizeExpression(t, tt.exprKeyword)
			parserKeyword := NewExprParser(tokensKeyword)
			exprKeyword, err := parserKeyword.ParseExpression()
			assert.NoError(t, err)

			ctx := NewExecutionContext(tt.context)
			resultKeyword, err := exprKeyword.Evaluate(ctx)
			assert.NoError(t, err)

			// Parse and evaluate symbol version
			tokensSymbol := tokenizeExpression(t, tt.exprSymbol)
			parserSymbol := NewExprParser(tokensSymbol)
			exprSymbol, err := parserSymbol.ParseExpression()
			assert.NoError(t, err)

			resultSymbol, err := exprSymbol.Evaluate(ctx)
			assert.NoError(t, err)

			// Results should be identical
			assert.Equal(t, resultKeyword.IsTrue(), resultSymbol.IsTrue(),
				"Keyword '%s' and symbol '%s' should produce same result",
				tt.exprKeyword, tt.exprSymbol)
		})
	}
}
