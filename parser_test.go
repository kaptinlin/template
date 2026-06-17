package template

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// parseTemplate is a test helper that tokenizes and parses the input.
func parseTemplate(t *testing.T, input string) ([]statement, error) {
	t.Helper()
	tokens, err := newLexer(input).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize(%q) returned unexpected error: %v", input, err)
	}
	return newParser(tokens).Parse()
}

// =============================================================================
// Test Group 1: Basic Text node Parsing
// =============================================================================

func TestParserTextNode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []statement
	}{
		{
			name:  "plain_text",
			input: "Hello, World!",
			want: []statement{
				&textNode{Text: "Hello, World!", Line: 1, Col: 1},
			},
		},
		{
			name:  "multiline_text",
			input: "Line 1\nLine 2\nLine 3",
			want: []statement{
				&textNode{Text: "Line 1\nLine 2\nLine 3", Line: 1, Col: 1},
			},
		},
		{
			name:  "special_characters",
			input: "Hello <html> & \"quotes\"",
			want: []statement{
				&textNode{Text: "Hello <html> & \"quotes\"", Line: 1, Col: 1},
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
// Test Group 2: Output node (Variable) Parsing
// =============================================================================

func TestParserOutputNode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []statement
	}{
		{
			name:  "simple_variable",
			input: "{{ name }}",
			want: []statement{
				&outputNode{
					Expr: &variableNode{Name: "name", Line: 1, Col: 4},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "number_literal",
			input: "{{ 42 }}",
			want: []statement{
				&outputNode{
					Expr: &literalNode{Value: 42.0, Line: 1, Col: 4},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "string_literal",
			input: `{{ "hello" }}`,
			want: []statement{
				&outputNode{
					Expr: &literalNode{Value: "hello", Line: 1, Col: 4},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "boolean_literal",
			input: "{{ true }}",
			want: []statement{
				&outputNode{
					Expr: &literalNode{Value: true, Line: 1, Col: 4},
					Line: 1,
					Col:  1,
				},
			},
		},
		{
			name:  "property_access",
			input: "{{ user.name }}",
			want: []statement{
				&outputNode{
					Expr: &propertyAccessNode{
						Object:   &variableNode{Name: "user", Line: 1, Col: 4},
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
			want: []statement{
				&outputNode{
					Expr: &subscriptNode{
						Object: &variableNode{Name: "items", Line: 1, Col: 4},
						Index:  &literalNode{Value: 0.0, Line: 1, Col: 10},
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
			want: []statement{
				&outputNode{
					Expr: &filterNode{
						Expr: &variableNode{Name: "name", Line: 1, Col: 4},
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
			want: []statement{
				&outputNode{
					Expr: &filterNode{
						Expr: &variableNode{Name: "price", Line: 1, Col: 4},
						Name: "add",
						Args: []expression{
							&literalNode{Value: 10.0, Line: 1, Col: 14},
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
			want: []statement{
				&outputNode{
					Expr: &binaryOpNode{
						Operator: "+",
						Left:     &variableNode{Name: "a", Line: 1, Col: 4},
						Right:    &variableNode{Name: "b", Line: 1, Col: 8},
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
			want: []statement{
				&outputNode{
					Expr: &binaryOpNode{
						Operator: ">",
						Left:     &variableNode{Name: "x", Line: 1, Col: 4},
						Right:    &literalNode{Value: 10.0, Line: 1, Col: 8},
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
		want  []statement
	}{
		{
			name:  "simple_if",
			input: "{% if x %}yes{% endif %}",
			want: []statement{
				&ifNode{
					Branches: []ifBranch{
						{
							Condition: &variableNode{Name: "x", Line: 1, Col: 7},
							Body: []node{
								&textNode{Text: "yes", Line: 1, Col: 11},
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
			want: []statement{
				&ifNode{
					Branches: []ifBranch{
						{
							Condition: &variableNode{Name: "x", Line: 1, Col: 7},
							Body: []node{
								&textNode{Text: "yes", Line: 1, Col: 11},
							},
						},
					},
					ElseBody: []node{
						&textNode{Text: "no", Line: 1, Col: 24},
					},
					Line: 1,
					Col:  4,
				},
			},
		},
		{
			name:  "if_elif",
			input: "{% if x %}a{% elif y %}b{% endif %}",
			want: []statement{
				&ifNode{
					Branches: []ifBranch{
						{
							Condition: &variableNode{Name: "x", Line: 1, Col: 7},
							Body: []node{
								&textNode{Text: "a", Line: 1, Col: 11},
							},
						},
						{
							Condition: &variableNode{Name: "y", Line: 1, Col: 20},
							Body: []node{
								&textNode{Text: "b", Line: 1, Col: 24},
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
			want: []statement{
				&ifNode{
					Branches: []ifBranch{
						{
							Condition: &variableNode{Name: "x", Line: 1, Col: 7},
							Body: []node{
								&textNode{Text: "a", Line: 1, Col: 11},
							},
						},
						{
							Condition: &variableNode{Name: "y", Line: 1, Col: 20},
							Body: []node{
								&textNode{Text: "b", Line: 1, Col: 24},
							},
						},
					},
					ElseBody: []node{
						&textNode{Text: "c", Line: 1, Col: 35},
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
		want  []statement
	}{
		{
			name:  "simple_for",
			input: "{% for item in items %}{{ item }}{% endfor %}",
			want: []statement{
				&forNode{
					Vars:       []string{"item"},
					Collection: &variableNode{Name: "items", Line: 1, Col: 16},
					Body: []node{
						&outputNode{
							Expr: &variableNode{Name: "item", Line: 1, Col: 27},
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
			want: []statement{
				&forNode{
					Vars:       []string{"k", "v"},
					Collection: &variableNode{Name: "dict", Line: 1, Col: 16},
					Body: []node{
						&outputNode{
							Expr: &variableNode{Name: "k", Line: 1, Col: 26},
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
		want  []statement
	}{
		{
			name:  "text_and_variable",
			input: "Hello {{ name }}!",
			want: []statement{
				&textNode{Text: "Hello ", Line: 1, Col: 1},
				&outputNode{
					Expr: &variableNode{Name: "name", Line: 1, Col: 10},
					Line: 1,
					Col:  7,
				},
				&textNode{Text: "!", Line: 1, Col: 17},
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
// Test Group 7: parser Helpers
// =============================================================================

func TestParserHelpers(t *testing.T) {
	t.Run("notEOF_empty", func(t *testing.T) {
		p := newParser(nil)
		if p.notEOF() {
			t.Error("notEOF() = true for empty parser, want false")
		}
	})

	t.Run("notEOF_with_eof", func(t *testing.T) {
		p := newParser([]*token{{Type: tokenEOF}})
		if p.notEOF() {
			t.Error("notEOF() = true for EOF token, want false")
		}
	})

	t.Run("notEOF_with_token", func(t *testing.T) {
		p := newParser([]*token{{Type: tokenText, value: "hi"}})
		if !p.notEOF() {
			t.Error("notEOF() = false for text token, want true")
		}
	})

	t.Run("peek_negative_offset", func(t *testing.T) {
		p := newParser([]*token{{Type: tokenText, value: "a"}})
		if got := p.peek(-1); got != nil {
			t.Errorf("peek(-1) = %v, want nil", got)
		}
	})

	t.Run("peek_beyond_end", func(t *testing.T) {
		p := newParser([]*token{{Type: tokenText, value: "a"}})
		if got := p.peek(5); got != nil {
			t.Errorf("peek(5) = %v, want nil", got)
		}
	})

	t.Run("peek_valid_offset", func(t *testing.T) {
		tokens := []*token{
			{Type: tokenText, value: "a"},
			{Type: tokenText, value: "b"},
		}
		p := newParser(tokens)
		if got := p.peek(1); got == nil || got.value != "b" {
			t.Errorf("peek(1) = %v, want token with value %q", got, "b")
		}
	})

	t.Run("remaining", func(t *testing.T) {
		tokens := []*token{
			{Type: tokenText, value: "a"},
			{Type: tokenText, value: "b"},
			{Type: tokenEOF},
		}
		p := newParser(tokens)
		if got, want := p.Remaining(), 3; got != want {
			t.Errorf("Remaining() = %d, want %d", got, want)
		}
		p.Advance()
		if got, want := p.Remaining(), 2; got != want {
			t.Errorf("Remaining() after Advance = %d, want %d", got, want)
		}
	})

	t.Run("advance_past_end", func(t *testing.T) {
		p := newParser([]*token{{Type: tokenText, value: "a"}})
		p.Advance()
		p.Advance() // Should not panic.
		p.Advance() // Should not panic.
		if got, want := p.Remaining(), 0; got != want {
			t.Errorf("Remaining() after over-advance = %d, want %d", got, want)
		}
	})

	t.Run("current_past_end", func(t *testing.T) {
		p := newParser([]*token{{Type: tokenText, value: "a"}})
		p.Advance()
		if got := p.Current(); got != nil {
			t.Errorf("Current() past end = %v, want nil", got)
		}
	})

	t.Run("match_success", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenSymbol, value: ","},
		})
		tok := p.Match(tokenSymbol, ",")
		if tok == nil {
			t.Fatal("Match(tokenSymbol, \",\") = nil, want token")
		}
		if got, want := tok.value, ","; got != want {
			t.Errorf("Match token value = %q, want %q", got, want)
		}
	})

	t.Run("match_failure_wrong_value", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenSymbol, value: "."},
		})
		if tok := p.Match(tokenSymbol, ","); tok != nil {
			t.Errorf("Match(tokenSymbol, \",\") = %v, want nil", tok)
		}
	})

	t.Run("match_failure_wrong_type", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenIdentifier, value: ","},
		})
		if tok := p.Match(tokenSymbol, ","); tok != nil {
			t.Errorf("Match wrong type = %v, want nil", tok)
		}
	})

	t.Run("match_nil_token", func(t *testing.T) {
		p := newParser(nil)
		if tok := p.Match(tokenSymbol, ","); tok != nil {
			t.Errorf("Match on empty parser = %v, want nil", tok)
		}
	})

	t.Run("collectUntil_empty", func(t *testing.T) {
		p := newParser(nil)
		got := p.collectUntil(tokenTagEnd)
		if got != nil {
			t.Errorf("collectUntil on empty parser = %v, want nil", got)
		}
	})

	t.Run("collectUntil_immediate_match", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenTagEnd, value: "%}"},
		})
		got := p.collectUntil(tokenTagEnd)
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
		p := newParser([]*token{
			{Type: tokenIdentifier, value: "foo", Line: 1, Col: 1},
		})
		tok, err := p.ExpectIdentifier()
		if err != nil {
			t.Fatalf("ExpectIdentifier() returned unexpected error: %v", err)
		}
		if tok.value != "foo" {
			t.Errorf("ExpectIdentifier() value = %q, want %q", tok.value, "foo")
		}
	})

	t.Run("wrong_type", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenNumber, value: "42", Line: 1, Col: 1},
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
		p := newParser(nil)
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
		p := newParser([]*token{
			{Type: tokenText, value: "x", Line: 3, Col: 7},
		})
		err := p.Error("something wrong")
		if !strings.Contains(err.Error(), "line 3") {
			t.Errorf("Error() = %q, want it to contain position info", err.Error())
		}
	})

	t.Run("error_without_token", func(t *testing.T) {
		p := newParser(nil)
		err := p.Error("something wrong")
		if !strings.Contains(err.Error(), "something wrong") {
			t.Errorf("Error() = %q, want it to contain message", err.Error())
		}
	})

	t.Run("errorf_formatting", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenText, value: "x", Line: 1, Col: 1},
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
		got := convertStatementsToNodes([]statement{})
		if got != nil {
			t.Errorf("convertStatementsToNodes([]) = %v, want nil", got)
		}
	})

	t.Run("with_statements", func(t *testing.T) {
		stmts := []statement{
			&textNode{Text: "a", Line: 1, Col: 1},
			&textNode{Text: "b", Line: 1, Col: 2},
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
		p := newParser([]*token{
			{Type: tokenTagBegin, value: "{%"},
			{Type: tokenIdentifier, value: "endif"},
			{Type: tokenTagEnd, value: "%}"},
		})
		if !p.isEndTag("endif", "else") {
			t.Error("isEndTag(endif, else) = false, want true")
		}
	})

	t.Run("isEndTag_no_match", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenTagBegin, value: "{%"},
			{Type: tokenIdentifier, value: "for"},
			{Type: tokenTagEnd, value: "%}"},
		})
		if p.isEndTag("endif", "else") {
			t.Error("isEndTag(endif, else) = true, want false")
		}
	})

	t.Run("isEndTag_not_tag_begin", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenText, value: "hello"},
		})
		if p.isEndTag("endif") {
			t.Error("isEndTag on text token = true, want false")
		}
	})

	t.Run("isEndTag_no_identifier_after_begin", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenTagBegin, value: "{%"},
			{Type: tokenTagEnd, value: "%}"},
		})
		if p.isEndTag("endif") {
			t.Error("isEndTag with no identifier = true, want false")
		}
	})

	t.Run("endTagName_valid", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenTagBegin, value: "{%"},
			{Type: tokenIdentifier, value: "endif"},
		})
		if got, want := p.endTagName(), "endif"; got != want {
			t.Errorf("endTagName() = %q, want %q", got, want)
		}
	})

	t.Run("endTagName_no_identifier", func(t *testing.T) {
		p := newParser([]*token{
			{Type: tokenTagBegin, value: "{%"},
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
	tokens := []*token{
		{Type: tokenVarEnd, value: "}}", Line: 1, Col: 1},
		{Type: tokenEOF, Line: 1, Col: 3},
	}
	p := newParser(tokens)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("Parse with unexpected token = nil error, want error")
	}
	if !strings.Contains(err.Error(), "unexpected token") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "unexpected token")
	}
}

// =============================================================================
// Test Group 14: ParseExpression on parser
// =============================================================================

func TestParserParseExpression(t *testing.T) {
	tokens := []*token{
		{Type: tokenIdentifier, value: "x", Line: 1, Col: 1},
		{Type: tokenEOF, Line: 1, Col: 2},
	}
	p := newParser(tokens)
	expr, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("ParseExpression() returned unexpected error: %v", err)
	}
	v, ok := expr.(*variableNode)
	if !ok {
		t.Fatalf("ParseExpression() = %T, want *variableNode", expr)
	}
	if v.Name != "x" {
		t.Errorf("variable name = %q, want %q", v.Name, "x")
	}
}

// =============================================================================
// Test Group 15: parser Edge Cases (coverage gaps)
// =============================================================================

func TestParserParseUntil(t *testing.T) {
	// ParseUntil is public API but unused by built-in tags (they use ParseUntilWithArgs).
	t.Run("basic ParseUntil", func(t *testing.T) {
		tokens, err := newLexer("hello{% endif %}").Tokenize()
		if err != nil {
			t.Fatal(err)
		}
		p := newParser(tokens)
		nodes, tag, err := p.ParseUntil("endif")
		if err != nil {
			t.Fatalf("ParseUntil() error = %v", err)
		}
		if tag != "endif" {
			t.Errorf("tag = %q, want %q", tag, "endif")
		}
		if len(nodes) != 1 {
			t.Errorf("len(nodes) = %d, want 1", len(nodes))
		}
	})

	t.Run("ParseUntil EOF", func(t *testing.T) {
		tokens, err := newLexer("hello world").Tokenize()
		if err != nil {
			t.Fatal(err)
		}
		p := newParser(tokens)
		_, _, err = p.ParseUntil("endif")
		if err == nil {
			t.Fatal("expected error for unexpected EOF")
		}
		if !strings.Contains(err.Error(), "unexpected EOF") {
			t.Errorf("error = %q, want 'unexpected EOF'", err.Error())
		}
	})
}

func TestParserParseNextEdgeCases(t *testing.T) {
	// Test the exhaustive switch branches for unexpected token types.
	unexpectedTypes := []struct {
		name string
		tok  *token
	}{
		{"tokenError", &token{Type: tokenError, value: "err", Line: 1, Col: 1}},
		{"tokenVarEnd", &token{Type: tokenVarEnd, value: "}}", Line: 1, Col: 1}},
		{"tokenTagEnd", &token{Type: tokenTagEnd, value: "%}", Line: 1, Col: 1}},
		{"tokenIdentifier", &token{Type: tokenIdentifier, value: "x", Line: 1, Col: 1}},
		{"tokenString", &token{Type: tokenString, value: "hello", Line: 1, Col: 1}},
		{"tokenNumber", &token{Type: tokenNumber, value: "42", Line: 1, Col: 1}},
		{"tokenSymbol", &token{Type: tokenSymbol, value: "+", Line: 1, Col: 1}},
	}

	for _, tt := range unexpectedTypes {
		t.Run(tt.name, func(t *testing.T) {
			tokens := []*token{tt.tok, {Type: tokenEOF, Line: 1, Col: 2}}
			p := newParser(tokens)
			_, err := p.Parse()
			if err == nil {
				t.Fatalf("expected error for %s as first token", tt.name)
			}
			if !strings.Contains(err.Error(), "unexpected token") {
				t.Errorf("error = %q, want 'unexpected token'", err.Error())
			}
		})
	}
}

func TestParserParseVariableEdgeCases(t *testing.T) {
	t.Run("empty variable expression", func(t *testing.T) {
		// {{ }}
		tokens := []*token{
			{Type: tokenVarBegin, value: "{{", Line: 1, Col: 1},
			{Type: tokenVarEnd, value: "}}", Line: 1, Col: 4},
			{Type: tokenEOF, Line: 1, Col: 6},
		}
		p := newParser(tokens)
		_, err := p.Parse()
		if err == nil {
			t.Fatal("expected error for empty variable expression")
		}
		if !strings.Contains(err.Error(), "empty variable expression") {
			t.Errorf("error = %q, want 'empty variable expression'", err.Error())
		}
	})

	t.Run("missing VarEnd", func(t *testing.T) {
		// {{ x <EOF>
		tokens := []*token{
			{Type: tokenVarBegin, value: "{{", Line: 1, Col: 1},
			{Type: tokenIdentifier, value: "x", Line: 1, Col: 4},
			{Type: tokenEOF, Line: 1, Col: 5},
		}
		p := newParser(tokens)
		_, err := p.Parse()
		if err == nil {
			t.Fatal("expected error for missing }}")
		}
		if !strings.Contains(err.Error(), "expected }}") {
			t.Errorf("error = %q, want 'expected }}'", err.Error())
		}
	})

	t.Run("expression parse error in variable", func(t *testing.T) {
		// {{ ]] }}
		tokens := []*token{
			{Type: tokenVarBegin, value: "{{", Line: 1, Col: 1},
			{Type: tokenSymbol, value: "]", Line: 1, Col: 4},
			{Type: tokenVarEnd, value: "}}", Line: 1, Col: 5},
			{Type: tokenEOF, Line: 1, Col: 7},
		}
		p := newParser(tokens)
		_, err := p.Parse()
		if err == nil {
			t.Fatal("expected error for invalid expression")
		}
	})
}

func TestParserParseExpressionError(t *testing.T) {
	// Test error propagation in parser.ParseExpression.
	tokens := []*token{
		{Type: tokenVarEnd, value: "}}", Line: 1, Col: 1},
		{Type: tokenEOF, Line: 1, Col: 3},
	}
	p := newParser(tokens)
	_, err := p.ParseExpression()
	if err == nil {
		t.Fatal("expected error from ParseExpression")
	}
}

func TestParserConsumeEndTagMissingClose(t *testing.T) {
	// Test consumeEndTag when %} is missing.
	// Construct: {% endif <EOF>
	tokens := []*token{
		{Type: tokenTagBegin, value: "{%", Line: 1, Col: 1},
		{Type: tokenIdentifier, value: "endif", Line: 1, Col: 4},
		{Type: tokenEOF, Line: 1, Col: 9},
	}
	p := newParser(tokens)
	// Manually call ParseUntilWithArgs which calls consumeEndTag.
	// But we need isEndTag to match first. Let's use ParseUntil instead.
	_, tag, err := p.ParseUntil("endif")
	if err != nil {
		// ParseUntil returns tag name from endTagName without consuming.
		// It just peeks, so it returns successfully but doesn't consume.
		// Actually ParseUntil checks isEndTag and returns immediately.
		// It returns the endTagName which is "endif".
		t.Fatalf("ParseUntil() unexpected error: %v", err)
	}
	if tag != "endif" {
		t.Errorf("tag = %q, want %q", tag, "endif")
	}
}

func TestParserParseTagEdgeCases(t *testing.T) {
	t.Run("missing tag name", func(t *testing.T) {
		// {% %} — no identifier after {%
		tokens := []*token{
			{Type: tokenTagBegin, value: "{%", Line: 1, Col: 1},
			{Type: tokenTagEnd, value: "%}", Line: 1, Col: 4},
			{Type: tokenEOF, Line: 1, Col: 6},
		}
		p := newParser(tokens)
		_, err := p.Parse()
		if err == nil {
			t.Fatal("expected error for missing tag name")
		}
		if !strings.Contains(err.Error(), "expected tag name") {
			t.Errorf("error = %q, want 'expected tag name'", err.Error())
		}
	})

	t.Run("unknown tag", func(t *testing.T) {
		// {% foobar %}
		tokens := []*token{
			{Type: tokenTagBegin, value: "{%", Line: 1, Col: 1},
			{Type: tokenIdentifier, value: "foobar", Line: 1, Col: 4},
			{Type: tokenTagEnd, value: "%}", Line: 1, Col: 11},
			{Type: tokenEOF, Line: 1, Col: 13},
		}
		p := newParser(tokens)
		_, err := p.Parse()
		if err == nil {
			t.Fatal("expected error for unknown tag")
		}
		if !strings.Contains(err.Error(), "unknown tag: foobar") {
			t.Errorf("error = %q, want 'unknown tag: foobar'", err.Error())
		}
	})

	t.Run("misused endif standalone", func(t *testing.T) {
		// {% endif %} at top level
		tokens := []*token{
			{Type: tokenTagBegin, value: "{%", Line: 1, Col: 1},
			{Type: tokenIdentifier, value: "endif", Line: 1, Col: 4},
			{Type: tokenTagEnd, value: "%}", Line: 1, Col: 10},
			{Type: tokenEOF, Line: 1, Col: 12},
		}
		p := newParser(tokens)
		_, err := p.Parse()
		if err == nil {
			t.Fatal("expected error for misused endif")
		}
		if !strings.Contains(err.Error(), "endif must match") {
			t.Errorf("error = %q, want hint about endif", err.Error())
		}
	})

	t.Run("missing TagEnd", func(t *testing.T) {
		// {% if x <EOF>
		tokens := []*token{
			{Type: tokenTagBegin, value: "{%", Line: 1, Col: 1},
			{Type: tokenIdentifier, value: "if", Line: 1, Col: 4},
			{Type: tokenIdentifier, value: "x", Line: 1, Col: 7},
			{Type: tokenEOF, Line: 1, Col: 8},
		}
		p := newParser(tokens)
		_, err := p.Parse()
		if err == nil {
			t.Fatal("expected error for missing %}")
		}
		if !strings.Contains(err.Error(), "expected %}") {
			t.Errorf("error = %q, want 'expected %%}'", err.Error())
		}
	})
}
