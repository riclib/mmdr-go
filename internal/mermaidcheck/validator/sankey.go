package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// SankeyRule is a validation rule for Sankey diagrams.
type SankeyRule interface {
	Validate(diagram *ast.SankeyDiagram) []*ValidationError
}

// ValidateSankey runs validation rules on a Sankey diagram.
func ValidateSankey(diagram *ast.SankeyDiagram, strict bool) []*ValidationError {
	rules := SankeyDefaultRules()
	if strict {
		rules = SankeyStrictRules()
	}

	var errors []*ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(diagram)...)
	}
	return errors
}

// SankeyDefaultRules returns the default validation rules for Sankey diagrams.
func SankeyDefaultRules() []SankeyRule {
	return []SankeyRule{
		&SankeyPositiveValuesRule{},
		&SankeyNoSelfLoopsRule{},
		&SankeyValidNodeReferencesRule{},
		&SankeyMinimumLinksRule{},
	}
}

// SankeyStrictRules returns strict validation rules for Sankey diagrams.
func SankeyStrictRules() []SankeyRule {
	rules := SankeyDefaultRules()
	// Add strict-only rules here if needed
	return rules
}

// SankeyPositiveValuesRule checks that all link values are positive.
type SankeyPositiveValuesRule struct{}

// Validate checks that all values are greater than zero.
func (r *SankeyPositiveValuesRule) Validate(diagram *ast.SankeyDiagram) []*ValidationError {
	var errors []*ValidationError

	for _, link := range diagram.Links {
		if link.Value <= 0 {
			errors = append(errors, &ValidationError{
				Line:     link.Pos.Line,
				Column:   link.Pos.Column,
				Message:  fmt.Sprintf("Sankey link value must be positive (got %f for %q -> %q)", link.Value, link.Source, link.Target),
				Severity: SeverityError,
			})
		}
	}

	return errors
}

// SankeyNoSelfLoopsRule checks that no link has the same source and target.
type SankeyNoSelfLoopsRule struct{}

// Validate checks that source and target are different for all links.
func (r *SankeyNoSelfLoopsRule) Validate(diagram *ast.SankeyDiagram) []*ValidationError {
	var errors []*ValidationError

	for _, link := range diagram.Links {
		if link.Source == link.Target {
			errors = append(errors, &ValidationError{
				Line:     link.Pos.Line,
				Column:   link.Pos.Column,
				Message:  fmt.Sprintf("self-loop detected: source and target cannot be the same (%q)", link.Source),
				Severity: SeverityError,
			})
		}
	}

	return errors
}

// SankeyValidNodeReferencesRule checks that all node names are non-empty strings.
type SankeyValidNodeReferencesRule struct{}

// Validate checks that source and target node names are valid.
func (r *SankeyValidNodeReferencesRule) Validate(diagram *ast.SankeyDiagram) []*ValidationError {
	var errors []*ValidationError

	for _, link := range diagram.Links {
		if link.Source == "" {
			errors = append(errors, &ValidationError{
				Line:     link.Pos.Line,
				Column:   link.Pos.Column,
				Message:  "source node name cannot be empty",
				Severity: SeverityError,
			})
		}
		if link.Target == "" {
			errors = append(errors, &ValidationError{
				Line:     link.Pos.Line,
				Column:   link.Pos.Column,
				Message:  "target node name cannot be empty",
				Severity: SeverityError,
			})
		}
	}

	return errors
}

// SankeyMinimumLinksRule checks that the diagram has at least one link.
type SankeyMinimumLinksRule struct{}

// Validate checks that there is at least one link.
func (r *SankeyMinimumLinksRule) Validate(diagram *ast.SankeyDiagram) []*ValidationError {
	var errors []*ValidationError

	if len(diagram.Links) == 0 {
		errors = append(errors, &ValidationError{
			Line:     1,
			Column:   1,
			Message:  "Sankey diagram must have at least one link",
			Severity: SeverityError,
		})
	}

	return errors
}
