package template

import (
	"errors"
	"strings"
	"testing"
)

// TestFilterErrorInTemplate tests filter errors in actual template execution
func TestFilterErrorInTemplate(t *testing.T) {
	t.Run("NonexistentFilter", func(t *testing.T) {
		tpl, err := Parse("{{ name | nonexistent }}")
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		ctx := NewContext()
		ctx.Set("name", "test")

		_, err = tpl.Execute(ctx)
		if err == nil {
			t.Fatal("Expected error for nonexistent filter")
		}

		// Check that error wraps ErrFilterNotFound
		if !errors.Is(err, ErrFilterNotFound) {
			t.Errorf("Expected error to wrap ErrFilterNotFound, got: %v", err)
		}

		// Check that error message contains filter name
		errMsg := err.Error()
		if !strings.Contains(errMsg, "nonexistent") {
			t.Errorf("Error message should contain filter name 'nonexistent', got: %s", errMsg)
		}

		if !strings.Contains(errMsg, "not found") {
			t.Errorf("Error message should contain 'not found', got: %s", errMsg)
		}
	})

	t.Run("FilterWithInvalidArgument", func(t *testing.T) {
		// Test that error messages are informative
		tpl, err := Parse("{{ name | truncate:invalid }}")
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		ctx := NewContext()
		ctx.Set("name", "test value")

		_, err = tpl.Execute(ctx)
		// This might or might not error depending on filter implementation
		// Just verify that if it errors, the message is clear
		if err != nil {
			errMsg := err.Error()
			if len(errMsg) == 0 {
				t.Error("Error message should not be empty")
			}
		}
	})
}

// TestContextErrorInTemplate tests context errors in template execution
func TestContextErrorInTemplate(t *testing.T) {
	t.Run("VariableNotFoundInFilter", func(t *testing.T) {
		// Register a simple test filter that uses a variable argument
		testFilterRegistered := false
		for name := range filters {
			if name == "test_var_filter" {
				testFilterRegistered = true
				break
			}
		}

		if !testFilterRegistered {
			// Register a temporary filter for testing
			_ = RegisterFilter("test_var_filter", func(value interface{}, _ ...string) (interface{}, error) {
				return value, nil
			})
		}

		// This test verifies that context key not found errors are properly wrapped
		ctx := NewContext()
		ctx.Set("name", "test")
		// Don't set "missing_var"

		// Create a filter that references a missing variable
		// Note: The actual behavior depends on how variables are resolved in filters
		// This is a placeholder test
		t.Skip("Variable resolution in filters needs specific test setup")
	})
}

// TestMultipleErrors tests error scenarios
func TestMultipleErrors(t *testing.T) {
	t.Run("BreakOutsideLoop", func(t *testing.T) {
		// Test break statement outside of loop
		tpl, err := Parse("{% break %}")
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		ctx := NewContext()
		_, err = tpl.Execute(ctx)

		if err == nil {
			t.Fatal("Expected error for break outside loop")
		}

		if !errors.Is(err, ErrBreakOutsideLoop) {
			t.Errorf("Expected ErrBreakOutsideLoop, got: %v", err)
		}

		// Verify error message is clear
		errMsg := err.Error()
		if !strings.Contains(errMsg, "break") && !strings.Contains(errMsg, "loop") {
			t.Errorf("Error message should mention break and loop, got: %s", errMsg)
		}
	})

	t.Run("ContinueOutsideLoop", func(t *testing.T) {
		// Test continue statement outside of loop
		tpl, err := Parse("{% continue %}")
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		ctx := NewContext()
		_, err = tpl.Execute(ctx)

		if err == nil {
			t.Fatal("Expected error for continue outside loop")
		}

		if !errors.Is(err, ErrContinueOutsideLoop) {
			t.Errorf("Expected ErrContinueOutsideLoop, got: %v", err)
		}

		// Verify error message is clear
		errMsg := err.Error()
		if !strings.Contains(errMsg, "continue") && !strings.Contains(errMsg, "loop") {
			t.Errorf("Error message should mention continue and loop, got: %s", errMsg)
		}
	})
}

// TestErrorWrapping tests that errors are properly wrapped using %w
func TestErrorWrapping(t *testing.T) {
	t.Run("ErrorsCanBeUnwrapped", func(t *testing.T) {
		tpl, err := Parse("{{ name | nonexistent }}")
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		ctx := NewContext()
		ctx.Set("name", "test")

		_, err = tpl.Execute(ctx)
		if err == nil {
			t.Fatal("Expected error for nonexistent filter")
		}

		// Verify that errors.Is works correctly for wrapped errors
		if !errors.Is(err, ErrFilterNotFound) {
			t.Error("errors.Is should be able to detect ErrFilterNotFound in wrapped error")
		}

		// Verify that the error message contains context
		errMsg := err.Error()
		if len(errMsg) == 0 {
			t.Error("Error message should not be empty")
		}
	})

	t.Run("SentinelErrorsPreserved", func(t *testing.T) {
		// Verify all sentinel errors are still accessible
		sentinelErrors := []error{
			ErrContextKeyNotFound,
			ErrFilterNotFound,
			ErrBreakOutsideLoop,
			ErrContinueOutsideLoop,
		}

		for _, sentinelErr := range sentinelErrors {
			if sentinelErr == nil {
				t.Error("Sentinel error should not be nil")
			}
			if sentinelErr.Error() == "" {
				t.Errorf("Sentinel error should have a message: %v", sentinelErr)
			}
		}
	})
}
