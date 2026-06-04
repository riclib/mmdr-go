package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// QuadrantRule is a validation rule for quadrant diagrams.
type QuadrantRule interface {
	Validate(diagram *ast.QuadrantDiagram) []*ValidationError
}

// ValidateQuadrant runs validation rules on a quadrant diagram.
func ValidateQuadrant(diagram *ast.QuadrantDiagram, strict bool) []*ValidationError {
	rules := QuadrantDefaultRules()
	if strict {
		rules = QuadrantStrictRules()
	}

	var errors []*ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(diagram)...)
	}
	return errors
}

// QuadrantDefaultRules returns the default validation rules for quadrant diagrams.
func QuadrantDefaultRules() []QuadrantRule {
	return []QuadrantRule{
		&ValidCoordinatesRule{},
		&NoDuplicatePointNamesRule{},
		&QuadrantXAxisDefinedRule{},
		&QuadrantYAxisDefinedRule{},
		&MinimumPointsRule{},
	}
}

// QuadrantStrictRules returns strict validation rules for quadrant diagrams.
func QuadrantStrictRules() []QuadrantRule {
	rules := QuadrantDefaultRules()
	// Add strict-only rules here if needed
	return rules
}

// ValidCoordinatesRule checks that all coordinates are between 0.0 and 1.0.
type ValidCoordinatesRule struct{}

// Validate checks that X and Y coordinates are within valid range.
func (r *ValidCoordinatesRule) Validate(diagram *ast.QuadrantDiagram) []*ValidationError {
	var errors []*ValidationError

	for _, point := range diagram.Points {
		if point.X < 0.0 || point.X > 1.0 {
			errors = append(errors, &ValidationError{
				Line:     point.Pos.Line,
				Column:   point.Pos.Column,
				Message:  fmt.Sprintf("X coordinate for %q must be between 0.0 and 1.0 (got %f)", point.Name, point.X),
				Severity: SeverityError,
			})
		}

		if point.Y < 0.0 || point.Y > 1.0 {
			errors = append(errors, &ValidationError{
				Line:     point.Pos.Line,
				Column:   point.Pos.Column,
				Message:  fmt.Sprintf("Y coordinate for %q must be between 0.0 and 1.0 (got %f)", point.Name, point.Y),
				Severity: SeverityError,
			})
		}
	}

	return errors
}

// NoDuplicatePointNamesRule checks for duplicate point names.
type NoDuplicatePointNamesRule struct{}

// Validate checks that all point names are unique.
func (r *NoDuplicatePointNamesRule) Validate(diagram *ast.QuadrantDiagram) []*ValidationError {
	checker := NewDuplicateChecker("point name")
	var errors []*ValidationError

	for _, point := range diagram.Points {
		if err := checker.Check(point.Name, point.Pos); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// QuadrantXAxisDefinedRule checks that x-axis is defined.
type QuadrantXAxisDefinedRule struct{}

// Validate checks that the x-axis has been defined.
func (r *QuadrantXAxisDefinedRule) Validate(diagram *ast.QuadrantDiagram) []*ValidationError {
	if diagram.XAxis.Min == "" && diagram.XAxis.Max == "" {
		return []*ValidationError{
			{
				Line:     diagram.Pos.Line,
				Column:   diagram.Pos.Column,
				Message:  "quadrant chart must define x-axis",
				Severity: SeverityError,
			},
		}
	}

	return nil
}

// QuadrantYAxisDefinedRule checks that y-axis is defined.
type QuadrantYAxisDefinedRule struct{}

// Validate checks that the y-axis has been defined.
func (r *QuadrantYAxisDefinedRule) Validate(diagram *ast.QuadrantDiagram) []*ValidationError {
	if diagram.YAxis.Min == "" && diagram.YAxis.Max == "" {
		return []*ValidationError{
			{
				Line:     diagram.Pos.Line,
				Column:   diagram.Pos.Column,
				Message:  "quadrant chart must define y-axis",
				Severity: SeverityError,
			},
		}
	}

	return nil
}

// MinimumPointsRule checks that at least one data point exists.
type MinimumPointsRule struct{}

// Validate checks that the diagram has at least one data point.
func (r *MinimumPointsRule) Validate(diagram *ast.QuadrantDiagram) []*ValidationError {
	if len(diagram.Points) == 0 {
		return []*ValidationError{
			{
				Line:     diagram.Pos.Line,
				Column:   diagram.Pos.Column,
				Message:  "quadrant chart must have at least one data point",
				Severity: SeverityError,
			},
		}
	}

	return nil
}
