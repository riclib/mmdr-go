package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// MindmapParser handles parsing of mindmap diagrams.
type MindmapParser struct{}

// NewMindmapParser creates a new mindmap parser.
func NewMindmapParser() *MindmapParser {
	return &MindmapParser{}
}

var (
	mindmapHeaderRegex = regexp.MustCompile(`^mindmap\s*$`)
	mindmapIconRegex   = regexp.MustCompile(`^\s*::icon\(([^)]+)\)\s*$`)
)

// Parse parses a mindmap diagram source.
func (p *MindmapParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	diagram := &ast.MindmapDiagram{
		Type:   "mindmap",
		Source: source,
		Pos:    ast.Position{Line: 1, Column: 1},
	}

	// Parse header line
	firstLine := strings.TrimSpace(lines[0])
	if !mindmapHeaderRegex.MatchString(firstLine) {
		return nil, fmt.Errorf("invalid mindmap diagram header: %s", firstLine)
	}

	// Build tree structure from indented lines
	nodeStack := make([]*ast.MindmapNode, 0)
	lastLevel := -1
	indentSize := 0  // Will be detected as 2 or 4
	rootIndent := -1 // Track root indentation

	for i := 1; i < len(lines); i++ {
		line := lines[i]

		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		// Calculate indentation level
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// First non-comment line is the root - track its indentation
		if diagram.Root == nil {
			rootIndent = indent
		}

		// Calculate relative indentation from root
		relativeIndent := indent - rootIndent

		// Detect indent size from first direct child of root (level 1)
		// Must be consistent - first child sets the pattern
		if indentSize == 0 && relativeIndent > 0 {
			// First child after root should be exactly one level deeper
			// Only accept 2 or 4 space indentation styles
			if relativeIndent == 2 || relativeIndent == 4 {
				// Additional check: if we're not at level 0, we can't set indent size
				// This catches cases where first indented line isn't directly after root
				if lastLevel != 0 {
					return nil, fmt.Errorf("line %d: cannot establish indentation pattern (not a direct child of root)", i+1)
				}
				indentSize = relativeIndent
			} else {
				return nil, fmt.Errorf("line %d: invalid indentation (expected 2 or 4 space indentation style, got %d spaces)", i+1, relativeIndent)
			}
		}

		// Calculate level
		var level int
		if relativeIndent == 0 {
			level = 0
		} else if relativeIndent > 0 {
			if indentSize == 0 {
				// This should never happen due to check above, but be defensive
				return nil, fmt.Errorf("line %d: unexpected indentation before establishing indent size", i+1)
			}
			if relativeIndent%indentSize != 0 {
				return nil, fmt.Errorf("line %d: inconsistent indentation (expected multiples of %d spaces)", i+1, indentSize)
			}
			level = relativeIndent / indentSize
		} else {
			return nil, fmt.Errorf("line %d: invalid indentation (less than root)", i+1)
		}

		// Check for icon line
		if iconMatches := mindmapIconRegex.FindStringSubmatch(trimmed); iconMatches != nil {
			if len(nodeStack) == 0 {
				return nil, fmt.Errorf("line %d: icon definition outside of node", i+1)
			}
			// Add icon to last node
			nodeStack[len(nodeStack)-1].Icon = strings.TrimSpace(iconMatches[1])
			continue
		}

		// Parse node text and shape
		text, shape := parseNodeText(trimmed)

		if text == "" {
			return nil, fmt.Errorf("line %d: node text cannot be empty", i+1)
		}

		// Create new node
		node := &ast.MindmapNode{
			Text:     text,
			Shape:    shape,
			Level:    level,
			Children: make([]*ast.MindmapNode, 0),
			Pos:      ast.Position{Line: i + 1, Column: 1},
		}

		// Determine parent and add to tree
		if level == 0 {
			// Root node
			if diagram.Root != nil {
				return nil, fmt.Errorf("line %d: multiple root nodes found", i+1)
			}
			diagram.Root = node
			nodeStack = []*ast.MindmapNode{node}
		} else {
			// Child node
			if len(nodeStack) == 0 {
				return nil, fmt.Errorf("line %d: child node before root node", i+1)
			}

			// Find parent (pop stack until we find correct level)
			for len(nodeStack) > level {
				nodeStack = nodeStack[:len(nodeStack)-1]
			}

			if len(nodeStack) == 0 {
				return nil, fmt.Errorf("line %d: invalid nesting level", i+1)
			}

			parent := nodeStack[len(nodeStack)-1]

			// Check level increment is valid (can only increase by 1)
			if level > lastLevel+1 {
				return nil, fmt.Errorf("line %d: invalid nesting (jumped from level %d to %d)", i+1, lastLevel, level)
			}

			parent.Children = append(parent.Children, node)
			nodeStack = append(nodeStack, node)
		}

		lastLevel = level
	}

	if diagram.Root == nil {
		return nil, fmt.Errorf("mindmap must have a root node")
	}

	return diagram, nil
}

// parseNodeText extracts the text and shape from a node line.
// Handles both: "((text))" and "id((text))" or "id[text]" formats
func parseNodeText(line string) (text string, shape string) {
	line = strings.TrimSpace(line)

	// Check for different shapes - order matters for overlapping patterns
	shapes := []struct {
		prefix string
		suffix string
		shape  string
	}{
		{"))", "((", "))(("}, // Hexagon - must check before ()
		{"((", "))", "(())"}, // Double circle
		{"{{", "}}", "{{}}"}, // Cloud
		{"[", "]", "[]"},     // Square
		{"(", ")", "()"},     // Circle
	}

	for _, s := range shapes {
		// For Hexagon shape ))text((, handle both with and without ID
		if s.shape == "))((" {
			// Try at start first: ))text((
			if strings.HasPrefix(line, s.prefix) && strings.HasSuffix(line, s.suffix) {
				text = strings.TrimSpace(line[len(s.prefix) : len(line)-len(s.suffix)])
				shape = s.shape
				return text, shape
			}
			// Try after an ID: id))text((
			shapeStartIdx := strings.Index(line, s.prefix)
			if shapeStartIdx > 0 && strings.HasSuffix(line, s.suffix) {
				shapeStart := shapeStartIdx + len(s.prefix)
				shapeEnd := len(line) - len(s.suffix)
				if shapeEnd >= shapeStart {
					text = strings.TrimSpace(line[shapeStart:shapeEnd])
					shape = s.shape
					return text, shape
				}
			}
		} else {
			// For other shapes, look for shape starting from beginning or after an ID
			// Try at start first
			if strings.HasPrefix(line, s.prefix) && strings.HasSuffix(line, s.suffix) {
				text = strings.TrimSpace(line[len(s.prefix) : len(line)-len(s.suffix)])
				shape = s.shape
				return text, shape
			}
			// Try after an ID (e.g., "id((text))")
			shapeStartIdx := strings.Index(line, s.prefix)
			if shapeStartIdx > 0 && strings.HasSuffix(line, s.suffix) {
				shapeStart := shapeStartIdx + len(s.prefix)
				shapeEnd := len(line) - len(s.suffix)
				if shapeEnd >= shapeStart {
					text = strings.TrimSpace(line[shapeStart:shapeEnd])
					shape = s.shape
					return text, shape
				}
			}
		}
	}

	// No shape markers - plain text
	return line, ""
}

// SupportedTypes returns the diagram types this parser supports.
func (p *MindmapParser) SupportedTypes() []string {
	return []string{"mindmap"}
}
