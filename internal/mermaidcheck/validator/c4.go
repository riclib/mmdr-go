package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// C4Rule is the interface for C4 diagram validation rules.
type C4Rule interface {
	Validate(d *ast.C4Diagram) []ValidationError
}

// ValidateC4 validates a C4 diagram using the provided rules.
func ValidateC4(d *ast.C4Diagram, rules []C4Rule) []ValidationError {
	var errors []ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(d)...)
	}
	return errors
}

// DefaultC4Rules returns the default validation rules for C4 diagrams.
func DefaultC4Rules() []C4Rule {
	return []C4Rule{
		&NoDuplicateElementIDsRule{},
		&C4ValidRelationshipReferencesRule{},
		&ValidBoundaryIDsRule{},
		&ValidStyleReferencesRule{},
	}
}

// StrictC4Rules returns strict validation rules for C4 diagrams.
func StrictC4Rules() []C4Rule {
	return DefaultC4Rules()
}

// NoDuplicateElementIDsRule checks that all element IDs are unique.
type NoDuplicateElementIDsRule struct{}

// Validate checks for duplicate element IDs across all elements and boundaries.
func (r *NoDuplicateElementIDsRule) Validate(d *ast.C4Diagram) []ValidationError {
	checker := NewDuplicateChecker("element")
	var errors []ValidationError

	// Check top-level elements
	for _, elem := range d.Elements {
		if err := checker.Check(elem.ID, elem.Pos); err != nil {
			errors = append(errors, *err)
		}
	}

	// Check boundary IDs and their nested elements recursively
	errors = append(errors, checkBoundaryElementDuplicates(d.Boundaries, checker)...)

	return errors
}

// checkBoundaryElementDuplicates recursively checks for duplicate IDs in boundaries and their elements.
func checkBoundaryElementDuplicates(boundaries []ast.C4Boundary, checker *DuplicateChecker) []ValidationError {
	var errors []ValidationError

	for _, boundary := range boundaries {
		// Check boundary ID
		if err := checker.Check(boundary.ID, boundary.Pos); err != nil {
			errors = append(errors, *err)
		}

		// Check elements in boundary
		for _, elem := range boundary.Elements {
			if err := checker.Check(elem.ID, elem.Pos); err != nil {
				errors = append(errors, *err)
			}
		}

		// Check nested boundaries recursively
		errors = append(errors, checkBoundaryElementDuplicates(boundary.Boundaries, checker)...)
	}

	return errors
}

// C4ValidRelationshipReferencesRule checks that all relationship references point to defined elements.
type C4ValidRelationshipReferencesRule struct{}

// Validate checks that relationship from/to references exist.
func (r *C4ValidRelationshipReferencesRule) Validate(d *ast.C4Diagram) []ValidationError {
	// Collect all valid element IDs
	validIDs := make(map[string]bool)

	for _, elem := range d.Elements {
		validIDs[elem.ID] = true
	}

	collectBoundaryIDs(d.Boundaries, validIDs)

	// Check all relationships
	var errors []ValidationError
	for _, rel := range d.Relationships {
		if !validIDs[rel.From] {
			errors = append(errors, ValidationError{
				Line:     rel.Pos.Line,
				Column:   rel.Pos.Column,
				Message:  fmt.Sprintf("relationship references undefined element '%s'", rel.From),
				Severity: SeverityError,
			})
		}
		if !validIDs[rel.To] {
			errors = append(errors, ValidationError{
				Line:     rel.Pos.Line,
				Column:   rel.Pos.Column,
				Message:  fmt.Sprintf("relationship references undefined element '%s'", rel.To),
				Severity: SeverityError,
			})
		}
	}

	return errors
}

// collectBoundaryIDs recursively collects all element IDs from boundaries.
func collectBoundaryIDs(boundaries []ast.C4Boundary, validIDs map[string]bool) {
	for _, boundary := range boundaries {
		validIDs[boundary.ID] = true
		for _, elem := range boundary.Elements {
			validIDs[elem.ID] = true
		}
		collectBoundaryIDs(boundary.Boundaries, validIDs)
	}
}

// ValidBoundaryIDsRule checks that all boundary IDs are unique.
type ValidBoundaryIDsRule struct{}

// Validate checks for duplicate boundary IDs.
func (r *ValidBoundaryIDsRule) Validate(d *ast.C4Diagram) []ValidationError {
	checker := NewDuplicateChecker("boundary")
	return checkBoundaryDuplicates(d.Boundaries, checker)
}

// checkBoundaryDuplicates recursively checks for duplicate boundary IDs.
func checkBoundaryDuplicates(boundaries []ast.C4Boundary, checker *DuplicateChecker) []ValidationError {
	var errors []ValidationError

	for _, boundary := range boundaries {
		if err := checker.Check(boundary.ID, boundary.Pos); err != nil {
			errors = append(errors, *err)
		}
		errors = append(errors, checkBoundaryDuplicates(boundary.Boundaries, checker)...)
	}

	return errors
}

// ValidStyleReferencesRule checks that style overrides reference defined elements.
type ValidStyleReferencesRule struct{}

// Validate checks that style references point to existing elements.
func (r *ValidStyleReferencesRule) Validate(d *ast.C4Diagram) []ValidationError {
	// Collect all valid element IDs
	validIDs := make(map[string]bool)

	for _, elem := range d.Elements {
		validIDs[elem.ID] = true
	}

	collectBoundaryIDs(d.Boundaries, validIDs)

	// Check all styles
	var errors []ValidationError
	for _, style := range d.Styles {
		switch style.StyleType {
		case "UpdateElementStyle":
			if !validIDs[style.ElementID] {
				errors = append(errors, ValidationError{
					Line:     style.Pos.Line,
					Column:   style.Pos.Column,
					Message:  fmt.Sprintf("style references undefined element '%s'", style.ElementID),
					Severity: SeverityError,
				})
			}
		case "UpdateRelStyle":
			if !validIDs[style.From] {
				errors = append(errors, ValidationError{
					Line:     style.Pos.Line,
					Column:   style.Pos.Column,
					Message:  fmt.Sprintf("relationship style references undefined element '%s'", style.From),
					Severity: SeverityError,
				})
			}
			if !validIDs[style.To] {
				errors = append(errors, ValidationError{
					Line:     style.Pos.Line,
					Column:   style.Pos.Column,
					Message:  fmt.Sprintf("relationship style references undefined element '%s'", style.To),
					Severity: SeverityError,
				})
			}
		}
	}

	return errors
}
