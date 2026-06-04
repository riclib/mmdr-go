// Package validator provides validation and linting for Mermaid diagrams.
package validator

import (
	"fmt"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// Severity represents the severity level of a validation error.
type Severity int

const (
	// SeverityError indicates a critical error that prevents diagram rendering.
	SeverityError Severity = iota
	// SeverityWarning indicates a potential issue that may affect diagram quality.
	SeverityWarning
	// SeverityInfo provides informational messages about the diagram.
	SeverityInfo
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

// ValidationError represents a validation error with position and context.
type ValidationError struct {
	Line     int      // Line number (1-indexed)
	Column   int      // Column number (1-indexed)
	Message  string   // Error message
	Severity Severity // Error severity
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("line %d: %s: %s", v.Line, v.Severity, v.Message)
}

// Rule represents a validation rule that can be applied to a flowchart.
type Rule interface {
	// Name returns the name of the rule.
	Name() string
	// Validate checks the flowchart and returns any validation errors.
	Validate(flowchart *ast.Flowchart) []ValidationError
}

// Validator validates Mermaid diagrams using a set of rules.
type Validator struct {
	rules         []Rule
	genericRules  []GenericRule
	sequenceRules []SequenceRule
	classRules    []ClassRule
	stateRules    []StateRule
}

// New creates a new validator with the given rules for flowcharts.
func New(rules ...Rule) *Validator {
	return &Validator{rules: rules}
}

// NewGeneric creates a new validator with generic rules for any diagram type.
func NewGeneric(rules ...GenericRule) *Validator {
	return &Validator{genericRules: rules}
}

// NewSequence creates a new validator with sequence diagram rules.
func NewSequence(rules ...SequenceRule) *Validator {
	return &Validator{sequenceRules: rules}
}

// Validate runs all validation rules on the flowchart.
func (v *Validator) Validate(flowchart *ast.Flowchart) []ValidationError {
	var errors []ValidationError
	for _, rule := range v.rules {
		errors = append(errors, rule.Validate(flowchart)...)
	}
	return errors
}

// ValidateDiagram validates any diagram type using the Diagram interface.
func (v *Validator) ValidateDiagram(diagram ast.Diagram) []ValidationError {
	switch d := diagram.(type) {
	case *ast.Flowchart:
		return v.Validate(d)
	case *ast.SequenceDiagram:
		var errors []ValidationError
		for _, rule := range v.sequenceRules {
			errors = append(errors, rule.ValidateSequence(d)...)
		}
		return errors
	case *ast.ClassDiagram:
		var errors []ValidationError
		for _, rule := range v.classRules {
			errors = append(errors, rule.ValidateClass(d)...)
		}
		return errors
	case *ast.StateDiagram:
		var errors []ValidationError
		for _, rule := range v.stateRules {
			errors = append(errors, rule.ValidateState(d)...)
		}
		return errors
	case *ast.GenericDiagram:
		var errors []ValidationError
		for _, rule := range v.genericRules {
			errors = append(errors, rule.ValidateGeneric(d)...)
		}
		return errors
	default:
		return []ValidationError{{
			Line:     1,
			Column:   1,
			Message:  fmt.Sprintf("unsupported diagram type: %T", diagram),
			Severity: SeverityError,
		}}
	}
}

// ValidDirection checks if the flowchart direction is valid.
type ValidDirection struct{}

// Name returns the name of this validation rule.
func (r *ValidDirection) Name() string { return "valid-direction" }

// Validate checks if the flowchart direction is one of the valid values.
func (r *ValidDirection) Validate(flowchart *ast.Flowchart) []ValidationError {
	validDirections := map[string]bool{
		"TB": true, "TD": true, "BT": true, "RL": true, "LR": true,
	}

	if !validDirections[flowchart.Direction] {
		return []ValidationError{{
			Line:     flowchart.Pos.Line,
			Column:   flowchart.Pos.Column,
			Message:  fmt.Sprintf("invalid direction '%s', must be one of: TB, TD, BT, RL, LR", flowchart.Direction),
			Severity: SeverityError,
		}}
	}

	return nil
}

// NoUndefinedNodes checks that all referenced nodes are defined.
type NoUndefinedNodes struct{}

// Name returns the name of this validation rule.
func (r *NoUndefinedNodes) Name() string { return "no-undefined-nodes" }

// Validate checks that all nodes referenced in links are defined.
func (r *NoUndefinedNodes) Validate(flowchart *ast.Flowchart) []ValidationError {
	definedNodes := make(map[string]bool)
	var errors []ValidationError

	// Collect all defined nodes
	r.collectDefinedNodes(flowchart.Statements, definedNodes)

	// Check all links reference defined nodes
	r.checkLinks(flowchart.Statements, definedNodes, &errors)

	return errors
}

func (r *NoUndefinedNodes) collectDefinedNodes(statements []ast.Statement, defined map[string]bool) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.NodeDef:
			defined[s.ID] = true
		case *ast.Link:
			// Links can also implicitly define nodes
			defined[s.From] = true
			defined[s.To] = true
		case *ast.Subgraph:
			r.collectDefinedNodes(s.Statements, defined)
		}
	}
}

func (r *NoUndefinedNodes) checkLinks(statements []ast.Statement, defined map[string]bool, errors *[]ValidationError) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.Link:
			if !defined[s.From] {
				*errors = append(*errors, ValidationError{
					Line:     s.Pos.Line,
					Column:   s.Pos.Column,
					Message:  fmt.Sprintf("undefined node '%s' in link", s.From),
					Severity: SeverityError,
				})
			}
			if !defined[s.To] {
				*errors = append(*errors, ValidationError{
					Line:     s.Pos.Line,
					Column:   s.Pos.Column,
					Message:  fmt.Sprintf("undefined node '%s' in link", s.To),
					Severity: SeverityError,
				})
			}
		case *ast.Subgraph:
			r.checkLinks(s.Statements, defined, errors)
		}
	}
}

// NoParenthesesInLabels checks that node labels don't contain parentheses.
type NoParenthesesInLabels struct{}

// Name returns the name of this validation rule.
func (r *NoParenthesesInLabels) Name() string { return "no-parentheses-in-labels" }

// Validate checks that no node labels contain parentheses.
func (r *NoParenthesesInLabels) Validate(flowchart *ast.Flowchart) []ValidationError {
	var errors []ValidationError
	r.checkStatements(flowchart.Statements, &errors)
	return errors
}

func (r *NoParenthesesInLabels) checkStatements(statements []ast.Statement, errors *[]ValidationError) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.NodeDef:
			if strings.Contains(s.Label, "(") || strings.Contains(s.Label, ")") {
				*errors = append(*errors, ValidationError{
					Line:     s.Pos.Line,
					Column:   s.Pos.Column,
					Message:  fmt.Sprintf("node label '%s' contains parentheses, use <br/> for line breaks instead", s.Label),
					Severity: SeverityWarning,
				})
			}
		case *ast.Subgraph:
			r.checkStatements(s.Statements, errors)
		}
	}
}

// NoDuplicateNodeIDs checks that node IDs are unique.
type NoDuplicateNodeIDs struct{}

// Name returns the name of this validation rule.
func (r *NoDuplicateNodeIDs) Name() string { return "no-duplicate-node-ids" }

// Validate checks that all node IDs are unique within the flowchart.
func (r *NoDuplicateNodeIDs) Validate(flowchart *ast.Flowchart) []ValidationError {
	nodePositions := make(map[string]ast.Position)
	var errors []ValidationError

	r.checkDuplicates(flowchart.Statements, nodePositions, &errors)

	return errors
}

func (r *NoDuplicateNodeIDs) checkDuplicates(statements []ast.Statement, positions map[string]ast.Position, errors *[]ValidationError) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.NodeDef:
			if firstPos, exists := positions[s.ID]; exists {
				*errors = append(*errors, ValidationError{
					Line:     s.Pos.Line,
					Column:   s.Pos.Column,
					Message:  fmt.Sprintf("duplicate node ID '%s' (first defined at line %d)", s.ID, firstPos.Line),
					Severity: SeverityWarning,
				})
			} else {
				positions[s.ID] = s.Pos
			}
		case *ast.Subgraph:
			r.checkDuplicates(s.Statements, positions, errors)
		}
	}
}

// DefaultRules returns the default set of validation rules.
func DefaultRules() []Rule {
	return []Rule{
		&ValidDirection{},
		&NoUndefinedNodes{},
		&NoDuplicateNodeIDs{},
	}
}

// StrictRules returns a strict set of validation rules including style checks.
func StrictRules() []Rule {
	return []Rule{
		&ValidDirection{},
		&NoUndefinedNodes{},
		&NoDuplicateNodeIDs{},
		&NoParenthesesInLabels{},
	}
}
