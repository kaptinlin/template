package template

// SafeString marks a string value as pre-escaped or otherwise trusted
// HTML content. When a template rendered via [NewHTMLSet] outputs a
// SafeString, the auto-escaper bypasses escaping. In [NewTextSet] or
// [Compile]-path templates, SafeString behaves identically to a plain
// string.
//
// Producing a SafeString from untrusted input is a security bug: the
// contents are emitted verbatim into HTML.
type SafeString string

// safeFilter tags its input as a SafeString so the HTML auto-escaper
// skips it. Nil input becomes the empty SafeString.
//
// Usage: {{ rendered_markdown | safe }}
func safeFilter(value any, _ ...any) (any, error) {
	if value == nil {
		return SafeString(""), nil
	}
	if s, ok := value.(SafeString); ok {
		return s, nil
	}
	return SafeString(toString(value)), nil
}
