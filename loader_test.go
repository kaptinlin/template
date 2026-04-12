package template

import (
	"errors"
	"testing"
)

// Phase A: Loader contract — name validation + happy path

func TestMemoryLoader_ValidPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"relative escape", "../a.html", ErrInvalidTemplateName},
		{"deep relative escape", "foo/../../bar", ErrInvalidTemplateName},
		{"absolute", "/etc/passwd", ErrInvalidTemplateName},
		{"backslash", "a\\b", ErrInvalidTemplateName},
		{"nul byte", "a\x00b", ErrInvalidTemplateName},
		{"trailing slash", "a/", ErrInvalidTemplateName},
		{"empty", "", ErrInvalidTemplateName},
	}

	loader := NewMemoryLoader(map[string]string{"a.html": "hi"})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := loader.Open(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Open(%q) err = %v, want %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestMemoryLoader_HappyPath(t *testing.T) {
	t.Parallel()

	loader := NewMemoryLoader(map[string]string{
		"a.html": "hello {{ name }}",
	})

	src, resolved, err := loader.Open("a.html")
	if err != nil {
		t.Fatalf("Open() err = %v", err)
	}
	if src != "hello {{ name }}" {
		t.Errorf("src = %q", src)
	}
	if resolved != "a.html" {
		t.Errorf("resolved = %q, want a.html", resolved)
	}
}

func TestMemoryLoader_NotFound(t *testing.T) {
	t.Parallel()

	loader := NewMemoryLoader(map[string]string{})
	_, _, err := loader.Open("missing.html")
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("err = %v, want ErrTemplateNotFound", err)
	}
}

// Phase M cycle 1: ChainLoader returns the first loader that has the file.
func TestChainLoader_FirstHitWins(t *testing.T) {
	t.Parallel()

	user := NewMemoryLoader(map[string]string{"x.txt": "USER"})
	theme := NewMemoryLoader(map[string]string{"x.txt": "THEME"})
	chain := NewChainLoader(user, theme)

	src, _, err := chain.Open("x.txt")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if src != "USER" {
		t.Errorf("src = %q, want USER", src)
	}
}

// Phase M cycle 2: ChainLoader falls through to the next loader on miss.
func TestChainLoader_FallthroughOnMiss(t *testing.T) {
	t.Parallel()

	user := NewMemoryLoader(map[string]string{})
	theme := NewMemoryLoader(map[string]string{"x.txt": "THEME"})
	chain := NewChainLoader(user, theme)

	src, _, err := chain.Open("x.txt")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if src != "THEME" {
		t.Errorf("src = %q, want THEME", src)
	}
}

// Phase M cycle 3: all misses return ErrTemplateNotFound.
func TestChainLoader_AllMiss_NotFound(t *testing.T) {
	t.Parallel()

	chain := NewChainLoader(
		NewMemoryLoader(map[string]string{}),
		NewMemoryLoader(map[string]string{}),
	)
	_, _, err := chain.Open("x.txt")
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("err = %v", err)
	}
}

// Phase M cycle 4: resolved name carries layer info so same-name files
// in different layers don't collide in cache.
func TestChainLoader_ResolvedNamePrefixDistinguishesLayers(t *testing.T) {
	t.Parallel()

	user := NewMemoryLoader(map[string]string{"x.txt": "USER"})
	theme := NewMemoryLoader(map[string]string{"x.txt": "THEME"})

	// User loader wins when present.
	chain1 := NewChainLoader(user, theme)
	_, r1, _ := chain1.Open("x.txt")

	// Theme loader wins when user is empty.
	chain2 := NewChainLoader(
		NewMemoryLoader(map[string]string{}),
		theme,
	)
	_, r2, _ := chain2.Open("x.txt")

	// The two resolved names should differ so their cache entries do
	// not collide.
	if r1 == r2 {
		t.Errorf("resolved names collided: %q == %q", r1, r2)
	}
}

// Phase M cycle 5: empty chain returns ErrTemplateNotFound.
func TestChainLoader_EmptyChain_NotFound(t *testing.T) {
	t.Parallel()

	chain := NewChainLoader()
	_, _, err := chain.Open("x.txt")
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("err = %v", err)
	}
}

// Phase M cycle 6: Engine using ChainLoader caches keyed on resolved name
// so the same user/theme layers do not clobber each other.
func TestChainLoader_EngineCacheKeyedOnResolvedName(t *testing.T) {
	t.Parallel()

	user := NewMemoryLoader(map[string]string{
		"a.txt": "USER",
	})
	theme := NewMemoryLoader(map[string]string{
		"a.txt": "THEME",
	})
	engine := New(
		WithLoader(NewChainLoader(user, theme)),
		WithFormat(FormatText),
		WithLayout(),
	)

	got, err := engine.Render("a.txt", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "USER" {
		t.Errorf("got %q, want USER", got)
	}
}
