package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// JourneyRule is a validation rule for journey diagrams.
type JourneyRule interface {
	Validate(diagram *ast.JourneyDiagram) []*ValidationError
}

// ValidateJourney runs validation rules on a journey diagram.
func ValidateJourney(diagram *ast.JourneyDiagram, strict bool) []*ValidationError {
	rules := JourneyDefaultRules()
	if strict {
		rules = JourneyStrictRules()
	}

	var errors []*ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(diagram)...)
	}
	return errors
}

// JourneyDefaultRules returns the default validation rules for journey diagrams.
func JourneyDefaultRules() []JourneyRule {
	return []JourneyRule{
		&ValidTaskScoresRule{},
		&TasksHaveActorsRule{},
	}
}

// JourneyStrictRules returns strict validation rules for journey diagrams.
func JourneyStrictRules() []JourneyRule {
	rules := JourneyDefaultRules()
	// Add strict-only rules here if needed
	return rules
}

// ValidTaskScoresRule checks that all task scores are within valid range (1-5).
type ValidTaskScoresRule struct{}

// Validate checks that all task scores are between 1 and 5.
func (r *ValidTaskScoresRule) Validate(diagram *ast.JourneyDiagram) []*ValidationError {
	var errors []*ValidationError

	for _, section := range diagram.Sections {
		for _, task := range section.Tasks {
			if task.Score < 1 || task.Score > 5 {
				errors = append(errors, &ValidationError{
					Line:     task.Pos.Line,
					Column:   task.Pos.Column,
					Message:  fmt.Sprintf("task %q has invalid score %d (must be between 1 and 5)", task.Name, task.Score),
					Severity: SeverityError,
				})
			}
		}
	}

	return errors
}

// TasksHaveActorsRule checks that all tasks have at least one actor.
type TasksHaveActorsRule struct{}

// Validate checks that all tasks have at least one actor assigned.
func (r *TasksHaveActorsRule) Validate(diagram *ast.JourneyDiagram) []*ValidationError {
	var errors []*ValidationError

	for _, section := range diagram.Sections {
		for _, task := range section.Tasks {
			if len(task.Actors) == 0 {
				errors = append(errors, &ValidationError{
					Line:     task.Pos.Line,
					Column:   task.Pos.Column,
					Message:  fmt.Sprintf("task %q must have at least one actor", task.Name),
					Severity: SeverityError,
				})
			}
		}
	}

	return errors
}
