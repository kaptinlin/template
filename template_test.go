package template

import (
	"strings"
	"testing"
	"time"
)

// TestTemplateExecution verifies the correct execution of a template with various node types.
func TestTemplateExecution(t *testing.T) {
	ctx := mockUserProfileContext()
	// Define test cases
	cases := []struct {
		name     string
		nodes    []*Node
		expected string
	}{
		{
			"TextOnly",
			[]*Node{{Type: "text", Text: "Hello, world!"}},
			"Hello, world!",
		},
		{
			"SingleVariable",
			[]*Node{
				{Type: "text", Text: "User: "},
				{Type: "variable", Variable: "userName"},
			},
			"User: JaneDoe",
		},
		{
			"NestedVariable",
			[]*Node{
				{Type: "variable", Variable: "profile.age"},
			},
			"29",
		},
		{
			"VariableNotFound",
			[]*Node{
				{Type: "variable", Variable: "nonexistent", Text: "{{nonexistent}}"},
			},
			"{{nonexistent}}",
		},
		{
			"MixedContent",
			[]*Node{
				{Type: "text", Text: "Welcome, "},
				{Type: "variable", Variable: "userName"},
				{Type: "text", Text: "! Age: "},
				{Type: "variable", Variable: "profile.age"},
			},
			"Welcome, JaneDoe! Age: 29",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup template
			tmpl := &Template{Nodes: tc.nodes}
			// Execute template
			result := tmpl.MustExecute(ctx)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestConvertToString(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "String",
			input:    "Hello, world!",
			expected: "Hello, world!",
		},
		{
			name:     "SliceOfString",
			input:    []string{"apple", "banana", "cherry"},
			expected: "[apple, banana, cherry]",
		},
		{
			name:     "SliceOfInt",
			input:    []int{1, 2, 3},
			expected: "[1, 2, 3]",
		},
		{
			name:     "SliceOfFloat64",
			input:    []float64{1.1, 2.2, 3.3},
			expected: "[1.1, 2.2, 3.3]",
		},
		{
			name:     "SliceOfBool",
			input:    []bool{true, false, true},
			expected: "[true, false, true]",
		},
		{
			name:     "Time",
			input:    time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "2020-01-01 12:00:00",
		},
		{
			name:     "ComplexTypeWithJSONFallback",
			input:    map[string]interface{}{"name": "John Doe", "age": 30},
			expected: "{\"age\":30,\"name\":\"John Doe\"}",
		},
		{
			name:     "HandleErrorInJSONFallback",
			input:    make(chan int),
			expected: "could not convert value to string: json: unsupported type: chan int",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := convertToString(tc.input)
			if err != nil {
				if err.Error() != tc.expected {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestForLoop(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Array loop",
			template: "{% for item in items %}{{ item }}{% endfor %}",
			context: map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			},
			expected: "abc",
		},
		{
			name:     "Array loop with index",
			template: "{% for item in items %}{{ loop.index }}.{{ item }} {% endfor %}",
			context: map[string]interface{}{
				"items": []interface{}{"apple", "banana", "cherry"},
			},
			expected: "1.apple 2.banana 3.cherry ",
		},
		{
			name:     "Empty collection loop",
			template: "{% for item in items %}{{ item }}{% endfor %}",
			context: map[string]interface{}{
				"items": []interface{}{},
			},
			expected: "",
		},
		{
			name:     "Nested loop",
			template: "{% for row in matrix %}{% for cell in row %}{{ cell }},{% endfor %}\n{% endfor %}",
			context: map[string]interface{}{
				"matrix": []interface{}{
					[]interface{}{1, 2, 3},
					[]interface{}{4, 5, 6},
					[]interface{}{7, 8, 9},
				},
			},
			expected: "1,2,3,\n4,5,6,\n7,8,9,\n",
		},
		{
			name:     "Loop index",
			template: "{% for item in items %}{{ loop.index0 }}:{{ item }}{% endfor %}",
			context: map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			},
			expected: "0:a1:b2:c",
		},
		{
			name:     "Loop reverse index",
			template: "{% for item in items %}{{ loop.revindex }}:{{ item }} {% endfor %}",
			context: map[string]interface{}{
				"items": []interface{}{"a", "b", "c", "d"},
			},
			expected: "4:a 3:b 2:c 1:d ",
		},
		{
			name:     "Complex object loop",
			template: "{% for user in users %}{{ user.name }} ({{ user.age }}), {% endfor %}",
			context: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{"name": "Alice", "age": 20},
					map[string]interface{}{"name": "Bob", "age": 25},
					map[string]interface{}{"name": "Charlie", "age": 30},
				},
			},
			expected: "Alice (20), Bob (25), Charlie (30), ",
		},
		{
			name:     "Three-level nested loop",
			template: "{% for x in cube %}{% for y in x %}{% for z in y %}{{ z }}{% endfor %};{% endfor %}\n{% endfor %}",
			context: map[string]interface{}{
				"cube": []interface{}{
					[]interface{}{
						[]interface{}{1, 2},
						[]interface{}{3, 4},
					},
					[]interface{}{
						[]interface{}{5, 6},
						[]interface{}{7, 8},
					},
				},
			},
			expected: "12;34;\n56;78;\n",
		},
		{
			name:     "Access outer variable in loop",
			template: "{% for item in items %}{{ prefix }}{{ item }}{% endfor %}",
			context: map[string]interface{}{
				"prefix": "Item: ",
				"items":  []interface{}{1, 2, 3},
			},
			expected: "Item: 1Item: 2Item: 3",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := NewTemplate()
			parser := NewParser()
			tpl, err := parser.Parse(tc.template)
			if err != nil {
				t.Fatalf("Template parsing failed: %v", err)
			}
			tmpl.Nodes = tpl.Nodes
			ctx := NewContext()
			for k, v := range tc.context {
				ctx.Set(k, v)
			}

			result, err := tmpl.Execute(ctx)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestIfConditions(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Split string",
			template: "{{ message | split:',' }}",
			context: map[string]interface{}{
				"message": "one,two,three",
			},
			expected: "[one, two, three]",
		},
		{
			name:     "Simple condition",
			template: "{% if age >= 18 %}Adult{% endif %}",
			context: map[string]interface{}{
				"age": 20,
			},
			expected: "Adult",
		},
		{
			name:     "Simple condition with filter",
			template: "{% if age >= 18 %}{{ name | upper }}{% endif %}",
			context: map[string]interface{}{
				"age":  20,
				"name": "alexander",
			},
			expected: "ALEXANDER",
		},
		{
			name:     "Condition with filter",
			template: "{% if name | length > 5 %}Name too long{% endif %}",
			context: map[string]interface{}{
				"name": "Alexander",
			},
			expected: "Name too long",
		},
		{
			name:     "Multiple conditions",
			template: "{% if score >= 90 %}Excellent{% if score == 100 %}Perfect{% endif %}{% endif %}",
			context: map[string]interface{}{
				"score": 100,
			},
			expected: "ExcellentPerfect",
		},
		{
			name:     "Multiple filters in condition",
			template: "{% if message | trim | upper | length > 0 %}Has content{% endif %}",
			context: map[string]interface{}{
				"message": "  Hello  ",
			},
			expected: "Has content",
		},
		{
			name:     "Complex object condition",
			template: "{% if user.age >= 18 && user.name | length > 3 %}{{ user.name | upper }} is adult{% endif %}",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Alice",
					"age":  20,
				},
			},
			expected: "ALICE is adult",
		},
		{
			name:     "String comparison",
			template: "{% if status | lower == 'active' %}Currently active{% endif %}",
			context: map[string]interface{}{
				"status": "ACTIVE",
			},
			expected: "Currently active",
		},
		{
			name:     "Empty value handling",
			template: "{% if content | default:'nil' == 'nil' %}No content{% endif %}",
			context: map[string]interface{}{
				"content": "",
			},
			expected: "No content",
		},
		{
			name:     "Basic boolean",
			template: "{% if isActive %}Active{% endif %}",
			context: map[string]interface{}{
				"isActive": true,
			},
			expected: "Active",
		},
		{
			name:     "Numeric equality",
			template: "{% if count == 0 %}Empty{% else %}Not empty{% endif %}",
			context: map[string]interface{}{
				"count": 0,
			},
			expected: "Empty",
		},
		{
			name:     "String equality",
			template: "{% if status == 'pending' %}Pending{% endif %}",
			context: map[string]interface{}{
				"status": "pending",
			},
			expected: "Pending",
		},
		{
			name:     "AND operation",
			template: "{% if isAdmin && isActive %}Admin online{% endif %}",
			context: map[string]interface{}{
				"isAdmin":  true,
				"isActive": true,
			},
			expected: "Admin online",
		},
		{
			name:     "OR operation",
			template: "{% if isVIP || isMember %}Access granted{% endif %}",
			context: map[string]interface{}{
				"isVIP":    false,
				"isMember": true,
			},
			expected: "Access granted",
		},
		{
			name:     "Simple condition with else",
			template: "{% if age >= 18 %}Adult{% else %}Minor{% endif %}",
			context: map[string]interface{}{
				"age": 16,
			},
			expected: "Minor",
		},
		{
			name:     "String comparison with else",
			template: "{% if status == 'active' %}Currently active{% else %}Inactive{% endif %}",
			context: map[string]interface{}{
				"status": "inactive",
			},
			expected: "Inactive",
		},
		{
			name:     "Numeric comparison with else",
			template: "{% if score >= 50 %}Pass{% else %}Fail{% endif %}",
			context: map[string]interface{}{
				"score": 45,
			},
			expected: "Fail",
		},
		{
			name:     "Complex condition with else",
			template: "{% if user.age >= 18 && user.name | length > 3 %}{{ user.name | upper }} is adult{% else %}Not an adult{% endif %}",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Bob",
					"age":  16,
				},
			},
			expected: "Not an adult",
		},
		{
			name:     "StringArraySizeCheck",
			template: "{% if names | size > 2 %}More than two names{% else %}Two or fewer names{% endif %}",
			context: map[string]interface{}{
				"names": []string{"John", "Alice", "Bob"},
			},
			expected: "More than two names",
		},
		{
			name:     "EmptyIntArrayCheck",
			template: "{% if !scores %}No scores recorded{% endif %}",
			context: map[string]interface{}{
				"scores": []int{},
			},
			expected: "No scores recorded",
		},
		{
			name:     "EmptyStringArrayCheck",
			template: "{% if !names %}No names available{% endif %}",
			context: map[string]interface{}{
				"names": []string{},
			},
			expected: "No names available",
		},
		{
			name:     "NestedMapWithEmployeeCountCheck",
			template: "{% if department.employees | size >= 3 %}Large department{% else %}Small department{% endif %}",
			context: map[string]interface{}{
				"department": map[string]interface{}{
					"name": "R&D",
					"employees": []interface{}{
						map[string]interface{}{"name": "John", "role": "Developer"},
						map[string]interface{}{"name": "Alice", "role": "Tester"},
						map[string]interface{}{"name": "Bob", "role": "Product Manager"},
					},
				},
			},
			expected: "Large department",
		},
		{
			name:     "NonEmptyMapCheck",
			template: "{% if !settings %}Has settings{% endif %}",
			context: map[string]interface{}{
				"settings": map[string]interface{}{
					"theme":   "Dark",
					"High":    "High priority",
					"enabled": "Feature enabled",
				},
			},
			expected: "",
		},
		{
			name:     "NumericTypeComparison",
			template: "{% if int_value > float_value %}Integer greater than float{% else %}Float greater than or equal to integer{% endif %}",
			context: map[string]interface{}{
				"int_value":   5,
				"float_value": 5.5,
			},
			expected: "Float greater than or equal to integer",
		},
		{
			name:     "EmptyMapCheck",
			template: "{% if !config %}Configuration needed{% endif %}",
			context: map[string]interface{}{
				"config": map[string]interface{}{},
			},
			expected: "Configuration needed",
		},
		{
			name:     "BooleanArrayWithTrueValueCheck",
			template: "{% if !flags %}At least one enabled{% else %}All disabled{% endif %}",
			context: map[string]interface{}{
				"flags": []bool{false, false, true, false},
			},
			expected: "All disabled",
		},
		{
			name:     "NestedStringArraySizeCheck",
			template: "{% if struct.names | size > 2 %}More than two names{% else %}Two or fewer names{% endif %}",
			context: map[string]interface{}{
				"struct": map[string]interface{}{
					"names": []string{"John", "Alice", "Bob"},
				},
			},
			expected: "More than two names",
		},
		{
			name:     "NonEmptyStringArrayCondition",
			template: "{% if struct.names %}name{% endif %}",
			context: map[string]interface{}{
				"struct": map[string]interface{}{
					"names": []string{"John", "Alice", "Bob"},
				},
			},
			expected: "name",
		},
		{
			name:     "EmptyStringArrayCondition",
			template: "{% if struct.names %}name{% endif %}",
			context: map[string]interface{}{
				"struct": map[string]interface{}{
					"names": []string{},
				},
			},
			expected: "",
		},
		{
			name:     "EmptyMapCondition",
			template: "{% if struct.names %}name{% endif %}",
			context: map[string]interface{}{
				"struct": map[string]interface{}{
					"names": map[string]interface{}{},
				},
			},
			expected: "",
		},
		{
			name:     "NonEmptyMapCondition",
			template: "{% if struct.names %}name{% endif %}",
			context: map[string]interface{}{
				"struct": map[string]interface{}{
					"names": map[string]interface{}{
						"theme":   "Dark",
						"High":    "High priority",
						"enabled": "Feature enabled",
					},
				},
			},
			expected: "name",
		},
		{
			name:     "ArrayLengthComparison",
			template: "{% if shortList | size < longList | size %}Shorter array{% else %}Equal or longer array{% endif %}",
			context: map[string]interface{}{
				"shortList": []string{"a", "b"},
				"longList":  []string{"x", "y", "z"},
			},
			expected: "Shorter array",
		},
		{
			name:     "MapKeysEqualityCheck",
			template: "{% if config1.debug == config2.debug %}Same debug setting{% else %}Different debug settings{% endif %}",
			context: map[string]interface{}{
				"config1": map[string]interface{}{"debug": true, "mode": "development"},
				"config2": map[string]interface{}{"debug": true, "mode": "production"},
			},
			expected: "Same debug setting",
		},
		{
			name:     "MapNegationCondition",
			template: "{% if !emptyMap && !nonEmptyMap %}Both false{% else %}Both true{% endif %}",
			context: map[string]interface{}{
				"emptyMap":    map[string]interface{}{},
				"nonEmptyMap": map[string]interface{}{"key": "value"},
			},
			expected: "Both true",
		},
		{
			name:     "ComplexNestedCollectionComparison",
			template: "{% if data.users | size > data.groups | size %}More users than groups{% else %}Equal or more groups than users{% endif %}",
			context: map[string]interface{}{
				"data": map[string]interface{}{
					"users": []string{"User1", "User2", "User3"},
					"groups": []interface{}{
						map[string]interface{}{"name": "Admins"},
						map[string]interface{}{"name": "Users"},
					},
				},
			},
			expected: "More users than groups",
		},
		{
			name:     "SliceWithANDOperator",
			template: "{% if numbers && numbers | sum > 10 %}Non-empty array with sum > 10{% endif %}",
			context: map[string]interface{}{
				"numbers": []int{2, 4, 6, 8},
			},
			expected: "Non-empty array with sum > 10",
		},
		{
			name:     "MapWithOROperator",
			template: "{% if emptyConfig || defaultConfig %}Using config{% endif %}",
			context: map[string]interface{}{
				"emptyConfig":   map[string]interface{}{},
				"defaultConfig": map[string]interface{}{"theme": "light"},
			},
			expected: "Using config",
		},
		{
			name:     "NestedArraysInNestedMaps",
			template: "{% if project.teams.developers | size > project.teams.designers | size %}More developers{% else %}Equal or more designers{% endif %}",
			context: map[string]interface{}{
				"project": map[string]interface{}{
					"teams": map[string]interface{}{
						"developers": []string{"Dev1", "Dev2", "Dev3", "Dev4"},
						"designers":  []string{"Des1", "Des2"},
					},
				},
			},
			expected: "More developers",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := NewTemplate()
			parser := NewParser()
			tpl, err := parser.Parse(tc.template)
			if err != nil {
				t.Fatalf("Template parsing failed: %v", err)
			}
			tmpl.Nodes = tpl.Nodes
			ctx := NewContext()
			for k, v := range tc.context {
				ctx.Set(k, v)
			}

			result, err := tmpl.Execute(ctx)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestComplexForLoop(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "Nested data structure",
			template: `Store: {{ store.name }}
Categories:{% for category in store.categories %}
  {{ loop.index }}. {{ category.name }}
  Products:{% for product in category.products %}
    - {{ product.name }}: ${{ product.price }}{% endfor %}{% endfor %}

Total Products: {{ store.products | size }}`,
			context: map[string]interface{}{
				"store": map[string]interface{}{
					"name": "My Store",
					"categories": []interface{}{
						map[string]interface{}{
							"name": "Electronics",
							"products": []interface{}{
								map[string]interface{}{
									"name":  "Phone",
									"price": 599,
								},
								map[string]interface{}{
									"name":  "Laptop",
									"price": 999,
								},
							},
						},
						map[string]interface{}{
							"name": "Books",
							"products": []interface{}{
								map[string]interface{}{
									"name":  "Python Book",
									"price": 39,
								},
							},
						},
					},
					"products": []interface{}{
						map[string]interface{}{
							"name":  "Phone",
							"price": 599,
						},
						map[string]interface{}{
							"name":  "Laptop",
							"price": 999,
						},
						map[string]interface{}{
							"name":  "Python Book",
							"price": 39,
						},
					},
				},
			},
			expected: `Store: My Store
Categories:
  1. Electronics
  Products:
    - Phone: $599
    - Laptop: $999
  2. Books
  Products:
    - Python Book: $39

Total Products: 3`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := NewTemplate()
			parser := NewParser()
			tpl, err := parser.Parse(tc.template)
			if err != nil {
				t.Fatalf("Template parsing failed: %v", err)
			}
			tmpl.Nodes = tpl.Nodes
			ctx := NewContext()
			for k, v := range tc.context {
				ctx.Set(k, v)
			}

			result, err := tmpl.Execute(ctx)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Template output mismatch.\nExpected:\n%s\nGot:\n%s", tc.expected, result)
			}
		})
	}
}

func TestNestedLoopWithConditions(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "Complex nested structure with conditions",
			template: `Store: {{ store.name }}
Categories:{% for category in store.categories %}
  {{ loop.index }}. {{ category.name }}
  Products:{% for product in category.products %}
    {% if product.featured == true %}
    - {{ product.name }}: ${{ product.price }}
    {% if product.discount > 0 %}(Sale: ${{ product.price }}){% endif %}
    Stock: {{ product.stock }} units
    {% endif %}{% endfor %}{% endfor %}

Featured Products:{% for product in store.products %}
  {% if product.featured == true %}- {{ product.name }}
  {% if product.featured == true %}123{% endif %}{% endif %}{% endfor %}`,
			context: map[string]interface{}{
				"store": map[string]interface{}{
					"name": "Tech Marketplace",
					"categories": []interface{}{
						map[string]interface{}{
							"name": "Smartphones",
							"products": []interface{}{
								map[string]interface{}{
									"name":     "Premium Phone",
									"price":    899,
									"stock":    15,
									"discount": 0.9,
									"rating":   4.7,
									"featured": true,
								},
								map[string]interface{}{
									"name":     "Budget Phone",
									"price":    299,
									"stock":    0,
									"rating":   4.2,
									"featured": false,
								},
							},
						},
						map[string]interface{}{
							"name": "Laptops",
							"products": []interface{}{
								map[string]interface{}{
									"name":     "Pro Laptop",
									"price":    1299,
									"stock":    8,
									"discount": 0.85,
									"rating":   4.8,
									"featured": true,
								},
							},
						},
					},
					"products": []interface{}{
						map[string]interface{}{
							"name":     "Premium Phone",
							"featured": true,
							"rating":   4.7,
						},
						map[string]interface{}{
							"name":     "Pro Laptop",
							"featured": true,
							"rating":   4.8,
						},
						map[string]interface{}{
							"name":     "Budget Phone",
							"featured": false,
							"rating":   4.2,
						},
					},
				},
			},
			expected: `Store: Tech Marketplace
Categories:
  1. Smartphones
  Products:
    
    - Premium Phone: $899
    (Sale: $899)
    Stock: 15 units
    
    
  2. Laptops
  Products:
    
    - Pro Laptop: $1299
    (Sale: $1299)
    Stock: 8 units
    

Featured Products:
  - Premium Phone
  123
  - Pro Laptop
  123
  `,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := NewTemplate()
			parser := NewParser()
			tpl, err := parser.Parse(tc.template)
			if err != nil {
				t.Fatalf("Template parsing failed: %v", err)
			}
			tmpl.Nodes = tpl.Nodes
			ctx := NewContext()
			for k, v := range tc.context {
				ctx.Set(k, v)
			}

			result, err := tmpl.Execute(ctx)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Template output mismatch.\nExpected (len=%d):\n%q\nGot (len=%d):\n%q",
					len(tc.expected), tc.expected, len(result), result)
			}
		})
	}
}

func TestComplexTemplateStructures(t *testing.T) {
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "Complex E-commerce Template",
			template: `Store: {{ store.name }}

Featured Categories:
{% for category in store.featured_categories %}
  {% if category.active %}Section: {{ category.name | upper }}
    {% if category.description %}Info: {{ category.description }}{% endif %}
    Available Products:
    {% for item in category.items %}
      {% if item.stock > 0 %}
      - {{ item.name }}
        {% if item.price>= 1000 %}[Premium]{% endif %}
        Price: ${{ item.price }}
        {% if item.discount > 0 %}
        Savings: {{ item.discount }}%
        {% endif %}
        {% if item.rating >= 4.1 %}
        Rating: {{ item.rating }}
        {% endif %}
      {% endif %}
    {% endfor %}
  {% endif %}
{% endfor %}

Special Offers:
{% for offer in store.special_offers %}
  {% if offer.active %}
  {{ offer.name }}
    Discount: {{ offer.discount }}% off
    {% if offer.ends_soon %}Limited Time Offer!{% endif %}
  {% endif %}
{% endfor %}`,
			context: map[string]interface{}{
				"store": map[string]interface{}{
					"name": "Electronics Emporium",
					"featured_categories": []interface{}{
						map[string]interface{}{
							"name":        "Laptops",
							"active":      true,
							"description": "High-performance computers",
							"items": []interface{}{
								map[string]interface{}{
									"name":     "Ultra Book Pro",
									"price":    1299,
									"stock":    5,
									"discount": 15,
									"rating":   4.8,
								},
								map[string]interface{}{
									"name":     "Budget Laptop",
									"price":    599,
									"stock":    10,
									"discount": 0,
									"rating":   4.2,
								},
							},
						},
						map[string]interface{}{
							"name":        "Smartphones",
							"active":      true,
							"description": "Latest mobile devices",
							"items": []interface{}{
								map[string]interface{}{
									"name":     "Pro Phone X",
									"price":    999,
									"stock":    15,
									"discount": 10,
									"rating":   4.9,
								},
							},
						},
					},
					"special_offers": []interface{}{
						map[string]interface{}{
							"name":      "summer sale",
							"active":    true,
							"discount":  20,
							"ends_soon": true,
						},
						map[string]interface{}{
							"name":      "clearance",
							"active":    true,
							"discount":  30,
							"ends_soon": false,
						},
					},
				},
			},
			expected: `Store: Electronics Emporium

Featured Categories:

  Section: LAPTOPS
    Info: High-performance computers
    Available Products:
    
      
      - Ultra Book Pro
        [Premium]
        Price: $1299
        
        Savings: 15%
        
        
        Rating: 4.8
        
      
    
      
      - Budget Laptop
        
        Price: $599
        
        
        Rating: 4.2
        
      
    
  

  Section: SMARTPHONES
    Info: Latest mobile devices
    Available Products:
    
      
      - Pro Phone X
        
        Price: $999
        
        Savings: 10%
        
        
        Rating: 4.9
        
      
    
  


Special Offers:

  
  summer sale
    Discount: 20% off
    Limited Time Offer!
  

  
  clearance
    Discount: 30% off
    
  
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := NewTemplate()
			parser := NewParser()
			tpl, err := parser.Parse(tc.template)
			if err != nil {
				t.Fatalf("Template parsing failed: %v", err)
			}
			tmpl.Nodes = tpl.Nodes
			ctx := NewContext()
			for k, v := range tc.context {
				ctx.Set(k, v)
			}

			result, err := tmpl.Execute(ctx)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Template output mismatch.\nExpected (len=%d):\n%q\nGot (len=%d):\n%q",
					len(tc.expected), tc.expected, len(result), result)
			}
		})
	}
}

// TestDepartmentInterface tests interface implementation in template context
type Department interface {
	GetName() string
}

type DepartmentAddress struct {
	Street   string                         `json:"street"`
	City     string                         `json:"city"`
	Map      map[string]string              `json:"map"`
	Slice    []string                       `json:"slice"`
	MapSlice map[string]map[string][]string `json:"mapSlice"`
	SliceMap [][]map[string]string          `json:"sliceMap"`
}

func (d DepartmentAddress) GetName() string {
	return d.Street + " " + d.City
}

// TestStructWithFilters tests template execution with filters applied to struct fields
func TestStructWithFilters(t *testing.T) {
	type Address struct {
		Street    string    `json:"street"`
		City      string    `json:"city"`
		Time      time.Time `json:"time"`
		ZipCode   string    `json:"zip_code"`
		CountryID int       `json:"country_id"`
	}

	type User struct {
		Name      string    `json:"name"`
		Age       int       `json:"age"`
		Score     float64   `json:"score"`
		IsActive  bool      `json:"is_active"`
		Address   Address   `json:"address"`
		Tags      []string  `json:"tags"`
		Scores    []int     `json:"scores"`
		CreatedAt time.Time `json:"created_at"`
	}

	now := time.Now()
	sixMonthsAgo := now.AddDate(0, -6, 0)

	testCases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "Date filter with structs",
			template: `// Date filter structure testing
Creation time: {{ user.created_at | date "2006-01-02" }}
Address time: {{ user.address.time | date "2006-01-02" }}`,
			context: map[string]interface{}{
				"user": User{
					CreatedAt: sixMonthsAgo,
					Address: Address{
						Time: now,
					},
				},
			},
			expected: `// Date filter structure testing
Creation time: {{ user.created_at | date "2006-01-02" }}
Address time: {{ user.address.time | date "2006-01-02" }}`,
		},
		{
			name: "Default filter with structs",
			template: `// Default value filter structure testing
City: {{ user.address.city | default:"Unknown city" }}
Name: {{ user.name | default:"Unknown user" }}`,
			context: map[string]interface{}{
				"user": User{
					Name: "",
					Address: Address{
						City: "",
					},
				},
			},
			expected: `// Default value filter structure testing
City: Unknown city
Name: Unknown user`,
		},
		{
			name: "Length filter with structs",
			template: `// Length filter structure testing
Name length: {{ user.name | length }}
City length: {{ user.address.city | length }}
Tag count: {{ user.tags | size }}
Score count: {{ user.scores | size }}`,
			context: map[string]interface{}{
				"user": User{
					Name: "Join",
					Address: Address{
						City: "shanghai",
					},
					Tags:   []string{"tag one", "tag two", "tag three"},
					Scores: []int{80, 90, 85},
				},
			},
			expected: `// Length filter structure testing
Name length: 4
City length: 8
Tag count: 3
Score count: 3`,
		},
		{
			name: "Basic conditions with filters on structs",
			template: `// Structure condition filter testing
{% if user.name | length > 0 %}User name is not empty{% endif %}
{% if user.address.city | length > 2 %}City name is longer than 2 characters{% endif %}
{% if user.tags | length >= 3 %}At least 3 tags{% endif %}`,
			context: map[string]interface{}{
				"user": User{
					Name: "Join",
					Address: Address{
						City: "shanghai",
					},
					Tags: []string{"tag one", "tag two", "tag three"},
				},
			},
			expected: `// Structure condition filter testing
User name is not empty
City name is longer than 2 characters
At least 3 tags`,
		},
		{
			name: "Split filter with structs",
			template: `// Split filter structure testing
Address split: {{ address_str | split:"," }}`,
			context: map[string]interface{}{
				"address_str": "Shanghai,Pudong,Zhangjiang",
			},
			expected: `// Split filter structure testing
Address split: [Shanghai, Pudong, Zhangjiang]`,
		},
		{
			name: "Trim filter with structs",
			template: `// Trim filter structure testing
Trim name: {{ user.name | trim }}
Trim city: {{ user.address.city | trim }}`,
			context: map[string]interface{}{
				"user": User{
					Name: "  Join  ",
					Address: Address{
						City: " shanghai ",
					},
				},
			},
			expected: `// Trim filter structure testing
Trim name: Join
Trim city: shanghai`,
		},
		{
			name: "For loops with structs and conditions",
			template: `// For loops with structs and conditions
{% for item in items %}
	{% if item.user.age > 18 %}
		Name: {{ item.user.name }}
		Age: {{ item.user.age }}
		City: {{ item.user.address.city }}
	{% endif %}
{% endfor %}`,
			context: map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"user": User{
							Name: "John",
							Age:  20,
							Address: Address{
								City: "Shanghai",
							},
						},
					},
					{
						"user": User{
							Name: "Peter",
							Age:  17,
							Address: Address{
								City: "Beijing",
							},
						},
					},
					{
						"user": User{
							Name: "Mark",
							Age:  25,
							Address: Address{
								City: "Guangzhou",
							},
						},
					},
				},
			},
			expected: `// For loops with structs and conditions

	
		Name: John
		Age: 20
		City: Shanghai
	

	

	
		Name: Mark
		Age: 25
		City: Guangzhou
	
`,
		},
		{
			name: "Upper and lower filters with structs",
			template: `// Upper and lower filters with structs
Upper name: {{ user.name | upper }}
Lower city: {{ user.address.city | lower }}`,
			context: map[string]interface{}{
				"user": User{
					Name: "Zhang San",
					Address: Address{
						City: "SHANGHAI",
					},
				},
			},
			expected: `// Upper and lower filters with structs
Upper name: ZHANG SAN
Lower city: shanghai`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := NewTemplate()
			parser := NewParser()
			tpl, err := parser.Parse(tc.template)
			if err != nil {
				t.Fatalf("Template parsing failed: %v", err)
			}
			tmpl.Nodes = tpl.Nodes

			ctx := NewContext()
			for k, v := range tc.context {
				ctx.Set(k, v)
			}

			result, err := tmpl.Execute(ctx)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Template output mismatch.\nExpected (len=%d):\n%q\nGot (len=%d):\n%q",
					len(tc.expected), tc.expected, len(result), result)
			}
		})
	}
}

// TestComplexNestedVariableAccess tests accessing nested variables in various complex data structures
func TestComplexNestedVariableAccess(t *testing.T) {
	sourceTemplate := `
	// Basic variable access tests
	Hello, 
	! {{ simpleMap.name }} ! {{ arrayMap.items }} ! {{ nestedMap.user.name }} ! {{ stringSliceMap.items }}
	! {{ nestedSliceMap.config.values }} ! {{ basicSlice.0 }} {{ objectSlice.0.name }} {{ objectSlice.1.nickname }} 
	
	// Nested arrays and map access tests
	! {{ nestedObjectArray.0.0.name }} {{ nestedObjectArray.0.1.nickname }} {{ nestedObjectArray.1.0.name }} {{ nestedObjectArray.1.1.nickname }}
	
	// Deep nested structure access tests
	{{ deepStructure.0.0.user.path.0.innerPath.0 }} {{ deepStructure.0.1.profile.path.0.innerPath.0 }} 
	{{ deepStructure.1.0.user.path.0.innerPath.0 }} {{ deepStructure.1.1.profile.path.0.innerPath.0 }}`

	tmpl, err := Parse(sourceTemplate)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Create context and set complex nested data
	ctx := NewContext()
	ctx.Set("simpleMap", map[string]string{
		"name": "SimpleMap",
	})
	ctx.Set("arrayMap", map[string]interface{}{
		"items": []string{"ArrayMap", "ArrayMap-2"},
	})
	ctx.Set("nestedMap", map[string]interface{}{
		"user": map[string]interface{}{
			"name": "NestedMap",
		},
	})
	ctx.Set("stringSliceMap", map[string][]string{
		"items": {"StringSliceMap", "StringSliceMap-2"},
	})
	ctx.Set("nestedSliceMap", map[string]map[string][]string{
		"config": {
			"values": []string{"NestedSliceMap", "NestedSliceMap-2"},
		},
	})
	ctx.Set("basicSlice", []string{"BasicSlice", "BasicSlice-2"})
	ctx.Set("objectSlice", []map[string]interface{}{
		{
			"name": "ObjectSlice1",
		},
		{
			"nickname": "ObjectSlice2",
		},
	})
	ctx.Set("nestedObjectArray", [][]map[string]interface{}{
		{
			{
				"name": "NestedArray1",
			},
			{
				"nickname": "NestedArray2",
			},
		},
		{
			{
				"name": "NestedArray3",
			},
			{
				"nickname": "NestedArray4",
			},
		},
	})
	ctx.Set("deepStructure", [][]map[string]map[string][]map[string][]string{
		{
			{
				"user": map[string][]map[string][]string{
					"path": {
						{
							"innerPath": []string{"DeepStructure1"},
						},
					},
				},
			},
			{
				"profile": map[string][]map[string][]string{
					"path": {
						{
							"innerPath": []string{"DeepStructure2"},
						},
					},
				},
			},
		},
		{
			{
				"user": map[string][]map[string][]string{
					"path": {
						{
							"innerPath": []string{"DeepStructure3"},
						},
					},
				},
			},
			{
				"profile": map[string][]map[string][]string{
					"path": {
						{
							"innerPath": []string{"DeepStructure4"},
						},
					},
				},
			},
		},
	})

	// Execute template
	output, err := tmpl.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Verify output contains all expected values
	expectedValues := []string{
		"SimpleMap", "ArrayMap", "NestedMap", "StringSliceMap", "NestedSliceMap",
		"BasicSlice", "ObjectSlice1", "ObjectSlice2",
		"NestedArray1", "NestedArray2", "NestedArray3", "NestedArray4",
		"DeepStructure1", "DeepStructure2", "DeepStructure3", "DeepStructure4",
	}

	for _, expected := range expectedValues {
		if !strings.Contains(output, expected) {
			t.Errorf("Output doesn't contain expected value: %s", expected)
		}
	}
}

// TestComplexNestedConditions tests conditions with deeply nested data structures
func TestComplexNestedConditions(t *testing.T) {
	sourceTemplate := `
	// Basic condition tests
	{% if userArray.0.0.credentials %}Hello, {{ userArray.1.0.profile }}!{% endif %}
	{% if userProfile.credentials %}Hello, {{ userProfile.credentials }}!{% endif %}
	
	// Array index condition tests
	{% if studentList.0.info %}Hello, {{ studentList.1.info }}!{% endif %}
	{% if teacherList.0.details %}Hello, {{ teacherList.1.details }}!{% endif %}
	
	// Complex nested condition tests
	{% if scoreData.0.subject.math.grades.0 %}Math score: {{ scoreData.0.subject.math.grades.1 }}!{% endif %}
	{% if scoreData.1.subject.math.grades.1 %}Language score: {{ scoreData.1.subject.math.grades.1 }}!{% endif %}
	
	// Deep nested condition tests
	{% if configData.0.system.settings.theme %}Theme: {{ configData.0.system.settings.theme }}!{% endif %}
	{% if configData.1.system.settings.theme %}Layout: {{ configData.1.system.settings.theme }}!{% endif %}
	
	// Extremely deep nested structure condition tests
	{% if appSettings.0.preferences.display.options.flags.0 %}Option1: {{ appSettings.0.preferences.display.options.flags.1 }}!{% endif %}
	{% if appSettings.1.preferences.display.options.flags.0 %}Option2: {{ appSettings.1.preferences.display.options.flags.1 }}!{% endif %}`

	tmpl, err := Parse(sourceTemplate)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Create context and set complex nested data
	ctx := NewContext()
	ctx.Set("userArray", [][]map[string][]string{
		{
			{
				"credentials": []string{"UserCredentials", "Password"},
			},
		},
		{
			{
				"profile": []string{"UserProfile", "PersonalInfo"},
			},
		},
	})
	ctx.Set("userProfile", map[string][]string{
		"credentials": {"AccountInfo", "SecurityInfo"},
	})
	ctx.Set("studentList", []map[string]interface{}{
		{
			"info": "StudentInfo1",
		},
		{
			"info": "StudentInfo2",
		},
	})
	ctx.Set("teacherList", []map[string]map[string]interface{}{
		{
			"details": {
				"info": "TeacherInfo1",
			},
		},
		{
			"details": {
				"info": "TeacherInfo2",
			},
		},
	})
	ctx.Set("scoreData", []map[string]map[string]map[string][]int{
		{
			"subject": {
				"math": {
					"grades": []int{80, 85, 90},
				},
			},
		},
		{
			"subject": {
				"math": {
					"grades": []int{75, 88, 92},
				},
			},
		},
	})
	ctx.Set("configData", []map[string]map[string]map[string]map[string]int{
		{
			"system": {
				"settings": {
					"theme": {
						"value": 1,
					},
				},
			},
		},
		{
			"system": {
				"settings": {
					"theme": {
						"value": 2,
					},
				},
			},
		},
	})
	ctx.Set("appSettings", []map[string]map[string]map[string]map[string][]bool{
		{
			"preferences": {
				"display": {
					"options": {
						"flags": []bool{true, false},
					},
				},
			},
		},
		{
			"preferences": {
				"display": {
					"options": {
						"flags": []bool{true, false},
					},
				},
			},
		},
	})

	// Execute template
	output, err := tmpl.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Verify output contains all expected values
	expectedValues := []string{
		"Hello, [UserProfile, PersonalInfo]!", "Hello, [AccountInfo, SecurityInfo]!", "Hello, StudentInfo2!",
		"Hello, {\"info\":\"TeacherInfo2\"}!", "Math score: 85!", "Language score: 88!",
		"Theme: {\"value\":1}!", "Layout: {\"value\":2}!", "Option1: false!", "Option2: false!",
	}

	for i, expected := range expectedValues {
		if !strings.Contains(output, expected) {
			t.Errorf("Output doesn't contain expected value[%d]: %s", i, expected)
		}
	}
}

// TestComplexLoops tests for loops with various complex data structures
func TestComplexLoops(t *testing.T) {
	sourceTemplate := `
	// Integer array loop tests
	{% for num in numberList %}{{ num }} {% endfor %}
	
	// String loop tests
	{% for char in textContent %}{{ char }} {% endfor %}
	
	// Slice element access by index tests
	Basic element access: {{ stringArray.0 }}, {{ stringArray.1 }}, {{ stringArray.2 }}
	2D array access: {{ matrixData.0.0 }}, {{ matrixData.1.2 }}
	Object array access: {{ productList.0.title }}, {{ productList.1.price }}
	
	// 2D string array tests
	{% for row in gridData %}
		{% for cell in row %}{{ cell }} {% endfor %}
	{% endfor %}
	
	// Object array tests
	{% for item in productList %}
		{{ item.title }} - {{ item.price }}
	{% endfor %}
	
	// Nested object tests
	{% for item in userSettings %}
		{{ item.key }}:
		{% for subItem in item.value %}
			{{ subItem.key }}: {{ subItem.value }}
		{% endfor %}
	{% endfor %}
	
	// Complex nested structure tests
	{% for item in academicRecords %}
		{{ item.key }}:
		{% for subItem in item.value %}
			{{ subItem.key }}: 
			{% for grade in subItem.value %}
				{{ grade }}
			{% endfor %}
		{% endfor %}
	{% endfor %}
	
	// Multi-level nested tests
	{% for item in schoolData %}
		{{ item.key }}:
		{% for category in item.value %}
			{{ category.key }}:
			{% for person in category.value %}
				{% for detail in person %}
					{{ detail.key }}: {{ detail.value }}
				{% endfor %}
			{% endfor %}
		{% endfor %}
	{% endfor %}`

	tmpl, err := Parse(sourceTemplate)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Create context and set complex nested data
	ctx := NewContext()

	// Basic data types
	ctx.Set("numberList", []int{1, 2, 3, 4, 5})
	ctx.Set("textContent", "HelloWorld")

	// Index access
	ctx.Set("stringArray", []string{"FirstItem", "SecondItem", "ThirdItem"})
	ctx.Set("matrixData", [][]string{
		{"A1", "A2", "A3"},
		{"B1", "B2", "B3"},
	})
	ctx.Set("productList", []map[string]interface{}{
		{"title": "Product1", "price": 100},
		{"title": "Product2", "price": 200},
	})

	// 2D array
	ctx.Set("gridData", [][]string{
		{"A1", "A2", "A3"},
		{"B1", "B2", "B3"},
		{"C1", "C2", "C3"},
	})

	// Object array
	ctx.Set("productList", []map[string]interface{}{
		{"title": "Product1", "price": 100},
		{"title": "Product2", "price": 200},
		{"title": "Product3", "price": 300},
	})

	// Nested objects
	userSettings := map[string]map[string]interface{}{
		"user1": {
			"name": "John",
			"age":  30,
		},
		"user2": {
			"name": "Jane",
			"age":  25,
		},
	}
	ctx.Set("userSettings", userSettings)

	// Complex nested structure
	academicRecords := map[string]map[string][]int{
		"semester1": {
			"math":    []int{90, 85, 92},
			"science": []int{88, 79, 95},
		},
		"semester2": {
			"math":    []int{78, 82, 89},
			"science": []int{92, 95, 88},
		},
	}
	ctx.Set("academicRecords", academicRecords)

	// Multi-level nested structure
	schoolData := map[string]map[string][]map[string]string{
		"class1": {
			"students": []map[string]string{
				{"name": "Student1", "grade": "A"},
				{"name": "Student2", "grade": "B"},
			},
			"teachers": []map[string]string{
				{"name": "Teacher1", "subject": "Math"},
				{"name": "Teacher2", "subject": "English"},
			},
		},
		"class2": {
			"students": []map[string]string{
				{"name": "Student3", "grade": "A+"},
				{"name": "Student4", "grade": "C"},
			},
			"teachers": []map[string]string{
				{"name": "Teacher3", "subject": "Science"},
				{"name": "Teacher4", "subject": "Physics"},
			},
		},
	}
	ctx.Set("schoolData", schoolData)

	// Execute template
	output, err := tmpl.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Check results of various loop and access patterns
	expectedValues := []string{
		// Basic array loop
		"1 2 3 4 5",
		// String loop
		"H e l l o W o r l d",
		// Index access
		"FirstItem, SecondItem, ThirdItem", "A1, B3", "Product1, 200",
		// Nested array loop
		"A1 A2 A3", "B1 B2 B3", "C1 C2 C3",
		// Object array loop
		"Product1 - 100", "Product2 - 200", "Product3 - 300",
		// Nested object data
		"John", "Jane", "30", "25",
		// Complex nested structure
		"90", "85", "92", "78", "82", "89",
		// Multi-level nested structure
		"Student1", "Student2", "Teacher1", "Teacher2", "Student3", "Student4", "Teacher3", "Teacher4",
		"A", "B", "A+", "C", "Math", "English", "Science", "Physics",
	}

	for _, expected := range expectedValues {
		if !strings.Contains(output, expected) {
			t.Errorf("Output doesn't contain expected value: %s", expected)
		}
	}
}

// TestComplexStructFields tests template execution with complex nested struct fields
func TestComplexStructFields(t *testing.T) {
	type Address struct {
		Street     string                         `json:"street"`
		City       string                         `json:"city"`
		Time       time.Time                      `json:"time"`
		Map        map[string]string              `json:"map"`
		Slice      []string                       `json:"slice"`
		MapSlice   map[string]map[string][]string `json:"mapSlice"`
		SliceMap   [][]map[string]string          `json:"sliceMap"`
		Department Department                     `json:"department"`
	}

	type User struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	type Team struct {
		Name    string `json:"name"`
		Members []User `json:"members"`
	}

	now := time.Now()

	// Create base map
	testMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	// Create base slice
	testSlice := []string{"item1", "item2", "item3"}

	// Create nested MapSlice
	testMapSlice := map[string]map[string][]string{
		"level1": {
			"sublevel": {"deepitem1", "deepitem2", "deepitem3"},
		},
	}

	// Create nested SliceMap
	testSliceMap := [][]map[string]string{
		{
			{"key1": "nestedvalue1", "key2": "nestedvalue2"},
			{"key1": "nestedvalue3", "key2": "nestedvalue4"},
		},
		{
			{"key1": "nestedvalue5", "key2": "nestedvalue6"},
		},
	}

	testCases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "Basic variable access",
			template: `Hello, {{ name }}!
Simple struct:
User: {{ user }}
Username: {{ user.name }}
User age: {{ user.age }}`,
			context: map[string]interface{}{
				"name": "World",
				"user": User{
					Name: "John",
					Age:  30,
				},
			},
			expected: `Hello, World!
Simple struct:
User: {"address":{"city":"","department":null,"map":null,"mapSlice":null,"slice":null,"sliceMap":null,"street":"","time":"0001-01-01T00:00:00Z"},"age":30,"name":"John"}
Username: John
User age: 30`,
		},
		{
			name: "Nested struct access",
			template: `Nested struct:
User address: {{ user.address }}
Street: {{ user.address.street }}
City: {{ user.address.city }}`,
			context: map[string]interface{}{
				"user": User{
					Name: "John",
					Age:  30,
					Address: Address{
						Street: "123 Main St",
						City:   "New York",
					},
				},
			},
			expected: `Nested struct:
User address: {"city":"New York","department":null,"map":null,"mapSlice":null,"slice":null,"sliceMap":null,"street":"123 Main St","time":"0001-01-01T00:00:00Z"}
Street: 123 Main St
City: New York`,
		},
		{
			name: "Time formatting",
			template: `Time test:
Time: {{ user.address.time }}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						Time: now,
					},
				},
			},
			expected: `Time test:
Time: ` + now.Format(time.RFC3339Nano),
		},
		{
			name: "Map access",
			template: `Map test:
Map: {{ user.address.map }}
Map key1: {{ user.address.map.key1 }}
Map key2: {{ user.address.map.key2 }}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						Map: testMap,
					},
				},
			},
			expected: `Map test:
Map: {"key1":"value1","key2":"value2"}
Map key1: value1
Map key2: value2`,
		},
		{
			name: "Slice access",
			template: `Slice test:
Slice: {{ user.address.slice }}
First element: {{ user.address.slice.0 }}
Second element: {{ user.address.slice.1 }}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						Slice: testSlice,
					},
				},
			},
			expected: `Slice test:
Slice: ["item1","item2","item3"]
First element: item1
Second element: item2`,
		},
		{
			name: "MapSlice access",
			template: `MapSlice test:
MapSlice: {{ user.address.mapSlice }}
Level 1 key: {{ user.address.mapSlice.level1 }}
Level 2 key: {{ user.address.mapSlice.level1.sublevel }}
Array in level 2: {{ user.address.mapSlice.level1.sublevel.0 }}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						MapSlice: testMapSlice,
					},
				},
			},
			expected: `MapSlice test:
MapSlice: {"level1":{"sublevel":["deepitem1","deepitem2","deepitem3"]}}
Level 1 key: {"sublevel":["deepitem1","deepitem2","deepitem3"]}
Level 2 key: ["deepitem1","deepitem2","deepitem3"]
Array in level 2: deepitem1`,
		},
		{
			name: "SliceMap access",
			template: `SliceMap test:
SliceMap: {{ user.address.sliceMap }}
First array: {{ user.address.sliceMap.0 }}
First element of first array: {{ user.address.sliceMap.0.0 }}
key1 of first element of first array: {{ user.address.sliceMap.0.0.key1 }}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						SliceMap: testSliceMap,
					},
				},
			},
			expected: `SliceMap test:
SliceMap: [[{"key1":"nestedvalue1","key2":"nestedvalue2"},{"key1":"nestedvalue3","key2":"nestedvalue4"}],[{"key1":"nestedvalue5","key2":"nestedvalue6"}]]
First array: [{"key1":"nestedvalue1","key2":"nestedvalue2"},{"key1":"nestedvalue3","key2":"nestedvalue4"}]
First element of first array: {"key1":"nestedvalue1","key2":"nestedvalue2"}
key1 of first element of first array: nestedvalue1`,
		},
		{
			name: "Team members access",
			template: `Team test:
Team name: {{ team.name }}
First member: {{ team.members.0.name }}
Second member city: {{ team.members.1.address.city }}`,
			context: map[string]interface{}{
				"team": Team{
					Name: "Development Team",
					Members: []User{
						{
							Name: "John",
							Address: Address{
								City: "New York",
							},
						},
						{
							Name: "Jane",
							Address: Address{
								City: "Boston",
							},
						},
					},
				},
			},
			expected: `Team test:
Team name: Development Team
First member: John
Second member city: Boston`,
		},
		{
			name: "Interface field access",
			template: `Department test:
Department street: {{ department.department.street }}	
Department city: {{ department.department.city }}
Department Map: {{ department.department.map }}
Department Slice: {{ department.department.slice }}
Department MapSlice: {{ department.department.mapSlice }}
Department SliceMap: {{ department.department.sliceMap }}`,
			context: map[string]interface{}{
				"department": Address{
					Department: DepartmentAddress{
						Street:   "123 Main St",
						City:     "New York",
						Map:      testMap,
						Slice:    testSlice,
						MapSlice: testMapSlice,
						SliceMap: testSliceMap,
					},
				},
			},
			expected: `Department test:
Department street: 123 Main St	
Department city: New York
Department Map: {"key1":"value1","key2":"value2"}
Department Slice: ["item1","item2","item3"]
Department MapSlice: {"level1":{"sublevel":["deepitem1","deepitem2","deepitem3"]}}
Department SliceMap: [[{"key1":"nestedvalue1","key2":"nestedvalue2"},{"key1":"nestedvalue3","key2":"nestedvalue4"}],[{"key1":"nestedvalue5","key2":"nestedvalue6"}]]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := NewTemplate()
			parser := NewParser()
			tpl, err := parser.Parse(tc.template)
			if err != nil {
				t.Fatalf("Template parsing failed: %v", err)
			}
			tmpl.Nodes = tpl.Nodes

			ctx := NewContext()
			for k, v := range tc.context {
				ctx.Set(k, v)
			}

			result, err := tmpl.Execute(ctx)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Template output mismatch.\nExpected (len=%d):\n%q\nGot (len=%d):\n%q",
					len(tc.expected), tc.expected, len(result), result)
			}
		})
	}
}

// TestIfConditionsWithComplexStructs tests if conditions with complex struct fields
func TestIfConditionsWithComplexStructs(t *testing.T) {
	type Address struct {
		Street     string                         `json:"street"`
		City       string                         `json:"city"`
		Time       time.Time                      `json:"time"`
		Map        map[string]string              `json:"map"`
		Slice      []string                       `json:"slice"`
		MapSlice   map[string]map[string][]string `json:"mapSlice"`
		SliceMap   [][]map[string]string          `json:"sliceMap"`
		Department Department                     `json:"department"`
	}

	type User struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	type Team struct {
		Name    string `json:"name"`
		Members []User `json:"members"`
	}

	now := time.Now()
	testSlice := []string{"item1", "item2", "item3"}

	testCases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "Basic condition checks",
			template: `// Basic condition tests
{% if user.name %}Username is: {{ user.name }}{% endif %}
{% if user.age > 20 %}User is older than 20{% endif %}`,
			context: map[string]interface{}{
				"user": User{
					Name: "John",
					Age:  30,
				},
			},
			expected: `// Basic condition tests
Username is: John
User is older than 20`,
		},
		{
			name: "Array and slice conditions",
			template: `// Array and slice condition tests
{% if user.address.slice.0 %}First slice element: {{ user.address.slice.0 }}{% endif %}
{% if user.address.sliceMap.0.0.key1 %}Value in nested structure: {{ user.address.sliceMap.0.0.key1 }}{% endif %}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						Slice: testSlice,
						SliceMap: [][]map[string]string{
							{
								{"key1": "nestedvalue1", "key2": "nestedvalue2"},
							},
						},
					},
				},
			},
			expected: `// Array and slice condition tests
First slice element: item1
Value in nested structure: nestedvalue1`,
		},
		{
			name: "Complex structure conditions",
			template: `// Complex structure condition tests
{% if team.members.0.name %}First team member: {{ team.members.0.name }}{% endif %}
{% if team.members.1.address.city %}Second member's city: {{ team.members.1.address.city }}{% endif %}`,
			context: map[string]interface{}{
				"team": Team{
					Name: "Development Team",
					Members: []User{
						{
							Name: "John",
							Address: Address{
								City: "New York",
							},
						},
						{
							Name: "Jane",
							Address: Address{
								City: "Boston",
							},
						},
					},
				},
			},
			expected: `// Complex structure condition tests
First team member: John
Second member's city: Boston`,
		},
		{
			name: "Interface type conditions",
			template: `// Interface type condition tests
{% if department.department.street %}Department street: {{ department.department.street }}{% endif %}
{% if department.department.city %}Department city: {{ department.department.city }}{% endif %}`,
			context: map[string]interface{}{
				"department": Address{
					Department: DepartmentAddress{
						Street: "123 Main St",
						City:   "New York",
					},
				},
			},
			expected: `// Interface type condition tests
Department street: 123 Main St
Department city: New York`,
		},
		{
			name: "Time type conditions",
			template: `// Time type condition tests
{% if user.address.time %}Time exists: {{ user.address.time }}{% endif %}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						Time: now,
					},
				},
			},
			expected: func() string {
				// Use ISO format directly
				return `// Time type condition tests
Time exists: ` + now.Format(time.RFC3339Nano)
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := NewTemplate()
			parser := NewParser()
			tpl, err := parser.Parse(tc.template)
			if err != nil {
				t.Fatalf("Template parsing failed: %v", err)
			}
			tmpl.Nodes = tpl.Nodes

			ctx := NewContext()
			for k, v := range tc.context {
				ctx.Set(k, v)
			}

			result, err := tmpl.Execute(ctx)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Template output mismatch.\nExpected (len=%d):\n%q\nGot (len=%d):\n%q",
					len(tc.expected), tc.expected, len(result), result)
			}
		})
	}
}

// TestForLoopsWithComplexStructs tests for loops with complex struct fields
func TestForLoopsWithComplexStructs(t *testing.T) {
	type Address struct {
		Street     string                         `json:"street"`
		City       string                         `json:"city"`
		Time       time.Time                      `json:"time"`
		Map        map[string]string              `json:"map"`
		Slice      []string                       `json:"slice"`
		MapSlice   map[string]map[string][]string `json:"mapSlice"`
		SliceMap   [][]map[string]string          `json:"sliceMap"`
		Department Department                     `json:"department"`
	}

	type User struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	type Team struct {
		Name    string `json:"name"`
		Members []User `json:"members"`
	}

	testSlice := []string{"item1", "item2", "item3"}
	testMapSlice := map[string]map[string][]string{
		"level1": {
			"sublevel": {"deepitem1", "deepitem2", "deepitem3"},
		},
	}
	testSliceMap := [][]map[string]string{
		{
			{"key1": "nestedvalue1", "key2": "nestedvalue2"},
			{"key1": "nestedvalue3", "key2": "nestedvalue4"},
		},
		{
			{"key1": "nestedvalue5", "key2": "nestedvalue6"},
		},
	}

	testCases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "Basic loop test",
			template: `// Basic loop tests
Slice loop:
{% for item in user.address.slice %}
		- {{ item }}
{% endfor %}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						Slice: testSlice,
					},
				},
			},
			expected: `// Basic loop tests
Slice loop:

		- item1

		- item2

		- item3
`,
		},
		{
			name: "Nested loop test with team members",
			template: `// Nested loop tests
Team members loop:
{% for member in team.members %}
		Member name: {{ member.name }}, age: {{ member.age }}
		Address info: {{ member.address.city }} {{ member.address.street }}
{% endfor %}`,
			context: map[string]interface{}{
				"team": Team{
					Members: []User{
						{
							Name: "John",
							Age:  30,
							Address: Address{
								Street: "123 Main St",
								City:   "New York",
							},
						},
						{
							Name: "Jane",
							Age:  25,
							Address: Address{
								Street: "456 Elm St",
								City:   "Boston",
							},
						},
					},
				},
			},
			expected: `// Nested loop tests
Team members loop:

		Member name: John, age: 30
		Address info: New York 123 Main St

		Member name: Jane, age: 25
		Address info: Boston 456 Elm St
`,
		},
		{
			name: "Complex SliceMap iteration",
			template: `// Complex data structure loops
SliceMap loop:
{% for arr in user.address.sliceMap %}
		Array index:
		{% for item in arr %}
			Key-value pair: {{ item.key1 }} - {{ item.key2 }}
		{% endfor %}
{% endfor %}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						SliceMap: testSliceMap,
					},
				},
			},
			expected: `// Complex data structure loops
SliceMap loop:

		Array index:
		
			Key-value pair: nestedvalue1 - nestedvalue2
		
			Key-value pair: nestedvalue3 - nestedvalue4
		

		Array index:
		
			Key-value pair: nestedvalue5 - nestedvalue6
		
`,
		},
		{
			name: "MapSlice iteration",
			template: `// MapSlice loops (Note: Map cannot be directly iterated with key,value, need to get the whole map first)
MapSlice loop:
{% for mkey in user.address.mapSlice %}
		Key: {{ mkey }}
		{% for sublevel in user.address.mapSlice.level1 %}
			Subkey: {{ sublevel }}
			{% for item in sublevel %}
				- {{ item }}
			{% endfor %}
		{% endfor %}
{% endfor %}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						MapSlice: testMapSlice,
					},
				},
			},
			expected: `// MapSlice loops (Note: Map cannot be directly iterated with key,value, need to get the whole map first)
MapSlice loop:

		Key: {"key":"level1","value":{"sublevel":["deepitem1","deepitem2","deepitem3"]}}
		
			Subkey: {"key":"sublevel","value":["deepitem1","deepitem2","deepitem3"]}
			
				- {"key":"key","value":"sublevel"}
			
				- {"key":"value","value":["deepitem1","deepitem2","deepitem3"]}
			
		
`,
		},
		{
			name: "Map key iteration",
			template: `// Map loops (Since key,value iteration is not supported, we need to know keys in advance)
Map loop:
key1: {{ user.address.map.key1 }}
key2: {{ user.address.map.key2 }}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						Map: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
			expected: `// Map loops (Since key,value iteration is not supported, we need to know keys in advance)
Map loop:
key1: value1
key2: value2`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := NewTemplate()
			parser := NewParser()
			tpl, err := parser.Parse(tc.template)
			if err != nil {
				t.Fatalf("Template parsing failed: %v", err)
			}
			tmpl.Nodes = tpl.Nodes

			ctx := NewContext()
			for k, v := range tc.context {
				ctx.Set(k, v)
			}

			result, err := tmpl.Execute(ctx)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Template output mismatch.\nExpected (len=%d):\n%q\nGot (len=%d):\n%q",
					len(tc.expected), tc.expected, len(result), result)
			}
		})
	}
}
