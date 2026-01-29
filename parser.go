package template

import (
	"regexp"
	"strconv"
	"strings"
)

// Regular expression to identify variables.
var variableRegex = regexp.MustCompile(
	`{{\s*(?:'[^']*'|"[\s\S]*?"|[\w\.]+)((?:\s*\|\s*[\w\:\,]+(?:\s*:\s*[^}]+)?)*)\s*}}`,
)
var ifRegex = regexp.MustCompile(`{%\s*if\s+` + // if start
	`([^%]+)` + // Capture any non-% characters (expression part)
	`\s*%}`) // if end
var endforRegex = regexp.MustCompile(`{%\s*endfor\s*%}`)
var endifRegex = regexp.MustCompile(`{%\s*endif\s*%}`)
var elseRegex = regexp.MustCompile(`{%\s*else\s*%}`)
var elifRegex = regexp.MustCompile(`{%\s*elif\s+` + // elif start
	`([^%]+)` + // Capture any non-% characters (expression part)
	`\s*%}`) // elif end
var forRegex = regexp.MustCompile(
	`{%\s*for\s+` +
		`(` +
		`([\w\.]+)` +
		`|` +
		`([\w\.]+)\s*,\s*([\w\.]+)` +
		`)` +
		`\s+in\s+` +
		`([\w\.$$$$'"]+)` +
		`\s*%}`,
)

// Regular expressions for break/continue control flow
var breakRegex = regexp.MustCompile(`{%\s*break\s*%}`)
var continueRegex = regexp.MustCompile(`{%\s*continue\s*%}`)

// Template delimiter constants
const (
	templateVarOpenLen   = 2 // length of "{{"
	templateVarCloseLen  = 2 // length of "}}"
	templateCtrlOpenLen  = 2 // length of "{%"
	templateCtrlCloseLen = 2 // length of "%}"
)

// Node type identifiers for control structures
const (
	nodeTypeFor      = 1
	nodeTypeIf       = 2
	nodeTypeElse     = 3
	nodeTypeEndFor   = 4
	nodeTypeEndIf    = 5
	nodeTypeBreak    = 6
	nodeTypeContinue = 7
	nodeTypeElif     = 8
)

// Parser analyzes template syntax.
type Parser struct{}

// NewParser creates a Parser with a compiled regular expression for efficiency.
func NewParser() *Parser {
	return &Parser{}
}

// addNode is a generic helper to add a node to either the template root or a parent node
func (p *Parser) addNode(newNode *Node, template *Template, parent *Node) {
	if template != nil {
		template.Nodes = append(template.Nodes, newNode)
	} else if parent != nil {
		parent.Children = append(parent.Children, newNode)
	}
}

// extractFiltersFromToken extracts variable name and filters from a variable token
func (p *Parser) extractFiltersFromToken(token string) (string, []Filter) {
	innerContent := strings.TrimSpace(token[templateVarOpenLen : len(token)-templateVarCloseLen])
	varNameRaw, filtersStr, hasFilters := strings.Cut(innerContent, "|")

	varName := strings.TrimSpace(varNameRaw)
	var filters []Filter

	if hasFilters {
		filters = parseFilters(filtersStr)
	}

	return varName, filters
}

// Updated addVariableNode processes a variable token, parses out any filters, and adds it to the template.
func (p *Parser) addVariableNode(token string, template *Template, node *Node) {
	varName, filters := p.extractFiltersFromToken(token)

	varNode := &Node{
		Type:     "variable",
		Variable: varName,
		Filters:  filters,
		Text:     token,
	}

	p.addNode(varNode, template, node)
}

func parseFilters(filterStr string) []Filter {
	filters := make([]Filter, 0)

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

		// Splitting the filter name from its arguments using strings.Cut (faster than SplitN)
		filterName, argsStr, hasArgs := strings.Cut(partTrimmed, ":")
		filter := Filter{Name: strings.TrimSpace(filterName)}

		// Handling arguments, if present
		if hasArgs {
			filter.Args = splitArgsConsideringQuotes(argsStr)
		}

		filters = append(filters, filter)
	}

	return filters
}

func splitArgsConsideringQuotes(argsStr string) []FilterArg {
	var args []FilterArg
	var currentArg strings.Builder
	currentArg.Grow(len(argsStr))
	var inQuotes bool
	var quoteChar rune

	appendArg := func() {
		arg := currentArg.String()
		currentArg.Reset()
		arg = strings.TrimSpace(arg)
		if arg == "" {
			return
		}

		if (strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"")) ||
			(strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'")) {
			if len(arg) >= 2 {
				args = append(args, StringArg{val: arg[1 : len(arg)-1]})
			} else {
				args = append(args, StringArg{val: ""})
			}
		} else if number, err := strconv.ParseFloat(arg, 64); err == nil {
			args = append(args, NumberArg{val: number})
		} else {
			args = append(args, VariableArg{name: arg})
		}
	}

	for i, char := range argsStr {
		switch {
		case (char == '"' || char == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = char
			currentArg.WriteRune(char)
		case inQuotes && char == quoteChar:
			inQuotes = false
			currentArg.WriteRune(char)
		case char == ',' && !inQuotes:
			appendArg()
		default:
			currentArg.WriteRune(char)
		}

		// If last char, append
		if i == len(argsStr)-1 {
			appendArg()
		}
	}

	return args
}

// addTextNode adds a text token to the template
func (p *Parser) addTextNode(text string, template *Template, node *Node) {
	if text == "" {
		return
	}
	p.addNode(&Node{Type: "text", Text: text}, template, node)
}

// addForNode adds a for node to the template
func (p *Parser) addForNode(text string, template *Template, parent *Node) {
	// If it's a start tag, create the For node immediately
	if forRegex.MatchString(text) {
		newNode := &Node{Type: "for", Text: text}
		analyzeForParameter(text, newNode)
		p.addNode(newNode, template, parent)
		return
	}

	// If it's an end tag, we expect the parent node to already be a For node
	if endforRegex.MatchString(text) && parent != nil && parent.Type == "for" {
		parent.EndText = text
	}
}

// analyzeForParameter analyzes the for parameter using regex
func analyzeForParameter(text string, node *Node) {
	matches := forRegex.FindStringSubmatch(text)
	if len(matches) >= 6 {
		node.Variable = strings.TrimSpace(matches[1])
		node.Collection = matches[5]
	}
}

// addIfNode adds an if node to the template
func (p *Parser) addIfNode(text string, template *Template, parent *Node) {
	// If it's a start tag, create the If node immediately
	if ifRegex.MatchString(text) {
		newNode := &Node{Type: "if", Text: text}
		p.addNode(newNode, template, parent)
		return
	}

	// If it's an end tag, we expect the parent node to already be an If node
	if endifRegex.MatchString(text) && parent != nil && parent.Type == "if" {
		parent.EndText = text
	}
}

// addControlFlowNode adds a control flow node (break/continue) to the template
func (p *Parser) addControlFlowNode(text string, template *Template, node *Node, nodeType string) {
	p.addNode(&Node{Type: nodeType, Text: text}, template, node)
}

// addConditionalBranchNode adds a conditional branch node (else/elif) to the template
func (p *Parser) addConditionalBranchNode(text string, template *Template, node *Node, nodeType string) *Node {
	newNode := &Node{Type: nodeType, Text: text}
	p.addNode(newNode, template, node)
	return newNode
}

// Parse converts a template string into a Template object.
// It recognizes the following syntax:
// - Text content
// - Variable expressions {{ variable }}
// - Control structures {% if/for ... %}{% endif/endfor %}
// Returns the parsed template and any error encountered.
func (p *Parser) Parse(src string) (*Template, error) {
	n := len(src)
	template := NewTemplate()
	prev := 0
	next := 0

	for next < n {
		switch {
		case src[next] == '{' && next+1 < n && src[next+1] == '{':
			// Handle variable
			next, prev = p.handleVariable(src, next, prev, template, nil)
		case next+1 < n && src[next] == '{' && src[next+1] == '%':
			// Handle for and if
			next, prev = p.parseControlBlock(src, next, prev, template)
		case next+1 < n && src[next] == '{' && src[next+1] == '#':
			// Handle explanatory note
			next, prev = p.parseExplanatoryNote(src, next, prev, template)
		default:
			next++
		}
	}
	// Handle any remaining text
	p.parseRemainingText(src, prev, next, template)

	return template, nil
}

// parseControlBlock processes control structures (if/for blocks) in the template.
// It handles the parsing of opening tags, their content, and closing tags,
// maintaining the proper nesting structure in the template tree.
// Returns updated next and prev positions in the source string.
func (p *Parser) parseControlBlock(src string, next, prev int, template *Template) (int, int) {
	n := len(src)
	// Try to match control structure opening tags and determine their type
	matched, temp, typ := p.matchAppropriateStrings(src, n, next+2, "%}")
	if matched && typ <= nodeTypeElif {
		// Extract the complete control structure token
		token := src[next : temp+1]
		// Handle any text content before the control structure
		if next > prev {
			token := src[prev:next]
			p.addTextNode(token, template, nil)
		}

		// Add appropriate node based on control structure type
		switch typ {
		case nodeTypeFor:
			p.addForNode(token, template, nil)
		case nodeTypeIf:
			p.addIfNode(token, template, nil)
		case nodeTypeBreak:
			p.addControlFlowNode(token, template, nil, NodeTypeBreak)
		case nodeTypeContinue:
			p.addControlFlowNode(token, template, nil, NodeTypeContinue)
		}

		// Handle control structures that need end tags (for/if)
		if typ < nodeTypeElse {
			// Get the last added node to process its children
			node := template.Nodes[len(template.Nodes)-1]
			// Parse the content between opening and closing tags
			tempPrev, tempNext := p.parser(src, temp+1, typ, node)
			tempToken := src[tempPrev : tempNext+1]

			// Process the closing tag of the control structure
			switch typ {
			case nodeTypeFor:
				p.addForNode(tempToken, nil, node)
			case nodeTypeIf:
				p.addIfNode(tempToken, nil, node)
			}

			// Update position markers to continue parsing after this block
			next = tempNext + 1
			prev = tempNext + 1
		} else {
			// For break/continue, just move past the current token
			next = temp + 1
			prev = temp + 1
		}
	} else {
		// If no match found, move to next character
		next++
	}
	return next, prev
}

func (p *Parser) parseRemainingText(src string, prev, next int, template *Template) {
	if next > prev {
		token := src[prev:next]
		p.addTextNode(token, template, nil)
	}
}

// matchAppropriateStrings matches specific syntax structures in the template.
// It handles two main syntax formats:
// - Control structure tags {% ... %}
// - Variable expression tags {{ ... }}
// Returns:
// - bool: whether the match was successful
// - int: ending position of the match
// - int: syntax type (0:variable, 1:for, 2:if)
func (p *Parser) matchAppropriateStrings(src string, n int, next int, format string) (bool, int, int) {
	// Store starting position (excluding the two opening characters)
	temp := next - 2

	switch format {
	case "%}":
		// Handle control structure cases
		for next+1 < n {
			// Look for "%}" ending tag
			if src[next] == '%' && src[next+1] == '}' {
				// Determine if it's a for or if statement
				matched, typ := p.judgeBranchingStatements(src, temp, next+2)
				if matched {
					return true, next + 1, typ
				}
				return false, 0, 0
			}
			next++
		}
	case "}}":
		// Handle variable expression cases
		for next+1 < n {
			// Look for "}}" ending tag
			if src[next] == '}' && src[next+1] == '}' {
				// Extract complete variable expression and validate format
				token := src[temp : next+2]
				matched := variableRegex.MatchString(token)
				return matched, next + 1, 0
			}
			next++
		}
	case "#}":
		for next+1 < n {
			if src[next] == '#' && src[next+1] == '}' {
				return true, next + 1, 0
			}
			next++
		}
	default:
		// Return failure for unsupported formats
		return false, 0, 0
	}
	// Return failure if no match found
	return false, 0, 0
}

// judgeBranchingStatements determines the type of control structure by matching against predefined regex patterns.
// It checks for the following control structures:
// - for/endfor statements (type 1/4)
// - if/endif statements (type 2/5)
// - else statements (type 3)
// - break statements (type 6)
// - continue statements (type 7)
// Returns:
// - bool: whether a valid control structure was matched
// - int: the type of control structure (1-7) or 0 if no match
func (p *Parser) judgeBranchingStatements(src string, temp int, next int) (bool, int) {
	token := src[temp:next]
	switch {
	case forRegex.MatchString(token):
		return true, nodeTypeFor
	case ifRegex.MatchString(token):
		return true, nodeTypeIf
	case elseRegex.MatchString(token):
		return true, nodeTypeElse
	case elifRegex.MatchString(token):
		return true, nodeTypeElif
	case endforRegex.MatchString(token):
		return true, nodeTypeEndFor
	case endifRegex.MatchString(token):
		return true, nodeTypeEndIf
	case breakRegex.MatchString(token):
		return true, nodeTypeBreak
	case continueRegex.MatchString(token):
		return true, nodeTypeContinue
	default:
		return false, 0
	}
}

// parser processes the content within control structure blocks, including nested variables,
// control structures, and plain text. It recursively handles all nested structures until
// it encounters the corresponding end tag.
// Parameters:
// - src: source template string
// - prev: current starting position
// - typ: current control structure type
// - node: current node being processed
// Returns:
// - ending position and position of matched end tag
func (p *Parser) parser(src string, prev int, typ int, node *Node) (int, int) {
	n := len(src)
	next := prev
	// Adjust type value to match end tags (e.g., 1->4 means for->endfor)
	typ += 3
	// Flag to track if we've entered an else branch
	markEnterElse := false
	// Store reference to if node for else branch handling
	enterIfNode := node

	for next < n {
		switch {
		case src[next] == '{' && next+1 < n && src[next+1] == '{':
			// Handle variable expressions
			next, prev = p.handleVariable(src, next, prev, nil, node)
		case next+1 < n && src[next] == '{' && src[next+1] == '%':
			// Handle control structures
			matched, temp, tempType := p.matchAppropriateStrings(src, n, next+2, "%}")
			switch {
			case matched && tempType < nodeTypeElse:
				next, prev = p.handleControlStructure(src, next, prev, temp, tempType, node)
			case matched && tempType == nodeTypeElse && typ == nodeTypeEndIf:
				if markEnterElse {
					// We're in an elif/else branch and found another else
					// Add any remaining text to current branch and return to if node
					if next > prev {
						token := src[prev:next]
						p.addTextNode(token, nil, node)
					}
					next, prev, markEnterElse, node = p.handleConditionalBranch(src, next, prev, temp, nodeTypeElse, enterIfNode)
				} else {
					next, prev, markEnterElse, node = p.handleConditionalBranch(src, next, prev, temp, nodeTypeElse, node)
				}
			case matched && tempType == nodeTypeElif && typ == nodeTypeEndIf:
				if markEnterElse {
					// We're in an elif/else branch and found another elif
					// Add any remaining text to current branch and return to if node
					if next > prev {
						token := src[prev:next]
						p.addTextNode(token, nil, node)
					}
					next, prev, markEnterElse, node = p.handleConditionalBranch(src, next, prev, temp, nodeTypeElif, enterIfNode)
				} else {
					next, prev, markEnterElse, node = p.handleConditionalBranch(src, next, prev, temp, nodeTypeElif, node)
				}
			case matched && (tempType == nodeTypeBreak || tempType == nodeTypeContinue):
				// Handle break/continue control flow
				token := src[next : temp+1]
				if next > prev {
					textToken := src[prev:next]
					p.addTextNode(textToken, nil, node)
				}
				// Add break or continue node
				if tempType == nodeTypeBreak {
					p.addControlFlowNode(token, nil, node, NodeTypeBreak)
				} else {
					p.addControlFlowNode(token, nil, node, NodeTypeContinue)
				}
				next = temp + 1
				prev = temp + 1
			case matched && tempType == typ:
				if next > prev {
					token := src[prev:next]
					p.addTextNode(token, nil, node)
				}
				if markEnterElse {
					node = enterIfNode //nolint:staticcheck,ineffassign
				}
				return next, temp
			default:
				next++
			}
		case next+1 < n && src[next] == '{' && src[next+1] == '#':
			next, prev = p.handleComment(src, n, next, prev, node)
		default:
			next++
		}
	}
	// No matching end tag found
	return 0, 0
}

// handleControlStructure processes nested for/if control structures within a parent node.
// It handles both the opening and closing tags, and processes all content in between.
// Parameters:
// - src: source template string
// - next, prev: current positions in the source
// - temp: position of the end of opening tag
// - typ: type of control structure (1:for, 2:if)
// - node: parent node where this structure belongs
// Returns updated next and prev positions
func (p *Parser) handleControlStructure(src string, next, prev, temp, typ int, node *Node) (int, int) {
	// Extract the complete opening tag
	token := src[next : temp+1]

	// Process any text content before this control structure
	if next > prev {
		token := src[prev:next]
		p.addTextNode(token, nil, node)
	}

	// Create appropriate node based on control structure type
	switch typ {
	case nodeTypeFor:
		p.addForNode(token, nil, node)
	case nodeTypeIf:
		p.addIfNode(token, nil, node)
	}

	// Get the newly created node to process its contents
	childNode := node.Children[len(node.Children)-1]

	// Recursively parse the content between opening and closing tags
	tempPrev, tempNext := p.parser(src, temp+1, typ, childNode)

	// Extract and process the closing tag
	tempToken := src[tempPrev : tempNext+1]
	switch typ {
	case nodeTypeFor:
		p.addForNode(tempToken, nil, childNode)
	case nodeTypeIf:
		p.addIfNode(tempToken, nil, childNode)
	}

	// Update positions to continue parsing after this structure
	next = tempNext + 1
	prev = tempNext + 1

	return next, prev
}

// handleConditionalBranch processes conditional branches (else/elif) within an if control structure.
// It handles the else/elif tag and prepares for processing the branch content.
// Parameters:
// - src: source template string
// - next, prev: current positions in the source
// - temp: position of the end of else/elif tag
// - branchType: type of conditional branch (3: else, 8: elif)
// - node: current if node being processed
// Returns:
// - updated next and prev positions
// - markEnterElse flag to indicate conditional branch
// - updated node reference for branch content
func (p *Parser) handleConditionalBranch(src string, next, prev, temp, branchType int, node *Node) (int, int, bool, *Node) {
	// Extract the complete else/elif tag
	token := src[next : temp+1]

	// Process any text content before the conditional branch tag
	if next > prev {
		textToken := src[prev:next]
		p.addTextNode(textToken, nil, node)
	}

	// Create appropriate node type based on branch type
	var nodeType string
	switch branchType {
	case nodeTypeElse:
		nodeType = "else"
	case nodeTypeElif:
		nodeType = "elif"
	default:
		nodeType = "text"
	}

	// Add conditional branch node as a child of the if node
	tempNode := p.addConditionalBranchNode(token, nil, node, nodeType)

	// Set flag to indicate we're in conditional branch
	markEnterElse := true

	// Return the newly created conditional branch node for content parsing
	// but don't update the main node reference
	next = temp + 1
	prev = temp + 1

	return next, prev, markEnterElse, tempNode
}

// handleVariable processes variable expressions {{ ... }} in the template.
// It can handle variables both at the root level (using template) and
// nested within control structures (using node).
// Parameters:
// - src: source template string
// - next, prev: current positions in the source
// - template: root template (nil if processing nested variable)
// - node: parent node (nil if processing root level variable)
// Returns updated next and prev positions
func (p *Parser) handleVariable(src string, next, prev int, template *Template, node *Node) (int, int) {
	n := len(src)
	// Try to match variable expression ending with "}}"
	matched, tempNext, _ := p.matchAppropriateStrings(src, n, next+2, "}}")

	if matched { //nolint:nestif
		// Extract the complete variable expression
		token := src[next : tempNext+1]

		// Process any text content before the variable
		if next > prev {
			token := src[prev:next]
			if template != nil {
				// Add text node to root template
				p.addTextNode(token, template, nil)
			} else {
				// Add text node to parent control structure
				p.addTextNode(token, nil, node)
			}
		}

		// Add the variable node to appropriate parent
		if template != nil {
			// Add variable to root template
			p.addVariableNode(token, template, nil)
		} else {
			// Add variable to parent control structure
			p.addVariableNode(token, nil, node)
		}

		// Update positions to continue after variable
		prev = tempNext + 1
		next = tempNext + 1
	} else {
		// No valid variable match found, move forward
		next++
	}

	return next, prev
}

// handleComment processes comment blocks {# ... #} in the template
func (p *Parser) handleComment(src string, n, next, prev int, node *Node) (int, int) {
	matched, tempNext, _ := p.matchAppropriateStrings(src, n, next+2, "#}")
	if matched {
		// Add text content before comment
		if next > prev {
			token := src[prev:next]
			p.addTextNode(token, nil, node)
		}

		// Comment content is not added to template, skip it
		// Update position to continue parsing after comment
		return tempNext + 1, tempNext + 1
	}

	// No valid comment match found, move forward
	return next + 1, prev
}

func (p *Parser) parseExplanatoryNote(src string, next, prev int, template *Template) (int, int) {
	n := len(src)
	// Find comment end marker "#}"
	endPos := next + 2
	for endPos < n-1 {
		if src[endPos] == '#' && src[endPos+1] == '}' {
			// Found comment end marker

			// Handle text content before comment
			if next > prev {
				textToken := src[prev:next]
				p.addTextNode(textToken, template, nil)
			}

			// Comment content is not added to template, skip it

			// Update position to continue parsing after comment
			next = endPos + 2
			prev = next
			return next, prev
		}
		endPos++
	}

	// If no end marker found, just advance one character
	next++

	return next, prev
}
