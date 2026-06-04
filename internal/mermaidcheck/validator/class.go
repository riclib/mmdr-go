package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// ClassRule defines a validation rule for class diagrams.
type ClassRule interface {
	Name() string
	ValidateClass(diagram *ast.ClassDiagram) []ValidationError
}

// NoDuplicateClasses checks for duplicate class names.
type NoDuplicateClasses struct{}

// Name returns the rule name.
func (r *NoDuplicateClasses) Name() string {
	return "no-duplicate-classes"
}

// ValidateClass validates the class diagram.
func (r *NoDuplicateClasses) ValidateClass(diagram *ast.ClassDiagram) []ValidationError {
	var errors []ValidationError
	seen := make(map[string]ast.Position)

	for _, stmt := range diagram.Statements {
		if class, ok := stmt.(*ast.Class); ok {
			if pos, exists := seen[class.Name]; exists {
				errors = append(errors, ValidationError{
					Line:     class.Pos.Line,
					Column:   class.Pos.Column,
					Message:  fmt.Sprintf("duplicate class name %q (first defined at line %d)", class.Name, pos.Line),
					Severity: SeverityError,
				})
			} else {
				seen[class.Name] = class.Pos
			}
		}
	}

	return errors
}

// ValidClassReferences checks that all classes referenced in relationships exist.
type ValidClassReferences struct{}

// Name returns the rule name.
func (r *ValidClassReferences) Name() string {
	return "valid-class-references"
}

// ValidateClass validates the class diagram.
func (r *ValidClassReferences) ValidateClass(diagram *ast.ClassDiagram) []ValidationError {
	var errors []ValidationError

	// Collect all classes (both explicitly defined and implicitly referenced in relationships)
	definedClasses := make(map[string]bool)

	// First pass: collect explicitly defined classes
	for _, stmt := range diagram.Statements {
		if class, ok := stmt.(*ast.Class); ok {
			definedClasses[class.Name] = true
		}
	}

	// Second pass: collect classes implicitly defined in relationships
	for _, stmt := range diagram.Statements {
		if rel, ok := stmt.(*ast.Relationship); ok {
			definedClasses[rel.From] = true
			definedClasses[rel.To] = true
		}
	}

	// Check notes
	for _, stmt := range diagram.Statements {
		if note, ok := stmt.(*ast.ClassNote); ok {
			if !definedClasses[note.ClassName] {
				errors = append(errors, ValidationError{
					Line:     note.Pos.Line,
					Column:   note.Pos.Column,
					Message:  fmt.Sprintf("note references undefined class %q", note.ClassName),
					Severity: SeverityError,
				})
			}
		}
	}

	return errors
}

// ValidMemberVisibility checks that member visibility modifiers are valid.
type ValidMemberVisibility struct{}

// Name returns the rule name.
func (r *ValidMemberVisibility) Name() string {
	return "valid-member-visibility"
}

// ValidateClass validates the class diagram.
func (r *ValidMemberVisibility) ValidateClass(diagram *ast.ClassDiagram) []ValidationError {
	var errors []ValidationError
	validVisibility := map[string]bool{
		"+": true, // public
		"-": true, // private
		"#": true, // protected
		"~": true, // package
	}

	for _, stmt := range diagram.Statements {
		if class, ok := stmt.(*ast.Class); ok {
			for _, member := range class.Members {
				if !validVisibility[member.Visibility] {
					errors = append(errors, ValidationError{
						Line:     member.Pos.Line,
						Column:   member.Pos.Column,
						Message:  fmt.Sprintf("invalid visibility modifier %q (must be +, -, #, or ~)", member.Visibility),
						Severity: SeverityError,
					})
				}
			}
		}
	}

	return errors
}

// ValidRelationshipType checks that relationship types are valid.
type ValidRelationshipType struct{}

// Name returns the rule name.
func (r *ValidRelationshipType) Name() string {
	return "valid-relationship-type"
}

// ValidateClass validates the class diagram.
func (r *ValidRelationshipType) ValidateClass(diagram *ast.ClassDiagram) []ValidationError {
	var errors []ValidationError
	validTypes := map[string]bool{
		"inheritance": true,
		"composition": true,
		"aggregation": true,
		"association": true,
		"dependency":  true,
		"realization": true,
	}

	for _, stmt := range diagram.Statements {
		if rel, ok := stmt.(*ast.Relationship); ok {
			if !validTypes[rel.Type] {
				errors = append(errors, ValidationError{
					Line:     rel.Pos.Line,
					Column:   rel.Pos.Column,
					Message:  fmt.Sprintf("invalid relationship type %q", rel.Type),
					Severity: SeverityError,
				})
			}
		}
	}

	return errors
}

// ClassDefaultRules returns the default set of validation rules for class diagrams.
func ClassDefaultRules() []ClassRule {
	return []ClassRule{
		&NoDuplicateClasses{},
		&ValidClassReferences{},
		&ValidMemberVisibility{},
		&ValidRelationshipType{},
	}
}

// ClassStrictRules returns a strict set of validation rules for class diagrams.
func ClassStrictRules() []ClassRule {
	return ClassDefaultRules()
}

// NewClass creates a new class diagram validator with the given rules.
func NewClass(rules ...ClassRule) *Validator {
	return &Validator{
		classRules: rules,
	}
}
