package validator

import (
	"fmt"
	"regexp"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// GanttRule is a validation rule for Gantt diagrams.
type GanttRule interface {
	Validate(diagram *ast.GanttDiagram) []*ValidationError
}

// ValidateGantt runs validation rules on a Gantt diagram.
func ValidateGantt(diagram *ast.GanttDiagram, strict bool) []*ValidationError {
	rules := GanttDefaultRules()
	if strict {
		rules = GanttStrictRules()
	}

	var errors []*ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(diagram)...)
	}
	return errors
}

// GanttDefaultRules returns the default validation rules for Gantt diagrams.
func GanttDefaultRules() []GanttRule {
	return []GanttRule{
		&NoDuplicateTaskIDsRule{},
		&ValidTaskReferencesRule{},
		&ValidDateFormatRule{},
		&ValidTaskStatusRule{},
	}
}

// GanttStrictRules returns strict validation rules for Gantt diagrams.
func GanttStrictRules() []GanttRule {
	rules := GanttDefaultRules()
	// Add strict-only rules here if needed
	return rules
}

// NoDuplicateTaskIDsRule checks for duplicate task IDs in Gantt chart.
type NoDuplicateTaskIDsRule struct{}

// Validate checks that all task IDs are unique.
func (r *NoDuplicateTaskIDsRule) Validate(diagram *ast.GanttDiagram) []*ValidationError {
	checker := NewDuplicateChecker("task ID")
	var errors []*ValidationError

	for _, section := range diagram.Sections {
		for _, task := range section.Tasks {
			if task.ID != "" {
				if err := checker.Check(task.ID, task.Pos); err != nil {
					errors = append(errors, err)
				}
			}
		}
	}

	return errors
}

// ValidTaskReferencesRule checks that task dependencies reference existing tasks.
type ValidTaskReferencesRule struct{}

// Validate checks that all task dependencies are valid.
func (r *ValidTaskReferencesRule) Validate(diagram *ast.GanttDiagram) []*ValidationError {
	refChecker := NewReferenceChecker("task")
	var errors []*ValidationError

	// First pass: collect all defined task IDs
	for _, section := range diagram.Sections {
		for _, task := range section.Tasks {
			if task.ID != "" {
				refChecker.Add(task.ID)
			}
		}
	}

	// Second pass: check all dependencies
	for _, section := range diagram.Sections {
		for _, task := range section.Tasks {
			for _, dep := range task.Dependencies {
				if err := refChecker.Check(dep, task.Pos, fmt.Sprintf("task %q", task.Name)); err != nil {
					errors = append(errors, err)
				}
			}
		}
	}

	return errors
}

// ValidDateFormatRule checks that the dateFormat is valid.
type ValidDateFormatRule struct{}

var validDateFormatRegex = regexp.MustCompile(`^[YMDHmsSs\-/:\. ]+$`)

// Validate checks that the date format is valid.
func (r *ValidDateFormatRule) Validate(diagram *ast.GanttDiagram) []*ValidationError {
	var errors []*ValidationError

	if diagram.DateFormat != "" {
		// Check for common valid formats
		validFormats := map[string]bool{
			"YYYY-MM-DD":          true,
			"DD-MM-YYYY":          true,
			"MM-DD-YYYY":          true,
			"YYYY/MM/DD":          true,
			"DD/MM/YYYY":          true,
			"MM/DD/YYYY":          true,
			"YYYY-MM-DD HH:mm":    true,
			"YYYY-MM-DD HH:mm:ss": true,
			"DD.MM.YYYY":          true,
			"HH:mm":               true,
		}

		// If not in the common formats, check if it matches the pattern
		if !validFormats[diagram.DateFormat] && !validDateFormatRegex.MatchString(diagram.DateFormat) {
			errors = append(errors, &ValidationError{
				Line:     diagram.Pos.Line,
				Column:   diagram.Pos.Column,
				Message:  fmt.Sprintf("invalid date format %q", diagram.DateFormat),
				Severity: SeverityError,
			})
		}
	}

	return errors
}

// ValidTaskStatusRule checks that task statuses are valid.
type ValidTaskStatusRule struct{}

// Validate checks that all task statuses are valid.
func (r *ValidTaskStatusRule) Validate(diagram *ast.GanttDiagram) []*ValidationError {
	statusValidator := NewEnumValidator("task status", "done", "active", "crit", "milestone")
	var errors []*ValidationError

	for _, section := range diagram.Sections {
		for _, task := range section.Tasks {
			if task.Status != "" {
				if err := statusValidator.Check(task.Status, task.Pos); err != nil {
					errors = append(errors, err)
				}
			}
		}
	}

	return errors
}
