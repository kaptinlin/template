package template

import (
	"errors"
	"reflect"
	"testing"
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
