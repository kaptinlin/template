package template

import (
	"fmt"
	"maps"
	"slices"
	"sync"
)

// tagParser is the parse function signature for template tags.
//
// Parameters:
//   - doc: The document-level parser used to parse nested tag bodies.
//     Example: an if tag parses content between {% if %} and {% endif %}.
//   - start: The tag-name token (for example "if", "for"), including source
//     position information used in error reporting.
//   - arguments: A dedicated parser for tag arguments.
//     Example: in {% if x > 5 %}, this parser sees "x > 5".
//
// Returns:
//   - statement: The parsed statement node.
//   - error: Any parse error encountered.
type tagParser func(doc *parser, start *token, arguments *parser) (statement, error)

// tagRegistry stores tag parsers. An optional parent registry is
// consulted when a lookup misses, enabling an Engine to layer its own
// private tags over the global registry without copying entries.
//
// tagRegistry is safe for concurrent use.
type tagRegistry struct {
	mu     sync.RWMutex
	tags   map[string]tagParser
	parent *tagRegistry
}

// newTagRegistry creates an empty [tagRegistry].
func newTagRegistry() *tagRegistry {
	return &tagRegistry{tags: make(map[string]tagParser)}
}

func (r *tagRegistry) validate(name string, parser tagParser) {
	if parser == nil {
		panic(fmt.Sprintf("template: nil tag parser for %q", name))
	}
}

// Register adds a tag parser. Duplicate names return [errTagAlreadyRegistered].
func (r *tagRegistry) Register(name string, parser tagParser) error {
	r.validate(name, parser)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tags[name]; exists {
		return fmt.Errorf("%w: %q", errTagAlreadyRegistered, name)
	}
	r.tags[name] = parser
	return nil
}

// Replace stores parser under name, overwriting any direct existing entry.
//
// Replace panics if parser is nil.
func (r *tagRegistry) Replace(name string, parser tagParser) {
	r.validate(name, parser)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tags[name] = parser
}

// MustRegister registers parser under name and panics on duplicate
// registration or nil parser.
func (r *tagRegistry) MustRegister(name string, parser tagParser) {
	if err := r.Register(name, parser); err != nil {
		panic(err)
	}
}

// Get looks up a tag parser by name. If not found and a parent registry
// is set, the parent is consulted.
func (r *tagRegistry) Get(name string) (tagParser, bool) {
	r.mu.RLock()
	fn, ok := r.tags[name]
	r.mu.RUnlock()
	if ok {
		return fn, true
	}
	if r.parent != nil {
		return r.parent.Get(name)
	}
	return nil, false
}

// Has reports whether a tag is registered (including in ancestors).
func (r *tagRegistry) Has(name string) bool {
	_, ok := r.Get(name)
	return ok
}

// List returns a sorted list of tag names registered directly in this
// registry (excluding parents).
func (r *tagRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Sorted(maps.Keys(r.tags))
}

// Unregister removes a tag from this registry (does not touch parents).
func (r *tagRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tags, name)
}

// Clone returns a shallow copy of the registry and its direct entries.
// The parent registry reference is preserved.
func (r *tagRegistry) Clone() *tagRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &tagRegistry{
		tags:   maps.Clone(r.tags),
		parent: r.parent,
	}
}

// defaultTagRegistry is the package-wide built-in tag registry used as the
// parent layer for engine-local registries.
var defaultTagRegistry = newTagRegistry()

// Built-in tag registration.

// builtinTag pairs a tag name with its parser for deterministic registration.
type builtinTag struct {
	name   string
	parser tagParser
}

// builtinTags lists tags that are registered into the built-in registry
// when the package is imported. These are the tags available in every
// engine by default.
//
// Layout tags (include, extends, block) are intentionally NOT in this
// list. They are registered per-engine only when FeatureLayout is enabled.
var builtinTags = []builtinTag{
	{"if", parseIfTag},
	{"for", parseForTag},
	{"break", parseBreakTag},
	{"continue", parseContinueTag},
}

// layoutTags are the tags that become available only after an engine
// enables FeatureLayout.
var layoutTags = []builtinTag{
	{"include", parseIncludeTag},
	{"extends", parseExtendsTag},
	{"block", parseBlockTag},
}

// init registers all built-in (non-layout) tags when the package is
// imported. Tag parsers are implemented in the corresponding tag_*.go
// files.
//
// Note: elif, else, and endif are not registered as independent tags.
// They are recognized only within if blocks by parseIfTag, which
// prevents them from appearing outside of if blocks.
func init() {
	for _, bt := range builtinTags {
		defaultTagRegistry.MustRegister(bt.name, bt.parser)
	}
}
