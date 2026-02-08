package template

import (
	"fmt"
	"sync"
	"testing"
)

// TestConcurrentTemplateExecution tests thread-safety of template execution
func TestConcurrentTemplateExecution(t *testing.T) {
	tpl, err := Parse("Hello {{ name }}!")
	if err != nil {
		t.Fatal(err)
	}

	// Run 100 concurrent executions
	var wg sync.WaitGroup
	results := make(chan string, 100)
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := NewContext()
			ctx.Set("name", fmt.Sprintf("User%d", id))
			result, err := tpl.Execute(ctx)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	var errCount int
	for err := range errors {
		t.Errorf("Execution error: %v", err)
		errCount++
	}

	// Check results
	var resultCount int
	for range results {
		resultCount++
	}

	if resultCount+errCount != 100 {
		t.Errorf("Expected 100 total results, got %d successful + %d errors", resultCount, errCount)
	}

	if errCount > 0 {
		t.Errorf("Found %d errors in concurrent execution", errCount)
	}
}

// TestConcurrentTemplateExecutionWithFilters tests concurrent execution with filters
func TestConcurrentTemplateExecutionWithFilters(t *testing.T) {
	tpl, err := Parse("{{ text | upper | truncate:10 }}")
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := NewContext()
			ctx.Set("text", fmt.Sprintf("test message %d", id))
			_, err := tpl.Execute(ctx)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent execution error: %v", err)
	}
}

// TestConcurrentParsing tests concurrent template parsing
func TestConcurrentParsing(t *testing.T) {
	templates := []string{
		"Hello {{ name }}!",
		"{{ text | upper }}",
		"{% for item in items %}{{ item }}{% endfor %}",
		"{% if active %}Active{% else %}Inactive{% endif %}",
		"{{ value | truncate:10 | lower }}",
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(templates)*10)

	for _, tmpl := range templates {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(template string) {
				defer wg.Done()
				_, err := Parse(template)
				if err != nil {
					errors <- err
				}
			}(tmpl)
		}
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent parsing error: %v", err)
	}
}

// TestFilterRegistryConcurrency tests concurrent filter access
func TestFilterRegistryConcurrency(t *testing.T) {
	// This test verifies that the filter registry can handle concurrent reads
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := NewContext()
			ctx.Set("value", "test")

			tpl, err := Parse("{{ value | upper }}")
			if err != nil {
				errors <- err
				return
			}

			_, err = tpl.Execute(ctx)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent filter access error: %v", err)
	}
}

// TestConcurrentContextAccess tests concurrent context operations
func TestConcurrentContextAccess(t *testing.T) {
	tpl, err := Parse("{{ user.name }} - {{ user.age }}")
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Each goroutine creates its own context
			ctx := NewContext()
			ctx.Set("user", map[string]any{
				"name": fmt.Sprintf("User%d", id),
				"age":  20 + id,
			})

			_, err := tpl.Execute(ctx)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent context access error: %v", err)
	}
}

// TestConcurrentLoopExecution tests concurrent execution of templates with loops
func TestConcurrentLoopExecution(t *testing.T) {
	tpl, err := Parse("{% for item in items %}{{ item }}{% endfor %}")
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 30)

	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := NewContext()
			items := make([]string, 10)
			for j := 0; j < 10; j++ {
				items[j] = fmt.Sprintf("item-%d-%d", id, j)
			}
			ctx.Set("items", items)

			_, err := tpl.Execute(ctx)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent loop execution error: %v", err)
	}
}

// TestConcurrentComplexTemplate tests concurrent execution of complex templates
func TestConcurrentComplexTemplate(t *testing.T) {
	template := `
	{% for user in users %}
		Name: {{ user.name | titleize }}
		Age: {{ user.age }}
		{% if user.active %}Active{% else %}Inactive{% endif %}
	{% endfor %}
	`

	tpl, err := Parse(template)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := NewContext()
			ctx.Set("users", []map[string]any{
				{"name": fmt.Sprintf("user%d-a", id), "age": 25, "active": true},
				{"name": fmt.Sprintf("user%d-b", id), "age": 30, "active": false},
				{"name": fmt.Sprintf("user%d-c", id), "age": 35, "active": true},
			})

			_, err := tpl.Execute(ctx)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent complex template error: %v", err)
	}
}

// BenchmarkConcurrentExecution benchmarks concurrent template execution
func BenchmarkConcurrentExecution(b *testing.B) {
	tpl, err := Parse("Hello {{ name }}!")
	if err != nil {
		b.Fatal(err)
	}

	b.RunParallel(func(pb *testing.PB) {
		id := 0
		for pb.Next() {
			ctx := NewContext()
			ctx.Set("name", fmt.Sprintf("User%d", id))
			_, _ = tpl.Execute(ctx)
			id++
		}
	})
}

// BenchmarkConcurrentParsingAndExecution benchmarks concurrent parsing and execution
func BenchmarkConcurrentParsingAndExecution(b *testing.B) {
	template := "{{ value | upper }}"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tpl, err := Parse(template)
			if err != nil {
				b.Fatal(err)
			}

			ctx := NewContext()
			ctx.Set("value", "test")
			_, _ = tpl.Execute(ctx)
		}
	})
}
