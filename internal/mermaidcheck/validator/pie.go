package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// PieRule is a validation rule for pie diagrams.
type PieRule interface {
	Validate(diagram *ast.PieDiagram) []*ValidationError
}

// ValidatePie runs validation rules on a pie diagram.
func ValidatePie(diagram *ast.PieDiagram, strict bool) []*ValidationError {
	rules := PieDefaultRules()
	if strict {
		rules = PieStrictRules()
	}

	var errors []*ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(diagram)...)
	}
	return errors
}

// PieDefaultRules returns the default validation rules for pie diagrams.
func PieDefaultRules() []PieRule {
	return []PieRule{
		&NoDuplicateLabelsRule{},
		&PositiveValuesRule{},
	}
}

// PieStrictRules returns strict validation rules for pie diagrams.
func PieStrictRules() []PieRule {
	rules := PieDefaultRules()
	// Add strict-only rules here if needed
	return rules
}

// NoDuplicateLabelsRule checks for duplicate labels in pie chart.
type NoDuplicateLabelsRule struct{}

// Validate checks that all labels are unique.
func (r *NoDuplicateLabelsRule) Validate(diagram *ast.PieDiagram) []*ValidationError {
	checker := NewDuplicateChecker("label")
	var errors []*ValidationError

	for _, entry := range diagram.DataEntries {
		if err := checker.Check(entry.Label, entry.Pos); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// PositiveValuesRule checks that all values are positive.
type PositiveValuesRule struct{}

// Validate checks that all values are greater than zero.
func (r *PositiveValuesRule) Validate(diagram *ast.PieDiagram) []*ValidationError {
	var errors []*ValidationError

	for _, entry := range diagram.DataEntries {
		if entry.Value <= 0 {
			errors = append(errors, &ValidationError{
				Line:     entry.Pos.Line,
				Column:   entry.Pos.Column,
				Message:  fmt.Sprintf("pie chart value for %q must be positive (got %f)", entry.Label, entry.Value),
				Severity: SeverityError,
			})
		}
	}

	return errors
}
