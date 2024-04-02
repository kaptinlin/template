package template

import (
	"regexp"
	"strconv"
	"strings"
)

// Regular expression to identify variables.
var variableRegex = regexp.MustCompile(`{{\s*([\w\.]+)((?:\s*\|\s*[\w\:\,]+(?:\s*:\s*[^}]+)?)*)\s*}}`)

// Parser analyzes template syntax.
type Parser struct{}

// NewParser creates a Parser with a compiled regular expression for efficiency.
func NewParser() *Parser {
	return &Parser{}
}

// Parse transforms a template string into a Template.
func (p *Parser) Parse(src string) (*Template, error) {
	template := NewTemplate()
	tokens := p.tokenize(src)
	for _, token := range tokens {
		if p.isVariable(token) {
			p.addVariableNode(token, template)
		} else {
			p.addTextNode(token, template)
		}
	}
	return template, nil
}

// tokenize divides the source string into tokens for easier parsing.
func (p *Parser) tokenize(src string) []string {
	var tokens []string
	matches := variableRegex.FindAllStringIndex(src, -1)
	start := 0
	for _, match := range matches {
		// Add text between variables as tokens
		if start < match[0] {
			tokens = append(tokens, src[start:match[0]])
		}
		// Add variable token
		tokens = append(tokens, src[match[0]:match[1]])
		start = match[1]
	}
	// Add remaining text as a token
	if start < len(src) {
		tokens = append(tokens, src[start:])
	}
	return tokens
}

// isVariable checks if a token represents a variable.
func (p *Parser) isVariable(token string) bool {
	return strings.HasPrefix(token, "{{") && strings.HasSuffix(token, "}}")
}

// Updated addVariableNode processes a variable token, parses out any filters, and adds it to the template.
func (p *Parser) addVariableNode(token string, tpl *Template) {
	// Extract the inner content of the variable token.
	innerContent := strings.TrimSpace(token[2 : len(token)-2])
	// Split the variable name from any filters.
	parts := strings.SplitN(innerContent, "|", 2)

	varName := strings.TrimSpace(parts[0])

	// Initialize filters slice.
	var filters []Filter

	// Check if there are filters to parse and use parseFilters if so.
	if len(parts) > 1 {
		filters = parseFilters(parts[1])
	}

	// Create a new variable node with the parsed variable name and filters.
	node := &Node{
		Type:     "variable",
		Variable: varName,
		Filters:  filters,
		Text:     token,
	}

	// Add the new node to the template.
	tpl.Nodes = append(tpl.Nodes, node)
}
func parseFilters(filterStr string) []Filter {
	var filters []Filter
	if filterStr == "" {
		return filters
	}

	// Splitting the entire filter string into individual filters
	filterParts := strings.Split(filterStr, "|")
	for _, part := range filterParts {
		partTrimmed := strings.TrimSpace(part)
		if partTrimmed == "" {
			continue
		}

		// Splitting the filter name from its arguments
		nameArgs := strings.SplitN(partTrimmed, ":", 2)
		filter := Filter{Name: strings.TrimSpace(nameArgs[0])}

		// Handling arguments, if present
		if len(nameArgs) == 2 {
			filter.Args = splitArgsConsideringQuotes(nameArgs[1])
		}

		filters = append(filters, filter)
	}

	return filters
}

func splitArgsConsideringQuotes(argsStr string) []FilterArg {
	var args []FilterArg
	var currentArg strings.Builder
	var inQuotes bool
	var quoteChar rune

	appendArg := func() {
		arg := currentArg.String()
		currentArg.Reset()
		if arg == `""` || arg == "''" {
			args = append(args, StringArg{val: ""})
		} else if trimmedArg := strings.TrimSpace(arg); len(trimmedArg) > 0 {
			if len(trimmedArg) >= 2 && (trimmedArg[0] == '"' || trimmedArg[0] == '\'') {
				// String argument
				args = append(args, StringArg{val: trimmedArg[1 : len(trimmedArg)-1]})
			} else {
				// Check if it's a number
				if number, err := strconv.ParseFloat(trimmedArg, 64); err == nil {
					args = append(args, NumberArg{val: number})
				} else {
					// Treat as variable
					args = append(args, VariableArg{name: trimmedArg})
				}
			}
		}
	}

	for i, char := range argsStr {
		switch {
		case char == '"' || char == '\'':
			if inQuotes && char == quoteChar {
				currentArg.WriteRune(char) // Include the quote for simplicity
				inQuotes = false
				if i == len(argsStr)-1 || argsStr[i+1] == ',' {
					appendArg()
				}
			} else if !inQuotes {
				inQuotes = true
				quoteChar = char
				currentArg.WriteRune(char) // Include the quote for parsing
			}
		case char == ',' && !inQuotes:
			appendArg()
		default:
			currentArg.WriteRune(char)
		}
	}

	if currentArg.Len() > 0 || inQuotes {
		appendArg()
	}

	return args
}

// addTextNode adds a text token to the template
func (p *Parser) addTextNode(text string, tpl *Template) {
	if text != "" {
		tpl.Nodes = append(tpl.Nodes, &Node{Type: "text", Text: text})
	}
}
