package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// ERParser handles parsing of ER diagrams.
type ERParser struct{}

// NewERParser creates a new ER parser.
func NewERParser() *ERParser {
	return &ERParser{}
}

var (
	erHeaderRegex     = regexp.MustCompile(`^erDiagram\s*(?:(TB|BT|LR|RL)\s*)?$`)
	entityHeaderRegex = regexp.MustCompile(`^([A-Z_][A-Z0-9_-]*)\s*(?:\[([^\]]+)\])?\s*\{?\s*$`)
	attributeRegex    = regexp.MustCompile(`^\s+([a-zA-Z][a-zA-Z0-9\-_\(\)\[\]]*)\s+([a-zA-Z_][a-zA-Z0-9_-]*|\*[a-zA-Z_][a-zA-Z0-9_-]*)\s*(?:([A-Z,]+))?\s*(?:"([^"]*)")?\s*$`)
	relationshipRegex = regexp.MustCompile(`^([A-Z_][A-Z0-9_-]*)\s+(\|\||\|o|\}\||\}o)(--|\.\.)(\|\||\|o|o\||o\{|\|\{|\}\||\}o)\s+([A-Z_][A-Z0-9_-]*)\s*(?::\s*(.+))?$`)
	simpleEntityRegex = regexp.MustCompile(`^([A-Z_][A-Z0-9_-]*)\s*(?:\[([^\]]+)\])?\s*$`)
)

// Parse parses an ER diagram source.
func (p *ERParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	diagram := &ast.ERDiagram{
		Type:          "er",
		Source:        source,
		Entities:      []ast.EREntity{},
		Relationships: []ast.ERRelationship{},
		Pos:           ast.Position{Line: 1, Column: 1},
	}

	// Parse header line
	firstLine := strings.TrimSpace(lines[0])
	matches := erHeaderRegex.FindStringSubmatch(firstLine)
	if matches == nil {
		return nil, fmt.Errorf("invalid ER diagram header: %s", firstLine)
	}

	// Extract direction if present
	if matches[1] != "" {
		diagram.Direction = matches[1]
	}

	// Parse diagram content
	var currentEntity *ast.EREntity
	inEntityBlock := false

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		// Check for closing brace
		if trimmed == "}" {
			if currentEntity != nil {
				diagram.Entities = append(diagram.Entities, *currentEntity)
				currentEntity = nil
				inEntityBlock = false
			}
			continue
		}

		// Try to parse as attribute if we're in an entity block
		if inEntityBlock && currentEntity != nil {
			attrMatches := attributeRegex.FindStringSubmatch(line)
			if attrMatches != nil {
				attr := ast.ERAttribute{
					Type: attrMatches[1],
					Name: attrMatches[2],
					Pos:  ast.Position{Line: i + 1, Column: 1},
				}

				// Handle asterisk notation for primary keys
				if after, ok := strings.CutPrefix(attr.Name, "*"); ok {
					attr.Name = after
					attr.Keys = []string{"PK"}
				}

				// Parse key indicators
				if attrMatches[3] != "" {
					keys := strings.SplitSeq(attrMatches[3], ",")
					for key := range keys {
						key = strings.TrimSpace(key)
						if key == "PK" || key == "FK" || key == "UK" {
							if len(attr.Keys) == 0 || (len(attr.Keys) == 1 && attr.Keys[0] != key) {
								attr.Keys = append(attr.Keys, key)
							}
						}
					}
				}

				// Parse comment
				if attrMatches[4] != "" {
					attr.Comment = attrMatches[4]
				}

				currentEntity.Attributes = append(currentEntity.Attributes, attr)
				continue
			}
		}

		// Try to parse as relationship
		relMatches := relationshipRegex.FindStringSubmatch(trimmed)
		if relMatches != nil {
			rel := ast.ERRelationship{
				From:     relMatches[1],
				FromCard: relMatches[2],
				Type:     relMatches[3],
				ToCard:   relMatches[4],
				To:       relMatches[5],
				Pos:      ast.Position{Line: i + 1, Column: 1},
			}
			if relMatches[6] != "" {
				rel.Label = strings.TrimSpace(relMatches[6])
			}
			diagram.Relationships = append(diagram.Relationships, rel)
			continue
		}

		// Try to parse as entity header
		entityMatches := entityHeaderRegex.FindStringSubmatch(trimmed)
		if entityMatches != nil {
			// Save previous entity if it exists
			if currentEntity != nil {
				diagram.Entities = append(diagram.Entities, *currentEntity)
			}

			currentEntity = &ast.EREntity{
				Name:       entityMatches[1],
				Attributes: []ast.ERAttribute{},
				Pos:        ast.Position{Line: i + 1, Column: 1},
			}

			if entityMatches[2] != "" {
				currentEntity.Alias = entityMatches[2]
			}

			// Check if this is a block entity (has opening brace)
			if strings.HasSuffix(trimmed, "{") {
				inEntityBlock = true
			} else {
				// Simple entity without attributes
				diagram.Entities = append(diagram.Entities, *currentEntity)
				currentEntity = nil
			}
			continue
		}

		// Try to parse as simple entity (no block)
		simpleMatches := simpleEntityRegex.FindStringSubmatch(trimmed)
		if simpleMatches != nil {
			entity := ast.EREntity{
				Name:       simpleMatches[1],
				Attributes: []ast.ERAttribute{},
				Pos:        ast.Position{Line: i + 1, Column: 1},
			}
			if simpleMatches[2] != "" {
				entity.Alias = simpleMatches[2]
			}
			diagram.Entities = append(diagram.Entities, entity)
			continue
		}

		return nil, fmt.Errorf("line %d: invalid ER diagram syntax: %s", i+1, trimmed)
	}

	// Save final entity if exists
	if currentEntity != nil {
		diagram.Entities = append(diagram.Entities, *currentEntity)
	}

	return diagram, nil
}

// SupportedTypes returns the diagram types this parser supports.
func (p *ERParser) SupportedTypes() []string {
	return []string{"erDiagram"}
}
