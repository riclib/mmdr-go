package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// QuadrantParser handles parsing of quadrant chart diagrams.
type QuadrantParser struct{}

// NewQuadrantParser creates a new quadrant parser.
func NewQuadrantParser() *QuadrantParser {
	return &QuadrantParser{}
}

var (
	quadrantHeaderRegex = regexp.MustCompile(`^quadrantChart\s*$`)
	quadrantTitleRegex  = regexp.MustCompile(`^\s*title\s+(.+)$`)
	quadrantXAxisRegex  = regexp.MustCompile(`^\s*x-axis\s+(.+?)\s+-->\s+(.+)$`)
	quadrantYAxisRegex  = regexp.MustCompile(`^\s*y-axis\s+(.+?)\s+-->\s+(.+)$`)
	quadrantLabelRegex  = regexp.MustCompile(`^\s*quadrant-([1-4])\s+(.+)$`)
	quadrantPointRegex  = regexp.MustCompile(`^\s*(.+?):\s*\[\s*([0-9]+(?:\.[0-9]+)?)\s*,\s*([0-9]+(?:\.[0-9]+)?)\s*\]$`)
)

// Parse parses a quadrant chart diagram source.
func (p *QuadrantParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	diagram := &ast.QuadrantDiagram{
		Type:   "quadrantChart",
		Source: source,
		Points: []ast.QuadrantPoint{},
		Pos:    ast.Position{Line: 1, Column: 1},
	}

	// Parse header line
	firstLine := strings.TrimSpace(lines[0])
	if !quadrantHeaderRegex.MatchString(firstLine) {
		return nil, fmt.Errorf("invalid quadrant chart header: %s", firstLine)
	}

	var xAxisDefined, yAxisDefined bool

	// Parse remaining lines
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		// Try to match title
		if matches := quadrantTitleRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.Title = strings.TrimSpace(matches[1])
			continue
		}

		// Try to match x-axis
		if matches := quadrantXAxisRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.XAxis = ast.QuadrantAxis{
				Min: strings.TrimSpace(matches[1]),
				Max: strings.TrimSpace(matches[2]),
			}
			xAxisDefined = true
			continue
		}

		// Try to match y-axis
		if matches := quadrantYAxisRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.YAxis = ast.QuadrantAxis{
				Min: strings.TrimSpace(matches[1]),
				Max: strings.TrimSpace(matches[2]),
			}
			yAxisDefined = true
			continue
		}

		// Try to match quadrant label
		if matches := quadrantLabelRegex.FindStringSubmatch(trimmed); matches != nil {
			quadrantNum, _ := strconv.Atoi(matches[1])
			diagram.QuadrantLabels[quadrantNum-1] = strings.TrimSpace(matches[2])
			continue
		}

		// Try to match data point
		if matches := quadrantPointRegex.FindStringSubmatch(trimmed); matches != nil {
			name := strings.TrimSpace(matches[1])
			xStr := matches[2]
			yStr := matches[3]

			x, err := strconv.ParseFloat(xStr, 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid X coordinate: %s", i+1, xStr)
			}

			y, err := strconv.ParseFloat(yStr, 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid Y coordinate: %s", i+1, yStr)
			}

			diagram.Points = append(diagram.Points, ast.QuadrantPoint{
				Name: name,
				X:    x,
				Y:    y,
				Pos:  ast.Position{Line: i + 1, Column: 1},
			})
			continue
		}

		// If we reach here, the line doesn't match any known pattern
		return nil, fmt.Errorf("line %d: invalid quadrant chart syntax: %s", i+1, trimmed)
	}

	// Validate required elements
	if !xAxisDefined {
		return nil, fmt.Errorf("quadrant chart must define x-axis")
	}

	if !yAxisDefined {
		return nil, fmt.Errorf("quadrant chart must define y-axis")
	}

	if len(diagram.Points) == 0 {
		return nil, fmt.Errorf("quadrant chart must have at least one data point")
	}

	return diagram, nil
}

// SupportedTypes returns the diagram types this parser supports.
func (p *QuadrantParser) SupportedTypes() []string {
	return []string{"quadrantChart"}
}
