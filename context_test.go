package template

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
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
			checkPath: "person.name",
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
			checkPath: "person.address.city",
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

// dereference 会自动解引用指针类型
func dereference(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		return val.Elem().Interface()
	}
	return v
}

// TestComplexStructConversion tests more complex struct conversion scenarios,
// including pointers, embedded structs, and time.Time fields.
func TestComplexStructConversion(t *testing.T) {
	// Define test structs with pointers and embedded types
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
			checkPath:   "project.metadata.tags.0",
			expected:    "important",
			compareFunc: func(a, b interface{}) bool {
				return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
			},
		},
		{
			description: "Access pointer value",
			checkPath:   "project.owner",
			expected:    "Project Owner",
			compareFunc: func(a, b interface{}) bool {
				// 对于指针类型，尝试解引用后比较
				a = dereference(a)
				b = dereference(b)
				return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
			},
		},
		{
			description: "Preserve time.Time",
			checkPath:   "project.created",
			expected:    creationTime.Format(time.RFC3339),
			compareFunc: func(a, b interface{}) bool {
				// 对于时间类型，使用字符串表示进行比较
				aTime, aOk := a.(time.Time)
				if aOk {
					return aTime.Format(time.RFC3339) == b
				}

				// 如果直接是字符串，也尝试比较
				aStr, aOk := a.(string)
				if aOk {
					return aStr == b
				}

				return false
			},
		},
		{
			description: "Access ID directly",
			checkPath:   "project.id",
			expected:    "project-123",
			compareFunc: func(a, b interface{}) bool {
				return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
			},
		},
		{
			description: "Access embedded field directly",
			checkPath:   "project.metadata.priority",
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
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
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
			checkPath:   "items.0.name",
			expected:    "Item 1",
			compareFunc: nil, // Use default comparison
		},
		{
			description: "Access second item price",
			checkPath:   "items.1.price",
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
			checkPath:   "items.2.id",
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
		Username string `json:"username"`
		Email    string `json:"email"`
		IsAdmin  bool   `json:"is_admin"`
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
			checkPath:   "users.user1.username",
			expected:    "johndoe",
		},
		{
			description: "Access user2 admin status",
			checkPath:   "users.user2.is_admin",
			expected:    true,
		},
		{
			description: "Access non-existent user",
			checkPath:   "users.user3.username",
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
		ID            string                 `json:"id"`
		CustomerID    string                 `json:"customerId"`
		Items         []OrderItem            `json:"items"`
		Total         float64                `json:"total"`
		ShippingInfo  map[string]string      `json:"shippingInfo"`
		PaymentMethod map[string]interface{} `json:"paymentMethod"`
		Status        string                 `json:"status"`
		CreatedAt     time.Time              `json:"createdAt"`
	}

	type Customer struct {
		ID             string                 `json:"id"`
		Name           string                 `json:"name"`
		Age            int                    `json:"age"`
		Email          string                 `json:"email"`
		Addresses      []Address              `json:"addresses"`
		Contacts       []Contact              `json:"contacts"`
		PreferredItems []string               `json:"preferredItems"`
		Orders         []Order                `json:"orders"`
		AccountBalance float64                `json:"accountBalance"`
		Metadata       map[string]interface{} `json:"metadata"`
		IsVerified     bool                   `json:"isVerified"`
		JoinDate       time.Time              `json:"joinDate"`
		LastLogin      time.Time              `json:"lastLogin"`
	}

	type Department struct {
		Name     string                 `json:"name"`
		Manager  string                 `json:"manager"`
		Budget   float64                `json:"budget"`
		Projects []string               `json:"projects"`
		Staff    map[string]interface{} `json:"staff"`
	}

	type Company struct {
		Name         string                 `json:"name"`
		Founded      time.Time              `json:"founded"`
		Departments  map[string]Department  `json:"departments"`
		Customers    map[string]Customer    `json:"customers"`
		Products     []Product              `json:"products"`
		Reviews      map[string][]Review    `json:"reviews"`
		Headquarters Address                `json:"headquarters"`
		Branches     []Address              `json:"branches"`
		Revenue      map[string]float64     `json:"revenue"`
		Employees    int                    `json:"employees"`
		Partners     []string               `json:"partners"`
		Settings     map[string]interface{} `json:"settings"`
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
			checkPath:   "company.name",
			expected:    "Acme Corporation",
		},
		{
			description: "Access headquarters city",
			checkPath:   "company.headquarters.city",
			expected:    "Business City",
		},
		{
			description: "Access 2022 revenue",
			checkPath:   "company.revenue.2022",
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
			checkPath:   "company.branches.0.postalCode",
			expected:    "20025",
		},
		{
			description: "Access engineering department budget",
			checkPath:   "company.departments.engineering.budget",
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
			checkPath:   "company.departments.engineering.projects.1",
			expected:    "Project Beta",
		},
		{
			description: "Access engineering senior staff",
			checkPath:   "company.departments.engineering.staff.senior.1",
			expected:    "Bob",
		},
		{
			description: "Access engineering testers count",
			checkPath:   "company.departments.engineering.staff.counts.testers",
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
			checkPath:   "company.customers.cust1.name",
			expected:    "XYZ Corp",
		},
		{
			description: "Access first customer's second address city",
			checkPath:   "company.customers.cust1.addresses.1.city",
			expected:    "Commerce City",
		},
		{
			description: "Access first customer's first contact type",
			checkPath:   "company.customers.cust1.contacts.0.type",
			expected:    "phone",
		},
		{
			description: "Access first customer's first order's second item quantity",
			checkPath:   "company.customers.cust1.orders.0.items.1.quantity",
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
			checkPath:   "company.customers.cust1.metadata.contacts.primary",
			expected:    "John Smith",
		},
		{
			description: "Access second customer's device OS",
			checkPath:   "company.customers.cust2.metadata.deviceInfo.os",
			expected:    "Windows",
		},
		{
			description: "Access first product's price",
			checkPath:   "company.products.0.price",
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
			checkPath:   "company.products.0.categories.0",
			expected:    "software",
		},
		{
			description: "Access second product's warranty tag",
			checkPath:   "company.products.1.tags.warranty",
			expected:    "2-year",
		},
		{
			description: "Access P-101 first review helpful count",
			checkPath:   "company.reviews.P-101.0.helpful",
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
			checkPath:   "company.reviews.P-102.0.responses.0",
			expected:    "We're glad you enjoy our product!",
		},
		{
			description: "Access security settings password minimum length",
			checkPath:   "company.settings.security.password_policy.min_length",
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
			checkPath:   "company.settings.notifications.email",
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

// TestSetPreservesOriginalTypes tests that Set method preserves original data types
// and jsonpointer can read them correctly
func TestSetPreservesOriginalTypes(t *testing.T) {
	type User struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	type Company struct {
		Name      string          `json:"name"`
		Employees []User          `json:"employees"`
		Settings  map[string]bool `json:"settings"`
	}

	tests := []struct {
		name     string
		key      string
		value    interface{}
		getKey   string
		expected interface{}
	}{
		{
			name:     "Set struct directly",
			key:      "user",
			value:    User{Name: "John", Age: 30, Email: "john@example.com"},
			getKey:   "user.name",
			expected: "John",
		},
		{
			name:     "Set slice of structs",
			key:      "users",
			value:    []User{{Name: "Alice", Age: 25}, {Name: "Bob", Age: 35}},
			getKey:   "users.1.name",
			expected: "Bob",
		},
		{
			name: "Set complex nested structure",
			key:  "company",
			value: Company{
				Name: "TechCorp",
				Employees: []User{
					{Name: "Charlie", Age: 28, Email: "charlie@techcorp.com"},
					{Name: "Diana", Age: 32, Email: "diana@techcorp.com"},
				},
				Settings: map[string]bool{"remote": true, "flexible": false},
			},
			getKey:   "company.employees.0.email",
			expected: "charlie@techcorp.com",
		},
		{
			name:     "Set map directly",
			key:      "config",
			value:    map[string]interface{}{"debug": true, "port": 8080},
			getKey:   "config.port",
			expected: 8080,
		},
		{
			name:     "Set nested with preserved types",
			key:      "nested.data",
			value:    []map[string]string{{"type": "test", "value": "data"}},
			getKey:   "nested.data.0.type",
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()

			// Set the value
			ctx.Set(tt.key, tt.value)

			// Try to get the nested value using jsonpointer-compatible Get method
			result, err := ctx.Get(tt.getKey)
			if err != nil {
				t.Fatalf("Failed to get value for key '%s': %v", tt.getKey, err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}

			// Also verify that the top-level value maintains its original type
			topLevelKey := strings.Split(tt.key, ".")[0]
			topLevelValue, err := ctx.Get(topLevelKey)
			if err != nil {
				t.Fatalf("Failed to get top-level value for key '%s': %v", topLevelKey, err)
			}

			// The type should be preserved
			switch tt.value.(type) {
			case User:
				if _, ok := topLevelValue.(User); !ok {
					t.Errorf("Expected User type to be preserved, got %T", topLevelValue)
				}
			case []User:
				if _, ok := topLevelValue.([]User); !ok {
					t.Errorf("Expected []User type to be preserved, got %T", topLevelValue)
				}
			case Company:
				if _, ok := topLevelValue.(Company); !ok {
					t.Errorf("Expected Company type to be preserved, got %T", topLevelValue)
				}
			}
		})
	}
}

// TestRenderingStability tests that template rendering produces consistent output
// across multiple executions for the same context data.
func TestRenderingStability(t *testing.T) {
	// Test multiple rendering iterations to catch ordering issues
	const iterations = 10

	testCases := []struct {
		name        string
		template    string
		setupFunc   func() Context
		description string
	}{
		{
			name:     "Map string iteration stability",
			template: `{% for key in data %}{{ key.key }}:{{ key.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]string{
					"zebra":   "last",
					"alpha":   "first",
					"beta":    "second",
					"gamma":   "third",
					"delta":   "fourth",
					"echo":    "fifth",
					"foxtrot": "sixth",
				})
				return ctx
			},
			description: "Map with string keys should maintain consistent iteration order",
		},
		{
			name:     "Map interface iteration stability",
			template: `{% for item in users %}{{ item.key }}:{{ item.value.name }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("users", map[string]interface{}{
					"user3": map[string]interface{}{"name": "Charlie", "age": 35},
					"user1": map[string]interface{}{"name": "Alice", "age": 25},
					"user4": map[string]interface{}{"name": "David", "age": 40},
					"user2": map[string]interface{}{"name": "Bob", "age": 30},
				})
				return ctx
			},
			description: "Map with interface{} values should maintain consistent iteration order",
		},
		{
			name:     "Nested map iteration stability",
			template: `{% for dept in company %}{{ dept.key }}:{% for emp in dept.value %}{{ emp.key }}-{{ emp.value }},{% endfor %};{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("company", map[string]map[string]string{
					"engineering": {
						"john":  "senior",
						"alice": "junior",
						"bob":   "lead",
					},
					"sales": {
						"mary": "manager",
						"tom":  "rep",
					},
					"hr": {
						"susan": "director",
					},
				})
				return ctx
			},
			description: "Nested maps should maintain consistent iteration order",
		},
		{
			name:     "Slice rendering stability",
			template: `{% for item in items %}{{ item }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("items", []string{"first", "second", "third", "fourth", "fifth"})
				return ctx
			},
			description: "Slice should maintain consistent order (this should be stable)",
		},
		{
			name:     "Struct slice rendering stability",
			template: `{% for person in people %}{{ person.name }}-{{ person.age }},{% endfor %}`,
			setupFunc: func() Context {
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
			description: "Slice of structs should maintain consistent order",
		},
		{
			name:     "Map variable rendering stability",
			template: `{{ data }}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]interface{}{
					"zebra": "value1",
					"alpha": "value2",
					"beta":  "value3",
					"gamma": "value4",
				})
				return ctx
			},
			description: "Direct map variable rendering should be consistent",
		},
		{
			name:     "Complex nested structure rendering stability",
			template: `{{ complex }}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("complex", map[string]interface{}{
					"users": map[string]interface{}{
						"user2": []string{"read", "write"},
						"user1": []string{"admin", "read", "write"},
					},
					"settings": map[string]interface{}{
						"theme": "dark",
						"lang":  "en",
					},
					"data": []map[string]string{
						{"name": "item1", "type": "A"},
						{"name": "item2", "type": "B"},
					},
				})
				return ctx
			},
			description: "Complex nested structures should render consistently",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse template once
			template, err := Parse(tc.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			var results []string

			// Run multiple iterations to detect inconsistencies
			for i := 0; i < iterations; i++ {
				ctx := tc.setupFunc()
				result, err := template.Execute(ctx)
				if err != nil {
					t.Fatalf("Iteration %d failed: %v", i, err)
				}
				results = append(results, result)
			}

			// Check that all results are identical
			firstResult := results[0]
			for i := 1; i < iterations; i++ {
				if results[i] != firstResult {
					t.Errorf("%s: Inconsistent output detected\n"+
						"Iteration 0: %q\n"+
						"Iteration %d: %q\n"+
						"All results: %v",
						tc.description, firstResult, i, results[i], results)
					break
				}
			}

			// Log the consistent result for manual verification
			t.Logf("%s - Consistent output: %q", tc.name, firstResult)
		})
	}
}

// TestMapKeyOrderStability specifically tests map key ordering in different scenarios
func TestMapKeyOrderStability(t *testing.T) {
	const iterations = 20 // More iterations for map ordering tests

	testCases := []struct {
		name string
		data map[string]interface{}
	}{
		{
			name: "String keys",
			data: map[string]interface{}{
				"zebra": "last",
				"alpha": "first",
				"beta":  "second",
				"gamma": "third",
				"delta": "fourth",
			},
		},
		{
			name: "Numeric-like string keys",
			data: map[string]interface{}{
				"10": "ten",
				"1":  "one",
				"5":  "five",
				"2":  "two",
				"20": "twenty",
			},
		},
		{
			name: "Mixed case keys",
			data: map[string]interface{}{
				"Apple":  "fruit1",
				"banana": "fruit2",
				"Cherry": "fruit3",
				"date":   "fruit4",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var keyOrders [][]string

			for i := 0; i < iterations; i++ {
				ctx := NewContext()
				ctx.Set("data", tc.data)

				var observedKeys []string
				// Use reflection to get map keys in the order they would be iterated
				rv := reflect.ValueOf(tc.data)
				for _, key := range rv.MapKeys() {
					observedKeys = append(observedKeys, key.String())
				}
				keyOrders = append(keyOrders, observedKeys)
			}

			// Check if all key orders are the same
			firstOrder := keyOrders[0]
			allSame := true
			for i := 1; i < iterations; i++ {
				if !reflect.DeepEqual(keyOrders[i], firstOrder) {
					allSame = false
					t.Logf("Key order difference detected:\n"+
						"Iteration 0: %v\n"+
						"Iteration %d: %v", firstOrder, i, keyOrders[i])
					break
				}
			}

			if allSame {
				t.Logf("%s: Key order is stable across %d iterations: %v",
					tc.name, iterations, firstOrder)
			} else {
				t.Logf("%s: Key order is NOT stable - this is expected behavior in Go", tc.name)
			}
		})
	}
}

// TestContextValueStability tests that different types of values maintain stability when stored and retrieved
func TestContextValueStability(t *testing.T) {
	const iterations = 10

	testCases := []struct {
		name     string
		key      string
		value    interface{}
		checkKey string
	}{
		{
			name:     "Map value stability",
			key:      "config",
			value:    map[string]interface{}{"a": 1, "z": 2, "m": 3, "b": 4},
			checkKey: "config",
		},
		{
			name:     "Slice value stability",
			key:      "items",
			value:    []interface{}{"third", "first", "second"},
			checkKey: "items",
		},
		{
			name: "Struct value stability",
			key:  "user",
			value: struct {
				Name   string            `json:"name"`
				Age    int               `json:"age"`
				Tags   map[string]string `json:"tags"`
				Skills []string          `json:"skills"`
			}{
				Name:   "John",
				Age:    30,
				Tags:   map[string]string{"level": "senior", "team": "backend"},
				Skills: []string{"Go", "Python", "SQL"},
			},
			checkKey: "user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var results []interface{}

			for i := 0; i < iterations; i++ {
				ctx := NewContext()
				ctx.Set(tc.key, tc.value)

				retrieved, err := ctx.Get(tc.checkKey)
				if err != nil {
					t.Fatalf("Iteration %d: Failed to get value: %v", i, err)
				}
				results = append(results, retrieved)
			}

			// For basic stability check, verify the retrieved values are deeply equal
			firstResult := results[0]
			for i := 1; i < iterations; i++ {
				if !reflect.DeepEqual(results[i], firstResult) {
					t.Errorf("Value stability failed at iteration %d\n"+
						"Expected: %+v\n"+
						"Got: %+v", i, firstResult, results[i])
				}
			}

			t.Logf("%s: Value remains stable across %d iterations", tc.name, iterations)
		})
	}
}

// TestTemplateRenderingConsistency tests full template rendering consistency
func TestTemplateRenderingConsistency(t *testing.T) {
	const iterations = 15

	type User struct {
		Name     string            `json:"name"`
		Age      int               `json:"age"`
		Roles    []string          `json:"roles"`
		Settings map[string]string `json:"settings"`
	}

	type Company struct {
		Name  string          `json:"name"`
		Users map[string]User `json:"users"`
		Tags  []string        `json:"tags"`
	}

	template := `Company: {{ company.name }}
Users:
{% for user in company.users %}  - {{ user.key }}: {{ user.value.name }} ({{ user.value.age }})
    Roles: {% for role in user.value.roles %}{{ role }}, {% endfor %}
    Settings: {% for setting in user.value.settings %}{{ setting.key }}={{ setting.value }}, {% endfor %}
{% endfor %}
Tags: {% for tag in company.tags %}{{ tag }}, {% endfor %}`

	setupData := func() Context {
		ctx := NewContext()
		company := Company{
			Name: "TechCorp",
			Users: map[string]User{
				"user3": {
					Name:     "Charlie",
					Age:      35,
					Roles:    []string{"developer", "lead"},
					Settings: map[string]string{"theme": "dark", "lang": "en"},
				},
				"user1": {
					Name:     "Alice",
					Age:      25,
					Roles:    []string{"admin", "developer"},
					Settings: map[string]string{"theme": "light", "notifications": "on"},
				},
				"user2": {
					Name:     "Bob",
					Age:      30,
					Roles:    []string{"developer"},
					Settings: map[string]string{"lang": "fr"},
				},
			},
			Tags: []string{"startup", "tech", "remote"},
		}
		ctx.Set("company", company)
		return ctx
	}

	tmpl, err := Parse(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	var results []string
	for i := 0; i < iterations; i++ {
		ctx := setupData()
		result, err := tmpl.Execute(ctx)
		if err != nil {
			t.Fatalf("Template execution failed at iteration %d: %v", i, err)
		}
		results = append(results, result)
	}

	// Check consistency
	firstResult := results[0]
	inconsistentFound := false
	for i := 1; i < iterations; i++ {
		if results[i] != firstResult {
			inconsistentFound = true
			t.Errorf("Template rendering inconsistency detected:\n"+
				"=== Iteration 0 ===\n%s\n"+
				"=== Iteration %d ===\n%s\n", firstResult, i, results[i])
			break
		}
	}

	if !inconsistentFound {
		t.Logf("Template rendering is consistent across %d iterations", iterations)
		t.Logf("Sample output:\n%s", firstResult)
	}
}

// TestRenderingStabilityEdgeCases tests rendering stability for edge cases and boundary conditions
func TestRenderingStabilityEdgeCases(t *testing.T) {
	const iterations = 10

	testCases := []struct {
		name        string
		template    string
		setupFunc   func() Context
		description string
	}{
		{
			name:     "Empty map iteration",
			template: `{% for item in data %}{{ item.key }}:{{ item.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]string{})
				return ctx
			},
			description: "Empty map should render consistently (empty result)",
		},
		{
			name:     "Single item map iteration",
			template: `{% for item in data %}{{ item.key }}:{{ item.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]string{"onlykey": "onlyvalue"})
				return ctx
			},
			description: "Single item map should render consistently",
		},
		{
			name:     "Empty slice iteration",
			template: `{% for item in data %}{{ item }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", []string{})
				return ctx
			},
			description: "Empty slice should render consistently (empty result)",
		},
		{
			name:     "Nil values in map",
			template: `{% for item in data %}{{ item.key }}:{{ item.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]interface{}{
					"nil_value": nil,
					"empty_str": "",
					"zero_int":  0,
					"false_val": false,
				})
				return ctx
			},
			description: "Map with nil and falsy values should render consistently",
		},
		{
			name:     "Special characters in keys",
			template: `{% for item in data %}{{ item.key }}:{{ item.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]string{
					"key-with-dashes":      "value1",
					"key_with_underscores": "value2",
					"key.with.dots":        "value3",
					"key with spaces":      "value4",
					"key@with#symbols":     "value5",
				})
				return ctx
			},
			description: "Keys with special characters should maintain stable order",
		},
		{
			name:     "Unicode keys and values",
			template: `{% for item in data %}{{ item.key }}:{{ item.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]string{
					"中文":      "中文值",
					"العربية": "قيمة",
					"русский": "значение",
					"日本語":     "値",
					"한국어":     "값",
					"emoji":   "🎉🎈🎊",
				})
				return ctx
			},
			description: "Unicode keys and values should maintain stable order",
		},
		{
			name:     "Numeric string keys",
			template: `{% for item in data %}{{ item.key }}:{{ item.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]string{
					"100": "hundred",
					"1":   "one",
					"10":  "ten",
					"2":   "two",
					"01":  "zero-one",
					"001": "zero-zero-one",
					"1.5": "one-point-five",
					"-1":  "negative-one",
					"0":   "zero",
				})
				return ctx
			},
			description: "Numeric string keys should sort lexicographically",
		},
		{
			name:     "Large map iteration",
			template: `{% for item in data %}{{ item.key }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				largeMap := make(map[string]string)
				for i := 0; i < 100; i++ {
					key := fmt.Sprintf("key_%03d", i)
					largeMap[key] = fmt.Sprintf("value_%d", i)
				}
				ctx.Set("data", largeMap)
				return ctx
			},
			description: "Large map should maintain stable iteration order",
		},
		{
			name:     "Deeply nested map structures",
			template: `{% for l1 in data %}{{ l1.key }}:{% for l2 in l1.value %}{{ l2.key }}:{% for l3 in l2.value %}{{ l3.key }}-{{ l3.value }},{% endfor %};{% endfor %};{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]map[string]map[string]string{
					"level1_b": {
						"level2_y": {"level3_z": "value_z", "level3_x": "value_x"},
						"level2_x": {"level3_a": "value_a"},
					},
					"level1_a": {
						"level2_z": {"level3_b": "value_b"},
						"level2_a": {"level3_y": "value_y", "level3_z": "value_z2"},
					},
				})
				return ctx
			},
			description: "Deeply nested maps should maintain stable order at all levels",
		},
		{
			name:     "Mixed types in map values",
			template: `{% for item in data %}{{ item.key }}:{{ item.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]interface{}{
					"string_val": "text",
					"int_val":    123,
					"float_val":  45.67,
					"bool_val":   true,
					"slice_val":  []string{"a", "b"},
					"map_val":    map[string]string{"nested": "value"},
				})
				return ctx
			},
			description: "Map with mixed value types should render consistently",
		},
		{
			name:     "Very long keys and values",
			template: `{% for item in data %}{{ item.key }}:{{ item.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				longKey := strings.Repeat("a", 1000) + "_key"
				longValue := strings.Repeat("x", 1000)
				ctx.Set("data", map[string]string{
					"short":  "short_value",
					longKey:  longValue,
					"medium": strings.Repeat("m", 100),
				})
				return ctx
			},
			description: "Maps with very long keys and values should be stable",
		},
		{
			name:     "Case sensitive key ordering",
			template: `{% for item in data %}{{ item.key }}:{{ item.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]string{
					"Apple":  "fruit1",
					"apple":  "fruit2",
					"APPLE":  "fruit3",
					"aPpLe":  "fruit4",
					"Banana": "fruit5",
					"banana": "fruit6",
				})
				return ctx
			},
			description: "Case sensitive keys should be sorted correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			template, err := Parse(tc.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			var results []string
			for i := 0; i < iterations; i++ {
				ctx := tc.setupFunc()
				result, err := template.Execute(ctx)
				if err != nil {
					t.Fatalf("Iteration %d failed: %v", i, err)
				}
				results = append(results, result)
			}

			// Check consistency
			firstResult := results[0]
			for i := 1; i < iterations; i++ {
				if results[i] != firstResult {
					t.Errorf("%s: Inconsistent output detected\n"+
						"Iteration 0: %q\n"+
						"Iteration %d: %q",
						tc.description, firstResult, i, results[i])
					break
				}
			}

			t.Logf("%s - Consistent output: %q", tc.name, firstResult)
		})
	}
}

// TestRenderingStabilityPerformance tests rendering stability with performance considerations
func TestRenderingStabilityPerformance(t *testing.T) {
	const iterations = 5 // Fewer iterations for performance tests

	testCases := []struct {
		name        string
		template    string
		setupFunc   func() Context
		description string
	}{
		{
			name:     "Very large map (1000 items)",
			template: `Count: {% for item in data %}1{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				largeMap := make(map[string]int)
				for i := 0; i < 1000; i++ {
					// Use random-ish keys to test sorting performance
					key := fmt.Sprintf("key_%04d_%d", 1000-i, i*7%13)
					largeMap[key] = i
				}
				ctx.Set("data", largeMap)
				return ctx
			},
			description: "Very large map should maintain stability without significant performance degradation",
		},
		{
			name:     "Deep nesting (10 levels)",
			template: `{{ deep }}`,
			setupFunc: func() Context {
				ctx := NewContext()
				// Create deeply nested structure
				current := make(map[string]interface{})
				root := current
				for i := 0; i < 10; i++ {
					next := make(map[string]interface{})
					current[fmt.Sprintf("level_%d_key_z", i)] = fmt.Sprintf("value_%d", i)
					current[fmt.Sprintf("level_%d_key_a", i)] = fmt.Sprintf("value_%d", i)
					current["nested"] = next
					current = next
				}
				current["final"] = "deep_value"
				ctx.Set("deep", root)
				return ctx
			},
			description: "Deep nesting should not cause performance issues with stability",
		},
		{
			name:     "Wide map (many keys at same level)",
			template: `{% for item in data %}{{ item.key }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				wideMap := make(map[string]string)
				// Create many keys with different prefixes to test sorting
				prefixes := []string{"apple", "banana", "cherry", "date", "elderberry"}
				for _, prefix := range prefixes {
					for i := 0; i < 50; i++ {
						key := fmt.Sprintf("%s_%03d", prefix, i)
						wideMap[key] = fmt.Sprintf("value_%s_%d", prefix, i)
					}
				}
				ctx.Set("data", wideMap)
				return ctx
			},
			description: "Wide map with many keys should maintain stable order efficiently",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			template, err := Parse(tc.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			var results []string
			start := time.Now()

			for i := 0; i < iterations; i++ {
				ctx := tc.setupFunc()
				result, err := template.Execute(ctx)
				if err != nil {
					t.Fatalf("Iteration %d failed: %v", i, err)
				}
				results = append(results, result)
			}

			duration := time.Since(start)

			// Check consistency
			firstResult := results[0]
			for i := 1; i < iterations; i++ {
				if results[i] != firstResult {
					t.Errorf("%s: Inconsistent output detected", tc.description)
					break
				}
			}

			t.Logf("%s - Completed %d iterations in %v (avg: %v per iteration)",
				tc.name, iterations, duration, duration/time.Duration(iterations))
		})
	}
}

// TestContextStabilityWithStructs tests rendering stability with various struct configurations
func TestContextStabilityWithStructs(t *testing.T) {
	const iterations = 10

	// Define test structs with different configurations
	type SimpleStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type StructWithTags struct {
		PublicField  string `json:"public"`
		privateField string // Unexported field should be ignored
		IgnoredField string `json:"-"` // Explicitly ignored field
		CustomName   string `json:"custom_name"`
	}

	type NestedStruct struct {
		ID       string                 `json:"id"`
		Simple   SimpleStruct           `json:"simple"`
		Settings map[string]interface{} `json:"settings"`
		Items    []string               `json:"items"`
	}

	type StructWithPointers struct {
		Name    *string            `json:"name"`
		Age     *int               `json:"age"`
		Config  *map[string]string `json:"config"`
		Enabled *bool              `json:"enabled"`
	}

	testCases := []struct {
		name        string
		template    string
		setupFunc   func() Context
		description string
	}{
		{
			name:     "Simple struct rendering",
			template: `{{ person }}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("person", SimpleStruct{Name: "Alice", Age: 30})
				return ctx
			},
			description: "Simple struct should render consistently",
		},
		{
			name:     "Struct with field tags",
			template: `{{ data }}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", StructWithTags{
					PublicField:  "public_value",
					privateField: "private_value", // Should not appear in JSON
					IgnoredField: "ignored_value", // Should not appear in JSON
					CustomName:   "custom_value",
				})
				return ctx
			},
			description: "Struct with JSON tags should render consistently",
		},
		{
			name:     "Nested struct iteration",
			template: `{% for item in data %}{{ item.key }}:{{ item.value }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				data := map[string]NestedStruct{
					"item_b": {
						ID:     "id_b",
						Simple: SimpleStruct{Name: "Bob", Age: 25},
						Settings: map[string]interface{}{
							"theme": "dark",
							"lang":  "en",
						},
						Items: []string{"item1", "item2"},
					},
					"item_a": {
						ID:     "id_a",
						Simple: SimpleStruct{Name: "Alice", Age: 30},
						Settings: map[string]interface{}{
							"theme":         "light",
							"notifications": true,
						},
						Items: []string{"item3", "item4"},
					},
				}
				ctx.Set("data", data)
				return ctx
			},
			description: "Map of nested structs should maintain stable order",
		},
		{
			name:     "Struct with pointers",
			template: `{{ data }}`,
			setupFunc: func() Context {
				ctx := NewContext()
				name := "John"
				age := 35
				config := map[string]string{"key": "value"}
				enabled := true

				ctx.Set("data", StructWithPointers{
					Name:    &name,
					Age:     &age,
					Config:  &config,
					Enabled: &enabled,
				})
				return ctx
			},
			description: "Struct with pointer fields should render consistently",
		},
		{
			name:     "Slice of structs with map fields",
			template: `{% for person in people %}{{ person.name }}:{% for setting in person.settings %}{{ setting.key }}={{ setting.value }},{% endfor %};{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				type PersonWithSettings struct {
					Name     string            `json:"name"`
					Settings map[string]string `json:"settings"`
				}

				people := []PersonWithSettings{
					{
						Name: "Charlie",
						Settings: map[string]string{
							"theme":         "auto",
							"lang":          "fr",
							"notifications": "off",
						},
					},
					{
						Name: "Alice",
						Settings: map[string]string{
							"theme": "light",
							"lang":  "en",
						},
					},
				}
				ctx.Set("people", people)
				return ctx
			},
			description: "Slice of structs with map fields should maintain stable order",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			template, err := Parse(tc.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			var results []string
			for i := 0; i < iterations; i++ {
				ctx := tc.setupFunc()
				result, err := template.Execute(ctx)
				if err != nil {
					t.Fatalf("Iteration %d failed: %v", i, err)
				}
				results = append(results, result)
			}

			// Check consistency
			firstResult := results[0]
			for i := 1; i < iterations; i++ {
				if results[i] != firstResult {
					t.Errorf("%s: Inconsistent output detected\n"+
						"Iteration 0: %q\n"+
						"Iteration %d: %q",
						tc.description, firstResult, i, results[i])
					break
				}
			}

			t.Logf("%s - Consistent output length: %d chars", tc.name, len(firstResult))
		})
	}
}

// TestRenderingStabilityErrorCases tests stability in error conditions
func TestRenderingStabilityErrorCases(t *testing.T) {
	const iterations = 5

	testCases := []struct {
		name        string
		template    string
		setupFunc   func() Context
		description string
		expectError bool
	}{
		{
			name:        "Missing variable in map iteration",
			template:    `{% for item in missing_var %}{{ item }},{% endfor %}`,
			setupFunc:   NewContext,
			description: "Missing variable should produce consistent error behavior",
			expectError: true,
		},
		{
			name:     "Type mismatch for iteration",
			template: `{% for item in not_iterable %}{{ item }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("not_iterable", "this is a string, not iterable as expected")
				return ctx
			},
			description: "Type mismatch should produce consistent behavior",
			expectError: false, // String iteration is actually supported
		},
		{
			name:     "Accessing non-existent nested properties",
			template: `{% for item in data %}{{ item.nonexistent.property }},{% endfor %}`,
			setupFunc: func() Context {
				ctx := NewContext()
				ctx.Set("data", map[string]interface{}{
					"item1": map[string]string{"name": "value1"},
					"item2": map[string]string{"name": "value2"},
				})
				return ctx
			},
			description: "Non-existent nested properties should render consistently",
			expectError: false, // Should render empty strings consistently
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			template, err := Parse(tc.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			var results []string
			var errors []error

			for i := 0; i < iterations; i++ {
				ctx := tc.setupFunc()
				result, err := template.Execute(ctx)
				results = append(results, result)
				errors = append(errors, err)
			}

			// Check error consistency
			firstError := errors[0]
			for i := 1; i < iterations; i++ {
				if (firstError == nil) != (errors[i] == nil) {
					t.Errorf("%s: Inconsistent error behavior\n"+
						"Iteration 0 error: %v\n"+
						"Iteration %d error: %v",
						tc.description, firstError, i, errors[i])
					return
				}
			}

			// Check result consistency (even with errors, partial results should be consistent)
			firstResult := results[0]
			for i := 1; i < iterations; i++ {
				if results[i] != firstResult {
					t.Errorf("%s: Inconsistent output detected\n"+
						"Iteration 0: %q\n"+
						"Iteration %d: %q",
						tc.description, firstResult, i, results[i])
					break
				}
			}

			if firstError != nil {
				t.Logf("%s - Consistent error: %v", tc.name, firstError)
			} else {
				t.Logf("%s - Consistent output: %q", tc.name, firstResult)
			}
		})
	}
}
