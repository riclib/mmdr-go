// Package parser provides parsing functionality for Mermaid diagrams.
package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

var (
	// Regex patterns for Mermaid syntax
	headerPattern        = regexp.MustCompile(`^\s*(flowchart|graph)\s+(TB|TD|BT|RL|LR)\s*$`)
	commentPattern       = regexp.MustCompile(`^\s*%%(.*)$`)
	subgraphStartPattern = regexp.MustCompile(`^\s*subgraph\s+(?:(\w+)\s*\[([^\]]+)\]|(\w+)|"([^"]+)")\s*$`)
	subgraphEndPattern   = regexp.MustCompile(`^\s*end\s*$`)
	classDefPattern      = regexp.MustCompile(`^\s*classDef\s+(\w+)\s+(.+)$`)
	classAssignPattern   = regexp.MustCompile(`^\s*class\s+([\w,\s]+?)\s+(\w+)\s*$`)

	// Node and link patterns
	// NOTE: Order matters in alternation - longer patterns must come before shorter ones
	nodeDefPattern = regexp.MustCompile(`^\s*(\w+)\s*(\{\{|\[\[|\(\(|\[\(|\(\[|\[|\(|\{|>)([^\])\}]*?)(\}\}|\]\]|\)\)|\)\]|\]\)|\]|\)|\})?\s*$`)

	// Pattern to match a node reference with optional inline definition
	// Captures: nodeID + optional (openBracket + label + closeBracket)
	// NOTE: Order matters in alternation - longer patterns must come before shorter ones
	nodeWithOptDef   = `(\w+)(?:\s*(\{\{|\[\[|\(\(|\[\(|\(\[|\[|\(|\{|>)([^\])\}]*?)(\}\}|\]\]|\)\)|\)\]|\]\)|\]|\)|\}))?`
	linkPattern      = regexp.MustCompile(`^` + nodeWithOptDef + `\s*(<)?(-{2,3}|-\.{1,2}-|={2,3})(>)?\s*(\|([^|]+)\|)?\s*` + nodeWithOptDef + `$`)
	biDirLinkPattern = regexp.MustCompile(`^` + nodeWithOptDef + `\s*(<)(--|==|-\.-)(>)\s*(\|([^|]+)\|)?\s*` + nodeWithOptDef + `$`)
)

// FlowchartParser parses Mermaid flowchart and graph diagrams.
type FlowchartParser struct {
	// Pending NodeDefs from link parsing (from and to nodes)
	pendingFromNode *ast.NodeDef
	pendingToNode   *ast.NodeDef
	// Track which nodes have been defined to avoid duplicates
	definedNodes map[string]bool
}

// SupportedTypes returns the diagram types this parser handles.
func (p *FlowchartParser) SupportedTypes() []string {
	return []string{"flowchart", "graph"}
}

// NewFlowchartParser creates a new flowchart parser.
func NewFlowchartParser() *FlowchartParser {
	return &FlowchartParser{
		definedNodes: make(map[string]bool),
	}
}

// Parse parses a Mermaid flowchart/graph diagram from a string.
func (p *FlowchartParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	flowchart, err := p.parseLines(lines)
	if err != nil {
		return nil, err
	}
	// Set the source field
	flowchart.Source = source
	return flowchart, nil
}

// ParseBytes parses a Mermaid flowchart/graph diagram from bytes.
func (p *FlowchartParser) ParseBytes(_ string, source []byte) (*ast.Flowchart, error) {
	diagram, err := p.Parse(string(source))
	if err != nil {
		return nil, err
	}
	flowchart, ok := diagram.(*ast.Flowchart)
	if !ok {
		return nil, fmt.Errorf("parsed diagram is not a flowchart: %T", diagram)
	}
	return flowchart, nil
}

func (p *FlowchartParser) parseLines(lines []string) (*ast.Flowchart, error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram")
	}

	// Parse header
	header := strings.TrimSpace(lines[0])
	matches := headerPattern.FindStringSubmatch(header)
	if matches == nil {
		return nil, fmt.Errorf("invalid diagram header: expected 'flowchart' or 'graph' followed by direction")
	}

	flowchart := &ast.Flowchart{
		Type:      matches[1],
		Direction: matches[2],
		Pos:       ast.Position{Line: 1, Column: 1},
	}

	// Parse statements
	statements, err := p.parseStatements(lines[1:], 1, false)
	if err != nil {
		return nil, err
	}
	flowchart.Statements = statements

	return flowchart, nil
}

func (p *FlowchartParser) parseStatements(lines []string, startLine int, inSubgraph bool) ([]ast.Statement, error) {
	var statements []ast.Statement
	lineNum := startLine

	for i := 0; i < len(lines); i++ {
		lineNum++
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Handle comments
		if commentPattern.MatchString(trimmed) {
			matches := commentPattern.FindStringSubmatch(trimmed)
			statements = append(statements, &ast.Comment{
				Text: strings.TrimSpace(matches[1]),
				Pos:  ast.Position{Line: lineNum, Column: 1},
			})
			continue
		}

		// Handle subgraph end
		if subgraphEndPattern.MatchString(trimmed) {
			if !inSubgraph {
				return nil, fmt.Errorf("line %d: 'end' without matching 'subgraph'", lineNum)
			}
			return statements, nil
		}

		// Handle subgraph start
		if matches := subgraphStartPattern.FindStringSubmatch(trimmed); matches != nil {
			// Find the matching 'end'
			nestedLines, consumed, err := p.extractSubgraphLines(lines[i+1:], lineNum+1)
			if err != nil {
				return nil, err
			}

			nestedStatements, err := p.parseStatements(nestedLines, lineNum, true)
			if err != nil {
				return nil, err
			}

			// Extract title from matches
			// matches[1]: ID (if using ID[display] syntax)
			// matches[2]: Display name in brackets (if using ID[display] syntax)
			// matches[3]: Quoted name (if using "name" syntax)
			var title string
			if matches[2] != "" {
				// Use display name from brackets, strip quotes if present
				title = strings.Trim(matches[2], `"`)
			} else if matches[1] != "" {
				// Use ID if no display name
				title = matches[1]
			} else if matches[3] != "" {
				// Use quoted name
				title = matches[3]
			}

			statements = append(statements, &ast.Subgraph{
				Title:      title,
				Statements: nestedStatements,
				Pos:        ast.Position{Line: lineNum, Column: 1},
			})

			i += consumed
			lineNum += consumed
			continue
		}

		// Handle classDef
		if matches := classDefPattern.FindStringSubmatch(trimmed); matches != nil {
			styles := p.parseStyles(matches[2])
			statements = append(statements, &ast.ClassDef{
				Name:   matches[1],
				Styles: styles,
				Pos:    ast.Position{Line: lineNum, Column: 1},
			})
			continue
		}

		// Handle class assignment
		if matches := classAssignPattern.FindStringSubmatch(trimmed); matches != nil {
			nodeIDs := strings.Split(matches[1], ",")
			for i, id := range nodeIDs {
				nodeIDs[i] = strings.TrimSpace(id)
			}
			statements = append(statements, &ast.ClassAssignment{
				NodeIDs:   nodeIDs,
				ClassName: matches[2],
				Pos:       ast.Position{Line: lineNum, Column: 1},
			})
			continue
		}

		// Try to parse as link (bidirectional or unidirectional)
		if stmt := p.parseLink(trimmed, lineNum); stmt != nil {
			// Insert inline NodeDefs in the correct order:
			// 1. "from" node definition (if present)
			// 2. Link statement
			// 3. "to" node definition (if present)
			if p.pendingFromNode != nil {
				statements = append(statements, p.pendingFromNode)
			}
			statements = append(statements, stmt)
			if p.pendingToNode != nil {
				statements = append(statements, p.pendingToNode)
			}
			// Clear pending nodes
			p.pendingFromNode = nil
			p.pendingToNode = nil
			continue
		}

		// Try to parse as node definition
		if stmt := p.parseNodeDef(trimmed, lineNum); stmt != nil {
			if nodeDef, ok := stmt.(*ast.NodeDef); ok {
				p.definedNodes[nodeDef.ID] = true
			}
			statements = append(statements, stmt)
			continue
		}

		// If we can't parse the line, skip it (for now - could return error in strict mode)
		continue
	}

	if inSubgraph {
		return nil, fmt.Errorf("unclosed subgraph")
	}

	return statements, nil
}

func (p *FlowchartParser) extractSubgraphLines(lines []string, startLine int) ([]string, int, error) {
	var subgraphLines []string
	depth := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		//nolint:gocritic // if-else chain is more readable here than a switch
		if subgraphStartPattern.MatchString(trimmed) {
			depth++
			subgraphLines = append(subgraphLines, line)
		} else if subgraphEndPattern.MatchString(trimmed) {
			if depth == 0 {
				// This is the end of our subgraph - include it
				subgraphLines = append(subgraphLines, line)
				return subgraphLines, i + 1, nil
			}
			depth--
			subgraphLines = append(subgraphLines, line)
		} else {
			subgraphLines = append(subgraphLines, line)
		}
	}

	return nil, 0, fmt.Errorf("line %d: unclosed subgraph", startLine)
}

// extractNodeDef extracts a NodeDef from a node reference that may include an inline definition
// e.g., "B[Label]" -> NodeDef{ID: "B", Label: "Label", Shape: "[]"}
// Returns nil if the node reference is just an ID without a definition
func (p *FlowchartParser) extractNodeDef(nodeID, openBracket, label, closeBracket string, lineNum int) *ast.NodeDef {
	// If no brackets, it's just a node reference, not a definition
	if openBracket == "" || closeBracket == "" {
		return nil
	}

	return &ast.NodeDef{
		ID:    nodeID,
		Shape: openBracket + closeBracket,
		Label: strings.TrimSpace(label),
		Pos:   ast.Position{Line: lineNum, Column: 1},
	}
}

func (p *FlowchartParser) parseLink(line string, lineNum int) ast.Statement {
	// Clear pending nodes from previous calls
	p.pendingFromNode = nil
	p.pendingToNode = nil

	// Try bidirectional link first
	if matches := biDirLinkPattern.FindStringSubmatch(line); matches != nil {
		// Updated match groups with inline node definitions:
		// 1: from ID
		// 2: from open bracket (optional)
		// 3: from label (optional)
		// 4: from close bracket (optional)
		// 5: left arrow part <
		// 6: arrow middle (-->, ===, ---)
		// 7: right arrow part >
		// 8: link label with pipes (optional)
		// 9: link label content (optional)
		// 10: to ID
		// 11: to open bracket (optional)
		// 12: to label (optional)
		// 13: to close bracket (optional)

		fromID := matches[1]
		toID := matches[10]

		// Extract inline NodeDefs if present and not already defined
		if !p.definedNodes[fromID] {
			p.pendingFromNode = p.extractNodeDef(matches[1], matches[2], matches[3], matches[4], lineNum)
			if p.pendingFromNode != nil {
				p.definedNodes[fromID] = true
			}
		}
		if !p.definedNodes[toID] {
			p.pendingToNode = p.extractNodeDef(matches[10], matches[11], matches[12], matches[13], lineNum)
			if p.pendingToNode != nil {
				p.definedNodes[toID] = true
			}
		}

		label := ""
		if len(matches) > 9 && matches[9] != "" {
			label = strings.TrimSpace(matches[9])
		}

		return &ast.Link{
			From:  fromID,
			To:    toID,
			Arrow: matches[5] + matches[6] + matches[7], // <-->
			Label: label,
			BiDir: true,
			Pos:   ast.Position{Line: lineNum, Column: 1},
		}
	}

	// Try unidirectional link
	if matches := linkPattern.FindStringSubmatch(line); matches != nil {
		// Updated match groups with inline node definitions:
		// 1: from ID
		// 2: from open bracket (optional)
		// 3: from label (optional)
		// 4: from close bracket (optional)
		// 5: left arrow part < (optional)
		// 6: arrow middle (--, ---, -.-, etc.)
		// 7: right arrow part > (optional)
		// 8: link label with pipes (optional)
		// 9: link label content (optional)
		// 10: to ID
		// 11: to open bracket (optional)
		// 12: to label (optional)
		// 13: to close bracket (optional)

		fromID := matches[1]
		toID := matches[10]

		// Extract inline NodeDefs if present and not already defined
		if !p.definedNodes[fromID] {
			p.pendingFromNode = p.extractNodeDef(matches[1], matches[2], matches[3], matches[4], lineNum)
			if p.pendingFromNode != nil {
				p.definedNodes[fromID] = true
			}
		}
		if !p.definedNodes[toID] {
			p.pendingToNode = p.extractNodeDef(matches[10], matches[11], matches[12], matches[13], lineNum)
			if p.pendingToNode != nil {
				p.definedNodes[toID] = true
			}
		}

		arrow := matches[6]
		if matches[5] == "<" {
			arrow = "<" + arrow
		}
		if matches[7] == ">" {
			arrow += ">"
		}

		label := ""
		if len(matches) > 9 && matches[9] != "" {
			label = strings.TrimSpace(matches[9])
		}

		return &ast.Link{
			From:  fromID,
			To:    toID,
			Arrow: arrow,
			Label: label,
			BiDir: false,
			Pos:   ast.Position{Line: lineNum, Column: 1},
		}
	}

	return nil
}

func (p *FlowchartParser) parseNodeDef(line string, lineNum int) ast.Statement {
	matches := nodeDefPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	id := matches[1]
	shape := ""
	label := ""

	if len(matches) > 2 && matches[2] != "" {
		// Shape is opening + closing brackets
		shape = matches[2]
		if len(matches) > 4 && matches[4] != "" {
			shape += matches[4]
		}
		if len(matches) > 3 {
			label = strings.TrimSpace(matches[3])
		}
	}

	return &ast.NodeDef{
		ID:    id,
		Shape: shape,
		Label: label,
		Pos:   ast.Position{Line: lineNum, Column: 1},
	}
}

func (p *FlowchartParser) parseStyles(styleStr string) map[string]string {
	styles := make(map[string]string)
	parts := strings.SplitSeq(styleStr, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		if kv := strings.SplitN(part, ":", 2); len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			styles[key] = value
		}
	}

	return styles
}
