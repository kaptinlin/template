package template

import (
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
			expected: "{\n  \"age\": 30,\n  \"name\": \"John Doe\"\n}",
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
			// Template needs to be parsed first
			// ... parsing logic ...
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
