package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// PieParser handles parsing of pie chart diagrams.
type PieParser struct{}

// NewPieParser creates a new pie parser.
func NewPieParser() *PieParser {
	return &PieParser{}
}

var (
	pieHeaderRegex = regexp.MustCompile(`^pie\s*(?:(showData)\s*)?(?:title\s+(.+))?$`)
	pieEntryRegex  = regexp.MustCompile(`^\s*"([^"]+)"\s*:\s*([0-9]+(?:\.[0-9]{1,2})?)\s*$`)
)

// Parse parses a pie chart diagram source.
func (p *PieParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	diagram := &ast.PieDiagram{
		Type:        "pie",
		Source:      source,
		DataEntries: []ast.PieEntry{},
		Pos:         ast.Position{Line: 1, Column: 1},
	}

	// Parse header line
	firstLine := strings.TrimSpace(lines[0])
	matches := pieHeaderRegex.FindStringSubmatch(firstLine)
	if matches == nil {
		return nil, fmt.Errorf("invalid pie chart header: %s", firstLine)
	}

	// Check for showData modifier
	if matches[1] == "showData" {
		diagram.ShowData = true
	}

	// Extract title if present
	if matches[2] != "" {
		diagram.Title = strings.TrimSpace(matches[2])
	}

	// Parse data entries
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		// Parse data entry
		entryMatches := pieEntryRegex.FindStringSubmatch(trimmed)
		if entryMatches == nil {
			return nil, fmt.Errorf("line %d: invalid pie entry format: %s", i+1, trimmed)
		}

		label := entryMatches[1]
		valueStr := entryMatches[2]

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid numeric value: %s", i+1, valueStr)
		}

		if value <= 0 {
			return nil, fmt.Errorf("line %d: pie chart values must be positive (got %f)", i+1, value)
		}

		diagram.DataEntries = append(diagram.DataEntries, ast.PieEntry{
			Label: label,
			Value: value,
			Pos:   ast.Position{Line: i + 1, Column: 1},
		})
	}

	if len(diagram.DataEntries) == 0 {
		return nil, fmt.Errorf("pie chart must have at least one data entry")
	}

	return diagram, nil
}

// SupportedTypes returns the diagram types this parser supports.
func (p *PieParser) SupportedTypes() []string {
	return []string{"pie"}
}
