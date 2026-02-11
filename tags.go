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

// Global tag registry.
var (
	tagRegistry = make(map[string]TagParser)
	tagMu       sync.RWMutex
)

// RegisterTag registers a tag parser.
// It is safe to call from multiple goroutines.
func RegisterTag(name string, parser TagParser) error {
	tagMu.Lock()
	defer tagMu.Unlock()

	if _, exists := tagRegistry[name]; exists {
		return fmt.Errorf("%w: %q", ErrTagAlreadyRegistered, name)
	}
	tagRegistry[name] = parser
	return nil
}

// Tag returns the parser registered for the given tag name.
// It is safe to call from multiple goroutines.
func Tag(name string) (TagParser, bool) {
	tagMu.RLock()
	defer tagMu.RUnlock()

	parser, ok := tagRegistry[name]
	return parser, ok
}

// ListTags returns a sorted list of all registered tag names.
// It is safe to call from multiple goroutines.
func ListTags() []string {
	tagMu.RLock()
	defer tagMu.RUnlock()

	return slices.Sorted(maps.Keys(tagRegistry))
}

// HasTag checks if a tag with the given name is registered.
// It is safe to call from multiple goroutines.
func HasTag(name string) bool {
	tagMu.RLock()
	defer tagMu.RUnlock()

	_, ok := tagRegistry[name]
	return ok
}

// UnregisterTag removes a tag from the registry.
// It is safe to call from multiple goroutines.
func UnregisterTag(name string) {
	tagMu.Lock()
	defer tagMu.Unlock()

	delete(tagRegistry, name)
}

// Built-in tag registration.

// builtinTag pairs a tag name with its parser for deterministic registration.
type builtinTag struct {
	name   string
	parser TagParser
}

// builtinTags lists all built-in tags in registration order.
var builtinTags = []builtinTag{
	{"if", parseIfTag},
	{"for", parseForTag},
	{"break", parseBreakTag},
	{"continue", parseContinueTag},
}

// init registers all built-in tags when the package is imported.
// Tag parsers are implemented in the corresponding tag_*.go files.
//
// Note: elif, else, and endif are not registered as independent tags.
// They are recognized only within if blocks by parseIfTag, which
// prevents them from appearing outside of if blocks.
func init() {
	for _, bt := range builtinTags {
		if err := RegisterTag(bt.name, bt.parser); err != nil {
			panic(err)
		}
	}
}
