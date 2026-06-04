package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// TimelineParser handles parsing of timeline diagrams.
type TimelineParser struct{}

// NewTimelineParser creates a new timeline parser.
func NewTimelineParser() *TimelineParser {
	return &TimelineParser{}
}

var (
	timelineTitleRegex   = regexp.MustCompile(`^\s*title\s+(.+)$`)
	timelineSectionRegex = regexp.MustCompile(`^\s*section\s+(.+)$`)
	timelinePeriodRegex  = regexp.MustCompile(`^\s*([^:]+?)\s*:\s*(.+)$`)
	timelineEventRegex   = regexp.MustCompile(`^\s*:\s*(.+)$`)
)

// Parse parses a timeline diagram source.
func (p *TimelineParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	// Verify header
	firstLine := strings.TrimSpace(lines[0])
	if firstLine != "timeline" {
		return nil, fmt.Errorf("invalid timeline header: expected 'timeline', got %q", firstLine)
	}

	diagram := &ast.TimelineDiagram{
		Type:     "timeline",
		Source:   source,
		Sections: []ast.TimelineSection{},
		Pos:      ast.Position{Line: 1, Column: 1},
	}

	// Start with default section (no name)
	currentSection := &ast.TimelineSection{
		Name:    "",
		Periods: []ast.TimelinePeriod{},
		Pos:     ast.Position{Line: 1, Column: 1},
	}

	var currentPeriod *ast.TimelinePeriod

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		lineNum := i + 1

		// Check for title
		if matches := timelineTitleRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.Title = strings.TrimSpace(matches[1])
			continue
		}

		// Check for section
		if matches := timelineSectionRegex.FindStringSubmatch(trimmed); matches != nil {
			// Save current period to current section before switching sections
			if currentPeriod != nil {
				currentSection.Periods = append(currentSection.Periods, *currentPeriod)
			}

			// Save current section if it has periods
			if len(currentSection.Periods) > 0 {
				diagram.Sections = append(diagram.Sections, *currentSection)
			}

			// Start new section
			currentSection = &ast.TimelineSection{
				Name:    strings.TrimSpace(matches[1]),
				Periods: []ast.TimelinePeriod{},
				Pos:     ast.Position{Line: lineNum, Column: 1},
			}
			currentPeriod = nil
			continue
		}

		// Check for continuation event (leading colon)
		if matches := timelineEventRegex.FindStringSubmatch(trimmed); matches != nil {
			if currentPeriod == nil {
				return nil, fmt.Errorf("line %d: event continuation without time period", lineNum)
			}
			event := strings.TrimSpace(matches[1])
			if event == "" {
				return nil, fmt.Errorf("line %d: empty event", lineNum)
			}
			currentPeriod.Events = append(currentPeriod.Events, event)
			continue
		}

		// Check for period with events
		if matches := timelinePeriodRegex.FindStringSubmatch(trimmed); matches != nil {
			// Save previous period if exists
			if currentPeriod != nil {
				currentSection.Periods = append(currentSection.Periods, *currentPeriod)
			}

			timePeriod := strings.TrimSpace(matches[1])
			if timePeriod == "" {
				return nil, fmt.Errorf("line %d: empty time period", lineNum)
			}

			// Parse events (colon-separated)
			eventsStr := strings.TrimSpace(matches[2])
			events := []string{}
			if eventsStr != "" {
				for event := range strings.SplitSeq(eventsStr, ":") {
					event = strings.TrimSpace(event)
					if event != "" {
						events = append(events, event)
					}
				}
			}

			if len(events) == 0 {
				return nil, fmt.Errorf("line %d: time period %q has no events", lineNum, timePeriod)
			}

			currentPeriod = &ast.TimelinePeriod{
				TimePeriod: timePeriod,
				Events:     events,
				Pos:        ast.Position{Line: lineNum, Column: 1},
			}
			continue
		}

		return nil, fmt.Errorf("line %d: invalid timeline syntax: %s", lineNum, trimmed)
	}

	// Save last period if exists
	if currentPeriod != nil {
		currentSection.Periods = append(currentSection.Periods, *currentPeriod)
	}

	// Save last section if it has periods
	if len(currentSection.Periods) > 0 {
		diagram.Sections = append(diagram.Sections, *currentSection)
	}

	// Validate we have at least one period
	if len(diagram.Sections) == 0 {
		return nil, fmt.Errorf("timeline must have at least one time period")
	}

	return diagram, nil
}

// SupportedTypes returns the diagram types this parser supports.
func (p *TimelineParser) SupportedTypes() []string {
	return []string{"timeline"}
}
