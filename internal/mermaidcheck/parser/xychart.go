package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// XYChartParser handles parsing of XY chart diagrams.
type XYChartParser struct{}

// NewXYChartParser creates a new XY chart parser.
func NewXYChartParser() *XYChartParser {
	return &XYChartParser{}
}

var (
	xyChartHeaderRegex     = regexp.MustCompile(`^xychart-beta\s*(horizontal|vertical)?\s*$`)
	xyChartTitleRegex      = regexp.MustCompile(`^\s*title\s+"([^"]+)"\s*$`)
	xyChartXAxisCatRegex   = regexp.MustCompile(`^\s*x-axis\s+\[(.+)\]\s*$`)
	xyChartYAxisCatRegex   = regexp.MustCompile(`^\s*y-axis\s+\[(.+)\]\s*$`)
	xyChartXAxisNumRegex   = regexp.MustCompile(`^\s*x-axis\s+"([^"]+)"\s+(-?[0-9]+(?:\.[0-9]+)?)\s+-->\s+(-?[0-9]+(?:\.[0-9]+)?)\s*$`)
	xyChartYAxisNumRegex   = regexp.MustCompile(`^\s*y-axis\s+"([^"]+)"\s+(-?[0-9]+(?:\.[0-9]+)?)\s+-->\s+(-?[0-9]+(?:\.[0-9]+)?)\s*$`)
	xyChartBarSeriesRegex  = regexp.MustCompile(`^\s*bar\s+\[(.+)\]\s*$`)
	xyChartLineSeriesRegex = regexp.MustCompile(`^\s*line\s+\[(.+)\]\s*$`)
)

// Parse parses an XY chart diagram source.
func (p *XYChartParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	diagram := &ast.XYChartDiagram{
		Type:        "xyChart",
		Orientation: "vertical", // Default orientation
		Source:      source,
		Series:      []ast.XYChartSeries{},
		Pos:         ast.Position{Line: 1, Column: 1},
	}

	// Parse header line
	firstLine := strings.TrimSpace(lines[0])
	matches := xyChartHeaderRegex.FindStringSubmatch(firstLine)
	if matches == nil {
		return nil, fmt.Errorf("invalid xychart header: %s", firstLine)
	}

	// Set orientation if specified
	if matches[1] != "" {
		diagram.Orientation = matches[1]
	}

	xAxisDefined := false
	yAxisDefined := false

	// Parse remaining lines
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		lineNum := i + 1

		// Try to parse title
		if matches := xyChartTitleRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.Title = matches[1]
			continue
		}

		// Try to parse categorical x-axis
		if matches := xyChartXAxisCatRegex.FindStringSubmatch(trimmed); matches != nil {
			if xAxisDefined {
				return nil, fmt.Errorf("line %d: x-axis already defined", lineNum)
			}
			categories := parseCategories(matches[1])
			diagram.XAxis = ast.XYChartAxis{
				Categories: categories,
				IsNumeric:  false,
				Pos:        ast.Position{Line: lineNum, Column: 1},
			}
			xAxisDefined = true
			continue
		}

		// Try to parse numeric x-axis
		if matches := xyChartXAxisNumRegex.FindStringSubmatch(trimmed); matches != nil {
			if xAxisDefined {
				return nil, fmt.Errorf("line %d: x-axis already defined", lineNum)
			}
			minVal, err := strconv.ParseFloat(matches[2], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid x-axis minimum: %s", lineNum, matches[2])
			}
			maxVal, err := strconv.ParseFloat(matches[3], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid x-axis maximum: %s", lineNum, matches[3])
			}
			diagram.XAxis = ast.XYChartAxis{
				Label:     matches[1],
				Min:       minVal,
				Max:       maxVal,
				IsNumeric: true,
				Pos:       ast.Position{Line: lineNum, Column: 1},
			}
			xAxisDefined = true
			continue
		}

		// Try to parse categorical y-axis
		if matches := xyChartYAxisCatRegex.FindStringSubmatch(trimmed); matches != nil {
			if yAxisDefined {
				return nil, fmt.Errorf("line %d: y-axis already defined", lineNum)
			}
			categories := parseCategories(matches[1])
			diagram.YAxis = ast.XYChartAxis{
				Categories: categories,
				IsNumeric:  false,
				Pos:        ast.Position{Line: lineNum, Column: 1},
			}
			yAxisDefined = true
			continue
		}

		// Try to parse numeric y-axis
		if matches := xyChartYAxisNumRegex.FindStringSubmatch(trimmed); matches != nil {
			if yAxisDefined {
				return nil, fmt.Errorf("line %d: y-axis already defined", lineNum)
			}
			minVal, err := strconv.ParseFloat(matches[2], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid y-axis minimum: %s", lineNum, matches[2])
			}
			maxVal, err := strconv.ParseFloat(matches[3], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid y-axis maximum: %s", lineNum, matches[3])
			}
			diagram.YAxis = ast.XYChartAxis{
				Label:     matches[1],
				Min:       minVal,
				Max:       maxVal,
				IsNumeric: true,
				Pos:       ast.Position{Line: lineNum, Column: 1},
			}
			yAxisDefined = true
			continue
		}

		// Try to parse bar series
		if matches := xyChartBarSeriesRegex.FindStringSubmatch(trimmed); matches != nil {
			values, err := parseValues(matches[1])
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", lineNum, err)
			}
			diagram.Series = append(diagram.Series, ast.XYChartSeries{
				Type:   "bar",
				Values: values,
				Pos:    ast.Position{Line: lineNum, Column: 1},
			})
			continue
		}

		// Try to parse line series
		if matches := xyChartLineSeriesRegex.FindStringSubmatch(trimmed); matches != nil {
			values, err := parseValues(matches[1])
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", lineNum, err)
			}
			diagram.Series = append(diagram.Series, ast.XYChartSeries{
				Type:   "line",
				Values: values,
				Pos:    ast.Position{Line: lineNum, Column: 1},
			})
			continue
		}

		// Unknown line format
		return nil, fmt.Errorf("line %d: unrecognised xychart syntax: %s", lineNum, trimmed)
	}

	// Validate required elements
	if !xAxisDefined {
		return nil, fmt.Errorf("xychart must define an x-axis")
	}
	if !yAxisDefined {
		return nil, fmt.Errorf("xychart must define a y-axis")
	}
	if len(diagram.Series) == 0 {
		return nil, fmt.Errorf("xychart must have at least one data series")
	}

	return diagram, nil
}

// parseCategories parses a comma-separated list of categories.
func parseCategories(input string) []string {
	parts := strings.Split(input, ",")
	categories := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			categories = append(categories, trimmed)
		}
	}
	return categories
}

// parseValues parses a comma-separated list of numeric values.
func parseValues(input string) ([]float64, error) {
	parts := strings.Split(input, ",")
	values := make([]float64, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		value, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid numeric value: %s", trimmed)
		}
		values = append(values, value)
	}
	return values, nil
}

// SupportedTypes returns the diagram types this parser supports.
func (p *XYChartParser) SupportedTypes() []string {
	return []string{"xyChart"}
}
