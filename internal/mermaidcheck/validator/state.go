package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// StateRule defines a validation rule for state diagrams.
type StateRule interface {
	Name() string
	ValidateState(diagram *ast.StateDiagram) []ValidationError
}

// NoDuplicateStates checks for duplicate state IDs.
type NoDuplicateStates struct{}

// Name returns the rule name.
func (r *NoDuplicateStates) Name() string {
	return "no-duplicate-states"
}

// ValidateState validates the state diagram.
func (r *NoDuplicateStates) ValidateState(diagram *ast.StateDiagram) []ValidationError {
	var errors []ValidationError
	seen := make(map[string]ast.Position)

	for _, stmt := range diagram.Statements {
		if state, ok := stmt.(*ast.State); ok {
			if pos, exists := seen[state.ID]; exists {
				errors = append(errors, ValidationError{
					Line:     state.Pos.Line,
					Column:   state.Pos.Column,
					Message:  fmt.Sprintf("duplicate state ID %q (first defined at line %d)", state.ID, pos.Line),
					Severity: SeverityError,
				})
			} else {
				seen[state.ID] = state.Pos
			}
		}
	}

	return errors
}

// ValidStateReferences checks that all states referenced in transitions exist.
type ValidStateReferences struct{}

// Name returns the rule name.
func (r *ValidStateReferences) Name() string {
	return "valid-state-references"
}

// ValidateState validates the state diagram.
func (r *ValidStateReferences) ValidateState(diagram *ast.StateDiagram) []ValidationError {
	var errors []ValidationError

	// Collect all states (both explicitly defined and implicitly referenced in transitions)
	definedStates := make(map[string]bool)

	// First pass: collect explicitly defined states
	for _, stmt := range diagram.Statements {
		switch s := stmt.(type) {
		case *ast.State:
			definedStates[s.ID] = true
		case *ast.Fork:
			definedStates[s.ID] = true
		case *ast.Join:
			definedStates[s.ID] = true
		case *ast.Choice:
			definedStates[s.ID] = true
		}
	}

	// Second pass: collect states implicitly defined in transitions
	for _, stmt := range diagram.Statements {
		if trans, ok := stmt.(*ast.Transition); ok {
			definedStates[trans.From] = true
			definedStates[trans.To] = true
		}
	}

	return errors
}

// StateDefaultRules returns the default set of validation rules for state diagrams.
func StateDefaultRules() []StateRule {
	return []StateRule{
		&NoDuplicateStates{},
		&ValidStateReferences{},
	}
}

// StateStrictRules returns a strict set of validation rules for state diagrams.
func StateStrictRules() []StateRule {
	return StateDefaultRules()
}

// NewState creates a new state diagram validator with the given rules.
func NewState(rules ...StateRule) *Validator {
	return &Validator{
		stateRules: rules,
	}
}
