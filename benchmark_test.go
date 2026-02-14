package template

import (
	"bytes"
	"testing"
)

// BenchmarkCompile benchmarks template compilation.
func BenchmarkCompile(b *testing.B) {
	source := `Hello {{ name }}! {% if age > 18 %}You are an adult.{% endif %}`
	for b.Loop() {
		_, err := Compile(source)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExecuteSimple benchmarks simple template execution.
func BenchmarkExecuteSimple(b *testing.B) {
	tmpl, err := Compile(`Hello {{ name }}!`)
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewContext()
	ctx.Set("name", "World")
	execCtx := NewExecutionContext(ctx)
	var buf bytes.Buffer

	b.ResetTimer()
	for b.Loop() {
		buf.Reset()
		if err := tmpl.Execute(execCtx, &buf); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExecuteWithFilters benchmarks template execution with filters.
func BenchmarkExecuteWithFilters(b *testing.B) {
	tmpl, err := Compile(`{{ name | upper | append: "!" }}`)
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewContext()
	ctx.Set("name", "world")
	execCtx := NewExecutionContext(ctx)
	var buf bytes.Buffer

	b.ResetTimer()
	for b.Loop() {
		buf.Reset()
		if err := tmpl.Execute(execCtx, &buf); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExecuteWithLoop benchmarks template execution with loops.
func BenchmarkExecuteWithLoop(b *testing.B) {
	tmpl, err := Compile(`{% for item in items %}{{ item }}{% endfor %}`)
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewContext()
	ctx.Set("items", []string{"a", "b", "c", "d", "e"})
	execCtx := NewExecutionContext(ctx)
	var buf bytes.Buffer

	b.ResetTimer()
	for b.Loop() {
		buf.Reset()
		if err := tmpl.Execute(execCtx, &buf); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExecuteWithConditional benchmarks template execution with conditionals.
func BenchmarkExecuteWithConditional(b *testing.B) {
	tmpl, err := Compile(`{% if age > 18 %}Adult{% else %}Minor{% endif %}`)
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewContext()
	ctx.Set("age", 25)
	execCtx := NewExecutionContext(ctx)
	var buf bytes.Buffer

	b.ResetTimer()
	for b.Loop() {
		buf.Reset()
		if err := tmpl.Execute(execCtx, &buf); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRender benchmarks one-shot rendering.
func BenchmarkRender(b *testing.B) {
	source := `Hello {{ name }}!`
	data := map[string]any{"name": "World"}

	for b.Loop() {
		_, err := Render(source, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValueConversion benchmarks Value type conversions.
func BenchmarkValueConversion(b *testing.B) {
	v := NewValue(42)

	for b.Loop() {
		_, _ = v.Int()
		_, _ = v.Float()
		_ = v.String()
		_ = v.Bool()
	}
}

// BenchmarkValueIterate benchmarks Value iteration over slices.
func BenchmarkValueIterate(b *testing.B) {
	v := NewValue([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	for b.Loop() {
		_ = v.Iterate(func(i, count int, key, val *Value) bool {
			return true
		})
	}
}

// BenchmarkContextGet benchmarks Context.Get with dot notation.
func BenchmarkContextGet(b *testing.B) {
	ctx := NewContext()
	ctx.Set("user.address.city", "New York")

	for b.Loop() {
		_, _ = ctx.Get("user.address.city")
	}
}

// BenchmarkContextSet benchmarks Context.Set with dot notation.
func BenchmarkContextSet(b *testing.B) {
	ctx := NewContext()

	for b.Loop() {
		ctx.Set("user.address.city", "New York")
	}
}

// BenchmarkFilterApplication benchmarks filter application.
func BenchmarkFilterApplication(b *testing.B) {
	filter, _ := GetFilter("upper")

	for b.Loop() {
		_, err := filter("hello world", nil...)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComplexTemplate benchmarks a complex real-world template.
func BenchmarkComplexTemplate(b *testing.B) {
	source := `
<!DOCTYPE html>
<html>
<head><title>{{ title }}</title></head>
<body>
	<h1>{{ heading }}</h1>
	{% if user %}
		<p>Welcome, {{ user.name | capitalize }}!</p>
		{% if user.age > 18 %}
			<p>You are an adult.</p>
		{% else %}
			<p>You are a minor.</p>
		{% endif %}
	{% endif %}
	<ul>
	{% for item in items %}
		<li>{{ item.name }}: {{ item.price | round: 2 }}</li>
	{% endfor %}
	</ul>
</body>
</html>
`
	tmpl, err := Compile(source)
	if err != nil {
		b.Fatal(err)
	}

	ctx := NewContext()
	ctx.Set("title", "My Store")
	ctx.Set("heading", "Welcome to My Store")
	ctx.Set("user.name", "john doe")
	ctx.Set("user.age", 25)
	ctx.Set("items", []map[string]any{
		{"name": "Apple", "price": 1.234},
		{"name": "Banana", "price": 0.567},
		{"name": "Orange", "price": 2.345},
	})
	execCtx := NewExecutionContext(ctx)
	var buf bytes.Buffer

	b.ResetTimer()
	for b.Loop() {
		buf.Reset()
		if err := tmpl.Execute(execCtx, &buf); err != nil {
			b.Fatal(err)
		}
	}
}
