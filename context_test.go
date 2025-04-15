package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestEmptyContextInitialization(t *testing.T) {
	context := NewContext()
	if len(context) != 0 {
		t.Errorf("NewContext() should create an empty context, got %v", context)
	}
}
func TestAddingAndRetrievingDiverseTypesWithRefactoring(t *testing.T) {
	testCases := []struct {
		description string
		key         string
		value       interface{}
		expected    interface{}
	}{
		{"Retrieve string value", "stringValue", "hello world", "hello world"},
		{"Retrieve integer value", "intValue", 28, 28},
		{"Retrieve boolean true value", "boolValueTrue", true, true},
		{"Retrieve float value", "floatValue", 1.75, 1.75},
		{"Retrieve slice of strings", "sliceOfString", []string{"Go", "Python", "JavaScript"}, []string{"Go", "Python", "JavaScript"}},
		{"Retrieve slice of integers", "sliceOfInt", []int{1, 2, 3}, []int{1, 2, 3}},
		{"Retrieve multi-dimensional slice", "multiDimSlice", [][]int{{1, 2}, {3, 4}}, [][]int{{1, 2}, {3, 4}}},
		{"Retrieve nil value", "nilValue", nil, nil},
		{"Retrieve empty string value", "emptyStringValue", "", ""},
		{"Retrieve zero integer value", "intValueZero", 0, 0},
		{"Retrieve map value", "mapValue", map[string]interface{}{"name": "John", "age": 30}, map[string]interface{}{"name": "John", "age": 30}},
		{"Retrieve boolean false value", "boolValueFalse", false, false},
		{"Retrieve zero float value", "floatValueZero", 0.0, 0.0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			context := NewContext() // Initialize context for each case to avoid data pollution
			context.Set(testCase.key, testCase.value)

			value, err := context.Get(testCase.key)
			if err != nil {
				t.Errorf("Unexpected error for '%s': %v", testCase.key, err)
				return
			}

			// Reflect is used for deep equality checks, particularly useful for slices and maps
			if !reflect.DeepEqual(value, testCase.expected) {
				t.Errorf("Get('%s') did not return the expected value. Got %v, expected %v", testCase.key, value, testCase.expected)
			}
		})
	}
}
func TestRetrievingValuesWithNestedKeys(t *testing.T) {
	testCases := []struct {
		description   string
		key           string
		value         interface{}
		retrieveKey   string
		expectedValue interface{}
	}{
		{
			description:   "Retrieve nested string",
			key:           "user.name",
			value:         "John Doe",
			retrieveKey:   "user.name",
			expectedValue: "John Doe",
		},
		{
			description:   "Retrieve nested integer",
			key:           "user.age",
			value:         30,
			retrieveKey:   "user.age",
			expectedValue: 30,
		},
		{
			description:   "Retrieve deeply nested boolean",
			key:           "user.details.employment.isEmployed",
			value:         true,
			retrieveKey:   "user.details.employment.isEmployed",
			expectedValue: true,
		},
		{
			description:   "Retrieve nested slice",
			key:           "user.favorites.colors",
			value:         []string{"blue", "green"},
			retrieveKey:   "user.favorites.colors",
			expectedValue: []string{"blue", "green"},
		},
		{
			description:   "Overwrite and retrieve nested value",
			key:           "user.name",
			value:         "Jane Doe",
			retrieveKey:   "user.name",
			expectedValue: "Jane Doe",
		},
		{
			description:   "Retrieve nested map",
			key:           "user.address",
			value:         map[string]interface{}{"city": "Metropolis", "zip": "12345"},
			retrieveKey:   "user.address",
			expectedValue: map[string]interface{}{"city": "Metropolis", "zip": "12345"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			context := NewContext() // Initialize context for each case
			context.Set(testCase.key, testCase.value)

			value, _ := context.Get(testCase.retrieveKey) // Error handling omitted per instructions

			// Using reflect.DeepEqual for complex type comparison
			if !reflect.DeepEqual(value, testCase.expectedValue) {
				t.Errorf("%s: expected %v, got %v", testCase.description, testCase.expectedValue, value)
			}
		})
	}
}
func TestRetrievingValuesWithDeepNestedKeys(t *testing.T) {
	testCases := []struct {
		description   string
		addKey        string
		addValue      interface{}
		retrieveKey   string
		expectedValue interface{}
	}{
		{
			description:   "Retrieve deep nested string",
			addKey:        "user.profile.bio",
			addValue:      "Software Developer",
			retrieveKey:   "user.profile.bio",
			expectedValue: "Software Developer",
		},
		{
			description:   "Retrieve deep nested integer",
			addKey:        "user.profile.experience",
			addValue:      10,
			retrieveKey:   "user.profile.experience",
			expectedValue: 10,
		},
		{
			description:   "Overwrite and retrieve deep nested value",
			addKey:        "user.profile.bio",
			addValue:      "Senior Software Developer",
			retrieveKey:   "user.profile.bio",
			expectedValue: "Senior Software Developer",
		},
		{
			description:   "Retrieve deep nested slice",
			addKey:        "user.interests",
			addValue:      []string{"Coding", "Music", "Gaming"},
			retrieveKey:   "user.interests",
			expectedValue: []string{"Coding", "Music", "Gaming"},
		},
		{
			description:   "Retrieve deep nested map",
			addKey:        "user.socialMedia",
			addValue:      map[string]string{"Twitter": "@johndoe", "GitHub": "johndoe"},
			retrieveKey:   "user.socialMedia",
			expectedValue: map[string]string{"Twitter": "@johndoe", "GitHub": "johndoe"},
		},
	}

	context := NewContext()
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			context.Set(testCase.addKey, testCase.addValue)

			value, _ := context.Get(testCase.retrieveKey)

			// Using reflect.DeepEqual for accurate comparison of complex types
			if !reflect.DeepEqual(value, testCase.expectedValue) {
				t.Errorf("%s: expected %v, got %v", testCase.description, testCase.expectedValue, value)
			}
		})
	}
}
func TestRetrievingSliceElementsWithIndices(t *testing.T) {
	testCases := []struct {
		description   string
		sliceKey      string
		sliceValue    interface{}
		retrieveKey   string
		expectedValue interface{}
		expectedError bool
	}{
		{
			description:   "Retrieve first task from string slice",
			sliceKey:      "tasks",
			sliceValue:    []string{"Code Review", "Write Documentation", "Update Dependencies"},
			retrieveKey:   "tasks.0",
			expectedValue: "Code Review",
			expectedError: false,
		},
		{
			description:   "Retrieve second task from string slice",
			sliceKey:      "tasks",
			sliceValue:    []string{"Code Review", "Write Documentation", "Update Dependencies"},
			retrieveKey:   "tasks.1",
			expectedValue: "Write Documentation",
			expectedError: false,
		},
		{
			description:   "Attempt to retrieve non-existing index in tasks slice",
			sliceKey:      "tasks",
			sliceValue:    []string{"Code Review", "Write Documentation", "Update Dependencies"},
			retrieveKey:   "tasks.3",
			expectedValue: nil,
			expectedError: true,
		},
		{
			description:   "Retrieve integer from slice of integers",
			sliceKey:      "numbers",
			sliceValue:    []int{1, 2, 3},
			retrieveKey:   "numbers.2",
			expectedValue: 3,
			expectedError: false,
		},
		{
			description:   "Retrieve boolean from slice of booleans",
			sliceKey:      "flags",
			sliceValue:    []bool{true, false, true},
			retrieveKey:   "flags.1",
			expectedValue: false,
			expectedError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			context := NewContext()
			context.Set(testCase.sliceKey, testCase.sliceValue)

			value, err := context.Get(testCase.retrieveKey)
			if testCase.expectedError {
				if err == nil {
					t.Errorf("%s: expected an error but got none", testCase.description)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", testCase.description, err)
				} else if !reflect.DeepEqual(value, testCase.expectedValue) {
					t.Errorf("%s: expected value %v, got %v", testCase.description, testCase.expectedValue, value)
				}
			}
		})
	}
}

func TestRetrievingValuesForNonExistentNestedKeys(t *testing.T) {
	testCases := []struct {
		description     string
		setupKeysValues map[string]interface{} // Key-value pairs to set up context
		nonExistentKey  string                 // Key to test for non-existence
	}{
		{
			description: "Non-existent top-level key",
			setupKeysValues: map[string]interface{}{
				"user.name": "John Doe",
			},
			nonExistentKey: "user.age",
		},
		{
			description: "Non-existent second-level key",
			setupKeysValues: map[string]interface{}{
				"user.details.location": "City",
			},
			nonExistentKey: "user.details.age",
		},
		{
			description: "Non-existent key in deeply nested structure",
			setupKeysValues: map[string]interface{}{
				"user.profile.education.primary": "School Name",
			},
			nonExistentKey: "user.profile.education.highSchool",
		},
		{
			description: "Completely non-existent nested key",
			setupKeysValues: map[string]interface{}{
				"existing.key": "value",
			},
			nonExistentKey: "completely.non.existent.key",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			context := NewContext() // Initialize context for each case

			// Setup context with predefined keys and values
			for key, value := range testCase.setupKeysValues {
				context.Set(key, value)
			}

			// Attempt to retrieve a non-existent key
			_, err := context.Get(testCase.nonExistentKey)

			// Verify that the correct error is returned
			if !errors.Is(err, ErrContextKeyNotFound) {
				t.Errorf("%s: expected ErrContextKeyNotFound for non-existent key '%s', got %v", testCase.description, testCase.nonExistentKey, err)
			}
		})
	}
}
func TestRetrievingValuesForIndexOutOfRange(t *testing.T) {
	testCases := []struct {
		description        string
		setupKeysValues    map[string]interface{} // Key-value pairs to set up context
		indexOutOfRangeKey string                 // Key to test for index out of range
	}{
		{
			description: "Index out of range in slice",
			setupKeysValues: map[string]interface{}{
				"user.hobbies": []string{"reading", "swimming"},
			},
			indexOutOfRangeKey: "user.hobbies.2",
		},
		{
			description: "Index out of range in nested array",
			setupKeysValues: map[string]interface{}{
				"team.members": []interface{}{
					map[string]interface{}{"name": "John", "skills": []string{"C++", "Go"}},
					map[string]interface{}{"name": "Jane", "skills": []string{"JavaScript", "Python"}},
				},
			},
			indexOutOfRangeKey: "team.members.1.skills.2",
		},
		// Add more test cases as necessary
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			context := NewContext() // Initialize context for each case

			// Setup context with predefined keys and values
			for key, value := range testCase.setupKeysValues {
				context.Set(key, value)
			}

			// Attempt to retrieve a value using a key that specifies an index out of range
			_, err := context.Get(testCase.indexOutOfRangeKey)

			// Verify that the correct error is returned
			if !errors.Is(err, ErrContextIndexOutOfRange) {
				t.Errorf("%s: expected ErrContextIndexOutOfRange for key '%s', got %v", testCase.description, testCase.indexOutOfRangeKey, err)
			}
		})
	}
}

func TestSimplifiedOverwritingValuesInContext(t *testing.T) {
	testCases := []struct {
		description    string
		key            string
		initialValue   interface{}
		overwriteValue interface{}
	}{
		{
			description:    "Overwrite string value",
			key:            "simpleString",
			initialValue:   "initial",
			overwriteValue: "overwrite",
		},
		{
			description:    "Overwrite integer value",
			key:            "integerValue",
			initialValue:   123,
			overwriteValue: 456,
		},
		{
			description:    "Overwrite boolean value",
			key:            "booleanValue",
			initialValue:   false,
			overwriteValue: true,
		},
		{
			description:    "Overwrite nested key value",
			key:            "user.profile.age",
			initialValue:   25,
			overwriteValue: 30,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			context := NewContext()                            // Initialize context for each test case
			context.Set(testCase.key, testCase.initialValue)   // Add initial value
			context.Set(testCase.key, testCase.overwriteValue) // Overwrite it

			// Retrieve to verify overwrite
			value, _ := context.Get(testCase.key)

			// Verify overwrite without explicitly stating expected value in test case
			if !reflect.DeepEqual(value, testCase.overwriteValue) {
				t.Errorf("%s: expected %v after overwrite, got %v", testCase.description, testCase.overwriteValue, value)
			}
		})
	}
}

// TestStructConversion tests the conversion of struct types to maps for template accessibility.
// It verifies that structs are properly converted to maps while maintaining their field values
// and that nested structs are recursively converted.
func TestStructConversion(t *testing.T) {
	// Define test structs
	type Address struct {
		Street  string
		City    string
		ZipCode string
	}

	type Person struct {
		Name    string
		Age     int
		Active  bool
		Address Address
	}

	type PersonWithTags struct {
		Name    string `json:"full_name"`
		Age     int    `json:"age"`
		Active  bool   `json:"is_active"`
		Private string `json:"-"` // Should be excluded from the map
	}

	// Test cases
	testCases := []struct {
		description string
		key         string
		value       interface{}
		checkPath   string
		expected    interface{}
	}{
		{
			description: "Basic struct conversion",
			key:         "person",
			value: Person{
				Name:   "John Doe",
				Age:    30,
				Active: true,
				Address: Address{
					Street:  "123 Main St",
					City:    "Metropolis",
					ZipCode: "12345",
				},
			},
			checkPath: "person.Name",
			expected:  "John Doe",
		},
		{
			description: "Nested struct access",
			key:         "person",
			value: Person{
				Name:   "Jane Smith",
				Age:    28,
				Active: false,
				Address: Address{
					Street:  "456 Oak Ave",
					City:    "Gotham",
					ZipCode: "54321",
				},
			},
			checkPath: "person.Address.City",
			expected:  "Gotham",
		},
		{
			description: "JSON tag respect",
			key:         "tagged_person",
			value: PersonWithTags{
				Name:    "Alice Johnson",
				Age:     35,
				Active:  true,
				Private: "should not be exposed",
			},
			checkPath: "tagged_person.full_name",
			expected:  "Alice Johnson",
		},
		{
			description: "Private field exclusion",
			key:         "tagged_person",
			value: PersonWithTags{
				Name:    "Bob Brown",
				Age:     42,
				Active:  false,
				Private: "should not be exposed",
			},
			checkPath: "tagged_person.Private",
			expected:  nil, // Should not be found
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			context := NewContext()
			context.Set(testCase.key, testCase.value)

			// Get the value using the check path
			value, err := context.Get(testCase.checkPath)

			// For the private field test, we expect an error
			if testCase.expected == nil {
				if err == nil || !errors.Is(err, ErrContextKeyNotFound) {
					t.Errorf("%s: expected ErrContextKeyNotFound for path '%s', got %v",
						testCase.description, testCase.checkPath, err)
				}
				return
			}

			// For all other tests, we don't expect an error
			if err != nil {
				t.Errorf("%s: unexpected error: %v", testCase.description, err)
				return
			}

			// Check if the value matches the expected value
			if !reflect.DeepEqual(value, testCase.expected) {
				t.Errorf("%s: expected %v, got %v",
					testCase.description, testCase.expected, value)
			}
		})
	}
}

// TestComplexStructConversion tests more complex struct conversion scenarios,
// including pointers, embedded structs, and time.Time fields.
func TestComplexStructConversion(t *testing.T) {
	// Define test structs with pointers and embedded types
	type Metadata struct {
		Tags     []string
		Priority int
	}

	type Project struct {
		ID       string
		Metadata // Embedded struct
		Owner    *string
		Created  time.Time
	}

	// Create test data
	ownerName := "Project Owner"
	creationTime := time.Date(2023, 4, 15, 10, 30, 0, 0, time.UTC)

	project := Project{
		ID: "project-123",
		Metadata: Metadata{
			Tags:     []string{"important", "urgent"},
			Priority: 1,
		},
		Owner:   &ownerName,
		Created: creationTime,
	}

	// Setup test cases
	testCases := []struct {
		description string
		checkPath   string
		expected    interface{}
		compareFunc func(interface{}, interface{}) bool
	}{
		{
			description: "Access embedded struct field",
			checkPath:   "project.Tags.0",
			expected:    "important",
			compareFunc: func(a, b interface{}) bool {
				return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
			},
		},
		{
			description: "Access pointer value",
			checkPath:   "project.Owner",
			expected:    "Project Owner",
			compareFunc: func(a, b interface{}) bool {
				return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
			},
		},
		{
			description: "Preserve time.Time",
			checkPath:   "project.Created",
			expected:    creationTime.Format(time.RFC3339),
			compareFunc: func(a, b interface{}) bool {
				return fmt.Sprintf("%v", a) == b
			},
		},
		{
			description: "Access ID directly",
			checkPath:   "project.ID",
			expected:    "project-123",
			compareFunc: func(a, b interface{}) bool {
				return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
			},
		},
		{
			description: "Access embedded field directly",
			checkPath:   "project.Priority",
			expected:    1,
			compareFunc: func(a, b interface{}) bool {
				aVal, aOk := toFloat64(a)
				bVal, bOk := toFloat64(b)
				if !aOk || !bOk {
					return false
				}
				return math.Abs(aVal-bVal) < 0.0001
			},
		},
	}

	// Prepare context with the test data
	context := NewContext()
	context.Set("project", project)

	// Run the tests
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			value, err := context.Get(testCase.checkPath)
			if err != nil {
				t.Errorf("%s: unexpected error: %v", testCase.description, err)
				return
			}

			// Use custom comparison function
			if testCase.compareFunc != nil {
				if !testCase.compareFunc(value, testCase.expected) {
					t.Errorf("%s: expected %v (%T), got %v (%T)",
						testCase.description, testCase.expected, testCase.expected, value, value)
				}
				return
			}

			// Default comparison
			if !reflect.DeepEqual(value, testCase.expected) {
				t.Errorf("%s: expected %v (%T), got %v (%T)",
					testCase.description, testCase.expected, testCase.expected, value, value)
			}
		})
	}
}

// toFloat64 attempts to convert any numeric type to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case json.Number:
		if f, err := val.Float64(); err == nil {
			return f, true
		}
		if i, err := val.Int64(); err == nil {
			return float64(i), true
		}
		return 0, false
	case string:
		// Try to parse number from string
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// TestSliceOfStructs tests how slices containing structs are processed.
func TestSliceOfStructs(t *testing.T) {
	// Define a test struct
	type Item struct {
		ID    int
		Name  string
		Price float64
	}

	// Create test data
	items := []Item{
		{ID: 1, Name: "Item 1", Price: 10.99},
		{ID: 2, Name: "Item 2", Price: 24.99},
		{ID: 3, Name: "Item 3", Price: 5.99},
	}

	// Setup test cases
	testCases := []struct {
		description string
		checkPath   string
		expected    interface{}
		compareFunc func(interface{}, interface{}) bool // Add custom comparison function
	}{
		{
			description: "Access first item name",
			checkPath:   "items.0.Name",
			expected:    "Item 1",
			compareFunc: nil, // Use default comparison
		},
		{
			description: "Access second item price",
			checkPath:   "items.1.Price",
			expected:    24.99,
			compareFunc: func(a, b interface{}) bool {
				// For floating point numbers, use approximate comparison
				aVal, aOk := toFloat64(a)
				bVal, bOk := toFloat64(b)
				if !aOk || !bOk {
					return false
				}
				return math.Abs(aVal-bVal) < 0.0001
			},
		},
		{
			description: "Access last item ID",
			checkPath:   "items.2.ID",
			expected:    3,
			compareFunc: func(a, b interface{}) bool {
				// For integers, convert then compare
				aVal, aOk := toFloat64(a)
				bVal, bOk := toFloat64(b)
				if !aOk || !bOk {
					return false
				}
				return math.Abs(aVal-bVal) < 0.0001
			},
		},
	}

	// Prepare context with the test data
	context := NewContext()
	context.Set("items", items)

	// Run the tests
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			value, err := context.Get(testCase.checkPath)
			if err != nil {
				t.Errorf("%s: unexpected error: %v", testCase.description, err)
				return
			}

			// Use custom comparison function or default comparison
			if testCase.compareFunc != nil {
				if !testCase.compareFunc(value, testCase.expected) {
					t.Errorf("%s: expected %v (%T), got %v (%T)",
						testCase.description, testCase.expected, testCase.expected, value, value)
				}
				return
			}

			// Default comparison
			if !reflect.DeepEqual(value, testCase.expected) {
				t.Errorf("%s: expected %v (%T), got %v (%T)",
					testCase.description, testCase.expected, testCase.expected, value, value)
			}
		})
	}
}

// TestMapWithStructValues tests maps containing struct values.
func TestMapWithStructValues(t *testing.T) {
	// Define a test struct
	type UserProfile struct {
		Username string
		Email    string
		IsAdmin  bool
	}

	// Create test data - a map of users
	users := map[string]UserProfile{
		"user1": {
			Username: "johndoe",
			Email:    "john@example.com",
			IsAdmin:  false,
		},
		"user2": {
			Username: "adminuser",
			Email:    "admin@example.com",
			IsAdmin:  true,
		},
	}

	// Setup test cases
	testCases := []struct {
		description string
		checkPath   string
		expected    interface{}
		compareFunc func(interface{}, interface{}) bool
	}{
		{
			description: "Access user1 username",
			checkPath:   "users.user1.Username",
			expected:    "johndoe",
		},
		{
			description: "Access user2 admin status",
			checkPath:   "users.user2.IsAdmin",
			expected:    true,
		},
		{
			description: "Access non-existent user",
			checkPath:   "users.user3.Username",
			expected:    nil, // Should not be found
		},
	}

	// Prepare context with the test data
	context := NewContext()
	context.Set("users", users)

	// Run the tests
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			value, err := context.Get(testCase.checkPath)

			// For the non-existent user test, we expect an error
			if testCase.expected == nil {
				if err == nil || !errors.Is(err, ErrContextKeyNotFound) {
					t.Errorf("%s: expected ErrContextKeyNotFound for path '%s', got %v",
						testCase.description, testCase.checkPath, err)
				}
				return
			}

			// For all other tests, we don't expect an error
			if err != nil {
				t.Errorf("%s: unexpected error: %v", testCase.description, err)
				return
			}

			// Use custom comparison function or default comparison
			if testCase.compareFunc != nil {
				if !testCase.compareFunc(value, testCase.expected) {
					t.Errorf("%s: expected %v (%T), got %v (%T)",
						testCase.description, testCase.expected, testCase.expected, value, value)
				}
				return
			}

			// Default comparison
			if !reflect.DeepEqual(value, testCase.expected) {
				t.Errorf("%s: expected %v (%T), got %v (%T)",
					testCase.description, testCase.expected, testCase.expected, value, value)
			}
		})
	}
}

// TestComplexNestedStructures tests deeply nested structures with a mix of
// maps, slices, and structs to ensure proper recursive conversion and access.
func TestComplexNestedStructures(t *testing.T) {
	// Define several nested types
	type Contact struct {
		Type  string
		Value string
	}

	type Address struct {
		Street     string
		City       string
		PostalCode string
		Country    string
		IsDefault  bool
	}

	type Product struct {
		ID          string
		Name        string
		Price       float64
		Description string
		Categories  []string
		Tags        map[string]string
		Inventory   map[string]int
	}

	type Review struct {
		UserID    string
		Rating    int
		Comment   string
		Date      time.Time
		Helpful   int
		Responses []string
	}

	type OrderItem struct {
		ProductID string
		Quantity  int
		UnitPrice float64
		Discount  float64
	}

	type Order struct {
		ID            string
		CustomerID    string
		Items         []OrderItem
		Total         float64
		ShippingInfo  map[string]string
		PaymentMethod map[string]interface{}
		Status        string
		CreatedAt     time.Time
	}

	type Customer struct {
		ID             string
		Name           string
		Age            int
		Email          string
		Addresses      []Address
		Contacts       []Contact
		PreferredItems []string
		Orders         []Order
		AccountBalance float64
		Metadata       map[string]interface{}
		IsVerified     bool
		JoinDate       time.Time
		LastLogin      time.Time
	}

	type Department struct {
		Name     string
		Manager  string
		Budget   float64
		Projects []string
		Staff    map[string]interface{}
	}

	type Company struct {
		Name         string
		Founded      time.Time
		Departments  map[string]Department
		Customers    map[string]Customer
		Products     []Product
		Reviews      map[string][]Review
		Headquarters Address
		Branches     []Address
		Revenue      map[string]float64
		Employees    int
		Partners     []string
		Settings     map[string]interface{}
	}

	// Create a deeply nested test data structure
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)
	lastMonth := now.Add(-30 * 24 * time.Hour)

	// Create complex nested data
	company := Company{
		Name:    "Acme Corporation",
		Founded: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		Departments: map[string]Department{
			"engineering": {
				Name:    "Engineering",
				Manager: "Jane Smith",
				Budget:  1000000.50,
				Projects: []string{
					"Project Alpha",
					"Project Beta",
					"Project Gamma",
				},
				Staff: map[string]interface{}{
					"senior": []string{"Alice", "Bob", "Charlie"},
					"junior": []string{"Dave", "Eve", "Frank"},
					"counts": map[string]int{
						"developers": 20,
						"testers":    10,
						"managers":   5,
					},
				},
			},
			"sales": {
				Name:    "Sales",
				Manager: "John Doe",
				Budget:  800000.75,
				Projects: []string{
					"North Region Campaign",
					"South Region Campaign",
				},
				Staff: map[string]interface{}{
					"senior": []string{"Grace", "Heidi"},
					"junior": []string{"Ivan", "Judy"},
					"counts": map[string]int{
						"representatives": 15,
						"managers":        3,
					},
				},
			},
		},
		Customers: map[string]Customer{
			"cust1": {
				ID:    "C-1001",
				Name:  "XYZ Corp",
				Age:   0, // Not applicable for companies
				Email: "contact@xyzcorp.com",
				Addresses: []Address{
					{
						Street:     "123 Business Ave",
						City:       "Commerce City",
						PostalCode: "10001",
						Country:    "USA",
						IsDefault:  true,
					},
					{
						Street:     "456 Enterprise Blvd",
						City:       "Commerce City",
						PostalCode: "10001",
						Country:    "USA",
						IsDefault:  false,
					},
				},
				Contacts: []Contact{
					{Type: "phone", Value: "555-1234"},
					{Type: "email", Value: "info@xyzcorp.com"},
				},
				PreferredItems: []string{"product1", "product3"},
				Orders: []Order{
					{
						ID:         "O-5001",
						CustomerID: "C-1001",
						Items: []OrderItem{
							{ProductID: "P-101", Quantity: 5, UnitPrice: 29.99, Discount: 0.1},
							{ProductID: "P-105", Quantity: 2, UnitPrice: 99.99, Discount: 0},
						},
						Total: 249.93,
						ShippingInfo: map[string]string{
							"method":   "express",
							"carrier":  "FedEx",
							"tracking": "FX123456789",
						},
						PaymentMethod: map[string]interface{}{
							"type":   "credit_card",
							"last4":  "1234",
							"expiry": "05/25",
						},
						Status:    "delivered",
						CreatedAt: lastMonth,
					},
				},
				AccountBalance: 5000.00,
				Metadata: map[string]interface{}{
					"sector":      "Technology",
					"size":        "Medium",
					"established": 2005,
					"contacts": map[string]string{
						"primary":   "John Smith",
						"secondary": "Jane Doe",
					},
				},
				IsVerified: true,
				JoinDate:   lastYear(now),
				LastLogin:  yesterday,
			},
			"cust2": {
				ID:    "C-1002",
				Name:  "John Consumer",
				Age:   35,
				Email: "john@example.com",
				Addresses: []Address{
					{
						Street:     "789 Residential St",
						City:       "Hometown",
						PostalCode: "20002",
						Country:    "USA",
						IsDefault:  true,
					},
				},
				Contacts: []Contact{
					{Type: "phone", Value: "555-5678"},
					{Type: "email", Value: "john@example.com"},
				},
				PreferredItems: []string{"product2", "product4"},
				Orders: []Order{
					{
						ID:         "O-5002",
						CustomerID: "C-1002",
						Items: []OrderItem{
							{ProductID: "P-102", Quantity: 1, UnitPrice: 59.99, Discount: 0},
						},
						Total: 59.99,
						ShippingInfo: map[string]string{
							"method":   "standard",
							"carrier":  "UPS",
							"tracking": "UPS987654321",
						},
						PaymentMethod: map[string]interface{}{
							"type":  "paypal",
							"email": "john@example.com",
						},
						Status:    "processing",
						CreatedAt: yesterday,
					},
				},
				AccountBalance: 150.50,
				Metadata: map[string]interface{}{
					"preferences": map[string]interface{}{
						"notifications": true,
						"theme":         "dark",
					},
					"deviceInfo": map[string]string{
						"browser": "Chrome",
						"os":      "Windows",
					},
				},
				IsVerified: true,
				JoinDate:   lastMonth,
				LastLogin:  now,
			},
		},
		Products: []Product{
			{
				ID:          "P-101",
				Name:        "Enterprise Software",
				Price:       299.99,
				Description: "Business solution software",
				Categories:  []string{"software", "business", "enterprise"},
				Tags: map[string]string{
					"level":        "premium",
					"subscription": "yearly",
					"support":      "24/7",
				},
				Inventory: map[string]int{
					"licenses": 500,
					"physical": 0,
				},
			},
			{
				ID:          "P-102",
				Name:        "Office Chair",
				Price:       199.99,
				Description: "Ergonomic office chair",
				Categories:  []string{"furniture", "office", "ergonomic"},
				Tags: map[string]string{
					"material": "leather",
					"color":    "black",
					"warranty": "2-year",
				},
				Inventory: map[string]int{
					"warehouse_a": 120,
					"warehouse_b": 85,
					"display":     5,
				},
			},
		},
		Reviews: map[string][]Review{
			"P-101": {
				{
					UserID:  "C-1001",
					Rating:  5,
					Comment: "Excellent software for our business needs",
					Date:    lastWeek,
					Helpful: 12,
					Responses: []string{
						"Thank you for your feedback!",
						"We appreciate your business.",
					},
				},
				{
					UserID:  "C-1002",
					Rating:  4,
					Comment: "Good software but a bit expensive",
					Date:    lastMonth,
					Helpful: 8,
					Responses: []string{
						"Thanks for your honest review.",
					},
				},
			},
			"P-102": {
				{
					UserID:  "C-1002",
					Rating:  5,
					Comment: "Very comfortable chair, worth every penny",
					Date:    lastWeek,
					Helpful: 15,
					Responses: []string{
						"We're glad you enjoy our product!",
					},
				},
			},
		},
		Headquarters: Address{
			Street:     "1 Corporate Plaza",
			City:       "Business City",
			PostalCode: "10005",
			Country:    "USA",
			IsDefault:  true,
		},
		Branches: []Address{
			{
				Street:     "25 East Business St",
				City:       "Eastern City",
				PostalCode: "20025",
				Country:    "USA",
				IsDefault:  false,
			},
			{
				Street:     "50 West Commerce Rd",
				City:       "Western City",
				PostalCode: "30050",
				Country:    "USA",
				IsDefault:  false,
			},
		},
		Revenue: map[string]float64{
			"2021": 5000000.00,
			"2022": 6250000.00,
			"2023": 7500000.00,
		},
		Employees: 250,
		Partners:  []string{"Partner A", "Partner B", "Partner C"},
		Settings: map[string]interface{}{
			"notifications": map[string]bool{
				"email":   true,
				"sms":     false,
				"desktop": true,
			},
			"security": map[string]interface{}{
				"mfa_required": true,
				"password_policy": map[string]interface{}{
					"min_length":      12,
					"require_special": true,
					"expiry_days":     90,
				},
			},
			"display": map[string]string{
				"logo":  "logo.png",
				"theme": "corporate",
			},
		},
	}

	// Setup test cases for deeply nested data
	testCases := []struct {
		description string
		checkPath   string
		expected    interface{}
		compareFunc func(interface{}, interface{}) bool
	}{
		{
			description: "Access company name",
			checkPath:   "company.Name",
			expected:    "Acme Corporation",
		},
		{
			description: "Access headquarters city",
			checkPath:   "company.Headquarters.City",
			expected:    "Business City",
		},
		{
			description: "Access 2022 revenue",
			checkPath:   "company.Revenue.2022",
			expected:    6250000.00,
			compareFunc: func(a, b interface{}) bool {
				aVal, aOk := toFloat64(a)
				bVal, bOk := toFloat64(b)
				if !aOk || !bOk {
					return false
				}
				return math.Abs(aVal-bVal) < 0.01
			},
		},
		{
			description: "Access first branch postal code",
			checkPath:   "company.Branches.0.PostalCode",
			expected:    "20025",
		},
		{
			description: "Access engineering department budget",
			checkPath:   "company.Departments.engineering.Budget",
			expected:    1000000.50,
			compareFunc: func(a, b interface{}) bool {
				aVal, aOk := toFloat64(a)
				bVal, bOk := toFloat64(b)
				if !aOk || !bOk {
					return false
				}
				return math.Abs(aVal-bVal) < 0.01
			},
		},
		{
			description: "Access second engineering project",
			checkPath:   "company.Departments.engineering.Projects.1",
			expected:    "Project Beta",
		},
		{
			description: "Access engineering senior staff",
			checkPath:   "company.Departments.engineering.Staff.senior.1",
			expected:    "Bob",
		},
		{
			description: "Access engineering testers count",
			checkPath:   "company.Departments.engineering.Staff.counts.testers",
			expected:    10,
			compareFunc: func(a, b interface{}) bool {
				aVal, aOk := toFloat64(a)
				bVal, bOk := toFloat64(b)
				if !aOk || !bOk {
					return false
				}
				return math.Abs(aVal-bVal) < 0.01
			},
		},
		{
			description: "Access first customer name",
			checkPath:   "company.Customers.cust1.Name",
			expected:    "XYZ Corp",
		},
		{
			description: "Access first customer's second address city",
			checkPath:   "company.Customers.cust1.Addresses.1.City",
			expected:    "Commerce City",
		},
		{
			description: "Access first customer's first contact type",
			checkPath:   "company.Customers.cust1.Contacts.0.Type",
			expected:    "phone",
		},
		{
			description: "Access first customer's first order's second item quantity",
			checkPath:   "company.Customers.cust1.Orders.0.Items.1.Quantity",
			expected:    2,
			compareFunc: func(a, b interface{}) bool {
				aVal, aOk := toFloat64(a)
				bVal, bOk := toFloat64(b)
				if !aOk || !bOk {
					return false
				}
				return math.Abs(aVal-bVal) < 0.01
			},
		},
		{
			description: "Access first customer's primary contact",
			checkPath:   "company.Customers.cust1.Metadata.contacts.primary",
			expected:    "John Smith",
		},
		{
			description: "Access second customer's device OS",
			checkPath:   "company.Customers.cust2.Metadata.deviceInfo.os",
			expected:    "Windows",
		},
		{
			description: "Access first product's price",
			checkPath:   "company.Products.0.Price",
			expected:    299.99,
			compareFunc: func(a, b interface{}) bool {
				aVal, aOk := toFloat64(a)
				bVal, bOk := toFloat64(b)
				if !aOk || !bOk {
					return false
				}
				return math.Abs(aVal-bVal) < 0.01
			},
		},
		{
			description: "Access first product's first category",
			checkPath:   "company.Products.0.Categories.0",
			expected:    "software",
		},
		{
			description: "Access second product's warranty tag",
			checkPath:   "company.Products.1.Tags.warranty",
			expected:    "2-year",
		},
		{
			description: "Access P-101 first review helpful count",
			checkPath:   "company.Reviews.P-101.0.Helpful",
			expected:    12,
			compareFunc: func(a, b interface{}) bool {
				aVal, aOk := toFloat64(a)
				bVal, bOk := toFloat64(b)
				if !aOk || !bOk {
					return false
				}
				return math.Abs(aVal-bVal) < 0.01
			},
		},
		{
			description: "Access P-102 first review response",
			checkPath:   "company.Reviews.P-102.0.Responses.0",
			expected:    "We're glad you enjoy our product!",
		},
		{
			description: "Access security settings password minimum length",
			checkPath:   "company.Settings.security.password_policy.min_length",
			expected:    12,
			compareFunc: func(a, b interface{}) bool {
				aVal, aOk := toFloat64(a)
				bVal, bOk := toFloat64(b)
				if !aOk || !bOk {
					return false
				}
				return math.Abs(aVal-bVal) < 0.01
			},
		},
		{
			description: "Access notification settings for email",
			checkPath:   "company.Settings.notifications.email",
			expected:    true,
		},
	}

	// Prepare context with the test data
	context := NewContext()
	context.Set("company", company)

	// Run the tests
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			value, err := context.Get(testCase.checkPath)
			if err != nil {
				t.Errorf("%s: unexpected error: %v", testCase.description, err)
				return
			}

			// Use custom comparison function or default comparison
			if testCase.compareFunc != nil {
				if !testCase.compareFunc(value, testCase.expected) {
					t.Errorf("%s: expected %v (%T), got %v (%T)",
						testCase.description, testCase.expected, testCase.expected, value, value)
				}
				return
			}

			// Default comparison
			if !reflect.DeepEqual(value, testCase.expected) {
				t.Errorf("%s: expected %v (%T), got %v (%T)",
					testCase.description, testCase.expected, testCase.expected, value, value)
			}
		})
	}
}

// lastYear returns a date from one year ago
func lastYear(t time.Time) time.Time {
	return t.AddDate(-1, 0, 0)
}
