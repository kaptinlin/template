package template

import (
	"fmt"
	"maps"
	"slices"
	"sync"
)

// TagParser is the parse function signature for template tags.
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
//   - Statement: The parsed statement node.
//   - error: Any parse error encountered.
type TagParser func(doc *Parser, start *Token, arguments *Parser) (Statement, error)

// TagRegistry stores tag parsers. An optional parent registry is
// consulted when a lookup misses, enabling an Engine to layer its own
// private tags over the global registry without copying entries.
//
// TagRegistry is safe for concurrent use.
type TagRegistry struct {
	mu     sync.RWMutex
	tags   map[string]TagParser
	parent *TagRegistry
}

// NewTagRegistry creates an empty [TagRegistry].
func NewTagRegistry() *TagRegistry {
	return &TagRegistry{tags: make(map[string]TagParser)}
}

func (r *TagRegistry) validate(name string, parser TagParser) {
	if parser == nil {
		panic(fmt.Sprintf("template: nil tag parser for %q", name))
	}
}

// Register adds a tag parser. Duplicate names return [ErrTagAlreadyRegistered].
func (r *TagRegistry) Register(name string, parser TagParser) error {
	r.validate(name, parser)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tags[name]; exists {
		return fmt.Errorf("%w: %q", ErrTagAlreadyRegistered, name)
	}
	r.tags[name] = parser
	return nil
}

// Replace stores parser under name, overwriting any direct existing entry.
//
// Replace panics if parser is nil.
func (r *TagRegistry) Replace(name string, parser TagParser) {
	r.validate(name, parser)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tags[name] = parser
}

// MustRegister registers parser under name and panics on duplicate
// registration or nil parser.
func (r *TagRegistry) MustRegister(name string, parser TagParser) {
	if err := r.Register(name, parser); err != nil {
		panic(err)
	}
}

// Get looks up a tag parser by name. If not found and a parent registry
// is set, the parent is consulted.
func (r *TagRegistry) Get(name string) (TagParser, bool) {
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
func (r *TagRegistry) Has(name string) bool {
	_, ok := r.Get(name)
	return ok
}

// List returns a sorted list of tag names registered directly in this
// registry (excluding parents).
func (r *TagRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Sorted(maps.Keys(r.tags))
}

// Unregister removes a tag from this registry (does not touch parents).
func (r *TagRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tags, name)
}

// Clone returns a shallow copy of the registry and its direct entries.
// The parent registry reference is preserved.
func (r *TagRegistry) Clone() *TagRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &TagRegistry{
		tags:   maps.Clone(r.tags),
		parent: r.parent,
	}
}

// defaultTagRegistry is the package-wide built-in tag registry used as the
// parent layer for engine-local registries.
var defaultTagRegistry = NewTagRegistry()

// Built-in tag registration.

// builtinTag pairs a tag name with its parser for deterministic registration.
type builtinTag struct {
	name   string
	parser TagParser
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
