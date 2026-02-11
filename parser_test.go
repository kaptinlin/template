package template

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// parseTemplate is a test helper that tokenizes and parses the input.
func parseTemplate(t *testing.T, input string) ([]Statement, error) {
	t.Helper()
	tokens, err := NewLexer(input).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize(%q) returned unexpected error: %v", input, err)
	}
	return NewParser(tokens).Parse()
}

// =============================================================================
// Test Group 1: Basic Text Node Parsing
// =============================================================================

func TestParserTextNode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Statement
	}{
		{
			name:  "plain_text",
			input: "Hello, World!",
			want: []Statement{
				&TextNode{Text: "Hello, World!", Line: 1, Col: 1},
			},
		},
		{
			name:  "multiline_text",
			input: "Line 1\nLine 2\nLine 3",
			want: []Statement{
				&TextNode{Text: "Line 1\nLine 2\nLine 3", Line: 1, Col: 1},
			},
		},
		{
			name:  "special_characters",
			input: "Hello <html> & \"quotes\"",
			want: []Statement{
				&TextNode{Text: "Hello <html> & \"quotes\"", Line: 1, Col: 1},
			},
		},
		{
			name:  "empty_template",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTemplate(t, tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.input, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Parse(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

// =============================================================================
// Test Group 2: Output Node (Variable) Parsing
// =============================================================================

func TestParserOutputNode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Statement
	}{
		{
			name:  "simple_variable",
			input: "{{ name }}",
			want: []Statement{
				&OutputNode{
					Expr: &VariableNode{Name: "name", Line: 1, Col: 4},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "number_literal",
			input: "{{ 42 }}",
			want: []Statement{
				&OutputNode{
					Expr: &LiteralNode{Value: 42.0, Line: 1, Col: 4},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "string_literal",
			input: `{{ "hello" }}`,
			want: []Statement{
				&OutputNode{
					Expr: &LiteralNode{Value: "hello", Line: 1, Col: 4},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "boolean_literal",
			input: "{{ true }}",
			want: []Statement{
				&OutputNode{
					Expr: &LiteralNode{Value: true, Line: 1, Col: 4},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "property_access",
			input: "{{ user.name }}",
			want: []Statement{
				&OutputNode{
					Expr: &PropertyAccessNode{
						Object:   &VariableNode{Name: "user", Line: 1, Col: 4},
						Property: "name",
						Line:     1,
						Col:      8,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "subscript_access",
			input: "{{ items[0] }}",
			want: []Statement{
				&OutputNode{
					Expr: &SubscriptNode{
						Object: &VariableNode{Name: "items", Line: 1, Col: 4},
						Index:  &LiteralNode{Value: 0.0, Line: 1, Col: 10},
						Line:   1,
						Col:    9,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "filter",
			input: "{{ name|upper }}",
			want: []Statement{
				&OutputNode{
					Expr: &FilterNode{
						Expr: &VariableNode{Name: "name", Line: 1, Col: 4},
						Name: "upper",
						Args: nil,
						Line: 1,
						Col:  8,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "filter_with_arg",
			input: "{{ price|add:10 }}",
			want: []Statement{
				&OutputNode{
					Expr: &FilterNode{
						Expr: &VariableNode{Name: "price", Line: 1, Col: 4},
						Name: "add",
						Args: []Expression{
							&LiteralNode{Value: 10.0, Line: 1, Col: 14},
						},
						Line: 1,
						Col:  9,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "binary_operation",
			input: "{{ a + b }}",
			want: []Statement{
				&OutputNode{
					Expr: &BinaryOpNode{
						Operator: "+",
						Left:     &VariableNode{Name: "a", Line: 1, Col: 4},
						Right:    &VariableNode{Name: "b", Line: 1, Col: 8},
						Line:     1,
						Col:      6,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "comparison",
			input: "{{ x > 10 }}",
			want: []Statement{
				&OutputNode{
					Expr: &BinaryOpNode{
						Operator: ">",
						Left:     &VariableNode{Name: "x", Line: 1, Col: 4},
						Right:    &LiteralNode{Value: 10.0, Line: 1, Col: 8},
						Line:     1,
						Col:      6,
					},
					Line: 1,
					Col:  1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTemplate(t, tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.input, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Parse(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

// =============================================================================
// Test Group 3: If Tag Parsing
// =============================================================================

func TestParserIfTag(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Statement
	}{
		{
			name:  "simple_if",
			input: "{% if x %}yes{% endif %}",
			want: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&TextNode{Text: "yes", Line: 1, Col: 11},
							},
						},
					},
					Line: 1,
					Col:  4,
				},
			},
		},
		{
			name:  "if_else",
			input: "{% if x %}yes{% else %}no{% endif %}",
			want: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&TextNode{Text: "yes", Line: 1, Col: 11},
							},
						},
					},
					ElseBody: []Node{
						&TextNode{Text: "no", Line: 1, Col: 24},
					},
					Line: 1,
					Col:  4,
				},
			},
		},
		{
			name:  "if_elif",
			input: "{% if x %}a{% elif y %}b{% endif %}",
			want: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&TextNode{Text: "a", Line: 1, Col: 11},
							},
						},
						{
							Condition: &VariableNode{Name: "y", Line: 1, Col: 20},
							Body: []Node{
								&TextNode{Text: "b", Line: 1, Col: 24},
							},
						},
					},
					Line: 1,
					Col:  4,
				},
			},
		},
		{
			name:  "if_elif_else",
			input: "{% if x %}a{% elif y %}b{% else %}c{% endif %}",
			want: []Statement{
				&IfNode{
					Branches: []IfBranch{
						{
							Condition: &VariableNode{Name: "x", Line: 1, Col: 7},
							Body: []Node{
								&TextNode{Text: "a", Line: 1, Col: 11},
							},
						},
						{
							Condition: &VariableNode{Name: "y", Line: 1, Col: 20},
							Body: []Node{
								&TextNode{Text: "b", Line: 1, Col: 24},
							},
						},
					},
					ElseBody: []Node{
						&TextNode{Text: "c", Line: 1, Col: 35},
					},
					Line: 1,
					Col:  4,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTemplate(t, tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.input, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Parse(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

// =============================================================================
// Test Group 4: For Tag Parsing
// =============================================================================

func TestParserForTag(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Statement
	}{
		{
			name:  "simple_for",
			input: "{% for item in items %}{{ item }}{% endfor %}",
			want: []Statement{
				&ForNode{
					Vars:       []string{"item"},
					Collection: &VariableNode{Name: "items", Line: 1, Col: 16},
					Body: []Node{
						&OutputNode{
							Expr: &VariableNode{Name: "item", Line: 1, Col: 27},
							Line: 1,
							Col:  24,
						},
					},
					Line: 1,
					Col:  4,
				},
			},
		},
		{
			name:  "for_with_key_value",
			input: "{% for k, v in dict %}{{ k }}{% endfor %}",
			want: []Statement{
				&ForNode{
					Vars:       []string{"k", "v"},
					Collection: &VariableNode{Name: "dict", Line: 1, Col: 16},
					Body: []Node{
						&OutputNode{
							Expr: &VariableNode{Name: "k", Line: 1, Col: 26},
							Line: 1,
							Col:  23,
						},
					},
					Line: 1,
					Col:  4,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTemplate(t, tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.input, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Parse(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

// =============================================================================
// Test Group 5: Mixed Content Parsing
// =============================================================================

func TestParserMixedContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Statement
	}{
		{
			name:  "text_and_variable",
			input: "Hello {{ name }}!",
			want: []Statement{
				&TextNode{Text: "Hello ", Line: 1, Col: 1},
				&OutputNode{
					Expr: &VariableNode{Name: "name", Line: 1, Col: 10},
					Line: 1,
					Col:  7,
				},
				&TextNode{Text: "!", Line: 1, Col: 17},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTemplate(t, tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.input, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Parse(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

// =============================================================================
// Test Group 6: Error Cases
// =============================================================================

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "standalone_elif",
			input:   "{% elif x %}",
			wantErr: "elif must be used inside an if block",
		},
		{
			name:    "standalone_else",
			input:   "{% else %}",
			wantErr: "else must be used inside an if block",
		},
		{
			name:    "standalone_endif",
			input:   "{% endif %}",
			wantErr: "endif must match a corresponding if tag",
		},
		{
			name:    "standalone_endfor",
			input:   "{% endfor %}",
			wantErr: "endfor must match a corresponding for tag",
		},
		{
			name:    "unknown_tag",
			input:   "{% foobar %}",
			wantErr: "unknown tag: foobar",
		},
		{
			name:    "unclosed_if",
			input:   "{% if x %}yes",
			wantErr: "unexpected EOF",
		},
		{
			name:    "unclosed_for",
			input:   "{% for x in items %}body",
			wantErr: "unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTemplate(t, tt.input)
			if err == nil {
				t.Fatalf("Parse(%q) = nil error, want error containing %q", tt.input, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Parse(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Test Group 7: Parser Helpers
// =============================================================================

func TestParserHelpers(t *testing.T) {
	t.Run("notEOF_empty", func(t *testing.T) {
		p := NewParser(nil)
		if p.notEOF() {
			t.Error("notEOF() = true for empty parser, want false")
		}
	})

	t.Run("notEOF_with_eof", func(t *testing.T) {
		p := NewParser([]*Token{{Type: TokenEOF}})
		if p.notEOF() {
			t.Error("notEOF() = true for EOF token, want false")
		}
	})

	t.Run("notEOF_with_token", func(t *testing.T) {
		p := NewParser([]*Token{{Type: TokenText, Value: "hi"}})
		if !p.notEOF() {
			t.Error("notEOF() = false for text token, want true")
		}
	})

	t.Run("peek_negative_offset", func(t *testing.T) {
		p := NewParser([]*Token{{Type: TokenText, Value: "a"}})
		if got := p.peek(-1); got != nil {
			t.Errorf("peek(-1) = %v, want nil", got)
		}
	})

	t.Run("peek_beyond_end", func(t *testing.T) {
		p := NewParser([]*Token{{Type: TokenText, Value: "a"}})
		if got := p.peek(5); got != nil {
			t.Errorf("peek(5) = %v, want nil", got)
		}
	})

	t.Run("peek_valid_offset", func(t *testing.T) {
		tokens := []*Token{
			{Type: TokenText, Value: "a"},
			{Type: TokenText, Value: "b"},
		}
		p := NewParser(tokens)
		if got := p.peek(1); got == nil || got.Value != "b" {
			t.Errorf("peek(1) = %v, want token with value %q", got, "b")
		}
	})

	t.Run("remaining", func(t *testing.T) {
		tokens := []*Token{
			{Type: TokenText, Value: "a"},
			{Type: TokenText, Value: "b"},
			{Type: TokenEOF},
		}
		p := NewParser(tokens)
		if got, want := p.Remaining(), 3; got != want {
			t.Errorf("Remaining() = %d, want %d", got, want)
		}
		p.Advance()
		if got, want := p.Remaining(), 2; got != want {
			t.Errorf("Remaining() after Advance = %d, want %d", got, want)
		}
	})

	t.Run("advance_past_end", func(t *testing.T) {
		p := NewParser([]*Token{{Type: TokenText, Value: "a"}})
		p.Advance()
		p.Advance() // Should not panic.
		p.Advance() // Should not panic.
		if got, want := p.Remaining(), 0; got != want {
			t.Errorf("Remaining() after over-advance = %d, want %d", got, want)
		}
	})

	t.Run("current_past_end", func(t *testing.T) {
		p := NewParser([]*Token{{Type: TokenText, Value: "a"}})
		p.Advance()
		if got := p.Current(); got != nil {
			t.Errorf("Current() past end = %v, want nil", got)
		}
	})

	t.Run("match_success", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenSymbol, Value: ","},
		})
		tok := p.Match(TokenSymbol, ",")
		if tok == nil {
			t.Fatal("Match(TokenSymbol, \",\") = nil, want token")
		}
		if got, want := tok.Value, ","; got != want {
			t.Errorf("Match token value = %q, want %q", got, want)
		}
	})

	t.Run("match_failure_wrong_value", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenSymbol, Value: "."},
		})
		if tok := p.Match(TokenSymbol, ","); tok != nil {
			t.Errorf("Match(TokenSymbol, \",\") = %v, want nil", tok)
		}
	})

	t.Run("match_failure_wrong_type", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenIdentifier, Value: ","},
		})
		if tok := p.Match(TokenSymbol, ","); tok != nil {
			t.Errorf("Match wrong type = %v, want nil", tok)
		}
	})

	t.Run("match_nil_token", func(t *testing.T) {
		p := NewParser(nil)
		if tok := p.Match(TokenSymbol, ","); tok != nil {
			t.Errorf("Match on empty parser = %v, want nil", tok)
		}
	})

	t.Run("collectUntil_empty", func(t *testing.T) {
		p := NewParser(nil)
		got := p.collectUntil(TokenTagEnd)
		if got != nil {
			t.Errorf("collectUntil on empty parser = %v, want nil", got)
		}
	})

	t.Run("collectUntil_immediate_match", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenTagEnd, Value: "%}"},
		})
		got := p.collectUntil(TokenTagEnd)
		if got != nil {
			t.Errorf("collectUntil immediate match = %v, want nil", got)
		}
	})
}

// =============================================================================
// Test Group 8: newParseError helper
// =============================================================================

func TestNewParseError(t *testing.T) {
	pe := newParseError("test error", 5, 10)
	if pe.Message != "test error" {
		t.Errorf("ParseError.Message = %q, want %q", pe.Message, "test error")
	}
	if pe.Line != 5 {
		t.Errorf("ParseError.Line = %d, want %d", pe.Line, 5)
	}
	if pe.Col != 10 {
		t.Errorf("ParseError.Col = %d, want %d", pe.Col, 10)
	}
}

// =============================================================================
// Test Group 9: ExpectIdentifier
// =============================================================================

func TestParserExpectIdentifier(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenIdentifier, Value: "foo", Line: 1, Col: 1},
		})
		tok, err := p.ExpectIdentifier()
		if err != nil {
			t.Fatalf("ExpectIdentifier() returned unexpected error: %v", err)
		}
		if tok.Value != "foo" {
			t.Errorf("ExpectIdentifier() value = %q, want %q", tok.Value, "foo")
		}
	})

	t.Run("wrong_type", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenNumber, Value: "42", Line: 1, Col: 1},
		})
		_, err := p.ExpectIdentifier()
		if err == nil {
			t.Fatal("ExpectIdentifier() = nil error, want error")
		}
		if !strings.Contains(err.Error(), "expected") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "expected")
		}
	})

	t.Run("empty_parser", func(t *testing.T) {
		p := NewParser(nil)
		_, err := p.ExpectIdentifier()
		if err == nil {
			t.Fatal("ExpectIdentifier() on empty = nil error, want error")
		}
		if !strings.Contains(err.Error(), "unexpected end of input") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "unexpected end of input")
		}
	})
}

// =============================================================================
// Test Group 10: Error and Errorf helpers
// =============================================================================

func TestParserErrorHelpers(t *testing.T) {
	t.Run("error_with_token", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenText, Value: "x", Line: 3, Col: 7},
		})
		err := p.Error("something wrong")
		if !strings.Contains(err.Error(), "line 3") {
			t.Errorf("Error() = %q, want it to contain position info", err.Error())
		}
	})

	t.Run("error_without_token", func(t *testing.T) {
		p := NewParser(nil)
		err := p.Error("something wrong")
		if !strings.Contains(err.Error(), "something wrong") {
			t.Errorf("Error() = %q, want it to contain message", err.Error())
		}
	})

	t.Run("errorf_formatting", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenText, Value: "x", Line: 1, Col: 1},
		})
		err := p.Errorf("got %s, want %s", "a", "b")
		if !strings.Contains(err.Error(), "got a, want b") {
			t.Errorf("Errorf() = %q, want formatted message", err.Error())
		}
	})
}

// =============================================================================
// Test Group 11: convertStatementsToNodes
// =============================================================================

func TestConvertStatementsToNodes(t *testing.T) {
	t.Run("nil_input", func(t *testing.T) {
		got := convertStatementsToNodes(nil)
		if got != nil {
			t.Errorf("convertStatementsToNodes(nil) = %v, want nil", got)
		}
	})

	t.Run("empty_input", func(t *testing.T) {
		got := convertStatementsToNodes([]Statement{})
		if got != nil {
			t.Errorf("convertStatementsToNodes([]) = %v, want nil", got)
		}
	})

	t.Run("with_statements", func(t *testing.T) {
		stmts := []Statement{
			&TextNode{Text: "a", Line: 1, Col: 1},
			&TextNode{Text: "b", Line: 1, Col: 2},
		}
		got := convertStatementsToNodes(stmts)
		if len(got) != 2 {
			t.Fatalf("convertStatementsToNodes() len = %d, want 2", len(got))
		}
		if got[0].String() != `Text("a")` {
			t.Errorf("node[0] = %s, want Text(\"a\")", got[0])
		}
	})
}

// =============================================================================
// Test Group 12: isEndTag and endTagName
// =============================================================================

func TestParserEndTagHelpers(t *testing.T) {
	t.Run("isEndTag_match", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenTagBegin, Value: "{%"},
			{Type: TokenIdentifier, Value: "endif"},
			{Type: TokenTagEnd, Value: "%}"},
		})
		if !p.isEndTag("endif", "else") {
			t.Error("isEndTag(endif, else) = false, want true")
		}
	})

	t.Run("isEndTag_no_match", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenTagBegin, Value: "{%"},
			{Type: TokenIdentifier, Value: "for"},
			{Type: TokenTagEnd, Value: "%}"},
		})
		if p.isEndTag("endif", "else") {
			t.Error("isEndTag(endif, else) = true, want false")
		}
	})

	t.Run("isEndTag_not_tag_begin", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenText, Value: "hello"},
		})
		if p.isEndTag("endif") {
			t.Error("isEndTag on text token = true, want false")
		}
	})

	t.Run("isEndTag_no_identifier_after_begin", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenTagBegin, Value: "{%"},
			{Type: TokenTagEnd, Value: "%}"},
		})
		if p.isEndTag("endif") {
			t.Error("isEndTag with no identifier = true, want false")
		}
	})

	t.Run("endTagName_valid", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenTagBegin, Value: "{%"},
			{Type: TokenIdentifier, Value: "endif"},
		})
		if got, want := p.endTagName(), "endif"; got != want {
			t.Errorf("endTagName() = %q, want %q", got, want)
		}
	})

	t.Run("endTagName_no_identifier", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenTagBegin, Value: "{%"},
		})
		if got := p.endTagName(); got != "" {
			t.Errorf("endTagName() = %q, want empty", got)
		}
	})
}

// =============================================================================
// Test Group 13: parseNext with unexpected token types
// =============================================================================

func TestParserUnexpectedToken(t *testing.T) {
	tokens := []*Token{
		{Type: TokenVarEnd, Value: "}}", Line: 1, Col: 1},
		{Type: TokenEOF, Line: 1, Col: 3},
	}
	p := NewParser(tokens)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("Parse with unexpected token = nil error, want error")
	}
	if !strings.Contains(err.Error(), "unexpected token") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "unexpected token")
	}
}

// =============================================================================
// Test Group 14: ParseExpression on Parser
// =============================================================================

func TestParserParseExpression(t *testing.T) {
	tokens := []*Token{
		{Type: TokenIdentifier, Value: "x", Line: 1, Col: 1},
		{Type: TokenEOF, Line: 1, Col: 2},
	}
	p := NewParser(tokens)
	expr, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("ParseExpression() returned unexpected error: %v", err)
	}
	v, ok := expr.(*VariableNode)
	if !ok {
		t.Fatalf("ParseExpression() = %T, want *VariableNode", expr)
	}
	if v.Name != "x" {
		t.Errorf("variable name = %q, want %q", v.Name, "x")
	}
}
