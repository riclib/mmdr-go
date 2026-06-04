package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// MindmapRule is a validation rule for mindmap diagrams.
type MindmapRule interface {
	Validate(diagram *ast.MindmapDiagram) []*ValidationError
}

// ValidateMindmap runs validation rules on a mindmap diagram.
func ValidateMindmap(diagram *ast.MindmapDiagram, strict bool) []*ValidationError {
	rules := MindmapDefaultRules()
	if strict {
		rules = MindmapStrictRules()
	}

	var errors []*ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(diagram)...)
	}
	return errors
}

// MindmapDefaultRules returns the default validation rules for mindmap diagrams.
func MindmapDefaultRules() []MindmapRule {
	return []MindmapRule{
		&RootNodeExistsRule{},
		&NoEmptyNodesRule{},
		&ValidShapeRule{},
	}
}

// MindmapStrictRules returns strict validation rules for mindmap diagrams.
func MindmapStrictRules() []MindmapRule {
	rules := MindmapDefaultRules()
	// Add strict-only rules here if needed
	return rules
}

// RootNodeExistsRule checks that the mindmap has a root node.
type RootNodeExistsRule struct{}

// Validate checks that a root node exists.
func (r *RootNodeExistsRule) Validate(diagram *ast.MindmapDiagram) []*ValidationError {
	if diagram.Root == nil {
		return []*ValidationError{
			{
				Line:     diagram.Pos.Line,
				Column:   diagram.Pos.Column,
				Message:  "mindmap must have a root node",
				Severity: SeverityError,
			},
		}
	}
	return nil
}

// NoEmptyNodesRule checks that all nodes have non-empty text.
type NoEmptyNodesRule struct{}

// Validate checks that all nodes have text content.
func (r *NoEmptyNodesRule) Validate(diagram *ast.MindmapDiagram) []*ValidationError {
	var errors []*ValidationError

	var checkNode func(*ast.MindmapNode)
	checkNode = func(node *ast.MindmapNode) {
		if node == nil {
			return
		}

		if node.Text == "" {
			errors = append(errors, &ValidationError{
				Line:     node.Pos.Line,
				Column:   node.Pos.Column,
				Message:  "mindmap node text cannot be empty",
				Severity: SeverityError,
			})
		}

		for _, child := range node.Children {
			checkNode(child)
		}
	}

	if diagram.Root != nil {
		checkNode(diagram.Root)
	}

	return errors
}

// ValidShapeRule checks that all node shapes are valid.
type ValidShapeRule struct{}

var validShapes = map[string]bool{
	"":     true, // Default (no shape)
	"()":   true,
	"(())": true,
	"[]":   true,
	"{{}}": true,
	"))((": true,
}

// Validate checks that all node shapes are recognised.
func (r *ValidShapeRule) Validate(diagram *ast.MindmapDiagram) []*ValidationError {
	var errors []*ValidationError

	var checkNode func(*ast.MindmapNode)
	checkNode = func(node *ast.MindmapNode) {
		if node == nil {
			return
		}

		if !validShapes[node.Shape] {
			errors = append(errors, &ValidationError{
				Line:     node.Pos.Line,
				Column:   node.Pos.Column,
				Message:  fmt.Sprintf("invalid node shape %q (valid shapes: (), (()), [], {{}}, ))((, or no shape)", node.Shape),
				Severity: SeverityError,
			})
		}

		for _, child := range node.Children {
			checkNode(child)
		}
	}

	if diagram.Root != nil {
		checkNode(diagram.Root)
	}

	return errors
}
