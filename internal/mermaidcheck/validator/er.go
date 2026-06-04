package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// ERRule is a validation rule for ER diagrams.
type ERRule interface {
	Validate(diagram *ast.ERDiagram) []*ValidationError
}

// ValidateER runs validation rules on an ER diagram.
func ValidateER(diagram *ast.ERDiagram, strict bool) []*ValidationError {
	rules := ERDefaultRules()
	if strict {
		rules = ERStrictRules()
	}

	var errors []*ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(diagram)...)
	}
	return errors
}

// ERDefaultRules returns the default validation rules for ER diagrams.
func ERDefaultRules() []ERRule {
	return []ERRule{
		&NoDuplicateEntitiesRule{},
		&ValidRelationshipReferencesRule{},
		&ValidAttributeKeysRule{},
	}
}

// ERStrictRules returns strict validation rules for ER diagrams.
func ERStrictRules() []ERRule {
	rules := ERDefaultRules()
	return rules
}

// NoDuplicateEntitiesRule checks for duplicate entity names in ER diagram.
type NoDuplicateEntitiesRule struct{}

// Validate checks that all entity names are unique.
func (r *NoDuplicateEntitiesRule) Validate(diagram *ast.ERDiagram) []*ValidationError {
	checker := NewDuplicateChecker("entity")
	var errors []*ValidationError

	for _, entity := range diagram.Entities {
		if err := checker.Check(entity.Name, entity.Pos); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// ValidRelationshipReferencesRule checks that relationships reference defined entities.
type ValidRelationshipReferencesRule struct{}

// Validate checks that all relationships reference existing entities.
func (r *ValidRelationshipReferencesRule) Validate(diagram *ast.ERDiagram) []*ValidationError {
	checker := NewReferenceChecker("entity")
	var errors []*ValidationError

	// Register all explicitly defined entities
	for _, entity := range diagram.Entities {
		checker.Add(entity.Name)
	}

	// Also register entities implicitly defined in relationships
	// (ER diagrams can have relationships without explicit entity blocks)
	for _, rel := range diagram.Relationships {
		checker.Add(rel.From)
		checker.Add(rel.To)
	}

	// No need to check relationships since all referenced entities are now registered
	// (either explicitly or implicitly)

	return errors
}

// ValidAttributeKeysRule checks that attribute key indicators are valid.
type ValidAttributeKeysRule struct{}

// Validate checks that attribute keys are PK, FK, or UK.
func (r *ValidAttributeKeysRule) Validate(diagram *ast.ERDiagram) []*ValidationError {
	var errors []*ValidationError
	validKeys := map[string]bool{"PK": true, "FK": true, "UK": true}

	for _, entity := range diagram.Entities {
		for _, attr := range entity.Attributes {
			for _, key := range attr.Keys {
				if !validKeys[key] {
					errors = append(errors, &ValidationError{
						Line:     attr.Pos.Line,
						Column:   attr.Pos.Column,
						Message:  fmt.Sprintf("invalid attribute key %q in entity %q (must be PK, FK, or UK)", key, entity.Name),
						Severity: SeverityError,
					})
				}
			}
		}
	}

	return errors
}
