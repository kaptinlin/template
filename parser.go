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

// Parser analyzes template syntax.
type Parser struct{}

// NewParser creates a Parser with a compiled regular expression for efficiency.
func NewParser() *Parser {
	return &Parser{}
}

// Updated addVariableNode processes a variable token, parses out any filters, and adds it to the template.
func (p *Parser) addVariableNode(token string, tpl *Template, node *Node) {
	if tpl != nil {
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
		node = &Node{
			Type:     "variable",
			Variable: varName,
			Filters:  filters,
			Text:     token,
		}

		// Add the new node to the template.
		tpl.Nodes = append(tpl.Nodes, node)
	} else {
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
		node.Children = append(node.Children, &Node{
			Type:     "variable",
			Variable: varName,
			Filters:  filters,
			Text:     token,
		})
	}
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
			} else if number, err := strconv.ParseFloat(trimmedArg, 64); err == nil {
				args = append(args, NumberArg{val: number})
			} else {
				// Treat as variable
				args = append(args, VariableArg{name: trimmedArg})
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
func (p *Parser) addTextNode(text string, tpl *Template, node *Node) {
	if text != "" && tpl != nil {
		tpl.Nodes = append(tpl.Nodes, &Node{Type: "text", Text: text})
	} else {
		node.Children = append(node.Children, &Node{Type: "text", Text: text})
	}
}

// addForNode adds a for node to the template
func (p *Parser) addForNode(text string, tpl *Template, node *Node) {
	matched := endforRegex.MatchString(text)
	if text != "" && tpl != nil {
		tpl.Nodes = append(tpl.Nodes, &Node{Type: "text", Text: text})
	} else if !matched {
		node.Children = append(node.Children, &Node{Type: "text", Text: text})
	}
	if matched {
		node.Type = "for"
		node.EndText = text
		analyzeForParameter(node.Text, node)
	}
}

// analyzeForParameter analyzes the for parameter
func analyzeForParameter(text string, node *Node) {
	n := len(text)
	next := 0
	for next < n {
		if text[next:next+4] == " in " {
			left := next
			right := next + 4
			for left > 0 {
				if text[0:left] == "{% for " {
					node.Variable = text[left:next]
				}
				left--
			}
			for right < n {
				if text[right:n] == " %}" {
					node.Collection = text[next+4 : right]
				}
				right++
			}
			break
		}
		next++
	}
}

// addIfNode adds an if node to the template
func (p *Parser) addIfNode(text string, tpl *Template, node *Node) {
	matched := endifRegex.MatchString(text)
	if text != "" && tpl != nil {
		tpl.Nodes = append(tpl.Nodes, &Node{Type: "if", Text: text})
	} else if !matched {
		node.Children = append(node.Children, &Node{Type: "text", Text: text})
	}
	if matched {
		node.Type = "if"
		node.EndText = text
	}
}

// addControlFlowNode adds a control flow node (break/continue) to the template
func (p *Parser) addControlFlowNode(text string, tpl *Template, node *Node, nodeType string) {
	newNode := &Node{Type: nodeType, Text: text}

	if tpl != nil {
		// Adding to template root
		tpl.Nodes = append(tpl.Nodes, newNode)
	} else if node != nil {
		// Adding to a parent node
		node.Children = append(node.Children, newNode)
	}
}

// addConditionalBranchNode adds a conditional branch node (else/elif) to the template
func (p *Parser) addConditionalBranchNode(text string, tpl *Template, node *Node, nodeType string) *Node {
	newNode := &Node{Type: nodeType, Text: text}

	if tpl != nil {
		// Adding to template root
		tpl.Nodes = append(tpl.Nodes, newNode)
	} else if node != nil {
		// Adding to a parent node
		node.Children = append(node.Children, newNode)
	}

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
	if matched && typ <= 8 {
		// Extract the complete control structure token
		token := src[next : temp+1]
		// Handle any text content before the control structure
		if next > prev {
			token := src[prev:next]
			p.addTextNode(token, template, nil)
		}

		// Add appropriate node based on control structure type
		switch typ {
		case 1:
			p.addForNode(token, template, nil)
		case 2:
			p.addIfNode(token, template, nil)
		case 6: // break
			p.addControlFlowNode(token, template, nil, NodeTypeBreak)
		case 7: // continue
			p.addControlFlowNode(token, template, nil, NodeTypeContinue)
		}

		// Handle control structures that need end tags (for/if)
		if typ < 3 {
			// Get the last added node to process its children
			node := template.Nodes[len(template.Nodes)-1]
			// Parse the content between opening and closing tags
			tempPrev, tempNext := p.parser(src, temp+1, typ, node)
			tempToken := src[tempPrev : tempNext+1]

			// Process the closing tag of the control structure
			switch typ {
			case 1:
				p.addForNode(tempToken, nil, node)
			case 2:
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
		return true, 1
	case ifRegex.MatchString(token):
		return true, 2
	case elseRegex.MatchString(token):
		return true, 3
	case elifRegex.MatchString(token):
		return true, 8
	case endforRegex.MatchString(token):
		return true, 4
	case endifRegex.MatchString(token):
		return true, 5
	case breakRegex.MatchString(token):
		return true, 6
	case continueRegex.MatchString(token):
		return true, 7
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
			case matched && tempType < 3:
				next, prev = p.handleControlStructure(src, next, prev, temp, tempType, node)
			case matched && tempType == 3 && typ == 5:
				if markEnterElse {
					// We're in an elif/else branch and found another else
					// Add any remaining text to current branch and return to if node
					if next > prev {
						token := src[prev:next]
						p.addTextNode(token, nil, node)
					}
					next, prev, markEnterElse, node = p.handleConditionalBranch(src, next, prev, temp, 3, enterIfNode)
				} else {
					next, prev, markEnterElse, node = p.handleConditionalBranch(src, next, prev, temp, 3, node)
				}
			case matched && tempType == 8 && typ == 5:
				if markEnterElse {
					// We're in an elif/else branch and found another elif
					// Add any remaining text to current branch and return to if node
					if next > prev {
						token := src[prev:next]
						p.addTextNode(token, nil, node)
					}
					next, prev, markEnterElse, node = p.handleConditionalBranch(src, next, prev, temp, 8, enterIfNode)
				} else {
					next, prev, markEnterElse, node = p.handleConditionalBranch(src, next, prev, temp, 8, node)
				}
			case matched && (tempType == 6 || tempType == 7):
				// Handle break/continue control flow
				token := src[next : temp+1]
				if next > prev {
					textToken := src[prev:next]
					p.addTextNode(textToken, nil, node)
				}
				// Add break or continue node
				if tempType == 6 {
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
			// 处理注释
			matched, tempNext, _ := p.matchAppropriateStrings(src, n, next+2, "#}")
			if matched {
				// 处理注释前的文本内容
				if next > prev {
					token := src[prev:next]
					p.addTextNode(token, nil, node)
				}

				// 注释内容不添加到模板中，直接跳过

				// 更新位置到注释之后继续解析
				next = tempNext + 1
				prev = tempNext + 1
			} else {
				next++
			}
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
	case 1:
		p.addForNode(token, nil, node)
	case 2:
		p.addIfNode(token, nil, node)
	}

	// Get the newly created node to process its contents
	node1 := node.Children[len(node.Children)-1]

	// Recursively parse the content between opening and closing tags
	tempPrev, tempNext := p.parser(src, temp+1, typ, node1)

	// Extract and process the closing tag
	tempToken := src[tempPrev : tempNext+1]
	switch typ {
	case 1:
		p.addForNode(tempToken, nil, node1)
	case 2:
		p.addIfNode(tempToken, nil, node1)
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
	case 3:
		nodeType = "else"
	case 8:
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
func (p *Parser) parseExplanatoryNote(src string, next, prev int, template *Template) (int, int) {
	n := len(src)
	// 寻找注释结束标记 "#}"
	endPos := next + 2
	for endPos < n-1 {
		if src[endPos] == '#' && src[endPos+1] == '}' {
			// 找到注释结束标记

			// 处理注释前的文本内容
			if next > prev {
				textToken := src[prev:next]
				p.addTextNode(textToken, template, nil)
			}

			// 注释内容不添加到模板中，直接跳过

			// 更新位置到注释之后继续解析
			next = endPos + 2
			prev = next
			return next, prev
		}
		endPos++
	}

	// 如果没有找到结束标记，只前进一个字符
	next++

	return next, prev
}
