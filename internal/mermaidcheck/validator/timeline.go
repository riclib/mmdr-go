package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// TimelineRule is a validation rule for timeline diagrams.
type TimelineRule interface {
	Validate(diagram *ast.TimelineDiagram) []*ValidationError
}

// ValidateTimeline runs validation rules on a timeline diagram.
func ValidateTimeline(diagram *ast.TimelineDiagram, strict bool) []*ValidationError {
	rules := TimelineDefaultRules()
	if strict {
		rules = TimelineStrictRules()
	}

	var errors []*ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(diagram)...)
	}
	return errors
}

// TimelineDefaultRules returns the default validation rules for timeline diagrams.
func TimelineDefaultRules() []TimelineRule {
	return []TimelineRule{
		&PeriodsHaveEventsRule{},
		&NoEmptyPeriodsRule{},
	}
}

// TimelineStrictRules returns strict validation rules for timeline diagrams.
func TimelineStrictRules() []TimelineRule {
	rules := TimelineDefaultRules()
	// Add strict-only rules here if needed
	return rules
}

// PeriodsHaveEventsRule checks that all periods have at least one event.
type PeriodsHaveEventsRule struct{}

// Validate checks that all periods have events.
func (r *PeriodsHaveEventsRule) Validate(diagram *ast.TimelineDiagram) []*ValidationError {
	var errors []*ValidationError

	for _, section := range diagram.Sections {
		for _, period := range section.Periods {
			if len(period.Events) == 0 {
				errors = append(errors, &ValidationError{
					Line:     period.Pos.Line,
					Column:   period.Pos.Column,
					Message:  fmt.Sprintf("time period %q has no events", period.TimePeriod),
					Severity: SeverityError,
				})
			}
		}
	}

	return errors
}

// NoEmptyPeriodsRule checks that period names and events are not empty strings.
type NoEmptyPeriodsRule struct{}

// Validate checks that periods and events have content.
func (r *NoEmptyPeriodsRule) Validate(diagram *ast.TimelineDiagram) []*ValidationError {
	var errors []*ValidationError

	for _, section := range diagram.Sections {
		for _, period := range section.Periods {
			// Check period name
			if period.TimePeriod == "" {
				errors = append(errors, &ValidationError{
					Line:     period.Pos.Line,
					Column:   period.Pos.Column,
					Message:  "time period cannot be empty",
					Severity: SeverityError,
				})
			}

			// Check events
			for i, event := range period.Events {
				if event == "" {
					errors = append(errors, &ValidationError{
						Line:     period.Pos.Line,
						Column:   period.Pos.Column,
						Message:  fmt.Sprintf("event %d in period %q is empty", i+1, period.TimePeriod),
						Severity: SeverityError,
					})
				}
			}
		}
	}

	return errors
}
