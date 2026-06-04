package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// SankeyParser handles parsing of Sankey diagrams.
type SankeyParser struct{}

// NewSankeyParser creates a new Sankey parser.
func NewSankeyParser() *SankeyParser {
	return &SankeyParser{}
}

// Parse parses a Sankey diagram source.
func (p *SankeyParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	diagram := &ast.SankeyDiagram{
		Type:   "sankey",
		Source: source,
		Links:  []ast.SankeyLink{},
		Pos:    ast.Position{Line: 1, Column: 1},
	}

	// Parse header line
	firstLine := strings.TrimSpace(lines[0])
	if firstLine != "sankey-beta" {
		return nil, fmt.Errorf("invalid Sankey diagram header: expected 'sankey-beta', got %q", firstLine)
	}

	// Parse link lines
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		// Parse CSV format: source,target,value
		parts := strings.Split(trimmed, ",")
		if len(parts) != 3 {
			return nil, fmt.Errorf("line %d: invalid Sankey link format: expected 'source,target,value', got %q", i+1, trimmed)
		}

		source := strings.TrimSpace(parts[0])
		target := strings.TrimSpace(parts[1])
		valueStr := strings.TrimSpace(parts[2])

		// Validate source and target are not empty
		if source == "" {
			return nil, fmt.Errorf("line %d: source node name cannot be empty", i+1)
		}
		if target == "" {
			return nil, fmt.Errorf("line %d: target node name cannot be empty", i+1)
		}

		// Validate source != target (no self-loops)
		if source == target {
			return nil, fmt.Errorf("line %d: self-loop detected: source and target cannot be the same (%q)", i+1, source)
		}

		// Parse and validate value
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid numeric value: %s", i+1, valueStr)
		}

		if value <= 0 {
			return nil, fmt.Errorf("line %d: Sankey link value must be positive (got %f)", i+1, value)
		}

		diagram.Links = append(diagram.Links, ast.SankeyLink{
			Source: source,
			Target: target,
			Value:  value,
			Pos:    ast.Position{Line: i + 1, Column: 1},
		})
	}

	if len(diagram.Links) == 0 {
		return nil, fmt.Errorf("sankey diagram must have at least one link")
	}

	return diagram, nil
}

// SupportedTypes returns the diagram types this parser supports.
func (p *SankeyParser) SupportedTypes() []string {
	return []string{"sankey"}
}
