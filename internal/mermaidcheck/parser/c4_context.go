package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// C4 element patterns (shared across all C4 diagram types)
var (
	c4TitlePattern         = regexp.MustCompile(`^\s*title\s+(.+)$`)
	c4CommentPattern       = regexp.MustCompile(`^\s*%%.*$`)
	c4PersonPattern        = regexp.MustCompile(`^\s*Person(?:_Ext)?\s*\(([^)]+)\)\s*$`)
	c4SystemPattern        = regexp.MustCompile(`^\s*System(?:_Ext)?\s*\(([^)]+)\)\s*$`)
	c4ContainerPattern     = regexp.MustCompile(`^\s*Container(?:Db|Queue)?\s*\(([^)]+)\)\s*$`)
	c4ComponentPattern     = regexp.MustCompile(`^\s*Component(?:Db|Queue)?\s*\(([^)]+)\)\s*$`)
	c4NodePattern          = regexp.MustCompile(`^\s*(?:Deployment_)?Node\s*\(([^)]+)\)\s*$`)
	c4RelPattern           = regexp.MustCompile(`^\s*(Rel|Rel_Back|Rel_Neighbor|Rel_Down|Rel_Up|Rel_Left|Rel_Right|BiRel)\s*\(([^)]+)\)\s*$`)
	c4BoundaryStartPattern = regexp.MustCompile(`^\s*(Boundary|Enterprise_Boundary|System_Boundary|Container_Boundary|Deployment_Node|Node)\s*\(([^)]+)\)\s*\{\s*$`)
	c4BoundaryEndPattern   = regexp.MustCompile(`^\s*\}\s*$`)
	c4ElementStylePattern  = regexp.MustCompile(`^\s*UpdateElementStyle\s*\(([^)]+)\)\s*$`)
	c4RelStylePattern      = regexp.MustCompile(`^\s*UpdateRelStyle\s*\(([^)]+)\)\s*$`)
)

// C4ContextParser parses C4 Context diagrams.
type C4ContextParser struct{}

// NewC4ContextParser creates a new C4 Context parser.
func NewC4ContextParser() *C4ContextParser {
	return &C4ContextParser{}
}

// Parse parses a C4 Context diagram and returns a C4Diagram AST.
func (p *C4ContextParser) Parse(source string) (ast.Diagram, error) {
	return parseC4Diagram(source, "c4Context", "C4Context")
}

// SupportedTypes returns the diagram types this parser supports.
func (p *C4ContextParser) SupportedTypes() []string {
	return []string{"c4Context"}
}

// parseC4Diagram is a shared parser for all C4 diagram types.
func parseC4Diagram(source, diagramType, expectedHeader string) (*ast.C4Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	// Check header
	firstLine := strings.TrimSpace(lines[0])
	if firstLine != expectedHeader {
		return nil, fmt.Errorf("expected %s header, got: %s", expectedHeader, firstLine)
	}

	diagram := &ast.C4Diagram{
		DiagramType:   diagramType,
		Elements:      []ast.C4Element{},
		Boundaries:    []ast.C4Boundary{},
		Relationships: []ast.C4Relationship{},
		Styles:        []ast.C4Style{},
		Source:        source,
		Pos:           ast.Position{Line: 1, Column: 1},
	}

	// Parse body (skip header)
	var err error
	diagram.Boundaries, err = parseC4Body(lines[1:], 2, diagram)
	if err != nil {
		return nil, err
	}

	return diagram, nil
}

// parseC4Body parses the body of a C4 diagram, handling nested boundaries.
func parseC4Body(lines []string, startLine int, diagram *ast.C4Diagram) ([]ast.C4Boundary, error) {
	var boundaries []ast.C4Boundary
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		lineNum := startLine + i

		// Skip empty lines and comments
		if trimmed == "" || c4CommentPattern.MatchString(trimmed) {
			i++
			continue
		}

		// Parse title
		if matches := c4TitlePattern.FindStringSubmatch(trimmed); matches != nil {
			diagram.Title = strings.TrimSpace(matches[1])
			i++
			continue
		}

		// Parse boundary start
		if matches := c4BoundaryStartPattern.FindStringSubmatch(trimmed); matches != nil {
			boundaryType := matches[1]
			params := parseC4Parameters(matches[2])

			if len(params) < 2 {
				return nil, fmt.Errorf("line %d: boundary requires at least id and label", lineNum)
			}

			boundary := ast.C4Boundary{
				BoundaryType: boundaryType,
				ID:           params[0],
				Label:        params[1],
				Elements:     []ast.C4Element{},
				Boundaries:   []ast.C4Boundary{},
				Pos:          ast.Position{Line: lineNum, Column: 1},
			}

			// For generic Boundary, third parameter is type
			if boundaryType == "Boundary" && len(params) >= 3 {
				boundary.Type = params[2]
			}

			// Find matching closing brace
			depth := 1
			boundaryEnd := i + 1
			for boundaryEnd < len(lines) && depth > 0 {
				if c4BoundaryStartPattern.MatchString(strings.TrimSpace(lines[boundaryEnd])) {
					depth++
				} else if c4BoundaryEndPattern.MatchString(strings.TrimSpace(lines[boundaryEnd])) {
					depth--
				}
				if depth > 0 {
					boundaryEnd++
				}
			}

			if depth > 0 {
				return nil, fmt.Errorf("line %d: unclosed boundary %s", lineNum, boundary.ID)
			}

			// Parse boundary contents recursively
			boundaryLines := lines[i+1 : boundaryEnd]
			nestedBoundaries, err := parseC4BoundaryContents(boundaryLines, lineNum+1, diagram, &boundary)
			if err != nil {
				return nil, err
			}
			boundary.Boundaries = nestedBoundaries

			boundaries = append(boundaries, boundary)
			i = boundaryEnd + 1
			continue
		}

		// Parse elements
		if elem, ok := parseC4Element(trimmed, lineNum); ok {
			diagram.Elements = append(diagram.Elements, elem)
			i++
			continue
		}

		// Parse relationships
		if rel, ok := parseC4Relationship(trimmed, lineNum); ok {
			diagram.Relationships = append(diagram.Relationships, rel)
			i++
			continue
		}

		// Parse styles
		if style, ok := parseC4Style(trimmed, lineNum); ok {
			diagram.Styles = append(diagram.Styles, style)
			i++
			continue
		}

		// Unknown line
		return nil, fmt.Errorf("line %d: unrecognised C4 syntax: %s", lineNum, trimmed)
	}

	return boundaries, nil
}

// parseC4BoundaryContents parses the contents of a boundary (elements and nested boundaries).
func parseC4BoundaryContents(lines []string, startLine int, diagram *ast.C4Diagram, boundary *ast.C4Boundary) ([]ast.C4Boundary, error) {
	var boundaries []ast.C4Boundary
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		lineNum := startLine + i

		// Skip empty lines and comments
		if trimmed == "" || c4CommentPattern.MatchString(trimmed) {
			i++
			continue
		}

		// Parse nested boundary
		if matches := c4BoundaryStartPattern.FindStringSubmatch(trimmed); matches != nil {
			boundaryType := matches[1]
			params := parseC4Parameters(matches[2])

			if len(params) < 2 {
				return nil, fmt.Errorf("line %d: boundary requires at least id and label", lineNum)
			}

			nestedBoundary := ast.C4Boundary{
				BoundaryType: boundaryType,
				ID:           params[0],
				Label:        params[1],
				Elements:     []ast.C4Element{},
				Boundaries:   []ast.C4Boundary{},
				Pos:          ast.Position{Line: lineNum, Column: 1},
			}

			if boundaryType == "Boundary" && len(params) >= 3 {
				nestedBoundary.Type = params[2]
			}

			// Find matching closing brace
			depth := 1
			boundaryEnd := i + 1
			for boundaryEnd < len(lines) && depth > 0 {
				if c4BoundaryStartPattern.MatchString(strings.TrimSpace(lines[boundaryEnd])) {
					depth++
				} else if c4BoundaryEndPattern.MatchString(strings.TrimSpace(lines[boundaryEnd])) {
					depth--
				}
				if depth > 0 {
					boundaryEnd++
				}
			}

			if depth > 0 {
				return nil, fmt.Errorf("line %d: unclosed boundary %s", lineNum, nestedBoundary.ID)
			}

			// Parse nested boundary contents
			nestedLines := lines[i+1 : boundaryEnd]
			nestedNestedBoundaries, err := parseC4BoundaryContents(nestedLines, lineNum+1, diagram, &nestedBoundary)
			if err != nil {
				return nil, err
			}
			nestedBoundary.Boundaries = nestedNestedBoundaries

			boundaries = append(boundaries, nestedBoundary)
			i = boundaryEnd + 1
			continue
		}

		// Parse elements in boundary
		if elem, ok := parseC4Element(trimmed, lineNum); ok {
			boundary.Elements = append(boundary.Elements, elem)
			i++
			continue
		}

		// Parse relationships in boundary
		if rel, ok := parseC4Relationship(trimmed, lineNum); ok {
			diagram.Relationships = append(diagram.Relationships, rel)
			i++
			continue
		}

		// Unknown line in boundary
		return nil, fmt.Errorf("line %d: unrecognised C4 syntax in boundary: %s", lineNum, trimmed)
	}

	return boundaries, nil
}

// parseC4Element parses a C4 element (Person, System, Container, Component, Node).
func parseC4Element(line string, lineNum int) (ast.C4Element, bool) {
	// Try Person
	if matches := c4PersonPattern.FindStringSubmatch(line); matches != nil {
		params := parseC4Parameters(matches[1])
		if len(params) < 2 {
			return ast.C4Element{}, false
		}
		return ast.C4Element{
			ElementType: "Person",
			ID:          params[0],
			Label:       params[1],
			Description: getParam(params, 2),
			Sprite:      getParam(params, 3),
			Tags:        getParam(params, 4),
			Link:        getParam(params, 5),
			External:    strings.Contains(line, "Person_Ext"),
			Pos:         ast.Position{Line: lineNum, Column: 1},
		}, true
	}

	// Try System
	if matches := c4SystemPattern.FindStringSubmatch(line); matches != nil {
		params := parseC4Parameters(matches[1])
		if len(params) < 2 {
			return ast.C4Element{}, false
		}
		return ast.C4Element{
			ElementType: "System",
			ID:          params[0],
			Label:       params[1],
			Description: getParam(params, 2),
			Sprite:      getParam(params, 3),
			Tags:        getParam(params, 4),
			Link:        getParam(params, 5),
			External:    strings.Contains(line, "System_Ext"),
			Pos:         ast.Position{Line: lineNum, Column: 1},
		}, true
	}

	// Try Container
	if matches := c4ContainerPattern.FindStringSubmatch(line); matches != nil {
		params := parseC4Parameters(matches[1])
		if len(params) < 2 {
			return ast.C4Element{}, false
		}
		return ast.C4Element{
			ElementType: "Container",
			ID:          params[0],
			Label:       params[1],
			Technology:  getParam(params, 2),
			Description: getParam(params, 3),
			Sprite:      getParam(params, 4),
			Tags:        getParam(params, 5),
			Link:        getParam(params, 6),
			Database:    strings.Contains(line, "ContainerDb"),
			Queue:       strings.Contains(line, "ContainerQueue"),
			Pos:         ast.Position{Line: lineNum, Column: 1},
		}, true
	}

	// Try Component
	if matches := c4ComponentPattern.FindStringSubmatch(line); matches != nil {
		params := parseC4Parameters(matches[1])
		if len(params) < 2 {
			return ast.C4Element{}, false
		}
		return ast.C4Element{
			ElementType: "Component",
			ID:          params[0],
			Label:       params[1],
			Technology:  getParam(params, 2),
			Description: getParam(params, 3),
			Sprite:      getParam(params, 4),
			Tags:        getParam(params, 5),
			Link:        getParam(params, 6),
			Database:    strings.Contains(line, "ComponentDb"),
			Queue:       strings.Contains(line, "ComponentQueue"),
			Pos:         ast.Position{Line: lineNum, Column: 1},
		}, true
	}

	// Try Node (leaf nodes without braces)
	if matches := c4NodePattern.FindStringSubmatch(line); matches != nil {
		params := parseC4Parameters(matches[1])
		if len(params) < 2 {
			return ast.C4Element{}, false
		}
		return ast.C4Element{
			ElementType: "Node",
			ID:          params[0],
			Label:       params[1],
			Technology:  getParam(params, 2),
			Description: getParam(params, 3),
			Sprite:      getParam(params, 4),
			Tags:        getParam(params, 5),
			Link:        getParam(params, 6),
			Pos:         ast.Position{Line: lineNum, Column: 1},
		}, true
	}

	return ast.C4Element{}, false
}

// parseC4Relationship parses a C4 relationship.
func parseC4Relationship(line string, lineNum int) (ast.C4Relationship, bool) {
	matches := c4RelPattern.FindStringSubmatch(line)
	if matches == nil {
		return ast.C4Relationship{}, false
	}

	relType := matches[1]
	params := parseC4Parameters(matches[2])

	if len(params) < 3 {
		return ast.C4Relationship{}, false
	}

	return ast.C4Relationship{
		RelType:     relType,
		From:        params[0],
		To:          params[1],
		Label:       params[2],
		Technology:  getParam(params, 3),
		Description: getParam(params, 4),
		Sprite:      getParam(params, 5),
		Tags:        getParam(params, 6),
		Link:        getParam(params, 7),
		Pos:         ast.Position{Line: lineNum, Column: 1},
	}, true
}

// parseC4Style parses a C4 style override.
func parseC4Style(line string, lineNum int) (ast.C4Style, bool) {
	// Try element style
	if matches := c4ElementStylePattern.FindStringSubmatch(line); matches != nil {
		params := parseC4Parameters(matches[1])
		if len(params) < 1 {
			return ast.C4Style{}, false
		}
		return ast.C4Style{
			StyleType:   "UpdateElementStyle",
			ElementID:   params[0],
			BgColor:     getParam(params, 1),
			FontColor:   getParam(params, 2),
			BorderColor: getParam(params, 3),
			Shadowing:   getParam(params, 4),
			Shape:       getParam(params, 5),
			Pos:         ast.Position{Line: lineNum, Column: 1},
		}, true
	}

	// Try relationship style
	if matches := c4RelStylePattern.FindStringSubmatch(line); matches != nil {
		params := parseC4Parameters(matches[1])
		if len(params) < 2 {
			return ast.C4Style{}, false
		}
		return ast.C4Style{
			StyleType: "UpdateRelStyle",
			From:      params[0],
			To:        params[1],
			TextColor: getParam(params, 2),
			LineColor: getParam(params, 3),
			OffsetX:   getParam(params, 4),
			OffsetY:   getParam(params, 5),
			Pos:       ast.Position{Line: lineNum, Column: 1},
		}, true
	}

	return ast.C4Style{}, false
}

// parseC4Parameters parses comma-separated parameters, handling quoted strings.
func parseC4Parameters(params string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	for i, ch := range params {
		switch {
		case escaped:
			current.WriteRune(ch)
			escaped = false
		case ch == '\\':
			escaped = true
		case ch == '"':
			inQuotes = !inQuotes
		case ch == ',' && !inQuotes:
			result = append(result, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(ch)
		}

		// Handle last character
		if i == len(params)-1 {
			result = append(result, strings.TrimSpace(current.String()))
		}
	}

	// Remove surrounding quotes from each parameter
	for i, p := range result {
		if len(p) >= 2 && p[0] == '"' && p[len(p)-1] == '"' {
			result[i] = p[1 : len(p)-1]
		}
	}

	return result
}

// getParam safely gets a parameter at index, returning empty string if out of bounds.
func getParam(params []string, index int) string {
	if index < len(params) {
		return params[index]
	}
	return ""
}
