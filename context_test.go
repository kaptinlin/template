package template

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEmptyContextInitialization(t *testing.T) {
	ctx := NewContext()
	if len(ctx) != 0 {
		t.Errorf("NewContext() = %v, want empty context", ctx)
	}
}

func TestContextSetAndGet(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  any
		want any
	}{
		{"string", "stringValue", "hello world", "hello world"},
		{"integer", "intValue", 28, 28},
		{"bool true", "boolValueTrue", true, true},
		{"float", "floatValue", 1.75, 1.75},
		{"string slice", "sliceOfString", []string{"Go", "Python", "JavaScript"}, []string{"Go", "Python", "JavaScript"}},
		{"int slice", "sliceOfInt", []int{1, 2, 3}, []int{1, 2, 3}},
		{"2d slice", "multiDimSlice", [][]int{{1, 2}, {3, 4}}, [][]int{{1, 2}, {3, 4}}},
		{"nil", "nilValue", nil, nil},
		{"empty string", "emptyStringValue", "", ""},
		{"zero int", "intValueZero", 0, 0},
		{"map", "mapValue", map[string]any{"name": "John", "age": 30}, map[string]any{"name": "John", "age": 30}},
		{"bool false", "boolValueFalse", false, false},
		{"zero float", "floatValueZero", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			ctx.Set(tt.key, tt.val)

			got, err := ctx.Get(tt.key)
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.key, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestContextNestedKeys(t *testing.T) {
	tests := []struct {
		name   string
		setKey string
		val    any
		getKey string
		want   any
	}{
		{
			name:   "nested string",
			setKey: "user.name", val: "John Doe",
			getKey: "user.name", want: "John Doe",
		},
		{
			name:   "nested integer",
			setKey: "user.age", val: 30,
			getKey: "user.age", want: 30,
		},
		{
			name:   "deeply nested boolean",
			setKey: "user.details.employment.isEmployed", val: true,
			getKey: "user.details.employment.isEmployed", want: true,
		},
		{
			name:   "nested slice",
			setKey: "user.favorites.colors", val: []string{"blue", "green"},
			getKey: "user.favorites.colors", want: []string{"blue", "green"},
		},
		{
			name:   "overwrite nested value",
			setKey: "user.name", val: "Jane Doe",
			getKey: "user.name", want: "Jane Doe",
		},
		{
			name:   "nested map",
			setKey: "user.address",
			val:    map[string]any{"city": "Metropolis", "zip": "12345"},
			getKey: "user.address",
			want:   map[string]any{"city": "Metropolis", "zip": "12345"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			ctx.Set(tt.setKey, tt.val)

			got, err := ctx.Get(tt.getKey)
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.getKey, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get(%q) = %v, want %v", tt.getKey, got, tt.want)
			}
		})
	}
}

func TestContextDeepNestedKeys(t *testing.T) {
	tests := []struct {
		name   string
		setKey string
		val    any
		getKey string
		want   any
	}{
		{
			name: "deep nested string", setKey: "user.profile.bio",
			val: "Software Developer", getKey: "user.profile.bio",
			want: "Software Developer",
		},
		{
			name: "deep nested integer", setKey: "user.profile.experience",
			val: 10, getKey: "user.profile.experience", want: 10,
		},
		{
			name: "overwrite deep nested", setKey: "user.profile.bio",
			val: "Senior Software Developer", getKey: "user.profile.bio",
			want: "Senior Software Developer",
		},
		{
			name: "deep nested slice", setKey: "user.interests",
			val:    []string{"Coding", "Music", "Gaming"},
			getKey: "user.interests",
			want:   []string{"Coding", "Music", "Gaming"},
		},
		{
			name: "deep nested map", setKey: "user.socialMedia",
			val:    map[string]string{"Twitter": "@johndoe", "GitHub": "johndoe"},
			getKey: "user.socialMedia",
			want:   map[string]string{"Twitter": "@johndoe", "GitHub": "johndoe"},
		},
	}

	ctx := NewContext()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx.Set(tt.setKey, tt.val)

			got, err := ctx.Get(tt.getKey)
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.getKey, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get(%q) = %v, want %v", tt.getKey, got, tt.want)
			}
		})
	}
}

func TestContextSliceIndexAccess(t *testing.T) {
	tests := []struct {
		name      string
		setKey    string
		val       any
		getKey    string
		want      any
		wantError bool
	}{
		{
			name: "first element", setKey: "tasks",
			val:    []string{"Code Review", "Write Documentation", "Update Dependencies"},
			getKey: "tasks.0", want: "Code Review",
		},
		{
			name: "second element", setKey: "tasks",
			val:    []string{"Code Review", "Write Documentation", "Update Dependencies"},
			getKey: "tasks.1", want: "Write Documentation",
		},
		{
			name: "out of bounds", setKey: "tasks",
			val:       []string{"Code Review", "Write Documentation", "Update Dependencies"},
			getKey:    "tasks.3",
			wantError: true,
		},
		{
			name: "int slice element", setKey: "numbers",
			val: []int{1, 2, 3}, getKey: "numbers.2", want: 3,
		},
		{
			name: "bool slice element", setKey: "flags",
			val: []bool{true, false, true}, getKey: "flags.1", want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			ctx.Set(tt.setKey, tt.val)

			got, err := ctx.Get(tt.getKey)
			if tt.wantError {
				if err == nil {
					t.Errorf("Get(%q) = %v, nil, want error", tt.getKey, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.getKey, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get(%q) = %v, want %v", tt.getKey, got, tt.want)
			}
		})
	}
}

func TestContextKeyNotFound(t *testing.T) {
	tests := []struct {
		name   string
		setup  map[string]any
		getKey string
	}{
		{
			name:   "missing sibling key",
			setup:  map[string]any{"user.name": "John Doe"},
			getKey: "user.age",
		},
		{
			name:   "missing second-level key",
			setup:  map[string]any{"user.details.location": "City"},
			getKey: "user.details.age",
		},
		{
			name:   "missing deeply nested key",
			setup:  map[string]any{"user.profile.education.primary": "School Name"},
			getKey: "user.profile.education.highSchool",
		},
		{
			name:   "completely non-existent path",
			setup:  map[string]any{"existing.key": "value"},
			getKey: "completely.non.existent.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			for k, v := range tt.setup {
				ctx.Set(k, v)
			}

			_, err := ctx.Get(tt.getKey)
			if !errors.Is(err, ErrContextKeyNotFound) {
				t.Errorf("Get(%q) error = %v, want %v", tt.getKey, err, ErrContextKeyNotFound)
			}
		})
	}
}

func TestContextIndexOutOfRange(t *testing.T) {
	tests := []struct {
		name   string
		setup  map[string]any
		getKey string
	}{
		{
			name:   "slice index out of range",
			setup:  map[string]any{"user.hobbies": []string{"reading", "swimming"}},
			getKey: "user.hobbies.2",
		},
		{
			name: "nested array index out of range",
			setup: map[string]any{
				"team.members": []any{
					map[string]any{"name": "John", "skills": []string{"C++", "Go"}},
					map[string]any{"name": "Jane", "skills": []string{"JavaScript", "Python"}},
				},
			},
			getKey: "team.members.1.skills.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			for k, v := range tt.setup {
				ctx.Set(k, v)
			}

			_, err := ctx.Get(tt.getKey)
			if !errors.Is(err, ErrContextIndexOutOfRange) {
				t.Errorf("Get(%q) error = %v, want %v", tt.getKey, err, ErrContextIndexOutOfRange)
			}
		})
	}
}

func TestContextOverwrite(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		initial  any
		override any
	}{
		{"string", "simpleString", "initial", "overwrite"},
		{"integer", "integerValue", 123, 456},
		{"boolean", "booleanValue", false, true},
		{"nested key", "user.profile.age", 25, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			ctx.Set(tt.key, tt.initial)
			ctx.Set(tt.key, tt.override)

			got, err := ctx.Get(tt.key)
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.key, err)
			}
			if !reflect.DeepEqual(got, tt.override) {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.override)
			}
		})
	}
}

func TestContextStructConversion(t *testing.T) {
	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		ZipCode string `json:"zip_code"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Active  bool    `json:"is_active"`
		Address Address `json:"address"`
	}

	type PersonWithTags struct {
		Name    string `json:"full_name"`
		Age     int    `json:"age"`
		Active  bool   `json:"is_active"`
		Private string `json:"-"`
	}

	tests := []struct {
		name      string
		key       string
		val       any
		getKey    string
		want      any
		wantError error
	}{
		{
			name: "basic struct field",
			key:  "person",
			val: Person{
				Name: "John Doe", Age: 30, Active: true,
				Address: Address{Street: "123 Main St", City: "Metropolis", ZipCode: "12345"},
			},
			getKey: "person.name", want: "John Doe",
		},
		{
			name: "nested struct field",
			key:  "person",
			val: Person{
				Name: "Jane Smith", Age: 28, Active: false,
				Address: Address{Street: "456 Oak Ave", City: "Gotham", ZipCode: "54321"},
			},
			getKey: "person.address.city", want: "Gotham",
		},
		{
			name: "json tag rename",
			key:  "tagged",
			val: PersonWithTags{
				Name: "Alice Johnson", Age: 35, Active: true, Private: "secret",
			},
			getKey: "tagged.full_name", want: "Alice Johnson",
		},
		{
			name: "excluded json field",
			key:  "tagged",
			val: PersonWithTags{
				Name: "Bob Brown", Age: 42, Active: false, Private: "secret",
			},
			getKey:    "tagged.Private",
			wantError: ErrContextKeyNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			ctx.Set(tt.key, tt.val)

			got, err := ctx.Get(tt.getKey)
			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Errorf("Get(%q) error = %v, want %v", tt.getKey, err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.getKey, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get(%q) = %v, want %v", tt.getKey, got, tt.want)
			}
		})
	}
}

func TestContextComplexStruct(t *testing.T) {
	type Metadata struct {
		Tags     []string `json:"tags"`
		Priority int      `json:"priority"`
	}

	type Project struct {
		ID       string    `json:"id"`
		Metadata Metadata  `json:"metadata"`
		Owner    *string   `json:"owner"`
		Created  time.Time `json:"created"`
	}

	owner := "Project Owner"
	created := time.Date(2023, 4, 15, 10, 30, 0, 0, time.UTC)

	project := Project{
		ID:       "project-123",
		Metadata: Metadata{Tags: []string{"important", "urgent"}, Priority: 1},
		Owner:    &owner,
		Created:  created,
	}

	tests := []struct {
		name   string
		getKey string
		want   any
	}{
		{"embedded struct field", "project.metadata.tags.0", "important"},
		{"pointer value", "project.owner", "Project Owner"},
		{"time field", "project.created", created.Format(time.RFC3339)},
		{"direct field", "project.id", "project-123"},
		{"embedded priority", "project.metadata.priority", 1},
	}

	ctx := NewContext()
	ctx.Set("project", project)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := ctx.Get(tt.getKey)
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.getKey, err)
			}

			// Normalize: dereference pointers, format time as RFC3339.
			got := raw
			if v := reflect.ValueOf(raw); v.Kind() == reflect.Ptr && !v.IsNil() {
				got = v.Elem().Interface()
			}
			if ts, ok := got.(time.Time); ok {
				got = ts.Format(time.RFC3339)
			}

			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("Get(%q) = %v (%T), want %v (%T)", tt.getKey, raw, raw, tt.want, tt.want)
			}
		})
	}
}

func TestContextSliceOfStructs(t *testing.T) {
	type Item struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	items := []Item{
		{ID: 1, Name: "Item 1", Price: 10.99},
		{ID: 2, Name: "Item 2", Price: 24.99},
		{ID: 3, Name: "Item 3", Price: 5.99},
	}

	tests := []struct {
		name   string
		getKey string
		want   any
	}{
		{"first item name", "items.0.name", "Item 1"},
		{"second item price", "items.1.price", 24.99},
		{"last item ID", "items.2.id", 3},
	}

	ctx := NewContext()
	ctx.Set("items", items)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ctx.Get(tt.getKey)
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.getKey, err)
			}
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("Get(%q) = %v (%T), want %v (%T)", tt.getKey, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestContextMapWithStructValues(t *testing.T) {
	type UserProfile struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		IsAdmin  bool   `json:"is_admin"`
	}

	users := map[string]UserProfile{
		"user1": {Username: "johndoe", Email: "john@example.com"},
		"user2": {Username: "adminuser", Email: "admin@example.com", IsAdmin: true},
	}

	tests := []struct {
		name      string
		getKey    string
		want      any
		wantError error
	}{
		{"user1 username", "users.user1.username", "johndoe", nil},
		{"user2 admin status", "users.user2.is_admin", true, nil},
		{"non-existent user", "users.user3.username", nil, ErrContextKeyNotFound},
	}

	ctx := NewContext()
	ctx.Set("users", users)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ctx.Get(tt.getKey)
			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Errorf("Get(%q) error = %v, want %v", tt.getKey, err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.getKey, err)
			}
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("Get(%q) = %v (%T), want %v (%T)", tt.getKey, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestContextComplexNestedStructures(t *testing.T) {
	type Contact struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}

	type Address struct {
		Street     string `json:"street"`
		City       string `json:"city"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
		IsDefault  bool   `json:"isDefault"`
	}

	type Product struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Price       float64           `json:"price"`
		Description string            `json:"description"`
		Categories  []string          `json:"categories"`
		Tags        map[string]string `json:"tags"`
		Inventory   map[string]int    `json:"inventory"`
	}

	type Review struct {
		UserID    string    `json:"userId"`
		Rating    int       `json:"rating"`
		Comment   string    `json:"comment"`
		Date      time.Time `json:"date"`
		Helpful   int       `json:"helpful"`
		Responses []string  `json:"responses"`
	}

	type OrderItem struct {
		ProductID string  `json:"productId"`
		Quantity  int     `json:"quantity"`
		UnitPrice float64 `json:"unitPrice"`
		Discount  float64 `json:"discount"`
	}

	type Order struct {
		ID            string            `json:"id"`
		CustomerID    string            `json:"customerId"`
		Items         []OrderItem       `json:"items"`
		Total         float64           `json:"total"`
		ShippingInfo  map[string]string `json:"shippingInfo"`
		PaymentMethod map[string]any    `json:"paymentMethod"`
		Status        string            `json:"status"`
		CreatedAt     time.Time         `json:"createdAt"`
	}

	type Customer struct {
		ID             string         `json:"id"`
		Name           string         `json:"name"`
		Age            int            `json:"age"`
		Email          string         `json:"email"`
		Addresses      []Address      `json:"addresses"`
		Contacts       []Contact      `json:"contacts"`
		PreferredItems []string       `json:"preferredItems"`
		Orders         []Order        `json:"orders"`
		AccountBalance float64        `json:"accountBalance"`
		Metadata       map[string]any `json:"metadata"`
		IsVerified     bool           `json:"isVerified"`
		JoinDate       time.Time      `json:"joinDate"`
		LastLogin      time.Time      `json:"lastLogin"`
	}

	type Department struct {
		Name     string         `json:"name"`
		Manager  string         `json:"manager"`
		Budget   float64        `json:"budget"`
		Projects []string       `json:"projects"`
		Staff    map[string]any `json:"staff"`
	}

	type Company struct {
		Name         string                `json:"name"`
		Founded      time.Time             `json:"founded"`
		Departments  map[string]Department `json:"departments"`
		Customers    map[string]Customer   `json:"customers"`
		Products     []Product             `json:"products"`
		Reviews      map[string][]Review   `json:"reviews"`
		Headquarters Address               `json:"headquarters"`
		Branches     []Address             `json:"branches"`
		Revenue      map[string]float64    `json:"revenue"`
		Employees    int                   `json:"employees"`
		Partners     []string              `json:"partners"`
		Settings     map[string]any        `json:"settings"`
	}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)
	lastMonth := now.Add(-30 * 24 * time.Hour)

	company := Company{
		Name:    "Acme Corporation",
		Founded: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		Departments: map[string]Department{
			"engineering": {
				Name: "Engineering", Manager: "Jane Smith",
				Budget:   1000000.50,
				Projects: []string{"Project Alpha", "Project Beta", "Project Gamma"},
				Staff: map[string]any{
					"senior": []string{"Alice", "Bob", "Charlie"},
					"junior": []string{"Dave", "Eve", "Frank"},
					"counts": map[string]int{
						"developers": 20, "testers": 10, "managers": 5,
					},
				},
			},
			"sales": {
				Name: "Sales", Manager: "John Doe",
				Budget:   800000.75,
				Projects: []string{"North Region Campaign", "South Region Campaign"},
				Staff: map[string]any{
					"senior": []string{"Grace", "Heidi"},
					"junior": []string{"Ivan", "Judy"},
					"counts": map[string]int{
						"representatives": 15, "managers": 3,
					},
				},
			},
		},
		Customers: map[string]Customer{
			"cust1": {
				ID: "C-1001", Name: "XYZ Corp",
				Email: "contact@xyzcorp.com",
				Addresses: []Address{
					{Street: "123 Business Ave", City: "Commerce City",
						PostalCode: "10001", Country: "USA", IsDefault: true},
					{Street: "456 Enterprise Blvd", City: "Commerce City",
						PostalCode: "10001", Country: "USA"},
				},
				Contacts: []Contact{
					{Type: "phone", Value: "555-1234"},
					{Type: "email", Value: "info@xyzcorp.com"},
				},
				PreferredItems: []string{"product1", "product3"},
				Orders: []Order{{
					ID: "O-5001", CustomerID: "C-1001",
					Items: []OrderItem{
						{ProductID: "P-101", Quantity: 5, UnitPrice: 29.99, Discount: 0.1},
						{ProductID: "P-105", Quantity: 2, UnitPrice: 99.99},
					},
					Total: 249.93,
					ShippingInfo: map[string]string{
						"method": "express", "carrier": "FedEx", "tracking": "FX123456789",
					},
					PaymentMethod: map[string]any{
						"type": "credit_card", "last4": "1234", "expiry": "05/25",
					},
					Status: "delivered", CreatedAt: lastMonth,
				}},
				AccountBalance: 5000.00,
				Metadata: map[string]any{
					"sector": "Technology", "size": "Medium", "established": 2005,
					"contacts": map[string]string{
						"primary": "John Smith", "secondary": "Jane Doe",
					},
				},
				IsVerified: true,
				JoinDate:   lastYear(now),
				LastLogin:  yesterday,
			},
			"cust2": {
				ID: "C-1002", Name: "John Consumer", Age: 35,
				Email: "john@example.com",
				Addresses: []Address{
					{Street: "789 Residential St", City: "Hometown",
						PostalCode: "20002", Country: "USA", IsDefault: true},
				},
				Contacts: []Contact{
					{Type: "phone", Value: "555-5678"},
					{Type: "email", Value: "john@example.com"},
				},
				PreferredItems: []string{"product2", "product4"},
				Orders: []Order{{
					ID: "O-5002", CustomerID: "C-1002",
					Items: []OrderItem{
						{ProductID: "P-102", Quantity: 1, UnitPrice: 59.99},
					},
					Total: 59.99,
					ShippingInfo: map[string]string{
						"method": "standard", "carrier": "UPS", "tracking": "UPS987654321",
					},
					PaymentMethod: map[string]any{
						"type": "paypal", "email": "john@example.com",
					},
					Status: "processing", CreatedAt: yesterday,
				}},
				AccountBalance: 150.50,
				Metadata: map[string]any{
					"preferences": map[string]any{
						"notifications": true, "theme": "dark",
					},
					"deviceInfo": map[string]string{
						"browser": "Chrome", "os": "Windows",
					},
				},
				IsVerified: true,
				JoinDate:   lastMonth,
				LastLogin:  now,
			},
		},
		Products: []Product{
			{
				ID: "P-101", Name: "Enterprise Software",
				Price: 299.99, Description: "Business solution software",
				Categories: []string{"software", "business", "enterprise"},
				Tags: map[string]string{
					"level": "premium", "subscription": "yearly", "support": "24/7",
				},
				Inventory: map[string]int{"licenses": 500, "physical": 0},
			},
			{
				ID: "P-102", Name: "Office Chair",
				Price: 199.99, Description: "Ergonomic office chair",
				Categories: []string{"furniture", "office", "ergonomic"},
				Tags: map[string]string{
					"material": "leather", "color": "black", "warranty": "2-year",
				},
				Inventory: map[string]int{
					"warehouse_a": 120, "warehouse_b": 85, "display": 5,
				},
			},
		},
		Reviews: map[string][]Review{
			"P-101": {
				{
					UserID: "C-1001", Rating: 5,
					Comment: "Excellent software for our business needs",
					Date:    lastWeek, Helpful: 12,
					Responses: []string{
						"Thank you for your feedback!",
						"We appreciate your business.",
					},
				},
				{
					UserID: "C-1002", Rating: 4,
					Comment: "Good software but a bit expensive",
					Date:    lastMonth, Helpful: 8,
					Responses: []string{"Thanks for your honest review."},
				},
			},
			"P-102": {{
				UserID: "C-1002", Rating: 5,
				Comment: "Very comfortable chair, worth every penny",
				Date:    lastWeek, Helpful: 15,
				Responses: []string{"We're glad you enjoy our product!"},
			}},
		},
		Headquarters: Address{
			Street: "1 Corporate Plaza", City: "Business City",
			PostalCode: "10005", Country: "USA", IsDefault: true,
		},
		Branches: []Address{
			{Street: "25 East Business St", City: "Eastern City",
				PostalCode: "20025", Country: "USA"},
			{Street: "50 West Commerce Rd", City: "Western City",
				PostalCode: "30050", Country: "USA"},
		},
		Revenue: map[string]float64{
			"2021": 5000000.00, "2022": 6250000.00, "2023": 7500000.00,
		},
		Employees: 250,
		Partners:  []string{"Partner A", "Partner B", "Partner C"},
		Settings: map[string]any{
			"notifications": map[string]bool{
				"email": true, "sms": false, "desktop": true,
			},
			"security": map[string]any{
				"mfa_required": true,
				"password_policy": map[string]any{
					"min_length": 12, "require_special": true, "expiry_days": 90,
				},
			},
			"display": map[string]string{"logo": "logo.png", "theme": "corporate"},
		},
	}

	tests := []struct {
		name   string
		getKey string
		want   any
	}{
		{"company name", "company.name", "Acme Corporation"},
		{"headquarters city", "company.headquarters.city", "Business City"},
		{"2022 revenue", "company.revenue.2022", 6250000.00},
		{"first branch postal code", "company.branches.0.postalCode", "20025"},
		{"engineering budget", "company.departments.engineering.budget", 1000000.50},
		{"second engineering project", "company.departments.engineering.projects.1", "Project Beta"},
		{"engineering senior staff", "company.departments.engineering.staff.senior.1", "Bob"},
		{"engineering testers count", "company.departments.engineering.staff.counts.testers", 10},
		{"first customer name", "company.customers.cust1.name", "XYZ Corp"},
		{"first customer second address city", "company.customers.cust1.addresses.1.city", "Commerce City"},
		{"first customer first contact type", "company.customers.cust1.contacts.0.type", "phone"},
		{"first customer order second item quantity", "company.customers.cust1.orders.0.items.1.quantity", 2},
		{"first customer primary contact", "company.customers.cust1.metadata.contacts.primary", "John Smith"},
		{"second customer device OS", "company.customers.cust2.metadata.deviceInfo.os", "Windows"},
		{"first product price", "company.products.0.price", 299.99},
		{"first product first category", "company.products.0.categories.0", "software"},
		{"second product warranty tag", "company.products.1.tags.warranty", "2-year"},
		{"P-101 first review helpful", "company.reviews.P-101.0.helpful", 12},
		{"P-102 first review response", "company.reviews.P-102.0.responses.0", "We're glad you enjoy our product!"},
		{"security password min length", "company.settings.security.password_policy.min_length", 12},
		{"email notification setting", "company.settings.notifications.email", true},
	}

	ctx := NewContext()
	ctx.Set("company", company)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ctx.Get(tt.getKey)
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.getKey, err)
			}
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("Get(%q) = %v (%T), want %v (%T)", tt.getKey, got, got, tt.want, tt.want)
			}
		})
	}
}

func lastYear(t time.Time) time.Time {
	return t.AddDate(-1, 0, 0)
}

func TestContextPreservesOriginalTypes(t *testing.T) {
	type User struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	type Corp struct {
		Name      string          `json:"name"`
		Employees []User          `json:"employees"`
		Settings  map[string]bool `json:"settings"`
	}

	tests := []struct {
		name   string
		key    string
		val    any
		getKey string
		want   any
	}{
		{
			name: "struct directly", key: "user",
			val:    User{Name: "John", Age: 30, Email: "john@example.com"},
			getKey: "user.name", want: "John",
		},
		{
			name: "slice of structs", key: "users",
			val:    []User{{Name: "Alice", Age: 25}, {Name: "Bob", Age: 35}},
			getKey: "users.1.name", want: "Bob",
		},
		{
			name: "complex nested structure", key: "company",
			val: Corp{
				Name: "TechCorp",
				Employees: []User{
					{Name: "Charlie", Age: 28, Email: "charlie@techcorp.com"},
					{Name: "Diana", Age: 32, Email: "diana@techcorp.com"},
				},
				Settings: map[string]bool{"remote": true, "flexible": false},
			},
			getKey: "company.employees.0.email", want: "charlie@techcorp.com",
		},
		{
			name: "map directly", key: "config",
			val:    map[string]any{"debug": true, "port": 8080},
			getKey: "config.port", want: 8080,
		},
		{
			name: "nested with preserved types", key: "nested.data",
			val:    []map[string]string{{"type": "test", "value": "data"}},
			getKey: "nested.data.0.type", want: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			ctx.Set(tt.key, tt.val)

			got, err := ctx.Get(tt.getKey)
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", tt.getKey, err)
			}
			if got != tt.want {
				t.Errorf("Get(%q) = %v, want %v", tt.getKey, got, tt.want)
			}

			// Verify top-level type is preserved.
			topKey := strings.Split(tt.key, ".")[0]
			top, err := ctx.Get(topKey)
			if err != nil {
				t.Fatalf("Get(%q) = _, %v, want nil error", topKey, err)
			}

			switch tt.val.(type) {
			case User:
				if _, ok := top.(User); !ok {
					t.Errorf("top-level type = %T, want User", top)
				}
			case []User:
				if _, ok := top.([]User); !ok {
					t.Errorf("top-level type = %T, want []User", top)
				}
			case Corp:
				if _, ok := top.(Corp); !ok {
					t.Errorf("top-level type = %T, want Corp", top)
				}
			}
		})
	}
}

func TestRenderingStability(t *testing.T) {
	const iterations = 10

	tests := []struct {
		name     string
		template string
		setup    func() Context
	}{
		{
			name:     "map string iteration",
			template: `{% for key, value in data %}{{ key }}:{{ value }},{% endfor %}`,
			setup: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]string{
					"zebra": "last", "alpha": "first", "beta": "second",
					"gamma": "third", "delta": "fourth", "echo": "fifth",
					"foxtrot": "sixth",
				})
				return ctx
			},
		},
		{
			name:     "map interface iteration",
			template: `{% for key, user in users %}{{ key }}:{{ user.name }},{% endfor %}`,
			setup: func() Context {
				ctx := NewContext()
				ctx.Set("users", map[string]any{
					"user3": map[string]any{"name": "Charlie", "age": 35},
					"user1": map[string]any{"name": "Alice", "age": 25},
					"user4": map[string]any{"name": "David", "age": 40},
					"user2": map[string]any{"name": "Bob", "age": 30},
				})
				return ctx
			},
		},
		{
			name:     "nested map iteration",
			template: `{% for deptKey, dept in company %}{{ deptKey }}:{% for empKey, emp in dept %}{{ empKey }}-{{ emp }},{% endfor %};{% endfor %}`,
			setup: func() Context {
				ctx := NewContext()
				ctx.Set("company", map[string]map[string]string{
					"engineering": {"john": "senior", "alice": "junior", "bob": "lead"},
					"sales":       {"mary": "manager", "tom": "rep"},
					"hr":          {"susan": "director"},
				})
				return ctx
			},
		},
		{
			name:     "slice rendering",
			template: `{% for item in items %}{{ item }},{% endfor %}`,
			setup: func() Context {
				ctx := NewContext()
				ctx.Set("items", []string{"first", "second", "third", "fourth", "fifth"})
				return ctx
			},
		},
		{
			name:     "struct slice rendering",
			template: `{% for person in people %}{{ person.name }}-{{ person.age }},{% endfor %}`,
			setup: func() Context {
				type Person struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}
				ctx := NewContext()
				ctx.Set("people", []Person{
					{Name: "Alice", Age: 25},
					{Name: "Bob", Age: 30},
					{Name: "Charlie", Age: 35},
				})
				return ctx
			},
		},
		{
			name:     "map variable rendering",
			template: `{{ data }}`,
			setup: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]any{
					"zebra": "value1", "alpha": "value2",
					"beta": "value3", "gamma": "value4",
				})
				return ctx
			},
		},
		{
			name:     "complex nested structure rendering",
			template: `{{ complex }}`,
			setup: func() Context {
				ctx := NewContext()
				ctx.Set("complex", map[string]any{
					"users": map[string]any{
						"user2": []string{"read", "write"},
						"user1": []string{"admin", "read", "write"},
					},
					"settings": map[string]any{"theme": "dark", "lang": "en"},
					"data": []map[string]string{
						{"name": "item1", "type": "A"},
						{"name": "item2", "type": "B"},
					},
				})
				return ctx
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Compile(tt.template)
			if err != nil {
				t.Fatalf("Compile(%q) = _, %v", tt.template, err)
			}

			var results []string
			for i := range iterations {
				ctx := tt.setup()
				result, err := tmpl.Render(map[string]any(ctx))
				if err != nil {
					t.Fatalf("Render() iteration %d = _, %v", i, err)
				}
				results = append(results, result)
			}

			first := results[0]
			for i := 1; i < iterations; i++ {
				if results[i] != first {
					t.Errorf("inconsistent output:\n  iteration 0: %q\n  iteration %d: %q",
						first, i, results[i])
					break
				}
			}
		})
	}
}
