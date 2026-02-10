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

	t.Run("match_failure", func(t *testing.T) {
		p := NewParser([]*Token{
			{Type: TokenSymbol, Value: "."},
		})
		if tok := p.Match(TokenSymbol, ","); tok != nil {
			t.Errorf("Match(TokenSymbol, \",\") = %v, want nil", tok)
		}
	})

	t.Run("collectUntil_empty", func(t *testing.T) {
		p := NewParser(nil)
		got := p.collectUntil(TokenTagEnd)
		if got != nil {
			t.Errorf("collectUntil on empty parser = %v, want nil", got)
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
