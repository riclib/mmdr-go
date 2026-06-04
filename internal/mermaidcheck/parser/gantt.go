package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// GanttParser handles parsing of Gantt chart diagrams.
type GanttParser struct{}

// NewGanttParser creates a new Gantt parser.
func NewGanttParser() *GanttParser {
	return &GanttParser{}
}

var (
	ganttHeaderRegex      = regexp.MustCompile(`^gantt\s*$`)
	ganttTitleRegex       = regexp.MustCompile(`^\s*title\s+(.+)$`)
	ganttDateFormatRegex  = regexp.MustCompile(`^\s*dateFormat\s+(.+)$`)
	ganttAxisFormatRegex  = regexp.MustCompile(`^\s*axisFormat\s+(.+)$`)
	ganttExcludesRegex    = regexp.MustCompile(`^\s*excludes\s+(.+)$`)
	ganttTodayMarkerRegex = regexp.MustCompile(`^\s*todayMarker\s+(on|off|#?[0-9a-fA-F]{3,6})\s*$`)
	ganttSectionRegex     = regexp.MustCompile(`^\s*section\s+(.+)$`)
	// Task format: name : [status,] [id,] [dependencies,] start, end/duration
	ganttTaskRegex = regexp.MustCompile(`^\s*([^:]+?)\s*:\s*(.+)$`)
)

// Parse parses a Gantt chart diagram source.
func (p *GanttParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	diagram := &ast.GanttDiagram{
		Type:       "gantt",
		DateFormat: "YYYY-MM-DD", // Default date format
		Source:     source,
		Sections:   []ast.GanttSection{},
		Pos:        ast.Position{Line: 1, Column: 1},
	}

	// Parse header line
	firstLine := strings.TrimSpace(lines[0])
	if !ganttHeaderRegex.MatchString(firstLine) {
		return nil, fmt.Errorf("invalid gantt diagram header: %s", firstLine)
	}

	var currentSection *ast.GanttSection
	hasContent := false

	// Parse subsequent lines
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		// Check for title
		if matches := ganttTitleRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.Title = strings.TrimSpace(matches[1])
			hasContent = true
			continue
		}

		// Check for dateFormat
		if matches := ganttDateFormatRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.DateFormat = strings.TrimSpace(matches[1])
			hasContent = true
			continue
		}

		// Check for axisFormat
		if matches := ganttAxisFormatRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.AxisFormat = strings.TrimSpace(matches[1])
			hasContent = true
			continue
		}

		// Check for excludes
		if matches := ganttExcludesRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.Excludes = strings.TrimSpace(matches[1])
			hasContent = true
			continue
		}

		// Check for todayMarker
		if matches := ganttTodayMarkerRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.TodayMarker = strings.TrimSpace(matches[1])
			hasContent = true
			continue
		}

		// Check for section
		if matches := ganttSectionRegex.FindStringSubmatch(trimmed); matches != nil {
			// Save previous section if exists
			if currentSection != nil {
				diagram.Sections = append(diagram.Sections, *currentSection)
			}
			// Create new section
			currentSection = &ast.GanttSection{
				Name:  strings.TrimSpace(matches[1]),
				Tasks: []ast.GanttTask{},
				Pos:   ast.Position{Line: i + 1, Column: 1},
			}
			hasContent = true
			continue
		}

		// Check for task
		if matches := ganttTaskRegex.FindStringSubmatch(trimmed); matches != nil {
			if currentSection == nil {
				return nil, fmt.Errorf("line %d: task defined outside of section", i+1)
			}

			taskName := strings.TrimSpace(matches[1])
			taskParams := strings.TrimSpace(matches[2])

			// Parse task parameters
			task, err := parseGanttTask(taskName, taskParams, i+1)
			if err != nil {
				return nil, err
			}

			currentSection.Tasks = append(currentSection.Tasks, task)
			hasContent = true
			continue
		}

		// If we get here, it's an invalid line
		return nil, fmt.Errorf("line %d: invalid gantt syntax: %s", i+1, trimmed)
	}

	// Save last section if exists
	if currentSection != nil {
		diagram.Sections = append(diagram.Sections, *currentSection)
	}

	// Validate we have at least some content
	if !hasContent {
		return nil, fmt.Errorf("gantt diagram must have at least a title or one section with tasks")
	}

	return diagram, nil
}

// parseGanttTask parses task parameters from the colon-separated format.
// Format can be: status, id, start, end/duration
// Or: status, id, after deps, end/duration
// Or: id, start, end/duration
// Or: start, end/duration
func parseGanttTask(name, params string, lineNum int) (ast.GanttTask, error) {
	task := ast.GanttTask{
		Name:         name,
		Dependencies: []string{},
		Pos:          ast.Position{Line: lineNum, Column: 1},
	}

	// Split parameters by comma
	parts := strings.Split(params, ",")
	if len(parts) < 1 {
		return task, fmt.Errorf("line %d: task must have at least duration", lineNum)
	}

	// Trim all parts
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// Special case: single parameter (duration only) - defaults to starting after previous task
	if len(parts) == 1 {
		// Validate it looks like a duration (ends with 'd' or 'w') or a date
		param := parts[0]
		if strings.HasSuffix(param, "d") || strings.HasSuffix(param, "w") || looksLikeDate(param) {
			task.EndDate = param
			return task, nil
		}
		// Otherwise it's likely an ID without duration, which is invalid
		return task, fmt.Errorf("line %d: task must have at least duration", lineNum)
	}

	// Parse parameters - we need to detect which format is being used
	// Status keywords: done, active, crit, milestone
	statusKeywords := map[string]bool{
		"done":      true,
		"active":    true,
		"crit":      true,
		"milestone": true,
	}

	currentIdx := 0

	// Check for status (first param is a status keyword)
	if len(parts) > 2 && statusKeywords[parts[0]] {
		task.Status = parts[0]
		currentIdx++
	}

	// Check for ID (next param doesn't start with "after" and isn't a date)
	// ID is typically a simple alphanumeric identifier
	if currentIdx < len(parts)-1 {
		nextParam := parts[currentIdx]
		// If it doesn't look like a date or "after", it's likely an ID
		if !strings.HasPrefix(nextParam, "after") && !looksLikeDate(nextParam) && !strings.HasSuffix(nextParam, "d") && !strings.HasSuffix(nextParam, "w") {
			task.ID = nextParam
			currentIdx++
		}
	}

	// Parse start date or dependency (required)
	if currentIdx >= len(parts) {
		return task, fmt.Errorf("line %d: task missing start date", lineNum)
	}

	// Check if this is an "after" dependency
	if after, ok := strings.CutPrefix(parts[currentIdx], "after "); ok {
		// Parse dependencies
		afterStr := after
		depParts := strings.Fields(afterStr)
		task.Dependencies = depParts
		task.StartDate = parts[currentIdx] // Store the full "after" clause
		currentIdx++
	} else {
		// It's a regular start date
		task.StartDate = parts[currentIdx]
		currentIdx++
	}

	// Parse end date/duration (required)
	if currentIdx >= len(parts) {
		return task, fmt.Errorf("line %d: task missing end date or duration", lineNum)
	}
	task.EndDate = parts[currentIdx]

	return task, nil
}

// looksLikeDate checks if a string looks like a date format.
func looksLikeDate(s string) bool {
	// Simple heuristic: contains dashes or slashes and numbers
	return (strings.Contains(s, "-") || strings.Contains(s, "/")) &&
		regexp.MustCompile(`\d`).MatchString(s)
}

// SupportedTypes returns the diagram types this parser supports.
func (p *GanttParser) SupportedTypes() []string {
	return []string{"gantt"}
}
