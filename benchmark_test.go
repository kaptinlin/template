package template

import (
	"slices"
	"strings"
	"testing"
)

// BenchmarkStringsCut benchmarks the strings.Cut() function for parsing filters
func BenchmarkStringsCut(b *testing.B) {
	input := "variable|filter1|filter2"
	b.ResetTimer()
	for b.Loop() {
		_, _, _ = strings.Cut(input, "|")
	}
}

// BenchmarkStringsSplitN benchmarks the old strings.SplitN() approach for comparison
func BenchmarkStringsSplitN(b *testing.B) {
	input := "variable|filter1|filter2"
	b.ResetTimer()
	for b.Loop() {
		parts := strings.SplitN(input, "|", 2)
		_ = parts[0]
		if len(parts) > 1 {
			_ = parts[1]
		}
	}
}

// BenchmarkSlicesClone benchmarks slices.Clone() for string slices
func BenchmarkSlicesClone(b *testing.B) {
	data := []string{"apple", "banana", "cherry", "date", "elderberry"}
	b.ResetTimer()
	for b.Loop() {
		_ = slices.Clone(data)
	}
}

// BenchmarkManualSliceCopy benchmarks the old make+copy approach for comparison
func BenchmarkManualSliceCopy(b *testing.B) {
	data := []string{"apple", "banana", "cherry", "date", "elderberry"}
	b.ResetTimer()
	for b.Loop() {
		newSlice := make([]string, len(data))
		copy(newSlice, data)
		_ = newSlice
	}
}

// BenchmarkSlicesCloneInt benchmarks slices.Clone() for int slices
func BenchmarkSlicesCloneInt(b *testing.B) {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer()
	for b.Loop() {
		_ = slices.Clone(data)
	}
}

// BenchmarkManualSliceCopyInt benchmarks the old make+copy approach for int slices
func BenchmarkManualSliceCopyInt(b *testing.B) {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer()
	for b.Loop() {
		newSlice := make([]int, len(data))
		copy(newSlice, data)
		_ = newSlice
	}
}

// BenchmarkStringBuilderPreallocated benchmarks strings.Builder with pre-allocation
func BenchmarkStringBuilderPreallocated(b *testing.B) {
	items := []string{"item1", "item2", "item3", "item4", "item5"}
	b.ResetTimer()
	for b.Loop() {
		var result strings.Builder
		result.Grow(len(items) * 10) // Pre-allocate
		result.WriteByte('[')
		for j, item := range items {
			if j > 0 {
				result.WriteByte(',')
			}
			result.WriteString(item)
		}
		result.WriteByte(']')
		_ = result.String()
	}
}

// BenchmarkStringBuilderNoPrealloc benchmarks strings.Builder without pre-allocation
func BenchmarkStringBuilderNoPrealloc(b *testing.B) {
	items := []string{"item1", "item2", "item3", "item4", "item5"}
	b.ResetTimer()
	for b.Loop() {
		var result strings.Builder
		result.WriteByte('[')
		for j, item := range items {
			if j > 0 {
				result.WriteByte(',')
			}
			result.WriteString(item)
		}
		result.WriteByte(']')
		_ = result.String()
	}
}

// BenchmarkStringJoin benchmarks the old strings.Join approach
func BenchmarkStringJoin(b *testing.B) {
	items := []string{"item1", "item2", "item3", "item4", "item5"}
	b.ResetTimer()
	for b.Loop() {
		parts := append([]string(nil), items...)
		_ = "[" + strings.Join(parts, ",") + "]"
	}
}

// BenchmarkTemplateExecuteSimple benchmarks simple template execution
func BenchmarkTemplateExecuteSimple(b *testing.B) {
	tpl, err := Parse("Hello {{ name }}!")
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewContext()
	ctx.Set("name", "World")

	b.ResetTimer()
	for b.Loop() {
		_, _ = tpl.Execute(ctx)
	}
}

// BenchmarkTemplateExecuteWithFilters benchmarks template with filters
func BenchmarkTemplateExecuteWithFilters(b *testing.B) {
	tpl, err := Parse("{{ text | upper | truncate:10 }}")
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewContext()
	ctx.Set("text", "hello world, this is a test")

	b.ResetTimer()
	for b.Loop() {
		_, _ = tpl.Execute(ctx)
	}
}

// BenchmarkTemplateExecuteForLoop benchmarks template with for loop
func BenchmarkTemplateExecuteForLoop(b *testing.B) {
	tpl, err := Parse("{% for item in items %}{{ item }}{% endfor %}")
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewContext()
	ctx.Set("items", []string{"a", "b", "c", "d", "e"})

	b.ResetTimer()
	for b.Loop() {
		_, _ = tpl.Execute(ctx)
	}
}

// BenchmarkTemplateExecuteComplex benchmarks complex template with nested structures
func BenchmarkTemplateExecuteComplex(b *testing.B) {
	template := `
	{% for user in users %}
		Name: {{ user.name | title }}
		Age: {{ user.age }}
		{% if user.active %}Active{% else %}Inactive{% endif %}
	{% endfor %}
	`
	tpl, err := Parse(template)
	if err != nil {
		b.Fatal(err)
	}

	ctx := NewContext()
	ctx.Set("users", []map[string]interface{}{
		{"name": "alice", "age": 30, "active": true},
		{"name": "bob", "age": 25, "active": false},
		{"name": "charlie", "age": 35, "active": true},
	})

	b.ResetTimer()
	for b.Loop() {
		_, _ = tpl.Execute(ctx)
	}
}

// BenchmarkParse benchmarks template parsing
func BenchmarkParse(b *testing.B) {
	template := "Hello {{ name | upper }}! {% if active %}Welcome{% endif %}"
	b.ResetTimer()
	for b.Loop() {
		_, _ = Parse(template)
	}
}

// BenchmarkParseComplex benchmarks parsing of complex template
func BenchmarkParseComplex(b *testing.B) {
	template := `
	{% for item in items %}
		{{ item.name | title }}
		{% if item.price > 100 %}
			Expensive: {{ item.price | format:"%.2f" }}
		{% else %}
			Affordable: {{ item.price | format:"%.2f" }}
		{% endif %}
	{% endfor %}
	`
	b.ResetTimer()
	for b.Loop() {
		_, _ = Parse(template)
	}
}
