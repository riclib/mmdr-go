package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// JourneyParser handles parsing of user journey diagrams.
type JourneyParser struct{}

// NewJourneyParser creates a new journey parser.
func NewJourneyParser() *JourneyParser {
	return &JourneyParser{}
}

var (
	journeyHeaderRegex  = regexp.MustCompile(`^journey\s*$`)
	journeyTitleRegex   = regexp.MustCompile(`^\s*title\s+(.+)$`)
	journeySectionRegex = regexp.MustCompile(`^\s*section\s+(.+)$`)
	journeyTaskRegex    = regexp.MustCompile(`^\s*([^:]+):\s*(\d+)\s*:\s*(.+)$`)
)

// Parse parses a user journey diagram source.
func (p *JourneyParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	diagram := &ast.JourneyDiagram{
		Type:     "journey",
		Source:   source,
		Sections: []ast.Section{},
		Pos:      ast.Position{Line: 1, Column: 1},
	}

	// Parse header line
	firstLine := strings.TrimSpace(lines[0])
	if !journeyHeaderRegex.MatchString(firstLine) {
		return nil, fmt.Errorf("invalid journey diagram header: %s", firstLine)
	}

	var currentSection *ast.Section
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
		if matches := journeyTitleRegex.FindStringSubmatch(trimmed); matches != nil {
			diagram.Title = strings.TrimSpace(matches[1])
			hasContent = true
			continue
		}

		// Check for section
		if matches := journeySectionRegex.FindStringSubmatch(trimmed); matches != nil {
			// Save previous section if exists
			if currentSection != nil {
				diagram.Sections = append(diagram.Sections, *currentSection)
			}
			// Create new section
			currentSection = &ast.Section{
				Name:  strings.TrimSpace(matches[1]),
				Tasks: []ast.Task{},
				Pos:   ast.Position{Line: i + 1, Column: 1},
			}
			hasContent = true
			continue
		}

		// Check for task
		if matches := journeyTaskRegex.FindStringSubmatch(trimmed); matches != nil {
			if currentSection == nil {
				return nil, fmt.Errorf("line %d: task defined outside of section", i+1)
			}

			taskName := strings.TrimSpace(matches[1])
			scoreStr := matches[2]
			actorsStr := strings.TrimSpace(matches[3])

			// Parse score
			score, err := strconv.Atoi(scoreStr)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid score value: %s", i+1, scoreStr)
			}

			// Validate score range
			if score < 1 || score > 5 {
				return nil, fmt.Errorf("line %d: score must be between 1 and 5 (got %d)", i+1, score)
			}

			// Parse actors (comma-separated)
			actorsParts := strings.Split(actorsStr, ",")
			actors := make([]string, 0, len(actorsParts))
			for _, actor := range actorsParts {
				trimmedActor := strings.TrimSpace(actor)
				if trimmedActor != "" {
					actors = append(actors, trimmedActor)
				}
			}

			if len(actors) == 0 {
				return nil, fmt.Errorf("line %d: task must have at least one actor", i+1)
			}

			currentSection.Tasks = append(currentSection.Tasks, ast.Task{
				Name:   taskName,
				Score:  score,
				Actors: actors,
				Pos:    ast.Position{Line: i + 1, Column: 1},
			})
			hasContent = true
			continue
		}

		// If we get here, it's an invalid line
		return nil, fmt.Errorf("line %d: invalid journey syntax: %s", i+1, trimmed)
	}

	// Save last section if exists
	if currentSection != nil {
		diagram.Sections = append(diagram.Sections, *currentSection)
	}

	// Validate we have at least some content
	if !hasContent {
		return nil, fmt.Errorf("journey diagram must have at least a title or one section with tasks")
	}

	return diagram, nil
}

// SupportedTypes returns the diagram types this parser supports.
func (p *JourneyParser) SupportedTypes() []string {
	return []string{"journey"}
}
