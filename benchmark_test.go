package template

import (
	"bytes"
	"testing"
)

type benchmarkFlatUser struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int    `json:"age"`
	Active bool   `json:"active"`
	Role   string `json:"role"`
}

type benchmarkProfile struct {
	Bio     string `json:"bio"`
	Website string `json:"website"`
	City    string `json:"city"`
}

type benchmarkNestedUser struct {
	Name    string           `json:"name"`
	Email   string           `json:"email"`
	Profile benchmarkProfile `json:"profile"`
	Tags    []string         `json:"tags"`
}

// BenchmarkParseString benchmarks source-string parsing.
func BenchmarkParseString(b *testing.B) {
	source := `Hello {{ name }}! {% if age > 18 %}You are an adult.{% endif %}`
	engine := New()
	for b.Loop() {
		_, err := engine.ParseString(source)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExecuteSimple benchmarks simple template execution.
func BenchmarkExecuteSimple(b *testing.B) {
	engine := New()
	tmpl, err := engine.ParseString(`Hello {{ name }}!`)
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewData()
	ctx.Set("name", "World")
	execCtx := NewRenderContext(ctx)
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
	engine := New()
	tmpl, err := engine.ParseString(`{{ name | upper | append: "!" }}`)
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewData()
	ctx.Set("name", "world")
	execCtx := NewRenderContext(ctx)
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
	engine := New()
	tmpl, err := engine.ParseString(`{% for item in items %}{{ item }}{% endfor %}`)
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewData()
	ctx.Set("items", []string{"a", "b", "c", "d", "e"})
	execCtx := NewRenderContext(ctx)
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
	engine := New()
	tmpl, err := engine.ParseString(`{% if age > 18 %}Adult{% else %}Minor{% endif %}`)
	if err != nil {
		b.Fatal(err)
	}
	ctx := NewData()
	ctx.Set("age", 25)
	execCtx := NewRenderContext(ctx)
	var buf bytes.Buffer

	b.ResetTimer()
	for b.Loop() {
		buf.Reset()
		if err := tmpl.Execute(execCtx, &buf); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseAndRender benchmarks one-shot parse+render.
func BenchmarkParseAndRender(b *testing.B) {
	source := `Hello {{ name }}!`
	data := map[string]any{"name": "World"}
	engine := New()

	for b.Loop() {
		tpl, err := engine.ParseString(source)
		if err != nil {
			b.Fatal(err)
		}
		_, err = tpl.Render(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEngineLoadCacheHit(b *testing.B) {
	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `Hello {{ name }}`,
		})),
		WithFormat(FormatText),
	)

	if _, err := engine.Load("a.txt"); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		if _, err := engine.Load("a.txt"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEngineLoadCacheMissAfterReset(b *testing.B) {
	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.txt": `Hello {{ name }}`,
		})),
		WithFormat(FormatText),
	)

	b.ResetTimer()
	for b.Loop() {
		engine.Reset()
		if _, err := engine.Load("a.txt"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEngineRenderLayout(b *testing.B) {
	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"page.html": `{% extends "base.html" %}{% block content %}{% include "card.html" with title=title %}{% endblock %}`,
			"base.html": `<html><body>{% block content %}{% endblock %}</body></html>`,
			"card.html": `<section>{{ title }}</section>`,
		})),
		WithFormat(FormatHTML),
		WithLayout(),
	)

	b.ResetTimer()
	for b.Loop() {
		if _, err := engine.Render("page.html", Data{"title": "<b>hello</b>"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEngineRenderCachedParallel(b *testing.B) {
	engine := New(
		WithLoader(NewMemoryLoader(map[string]string{
			"a.html": `{{ x }}-{{ y }}`,
		})),
		WithFormat(FormatHTML),
	)

	if _, err := engine.Load("a.html"); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := engine.Render("a.html", Data{"x": "left", "y": "<right>"}); err != nil {
				b.Fatal(err)
			}
		}
	})
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

// BenchmarkDataGet benchmarks Data.Get with dot notation.
func BenchmarkDataGet(b *testing.B) {
	ctx := NewData()
	ctx.Set("user.address.city", "New York")

	for b.Loop() {
		_, _ = ctx.Get("user.address.city")
	}
}

// BenchmarkDataSet benchmarks Data.Set with dot notation.
func BenchmarkDataSet(b *testing.B) {
	ctx := NewData()

	for b.Loop() {
		ctx.Set("user.address.city", "New York")
	}
}

func BenchmarkDataBuilderKeyValue(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		builder := NewDataBuilder().
			KeyValue("name", "Alice").
			KeyValue("email", "alice@example.com").
			KeyValue("age", 30).
			KeyValue("active", true).
			KeyValue("role", "admin")
		if _, err := builder.Build(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDataBuilderStructFlat(b *testing.B) {
	user := benchmarkFlatUser{
		Name:   "Alice",
		Email:  "alice@example.com",
		Age:    30,
		Active: true,
		Role:   "admin",
	}

	b.ResetTimer()
	for b.Loop() {
		if _, err := NewDataBuilder().Struct(user).Build(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDataBuilderStructNested(b *testing.B) {
	user := benchmarkNestedUser{
		Name:  "Alice",
		Email: "alice@example.com",
		Profile: benchmarkProfile{
			Bio:     "Engineer",
			Website: "https://example.com",
			City:    "Hong Kong",
		},
		Tags: []string{"go", "template", "bench"},
	}

	b.ResetTimer()
	for b.Loop() {
		if _, err := NewDataBuilder().Struct(user).Build(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFilterApplication benchmarks filter application.
func BenchmarkFilterApplication(b *testing.B) {
	filter, _ := defaultRegistry.Filter("upper")

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
	engine := New()
	tmpl, err := engine.ParseString(source)
	if err != nil {
		b.Fatal(err)
	}

	ctx := NewData()
	ctx.Set("title", "My Store")
	ctx.Set("heading", "Welcome to My Store")
	ctx.Set("user.name", "john doe")
	ctx.Set("user.age", 25)
	ctx.Set("items", []map[string]any{
		{"name": "Apple", "price": 1.234},
		{"name": "Banana", "price": 0.567},
		{"name": "Orange", "price": 2.345},
	})
	execCtx := NewRenderContext(ctx)
	var buf bytes.Buffer

	b.ResetTimer()
	for b.Loop() {
		buf.Reset()
		if err := tmpl.Execute(execCtx, &buf); err != nil {
			b.Fatal(err)
		}
	}
}
