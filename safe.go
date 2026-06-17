package template

// SafeHTML marks a string value as pre-escaped or otherwise trusted
// HTML content. When a template rendered by an engine with [FormatHTML]
// outputs a SafeHTML, the auto-escaper bypasses escaping. In text-format
// engine renders or source-string templates without HTML format, SafeHTML
// behaves identically to a plain string.
//
// Producing a SafeHTML from untrusted input is a security bug: the
// contents are emitted verbatim into HTML.
type SafeHTML string

// safeFilter tags its input as a SafeHTML so the HTML auto-escaper
// skips it. Nil input becomes the empty SafeHTML.
//
// Usage: {{ rendered_markdown | safe }}
func safeFilter(value any, _ ...any) (any, error) {
	if value == nil {
		return SafeHTML(""), nil
	}
	if s, ok := value.(SafeHTML); ok {
		return s, nil
	}
	return SafeHTML(toString(value)), nil
}
