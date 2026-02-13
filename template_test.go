package template

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestIfConditions(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Split string",
			tmpl: "{{ message | split:',' }}",
			context: map[string]any{
				"message": "one,two,three",
			},
			expected: "[one,two,three]",
		},
		{
			name: "Simple condition",
			tmpl: "{% if age >= 18 %}Adult{% endif %}",
			context: map[string]any{
				"age": 20,
			},
			expected: "Adult",
		},
		{
			name: "Simple condition with filter",
			tmpl: "{% if age >= 18 %}{{ name | upper }}{% endif %}",
			context: map[string]any{
				"age":  20,
				"name": "alexander",
			},
			expected: "ALEXANDER",
		},
		{
			name: "Condition with filter",
			tmpl: "{% if name | length > 5 %}Name too long{% endif %}",
			context: map[string]any{
				"name": "Alexander",
			},
			expected: "Name too long",
		},
		{
			name: "Multiple conditions",
			tmpl: "{% if score >= 90 %}Excellent{% if score == 100 %}Perfect{% endif %}{% endif %}",
			context: map[string]any{
				"score": 100,
			},
			expected: "ExcellentPerfect",
		},
		{
			name: "Multiple filters in condition",
			tmpl: "{% if message | trim | upper | length > 0 %}Has content{% endif %}",
			context: map[string]any{
				"message": "  Hello  ",
			},
			expected: "Has content",
		},
		{
			name: "Complex object condition",
			tmpl: "{% if user.age >= 18 && user.name | length > 3 %}{{ user.name | upper }} is adult{% endif %}",
			context: map[string]any{
				"user": map[string]any{
					"name": "Alice",
					"age":  20,
				},
			},
			expected: "ALICE is adult",
		},
		{
			name: "String comparison",
			tmpl: "{% if status | lower == 'active' %}Currently active{% endif %}",
			context: map[string]any{
				"status": "ACTIVE",
			},
			expected: "Currently active",
		},
		{
			name: "Empty value handling",
			tmpl: "{% if content | default:'nil' == 'nil' %}No content{% endif %}",
			context: map[string]any{
				"content": "",
			},
			expected: "No content",
		},
		{
			name: "Basic boolean",
			tmpl: "{% if isActive %}Active{% endif %}",
			context: map[string]any{
				"isActive": true,
			},
			expected: "Active",
		},
		{
			name: "Numeric equality",
			tmpl: "{% if count == 0 %}Empty{% else %}Not empty{% endif %}",
			context: map[string]any{
				"count": 0,
			},
			expected: "Empty",
		},
		{
			name: "String equality",
			tmpl: "{% if status == 'pending' %}Pending{% endif %}",
			context: map[string]any{
				"status": "pending",
			},
			expected: "Pending",
		},
		{
			name: "AND operation",
			tmpl: "{% if isAdmin && isActive %}Admin online{% endif %}",
			context: map[string]any{
				"isAdmin":  true,
				"isActive": true,
			},
			expected: "Admin online",
		},
		{
			name: "OR operation",
			tmpl: "{% if isVIP || isMember %}Access granted{% endif %}",
			context: map[string]any{
				"isVIP":    false,
				"isMember": true,
			},
			expected: "Access granted",
		},
		{
			name: "Simple condition with else",
			tmpl: "{% if age >= 18 %}Adult{% else %}Minor{% endif %}",
			context: map[string]any{
				"age": 16,
			},
			expected: "Minor",
		},
		{
			name: "String comparison with else",
			tmpl: "{% if status == 'active' %}Currently active{% else %}Inactive{% endif %}",
			context: map[string]any{
				"status": "inactive",
			},
			expected: "Inactive",
		},
		{
			name: "Numeric comparison with else",
			tmpl: "{% if score >= 50 %}Pass{% else %}Fail{% endif %}",
			context: map[string]any{
				"score": 45,
			},
			expected: "Fail",
		},
		{
			name: "Complex condition with else",
			tmpl: "{% if user.age >= 18 && user.name | length > 3 %}{{ user.name | upper }} is adult{% else %}Not an adult{% endif %}",
			context: map[string]any{
				"user": map[string]any{
					"name": "Bob",
					"age":  16,
				},
			},
			expected: "Not an adult",
		},
		{
			name: "StringArraySizeCheck",
			tmpl: "{% if names | size > 2 %}More than two names{% else %}Two or fewer names{% endif %}",
			context: map[string]any{
				"names": []string{"John", "Alice", "Bob"},
			},
			expected: "More than two names",
		},
		{
			name: "EmptyIntArrayCheck",
			tmpl: "{% if !scores %}No scores recorded{% endif %}",
			context: map[string]any{
				"scores": []int{},
			},
			expected: "No scores recorded",
		},
		{
			name: "EmptyStringArrayCheck",
			tmpl: "{% if !names %}No names available{% endif %}",
			context: map[string]any{
				"names": []string{},
			},
			expected: "No names available",
		},
		{
			name: "NestedMapWithEmployeeCountCheck",
			tmpl: "{% if department.employees | size >= 3 %}Large department{% else %}Small department{% endif %}",
			context: map[string]any{
				"department": map[string]any{
					"name": "R&D",
					"employees": []any{
						map[string]any{"name": "John", "role": "Developer"},
						map[string]any{"name": "Alice", "role": "Tester"},
						map[string]any{"name": "Bob", "role": "Product Manager"},
					},
				},
			},
			expected: "Large department",
		},
		{
			name: "NonEmptyMapCheck",
			tmpl: "{% if !settings %}Has settings{% endif %}",
			context: map[string]any{
				"settings": map[string]any{
					"theme":   "Dark",
					"High":    "High priority",
					"enabled": "Feature enabled",
				},
			},
			expected: "",
		},
		{
			name: "NumericTypeComparison",
			tmpl: "{% if int_value > float_value %}Integer greater than float{% else %}Float greater than or equal to integer{% endif %}",
			context: map[string]any{
				"int_value":   5,
				"float_value": 5.5,
			},
			expected: "Float greater than or equal to integer",
		},
		{
			name: "EmptyMapCheck",
			tmpl: "{% if !config %}Configuration needed{% endif %}",
			context: map[string]any{
				"config": map[string]any{},
			},
			expected: "Configuration needed",
		},
		{
			name: "BooleanArrayWithTrueValueCheck",
			tmpl: "{% if !flags %}At least one enabled{% else %}All disabled{% endif %}",
			context: map[string]any{
				"flags": []bool{false, false, true, false},
			},
			expected: "All disabled",
		},
		{
			name: "NestedStringArraySizeCheck",
			tmpl: "{% if struct.names | size > 2 %}More than two names{% else %}Two or fewer names{% endif %}",
			context: map[string]any{
				"struct": map[string]any{
					"names": []string{"John", "Alice", "Bob"},
				},
			},
			expected: "More than two names",
		},
		{
			name: "NonEmptyStringArrayCondition",
			tmpl: "{% if struct.names %}name{% endif %}",
			context: map[string]any{
				"struct": map[string]any{
					"names": []string{"John", "Alice", "Bob"},
				},
			},
			expected: "name",
		},
		{
			name: "EmptyStringArrayCondition",
			tmpl: "{% if struct.names %}name{% endif %}",
			context: map[string]any{
				"struct": map[string]any{
					"names": []string{},
				},
			},
			expected: "",
		},
		{
			name: "EmptyMapCondition",
			tmpl: "{% if struct.names %}name{% endif %}",
			context: map[string]any{
				"struct": map[string]any{
					"names": map[string]any{},
				},
			},
			expected: "",
		},
		{
			name: "NonEmptyMapCondition",
			tmpl: "{% if struct.names %}name{% endif %}",
			context: map[string]any{
				"struct": map[string]any{
					"names": map[string]any{
						"theme":   "Dark",
						"High":    "High priority",
						"enabled": "Feature enabled",
					},
				},
			},
			expected: "name",
		},
		{
			name: "ArrayLengthComparison",
			tmpl: "{% if shortList | size < longList | size %}Shorter array{% else %}Equal or longer array{% endif %}",
			context: map[string]any{
				"shortList": []string{"a", "b"},
				"longList":  []string{"x", "y", "z"},
			},
			expected: "Shorter array",
		},
		{
			name: "MapKeysEqualityCheck",
			tmpl: "{% if config1.debug == config2.debug %}Same debug setting{% else %}Different debug settings{% endif %}",
			context: map[string]any{
				"config1": map[string]any{"debug": true, "mode": "development"},
				"config2": map[string]any{"debug": true, "mode": "production"},
			},
			expected: "Same debug setting",
		},
		{
			name: "MapNegationCondition",
			tmpl: "{% if !emptyMap && !nonEmptyMap %}Both false{% else %}Both true{% endif %}",
			context: map[string]any{
				"emptyMap":    map[string]any{},
				"nonEmptyMap": map[string]any{"key": "value"},
			},
			expected: "Both true",
		},
		{
			name: "ComplexNestedCollectionComparison",
			tmpl: "{% if data.users | size > data.groups | size %}More users than groups{% else %}Equal or more groups than users{% endif %}",
			context: map[string]any{
				"data": map[string]any{
					"users": []string{"User1", "User2", "User3"},
					"groups": []any{
						map[string]any{"name": "Admins"},
						map[string]any{"name": "Users"},
					},
				},
			},
			expected: "More users than groups",
		},
		{
			name: "SliceWithANDOperator",
			tmpl: "{% if numbers && numbers | sum > 10 %}Non-empty array with sum > 10{% endif %}",
			context: map[string]any{
				"numbers": []int{2, 4, 6, 8},
			},
			expected: "Non-empty array with sum > 10",
		},
		{
			name: "MapWithOROperator",
			tmpl: "{% if emptyConfig || defaultConfig %}Using config{% endif %}",
			context: map[string]any{
				"emptyConfig":   map[string]any{},
				"defaultConfig": map[string]any{"theme": "light"},
			},
			expected: "Using config",
		},
		{
			name: "NestedArraysInNestedMaps",
			tmpl: "{% if project.teams.developers | size > project.teams.designers | size %}More developers{% else %}Equal or more designers{% endif %}",
			context: map[string]any{
				"project": map[string]any{
					"teams": map[string]any{
						"developers": []string{"Dev1", "Dev2", "Dev3", "Dev4"},
						"designers":  []string{"Des1", "Des2"},
					},
				},
			},
			expected: "More developers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestComplexForLoop(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Nested data structure",
			tmpl: `Store: {{ store.name }}
Categories:{% for k, category in store.categories %}
  {{ k | plus:1 }}. {{ category.name }}
  Products:{% for product in category.products %}
    - {{ product.name }}: ${{ product.price }}{% endfor %}{% endfor %}

Total Products: {{ store.products | size }}`,
			context: map[string]any{
				"store": map[string]any{
					"name": "My Store",
					"categories": []any{
						map[string]any{
							"name": "Electronics",
							"products": []any{
								map[string]any{
									"name":  "Phone",
									"price": 599,
								},
								map[string]any{
									"name":  "Laptop",
									"price": 999,
								},
							},
						},
						map[string]any{
							"name": "Books",
							"products": []any{
								map[string]any{
									"name":  "Python Book",
									"price": 39,
								},
							},
						},
					},
					"products": []any{
						map[string]any{
							"name":  "Phone",
							"price": 599,
						},
						map[string]any{
							"name":  "Laptop",
							"price": 999,
						},
						map[string]any{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNestedLoopWithConditions(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Complex nested structure with conditions",
			tmpl: `Store: {{ store.name }}
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
			context: map[string]any{
				"store": map[string]any{
					"name": "Tech Marketplace",
					"categories": []any{
						map[string]any{
							"name": "Smartphones",
							"products": []any{
								map[string]any{
									"name":     "Premium Phone",
									"price":    899,
									"stock":    15,
									"discount": 0.9,
									"rating":   4.7,
									"featured": true,
								},
								map[string]any{
									"name":     "Budget Phone",
									"price":    299,
									"stock":    0,
									"rating":   4.2,
									"featured": false,
								},
							},
						},
						map[string]any{
							"name": "Laptops",
							"products": []any{
								map[string]any{
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
					"products": []any{
						map[string]any{
							"name":     "Premium Phone",
							"featured": true,
							"rating":   4.7,
						},
						map[string]any{
							"name":     "Pro Laptop",
							"featured": true,
							"rating":   4.8,
						},
						map[string]any{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("output mismatch (want len=%d, got len=%d):\nwant: %q\ngot:  %q",
					len(tt.expected), len(result), tt.expected, result)
			}
		})
	}
}

func TestComplexTemplateStructures(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Complex E-commerce Template",
			tmpl: `Store: {{ store.name }}

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
			context: map[string]any{
				"store": map[string]any{
					"name": "Electronics Emporium",
					"featured_categories": []any{
						map[string]any{
							"name":        "Laptops",
							"active":      true,
							"description": "High-performance computers",
							"items": []any{
								map[string]any{
									"name":     "Ultra Book Pro",
									"price":    1299,
									"stock":    5,
									"discount": 15,
									"rating":   4.8,
								},
								map[string]any{
									"name":     "Budget Laptop",
									"price":    599,
									"stock":    10,
									"discount": 0,
									"rating":   4.2,
								},
							},
						},
						map[string]any{
							"name":        "Smartphones",
							"active":      true,
							"description": "Latest mobile devices",
							"items": []any{
								map[string]any{
									"name":     "Pro Phone X",
									"price":    999,
									"stock":    15,
									"discount": 10,
									"rating":   4.9,
								},
							},
						},
					},
					"special_offers": []any{
						map[string]any{
							"name":      "summer sale",
							"active":    true,
							"discount":  20,
							"ends_soon": true,
						},
						map[string]any{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("output mismatch (want len=%d, got len=%d):\nwant: %q\ngot:  %q",
					len(tt.expected), len(result), tt.expected, result)
			}
		})
	}
}

// Department is a test interface for template context.
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

	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Date filter with structs",
			tmpl: `// Date filter structure testing
Creation time: {{ user.created_at | date:"Y-m-d" }}
Address time: {{ user.address.time | date:"Y-m-d" }}`,
			context: map[string]any{
				"user": User{
					CreatedAt: sixMonthsAgo,
					Address: Address{
						Time: now,
					},
				},
			},
			expected: fmt.Sprintf(`// Date filter structure testing
Creation time: %s
Address time: %s`, sixMonthsAgo.Format("2006-01-02"), now.Format("2006-01-02")),
		},
		{
			name: "Default filter with structs",
			tmpl: `// Default value filter structure testing
City: {{ user.address.city | default:"Unknown city" }}
Name: {{ user.name | default:"Unknown user" }}`,
			context: map[string]any{
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
			tmpl: `// Length filter structure testing
Name length: {{ user.name | length }}
City length: {{ user.address.city | length }}
Tag count: {{ user.tags | size }}
Score count: {{ user.scores | size }}`,
			context: map[string]any{
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
			tmpl: `// Structure condition filter testing
{% if user.name | length > 0 %}User name is not empty{% endif %}
{% if user.address.city | length > 2 %}City name is longer than 2 characters{% endif %}
{% if user.tags | length >= 3 %}At least 3 tags{% endif %}`,
			context: map[string]any{
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
			tmpl: `// Split filter structure testing
Address split: {{ address_str | split:"," }}`,
			context: map[string]any{
				"address_str": "Shanghai,Pudong,Zhangjiang",
			},
			expected: `// Split filter structure testing
Address split: [Shanghai,Pudong,Zhangjiang]`,
		},
		{
			name: "Trim filter with structs",
			tmpl: `// Trim filter structure testing
Trim name: {{ user.name | trim }}
Trim city: {{ user.address.city | trim }}`,
			context: map[string]any{
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
			tmpl: `// For loops with structs and conditions
{% for item in items %}
	{% if item.user.age > 18 %}
		Name: {{ item.user.name }}
		Age: {{ item.user.age }}
		City: {{ item.user.address.city }}
	{% endif %}
{% endfor %}`,
			context: map[string]any{
				"items": []map[string]any{
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
			tmpl: `// Upper and lower filters with structs
Upper name: {{ user.name | upper }}
Lower city: {{ user.address.city | lower }}`,
			context: map[string]any{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("output mismatch (want len=%d, got len=%d):\nwant: %q\ngot:  %q",
					len(tt.expected), len(result), tt.expected, result)
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

	tmpl, err := Compile(sourceTemplate)
	if err != nil {
		t.Fatalf("compiling template: %v", err)
	}

	// Create context and set complex nested data
	ctx := make(map[string]any)
	ctx["simpleMap"] = map[string]string{
		"name": "SimpleMap",
	}
	ctx["arrayMap"] = map[string]any{
		"items": []string{"ArrayMap", "ArrayMap-2"},
	}
	ctx["nestedMap"] = map[string]any{
		"user": map[string]any{
			"name": "NestedMap",
		},
	}
	ctx["stringSliceMap"] = map[string][]string{
		"items": {"StringSliceMap", "StringSliceMap-2"},
	}
	ctx["nestedSliceMap"] = map[string]map[string][]string{
		"config": {
			"values": []string{"NestedSliceMap", "NestedSliceMap-2"},
		},
	}
	ctx["basicSlice"] = []string{"BasicSlice", "BasicSlice-2"}
	ctx["objectSlice"] = []map[string]any{
		{
			"name": "ObjectSlice1",
		},
		{
			"nickname": "ObjectSlice2",
		},
	}
	ctx["nestedObjectArray"] = [][]map[string]any{
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
	}
	ctx["deepStructure"] = [][]map[string]map[string][]map[string][]string{
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
	}

	// Execute template
	output, err := tmpl.Render(ctx)
	if err != nil {
		t.Fatalf("rendering template: %v", err)
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

	tmpl, err := Compile(sourceTemplate)
	if err != nil {
		t.Fatalf("compiling template: %v", err)
	}

	// Create context and set complex nested data
	ctx := make(map[string]any)
	ctx["userArray"] = [][]map[string][]string{
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
	}
	ctx["userProfile"] = map[string][]string{
		"credentials": {"AccountInfo", "SecurityInfo"},
	}
	ctx["studentList"] = []map[string]any{
		{
			"info": "StudentInfo1",
		},
		{
			"info": "StudentInfo2",
		},
	}
	ctx["teacherList"] = []map[string]map[string]any{
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
	}
	ctx["scoreData"] = []map[string]map[string]map[string][]int{
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
	}
	ctx["configData"] = []map[string]map[string]map[string]map[string]int{
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
	}
	ctx["appSettings"] = []map[string]map[string]map[string]map[string][]bool{
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
	}

	// Execute template
	output, err := tmpl.Render(ctx)
	if err != nil {
		t.Fatalf("rendering template: %v", err)
	}
	// Verify output contains all expected values
	expectedValues := []string{
		"Hello, [UserProfile,PersonalInfo]!", "Hello, [AccountInfo,SecurityInfo]!", "Hello, StudentInfo2!",
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

// 	tmpl, err := Compile(sourceTemplate)
// 	if err != nil {
// 		t.Fatalf("compiling template: %v", err)
// 	}

// 	// Create context and set complex nested data
// 	ctx := make(map[string]any)

// 	// Basic data types
// 	ctx["numberList"] = []int{1, 2, 3, 4, 5}
// 	ctx["textContent"] = "HelloWorld"

// 	// Index access
// 	ctx["stringArray"] = []string{"FirstItem", "SecondItem", "ThirdItem"}
// 	ctx["matrixData"] = [][]string{
// 		{"A1", "A2", "A3"},
// 		{"B1", "B2", "B3"},
// 	}
// 	ctx["productList"] = []map[string]any{
// 		{"title": "Product1", "price": 100},
// 		{"title": "Product2", "price": 200},
// 	}

// 	// 2D array
// 	ctx["gridData"] = [][]string{
// 		{"A1", "A2", "A3"},
// 		{"B1", "B2", "B3"},
// 		{"C1", "C2", "C3"},
// 	}

// 	// Object array
// 	ctx["productList"] = []map[string]any{
// 		{"title": "Product1", "price": 100},
// 		{"title": "Product2", "price": 200},
// 		{"title": "Product3", "price": 300},
// 	}

// 	// Nested objects
// 	userSettings := map[string]map[string]any{
// 		"user1": {
// 			"name": "John",
// 			"age":  30,
// 		},
// 		"user2": {
// 			"name": "Jane",
// 			"age":  25,
// 		},
// 	}
// 	ctx["userSettings"] = userSettings

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
// 	ctx["academicRecords"] = academicRecords

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
// 	ctx["schoolData"] = schoolData

// 	// Execute template
// 	output, err := tmpl.Render(ctx)
// 	if err != nil {
// 		t.Fatalf("rendering template: %v", err)
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

	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Basic variable access",
			tmpl: `Hello, {{ name }}!
Simple struct:
User: {{ user }}
Username: {{ user.name }}
User age: {{ user.age }}`,
			context: map[string]any{
				"name": "World",
				"user": User{
					Name: "John",
					Age:  30,
				},
			},
			expected: `Hello, World!
Simple struct:
User: {"name":"John","age":30,"address":{"street":"","city":"","time":"0001-01-01T00:00:00Z","map":{},"slice":[],"mapSlice":{},"sliceMap":[],"department":null}}
Username: John
User age: 30`,
		},
		{
			name: "Nested struct access",
			tmpl: `Nested struct:
User address: {{ user.address }}
Street: {{ user.address.street }}
City: {{ user.address.city }}`,
			context: map[string]any{
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
User address: {"street":"123 Main St","city":"New York","time":"0001-01-01T00:00:00Z","map":{},"slice":[],"mapSlice":{},"sliceMap":[],"department":null}
Street: 123 Main St
City: New York`,
		},
		{
			name: "Time formatting",
			tmpl: `Time test:
Time: {{ user.address.time }}`,
			context: map[string]any{
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
			tmpl: `Map test:
Map: {{ user.address.map }}
Map key1: {{ user.address.map.key1 }}
Map key2: {{ user.address.map.key2 }}`,
			context: map[string]any{
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
			tmpl: `Slice test:
Slice: {{ user.address.slice }}
First element: {{ user.address.slice.0 }}
Second element: {{ user.address.slice.1 }}`,
			context: map[string]any{
				"user": User{
					Address: Address{
						Slice: testSlice,
					},
				},
			},
			expected: `Slice test:
Slice: [item1,item2,item3]
First element: item1
Second element: item2`,
		},
		{
			name: "MapSlice access",
			tmpl: `MapSlice test:
MapSlice: {{ user.address.mapSlice }}
Level 1 key: {{ user.address.mapSlice.level1 }}
Level 2 key: {{ user.address.mapSlice.level1.sublevel }}
Array in level 2: {{ user.address.mapSlice.level1.sublevel.0 }}`,
			context: map[string]any{
				"user": User{
					Address: Address{
						MapSlice: testMapSlice,
					},
				},
			},
			expected: `MapSlice test:
MapSlice: {"level1":{"sublevel":["deepitem1","deepitem2","deepitem3"]}}
Level 1 key: {"sublevel":["deepitem1","deepitem2","deepitem3"]}
Level 2 key: [deepitem1,deepitem2,deepitem3]
Array in level 2: deepitem1`,
		},
		{
			name: "SliceMap access",
			tmpl: `SliceMap test:
SliceMap: {{ user.address.sliceMap }}
First array: {{ user.address.sliceMap.0 }}
First element of first array: {{ user.address.sliceMap.0.0 }}
key1 of first element of first array: {{ user.address.sliceMap.0.0.key1 }}`,
			context: map[string]any{
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
			tmpl: `Team test:
Team name: {{ team.name }}
First member: {{ team.members.0.name }}
Second member city: {{ team.members.1.address.city }}`,
			context: map[string]any{
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
			tmpl: `Department test:
Department street: {{ user.address.department.street }}	
Department city: {{ user.address.department.city }}
Department Map: {{ user.address.department.map }}
Department Slice: {{ user.address.department.slice }}
Department MapSlice: {{ user.address.department.mapSlice }}
Department SliceMap: {{ user.address.department.sliceMap }}`,
			context: map[string]any{
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
Department Slice: [item1,item2,item3]
Department MapSlice: {"level1":{"sublevel":["deepitem1","deepitem2","deepitem3"]}}
Department SliceMap: [[{"key1":"nestedvalue1","key2":"nestedvalue2"},{"key1":"nestedvalue3","key2":"nestedvalue4"}],[{"key1":"nestedvalue5","key2":"nestedvalue6"}]]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			// For tests with JSON output, check key content rather than exact string match
			// due to map key ordering variations in go-json-experiment
			if tt.name == "Interface_field_access" || tt.name == "SliceMap_access" {
				// Verify key components are present
				if !strings.Contains(result, "nestedvalue1") || !strings.Contains(result, "nestedvalue2") ||
					!strings.Contains(result, "nestedvalue3") || !strings.Contains(result, "nestedvalue4") ||
					!strings.Contains(result, "nestedvalue5") || !strings.Contains(result, "nestedvalue6") {
					t.Errorf("Template output missing expected content.\nGot:\n%q", result)
				}
			} else if result != tt.expected {
				t.Errorf("output mismatch (want len=%d, got len=%d):\nwant: %q\ngot:  %q",
					len(tt.expected), len(result), tt.expected, result)
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

	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Basic condition checks",
			tmpl: `// Basic condition tests
{% if user.name %}Username is: {{ user.name }}{% endif %}
{% if user.age > 20 %}User is older than 20{% endif %}`,
			context: map[string]any{
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
			tmpl: `// Array and slice condition tests
{% if user.address.slice.0 %}First slice element: {{ user.address.slice.0 }}{% endif %}
{% if user.address.sliceMap.0.0.key1 %}Value in nested structure: {{ user.address.sliceMap.0.0.key1 }}{% endif %}`,
			context: map[string]any{
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
			tmpl: `// Complex structure condition tests
{% if team.members.0.name %}First team member: {{ team.members.0.name }}{% endif %}
{% if team.members.1.address.city %}Second member's city: {{ team.members.1.address.city }}{% endif %}`,
			context: map[string]any{
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
			tmpl: `// Interface type condition tests
{% if department.department.street %}Department street: {{ department.department.street }}{% endif %}
{% if department.department.city %}Department city: {{ department.department.city }}{% endif %}`,
			context: map[string]any{
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
			tmpl: `// Time type condition tests
{% if user.address.time_str %}Time exists: {{ user.address.time_str }}{% endif %}`,
			context: map[string]any{
				"user": map[string]any{
					"address": map[string]any{
						"time_str": now.Format(time.RFC3339),
					},
				},
			},
			expected: `// Time type condition tests
Time exists: ` + now.Format(time.RFC3339),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("output mismatch (want len=%d, got len=%d):\nwant: %q\ngot:  %q",
					len(tt.expected), len(result), tt.expected, result)
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

	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Basic loop test",
			tmpl: `// Basic loop tests
// Slice loop:
// {% for item in user.address.slice %}
// 		- {{ item }}
// {% endfor %}`,
			context: map[string]any{
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
			tmpl: `// Nested loop tests
// Team members loop:
// {% for member in team.members %}
// 		Member name: {{ member.name }}, age: {{ member.age }}
// 		Address info: {{ member.address.city }} {{ member.address.street }}
// {% endfor %}`,
			context: map[string]any{
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
			tmpl: `// Complex data structure loops
// SliceMap loop:
// {% for arr in user.address.sliceMap %}
// 		Array index:
// 		{% for item in arr %}
// 			Key-value pair: {{ item.key1 }} - {{ item.key2 }}
// 		{% endfor %}
// {% endfor %}`,
			context: map[string]any{
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
			tmpl: `// MapSlice direct access (more reliable than iteration)
MapSlice access:
Level1 data: {{ user.address.mapSlice.level1 }}
Sublevel data: {{ user.address.mapSlice.level1.sublevel }}
First item: {{ user.address.mapSlice.level1.sublevel.0 }}`,
			context: map[string]any{
				"user": User{
					Address: Address{
						MapSlice: testMapSlice,
					},
				},
			},
			expected: `// MapSlice direct access (more reliable than iteration)
MapSlice access:
Level1 data: {"sublevel":["deepitem1","deepitem2","deepitem3"]}
Sublevel data: [deepitem1,deepitem2,deepitem3]
First item: deepitem1`,
		},
		{
			name: "Map key iteration",
			tmpl: `// Map loops (Since key,value iteration is not supported, we need to know keys in advance)
Map loop:
key1: {{ user.address.map.key1 }}
key2: {{ user.address.map.key2 }}`,
			context: map[string]any{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
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

	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Basic pointer type access",
			tmpl: `ID: {{ person.id }}
Name: {{ person.name }}
Age: {{ person.age }}
Status: {{ person.isActive }}`,
			context: map[string]any{
				"person": person,
			},
			expected: `ID: EMP001
Name: John Doe
Age: 30
Status: true`,
		},
		{
			name: "Nested pointer structure access",
			tmpl: `Contact Information:
Phone: {% if person.contact.phone %}{{ person.contact.phone }}{% else %}Not set{% endif %}
Email: {% if person.contact.email %}{{ person.contact.email }}{% else %}Not set{% endif %}
Fax: {% if person.contact.fax %}{{ person.contact.fax }}{% else %}Not set{% endif %}`,
			context: map[string]any{
				"person": person,
			},
			expected: `Contact Information:
Phone: 123-456-7890
Email: john.doe@company.com
Fax: Not set`,
		},
		{
			name: "Pointer slice access",
			tmpl: `Tags: {{ person.tags }}
First tag: {% if person.tags.0 %}{{ person.tags.0 }}{% else %}Empty{% endif %}
Second tag: {% if person.tags.1 %}{{ person.tags.1 }}{% else %}Empty{% endif %}
Third tag: {% if person.tags.2 %}{{ person.tags.2 }}{% else %}Empty{% endif %}`,
			context: map[string]any{
				"person": person,
			},
			expected: `Tags: ["Go Programming","Python Development",null]
First tag: Go Programming
Second tag: Python Development
Third tag: Empty`,
		},
		{
			name: "Fixed size pointer array access",
			tmpl: `Projects: {{ person.projects }}
Project 1: {% if person.projects.0 %}{{ person.projects.0 }}{% else %}None{% endif %}
Project 2: {% if person.projects.1 %}{{ person.projects.1 }}{% else %}None{% endif %}
Project 3: {% if person.projects.2 %}{{ person.projects.2 }}{% else %}None{% endif %}`,
			context: map[string]any{
				"person": person,
			},
			expected: `Projects: ["Project Alpha","Project Beta",null]
Project 1: Project Alpha
Project 2: Project Beta
Project 3: None`,
		},
		{
			name: "Map with pointer values access",
			tmpl: `Metadata:
Department: {% if person.metaData.department %}{{ person.metaData.department }}{% else %}Not set{% endif %}
Empty field: {% if person.metaData.empty %}{{ person.metaData.empty }}{% else %}Not set{% endif %}`,
			context: map[string]any{
				"person": person,
			},
			expected: `Metadata:
Department: Engineering Department
Empty field: Not set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
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

	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Basic pointer condition checks",
			tmpl: `{% if user.name %}Name: {{ user.name }}{% else %}Name: Not set{% endif %}
{% if user.age %}Age: {{ user.age }}{% else %}Age: Unknown{% endif %}
{% if user.isActive %}Status: Active{% else %}Status: Inactive{% endif %}`,
			context: map[string]any{
				"user": user,
			},
			expected: `Name: Test User
Age: 25
Status: Active`,
		},
		{
			name: "Nested structure condition checks",
			tmpl: `{% if user.address %}
Address information exists:
{% if user.address.street %}  Street: {{ user.address.street }}{% endif %}
{% if user.address.city %}  City: {{ user.address.city }}{% endif %}
{% if user.address.country %}  Country: {{ user.address.country }}{% endif %}
{% else %}
Address information does not exist
{% endif %}`,
			context: map[string]any{
				"user": user,
			},
			expected: `
Address information exists:
  Street: 123 Main Street
  City: New York
  Country: USA
`,
		},
		{
			name: "Pointer slice condition checks",
			tmpl: `{% if user.skills %}
Skills list:
{% if user.skills.0 %}  - {{ user.skills.0 }}{% endif %}
{% if user.skills.1 %}  - {{ user.skills.1 }}{% endif %}
{% if user.skills.2 %}  - {{ user.skills.2 }}{% else %}  - Third skill is empty{% endif %}
{% else %}
No skills recorded
{% endif %}`,
			context: map[string]any{
				"user": user,
			},
			expected: `
Skills list:
  - Go
  - Python
  - Third skill is empty
`,
		},
		{
			name: "Complex condition checks",
			tmpl: `{% if user.name && user.age && user.isActive %}
Complete user info: {{ user.name }} ({{ user.age }} years old)
{% else %}
User information incomplete
{% endif %}

{% if user.address && user.address.city %}
User is from: {{ user.address.city }}
{% else %}
City information unknown
{% endif %}`,
			context: map[string]any{
				"user": user,
			},
			expected: `
Complete user info: Test User (25 years old)



User is from: New York
`,
		},
		{
			name: "Nil value condition tests",
			tmpl: `{% if emptyUser.name %}Name: {{ emptyUser.name }}{% else %}Name: Not set{% endif %}
{% if emptyUser.address %}Has address info{% else %}No address info{% endif %}
{% if emptyUser.skills %}Skill count: {{ emptyUser.skills | size }}{% else %}No skills recorded{% endif %}`,
			context: map[string]any{
				"emptyUser": emptyUser,
			},
			expected: `Name: Not set
No address info
No skills recorded`,
		},
		{
			name: "Array length condition checks",
			tmpl: `{% if user.skills | size > 1 %}
Multi-skill user: {{ user.skills | size }} skills
{% else %}
Single-skill or no-skill user
{% endif %}

{% if user.scores | size >= 2 %}
Sufficient scores: has {{ user.scores | size }} scores
{% else %}
Insufficient scores
{% endif %}`,
			context: map[string]any{
				"user": user,
			},
			expected: `
Multi-skill user: 3 skills



Sufficient scores: has 2 scores
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
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

	tests := []struct {
		name                string
		tmpl                string
		context             map[string]any
		containsAll         []string // must contain all these strings
		containsNone        []string // must not contain these strings
		exactMatch          string   // exact match (for non-map iteration cases)
		useContainsChecking bool     // whether to use contains checking instead of exact match
	}{
		{
			name: "Pointer slice loops",
			tmpl: `Team Members:
{% for team in company.teams %}
Team: {% if team.name %}{{ team.name }}{% endif %}
Lead: {% if team.lead %}{{ team.lead }}{% endif %}
Members:
{% for member in team.members %}
  - {% if member %}{{ member }}{% else %}Vacant{% endif %}
{% endfor %}
---
{% endfor %}`,
			context: map[string]any{
				"company": company,
			},
			exactMatch: `Team Members:

Team: Development Team
Lead: John Smith
Members:

  - Alice Johnson

  - Bob Wilson

  - Vacant

---

Team: QA Team
Lead: Jane Doe
Members:

  - Carol Brown

---
`,
			useContainsChecking: false,
		},
		{
			name: "Pointer struct slice loops",
			tmpl: `Product List:
{% for product in company.products %}
Product: {% if product.name %}{{ product.name }}{% endif %} - Price: {% if product.price %}${{ product.price }}{% endif %}
{% endfor %}`,
			context: map[string]any{
				"company": company,
			},
			exactMatch: `Product List:

Product: Product Alpha - Price: $100

Product: Product Beta - Price: $200
`,
			useContainsChecking: false,
		},
		{
			name: "Map iteration test (contains checking)",
			tmpl: `Company Settings:
{% for key, value in company.settings %}
{{ key }}: {% if value %}{{ value }}{% else %}Not set{% endif %}
{% endfor %}`,
			context: map[string]any{
				"company": company,
			},
			containsAll: []string{
				"theme: Dark Theme",
				"locale: English",
				"empty: Not set",
				"Company Settings:",
			},
			containsNone:        []string{},
			useContainsChecking: true,
		},
		{
			name: "Nested loop test",
			tmpl: `Detailed Team Information:
{% for team in company.teams %}
=== {% if team.name %}{{ team.name }}{% endif %} ===
{% for member in team.members %}
{% if member %}Member: {{ member }}{% else %}Vacant{% endif %}
{% endfor %}
{% endfor %}`,
			context: map[string]any{
				"company": company,
			},
			exactMatch: `Detailed Team Information:

=== Development Team ===

Member: Alice Johnson

Member: Bob Wilson

Vacant


=== QA Team ===

Member: Carol Brown

`,
			useContainsChecking: false,
		},
		{
			name: "Map with pointer values loop test",
			tmpl: `Budget Allocation:
{% for dept, budget in company.budgets %}
{{ dept }}: {% if budget %}${{ budget }}{% else %}Not allocated{% endif %}
{% endfor %}`,
			context: map[string]any{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if tt.useContainsChecking {
				// Use contains checking to handle map iteration order uncertainty
				for _, required := range tt.containsAll {
					if !strings.Contains(result, required) {
						t.Errorf("Output does not contain required string: %s\nGot:\n%s", required, result)
					}
				}
				for _, forbidden := range tt.containsNone {
					if strings.Contains(result, forbidden) {
						t.Errorf("Output contains forbidden string: %s\nGot:\n%s", forbidden, result)
					}
				}
			} else if result != tt.exactMatch {
				// Exact match
				t.Errorf("got %q, want %q", result, tt.exactMatch)
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
	allProjects := make([]*Project, 0, len(company.Projects["active"])+len(company.Projects["completed"]))
	allProjects = append(allProjects, company.Projects["active"]...)
	allProjects = append(allProjects, company.Projects["completed"]...)
	company.Departments[0].Projects = allProjects

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
	ctx := make(map[string]any)
	ctx["company"] = company

	// Test cases - mixed exact match and contains checks
	tests := []struct {
		name         string
		tmpl         string
		expected     string
		containsAll  []string // Must contain all these strings
		containsNone []string // Must not contain these strings
		useContains  bool     // Whether to use contains checking instead of exact match
	}{
		{
			name: "Basic company information",
			tmpl: `Company name: {{ company.name }}
Founded year: {{ company.founded }}
Is public: {{ company.is_public }}
Annual revenue: ${{ company.revenue }}`,
			expected: `Company name: TechCorp
Founded year: 2010
Is public: true
Annual revenue: $1000000.5`,
			useContains: false,
		},
		{
			name: "Test if conditions with pointer fields",
			tmpl: `{% if company.is_public %}{{ company.name }} is a public company{% endif %}
{% if company.revenue > 500000 %}Company revenue exceeds 500,000 USD{% endif %}
{% if company.founded < 2015 %}Company was founded early{% endif %}`,
			expected: `TechCorp is a public company
Company revenue exceeds 500,000 USD
Company was founded early`,
			useContains: false,
		},
		{
			name: "for loop - single variable iteration over employees",
			tmpl: `Employee list:
{% for employee in company.employees %}
- {{ employee.name }} ({{ employee.age }} years, ${{ employee.salary }})
{% endfor %}`,
			expected: `Employee list:

- Alice Johnson (30 years, $120000)

- Bob Smith (28 years, $95000)
`,
			useContains: false,
		},
		{
			name: "for loop - double variable iteration over company settings (using contains check)",
			tmpl: `Company settings:
{% for key, value in company.settings %}
{{ key }}: {{ value }}
{% endfor %}`,
			containsAll: []string{
				"Company settings:",
				"timezone: UTC",
				"currency: USD",
				"work_mode: hybrid",
				"dress_code: casual",
			},
			useContains: true,
		},
		{
			name: "for loop - double variable iteration over locations (using contains check)",
			tmpl: `Office locations:
{% for location_name, location in company.locations %}
{{ location_name }}: {{ location.city }}, {{ location.country }}
Address: {{ location.street }}
{% endfor %}`,
			containsAll: []string{
				"Office locations:",
				"hq: San Francisco, USA",
				"Address: 123 Tech Street",
				"branch: Austin, USA",
				"Address: 456 Innovation Ave",
			},
			useContains: true,
		},
		{
			name: "nested for loop - employee skills",
			tmpl: `Employee skill details:
{% for employee in company.employees %}
{{ employee.name }}'s skills:
{% for skill in employee.skills %}
  - {{ skill.name }}: Level {{ skill.level }}, {{ skill.years_exp }} years experience, Certified: {{ skill.certified }}
{% endfor %}
{% endfor %}`,
			expected: `Employee skill details:

Alice Johnson's skills:

  - Python: Level 9, 5.5 years experience, Certified: true

  - Machine Learning: Level 8, 3 years experience, Certified: true


Bob Smith's skills:

  - JavaScript: Level 8, 4 years experience, Certified: false

  - React: Level 7, 2.5 years experience, Certified: true

`,
			useContains: false,
		},
		{
			name: "complex nested access - employee contact information (using contains check)",
			tmpl: `Employee contact information:
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
				"Alice Johnson:",
				"Email: alice@techcorp.com",
				"Phone: +1-555-0101",
				"Social media:",
				"linkedin: alice-johnson",
				"github: alicejohnson",
				"twitter: @alice_codes",
				"Bob Smith:",
				"Email: bob@techcorp.com",
				"Phone: +1-555-0201",
				"linkedin: bob-smith",
				"github: bobsmith",
			},
			useContains: true,
		},
		{
			name: "fixed size array access",
			tmpl: `Employee ratings:
{% for employee in company.employees %}
{{ employee.name }} ratings: [{{ employee.scores.0 }}, {{ employee.scores.1 }}, {{ employee.scores.2 }}]
{% endfor %}`,
			expected: `Employee ratings:

Alice Johnson ratings: [95, 87, 92]

Bob Smith ratings: [82, 90, 85]
`,
			useContains: false,
		},
		{
			name: "map value as complex structure (using contains check)",
			tmpl: `Project status report:
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
				"- AI Platform (Budget: $250000)",
				"Team members: emp-001 emp-002",
				"- Mobile App (Budget: $150000)",
				"Team members: emp-003 emp-001",
				"completed project:",
				"- Website Redesign (Budget: $75000)",
				"Team members: emp-002",
			},
			useContains: true,
		},
		{
			name: "complex if condition nested access",
			tmpl: `Employee evaluation:
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


High-salary employee Alice Johnson:
  - Living city: San Francisco
  - Department: Engineering
  
   Core skills:
  
    
    - Python (Expert Level 9)
    
  
    
    - Machine Learning (Expert Level 8)
    
  
  


`,
			useContains: false,
		},
		{
			name: "complex nested array and coordinate access (using contains check)",
			tmpl: `Office location coordinates:
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
				"hq (San Francisco):",
				"Coordinates: [37.7749, -122.4194]",
				"Metadata:",
				"building_type: office",
				"floor_count: 10",
				"parking: available",
				"branch (Austin):",
				"Coordinates: [30.2672, -97.7431]",
				"building_type: co-working",
				"floor_count: 5",
				"parking: limited",
			},
			useContains: true,
		},
		{
			name: "employee achievements record (using contains check)",
			tmpl: `Employee achievements:
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
				"Alice Johnson:",
				"2023 year:",
				"- best_performer (Priority: 1)",
				"- innovation_award (Priority: 2)",
				"2024 year:",
				"- team_lead (Priority: 1)",
				"- mentor (Priority: 3)",
				"Bob Smith:",
				"- fast_learner (Priority: 1)",
				"- team_player (Priority: 2)",
			},
			useContains: true,
		},
		{
			name: "map value as pointer struct - department teams",
			tmpl: `Department team information:
{% for dept in company.departments %}
Department: {{ dept.name }}
Team:
{% for team_name, team in dept.teams %}
  {{ team_name }}: {{ team.name }} (Lead: {{ team.lead }}, Budget: ${{ team.budget }})
{% endfor %}
{% endfor %}`,
			containsAll: []string{
				"Department team information:",
				"Department: Engineering",
				"Team:",
				"development: Development Team (Lead: emp-001, Budget: $300000)",
				"qa: QA Team (Lead: emp-004, Budget: $150000)",
			},
			useContains: true,
		},
		{
			name: "deeply nested - team certification information",
			tmpl: `Team certification details:
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
				"Development Team certification:",
				"- Java Certified Developer (Oracle, Valid until: 2025-12-31)",
				"Tags: java backend",
				"- AWS Solutions Architect (Amazon, Valid until: 2026-06-15)",
				"Tags: aws cloud",
				"QA Team certification:",
				"- ISTQB Foundation (ISTQB, Valid until: 2027-01-01)",
				"Tags: testing quality",
			},
			useContains: true,
		},
		{
			name: "multi-level nested - department subsidiary addresses",
			tmpl: `Department subsidiary information:
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
				"Engineering subsidiary:",
				"lab: Palo Alto, 789 Research Blvd",
				"Coordinates: [37.4419, -122.143]",
				"type: research_lab",
				"security: high",
				"datacenter: Denver, 321 Server Farm Rd",
				"Coordinates: [39.7392, -104.9903]",
				"type: datacenter",
				"power_backup: redundant",
			},
			useContains: true,
		},
		{
			name: "extremely deeply nested - organizational settings",
			tmpl: `Organizational settings details:
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
				"North America Division:",
				"hr configuration:",
				"vacation_days: 25",
				"sick_days: 10",
				"it configuration:",
				"laptop_budget: 2000",
				"software_budget: 500",
			},
			useContains: true,
		},
		{
			name: "three-level nested map - organizational teams",
			tmpl: `Organizational team structure:
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
				"North America Division:",
				"engineering department:",
				"dev: Development Team (3 members)",
				"qa: QA Team (2 members)",
			},
			useContains: true,
		},
	}

	// Execute tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, ctx)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}

			if tt.useContains {
				// Use contains checking mode, suitable for map iteration order uncertainty
				for _, required := range tt.containsAll {
					if !strings.Contains(result, required) {
						t.Errorf("Missing required content: %s\nActual output:\n%s", required, result)
					}
				}
				for _, forbidden := range tt.containsNone {
					if strings.Contains(result, forbidden) {
						t.Errorf("Output contains forbidden content: %s\nActual output:\n%s", forbidden, result)
					}
				}
			} else if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				// Exact match
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestAliasTypes tests the handling of alias types in templates
func TestAliasTypes(t *testing.T) {
	// Define alias types
	type UserID string
	type Age int
	type Score float64
	type IsActive bool
	type Department string
	type Priority int

	// Define pointer alias types
	type UserIDPtr *string
	type AgePtr *int
	type ScorePtr *float64
	type IsActivePtr *bool
	type DepartmentPtr *string

	// Define struct using alias types
	type User struct {
		ID         UserID     `json:"id"`
		Name       string     `json:"name"`
		Age        Age        `json:"age"`
		Score      Score      `json:"score"`
		IsActive   IsActive   `json:"is_active"`
		Department Department `json:"department"`
		Priority   Priority   `json:"priority"`
	}

	// Define struct using pointer alias types
	type UserWithPointers struct {
		ID         UserIDPtr     `json:"id"`
		Name       *string       `json:"name"`
		Age        AgePtr        `json:"age"`
		Score      ScorePtr      `json:"score"`
		IsActive   IsActivePtr   `json:"is_active"`
		Department DepartmentPtr `json:"department"`
		Manager    *User         `json:"manager"`
	}

	// Define slice of alias types
	type TagList []string
	type ScoreList []Score
	type IDList []UserID
	type StatusList []IsActive

	// Create test data with alias types
	users := []User{
		{
			ID:         UserID("user-001"),
			Name:       "Alice Johnson",
			Age:        Age(30),
			Score:      Score(95.5),
			IsActive:   IsActive(true),
			Department: Department("Engineering"),
			Priority:   Priority(1),
		},
		{
			ID:         UserID("user-002"),
			Name:       "Bob Smith",
			Age:        Age(25),
			Score:      Score(87.2),
			IsActive:   IsActive(false),
			Department: Department("Marketing"),
			Priority:   Priority(2),
		},
		{
			ID:         UserID("user-003"),
			Name:       "Carol Brown",
			Age:        Age(35),
			Score:      Score(92.8),
			IsActive:   IsActive(true),
			Department: Department("HR"),
			Priority:   Priority(3),
		},
	}

	// Test data with alias slices
	tags := TagList{"important", "urgent", "high-priority"}
	scores := ScoreList{Score(88.5), Score(92.0), Score(85.7)}
	userIDs := IDList{UserID("admin-001"), UserID("admin-002")}
	statusList := StatusList{IsActive(true), IsActive(false), IsActive(true)}

	// Helper function to create string pointers
	stringPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }
	float64Ptr := func(f float64) *float64 { return &f }
	boolPtr := func(b bool) *bool { return &b }

	// Test data with pointer aliases
	usersWithPointers := []UserWithPointers{
		{
			ID:         UserIDPtr(stringPtr("ptr-user-001")),
			Name:       stringPtr("Alice Pointer"),
			Age:        AgePtr(intPtr(28)),
			Score:      ScorePtr(float64Ptr(93.5)),
			IsActive:   IsActivePtr(boolPtr(true)),
			Department: DepartmentPtr(stringPtr("Engineering")),
			Manager:    &users[0],
		},
		{
			ID:         UserIDPtr(stringPtr("ptr-user-002")),
			Name:       stringPtr("Bob Pointer"),
			Age:        AgePtr(intPtr(32)),
			Score:      ScorePtr(float64Ptr(89.2)),
			IsActive:   IsActivePtr(boolPtr(false)),
			Department: DepartmentPtr(stringPtr("Marketing")),
			Manager:    &users[1],
		},
		{
			ID:         nil, // Test nil pointer
			Name:       stringPtr("Carol Pointer"),
			Age:        nil, // Test nil pointer
			Score:      ScorePtr(float64Ptr(91.8)),
			IsActive:   IsActivePtr(boolPtr(true)),
			Department: nil, // Test nil pointer
			Manager:    &users[2],
		},
	}

	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Basic alias type rendering",
			tmpl: `User Information:
ID: {{ user.id }}
Name: {{ user.name }}
Age: {{ user.age }}
Score: {{ user.score }}
Active: {{ user.is_active }}
Department: {{ user.department }}
Priority: {{ user.priority }}`,
			context: map[string]any{
				"user": users[0],
			},
			expected: `User Information:
ID: user-001
Name: Alice Johnson
Age: 30
Score: 95.5
Active: true
Department: Engineering
Priority: 1`,
		},
		{
			name: "Alias types in if conditions",
			tmpl: `User Status Report:
{% if user.is_active %}{{ user.name }} is active{% else %}{{ user.name }} is inactive{% endif %}
{% if user.age >= 30 %}{{ user.name }} is senior ({{ user.age }} years old){% else %}{{ user.name }} is junior ({{ user.age }} years old){% endif %}
{% if user.score > 90 %}{{ user.name }} has excellent score: {{ user.score }}{% endif %}
{% if user.department == "Engineering" %}{{ user.name }} works in {{ user.department }}{% endif %}
{% if user.priority <= 2 %}{{ user.name }} has high priority: {{ user.priority }}{% endif %}`,
			context: map[string]any{
				"user": users[0],
			},
			expected: `User Status Report:
Alice Johnson is active
Alice Johnson is senior (30 years old)
Alice Johnson has excellent score: 95.5
Alice Johnson works in Engineering
Alice Johnson has high priority: 1`,
		},
		{
			name: "Alias types in for loops",
			tmpl: `All Users:
{% for user in users %}
- {{ user.name }} ({{ user.id }}): Age {{ user.age }}, Score {{ user.score }}, Department {{ user.department }}
  {% if user.is_active %}Status: Active{% else %}Status: Inactive{% endif %}
{% endfor %}`,
			context: map[string]any{
				"users": users,
			},
			expected: `All Users:

- Alice Johnson (user-001): Age 30, Score 95.5, Department Engineering
  Status: Active

- Bob Smith (user-002): Age 25, Score 87.2, Department Marketing
  Status: Inactive

- Carol Brown (user-003): Age 35, Score 92.8, Department HR
  Status: Active
`,
		},
		{
			name: "Alias slice types in for loops",
			tmpl: `Tags:
{% for tag in tags %}
- {{ tag }}
{% endfor %}

Scores:
{% for score in scores %}
- {{ score }}
{% endfor %}

User IDs:
{% for id in userIDs %}
- {{ id }}
{% endfor %}

Status List:
{% for status in statusList %}
- {{ status }}
{% endfor %}`,
			context: map[string]any{
				"tags":       tags,
				"scores":     scores,
				"userIDs":    userIDs,
				"statusList": statusList,
			},
			expected: `Tags:

- important

- urgent

- high-priority


Scores:

- 88.5

- 92

- 85.7


User IDs:

- admin-001

- admin-002


Status List:

- true

- false

- true
`,
		},
		{
			name: "Complex conditions with alias types",
			tmpl: `Active High-Performing Users:
{% for user in users %}
{% if user.is_active && user.score > 90 && user.age >= 25 %}

★ {{ user.name }}
  ID: {{ user.id }}
  Age: {{ user.age }} years
  Score: {{ user.score }}
  Department: {{ user.department }}
  Priority: {{ user.priority }}
{% endif %}
{% endfor %}`,
			context: map[string]any{
				"users": users,
			},
			expected: `Active High-Performing Users:



★ Alice Johnson
  ID: user-001
  Age: 30 years
  Score: 95.5
  Department: Engineering
  Priority: 1






★ Carol Brown
  ID: user-003
  Age: 35 years
  Score: 92.8
  Department: HR
  Priority: 3

`,
		},
		{
			name: "Alias types with filters",
			tmpl: `User Summary:
{% for user in users %}
{{ user.name | upper }}:
  ID: {{ user.id | upper }}
  Department: {{ user.department | lower }}
  {% if user.score >= 85 %}Long score: {{ user.score }}{% endif %}
{% endfor %}`,
			context: map[string]any{
				"users": users,
			},
			expected: `User Summary:

ALICE JOHNSON:
  ID: USER-001
  Department: engineering
  Long score: 95.5

BOB SMITH:
  ID: USER-002
  Department: marketing
  Long score: 87.2

CAROL BROWN:
  ID: USER-003
  Department: hr
  Long score: 92.8
`,
		},
		{
			name: "Alias types comparison and arithmetic",
			tmpl: `User Comparison:
{% for user in users %}
{{ user.name }}:
  {% if user.age > 28 %}Age category: Senior{% else %}Age category: Junior{% endif %}
  {% if user.score >= 90 %}Performance: Excellent{% else %}Performance: Good{% endif %}
  {% if user.priority == 1 %}Priority: Highest{% else %}Priority: {{ user.priority }}{% endif %}
{% endfor %}`,
			context: map[string]any{
				"users": users,
			},
			expected: `User Comparison:

Alice Johnson:
  Age category: Senior
  Performance: Excellent
  Priority: Highest

Bob Smith:
  Age category: Junior
  Performance: Good
  Priority: 2

Carol Brown:
  Age category: Senior
  Performance: Excellent
  Priority: 3
`,
		},
		{
			name: "Mixed alias and regular types",
			tmpl: `Department Statistics:
{% for user in users %}
{{ user.department }} Department:
  Employee: {{ user.name }} (ID: {{ user.id }})
  Details: {{ user.age }} years old, Score: {{ user.score }}
  Status: {% if user.is_active %}Active{% else %}Inactive{% endif %}
  Priority Level: {{ user.priority }}
{% endfor %}`,
			context: map[string]any{
				"users": users,
			},
			expected: `Department Statistics:

Engineering Department:
  Employee: Alice Johnson (ID: user-001)
  Details: 30 years old, Score: 95.5
  Status: Active
  Priority Level: 1

Marketing Department:
  Employee: Bob Smith (ID: user-002)
  Details: 25 years old, Score: 87.2
  Status: Inactive
  Priority Level: 2

HR Department:
  Employee: Carol Brown (ID: user-003)
  Details: 35 years old, Score: 92.8
  Status: Active
  Priority Level: 3
`,
		},
		{
			name: "Pointer alias types basic rendering",
			tmpl: `User with Pointers:
ID: {{ user.id }}
Name: {{ user.name }}
Age: {{ user.age }}
Score: {{ user.score }}
Active: {{ user.is_active }}
Department: {{ user.department }}
Manager: {{ user.manager.name }}`,
			context: map[string]any{
				"user": usersWithPointers[0],
			},
			expected: `User with Pointers:
ID: ptr-user-001
Name: Alice Pointer
Age: 28
Score: 93.5
Active: true
Department: Engineering
Manager: Alice Johnson`,
		},
		{
			name: "Pointer alias types with nil values",
			tmpl: `User with Nil Pointers:
ID: {% if user.id %}{{ user.id }}{% else %}N/A{% endif %}
Name: {{ user.name }}
Age: {% if user.age %}{{ user.age }}{% else %}N/A{% endif %}
Score: {{ user.score }}
Active: {{ user.is_active }}
Department: {% if user.department %}{{ user.department }}{% else %}N/A{% endif %}
Manager: {{ user.manager.name }}`,
			context: map[string]any{
				"user": usersWithPointers[2],
			},
			expected: `User with Nil Pointers:
ID: N/A
Name: Carol Pointer
Age: N/A
Score: 91.8
Active: true
Department: N/A
Manager: Carol Brown`,
		},
		{
			name: "Pointer alias types in for loops",
			tmpl: `All Users with Pointers:
{% for user in users %}
- {{ user.name }}{% if user.id %} ({{ user.id }}){% endif %}
  Age: {% if user.age %}{{ user.age }}{% else %}N/A{% endif %}
  Score: {{ user.score }}
  Department: {% if user.department %}{{ user.department }}{% else %}N/A{% endif %}
  Manager: {{ user.manager.name }}
{% endfor %}`,
			context: map[string]any{
				"users": usersWithPointers,
			},
			expected: `All Users with Pointers:

- Alice Pointer (ptr-user-001)
  Age: 28
  Score: 93.5
  Department: Engineering
  Manager: Alice Johnson

- Bob Pointer (ptr-user-002)
  Age: 32
  Score: 89.2
  Department: Marketing
  Manager: Bob Smith

- Carol Pointer
  Age: N/A
  Score: 91.8
  Department: N/A
  Manager: Carol Brown
`,
		},
		{
			name: "Pointer alias types in conditions",
			tmpl: `Pointer Alias Conditions:
{% for user in users %}
{{ user.name }}:
  {% if user.id %}Has ID: {{ user.id }}{% else %}No ID{% endif %}
  {% if user.age %}Age: {{ user.age }}{% else %}Age: Unknown{% endif %}
  {% if user.department %}Department: {{ user.department }}{% else %}Department: Unknown{% endif %}
  {% if user.is_active %}Status: Active{% else %}Status: Inactive{% endif %}
{% endfor %}`,
			context: map[string]any{
				"users": usersWithPointers,
			},
			expected: `Pointer Alias Conditions:

Alice Pointer:
  Has ID: ptr-user-001
  Age: 28
  Department: Engineering
  Status: Active

Bob Pointer:
  Has ID: ptr-user-002
  Age: 32
  Department: Marketing
  Status: Inactive

Carol Pointer:
  No ID
  Age: Unknown
  Department: Unknown
  Status: Active
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBasicBreakContinue(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  Context
		expected string
	}{
		{
			name:     "Basic break in for loop",
			tmpl:     `{% for i in nums %}{% if i == 3 %}{% break %}{% endif %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "12",
		},
		{
			name:     "Basic continue in for loop",
			tmpl:     `{% for i in nums %}{% if i == 3 %}{% continue %}{% endif %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "1245",
		},
		{
			name:     "Break at beginning",
			tmpl:     `{% for i in nums %}{% break %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "",
		},
		{
			name:     "Continue all iterations",
			tmpl:     `{% for i in nums %}{% continue %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "",
		},
		{
			name:     "Multiple breaks (only first one executes)",
			tmpl:     `{% for i in nums %}{% if i == 2 %}{% break %}{% endif %}{{ i }}{% if i == 4 %}{% break %}{% endif %}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "1",
		},
		{
			name:     "Break in string iteration",
			tmpl:     `{% for char in str %}{% if char == 'c' %}{% break %}{% endif %}{{ char }}{% endfor %}`,
			context:  Context{"str": "abcdef"},
			expected: "ab",
		},
		{
			name:     "Continue in string iteration",
			tmpl:     `{% for char in str %}{% if char == 'c' %}{% continue %}{% endif %}{{ char }}{% endfor %}`,
			context:  Context{"str": "abcdef"},
			expected: "abdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNestedBreakContinue(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  Context
		expected string
	}{
		{
			name:     "Break in inner loop only",
			tmpl:     `{% for i in nums %}{% for j in nums %}{% if j == 2 %}{% break %}{% endif %}{{ i }}-{{ j }}|{% endfor %}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3}},
			expected: "1-1|2-1|3-1|",
		},
		{
			name:     "Continue in inner loop only",
			tmpl:     `{% for i in nums %}{% for j in nums %}{% if j == 2 %}{% continue %}{% endif %}{{ i }}-{{ j }}|{% endfor %}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3}},
			expected: "1-1|1-3|2-1|2-3|3-1|3-3|",
		},
		{
			name:     "Break in outer loop",
			tmpl:     `{% for i in nums %}{% if i == 2 %}{% break %}{% endif %}{% for j in nums %}{{ i }}-{{ j }}|{% endfor %}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3}},
			expected: "1-1|1-2|1-3|",
		},
		{
			name:     "Continue in outer loop",
			tmpl:     `{% for i in nums %}{% if i == 2 %}{% continue %}{% endif %}{% for j in nums %}{{ i }}-{{ j }}|{% endfor %}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3}},
			expected: "1-1|1-2|1-3|3-1|3-2|3-3|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBreakContinueErrors(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        string
		context     Context
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Break outside of loop",
			tmpl:        `Hello {% break %} World`,
			context:     Context{},
			expectError: true,
			errorMsg:    "break statement outside of loop",
		},
		{
			name:        "Continue outside of loop",
			tmpl:        `Hello {% continue %} World`,
			context:     Context{},
			expectError: true,
			errorMsg:    "continue statement outside of loop",
		},
		{
			name:        "Break in if condition outside loop",
			tmpl:        `{% if true %}{% break %}{% endif %}`,
			context:     Context{},
			expectError: true,
			errorMsg:    "break statement outside of loop",
		},
		{
			name:        "Continue in if condition outside loop",
			tmpl:        `{% if true %}{% continue %}{% endif %}`,
			context:     Context{},
			expectError: true,
			errorMsg:    "continue statement outside of loop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil error and result %q", tt.errorMsg, result)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestAdvancedBreakContinue(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  Context
		expected string
	}{
		{
			name:     "Break and continue in same loop",
			tmpl:     `{% for i in nums %}{% if i == 2 %}{% continue %}{% endif %}{% if i == 4 %}{% break %}{% endif %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "13",
		},
		{
			name:     "Break with complex condition",
			tmpl:     `{% for i in nums %}{% if i > 2 && i < 5 %}{% break %}{% endif %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "12",
		},
		{
			name:     "Continue with complex condition",
			tmpl:     `{% for i in nums %}{% if i > 1 && i < 4 %}{% continue %}{% endif %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "145",
		},
		{
			name:     "Break in empty array",
			tmpl:     `{% for i in empty %}{% break %}{{ i }}{% endfor %}Done`,
			context:  Context{"empty": []int{}},
			expected: "Done",
		},
		{
			name:     "Continue in single element",
			tmpl:     `{% for i in single %}{% continue %}{{ i }}{% endfor %}Done`,
			context:  Context{"single": []int{1}},
			expected: "Done",
		},
		{
			name:     "Break with filter",
			tmpl:     `{% for item in items %}{% if item | length > 3 %}{% break %}{% endif %}{{ item }}|{% endfor %}`,
			context:  Context{"items": []string{"a", "bb", "ccc", "dddd", "eeeee"}},
			expected: "a|bb|ccc|",
		},
		{
			name:     "Continue with filter",
			tmpl:     `{% for item in items %}{% if item | length == 2 %}{% continue %}{% endif %}{{ item }}|{% endfor %}`,
			context:  Context{"items": []string{"a", "bb", "ccc", "dd", "e"}},
			expected: "a|ccc|e|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDeepNestedBreakContinue(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  Context
		expected string
	}{
		{
			name:     "Three level nested with break",
			tmpl:     `{% for i in nums %}{% for j in nums %}{% for k in nums %}{% if k == 2 %}{% break %}{% endif %}{{ i }}-{{ j }}-{{ k }}|{% endfor %}{% endfor %}{% endfor %}`,
			context:  Context{"nums": []int{1, 2}},
			expected: "1-1-1|1-2-1|2-1-1|2-2-1|",
		},
		{
			name:     "Three level nested with continue",
			tmpl:     `{% for i in nums %}{% for j in nums %}{% for k in nums %}{% if k == 2 %}{% continue %}{% endif %}{{ i }}-{{ j }}-{{ k }}|{% endfor %}{% endfor %}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3}},
			expected: "1-1-1|1-1-3|1-2-1|1-2-3|1-3-1|1-3-3|2-1-1|2-1-3|2-2-1|2-2-3|2-3-1|2-3-3|3-1-1|3-1-3|3-2-1|3-2-3|3-3-1|3-3-3|",
		},
		{
			name:     "Middle level break affects only that level",
			tmpl:     `{% for i in nums %}{% for j in nums %}{% if j == 2 %}{% break %}{% endif %}{% for k in nums %}{{ i }}-{{ j }}-{{ k }}|{% endfor %}{% endfor %}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3}},
			expected: "1-1-1|1-1-2|1-1-3|2-1-1|2-1-2|2-1-3|3-1-1|3-1-2|3-1-3|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestComplexControlFlowScenarios(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  Context
		expected string
	}{
		{
			name:     "Break in nested if conditions",
			tmpl:     `{% for i in nums %}{% if i > 1 %}{% if i < 4 %}{% break %}{% endif %}{% endif %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "1",
		},
		{
			name:     "Continue in nested if conditions",
			tmpl:     `{% for i in nums %}{% if i > 1 %}{% if i < 4 %}{% continue %}{% endif %}{% endif %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "145",
		},
		{
			name:     "Break in if-else",
			tmpl:     `{% for i in nums %}{% if i == 3 %}{% break %}{% else %}{{ i }}{% endif %}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "12",
		},
		{
			name:     "Continue in if-else",
			tmpl:     `{% for i in nums %}{% if i == 3 %}{% continue %}{% else %}{{ i }}{% endif %}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "1245",
		},
		{
			name:     "Multiple loops with different control flow",
			tmpl:     `{% for i in nums %}{% if i == 2 %}{% break %}{% endif %}{{ i }}{% endfor %}-{% for j in nums %}{% if j == 2 %}{% continue %}{% endif %}{{ j }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3}},
			expected: "1-13",
		},
		{
			name:     "Break in loop inside if",
			tmpl:     `{% if true %}{% for i in nums %}{% if i == 3 %}{% break %}{% endif %}{{ i }}{% endfor %}{% endif %}`,
			context:  Context{"nums": []int{1, 2, 3, 4, 5}},
			expected: "12",
		},
		{
			name: "Array of maps with break",
			tmpl: `{% for item in items %}{% if item.skip %}{% break %}{% endif %}{{ item.name }}|{% endfor %}`,
			context: Context{
				"items": []map[string]any{
					{"name": "a", "skip": false},
					{"name": "b", "skip": false},
					{"name": "c", "skip": true},
					{"name": "d", "skip": false},
				},
			},
			expected: "a|b|",
		},
		{
			name: "Array of maps with continue",
			tmpl: `{% for item in items %}{% if item.skip %}{% continue %}{% endif %}{{ item.name }}|{% endfor %}`,
			context: Context{
				"items": []map[string]any{
					{"name": "a", "skip": false},
					{"name": "b", "skip": true},
					{"name": "c", "skip": false},
					{"name": "d", "skip": true},
				},
			},
			expected: "a|c|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBreakContinueEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  Context
		expected string
	}{
		{
			name:     "Break immediately in nested loop",
			tmpl:     `{% for i in nums %}{% for j in nums %}{% break %}{{ j }}{% endfor %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3}},
			expected: "123",
		},
		{
			name:     "Continue immediately in nested loop",
			tmpl:     `{% for i in nums %}{% for j in nums %}{% continue %}{{ j }}{% endfor %}{{ i }}{% endfor %}`,
			context:  Context{"nums": []int{1, 2, 3}},
			expected: "123",
		},
		{
			name:     "Break with boolean array",
			tmpl:     `{% for flag in flags %}{% if flag %}{% break %}{% endif %}{{ flag }}{% endfor %}`,
			context:  Context{"flags": []bool{false, false, true, false}},
			expected: "falsefalse",
		},
		{
			name:     "Continue with boolean array",
			tmpl:     `{% for flag in flags %}{% if flag %}{% continue %}{% endif %}{{ flag }}{% endfor %}`,
			context:  Context{"flags": []bool{false, true, false, true}},
			expected: "falsefalse",
		},
		{
			name:     "Break in string with unicode",
			tmpl:     `{% for char in str %}{% if char == '界' %}{% break %}{% endif %}{{ char }}{% endfor %}`,
			context:  Context{"str": "世界你好"},
			expected: "世",
		},
		{
			name:     "Continue in string with unicode",
			tmpl:     `{% for char in str %}{% if char == '界' %}{% continue %}{% endif %}{{ char }}{% endfor %}`,
			context:  Context{"str": "世界你好"},
			expected: "世你好",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestElifBasic tests basic elif functionality with different branch conditions
func TestElifBasic(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "If true, elif ignored",
			tmpl: "{% if score >= 90 %}Excellent{% elif score >= 80 %}Good{% elif score >= 60 %}Pass{% else %}Fail{% endif %}",
			context: map[string]any{
				"score": 95,
			},
			expected: "Excellent",
		},
		{
			name: "If false, first elif true",
			tmpl: "{% if score >= 90 %}Excellent{% elif score >= 80 %}Good{% elif score >= 60 %}Pass{% else %}Fail{% endif %}",
			context: map[string]any{
				"score": 85,
			},
			expected: "Good",
		},
		{
			name: "If false, first elif false, second elif true",
			tmpl: "{% if score >= 90 %}Excellent{% elif score >= 80 %}Good{% elif score >= 60 %}Pass{% else %}Fail{% endif %}",
			context: map[string]any{
				"score": 65,
			},
			expected: "Pass",
		},
		{
			name: "All conditions false, else branch",
			tmpl: "{% if score >= 90 %}Excellent{% elif score >= 80 %}Good{% elif score >= 60 %}Pass{% else %}Fail{% endif %}",
			context: map[string]any{
				"score": 45,
			},
			expected: "Fail",
		},
		{
			name: "No else branch, all conditions false",
			tmpl: "{% if score >= 90 %}Excellent{% elif score >= 80 %}Good{% elif score >= 60 %}Pass{% endif %}",
			context: map[string]any{
				"score": 45,
			},
			expected: "",
		},
		{
			name: "Single elif with variables",
			tmpl: "{% if age < 18 %}Minor{% elif age >= 65 %}Senior{% else %}Adult{% endif %}",
			context: map[string]any{
				"age": 70,
			},
			expected: "Senior",
		},
		{
			name: "Complex conditions with filters",
			tmpl: "{% if name | length > 10 %}Long name{% elif name | length > 5 %}Medium name{% else %}Short name{% endif %}",
			context: map[string]any{
				"name": "Alexander",
			},
			expected: "Medium name",
		},
		{
			name: "Nested variables in elif",
			tmpl: "{% if user.age < 18 %}Minor{% elif user.age >= 65 %}Senior{% else %}Adult{% endif %}",
			context: map[string]any{
				"user": map[string]any{
					"age": 25,
				},
			},
			expected: "Adult",
		},
		{
			name: "String comparison in elif",
			tmpl: "{% if status == 'active' %}Active User{% elif status == 'inactive' %}Inactive User{% else %}Unknown Status{% endif %}",
			context: map[string]any{
				"status": "inactive",
			},
			expected: "Inactive User",
		},
		{
			name: "Boolean conditions in elif",
			tmpl: "{% if is_admin %}Admin{% elif is_moderator %}Moderator{% else %}User{% endif %}",
			context: map[string]any{
				"is_admin":     false,
				"is_moderator": true,
			},
			expected: "Moderator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Context(tt.context)
			result, err := Render(tt.tmpl, ctx)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestElifNested tests nested elif structures and if-elif-else combinations
func TestElifNested(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Nested if-elif inside if",
			tmpl: `{% if category == 'student' %}
				{% if age < 18 %}Child student{% elif age < 25 %}Young student{% else %}Adult student{% endif %}
			{% else %}
				Not a student
			{% endif %}`,
			context: map[string]any{
				"category": "student",
				"age":      20,
			},
			expected: "\n\t\t\t\tYoung student\n\t\t\t",
		},
		{
			name: "Nested if-elif inside elif",
			tmpl: `{% if role == 'admin' %}
				Administrator
			{% elif role == 'user' %}
				{% if level == 'premium' %}Premium user{% elif level == 'basic' %}Basic user{% else %}Guest user{% endif %}
			{% else %}
				Unknown role
			{% endif %}`,
			context: map[string]any{
				"role":  "user",
				"level": "premium",
			},
			expected: "\n\t\t\t\tPremium user\n\t\t\t",
		},
		{
			name: "Nested if-elif inside else",
			tmpl: `{% if type == 'vip' %}
				VIP member
			{% else %}
				{% if score >= 90 %}Excellent{% elif score >= 70 %}Good{% else %}Average{% endif %}
			{% endif %}`,
			context: map[string]any{
				"type":  "regular",
				"score": 85,
			},
			expected: "\n\t\t\t\tGood\n\t\t\t",
		},
		{
			name: "Multiple nested levels",
			tmpl: `{% if department == 'engineering' %}
				{% if level == 'senior' %}
					{% if experience >= 5 %}Senior Engineer{% elif experience >= 3 %}Mid-level Engineer{% else %}Junior Engineer{% endif %}
				{% elif level == 'junior' %}
					Junior Engineer
				{% else %}
					Engineer
				{% endif %}
			{% else %}
				Non-engineering role
			{% endif %}`,
			context: map[string]any{
				"department": "engineering",
				"level":      "senior",
				"experience": 4,
			},
			expected: "\n\t\t\t\t\n\t\t\t\t\tMid-level Engineer\n\t\t\t\t\n\t\t\t",
		},
		{
			name: "Nested with complex conditions",
			tmpl: `{% if user.type == 'premium' %}
				{% if user.subscription.active %}
					{% if user.subscription.plan == 'pro' %}Pro subscriber{% elif user.subscription.plan == 'basic' %}Basic subscriber{% else %}Unknown plan{% endif %}
				{% else %}
					Inactive premium user
				{% endif %}
			{% elif user.type == 'free' %}
				Free user
			{% else %}
				Unknown user type
			{% endif %}`,
			context: map[string]any{
				"user": map[string]any{
					"type": "premium",
					"subscription": map[string]any{
						"active": true,
						"plan":   "pro",
					},
				},
			},
			expected: "\n\t\t\t\t\n\t\t\t\t\tPro subscriber\n\t\t\t\t\n\t\t\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Context(tt.context)
			result, err := Render(tt.tmpl, ctx)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestElifComplexExpressions tests elif with complex condition expressions
func TestElifComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Multiple AND conditions in elif",
			tmpl: "{% if score < 50 %}Fail{% elif score >= 80 && grade == 'A' %}Excellent{% elif score >= 60 && grade == 'B' %}Good{% else %}Average{% endif %}",
			context: map[string]any{
				"score": 85,
				"grade": "A",
			},
			expected: "Excellent",
		},
		{
			name: "Multiple OR conditions in elif",
			tmpl: "{% if age < 18 %}Minor{% elif status == 'vip' || points >= 1000 %}Premium{% elif status == 'member' || points >= 500 %}Regular{% else %}Guest{% endif %}",
			context: map[string]any{
				"age":    25,
				"status": "member",
				"points": 300,
			},
			expected: "Regular",
		},
		{
			name: "Complex logical expressions",
			tmpl: "{% if (age >= 18 && age <= 65) && (income > 50000 || education == 'college') %}Qualified{% elif age > 65 && savings >= 100000 %}Senior qualified{% else %}Not qualified{% endif %}",
			context: map[string]any{
				"age":       30,
				"income":    45000,
				"education": "college",
				"savings":   50000,
			},
			expected: "Qualified",
		},
		{
			name: "Filter expressions in elif",
			tmpl: "{% if name | length < 3 %}Too short{% elif name | length > 10 %}Too long{% elif name | upper == 'ADMIN' %}Administrator{% else %}Valid name{% endif %}",
			context: map[string]any{
				"name": "admin",
			},
			expected: "Administrator",
		},
		{
			name: "Multiple filters in elif",
			tmpl: "{% if text | trim | length == 0 %}Empty{% elif text | trim | upper | length > 5 %}Long text{% elif text | trim | lower == 'hello' %}Greeting{% else %}Regular text{% endif %}",
			context: map[string]any{
				"text": "  HELLO  ",
			},
			expected: "Greeting",
		},
		{
			name: "Nested object access in elif",
			tmpl: "{% if user.profile.level == 'admin' %}Admin{% elif user.profile.level == 'moderator' && user.profile.permissions.write %}Moderator{% elif user.profile.active %}Active user{% else %}Inactive user{% endif %}",
			context: map[string]any{
				"user": map[string]any{
					"profile": map[string]any{
						"level":  "moderator",
						"active": true,
						"permissions": map[string]any{
							"write": true,
						},
					},
				},
			},
			expected: "Moderator",
		},
		{
			name: "Array operations in elif",
			tmpl: "{% if items | size == 0 %}Empty list{% elif items | size > 10 %}Large list{% elif items | size > 5 %}Medium list{% else %}Small list{% endif %}",
			context: map[string]any{
				"items": []string{"a", "b", "c", "d", "e", "f", "g"},
			},
			expected: "Medium list",
		},
		{
			name: "Mixed arithmetic and comparison in elif",
			tmpl: "{% if score * 2 < 100 %}Low{% elif score + bonus >= 150 %}High{% elif score - penalty > 80 %}Medium{% else %}Average{% endif %}",
			context: map[string]any{
				"score":   90,
				"bonus":   70,
				"penalty": 5,
			},
			expected: "High",
		},
		{
			name: "String comparison with filters",
			tmpl: "{% if status | lower == 'active' %}Active{% elif status | lower == 'pending' %}Pending{% elif status | lower == 'inactive' %}Inactive{% else %}Unknown{% endif %}",
			context: map[string]any{
				"status": "PENDING",
			},
			expected: "Pending",
		},
		{
			name: "Complex boolean expressions",
			tmpl: "{% if (is_admin || is_moderator) && is_active %}Privileged{% elif is_user && (has_subscription || trial_active) %}Subscriber{% elif is_guest %}Guest{% else %}Unknown{% endif %}",
			context: map[string]any{
				"is_admin":         false,
				"is_moderator":     false,
				"is_user":          true,
				"is_active":        true,
				"has_subscription": false,
				"trial_active":     true,
				"is_guest":         false,
			},
			expected: "Subscriber",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Context(tt.context)
			result, err := Render(tt.tmpl, ctx)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestElifComplexNestedWithLoops tests complex nested if-elif-else structures with for loops and control flow
func TestElifComplexNestedWithLoops(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        string
		context     map[string]any
		expected    string
		expectError bool
	}{
		{
			name: "Complex nested if-elif-else with for loops and break/continue",
			tmpl: `{% for user in users %}
User: {{ user.name }}
{% if user.role == 'admin' %}
  Admin Level: {% if user.level >= 3 %}Senior Admin{% elif user.level >= 2 %}Mid Admin{% else %}Junior Admin{% endif %}
  {% for permission in user.permissions %}
    {% if permission == 'read' %}{% continue %}{% endif %}
    {% if permission == 'delete' %}DELETE ACCESS{% break %}{% endif %}
    Permission: {{ permission }}
  {% endfor %}
{% elif user.role == 'user' %}
  User Type: {% if user.active %}Active{% else %}Inactive{% endif %}
  {% if user.subscription %}
    Plan: {% if user.subscription.type == 'premium' %}Premium{% elif user.subscription.type == 'basic' %}Basic{% else %}Free{% endif %}
  {% endif %}
{% else %}
  Unknown Role
{% endif %}
---
{% endfor %}`,
			context: map[string]any{
				"users": []map[string]any{
					{
						"name":        "Alice",
						"role":        "admin",
						"level":       3,
						"permissions": []string{"read", "write", "delete", "admin"},
					},
					{
						"name":   "Bob",
						"role":   "user",
						"active": true,
						"subscription": map[string]any{
							"type": "premium",
						},
					},
					{
						"name":   "Charlie",
						"role":   "user",
						"active": false,
						"subscription": map[string]any{
							"type": "basic",
						},
					},
				},
			},
			expected: `
User: Alice

  Admin Level: Senior Admin
  
    
    
    
    Permission: write
  
    
    DELETE ACCESS

---

User: Bob

  User Type: Active
  
    Plan: Premium
  

---

User: Charlie

  User Type: Inactive
  
    Plan: Basic
  

---
`,
			expectError: false,
		},
		{
			name: "Nested elif with complex conditions and loops",
			tmpl: `{% for item in items %}
Item: {{ item.name }}
{% if item.category == 'electronics' %}
  {% if item.price >= 1000 %}
    Expensive Electronics
    {% for feature in item.features %}
      {% if feature | length < 3 %}{% continue %}{% endif %}
      Feature: {{ feature }}
      {% if feature == 'wireless' %}{% break %}{% endif %}
    {% endfor %}
  {% elif item.price >= 500 %}
    Mid-range Electronics
  {% else %}
    Budget Electronics
  {% endif %}
{% elif item.category == 'books' %}
  {% if item.pages >= 500 %}Heavy Read{% elif item.pages >= 200 %}Medium Read{% else %}Light Read{% endif %}
{% elif item.category == 'clothing' %}
  {% if item.size == 'XL' || item.size == 'XXL' %}Large Size{% else %}Regular Size{% endif %}
{% else %}
  Unknown Category
{% endif %}
---
{% endfor %}`,
			context: map[string]any{
				"items": []map[string]any{
					{
						"name":     "Laptop",
						"category": "electronics",
						"price":    1200,
						"features": []string{"SSD", "4K", "wireless", "bluetooth"},
					},
					{
						"name":     "Novel",
						"category": "books",
						"pages":    350,
					},
					{
						"name":     "Shirt",
						"category": "clothing",
						"size":     "L",
					},
				},
			},
			expected: `
Item: Laptop

  
    Expensive Electronics
    
      
      Feature: SSD
      
    
      
      
      Feature: wireless
      
  

---

Item: Novel

  Medium Read

---

Item: Shirt

  Regular Size

---
`,
			expectError: false,
		},
		{
			name: "Deep nested if-elif-else with multiple conditions",
			tmpl: `{% for company in companies %}
Company: {{ company.name }}
{% if company.size == 'large' %}
  Large Company
  {% for dept in company.departments %}
    Dept: {{ dept.name }}
    {% if dept.budget >= 1000000 %}
      High Budget Department
      {% if dept.manager.experience >= 10 %}
        Experienced Manager
      {% elif dept.manager.experience >= 5 %}
        Mid-level Manager
      {% else %}
        Junior Manager
      {% endif %}
    {% elif dept.budget >= 500000 %}
      Medium Budget Department
    {% else %}
      Low Budget Department
    {% endif %}
  {% endfor %}
{% elif company.size == 'medium' %}
  Medium Company
{% else %}
  Small Company
{% endif %}
---
{% endfor %}`,
			context: map[string]any{
				"companies": []map[string]any{
					{
						"name": "TechCorp",
						"size": "large",
						"departments": []map[string]any{
							{
								"name":   "Engineering",
								"budget": 1500000,
								"manager": map[string]any{
									"experience": 12,
								},
							},
							{
								"name":   "Marketing",
								"budget": 300000,
								"manager": map[string]any{
									"experience": 3,
								},
							},
						},
					},
				},
			},
			expected: `
Company: TechCorp

  Large Company
  
    Dept: Engineering
    
      High Budget Department
      
        Experienced Manager
      
    
  
    Dept: Marketing
    
      Low Budget Department
    
  

---
`,
			expectError: false,
		},
		{
			name: "Multiple nested loops with complex elif chains",
			tmpl: `{% for category in categories %}
Category: {{ category.name }}
{% for product in category.products %}
  Product: {{ product.name }}
  {% if product.rating >= 4.5 %}
    Excellent Product
    {% for review in product.reviews %}
      {% if review.rating < 4 %}{% continue %}{% endif %}
      {% if review.verified %}
        Verified Review: {{ review.comment }}
      {% elif review.helpful_count > 10 %}
        Helpful Review: {{ review.comment }}
      {% else %}
        Regular Review: {{ review.comment }}
      {% endif %}
      {% if review.featured %}{% break %}{% endif %}
    {% endfor %}
  {% elif product.rating >= 3.5 %}
    Good Product
  {% elif product.rating >= 2.5 %}
    Average Product
  {% else %}
    Poor Product
  {% endif %}
{% endfor %}
---
{% endfor %}`,
			context: map[string]any{
				"categories": []map[string]any{
					{
						"name": "Electronics",
						"products": []map[string]any{
							{
								"name":   "Smartphone",
								"rating": 4.6,
								"reviews": []map[string]any{
									{
										"rating":        5,
										"comment":       "Great phone!",
										"verified":      true,
										"helpful_count": 5,
										"featured":      false,
									},
									{
										"rating":        4,
										"comment":       "Good value",
										"verified":      false,
										"helpful_count": 15,
										"featured":      true,
									},
								},
							},
						},
					},
				},
			},
			expected: `
Category: Electronics

  Product: Smartphone
  
    Excellent Product
    
      
      
        Verified Review: Great phone!
      
      
    
      
      
        Helpful Review: Good value
      
      
  

---
`,
			expectError: false,
		},
		{
			name: "Error case: elif without matching if",
			tmpl: `{% for item in items %}
  Item: {{ item.name }}
  {% elif item.active %}
    Active item
  {% endif %}
{% endfor %}`,
			context: map[string]any{
				"items": []map[string]any{
					{"name": "Item1", "active": true},
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "Error case: else without matching if",
			tmpl: `{% for user in users %}
  User: {{ user.name }}
  {% else %}
    No users
  {% endif %}
{% endfor %}`,
			context: map[string]any{
				"users": []map[string]any{
					{"name": "User1"},
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "Error case: nested elif without matching if",
			tmpl: `{% for category in categories %}
  Category: {{ category.name }}
  {% for product in category.products %}
    Product: {{ product.name }}
    {% elif product.featured %}
      Featured product
    {% endif %}
  {% endfor %}
{% endfor %}`,
			context: map[string]any{
				"categories": []map[string]any{
					{
						"name": "Electronics",
						"products": []map[string]any{
							{"name": "Phone", "featured": true},
						},
					},
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "Error case: multiple else statements",
			tmpl: `{% for item in items %}
  {% if item.active %}
    Active
  {% else %}
    Inactive
  {% else %}
    Another else
  {% endif %}
{% endfor %}`,
			context: map[string]any{
				"items": []map[string]any{
					{"active": true},
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "Error case: multiple else statements in nested if",
			tmpl: `{% if condition1 %}
  {% if condition2 %}
    Inner if
  {% else %}
    First else
  {% else %}
    Second else
  {% endif %}
{% endif %}`,
			context: map[string]any{
				"condition1": true,
				"condition2": false,
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "Error case: three else statements",
			tmpl: `{% if condition %}
  If branch
{% else %}
  First else
{% else %}
  Second else
{% else %}
  Third else
{% endif %}`,
			context: map[string]any{
				"condition": false,
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "Valid case: single else with multiple elif",
			tmpl: `{% if score >= 90 %}
  Excellent
{% elif score >= 80 %}
  Good
{% elif score >= 70 %}
  Average
{% else %}
  Poor
{% endif %}`,
			context: map[string]any{
				"score": 65,
			},
			expected: `
  Poor
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Context(tt.context)
			result, err := Render(tt.tmpl, ctx)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none. Result: %s", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCompleteParameterGeneration(t *testing.T) {
	type Schema struct {
		Type string `json:"type"`
	}

	type Param struct {
		In     string `json:"in"`
		Schema Schema `json:"schema"`
	}

	type Operation struct {
		RequestBody bool             `json:"requestBody"`
		Parameters  map[string]Param `json:"parameters"`
	}

	type Data struct {
		Operation Operation `json:"operation"`
	}

	tests := []struct {
		name        string
		context     map[string]any
		tmpl        string
		expected    string
		expectError bool
	}{
		{
			name: "String query parameter with request body",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: true,
						Parameters: map[string]Param{
							"username": {
								In:     "query",
								Schema: Schema{Type: "string"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
			// Apply {{ paramName }} parameter ({{ param.in }})
    {% if param.schema.type == "string" %}
    if r.{{ paramName }} != "" {
    {% else %}
    if r.{{ paramName }} != nil {
    {% endif %}
        {% if param.in == "query" %}
        req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% elif param.in == "path" %}
        // Path parameters are handled in the URL construction
        {% elif param.in == "header" %}
        req = req.Header("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% elif param.in == "cookie" %}
        req = req.Cookie("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% else %}
        req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% endif %}
    }
    {% endfor %}
    
    {% if data.operation.requestBody %}
    // Apply request body
    req = req.JSONBody(r.Body)
    {% endif %}`,
			expected: `
			// Apply username parameter (query)
    
    if r.username != "" {
    
        
        req = req.Query("username", fmt.Sprintf("%v", r.username))
        
    }
    
    
    
    // Apply request body
    req = req.JSONBody(r.Body)
    `,
		},
		{
			name: "Non-string header parameter without request body",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: false,
						Parameters: map[string]Param{
							"authorization": {
								In:     "header",
								Schema: Schema{Type: "integer"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
			// Apply {{ paramName }} parameter ({{ param.in }})
    {% if param.schema.type == "string" %}
    if r.{{ paramName }} != "" {
    {% else %}
    if r.{{ paramName }} != nil {
    {% endif %}
        {% if param.in == "query" %}
        req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% elif param.in == "path" %}
        // Path parameters are handled in the URL construction
        {% elif param.in == "header" %}
        req = req.Header("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% elif param.in == "cookie" %}
        req = req.Cookie("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% else %}
        req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% endif %}
    }
    {% endfor %}
    
    {% if data.operation.requestBody %}
    // Apply request body
    req = req.JSONBody(r.Body)
    {% endif %}`,
			expected: `
			// Apply authorization parameter (header)
    
    if r.authorization != nil {
    
        
        req = req.Header("authorization", fmt.Sprintf("%v", r.authorization))
        
    }
    
    
    `,
		},
		{
			name: "Path parameter",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: false,
						Parameters: map[string]Param{
							"userId": {
								In:     "path",
								Schema: Schema{Type: "string"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
			// Apply {{ paramName }} parameter ({{ param.in }})
    {% if param.schema.type == "string" %}
    if r.{{ paramName }} != "" {
    {% else %}
    if r.{{ paramName }} != nil {
    {% endif %}
        {% if param.in == "query" %}
        req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% elif param.in == "path" %}
        // Path parameters are handled in the URL construction
        {% elif param.in == "header" %}
        req = req.Header("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% elif param.in == "cookie" %}
        req = req.Cookie("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% else %}
        req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% endif %}
    }
    {% endfor %}
    
    {% if data.operation.requestBody %}
    // Apply request body
    req = req.JSONBody(r.Body)
    {% endif %}`,
			expected: `
			// Apply userId parameter (path)
    
    if r.userId != "" {
    
        
        // Path parameters are handled in the URL construction
        
    }
    
    
    `,
		},
		{
			name: "Cookie parameter",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: false,
						Parameters: map[string]Param{
							"sessionId": {
								In:     "cookie",
								Schema: Schema{Type: "string"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
			// Apply {{ paramName }} parameter ({{ param.in }})
    {% if param.schema.type == "string" %}
    if r.{{ paramName }} != "" {
    {% else %}
    if r.{{ paramName }} != nil {
    {% endif %}
        {% if param.in == "query" %}
        req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% elif param.in == "path" %}
        // Path parameters are handled in the URL construction
        {% elif param.in == "header" %}
        req = req.Header("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% elif param.in == "cookie" %}
        req = req.Cookie("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% else %}
        req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% endif %}
    }
    {% endfor %}
    
    {% if data.operation.requestBody %}
    // Apply request body
    req = req.JSONBody(r.Body)
    {% endif %}`,
			expected: `
			// Apply sessionId parameter (cookie)
    
    if r.sessionId != "" {
    
        
        req = req.Cookie("sessionId", fmt.Sprintf("%v", r.sessionId))
        
    }
    
    
    `,
		},
		{
			name: "Default case (unknown parameter type)",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: true,
						Parameters: map[string]Param{
							"data": {
								In:     "body",
								Schema: Schema{Type: "object"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
			// Apply {{ paramName }} parameter ({{ param.in }})
    {% if param.schema.type == "string" %}
    if r.{{ paramName }} != "" {
    {% else %}
    if r.{{ paramName }} != nil {
    {% endif %}
        {% if param.in == "query" %}
        req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% elif param.in == "path" %}
        // Path parameters are handled in the URL construction
        {% elif param.in == "header" %}
        req = req.Header("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% elif param.in == "cookie" %}
        req = req.Cookie("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% else %}
        req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% endif %}
    }
    {% endfor %}
    
    {% if data.operation.requestBody %}
    // Apply request body
    req = req.JSONBody(r.Body)
    {% endif %}`,
			expected: `
			// Apply data parameter (body)
    
    if r.data != nil {
    
        
        req = req.Query("data", fmt.Sprintf("%v", r.data))
        
    }
    
    
    
    // Apply request body
    req = req.JSONBody(r.Body)
    `,
		},
		{
			name: "Simple elif test with single parameter",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: true,
						Parameters: map[string]Param{
							"username": {
								In:     "query",
								Schema: Schema{Type: "string"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
{% if param.schema.type == "string" %}
String param: {{ paramName }}
{% elif param.schema.type == "integer" %}
Integer param: {{ paramName }}
{% else %}
Other param: {{ paramName }}
{% endif %}
{% endfor %}`,
			expected: `

String param: username

`,
		},
		{
			name: "Integer parameter test",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: false,
						Parameters: map[string]Param{
							"count": {
								In:     "query",
								Schema: Schema{Type: "integer"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
{% if param.schema.type == "string" %}
String param: {{ paramName }}
{% elif param.schema.type == "integer" %}
Integer param: {{ paramName }}
{% else %}
Other param: {{ paramName }}
{% endif %}
{% endfor %}`,
			expected: `

Integer param: count

`,
		},
		{
			name: "Empty parameters map",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: false,
						Parameters:  map[string]Param{},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
    // Process {{ paramName }}
{% endfor %}    
    // No request body
    `,
			expected: `    
    // No request body
    `,
		},
		{
			name: "String query parameter with nested if elif",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: false,
						Parameters: map[string]Param{
							"status": {
								In:     "query",
								Schema: Schema{Type: "string"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
    {% if param.in == "query" %}
        {% if param.schema.type == "string" %}
            // String query param
            if r.{{ paramName }} != "" {
                
                    req = req.Query("{{ paramName }}", r.{{ paramName }})
                    
            }
        {% elif param.schema.type == "integer" %}
            // Integer query param
            req = req.Query("{{ paramName }}", fmt.Sprintf("%d", r.{{ paramName }}))
        {% else %}
            // Other query param
            req = req.Query("{{ paramName }}", fmt.Sprintf("%v", r.{{ paramName }}))
        {% endif %}
    {% elif param.in == "header" %}
        // Header param
        req = req.Header("{{ paramName }}", r.{{ paramName }})
    {% else %}
        // Other param
        req = req.Query("{{ paramName }}", r.{{ paramName }})
    {% endif %}
{% endfor %}`,
			expected: `
    
        
            // String query param
            if r.status != "" {
                
                    req = req.Query("status", r.status)
                    
            }
        
    
`,
		},
		{
			name: "Unmatched elif in for loop (should error)",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: false,
						Parameters: map[string]Param{
							"test": {
								In:     "query",
								Schema: Schema{Type: "string"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
    // Process {{ paramName }}
    {% elif param.schema.type == "string" %}
    String parameter
    {% endif %}
{% endfor %}`,
			expected:    "",
			expectError: true,
		},
		{
			name: "Unmatched else in for loop (should error)",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: false,
						Parameters: map[string]Param{
							"test": {
								In:     "query",
								Schema: Schema{Type: "string"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
    // Process {{ paramName }}
    {% else %}
    No parameters
    {% endif %}
{% endfor %}`,
			expected:    "",
			expectError: true,
		},
		{
			name: "Error case: multiple else statements in nested structure",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: false,
						Parameters: map[string]Param{
							"test": {
								In:     "query",
								Schema: Schema{Type: "string"},
							},
						},
					},
				},
			},
			tmpl: `{% for paramName, param in data.operation.parameters %}
    {% if param.schema.type == "string" %}
        String type
    {% else %}
        Not string
    {% else %}
        Another else
    {% endif %}
{% endfor %}`,
			expected:    "",
			expectError: true,
		},
		{
			name: "Deep nested for loops with if elif",
			context: map[string]any{
				"category": map[string]any{
					"name": "API",
					"operations": []map[string]any{
						{
							"name": "GetUser",
							"parameters": map[string]Param{
								"format": {
									In:     "query",
									Schema: Schema{Type: "string"},
								},
							},
						},
					},
				},
			},
			tmpl: `Category: {{ category.name }}
{% for operation in category.operations %}
  Operation: {{ operation.name }}
  {% for paramName, param in operation.parameters %}
    {% if param.schema.type == "string" %}
      {% if param.in == "query" %}
      - Query string param: {{ paramName }}
      {% elif param.in == "header" %}
      - Header string param: {{ paramName }}
      {% else %}
      - Other string param: {{ paramName }}
      {% endif %}
    {% elif param.schema.type == "integer" %}
      {% if param.in == "path" %}
      - Path integer param: {{ paramName }}
      {% elif param.in == "query" %}
      - Query integer param: {{ paramName }}
      {% else %}
      - Other integer param: {{ paramName }}
      {% endif %}
    {% else %}
      - Unknown param type: {{ paramName }}
    {% endif %}
  {% endfor %}
{% endfor %}`,
			expected: `Category: API

  Operation: GetUser
  
    
      
      - Query string param: format
      
    
  
`,
		},
		{
			name: "Complex validation with nested conditions",
			context: map[string]any{
				"data": Data{
					Operation: Operation{
						RequestBody: true,
						Parameters: map[string]Param{
							"required": {
								In:     "header",
								Schema: Schema{Type: "string"},
							},
						},
					},
				},
				"config": map[string]any{
					"validateParams": true,
					"debugMode":      false,
				},
			},
			tmpl: `{% if config.debugMode %}
// Debug mode enabled
{% endif %}

{% for paramName, param in data.operation.parameters %}
    {% if config.validateParams %}
        {% if param.schema.type == "string" %}
            {% if paramName == "required" %}
            // Required parameter validation
            if r.{{ paramName }} == "" {
                return errors.New("{{ paramName }} is required")
            }
            {% elif paramName == "optional" %}
            // Optional parameter
            {% else %}
            // Regular parameter
            {% endif %}
        {% endif %}
    {% endif %}
    
    {% if param.in == "query" %}
    req = req.Query("{{ paramName }}", r.{{ paramName }})
    {% elif param.in == "header" %}
    req = req.Header("{{ paramName }}", r.{{ paramName }})
    {% endif %}
{% endfor %}

{% if data.operation.requestBody %}
req = req.JSONBody(r.Body) 
{% endif %}`,
			expected: `


    
        
            
            // Required parameter validation
            if r.required == "" {
                return errors.New("required is required")
            }
            
        
    
    
    
    req = req.Header("required", r.required)
    



req = req.JSONBody(r.Body) 
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Context(tt.context)
			result, err := Render(tt.tmpl, ctx)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none. Result: %s", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestLoopContextFeatures tests the comprehensive loop context functionality
func TestLoopContextFeatures(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        string
		context     map[string]any
		expected    string
		expectError bool
	}{
		{
			name: "Basic loop context properties",
			tmpl: `{% for item in items %}{{ loop.Index }}-{{ loop.Revindex }}-{{ loop.First }}-{{ loop.Last }}-{{ loop.Length }}: {{ item }}
{% endfor %}`,
			context: map[string]any{
				"items": []string{"apple", "banana", "cherry"},
			},
			expected: `0-2-true-false-3: apple
1-1-false-false-3: banana
2-0-false-true-3: cherry
`,
		},
		{
			name: "Loop context with string iteration",
			tmpl: `{% for char in word %}[{{ loop.Index }}:{{ char }}]{% if loop.Last == false %}-{% endif %}{% endfor %}`,
			context: map[string]any{
				"word": "hello",
			},
			expected: `[0:h]-[1:e]-[2:l]-[3:l]-[4:o]`,
		},
		{
			name: "Loop context with map iteration",
			tmpl: `{% for key, value in data %}{{ loop.Index }}: {{ key }}={{ value }}{% if loop.Last == false %}, {% endif %}{% endfor %}`,
			context: map[string]any{
				"data": map[string]string{"a": "1", "b": "2", "c": "3"},
			},
			expected: `0: a=1, 1: b=2, 2: c=3`,
		},
		{
			name: "Simple nested loops with loop context",
			tmpl: `{% for outer in outers %}Outer[{{ loop.Index }}]:
{% for inner in inners %}  Inner[{{ loop.Index }}]: {{ outer }}-{{ inner }}
{% endfor %}{% endfor %}`,
			context: map[string]any{
				"outers": []string{"A", "B"},
				"inners": []string{"1", "2"},
			},
			expected: `Outer[0]:
  Inner[0]: A-1
  Inner[1]: A-2
Outer[1]:
  Inner[0]: B-1
  Inner[1]: B-2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Context(tt.context)
			result, err := Render(tt.tmpl, ctx)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none. Result: %s", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLoopContext(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		ctx      Context
		expected string
	}{
		{
			name:     "Simple loop with loop.Index",
			tmpl:     "{% for item in items %}{{ loop.Index }}: {{ item }}\n{% endfor %}",
			ctx:      Context{"items": []string{"a", "b", "c"}},
			expected: "0: a\n1: b\n2: c\n",
		},
		{
			name:     "Loop with all loop properties",
			tmpl:     "{% for item in items %}{{ loop.Index }}-{{ loop.Revindex }}-{{ loop.First }}-{{ loop.Last }}-{{ loop.Length }}: {{ item }}\n{% endfor %}",
			ctx:      Context{"items": []string{"x", "y"}},
			expected: "0-1-true-false-2: x\n1-0-false-true-2: y\n",
		},
		{
			name:     "Nested loops",
			tmpl:     "{% for outer in outers %}Outer {{ loop.Index }}:\n{% for inner in inners %}  Inner {{ loop.Index }} (outer was {{ outer }})\n{% endfor %}{% endfor %}",
			ctx:      Context{"outers": []string{"A", "B"}, "inners": []string{"1", "2"}},
			expected: "Outer 0:\n  Inner 0 (outer was A)\n  Inner 1 (outer was A)\nOuter 1:\n  Inner 0 (outer was B)\n  Inner 1 (outer was B)\n",
		},
		{
			name:     "String iteration with loop",
			tmpl:     "{% for char in word %}{{ loop.Index }}: {{ char }}{% if loop.Last == false %}, {% endif %}{% endfor %}",
			ctx:      Context{"word": "hello"},
			expected: "0: h, 1: e, 2: l, 3: l, 4: o",
		},
		{
			name:     "Map iteration with loop",
			tmpl:     "{% for k, v in map %}{{ loop.Index }}: {{ k }}={{ v }}{% if loop.Last == false %}, {% endif %}{% endfor %}",
			ctx:      Context{"map": map[string]string{"a": "1", "b": "2"}},
			expected: "0: a=1, 1: b=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.tmpl)
			if err != nil {
				t.Fatalf("compiling template: %v", err)
			}

			result, err := tmpl.Render(tt.ctx)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNestedLoopContext(t *testing.T) {
	tmpl := `{% for outer in outers %}
Outer loop {{ loop.Index }} (first: {{ loop.First }}, last: {{ loop.Last }}):
{% for inner in inners %}
  Inner loop {{ loop.Index }} (first: {{ loop.First }}, last: {{ loop.Last }})
{% endfor %}
{% endfor %}`

	ctx := Context{
		"outers": []string{"A", "B"},
		"inners": []string{"1", "2", "3"},
	}

	expected := `
Outer loop 0 (first: true, last: false):

  Inner loop 0 (first: true, last: false)

  Inner loop 1 (first: false, last: false)

  Inner loop 2 (first: false, last: true)


Outer loop 1 (first: false, last: true):

  Inner loop 0 (first: true, last: false)

  Inner loop 1 (first: false, last: false)

  Inner loop 2 (first: false, last: true)

`

	compiled, err := Compile(tmpl)
	if err != nil {
		t.Fatalf("compiling template: %v", err)
	}

	result, err := compiled.Render(ctx)
	if err != nil {
		t.Fatalf("rendering template: %v", err)
	}

	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestHighlyComplexNestedLoops(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		ctx      Context
		expected string
	}{
		{
			name: "5-level nested for loops with conditional logic",
			tmpl: `{% for level1 in data.level1 %}L1[{{ loop.Index }}/{{ loop.Length }}]{% if loop.First %}:FIRST{% endif %}
{% for level2 in data.level2 %}  L2[{{ loop.Index }}/{{ loop.Length }}]{% if loop.Last %}:LAST{% endif %}
{% if loop.Index < 2 %}{% for level3 in data.level3 %}    L3[{{ loop.Index }}/{{ loop.Length }}]{% if loop.First %}{% if loop.Last %}:ONLY{% endif %}{% endif %}
{% if loop.Revindex > 0 %}{% for level4 in data.level4 %}      L4[{{ loop.Index }}/{{ loop.Length }}]
{% if loop.Index == 1 %}{% for level5 in data.level5 %}        L5[{{ loop.Index }}/{{ loop.Length }}]{% if loop.Last %}:END{% endif %}
{% endfor %}{% endif %}{% endfor %}{% endif %}{% endfor %}{% endif %}{% endfor %}{% endfor %}`,
			ctx: Context{
				"data": map[string]any{
					"level1": []int{10, 20},
					"level2": []string{"A", "B"},
					"level3": []bool{true},
					"level4": []string{"X", "Y"},
					"level5": []int{100, 200},
				},
			},
			expected: `L1[0/2]:FIRST
  L2[0/2]
    L3[0/1]:ONLY
  L2[1/2]:LAST
    L3[0/1]:ONLY
L1[1/2]
  L2[0/2]
    L3[0/1]:ONLY
  L2[1/2]:LAST
    L3[0/1]:ONLY
`,
		},
		{
			name: "6-level ultra-complex nested loops",
			tmpl: `{% for a in arrays.a %}A{{ loop.Index }}{% if loop.First %}(start){% endif %}{% if loop.Last %}(end){% endif %}
{% for b in arrays.b %}  B{{ loop.Index }}/{{ loop.Revindex }}
{% for c in arrays.c %}    C{{ loop.Index }}-{{ loop.First }}-{{ loop.Last }}
{% for d in arrays.d %}      D{{ loop.Index }}:{{ loop.Length }}
{% if loop.First %}{% for e in arrays.e %}        E{{ loop.Index }}{% if loop.First %}!{% endif %}
{% for f in arrays.f %}          F{{ loop.Index }}/{{ loop.Length }}{% if loop.Last %}*{% endif %}
{% endfor %}{% endfor %}{% endif %}{% endfor %}{% endfor %}{% endfor %}{% endfor %}`,
			ctx: Context{
				"arrays": map[string]any{
					"a": []int{1, 2},
					"b": []string{"x", "y"},
					"c": []bool{true, false},
					"d": []int{10, 20},
					"e": []string{"p", "q"},
					"f": []int{100, 200, 300},
				},
			},
			expected: `A0(start)
  B0/1
    C0-true-false
      D0:2
        E0!
          F0/3
          F1/3
          F2/3*
        E1
          F0/3
          F1/3
          F2/3*
      D1:2
    C1-false-true
      D0:2
        E0!
          F0/3
          F1/3
          F2/3*
        E1
          F0/3
          F1/3
          F2/3*
      D1:2
  B1/0
    C0-true-false
      D0:2
        E0!
          F0/3
          F1/3
          F2/3*
        E1
          F0/3
          F1/3
          F2/3*
      D1:2
    C1-false-true
      D0:2
        E0!
          F0/3
          F1/3
          F2/3*
        E1
          F0/3
          F1/3
          F2/3*
      D1:2
A1(end)
  B0/1
    C0-true-false
      D0:2
        E0!
          F0/3
          F1/3
          F2/3*
        E1
          F0/3
          F1/3
          F2/3*
      D1:2
    C1-false-true
      D0:2
        E0!
          F0/3
          F1/3
          F2/3*
        E1
          F0/3
          F1/3
          F2/3*
      D1:2
  B1/0
    C0-true-false
      D0:2
        E0!
          F0/3
          F1/3
          F2/3*
        E1
          F0/3
          F1/3
          F2/3*
      D1:2
    C1-false-true
      D0:2
        E0!
          F0/3
          F1/3
          F2/3*
        E1
          F0/3
          F1/3
          F2/3*
      D1:2
`,
		},
		{
			name: "Complex matrix iteration with all loop properties",
			tmpl: `{% for row in matrix %}ROW[{{ loop.Index }}]{% if loop.First %}:HEADER{% endif %}{% if loop.Last %}:FOOTER{% endif %} (rev:{{ loop.Revindex }})
{% for col in row %}  COL[{{ loop.Index }}]{% if loop.First %}:LEFT{% endif %}{% if loop.Last %}:RIGHT{% endif %}
{% if loop.Index == 1 %}{% for item in col %}    ITEM[{{ loop.Index }}/{{ loop.Length }}] = {{ item }}
{% if loop.Revindex == 1 %}{% for detail in details %}      DETAIL[{{ loop.Index }}]{% if loop.First %}:START{% endif %}{% if loop.Last %}:END{% endif %}
{% if loop.Index == 1 %}{% for sub in subs %}        SUB[{{ loop.Index }}]: {{ sub }}{% if loop.Last %} (FINAL){% endif %}
{% if loop.First %}{% for meta in metas %}          META[{{ loop.Index }}]: {{ meta }} (pos:{{ loop.Revindex }})
{% endfor %}{% endif %}{% endfor %}{% endif %}{% endfor %}{% endif %}{% endfor %}{% endif %}{% endfor %}{% endfor %}`,
			ctx: Context{
				"matrix": [][][]string{
					{{"a1"}, {"b1", "b2"}, {"c1"}},
					{{"d1"}, {"e1", "e2", "e3"}, {"f1"}},
				},
				"details": []string{"detail1", "detail2", "detail3"},
				"subs":    []string{"sub1", "sub2"},
				"metas":   []string{"meta1", "meta2"},
			},
			expected: `ROW[0]:HEADER (rev:1)
  COL[0]:LEFT
  COL[1]
    ITEM[0/2] = b1
      DETAIL[0]:START
      DETAIL[1]
        SUB[0]: sub1
          META[0]: meta1 (pos:1)
          META[1]: meta2 (pos:0)
        SUB[1]: sub2 (FINAL)
      DETAIL[2]:END
    ITEM[1/2] = b2
  COL[2]:RIGHT
ROW[1]:FOOTER (rev:0)
  COL[0]:LEFT
  COL[1]
    ITEM[0/3] = e1
    ITEM[1/3] = e2
      DETAIL[0]:START
      DETAIL[1]
        SUB[0]: sub1
          META[0]: meta1 (pos:1)
          META[1]: meta2 (pos:0)
        SUB[1]: sub2 (FINAL)
      DETAIL[2]:END
    ITEM[2/3] = e3
  COL[2]:RIGHT
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.tmpl)
			if err != nil {
				t.Fatalf("compiling template: %v", err)
			}

			result, err := tmpl.Render(tt.ctx)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// Moved from loop_features_test.go
func TestLoopContextFullFeatures(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		context  map[string]any
		expected string
	}{
		{
			name: "Index (0-indexed)",
			tmpl: `{% for item in items %}{{ loop.Index }},{% endfor %}`,
			context: map[string]any{
				"items": []string{"a", "b", "c"},
			},
			expected: "0,1,2,",
		},
		{
			name: "Counter (1-indexed)",
			tmpl: `{% for item in items %}{{ loop.Counter }},{% endfor %}`,
			context: map[string]any{
				"items": []string{"a", "b", "c"},
			},
			expected: "1,2,3,",
		},
		{
			name: "Revindex",
			tmpl: `{% for item in items %}{{ loop.Revindex }},{% endfor %}`,
			context: map[string]any{
				"items": []string{"a", "b", "c"},
			},
			expected: "2,1,0,",
		},
		{
			name: "Revcounter",
			tmpl: `{% for item in items %}{{ loop.Revcounter }},{% endfor %}`,
			context: map[string]any{
				"items": []string{"a", "b", "c"},
			},
			expected: "3,2,1,",
		},
		{
			name: "First and Last",
			tmpl: `{% for item in items %}{% if loop.First %}FIRST{% endif %}{{ item }}{% if loop.Last %}LAST{% endif %},{% endfor %}`,
			context: map[string]any{
				"items": []string{"a", "b", "c"},
			},
			expected: "FIRSTa,b,cLAST,",
		},
		{
			name: "Length",
			tmpl: `{% for item in items %}{{ loop.Index }}/{{ loop.Length }},{% endfor %}`,
			context: map[string]any{
				"items": []string{"a", "b", "c"},
			},
			expected: "0/3,1/3,2/3,",
		},
		{
			name: "Nested loop with Parent",
			tmpl: `{% for outer in outers %}[O{{ loop.Counter }}]{% for inner in inners %}{{ loop.Parent.Counter }}-{{ loop.Counter }},{% endfor %};{% endfor %}`,
			context: map[string]any{
				"outers": []string{"A", "B"},
				"inners": []string{"x", "y"},
			},
			expected: "[O1]1-1,1-2,;[O2]2-1,2-2,;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.tmpl, tt.context)
			if err != nil {
				t.Fatalf("rendering template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}
