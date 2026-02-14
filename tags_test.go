package template

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// saveTagRegistry returns a snapshot of the current tag registry.
func saveTagRegistry() map[string]TagParser {
	tagMu.RLock()
	defer tagMu.RUnlock()
	saved := make(map[string]TagParser, len(tagRegistry))
	maps.Copy(saved, tagRegistry)
	return saved
}

// restoreTagRegistry restores the tag registry from a snapshot.
func restoreTagRegistry(saved map[string]TagParser) {
	tagMu.Lock()
	defer tagMu.Unlock()
	tagRegistry = saved
}

func TestRegisterTag(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	tests := []struct {
		name    string
		setup   func()
		tagName string
		parser  TagParser
		wantErr error
	}{
		{
			name: "new tag",
			setup: func() {
				tagRegistry = make(map[string]TagParser)
			},
			tagName: "customtag",
			parser:  parseIfTag,
		},
		{
			name: "duplicate tag",
			setup: func() {
				tagRegistry = make(map[string]TagParser)
				tagRegistry["duplicate"] = parseIfTag
			},
			tagName: "duplicate",
			parser:  parseForTag,
			wantErr: ErrTagAlreadyRegistered,
		},
		{
			name: "multiple different tags",
			setup: func() {
				tagRegistry = make(map[string]TagParser)
			},
			tagName: "tag1",
			parser:  parseIfTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := RegisterTag(tt.tagName, tt.parser)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("RegisterTag(%q) err = %v, want %v", tt.tagName, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("RegisterTag(%q) = %v, want nil", tt.tagName, err)
			}
			if _, ok := tagRegistry[tt.tagName]; !ok {
				t.Errorf("RegisterTag(%q) did not register tag", tt.tagName)
			}
		})
	}
}

func TestTag(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	tests := []struct {
		name    string
		setup   func()
		tagName string
		want    bool
	}{
		{
			name: "existing tag",
			setup: func() {
				tagRegistry = make(map[string]TagParser)
				tagRegistry["if"] = parseIfTag
			},
			tagName: "if",
			want:    true,
		},
		{
			name: "non-existing tag",
			setup: func() {
				tagRegistry = make(map[string]TagParser)
			},
			tagName: "nonexistent",
		},
		{
			name: "built-in for tag",
			setup: func() {
				tagRegistry = make(map[string]TagParser)
				tagRegistry["for"] = parseForTag
			},
			tagName: "for",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			p, ok := Tag(tt.tagName)
			if ok != tt.want {
				t.Errorf("Tag(%q) ok = %v, want %v", tt.tagName, ok, tt.want)
			}
			if tt.want && p == nil {
				t.Errorf("Tag(%q) parser = nil, want non-nil", tt.tagName)
			}
			if !tt.want && p != nil {
				t.Errorf("Tag(%q) parser = non-nil, want nil", tt.tagName)
			}
		})
	}
}

func TestParseIfTag(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        string
		wantErr     bool
		errContains string
		validate    func(*testing.T, Statement)
	}{
		{
			name: "simple if",
			tmpl: `{% if x %}yes{% endif %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*IfNode)
				if !ok {
					t.Fatalf("got %T, want *IfNode", s)
				}
				if got := len(n.Branches); got != 1 {
					t.Errorf("len(Branches) = %d, want 1", got)
				}
				if got := len(n.ElseBody); got != 0 {
					t.Errorf("len(ElseBody) = %d, want 0", got)
				}
				if _, ok := n.Branches[0].Condition.(*VariableNode); !ok {
					t.Errorf("condition type = %T, want *VariableNode", n.Branches[0].Condition)
				}
				if got := len(n.Branches[0].Body); got != 1 {
					t.Errorf("len(Body) = %d, want 1", got)
				}
			},
		},
		{
			name: "if-else",
			tmpl: `{% if x %}yes{% else %}no{% endif %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*IfNode)
				if !ok {
					t.Fatalf("got %T, want *IfNode", s)
				}
				if got := len(n.Branches); got != 1 {
					t.Errorf("len(Branches) = %d, want 1", got)
				}
				if got := len(n.ElseBody); got != 1 {
					t.Errorf("len(ElseBody) = %d, want 1", got)
				}
			},
		},
		{
			name: "if-elif-else",
			tmpl: `{% if x > 5 %}big{% elif x > 0 %}small{% else %}zero{% endif %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*IfNode)
				if !ok {
					t.Fatalf("got %T, want *IfNode", s)
				}
				if got := len(n.Branches); got != 2 {
					t.Errorf("len(Branches) = %d, want 2", got)
				}
				if got := len(n.ElseBody); got != 1 {
					t.Errorf("len(ElseBody) = %d, want 1", got)
				}
			},
		},
		{
			name: "multiple elif",
			tmpl: `{% if x == 1 %}one{% elif x == 2 %}two{% elif x == 3 %}three{% endif %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*IfNode)
				if !ok {
					t.Fatalf("got %T, want *IfNode", s)
				}
				if got := len(n.Branches); got != 3 {
					t.Errorf("len(Branches) = %d, want 3", got)
				}
				if got := len(n.ElseBody); got != 0 {
					t.Errorf("len(ElseBody) = %d, want 0", got)
				}
			},
		},
		{
			name: "if with comparison",
			tmpl: `{% if count > 10 %}many{% endif %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*IfNode)
				if !ok {
					t.Fatalf("got %T, want *IfNode", s)
				}
				if got := len(n.Branches); got != 1 {
					t.Errorf("len(Branches) = %d, want 1", got)
				}
				if _, ok := n.Branches[0].Condition.(*BinaryOpNode); !ok {
					t.Errorf("condition type = %T, want *BinaryOpNode", n.Branches[0].Condition)
				}
			},
		},
		{
			name: "if with logical and",
			tmpl: `{% if x > 0 and x < 10 %}valid{% endif %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*IfNode)
				if !ok {
					t.Fatalf("got %T, want *IfNode", s)
				}
				if got := len(n.Branches); got != 1 {
					t.Errorf("len(Branches) = %d, want 1", got)
				}
				if _, ok := n.Branches[0].Condition.(*BinaryOpNode); !ok {
					t.Errorf("condition type = %T, want *BinaryOpNode", n.Branches[0].Condition)
				}
			},
		},
		{
			name: "empty if body",
			tmpl: `{% if x %}{% endif %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*IfNode)
				if !ok {
					t.Fatalf("got %T, want *IfNode", s)
				}
				if got := len(n.Branches); got != 1 {
					t.Errorf("len(Branches) = %d, want 1", got)
				}
				if got := len(n.Branches[0].Body); got != 0 {
					t.Errorf("len(Body) = %d, want 0", got)
				}
			},
		},
		{
			name:        "extra tokens after condition",
			tmpl:        `{% if x y %}yes{% endif %}`,
			wantErr:     true,
			errContains: ErrUnexpectedTokensAfterCondition.Error(),
		},
		{
			name:        "else with arguments",
			tmpl:        `{% if x %}yes{% else x %}no{% endif %}`,
			wantErr:     true,
			errContains: ErrElseNoArgs.Error(),
		},
		{
			name:        "endif with arguments",
			tmpl:        `{% if x %}yes{% endif x %}`,
			wantErr:     true,
			errContains: ErrEndifNoArgs.Error(),
		},
		{
			name:    "missing endif",
			tmpl:    `{% if x %}yes`,
			wantErr: true,
		},
		{
			name:        "elif after else",
			tmpl:        `{% if x %}a{% else %}b{% elif y %}c{% endif %}`,
			wantErr:     true,
			errContains: "unknown tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.tmpl)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Compile(%q) err = nil, want error", tt.tmpl)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Compile(%q) err = %q, want containing %q", tt.tmpl, err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("Compile(%q) = %v, want nil", tt.tmpl, err)
			}
			if tt.validate != nil {
				if len(tmpl.root) != 1 {
					t.Fatalf("len(root) = %d, want 1", len(tmpl.root))
				}
				tt.validate(t, tmpl.root[0])
			}
		})
	}
}

func TestParseForTag(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        string
		wantErr     bool
		errContains string
		validate    func(*testing.T, Statement)
	}{
		{
			name: "simple for loop",
			tmpl: `{% for item in items %}{{ item }}{% endfor %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*ForNode)
				if !ok {
					t.Fatalf("got %T, want *ForNode", s)
				}
				if diff := cmp.Diff([]string{"item"}, n.Vars); diff != "" {
					t.Errorf("Vars mismatch (-want +got):\n%s", diff)
				}
				if got := len(n.Body); got != 1 {
					t.Errorf("len(Body) = %d, want 1", got)
				}
				if _, ok := n.Collection.(*VariableNode); !ok {
					t.Errorf("collection type = %T, want *VariableNode", n.Collection)
				}
			},
		},
		{
			name: "for loop with key-value",
			tmpl: `{% for key, value in dict %}{{ key }}: {{ value }}{% endfor %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*ForNode)
				if !ok {
					t.Fatalf("got %T, want *ForNode", s)
				}
				if diff := cmp.Diff([]string{"key", "value"}, n.Vars); diff != "" {
					t.Errorf("Vars mismatch (-want +got):\n%s", diff)
				}
				if got := len(n.Body); got != 3 {
					t.Errorf("len(Body) = %d, want 3", got)
				}
			},
		},
		{
			name: "for loop with index-item",
			tmpl: `{% for i, item in list %}{{ i }}{% endfor %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*ForNode)
				if !ok {
					t.Fatalf("got %T, want *ForNode", s)
				}
				if diff := cmp.Diff([]string{"i", "item"}, n.Vars); diff != "" {
					t.Errorf("Vars mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name: "empty for body",
			tmpl: `{% for item in items %}{% endfor %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*ForNode)
				if !ok {
					t.Fatalf("got %T, want *ForNode", s)
				}
				if got := len(n.Body); got != 0 {
					t.Errorf("len(Body) = %d, want 0", got)
				}
			},
		},
		{
			name: "for with complex collection",
			tmpl: `{% for item in user.items %}{{ item }}{% endfor %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*ForNode)
				if !ok {
					t.Fatalf("got %T, want *ForNode", s)
				}
				if _, ok := n.Collection.(*PropertyAccessNode); !ok {
					t.Errorf("collection type = %T, want *PropertyAccessNode", n.Collection)
				}
			},
		},
		{
			name:        "missing variable name",
			tmpl:        `{% for in items %}{% endfor %}`,
			wantErr:     true,
			errContains: ErrExpectedInKeyword.Error(),
		},
		{
			name:        "missing in keyword",
			tmpl:        `{% for item items %}{% endfor %}`,
			wantErr:     true,
			errContains: ErrExpectedInKeyword.Error(),
		},
		{
			name:        "missing second variable after comma",
			tmpl:        `{% for item, in items %}{% endfor %}`,
			wantErr:     true,
			errContains: ErrExpectedInKeyword.Error(),
		},
		{
			name:        "extra tokens after collection",
			tmpl:        `{% for item in items extra %}{% endfor %}`,
			wantErr:     true,
			errContains: ErrUnexpectedTokensAfterCollection.Error(),
		},
		{
			name:        "endfor with arguments",
			tmpl:        `{% for item in items %}{% endfor x %}`,
			wantErr:     true,
			errContains: ErrEndforNoArgs.Error(),
		},
		{
			name:    "missing endfor",
			tmpl:    `{% for item in items %}{{ item }}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.tmpl)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Compile(%q) err = nil, want error", tt.tmpl)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Compile(%q) err = %q, want containing %q", tt.tmpl, err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("Compile(%q) = %v, want nil", tt.tmpl, err)
			}
			if tt.validate != nil {
				if len(tmpl.root) != 1 {
					t.Fatalf("len(root) = %d, want 1", len(tmpl.root))
				}
				tt.validate(t, tmpl.root[0])
			}
		})
	}
}

func TestParseBreakTag(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        string
		wantErr     bool
		errContains string
		validate    func(*testing.T, Statement)
	}{
		{
			name: "break without arguments",
			tmpl: `{% for item in items %}{% break %}{% endfor %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*ForNode)
				if !ok {
					t.Fatalf("got %T, want *ForNode", s)
				}
				found := false
				for _, node := range n.Body {
					if _, ok := node.(*BreakNode); ok {
						found = true
					}
				}
				if !found {
					t.Error("BreakNode not found in for body")
				}
			},
		},
		{
			name:        "break with arguments should fail",
			tmpl:        `{% for item in items %}{% break now %}{% endfor %}`,
			wantErr:     true,
			errContains: ErrBreakNoArgs.Error(),
		},
		{
			name: "multiple breaks in loop",
			tmpl: `{% for item in items %}{% if item %}{% break %}{% endif %}{% break %}{% endfor %}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.tmpl)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Compile(%q) err = nil, want error", tt.tmpl)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Compile(%q) err = %q, want containing %q", tt.tmpl, err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("Compile(%q) = %v, want nil", tt.tmpl, err)
			}
			if tt.validate != nil {
				if len(tmpl.root) != 1 {
					t.Fatalf("len(root) = %d, want 1", len(tmpl.root))
				}
				tt.validate(t, tmpl.root[0])
			}
		})
	}
}

func TestParseContinueTag(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        string
		wantErr     bool
		errContains string
		validate    func(*testing.T, Statement)
	}{
		{
			name: "continue without arguments",
			tmpl: `{% for item in items %}{% continue %}{% endfor %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*ForNode)
				if !ok {
					t.Fatalf("got %T, want *ForNode", s)
				}
				found := false
				for _, node := range n.Body {
					if _, ok := node.(*ContinueNode); ok {
						found = true
					}
				}
				if !found {
					t.Error("ContinueNode not found in for body")
				}
			},
		},
		{
			name:        "continue with arguments should fail",
			tmpl:        `{% for item in items %}{% continue now %}{% endfor %}`,
			wantErr:     true,
			errContains: ErrContinueNoArgs.Error(),
		},
		{
			name: "multiple continues in loop",
			tmpl: `{% for item in items %}{% if item %}{% continue %}{% endif %}{% continue %}{% endfor %}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.tmpl)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Compile(%q) err = nil, want error", tt.tmpl)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Compile(%q) err = %q, want containing %q", tt.tmpl, err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("Compile(%q) = %v, want nil", tt.tmpl, err)
			}
			if tt.validate != nil {
				if len(tmpl.root) != 1 {
					t.Fatalf("len(root) = %d, want 1", len(tmpl.root))
				}
				tt.validate(t, tmpl.root[0])
			}
		})
	}
}

func TestIfTagExecution(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "if true condition",
			tmpl: `{% if x %}yes{% endif %}`,
			data: map[string]any{"x": true},
			want: "yes",
		},
		{
			name: "if false condition",
			tmpl: `{% if x %}yes{% endif %}`,
			data: map[string]any{"x": false},
			want: "",
		},
		{
			name: "if-else true",
			tmpl: `{% if x %}yes{% else %}no{% endif %}`,
			data: map[string]any{"x": true},
			want: "yes",
		},
		{
			name: "if-else false",
			tmpl: `{% if x %}yes{% else %}no{% endif %}`,
			data: map[string]any{"x": false},
			want: "no",
		},
		{
			name: "if-elif first true",
			tmpl: `{% if first %}one{% elif second %}two{% else %}other{% endif %}`,
			data: map[string]any{"first": true, "second": false},
			want: "one",
		},
		{
			name: "if-elif second true",
			tmpl: `{% if first %}one{% elif second %}two{% else %}other{% endif %}`,
			data: map[string]any{"first": false, "second": true},
			want: "two",
		},
		{
			name: "if-elif else",
			tmpl: `{% if first %}one{% elif second %}two{% else %}other{% endif %}`,
			data: map[string]any{"first": false, "second": false},
			want: "other",
		},
		{
			name: "if with comparison",
			tmpl: `{% if count > 5 %}many{% else %}few{% endif %}`,
			data: map[string]any{"count": 10},
			want: "many",
		},
		{
			name: "nested if",
			tmpl: `{% if outer %}{% if inner %}yes{% endif %}{% endif %}`,
			data: map[string]any{"outer": true, "inner": true},
			want: "yes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestForTagExecution(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		data     map[string]any
		want     string
		contains []string
	}{
		{
			name: "simple for loop",
			tmpl: `{% for item in items %}{{ item }}{% endfor %}`,
			data: map[string]any{"items": []int{1, 2, 3}},
			want: "123",
		},
		{
			name: "for loop with separator",
			tmpl: `{% for item in items %}{{ item }},{% endfor %}`,
			data: map[string]any{"items": []string{"a", "b", "c"}},
			want: "a,b,c,",
		},
		{
			name: "empty collection",
			tmpl: `{% for item in items %}{{ item }}{% endfor %}`,
			data: map[string]any{"items": []int{}},
			want: "",
		},
		{
			name:     "for with key-value",
			tmpl:     `{% for k, v in dict %}{{ k }}:{{ v }};{% endfor %}`,
			data:     map[string]any{"dict": map[string]int{"a": 1, "b": 2}},
			contains: []string{"a:1", "b:2"},
		},
		{
			name:     "for map single var binds key",
			tmpl:     `{% for item in dict %}{{ item }}{% endfor %}`,
			data:     map[string]any{"dict": map[string]int{"a": 1, "b": 2}},
			contains: []string{"a", "b"},
		},
		{
			name: "nested for loops",
			tmpl: `{% for i in outer %}{% for j in inner %}{{ i }}{{ j }}{% endfor %}{% endfor %}`,
			data: map[string]any{"outer": []int{1, 2}, "inner": []string{"a", "b"}},
			want: "1a1b2a2b",
		},
		{
			name: "for loop with index",
			tmpl: `{% for i, item in items %}{{ i }}:{{ item }};{% endfor %}`,
			data: map[string]any{"items": []string{"x", "y"}},
			want: "0:x;1:y;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if tt.want != "" && got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
			for _, c := range tt.contains {
				if !strings.Contains(got, c) {
					t.Errorf("Render(%q) = %q, want containing %q", tt.tmpl, got, c)
				}
			}
		})
	}
}

func TestBreakTagExecution(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "break immediately",
			tmpl: `{% for item in items %}{% break %}{{ item }}{% endfor %}`,
			data: map[string]any{"items": []int{1, 2, 3}},
			want: "",
		},
		{
			name: "break after first item",
			tmpl: `{% for item in items %}{{ item }}{% break %}{% endfor %}`,
			data: map[string]any{"items": []int{1, 2, 3}},
			want: "1",
		},
		{
			name: "conditional break",
			tmpl: `{% for item in items %}{% if item > 2 %}{% break %}{% endif %}{{ item }}{% endfor %}`,
			data: map[string]any{"items": []int{1, 2, 3, 4}},
			want: "12",
		},
		{
			name: "break in nested loop",
			tmpl: `{% for i in outer %}{% for j in inner %}{{ i }}{{ j }}{% if j == 'b' %}{% break %}{% endif %}{% endfor %}{% endfor %}`,
			data: map[string]any{"outer": []int{1, 2}, "inner": []string{"a", "b", "c"}},
			want: "1a1b2a2b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestContinueTagExecution(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "continue immediately",
			tmpl: `{% for item in items %}{% continue %}{{ item }}{% endfor %}`,
			data: map[string]any{"items": []int{1, 2, 3}},
			want: "",
		},
		{
			name: "continue after output",
			tmpl: `{% for item in items %}{{ item }}{% continue %}X{% endfor %}`,
			data: map[string]any{"items": []int{1, 2, 3}},
			want: "123",
		},
		{
			name: "conditional continue",
			tmpl: `{% for item in items %}{% if item > 1 and item < 3 %}{% continue %}{% endif %}{{ item }}{% endfor %}`,
			data: map[string]any{"items": []int{1, 2, 3}},
			want: "13",
		},
		{
			name: "continue in nested loop",
			tmpl: `{% for i in outer %}{% for j in inner %}{% if j %}{% continue %}{% endif %}X{% endfor %}{% endfor %}`,
			data: map[string]any{"outer": []int{1, 2}, "inner": []bool{true, true}},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestComplexTagCombinations(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "if inside for",
			tmpl: `{% for item in items %}{% if item > 2 %}{{ item }}{% endif %}{% endfor %}`,
			data: map[string]any{"items": []int{1, 2, 3, 4}},
			want: "34",
		},
		{
			name: "for inside if",
			tmpl: `{% if show %}{% for item in items %}{{ item }}{% endfor %}{% endif %}`,
			data: map[string]any{"show": true, "items": []int{1, 2, 3}},
			want: "123",
		},
		{
			name: "nested if-for with break",
			tmpl: `{% if show %}{% for item in items %}{% if item > 2 %}{% break %}{% endif %}{{ item }}{% endfor %}{% endif %}`,
			data: map[string]any{"show": true, "items": []int{1, 2, 3, 4}},
			want: "12",
		},
		{
			name: "nested for with continue",
			tmpl: `{% for i in outer %}{% for j in inner %}{% if j > 1 and j < 3 %}{% continue %}{% endif %}{{ i }}{{ j }}{% endfor %}{% endfor %}`,
			data: map[string]any{"outer": []int{1, 2}, "inner": []int{1, 2, 3}},
			want: "11132123",
		},
		{
			name: "multiple elif with for",
			tmpl: `{% if first %}one{% elif second %}{% for i in items %}{{ i }}{% endfor %}{% else %}other{% endif %}`,
			data: map[string]any{"first": false, "second": true, "items": []string{"a", "b"}},
			want: "ab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestTagNodeStructure(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		validate func(*testing.T, Statement)
	}{
		{
			name: "if node structure",
			tmpl: `{% if x %}yes{% endif %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*IfNode)
				if !ok {
					t.Fatalf("got %T, want *IfNode", s)
				}
				if got := len(n.Branches); got != 1 {
					t.Errorf("len(Branches) = %d, want 1", got)
				}
				if got := len(n.ElseBody); got != 0 {
					t.Errorf("len(ElseBody) = %d, want 0", got)
				}
				line, _ := n.Position()
				if line < 1 {
					t.Errorf("Position() line = %d, want >= 1", line)
				}
			},
		},
		{
			name: "for node structure",
			tmpl: `{% for item in items %}x{% endfor %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*ForNode)
				if !ok {
					t.Fatalf("got %T, want *ForNode", s)
				}
				if diff := cmp.Diff([]string{"item"}, n.Vars); diff != "" {
					t.Errorf("Vars mismatch (-want +got):\n%s", diff)
				}
				if n.Collection == nil {
					t.Error("Collection = nil, want non-nil")
				}
				if got := len(n.Body); got < 1 {
					t.Errorf("len(Body) = %d, want >= 1", got)
				}
				line, _ := n.Position()
				if line < 1 {
					t.Errorf("Position() line = %d, want >= 1", line)
				}
			},
		},
		{
			name: "break node structure",
			tmpl: `{% for item in items %}{% break %}{% endfor %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*ForNode)
				if !ok {
					t.Fatalf("got %T, want *ForNode", s)
				}
				var brk *BreakNode
				for _, node := range n.Body {
					if b, ok := node.(*BreakNode); ok {
						brk = b
					}
				}
				if brk == nil {
					t.Fatal("BreakNode not found in for body")
				}
				line, col := brk.Position()
				if line != 1 {
					t.Errorf("Position() line = %d, want 1", line)
				}
				if col <= 0 {
					t.Errorf("Position() col = %d, want > 0", col)
				}
				if got := brk.String(); got != "Break" {
					t.Errorf("String() = %q, want %q", got, "Break")
				}
			},
		},
		{
			name: "continue node structure",
			tmpl: `{% for item in items %}{% continue %}{% endfor %}`,
			validate: func(t *testing.T, s Statement) {
				n, ok := s.(*ForNode)
				if !ok {
					t.Fatalf("got %T, want *ForNode", s)
				}
				var cont *ContinueNode
				for _, node := range n.Body {
					if c, ok := node.(*ContinueNode); ok {
						cont = c
					}
				}
				if cont == nil {
					t.Fatal("ContinueNode not found in for body")
				}
				line, col := cont.Position()
				if line != 1 {
					t.Errorf("Position() line = %d, want 1", line)
				}
				if col <= 0 {
					t.Errorf("Position() col = %d, want > 0", col)
				}
				if got := cont.String(); got != "Continue" {
					t.Errorf("String() = %q, want %q", got, "Continue")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.tmpl)
			if err != nil {
				t.Fatalf("Compile(%q) = %v, want nil", tt.tmpl, err)
			}
			if len(tmpl.root) != 1 {
				t.Fatalf("len(root) = %d, want 1", len(tmpl.root))
			}
			tt.validate(t, tmpl.root[0])
		})
	}
}

func TestIntegration_BasicTemplateRendering(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "plain text",
			tmpl: "Hello, world!",
			data: map[string]any{},
			want: "Hello, world!",
		},
		{
			name: "simple variable",
			tmpl: "Hello, {{ name }}!",
			data: map[string]any{"name": "Alice"},
			want: "Hello, Alice!",
		},
		{
			name: "multiple variables",
			tmpl: "{{ greeting }}, {{ name }}!",
			data: map[string]any{"greeting": "Hello", "name": "Bob"},
			want: "Hello, Bob!",
		},
		{
			name: "variable with number",
			tmpl: "Age: {{ age }}",
			data: map[string]any{"age": 30},
			want: "Age: 30",
		},
		{
			name: "nested property",
			tmpl: "{{ user.name }}",
			data: map[string]any{"user": map[string]any{"name": "Charlie"}},
			want: "Charlie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestIntegration_Filters(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "upper filter",
			tmpl: "{{ name|upper }}",
			data: map[string]any{"name": "alice"},
			want: "ALICE",
		},
		{
			name: "lower filter",
			tmpl: "{{ name|lower }}",
			data: map[string]any{"name": "ALICE"},
			want: "alice",
		},
		{
			name: "capitalize filter",
			tmpl: "{{ name|capitalize }}",
			data: map[string]any{"name": "alice"},
			want: "Alice",
		},
		{
			name: "length filter",
			tmpl: "{{ items|length }}",
			data: map[string]any{"items": []int{1, 2, 3, 4, 5}},
			want: "5",
		},
		{
			name: "default filter with value",
			tmpl: `{{ name|default:"Anonymous" }}`,
			data: map[string]any{"name": "Alice"},
			want: "Alice",
		},
		{
			name: "default filter without value",
			tmpl: `{{ name|default:"Anonymous" }}`,
			data: map[string]any{},
			want: "Anonymous",
		},
		{
			name: "chained filters",
			tmpl: "{{ name|lower|capitalize }}",
			data: map[string]any{"name": "ALICE"},
			want: "Alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestIntegration_ComplexStructures(t *testing.T) {
	type User struct {
		Name    string
		Age     int
		Email   string
		Active  bool
		Profile struct {
			Bio     string
			Website string
		}
	}

	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "struct field access",
			tmpl: "{{ user.Name }}",
			data: map[string]any{
				"user": User{Name: "Alice", Age: 30},
			},
			want: "Alice",
		},
		{
			name: "nested struct field",
			tmpl: "{{ user.Profile.Bio }}",
			data: map[string]any{
				"user": User{
					Name: "Alice",
					Profile: struct {
						Bio     string
						Website string
					}{Bio: "Software Engineer", Website: "https://example.com"},
				},
			},
			want: "Software Engineer",
		},
		{
			name: "struct in loop",
			tmpl: "{% for user in users %}{{ user.Name }},{% endfor %}",
			data: map[string]any{
				"users": []User{
					{Name: "Alice"},
					{Name: "Bob"},
					{Name: "Charlie"},
				},
			},
			want: "Alice,Bob,Charlie,",
		},
		{
			name: "struct with if condition",
			tmpl: "{% if user.Active %}{{ user.Name }} is active{% endif %}",
			data: map[string]any{
				"user": User{Name: "Alice", Active: true},
			},
			want: "Alice is active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestIntegration_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    string
		data    map[string]any
		wantErr bool
		want    string
	}{
		{
			name: "empty template",
			tmpl: "",
			data: map[string]any{},
			want: "",
		},
		{
			name: "whitespace only",
			tmpl: "   \n\t  ",
			data: map[string]any{},
			want: "   \n\t  ",
		},
		{
			name:    "unclosed variable tag",
			tmpl:    "{{ name",
			data:    map[string]any{"name": "Alice"},
			wantErr: true,
		},
		{
			name:    "unclosed block tag",
			tmpl:    "{% if x",
			data:    map[string]any{"x": true},
			wantErr: true,
		},
		{
			name:    "missing endif",
			tmpl:    "{% if x %}yes",
			data:    map[string]any{"x": true},
			wantErr: true,
		},
		{
			name:    "missing endfor",
			tmpl:    "{% for item in items %}{{ item }}",
			data:    map[string]any{"items": []int{1, 2}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Render(%q) err = nil, want error", tt.tmpl)
				}
				return
			}
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestIntegration_ComplexTemplates(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "user profile template",
			tmpl: "\nName: {{ user.name }}\nAge: {{ user.age }}\n{% if user.active %}Status: Active{% else %}Status: Inactive{% endif %}\n",
			data: map[string]any{
				"user": map[string]any{
					"name":   "Alice",
					"age":    30,
					"active": true,
				},
			},
			want: "\nName: Alice\nAge: 30\nStatus: Active\n",
		},
		{
			name: "list rendering with conditions",
			tmpl: "{% for item in items %}{% if item > 5 %}{{ item }} is big\n{% else %}{{ item }} is small\n{% endif %}{% endfor %}",
			data: map[string]any{
				"items": []int{3, 7, 4, 9},
			},
			want: "3 is small\n7 is big\n4 is small\n9 is big\n",
		},
		{
			name: "nested loops with break",
			tmpl: "{% for i in rows %}Row {{ i }}:{% for j in cols %}{% if j > 2 %}{% break %}{% endif %} {{ j }}{% endfor %}\n{% endfor %}",
			data: map[string]any{
				"rows": []int{1, 2},
				"cols": []int{1, 2, 3, 4},
			},
			want: "Row 1: 1 2\nRow 2: 1 2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestIntegration_ArithmeticAndComparison(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		data map[string]any
		want string
	}{
		{
			name: "addition",
			tmpl: "{{ a + b }}",
			data: map[string]any{"a": 5, "b": 3},
			want: "8",
		},
		{
			name: "subtraction",
			tmpl: "{{ a - b }}",
			data: map[string]any{"a": 10, "b": 3},
			want: "7",
		},
		{
			name: "multiplication",
			tmpl: "{{ a * b }}",
			data: map[string]any{"a": 4, "b": 3},
			want: "12",
		},
		{
			name: "division",
			tmpl: "{{ a / b }}",
			data: map[string]any{"a": 10, "b": 2},
			want: "5",
		},
		{
			name: "modulo",
			tmpl: "{{ a % b }}",
			data: map[string]any{"a": 10, "b": 3},
			want: "1",
		},
		{
			name: "comparison greater than",
			tmpl: "{% if a > b %}yes{% else %}no{% endif %}",
			data: map[string]any{"a": 10, "b": 5},
			want: "yes",
		},
		{
			name: "comparison less than",
			tmpl: "{% if a < b %}yes{% else %}no{% endif %}",
			data: map[string]any{"a": 3, "b": 5},
			want: "yes",
		},
		{
			name: "logical and",
			tmpl: "{% if a and b %}yes{% else %}no{% endif %}",
			data: map[string]any{"a": true, "b": true},
			want: "yes",
		},
		{
			name: "logical or",
			tmpl: "{% if a or b %}yes{% else %}no{% endif %}",
			data: map[string]any{"a": false, "b": true},
			want: "yes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("Render(%q) = %v, want nil", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

// testSetNode is a Statement implementation used in TestRegisterTagExternalStatement.
type testSetNode struct {
	varName string
	expr    Expression
	line    int
	col     int
}

func (n *testSetNode) Position() (int, int) { return n.line, n.col }
func (n *testSetNode) String() string       { return fmt.Sprintf("Set(%s)", n.varName) }

func (n *testSetNode) Execute(ctx *ExecutionContext, _ io.Writer) error {
	val, err := n.expr.Evaluate(ctx)
	if err != nil {
		return err
	}
	ctx.Set(n.varName, val.Interface())
	return nil
}

func TestRegisterTagExternalStatement(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	UnregisterTag("set")
	err := RegisterTag("set", func(_ *Parser, start *Token, args *Parser) (Statement, error) {
		v, err := args.ExpectIdentifier()
		if err != nil {
			return nil, args.Error("expected variable name after 'set'")
		}
		if args.Match(TokenSymbol, "=") == nil {
			return nil, args.Error("expected '=' after variable name")
		}
		expr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		if args.Remaining() > 0 {
			return nil, args.Error("unexpected tokens after expression")
		}
		return &testSetNode{
			varName: v.Value,
			expr:    expr,
			line:    start.Line,
			col:     start.Col,
		}, nil
	})
	if err != nil {
		t.Fatalf("RegisterTag(\"set\") = %v, want nil", err)
	}

	got, err := Render(`{% set greeting = "Hello" %}{{ greeting }}, World!`, nil)
	if err != nil {
		t.Fatalf("Render() = %v, want nil", err)
	}
	if want := "Hello, World!"; got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestListTags(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	tags := ListTags()
	for _, want := range []string{"if", "for", "break", "continue"} {
		if !slices.Contains(tags, want) {
			t.Errorf("ListTags() missing %q", want)
		}
	}
	if !slices.IsSorted(tags) {
		t.Errorf("ListTags() = %v, want sorted", tags)
	}
}

func TestHasTag(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{"registered if", "if", true},
		{"registered for", "for", true},
		{"registered break", "break", true},
		{"registered continue", "continue", true},
		{"nonexistent", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasTag(tt.tag); got != tt.want {
				t.Errorf("HasTag(%q) = %v, want %v", tt.tag, got, tt.want)
			}
		})
	}
}

func TestUnregisterTag(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	if !HasTag("if") {
		t.Fatal("HasTag(\"if\") = false, want true (before unregister)")
	}
	UnregisterTag("if")
	if HasTag("if") {
		t.Error("HasTag(\"if\") = true, want false (after unregister)")
	}
	// Unregistering a non-existent tag is a no-op.
	UnregisterTag("nonexistent")
}

func TestDuplicateRegistration(t *testing.T) {
	saved := saveTagRegistry()
	defer restoreTagRegistry(saved)

	tests := []struct {
		name string
		tag  string
		fn   TagParser
	}{
		{"duplicate if", "if", parseIfTag},
		{"duplicate for", "for", parseForTag},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RegisterTag(tt.tag, tt.fn); !errors.Is(err, ErrTagAlreadyRegistered) {
				t.Errorf("RegisterTag(%q) err = %v, want %v", tt.tag, err, ErrTagAlreadyRegistered)
			}
		})
	}
}

func TestTagRegistrySaveRestore(t *testing.T) {
	saved := saveTagRegistry()

	err := RegisterTag("mytesttag", func(_ *Parser, start *Token, _ *Parser) (Statement, error) {
		return NewTextNode("", start.Line, start.Col), nil
	})
	if err != nil {
		t.Fatalf("RegisterTag(\"mytesttag\") = %v, want nil", err)
	}
	if !HasTag("mytesttag") {
		t.Error("HasTag(\"mytesttag\") = false, want true (before restore)")
	}

	restoreTagRegistry(saved)

	if HasTag("mytesttag") {
		t.Error("HasTag(\"mytesttag\") = true, want false (after restore)")
	}
	if !HasTag("if") {
		t.Error("HasTag(\"if\") = false, want true (after restore)")
	}
}

func TestTagIfElseAfterElse(t *testing.T) {
	// {% if x %}...{% else %}...{% else %}...{% endif %}
	_, err := Compile("{% if x %}a{% else %}b{% else %}c{% endif %}")
	if err == nil {
		t.Fatal("expected error for else after else")
	}
}

func TestTagIfElifAfterElse(t *testing.T) {
	// {% if x %}...{% else %}...{% elif y %}...{% endif %}
	_, err := Compile("{% if x %}a{% else %}b{% elif y %}c{% endif %}")
	if err == nil {
		t.Fatal("expected error for elif after else")
	}
}

func TestTagForMissingIn(t *testing.T) {
	// {% for x items %}...{% endfor %} — missing "in"
	_, err := Compile("{% for x items %}a{% endfor %}")
	if err == nil {
		t.Fatal("expected error for missing 'in' keyword")
	}
}

func TestTagForMissingVariable(t *testing.T) {
	// {% for in items %}...{% endfor %} — variable missing/ambiguous
	_, err := Compile("{% for in items %}a{% endfor %}")
	if err == nil {
		t.Fatal("expected error for missing variable")
	}
}

func TestTagForSecondVariableMissing(t *testing.T) {
	// {% for k, in items %}...{% endfor %} — second variable missing
	_, err := Compile("{% for k, in items %}a{% endfor %}")
	if err == nil {
		t.Fatal("expected error for missing second variable")
	}
}
