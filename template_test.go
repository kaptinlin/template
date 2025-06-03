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
Categories:{% for k, category in store.categories %}
  {{ k | plus:1 }}. {{ category.name }}
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
Categories:{% for k, category in store.categories %}
  {{ k | plus:1 }}. {{ category.name }}
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

// // TestComplexLoops tests for loops with various complex data structures
// func TestComplexLoops(t *testing.T) {
// 	sourceTemplate := `
// 	// Integer array loop tests
// 	{% for num in numberList %}{{ num }} {% endfor %}

// 	// String loop tests
// 	{% for char in textContent %}{{ char }} {% endfor %}

// 	// Slice element access by index tests
// 	Basic element access: {{ stringArray.0 }}, {{ stringArray.1 }}, {{ stringArray.2 }}
// 	2D array access: {{ matrixData.0.0 }}, {{ matrixData.1.2 }}
// 	Object array access: {{ productList.0.title }}, {{ productList.1.price }}

// 	// 2D string array tests
// 	{% for row in gridData %}
// 		{% for cell in row %}{{ cell }} {% endfor %}
// 	{% endfor %}

// 	// Object array tests
// 	{% for item in productList %}
// 		{{ item.title }} - {{ item.price }}
// 	{% endfor %}

// 	// Nested object tests
// 	{% for item in userSettings %}
// 		{{ item.key }}:
// 		{% for subItem in item.value %}
// 			{{ subItem.key }}: {{ subItem.value }}
// 		{% endfor %}
// 	{% endfor %}

// 	// Complex nested structure tests
// 	{% for item in academicRecords %}
// 		{{ item.key }}:
// 		{% for subItem in item.value %}
// 			{{ subItem.key }}:
// 			{% for grade in subItem.value %}
// 				{{ grade }}
// 			{% endfor %}
// 		{% endfor %}
// 	{% endfor %}

// 	// Multi-level nested tests
// 	{% for item in schoolData %}
// 		{{ item.key }}:
// 		{% for category in item.value %}
// 			{{ category.key }}:
// 			{% for person in category.value %}
// 				{% for detail in person %}
// 					{{ detail.key }}: {{ detail.value }}
// 				{% endfor %}
// 			{% endfor %}
// 		{% endfor %}`

// 	tmpl, err := Parse(sourceTemplate)
// 	if err != nil {
// 		t.Fatalf("Failed to parse template: %v", err)
// 	}

// 	// Create context and set complex nested data
// 	ctx := NewContext()

// 	// Basic data types
// 	ctx.Set("numberList", []int{1, 2, 3, 4, 5})
// 	ctx.Set("textContent", "HelloWorld")

// 	// Index access
// 	ctx.Set("stringArray", []string{"FirstItem", "SecondItem", "ThirdItem"})
// 	ctx.Set("matrixData", [][]string{
// 		{"A1", "A2", "A3"},
// 		{"B1", "B2", "B3"},
// 	})
// 	ctx.Set("productList", []map[string]interface{}{
// 		{"title": "Product1", "price": 100},
// 		{"title": "Product2", "price": 200},
// 	})

// 	// 2D array
// 	ctx.Set("gridData", [][]string{
// 		{"A1", "A2", "A3"},
// 		{"B1", "B2", "B3"},
// 		{"C1", "C2", "C3"},
// 	})

// 	// Object array
// 	ctx.Set("productList", []map[string]interface{}{
// 		{"title": "Product1", "price": 100},
// 		{"title": "Product2", "price": 200},
// 		{"title": "Product3", "price": 300},
// 	})

// 	// Nested objects
// 	userSettings := map[string]map[string]interface{}{
// 		"user1": {
// 			"name": "John",
// 			"age":  30,
// 		},
// 		"user2": {
// 			"name": "Jane",
// 			"age":  25,
// 		},
// 	}
// 	ctx.Set("userSettings", userSettings)

// 	// Complex nested structure
// 	academicRecords := map[string]map[string][]int{
// 		"semester1": {
// 			"math":    []int{90, 85, 92},
// 			"science": []int{88, 79, 95},
// 		},
// 		"semester2": {
// 			"math":    []int{78, 82, 89},
// 			"science": []int{92, 95, 88},
// 		},
// 	}
// 	ctx.Set("academicRecords", academicRecords)

// 	// Multi-level nested structure
// 	schoolData := map[string]map[string][]map[string]string{
// 		"class1": {
// 			"students": []map[string]string{
// 				{"name": "Student1", "grade": "A"},
// 				{"name": "Student2", "grade": "B"},
// 			},
// 			"teachers": []map[string]string{
// 				{"name": "Teacher1", "subject": "Math"},
// 				{"name": "Teacher2", "subject": "English"},
// 			},
// 		},
// 		"class2": {
// 			"students": []map[string]string{
// 				{"name": "Student3", "grade": "A+"},
// 				{"name": "Student4", "grade": "C"},
// 			},
// 			"teachers": []map[string]string{
// 				{"name": "Teacher3", "subject": "Science"},
// 				{"name": "Teacher4", "subject": "Physics"},
// 			},
// 		},
// 	}
// 	ctx.Set("schoolData", schoolData)

// 	// Execute template
// 	output, err := tmpl.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Failed to execute template: %v", err)
// 	}

// 	// Check results of various loop and access patterns
// 	expectedValues := []string{
// 		// Basic array loop
// 		"1 2 3 4 5",
// 		// String loop
// 		"H e l l o W o r l d",
// 		// Index access
// 		"FirstItem, SecondItem, ThirdItem", "A1, B3", "Product1, 200",
// 		// Nested array loop
// 		"A1 A2 A3", "B1 B2 B3", "C1 C2 C3",
// 		// Object array loop
// 		"Product1 - 100", "Product2 - 200", "Product3 - 300",
// 		// Nested object data
// 		"John", "Jane", "30", "25",
// 		// Complex nested structure
// 		"90", "85", "92", "78", "82", "89",
// 		// Multi-level nested structure
// 		"Student1", "Student2", "Teacher1", "Teacher2", "Student3", "Student4", "Teacher3", "Teacher4",
// 		"A", "B", "A+", "C", "Math", "English", "Science", "Physics",
// 	}

// 	for _, expected := range expectedValues {
// 		if !strings.Contains(output, expected) {
// 			t.Errorf("Output doesn't contain expected value: %s", expected)
// 		}
// 	}
// }

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
User: {"name":"John","age":30,"address":{"street":"","city":"","time":"0001-01-01T00:00:00Z","map":null,"slice":null,"mapSlice":null,"sliceMap":null,"department":null}}
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
User address: {"street":"123 Main St","city":"New York","time":"0001-01-01T00:00:00Z","map":null,"slice":null,"mapSlice":null,"sliceMap":null,"department":null}
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
Time: ` + now.Format("2006-01-02 15:04:05"),
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
Slice: [item1, item2, item3]
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
Level 2 key: [deepitem1, deepitem2, deepitem3]
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
Department street: {{ user.address.department.street }}	
Department city: {{ user.address.department.city }}
Department Map: {{ user.address.department.map }}
Department Slice: {{ user.address.department.slice }}
Department MapSlice: {{ user.address.department.mapSlice }}
Department SliceMap: {{ user.address.department.sliceMap }}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
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
			},
			expected: `Department test:
Department street: 123 Main St	
Department city: New York
Department Map: {"key1":"value1","key2":"value2"}
Department Slice: [item1, item2, item3]
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
{% if user.address.time_str %}Time exists: {{ user.address.time_str }}{% endif %}`,
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"address": map[string]interface{}{
						"time_str": now.Format(time.RFC3339),
					},
				},
			},
			expected: `// Time type condition tests
Time exists: ` + now.Format(time.RFC3339),
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
// Slice loop:
// {% for item in user.address.slice %}
// 		- {{ item }}
// {% endfor %}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						Slice: testSlice,
					},
				},
			},
			expected: `// Basic loop tests
// Slice loop:
// 
// 		- item1
// 
// 		- item2
// 
// 		- item3
// `,
		},
		{
			name: "Nested loop test with team members",
			template: `// Nested loop tests
// Team members loop:
// {% for member in team.members %}
// 		Member name: {{ member.name }}, age: {{ member.age }}
// 		Address info: {{ member.address.city }} {{ member.address.street }}
// {% endfor %}`,
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
// Team members loop:
// 
// 		Member name: John, age: 30
// 		Address info: New York 123 Main St
// 
// 		Member name: Jane, age: 25
// 		Address info: Boston 456 Elm St
// `,
		},
		{
			name: "Complex SliceMap iteration",
			template: `// Complex data structure loops
// SliceMap loop:
// {% for arr in user.address.sliceMap %}
// 		Array index:
// 		{% for item in arr %}
// 			Key-value pair: {{ item.key1 }} - {{ item.key2 }}
// 		{% endfor %}
// {% endfor %}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						SliceMap: testSliceMap,
					},
				},
			},
			expected: `// Complex data structure loops
// SliceMap loop:
// 
// 		Array index:
// 		
// 			Key-value pair: nestedvalue1 - nestedvalue2
// 		
// 			Key-value pair: nestedvalue3 - nestedvalue4
// 		
// 
// 		Array index:
// 		
// 			Key-value pair: nestedvalue5 - nestedvalue6
// 		
// `,
		},
		{
			name: "MapSlice iteration",
			template: `// MapSlice direct access (more reliable than iteration)
MapSlice access:
Level1 data: {{ user.address.mapSlice.level1 }}
Sublevel data: {{ user.address.mapSlice.level1.sublevel }}
First item: {{ user.address.mapSlice.level1.sublevel.0 }}`,
			context: map[string]interface{}{
				"user": User{
					Address: Address{
						MapSlice: testMapSlice,
					},
				},
			},
			expected: `// MapSlice direct access (more reliable than iteration)
MapSlice access:
Level1 data: {"sublevel":["deepitem1","deepitem2","deepitem3"]}
Sublevel data: [deepitem1, deepitem2, deepitem3]
First item: deepitem1`,
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
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestPointerTypesAndNesting tests various pointer types and their nesting with maps, slices, and arrays
func TestPointerTypesAndNesting(t *testing.T) {
	// Define basic structures
	type Contact struct {
		Phone *string `json:"phone"`
		Email *string `json:"email"`
		Fax   *string `json:"fax"`
	}

	type Person struct {
		ID       *string            `json:"id"`
		Name     *string            `json:"name"`
		Age      *int               `json:"age"`
		IsActive *bool              `json:"isActive"`
		Contact  *Contact           `json:"contact"`
		Tags     []*string          `json:"tags"`     // pointer slice
		MetaData map[string]*string `json:"metaData"` // map with pointer values
		Projects [3]*string         `json:"projects"` // fixed size pointer array
		Scores   []*int             `json:"scores"`   // pointer slice
	}

	// Create test data with meaningful variable names
	personName := "John Doe"
	personID := "EMP001"
	phoneNumber := "123-456-7890"
	emailAddress := "john.doe@company.com"
	projectAlpha := "Project Alpha"
	projectBeta := "Project Beta"
	skillGo := "Go Programming"
	skillPython := "Python Development"
	departmentInfo := "Engineering Department"

	personAge := 30
	isActive := true
	score1 := 95
	score2 := 88

	contact := &Contact{
		Phone: &phoneNumber,
		Email: &emailAddress,
		Fax:   nil, // test nil pointer
	}

	person := &Person{
		ID:       &personID,
		Name:     &personName,
		Age:      &personAge,
		IsActive: &isActive,
		Contact:  contact,
		Tags:     []*string{&skillGo, &skillPython, nil}, // pointer slice with nil
		MetaData: map[string]*string{
			"department": &departmentInfo,
			"empty":      nil, // test nil pointer
		},
		Projects: [3]*string{&projectAlpha, &projectBeta, nil}, // fixed array with nil
		Scores:   []*int{&score1, &score2},
	}

	testCases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "Basic pointer type access",
			template: `ID: {{ person.id }}
Name: {{ person.name }}
Age: {{ person.age }}
Status: {{ person.isActive }}`,
			context: map[string]interface{}{
				"person": person,
			},
			expected: `ID: "EMP001"
Name: "John Doe"
Age: 30
Status: true`,
		},
		{
			name: "Nested pointer structure access",
			template: `Contact Information:
Phone: {% if person.contact.phone %}{{ person.contact.phone }}{% else %}Not set{% endif %}
Email: {% if person.contact.email %}{{ person.contact.email }}{% else %}Not set{% endif %}
Fax: {% if person.contact.fax %}{{ person.contact.fax }}{% else %}Not set{% endif %}`,
			context: map[string]interface{}{
				"person": person,
			},
			expected: `Contact Information:
Phone: "123-456-7890"
Email: "john.doe@company.com"
Fax: Not set`,
		},
		{
			name: "Pointer slice access",
			template: `Tags: {{ person.tags }}
First tag: {% if person.tags.0 %}{{ person.tags.0 }}{% else %}Empty{% endif %}
Second tag: {% if person.tags.1 %}{{ person.tags.1 }}{% else %}Empty{% endif %}
Third tag: {% if person.tags.2 %}{{ person.tags.2 }}{% else %}Empty{% endif %}`,
			context: map[string]interface{}{
				"person": person,
			},
			expected: `Tags: ["Go Programming","Python Development",null]
First tag: "Go Programming"
Second tag: "Python Development"
Third tag: Empty`,
		},
		{
			name: "Fixed size pointer array access",
			template: `Projects: {{ person.projects }}
Project 1: {% if person.projects.0 %}{{ person.projects.0 }}{% else %}None{% endif %}
Project 2: {% if person.projects.1 %}{{ person.projects.1 }}{% else %}None{% endif %}
Project 3: {% if person.projects.2 %}{{ person.projects.2 }}{% else %}None{% endif %}`,
			context: map[string]interface{}{
				"person": person,
			},
			expected: `Projects: ["Project Alpha","Project Beta",null]
Project 1: "Project Alpha"
Project 2: "Project Beta"
Project 3: None`,
		},
		{
			name: "Map with pointer values access",
			template: `Metadata:
Department: {% if person.metaData.department %}{{ person.metaData.department }}{% else %}Not set{% endif %}
Empty field: {% if person.metaData.empty %}{{ person.metaData.empty }}{% else %}Not set{% endif %}`,
			context: map[string]interface{}{
				"person": person,
			},
			expected: `Metadata:
Department: "Engineering Department"
Empty field: Not set`,
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
				t.Errorf("Template output mismatch.\nExpected:\n%s\nGot:\n%s", tc.expected, result)
			}
		})
	}
}

// TestNestedTypesWithConditions tests nested types used together with if conditions
func TestNestedTypesWithConditions(t *testing.T) {
	type Address struct {
		Street  *string   `json:"street"`
		City    *string   `json:"city"`
		Country *string   `json:"country"`
		Tags    []*string `json:"tags"`
	}

	type User struct {
		Name     *string            `json:"name"`
		Age      *int               `json:"age"`
		IsActive *bool              `json:"isActive"`
		Address  *Address           `json:"address"`
		Skills   []*string          `json:"skills"`
		MetaData map[string]*string `json:"metaData"`
		Scores   []*int             `json:"scores"`
	}

	// Create test data with meaningful variable names
	userName := "Test User"
	userAge := 25
	isActiveUser := true
	streetAddress := "123 Main Street"
	cityName := "New York"
	countryName := "USA"
	skillGo := "Go"
	skillPython := "Python"
	tagImportant := "Important"
	tagVIP := "VIP"
	metaValue := "Test Data"
	score1 := 90
	score2 := 85

	user := &User{
		Name:     &userName,
		Age:      &userAge,
		IsActive: &isActiveUser,
		Address: &Address{
			Street:  &streetAddress,
			City:    &cityName,
			Country: &countryName,
			Tags:    []*string{&tagImportant, &tagVIP, nil},
		},
		Skills: []*string{&skillGo, &skillPython, nil}, // add third nil element for testing
		MetaData: map[string]*string{
			"level": &metaValue,
			"empty": nil,
		},
		Scores: []*int{&score1, &score2},
	}

	// Create empty user for testing nil cases
	emptyUser := &User{
		Name:     nil,
		Age:      nil,
		IsActive: nil,
		Address:  nil,
		Skills:   []*string{},
		MetaData: map[string]*string{},
		Scores:   []*int{},
	}

	testCases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "Basic pointer condition checks",
			template: `{% if user.name %}Name: {{ user.name }}{% else %}Name: Not set{% endif %}
{% if user.age %}Age: {{ user.age }}{% else %}Age: Unknown{% endif %}
{% if user.isActive %}Status: Active{% else %}Status: Inactive{% endif %}`,
			context: map[string]interface{}{
				"user": user,
			},
			expected: `Name: "Test User"
Age: 25
Status: Active`,
		},
		{
			name: "Nested structure condition checks",
			template: `{% if user.address %}
Address information exists:
{% if user.address.street %}  Street: {{ user.address.street }}{% endif %}
{% if user.address.city %}  City: {{ user.address.city }}{% endif %}
{% if user.address.country %}  Country: {{ user.address.country }}{% endif %}
{% else %}
Address information does not exist
{% endif %}`,
			context: map[string]interface{}{
				"user": user,
			},
			expected: `
Address information exists:
  Street: "123 Main Street"
  City: "New York"
  Country: "USA"
`,
		},
		{
			name: "Pointer slice condition checks",
			template: `{% if user.skills %}
Skills list:
{% if user.skills.0 %}  - {{ user.skills.0 }}{% endif %}
{% if user.skills.1 %}  - {{ user.skills.1 }}{% endif %}
{% if user.skills.2 %}  - {{ user.skills.2 }}{% else %}  - Third skill is empty{% endif %}
{% else %}
No skills recorded
{% endif %}`,
			context: map[string]interface{}{
				"user": user,
			},
			expected: `
Skills list:
  - "Go"
  - "Python"
  - Third skill is empty
`,
		},
		{
			name: "Complex condition checks",
			template: `{% if user.name && user.age && user.isActive %}
Complete user info: {{ user.name }} ({{ user.age }} years old)
{% else %}
User information incomplete
{% endif %}

{% if user.address && user.address.city %}
User is from: {{ user.address.city }}
{% else %}
City information unknown
{% endif %}`,
			context: map[string]interface{}{
				"user": user,
			},
			expected: `
Complete user info: "Test User" (25 years old)



User is from: "New York"
`,
		},
		{
			name: "Nil value condition tests",
			template: `{% if emptyUser.name %}Name: {{ emptyUser.name }}{% else %}Name: Not set{% endif %}
{% if emptyUser.address %}Has address info{% else %}No address info{% endif %}
{% if emptyUser.skills %}Skill count: {{ emptyUser.skills | size }}{% else %}No skills recorded{% endif %}`,
			context: map[string]interface{}{
				"emptyUser": emptyUser,
			},
			expected: `Name: Not set
No address info
No skills recorded`,
		},
		{
			name: "Array length condition checks",
			template: `{% if user.skills | size > 1 %}
Multi-skill user: {{ user.skills | size }} skills
{% else %}
Single-skill or no-skill user
{% endif %}

{% if user.scores | size >= 2 %}
Sufficient scores: has {{ user.scores | size }} scores
{% else %}
Insufficient scores
{% endif %}`,
			context: map[string]interface{}{
				"user": user,
			},
			expected: `
Multi-skill user: 3 skills



Sufficient scores: has 2 scores
`,
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
				t.Errorf("Template output mismatch.\nExpected:\n%s\nGot:\n%s", tc.expected, result)
			}
		})
	}
}

// TestNestedTypesWithLoops tests nested types used together with for loops
func TestNestedTypesWithLoops(t *testing.T) {
	type Product struct {
		Name  *string  `json:"name"`
		Price *float64 `json:"price"`
	}

	type Team struct {
		Name    *string   `json:"name"`
		Members []*string `json:"members"`
		Lead    *string   `json:"lead"`
	}

	type Company struct {
		Name     *string             `json:"name"`
		Teams    []*Team             `json:"teams"`
		Products []*Product          `json:"products"`
		Settings map[string]*string  `json:"settings"`
		Budgets  map[string]*float64 `json:"budgets"`
	}

	// Create test data with meaningful variable names
	companyName := "Tech Corp"
	devTeamName := "Development Team"
	qaTeamName := "QA Team"
	teamLead1 := "John Smith"
	teamLead2 := "Jane Doe"
	member1 := "Alice Johnson"
	member2 := "Bob Wilson"
	member3 := "Carol Brown"
	productName1 := "Product Alpha"
	productName2 := "Product Beta"
	price1 := 100.0
	price2 := 200.0
	themeSetting := "Dark Theme"
	localeSetting := "English"
	devBudget := 50000.0
	qaBudget := 30000.0

	company := &Company{
		Name: &companyName,
		Teams: []*Team{
			{
				Name:    &devTeamName,
				Lead:    &teamLead1,
				Members: []*string{&member1, &member2, nil}, // include nil
			},
			{
				Name:    &qaTeamName,
				Lead:    &teamLead2,
				Members: []*string{&member3},
			},
		},
		Products: []*Product{
			{
				Name:  &productName1,
				Price: &price1,
			},
			{
				Name:  &productName2,
				Price: &price2,
			},
		},
		Settings: map[string]*string{
			"theme":  &themeSetting,
			"locale": &localeSetting,
			"empty":  nil, // test nil value
		},
		Budgets: map[string]*float64{
			"development": &devBudget,
			"testing":     &qaBudget,
		},
	}

	testCases := []struct {
		name                string
		template            string
		context             map[string]interface{}
		containsAll         []string // must contain all these strings
		containsNone        []string // must not contain these strings
		exactMatch          string   // exact match (for non-map iteration cases)
		useContainsChecking bool     // whether to use contains checking instead of exact match
	}{
		{
			name: "Pointer slice loops",
			template: `Team Members:
{% for team in company.teams %}
Team: {% if team.name %}{{ team.name }}{% endif %}
Lead: {% if team.lead %}{{ team.lead }}{% endif %}
Members:
{% for member in team.members %}
  - {% if member %}{{ member }}{% else %}Vacant{% endif %}
{% endfor %}
---
{% endfor %}`,
			context: map[string]interface{}{
				"company": company,
			},
			exactMatch: `Team Members:

Team: "Development Team"
Lead: "John Smith"
Members:

  - "Alice Johnson"

  - "Bob Wilson"

  - Vacant

---

Team: "QA Team"
Lead: "Jane Doe"
Members:

  - "Carol Brown"

---
`,
			useContainsChecking: false,
		},
		{
			name: "Pointer struct slice loops",
			template: `Product List:
{% for product in company.products %}
Product: {% if product.name %}{{ product.name }}{% endif %} - Price: {% if product.price %}${{ product.price }}{% endif %}
{% endfor %}`,
			context: map[string]interface{}{
				"company": company,
			},
			exactMatch: `Product List:

Product: "Product Alpha" - Price: $100

Product: "Product Beta" - Price: $200
`,
			useContainsChecking: false,
		},
		{
			name: "Map iteration test (contains checking)",
			template: `Company Settings:
{% for key, value in company.settings %}
{{ key }}: {% if value %}{{ value }}{% else %}Not set{% endif %}
{% endfor %}`,
			context: map[string]interface{}{
				"company": company,
			},
			containsAll: []string{
				"theme: \"Dark Theme\"",
				"locale: \"English\"",
				"empty: Not set",
				"Company Settings:",
			},
			containsNone:        []string{},
			useContainsChecking: true,
		},
		{
			name: "Nested loop test",
			template: `Detailed Team Information:
{% for team in company.teams %}
=== {% if team.name %}{{ team.name }}{% endif %} ===
{% for member in team.members %}
{% if member %}Member: {{ member }}{% else %}Vacant{% endif %}
{% endfor %}
{% endfor %}`,
			context: map[string]interface{}{
				"company": company,
			},
			exactMatch: `Detailed Team Information:

=== "Development Team" ===

Member: "Alice Johnson"

Member: "Bob Wilson"

Vacant


=== "QA Team" ===

Member: "Carol Brown"

`,
			useContainsChecking: false,
		},
		{
			name: "Map with pointer values loop test",
			template: `Budget Allocation:
{% for dept, budget in company.budgets %}
{{ dept }}: {% if budget %}${{ budget }}{% else %}Not allocated{% endif %}
{% endfor %}`,
			context: map[string]interface{}{
				"company": company,
			},
			containsAll: []string{
				"development: $50000",
				"testing: $30000",
				"Budget Allocation:",
			},
			containsNone:        []string{},
			useContainsChecking: true,
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

			if tc.useContainsChecking {
				// Use contains checking to handle map iteration order uncertainty
				for _, required := range tc.containsAll {
					if !strings.Contains(result, required) {
						t.Errorf("Output does not contain required string: %s\nGot:\n%s", required, result)
					}
				}
				for _, forbidden := range tc.containsNone {
					if strings.Contains(result, forbidden) {
						t.Errorf("Output contains forbidden string: %s\nGot:\n%s", forbidden, result)
					}
				}
			} else {
				// Exact match
				if result != tc.exactMatch {
					t.Errorf("Template output mismatch.\nExpected:\n%s\nGot:\n%s", tc.exactMatch, result)
				}
			}
		})
	}
}

// TestComplexNestedDataStructuresWithTemplateFeatures tests comprehensive application of complex nested data structures with template features
func TestComplexNestedDataStructuresWithTemplateFeatures(t *testing.T) {
	// Define complex nested structures
	type Tag struct {
		Name     *string `json:"name"`
		Priority *int    `json:"priority"`
	}

	type Contact struct {
		Email  *string            `json:"email"`
		Phone  *string            `json:"phone"`
		Social map[string]*string `json:"social"`
		Backup []*string          `json:"backup"`
	}

	type Skill struct {
		Name         *string   `json:"name"`
		Level        *int      `json:"level"`
		YearsExp     *float64  `json:"years_exp"`
		Certified    *bool     `json:"certified"`
		Endorsements []*string `json:"endorsements"`
	}

	type Address struct {
		Street   *string            `json:"street"`
		City     *string            `json:"city"`
		Country  *string            `json:"country"`
		PostCode *string            `json:"post_code"`
		Metadata map[string]*string `json:"metadata"`
		Tags     []*Tag             `json:"tags"`
		Coords   [2]*float64        `json:"coords"` // Fixed size pointer array
	}

	type Project struct {
		ID          *string            `json:"id"`
		Name        *string            `json:"name"`
		Budget      *float64           `json:"budget"`
		IsActive    *bool              `json:"is_active"`
		Tags        []*Tag             `json:"tags"`
		TeamMembers []*string          `json:"team_members"`
		Settings    map[string]*string `json:"settings"`
		Milestones  []*string          `json:"milestones"`
	}

	// Additional: More complex nested structures
	type Certification struct {
		Name       *string            `json:"name"`
		IssuedBy   *string            `json:"issued_by"`
		ValidUntil *string            `json:"valid_until"`
		Metadata   map[string]*string `json:"metadata"`
		Tags       []*Tag             `json:"tags"`
	}

	type Team struct {
		Name           *string                   `json:"name"`
		Lead           *string                   `json:"lead"`
		Members        []*string                 `json:"members"`
		Certifications map[string]*Certification `json:"certifications"` // map values are pointer structs
		Projects       map[string]*Project       `json:"projects"`       // map values are pointer structs
		Budget         *float64                  `json:"budget"`
		Settings       map[string]*string        `json:"settings"`
	}

	type Department struct {
		Name         *string             `json:"name"`
		Budget       *float64            `json:"budget"`
		Head         *string             `json:"head"`
		Teams        map[string]*Team    `json:"teams"` // map values are pointer structs
		Projects     []*Project          `json:"projects"`
		Settings     map[string]*string  `json:"settings"`
		Location     *Address            `json:"location"`
		Subsidiaries map[string]*Address `json:"subsidiaries"` // map values are pointer structs
	}

	type Employee struct {
		ID           *string            `json:"id"`
		Name         *string            `json:"name"`
		Age          *int               `json:"age"`
		Salary       *float64           `json:"salary"`
		IsActive     *bool              `json:"is_active"`
		Contact      *Contact           `json:"contact"`
		Address      *Address           `json:"address"`
		Skills       []*Skill           `json:"skills"`
		Tags         []*Tag             `json:"tags"`
		Projects     []*Project         `json:"projects"`
		Department   *Department        `json:"department"`
		Preferences  map[string]*string `json:"preferences"`
		Scores       [3]*int            `json:"scores"`       // Fixed size pointer array
		Achievements map[string][]*Tag  `json:"achievements"` // map values are pointer slices
	}

	// Additional: Deeper nested organizational structure
	type Organization struct {
		Name        *string                        `json:"name"`
		Departments map[string]*Department         `json:"departments"` // map values are pointer structs
		Locations   map[string]*Address            `json:"locations"`
		Teams       map[string]map[string]*Team    `json:"teams"`     // Double-nested map
		Employees   map[string]*Employee           `json:"employees"` // map values are pointer structs
		Partners    map[string]*Organization       `json:"partners"`  // Recursive nested structure
		Settings    map[string]map[string]*string  `json:"settings"`  // Nested map
		Budgets     map[string]map[string]*float64 `json:"budgets"`   // Nested map
	}

	type Company struct {
		Name          *string                        `json:"name"`
		Founded       *int                           `json:"founded"`
		IsPublic      *bool                          `json:"is_public"`
		Revenue       *float64                       `json:"revenue"`
		Employees     []*Employee                    `json:"employees"`
		Departments   []*Department                  `json:"departments"`
		Settings      map[string]*string             `json:"settings"`
		Locations     map[string]*Address            `json:"locations"`
		Projects      map[string][]*Project          `json:"projects"`      // map values are pointer slices
		Budgets       map[string]map[string]*float64 `json:"budgets"`       // Nested map
		Organizations map[string]*Organization       `json:"organizations"` // map values are pointer structs
	}

	// Create test data
	str := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }
	floatPtr := func(f float64) *float64 { return &f }
	boolPtr := func(b bool) *bool { return &b }

	// Create certification data
	javaCert := &Certification{
		Name:       str("Java Certified Developer"),
		IssuedBy:   str("Oracle"),
		ValidUntil: str("2025-12-31"),
		Metadata: map[string]*string{
			"level":     str("Professional"),
			"exam_code": str("1Z0-809"),
		},
		Tags: []*Tag{
			{Name: str("java"), Priority: intPtr(1)},
			{Name: str("backend"), Priority: intPtr(2)},
		},
	}

	awsCert := &Certification{
		Name:       str("AWS Solutions Architect"),
		IssuedBy:   str("Amazon"),
		ValidUntil: str("2026-06-15"),
		Metadata: map[string]*string{
			"level":     str("Associate"),
			"exam_code": str("SAA-C03"),
		},
		Tags: []*Tag{
			{Name: str("aws"), Priority: intPtr(1)},
			{Name: str("cloud"), Priority: intPtr(2)},
		},
	}

	// Create team data
	devTeam := &Team{
		Name:    str("Development Team"),
		Lead:    str("emp-001"),
		Members: []*string{str("emp-001"), str("emp-002"), str("emp-003")},
		Certifications: map[string]*Certification{
			"java": javaCert,
			"aws":  awsCert,
		},
		Projects: map[string]*Project{}, // Set later
		Budget:   floatPtr(300000.0),
		Settings: map[string]*string{
			"methodology": str("agile"),
			"tools":       str("jira,confluence"),
			"remote_work": str("hybrid"),
		},
	}

	qaTeam := &Team{
		Name:    str("QA Team"),
		Lead:    str("emp-004"),
		Members: []*string{str("emp-004"), str("emp-005")},
		Certifications: map[string]*Certification{
			"istqb": {
				Name:       str("ISTQB Foundation"),
				IssuedBy:   str("ISTQB"),
				ValidUntil: str("2027-01-01"),
				Metadata: map[string]*string{
					"level": str("Foundation"),
				},
				Tags: []*Tag{
					{Name: str("testing"), Priority: intPtr(1)},
					{Name: str("quality"), Priority: intPtr(2)},
				},
			},
		},
		Projects: map[string]*Project{}, // Set later
		Budget:   floatPtr(150000.0),
		Settings: map[string]*string{
			"testing_framework": str("selenium"),
			"automation_level":  str("80%"),
		},
	}

	// Create complex test data
	company := &Company{
		Name:     str("TechCorp"),
		Founded:  intPtr(2010),
		IsPublic: boolPtr(true),
		Revenue:  floatPtr(1000000.50),
		Settings: map[string]*string{
			"timezone":   str("UTC"),
			"currency":   str("USD"),
			"work_mode":  str("hybrid"),
			"dress_code": str("casual"),
		},
		Locations: map[string]*Address{
			"hq": {
				Street:   str("123 Tech Street"),
				City:     str("San Francisco"),
				Country:  str("USA"),
				PostCode: str("94102"),
				Metadata: map[string]*string{
					"building_type": str("office"),
					"floor_count":   str("10"),
					"parking":       str("available"),
				},
				Tags: []*Tag{
					{Name: str("headquarters"), Priority: intPtr(1)},
					{Name: str("main_office"), Priority: intPtr(2)},
				},
				Coords: [2]*float64{floatPtr(37.7749), floatPtr(-122.4194)},
			},
			"branch": {
				Street:   str("456 Innovation Ave"),
				City:     str("Austin"),
				Country:  str("USA"),
				PostCode: str("73301"),
				Metadata: map[string]*string{
					"building_type": str("co-working"),
					"floor_count":   str("5"),
					"parking":       str("limited"),
				},
				Tags: []*Tag{
					{Name: str("branch"), Priority: intPtr(3)},
					{Name: str("development"), Priority: intPtr(1)},
				},
				Coords: [2]*float64{floatPtr(30.2672), floatPtr(-97.7431)},
			},
		},
		Budgets: map[string]map[string]*float64{
			"2024": {
				"engineering": floatPtr(500000.0),
				"marketing":   floatPtr(200000.0),
				"operations":  floatPtr(150000.0),
			},
			"2025": {
				"engineering": floatPtr(600000.0),
				"marketing":   floatPtr(250000.0),
				"operations":  floatPtr(180000.0),
			},
		},
		Projects: map[string][]*Project{
			"active": {
				{
					ID:       str("proj-001"),
					Name:     str("AI Platform"),
					Budget:   floatPtr(250000.0),
					IsActive: boolPtr(true),
					Tags: []*Tag{
						{Name: str("ai"), Priority: intPtr(1)},
						{Name: str("platform"), Priority: intPtr(2)},
					},
					TeamMembers: []*string{str("emp-001"), str("emp-002")},
					Settings: map[string]*string{
						"framework": str("tensorflow"),
						"language":  str("python"),
					},
					Milestones: []*string{str("MVP"), str("Beta"), str("GA")},
				},
				{
					ID:       str("proj-002"),
					Name:     str("Mobile App"),
					Budget:   floatPtr(150000.0),
					IsActive: boolPtr(true),
					Tags: []*Tag{
						{Name: str("mobile"), Priority: intPtr(1)},
						{Name: str("app"), Priority: intPtr(2)},
					},
					TeamMembers: []*string{str("emp-003"), str("emp-001")},
					Settings: map[string]*string{
						"platform": str("react-native"),
						"language": str("javascript"),
					},
					Milestones: []*string{str("Design"), str("Development"), str("Testing")},
				},
			},
			"completed": {
				{
					ID:       str("proj-003"),
					Name:     str("Website Redesign"),
					Budget:   floatPtr(75000.0),
					IsActive: boolPtr(false),
					Tags: []*Tag{
						{Name: str("web"), Priority: intPtr(1)},
						{Name: str("design"), Priority: intPtr(2)},
					},
					TeamMembers: []*string{str("emp-002")},
					Settings: map[string]*string{
						"framework": str("react"),
						"language":  str("typescript"),
					},
					Milestones: []*string{str("Research"), str("Design"), str("Implementation")},
				},
			},
		},
		Employees: []*Employee{
			{
				ID:       str("emp-001"),
				Name:     str("Alice Johnson"),
				Age:      intPtr(30),
				Salary:   floatPtr(120000.0),
				IsActive: boolPtr(true),
				Contact: &Contact{
					Email: str("alice@techcorp.com"),
					Phone: str("+1-555-0101"),
					Social: map[string]*string{
						"linkedin": str("alice-johnson"),
						"github":   str("alicejohnson"),
						"twitter":  str("@alice_codes"),
					},
					Backup: []*string{str("alice.personal@gmail.com"), str("+1-555-0102")},
				},
				Address: &Address{
					Street:   str("789 Maple St"),
					City:     str("San Francisco"),
					Country:  str("USA"),
					PostCode: str("94103"),
					Metadata: map[string]*string{
						"building_type": str("apartment"),
						"floor":         str("3"),
					},
					Tags: []*Tag{
						{Name: str("residential"), Priority: intPtr(1)},
					},
					Coords: [2]*float64{floatPtr(37.7849), floatPtr(-122.4094)},
				},
				Skills: []*Skill{
					{
						Name:         str("Python"),
						Level:        intPtr(9),
						YearsExp:     floatPtr(5.5),
						Certified:    boolPtr(true),
						Endorsements: []*string{str("John Doe"), str("Jane Smith")},
					},
					{
						Name:         str("Machine Learning"),
						Level:        intPtr(8),
						YearsExp:     floatPtr(3.0),
						Certified:    boolPtr(true),
						Endorsements: []*string{str("Dr. Brown"), str("Prof. Wilson")},
					},
				},
				Tags: []*Tag{
					{Name: str("senior"), Priority: intPtr(1)},
					{Name: str("fullstack"), Priority: intPtr(2)},
				},
				Preferences: map[string]*string{
					"work_hours":   str("flexible"),
					"remote_work":  str("yes"),
					"team_size":    str("small"),
					"project_type": str("technical"),
				},
				Scores: [3]*int{intPtr(95), intPtr(87), intPtr(92)},
				Achievements: map[string][]*Tag{
					"2023": {
						{Name: str("best_performer"), Priority: intPtr(1)},
						{Name: str("innovation_award"), Priority: intPtr(2)},
					},
					"2024": {
						{Name: str("team_lead"), Priority: intPtr(1)},
						{Name: str("mentor"), Priority: intPtr(3)},
					},
				},
			},
			{
				ID:       str("emp-002"),
				Name:     str("Bob Smith"),
				Age:      intPtr(28),
				Salary:   floatPtr(95000.0),
				IsActive: boolPtr(true),
				Contact: &Contact{
					Email: str("bob@techcorp.com"),
					Phone: str("+1-555-0201"),
					Social: map[string]*string{
						"linkedin": str("bob-smith"),
						"github":   str("bobsmith"),
					},
					Backup: []*string{str("bob.personal@yahoo.com")},
				},
				Address: &Address{
					Street:   str("321 Oak Ave"),
					City:     str("Austin"),
					Country:  str("USA"),
					PostCode: str("73302"),
					Metadata: map[string]*string{
						"building_type": str("house"),
						"garage":        str("yes"),
					},
					Tags: []*Tag{
						{Name: str("residential"), Priority: intPtr(1)},
						{Name: str("suburb"), Priority: intPtr(2)},
					},
					Coords: [2]*float64{floatPtr(30.2772), floatPtr(-97.7531)},
				},
				Skills: []*Skill{
					{
						Name:         str("JavaScript"),
						Level:        intPtr(8),
						YearsExp:     floatPtr(4.0),
						Certified:    boolPtr(false),
						Endorsements: []*string{str("Alice Johnson")},
					},
					{
						Name:         str("React"),
						Level:        intPtr(7),
						YearsExp:     floatPtr(2.5),
						Certified:    boolPtr(true),
						Endorsements: []*string{str("Senior Dev"), str("Tech Lead")},
					},
				},
				Tags: []*Tag{
					{Name: str("frontend"), Priority: intPtr(1)},
					{Name: str("junior"), Priority: intPtr(2)},
				},
				Preferences: map[string]*string{
					"work_hours":   str("standard"),
					"remote_work":  str("hybrid"),
					"team_size":    str("medium"),
					"project_type": str("frontend"),
				},
				Scores: [3]*int{intPtr(82), intPtr(90), intPtr(85)},
				Achievements: map[string][]*Tag{
					"2024": {
						{Name: str("fast_learner"), Priority: intPtr(1)},
						{Name: str("team_player"), Priority: intPtr(2)},
					},
				},
			},
		},
		Departments: []*Department{
			{
				Name:   str("Engineering"),
				Budget: floatPtr(500000.0),
				Head:   str("emp-001"),
				Teams: map[string]*Team{
					"development": devTeam,
					"qa":          qaTeam,
				},
				Settings: map[string]*string{
					"methodology": str("agile"),
					"tools":       str("jira,github"),
					"code_review": str("required"),
				},
				Location: &Address{
					Street:   str("123 Tech Street, Floor 5"),
					City:     str("San Francisco"),
					Country:  str("USA"),
					PostCode: str("94102"),
					Metadata: map[string]*string{
						"floor":     str("5"),
						"capacity":  str("50"),
						"equipment": str("high-end"),
					},
					Tags: []*Tag{
						{Name: str("engineering"), Priority: intPtr(1)},
						{Name: str("development"), Priority: intPtr(2)},
					},
					Coords: [2]*float64{floatPtr(37.7749), floatPtr(-122.4194)},
				},
				Subsidiaries: map[string]*Address{
					"lab": {
						Street:  str("789 Research Blvd"),
						City:    str("Palo Alto"),
						Country: str("USA"),
						Metadata: map[string]*string{
							"type":     str("research_lab"),
							"security": str("high"),
						},
						Coords: [2]*float64{floatPtr(37.4419), floatPtr(-122.1430)},
					},
					"datacenter": {
						Street:  str("321 Server Farm Rd"),
						City:    str("Denver"),
						Country: str("USA"),
						Metadata: map[string]*string{
							"type":         str("datacenter"),
							"power_backup": str("redundant"),
						},
						Coords: [2]*float64{floatPtr(39.7392), floatPtr(-104.9903)},
					},
				},
			},
		},
	}

	// Set team project references
	devTeam.Projects = map[string]*Project{
		"ai_platform": company.Projects["active"][0],
		"mobile_app":  company.Projects["active"][1],
	}
	qaTeam.Projects = map[string]*Project{
		"mobile_app": company.Projects["active"][1],
		"website":    company.Projects["completed"][0],
	}

	// Set employee department and project references
	company.Employees[0].Department = company.Departments[0]
	company.Employees[1].Department = company.Departments[0]
	company.Employees[0].Projects = []*Project{company.Projects["active"][0], company.Projects["active"][1]}
	company.Employees[1].Projects = []*Project{company.Projects["active"][1], company.Projects["completed"][0]}

	// Set department project references
	company.Departments[0].Projects = append(company.Projects["active"], company.Projects["completed"]...)

	// Create organizational structure (after company creation)
	company.Organizations = map[string]*Organization{
		"north_america": {
			Name: str("North America Division"),
			Departments: map[string]*Department{
				"engineering": company.Departments[0], // Now safe to reference
			},
			Teams: map[string]map[string]*Team{
				"engineering": {
					"dev": devTeam,
					"qa":  qaTeam,
				},
			},
			Settings: map[string]map[string]*string{
				"hr": {
					"vacation_days": str("25"),
					"sick_days":     str("10"),
				},
				"it": {
					"laptop_budget":   str("2000"),
					"software_budget": str("500"),
				},
			},
		},
	}

	// Create context
	ctx := NewContext()
	ctx.Set("company", company)

	// Test cases - mixed exact match and contains checks
	cases := []struct {
		name         string
		template     string
		expected     string
		containsAll  []string // Must contain all these strings
		containsNone []string // Must not contain these strings
		useContains  bool     // Whether to use contains checking instead of exact match
	}{
		{
			name: "Basic company information",
			template: `Company name: {{ company.name }}
Founded year: {{ company.founded }}
Is public: {{ company.is_public }}
Annual revenue: ${{ company.revenue }}`,
			expected: `Company name: "TechCorp"
Founded year: 2010
Is public: true
Annual revenue: $1000000.5`,
			useContains: false,
		},
		{
			name: "Test if conditions with pointer fields",
			template: `{% if company.is_public %}{{ company.name }} is a public company{% endif %}
{% if company.revenue > 500000 %}Company revenue exceeds 500,000 USD{% endif %}
{% if company.founded < 2015 %}Company was founded early{% endif %}`,
			expected: `"TechCorp" is a public company
Company revenue exceeds 500,000 USD
Company was founded early`,
			useContains: false,
		},
		{
			name: "for loop - single variable iteration over employees",
			template: `Employee list:
{% for employee in company.employees %}
- {{ employee.name }} ({{ employee.age }} years, ${{ employee.salary }})
{% endfor %}`,
			expected: `Employee list:

- "Alice Johnson" (30 years, $120000)

- "Bob Smith" (28 years, $95000)
`,
			useContains: false,
		},
		{
			name: "for loop - double variable iteration over company settings (using contains check)",
			template: `Company settings:
{% for key, value in company.settings %}
{{ key }}: {{ value }}
{% endfor %}`,
			containsAll: []string{
				"Company settings:",
				"timezone: \"UTC\"",
				"currency: \"USD\"",
				"work_mode: \"hybrid\"",
				"dress_code: \"casual\"",
			},
			useContains: true,
		},
		{
			name: "for loop - double variable iteration over locations (using contains check)",
			template: `Office locations:
{% for location_name, location in company.locations %}
{{ location_name }}: {{ location.city }}, {{ location.country }}
Address: {{ location.street }}
{% endfor %}`,
			containsAll: []string{
				"Office locations:",
				"hq: \"San Francisco\", \"USA\"",
				"Address: \"123 Tech Street\"",
				"branch: \"Austin\", \"USA\"",
				"Address: \"456 Innovation Ave\"",
			},
			useContains: true,
		},
		{
			name: "nested for loop - employee skills",
			template: `Employee skill details:
{% for employee in company.employees %}
{{ employee.name }}'s skills:
{% for skill in employee.skills %}
  - {{ skill.name }}: Level {{ skill.level }}, {{ skill.years_exp }} years experience, Certified: {{ skill.certified }}
{% endfor %}
{% endfor %}`,
			expected: `Employee skill details:

"Alice Johnson"'s skills:

  - "Python": Level 9, 5.5 years experience, Certified: true

  - "Machine Learning": Level 8, 3 years experience, Certified: true


"Bob Smith"'s skills:

  - "JavaScript": Level 8, 4 years experience, Certified: false

  - "React": Level 7, 2.5 years experience, Certified: true

`,
			useContains: false,
		},
		{
			name: "complex nested access - employee contact information (using contains check)",
			template: `Employee contact information:
{% for employee in company.employees %}
{{ employee.name }}:
   Email: {{ employee.contact.email }}
   Phone: {{ employee.contact.phone }}
   Social media:
{% for platform, handle in employee.contact.social %}
    {{ platform }}: {{ handle }}
{% endfor %}
{% endfor %}`,
			containsAll: []string{
				"Employee contact information:",
				"\"Alice Johnson\":",
				"Email: \"alice@techcorp.com\"",
				"Phone: \"+1-555-0101\"",
				"Social media:",
				"linkedin: \"alice-johnson\"",
				"github: \"alicejohnson\"",
				"twitter: \"@alice_codes\"",
				"\"Bob Smith\":",
				"Email: \"bob@techcorp.com\"",
				"Phone: \"+1-555-0201\"",
				"linkedin: \"bob-smith\"",
				"github: \"bobsmith\"",
			},
			useContains: true,
		},
		{
			name: "fixed size array access",
			template: `Employee ratings:
{% for employee in company.employees %}
{{ employee.name }} ratings: [{{ employee.scores.0 }}, {{ employee.scores.1 }}, {{ employee.scores.2 }}]
{% endfor %}`,
			expected: `Employee ratings:

"Alice Johnson" ratings: [95, 87, 92]

"Bob Smith" ratings: [82, 90, 85]
`,
			useContains: false,
		},
		{
			name: "map value as complex structure (using contains check)",
			template: `Project status report:
{% for status, projects in company.projects %}
{{ status }} project:
{% for project in projects %}
  - {{ project.name }} (Budget: ${{ project.budget }})
     Team members: {% for member in project.team_members %}{{ member }} {% endfor %}
{% endfor %}
{% endfor %}`,
			containsAll: []string{
				"Project status report:",
				"active project:",
				"- \"AI Platform\" (Budget: $250000)",
				"Team members: \"emp-001\" \"emp-002\"",
				"- \"Mobile App\" (Budget: $150000)",
				"Team members: \"emp-003\" \"emp-001\"",
				"completed project:",
				"- \"Website Redesign\" (Budget: $75000)",
				"Team members: \"emp-002\"",
			},
			useContains: true,
		},
		{
			name: "complex if condition nested access",
			template: `Employee evaluation:
{% for employee in company.employees %}
{% if employee.salary > 100000 %}
High-salary employee {{ employee.name }}:
  - Living city: {{ employee.address.city }}
  - Department: {{ employee.department.name }}
  {% if employee.skills %}
   Core skills:
  {% for skill in employee.skills %}
    {% if skill.level >= 8 %}
    - {{ skill.name }} (Expert Level {{ skill.level }})
    {% endif %}
  {% endfor %}
  {% endif %}
{% endif %}
{% endfor %}`,
			expected: `Employee evaluation:


High-salary employee "Alice Johnson":
  - Living city: "San Francisco"
  - Department: "Engineering"
  
   Core skills:
  
    
    - "Python" (Expert Level 9)
    
  
    
    - "Machine Learning" (Expert Level 8)
    
  
  


`,
			useContains: false,
		},
		{
			name: "complex nested array and coordinate access (using contains check)",
			template: `Office location coordinates:
{% for name, location in company.locations %}
{{ name }} ({{ location.city }}): 
   Coordinates: [{{ location.coords.0 }}, {{ location.coords.1 }}]
   Metadata:
{% for key, value in location.metadata %}
    {{ key }}: {{ value }}
{% endfor %}
{% endfor %}`,
			containsAll: []string{
				"Office location coordinates:",
				"hq (\"San Francisco\"):",
				"Coordinates: [37.7749, -122.4194]",
				"Metadata:",
				"building_type: \"office\"",
				"floor_count: \"10\"",
				"parking: \"available\"",
				"branch (\"Austin\"):",
				"Coordinates: [30.2672, -97.7431]",
				"building_type: \"co-working\"",
				"floor_count: \"5\"",
				"parking: \"limited\"",
			},
			useContains: true,
		},
		{
			name: "employee achievements record (using contains check)",
			template: `Employee achievements:
{% for employee in company.employees %}
{{ employee.name }}:
{% for year, achievements in employee.achievements %}
  {{ year }} year:
{% for achievement in achievements %}
    - {{ achievement.name }} (Priority: {{ achievement.priority }})
{% endfor %}
{% endfor %}
{% endfor %}`,
			containsAll: []string{
				"Employee achievements:",
				"\"Alice Johnson\":",
				"2023 year:",
				"- \"best_performer\" (Priority: 1)",
				"- \"innovation_award\" (Priority: 2)",
				"2024 year:",
				"- \"team_lead\" (Priority: 1)",
				"- \"mentor\" (Priority: 3)",
				"\"Bob Smith\":",
				"- \"fast_learner\" (Priority: 1)",
				"- \"team_player\" (Priority: 2)",
			},
			useContains: true,
		},
		{
			name: "map value as pointer struct - department teams",
			template: `Department team information:
{% for dept in company.departments %}
Department: {{ dept.name }}
Team:
{% for team_name, team in dept.teams %}
  {{ team_name }}: {{ team.name }} (Lead: {{ team.lead }}, Budget: ${{ team.budget }})
{% endfor %}
{% endfor %}`,
			containsAll: []string{
				"Department team information:",
				"Department: \"Engineering\"",
				"Team:",
				"development: \"Development Team\" (Lead: \"emp-001\", Budget: $300000)",
				"qa: \"QA Team\" (Lead: \"emp-004\", Budget: $150000)",
			},
			useContains: true,
		},
		{
			name: "deeply nested - team certification information",
			template: `Team certification details:
{% for dept in company.departments %}
{% for team_name, team in dept.teams %}
{{ team.name }} certification:
{% for cert_name, cert in team.certifications %}
  - {{ cert.name }} ({{ cert.issued_by }}, Valid until: {{ cert.valid_until }})
     Tags:{% for tag in cert.tags %} {{ tag.name }}{% endfor %}
{% endfor %}
{% endfor %}
{% endfor %}`,
			containsAll: []string{
				"Team certification details:",
				"\"Development Team\" certification:",
				"- \"Java Certified Developer\" (\"Oracle\", Valid until: \"2025-12-31\")",
				"Tags: \"java\" \"backend\"",
				"- \"AWS Solutions Architect\" (\"Amazon\", Valid until: \"2026-06-15\")",
				"Tags: \"aws\" \"cloud\"",
				"\"QA Team\" certification:",
				"- \"ISTQB Foundation\" (\"ISTQB\", Valid until: \"2027-01-01\")",
				"Tags: \"testing\" \"quality\"",
			},
			useContains: true,
		},
		{
			name: "multi-level nested - department subsidiary addresses",
			template: `Department subsidiary information:
{% for dept in company.departments %}
{{ dept.name }} subsidiary:
{% for sub_name, address in dept.subsidiaries %}
  {{ sub_name }}: {{ address.city }}, {{ address.street }}
   Coordinates: [{{ address.coords.0 }}, {{ address.coords.1 }}]
   Metadata:
{% for key, value in address.metadata %}
    {{ key }}: {{ value }}
{% endfor %}
{% endfor %}
{% endfor %}`,
			containsAll: []string{
				"Department subsidiary information:",
				"\"Engineering\" subsidiary:",
				"lab: \"Palo Alto\", \"789 Research Blvd\"",
				"Coordinates: [37.4419, -122.143]",
				"type: \"research_lab\"",
				"security: \"high\"",
				"datacenter: \"Denver\", \"321 Server Farm Rd\"",
				"Coordinates: [39.7392, -104.9903]",
				"type: \"datacenter\"",
				"power_backup: \"redundant\"",
			},
			useContains: true,
		},
		{
			name: "extremely deeply nested - organizational settings",
			template: `Organizational settings details:
{% for org_name, org in company.organizations %}
{{ org.name }}:
{% for category, settings in org.settings %}
  {{ category }} configuration:
{% for key, value in settings %}
    {{ key }}: {{ value }}
{% endfor %}
{% endfor %}
{% endfor %}`,
			containsAll: []string{
				"Organizational settings details:",
				"\"North America Division\":",
				"hr configuration:",
				"vacation_days: \"25\"",
				"sick_days: \"10\"",
				"it configuration:",
				"laptop_budget: \"2000\"",
				"software_budget: \"500\"",
			},
			useContains: true,
		},
		{
			name: "three-level nested map - organizational teams",
			template: `Organizational team structure:
{% for org_name, org in company.organizations %}
{{ org.name }}:
{% for dept_name, teams in org.teams %}
  {{ dept_name }} department:
{% for team_name, team in teams %}
    {{ team_name }}: {{ team.name }} ({{ team.members | size }} members)
{% endfor %}
{% endfor %}
{% endfor %}`,
			containsAll: []string{
				"Organizational team structure:",
				"\"North America Division\":",
				"engineering department:",
				"dev: \"Development Team\" (3 members)",
				"qa: \"QA Team\" (2 members)",
			},
			useContains: true,
		},
	}

	// Execute tests
	parser := NewParser()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			template, err := parser.Parse(tc.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := template.Execute(ctx)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if tc.useContains {
				// Use contains checking mode, suitable for map iteration order uncertainty
				for _, required := range tc.containsAll {
					if !strings.Contains(result, required) {
						t.Errorf("Missing required content: %s\nActual output:\n%s", required, result)
					}
				}
				for _, forbidden := range tc.containsNone {
					if strings.Contains(result, forbidden) {
						t.Errorf("Output contains forbidden content: %s\nActual output:\n%s", forbidden, result)
					}
				}
			} else {
				// Exact match
				if strings.TrimSpace(result) != strings.TrimSpace(tc.expected) {
					t.Errorf("Template output mismatch\nExpected:\n%s\nActual:\n%s", tc.expected, result)
				}
			}
		})
	}
}
