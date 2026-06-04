package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// DuplicateChecker helps detect duplicate identifiers in diagrams.
type DuplicateChecker struct {
	seen     map[string]ast.Position
	itemType string // e.g., "class", "state", "participant"
}

// NewDuplicateChecker creates a new duplicate checker for the given item type.
func NewDuplicateChecker(itemType string) *DuplicateChecker {
	return &DuplicateChecker{
		seen:     make(map[string]ast.Position),
		itemType: itemType,
	}
}

// Check checks if an identifier is a duplicate and returns an error if so.
// Returns nil if the identifier is new (not a duplicate).
func (dc *DuplicateChecker) Check(id string, pos ast.Position) *ValidationError {
	if firstPos, exists := dc.seen[id]; exists {
		return &ValidationError{
			Line:     pos.Line,
			Column:   pos.Column,
			Message:  fmt.Sprintf("duplicate %s %q (first defined at line %d)", dc.itemType, id, firstPos.Line),
			Severity: SeverityError,
		}
	}
	dc.seen[id] = pos
	return nil
}

// ReferenceChecker helps validate that referenced items exist.
type ReferenceChecker struct {
	defined  map[string]bool
	itemType string // e.g., "class", "state", "node"
}

// NewReferenceChecker creates a new reference checker for the given item type.
func NewReferenceChecker(itemType string) *ReferenceChecker {
	return &ReferenceChecker{
		defined:  make(map[string]bool),
		itemType: itemType,
	}
}

// Add marks an identifier as defined.
func (rc *ReferenceChecker) Add(id string) {
	rc.defined[id] = true
}

// Check checks if an identifier is defined and returns an error if not.
// Returns nil if the identifier exists.
func (rc *ReferenceChecker) Check(id string, pos ast.Position, context string) *ValidationError {
	if !rc.defined[id] {
		message := fmt.Sprintf("%s references undefined %s %q", context, rc.itemType, id)
		return &ValidationError{
			Line:     pos.Line,
			Column:   pos.Column,
			Message:  message,
			Severity: SeverityError,
		}
	}
	return nil
}

// EnumValidator validates that values are in an allowed set.
type EnumValidator struct {
	allowed   map[string]bool
	valueType string // e.g., "visibility", "relationship type", "direction"
}

// NewEnumValidator creates a new enum validator for the given value type and allowed values.
func NewEnumValidator(valueType string, allowedValues ...string) *EnumValidator {
	allowed := make(map[string]bool)
	for _, v := range allowedValues {
		allowed[v] = true
	}
	return &EnumValidator{
		allowed:   allowed,
		valueType: valueType,
	}
}

// Check checks if a value is in the allowed set and returns an error if not.
// Returns nil if the value is valid.
func (ev *EnumValidator) Check(value string, pos ast.Position) *ValidationError {
	if !ev.allowed[value] {
		return &ValidationError{
			Line:     pos.Line,
			Column:   pos.Column,
			Message:  fmt.Sprintf("invalid %s %q", ev.valueType, value),
			Severity: SeverityError,
		}
	}
	return nil
}
