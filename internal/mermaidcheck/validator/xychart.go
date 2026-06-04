package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// XYChartRule is a validation rule for XY chart diagrams.
type XYChartRule interface {
	Validate(diagram *ast.XYChartDiagram) []*ValidationError
}

// ValidateXYChart runs validation rules on an XY chart diagram.
func ValidateXYChart(diagram *ast.XYChartDiagram, strict bool) []*ValidationError {
	rules := XYChartDefaultRules()
	if strict {
		rules = XYChartStrictRules()
	}

	var errors []*ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(diagram)...)
	}
	return errors
}

// XYChartDefaultRules returns the default validation rules for XY chart diagrams.
func XYChartDefaultRules() []XYChartRule {
	return []XYChartRule{
		&XYChartXAxisDefinedRule{},
		&XYChartYAxisDefinedRule{},
		&XYChartMinimumSeriesRule{},
		&XYChartValidSeriesLengthRule{},
		&XYChartValidOrientationRule{},
	}
}

// XYChartStrictRules returns strict validation rules for XY chart diagrams.
func XYChartStrictRules() []XYChartRule {
	rules := XYChartDefaultRules()
	// Add strict-only rules here if needed
	return rules
}

// XYChartXAxisDefinedRule checks that x-axis is defined.
type XYChartXAxisDefinedRule struct{}

// Validate checks that the x-axis is defined.
func (r *XYChartXAxisDefinedRule) Validate(diagram *ast.XYChartDiagram) []*ValidationError {
	if !diagram.XAxis.IsNumeric && len(diagram.XAxis.Categories) == 0 {
		return []*ValidationError{
			{
				Line:     diagram.Pos.Line,
				Column:   diagram.Pos.Column,
				Message:  "xychart must define an x-axis",
				Severity: SeverityError,
			},
		}
	}
	return nil
}

// XYChartYAxisDefinedRule checks that y-axis is defined.
type XYChartYAxisDefinedRule struct{}

// Validate checks that the y-axis is defined.
func (r *XYChartYAxisDefinedRule) Validate(diagram *ast.XYChartDiagram) []*ValidationError {
	if !diagram.YAxis.IsNumeric && len(diagram.YAxis.Categories) == 0 {
		return []*ValidationError{
			{
				Line:     diagram.Pos.Line,
				Column:   diagram.Pos.Column,
				Message:  "xychart must define a y-axis",
				Severity: SeverityError,
			},
		}
	}
	return nil
}

// XYChartMinimumSeriesRule checks that at least one data series is defined.
type XYChartMinimumSeriesRule struct{}

// Validate checks that at least one data series is defined.
func (r *XYChartMinimumSeriesRule) Validate(diagram *ast.XYChartDiagram) []*ValidationError {
	if len(diagram.Series) == 0 {
		return []*ValidationError{
			{
				Line:     diagram.Pos.Line,
				Column:   diagram.Pos.Column,
				Message:  "xychart must have at least one data series",
				Severity: SeverityError,
			},
		}
	}
	return nil
}

// XYChartValidSeriesLengthRule checks that all series have the same number of values.
type XYChartValidSeriesLengthRule struct{}

// Validate checks that all data series have consistent lengths.
func (r *XYChartValidSeriesLengthRule) Validate(diagram *ast.XYChartDiagram) []*ValidationError {
	if len(diagram.Series) == 0 {
		return nil
	}

	var errors []*ValidationError
	expectedLength := len(diagram.Series[0].Values)

	for i, series := range diagram.Series {
		if len(series.Values) != expectedLength {
			errors = append(errors, &ValidationError{
				Line:     series.Pos.Line,
				Column:   series.Pos.Column,
				Message:  fmt.Sprintf("series %d has %d values, expected %d values to match first series", i+1, len(series.Values), expectedLength),
				Severity: SeverityError,
			})
		}
	}

	// Check against categorical axis if present
	if !diagram.XAxis.IsNumeric && len(diagram.XAxis.Categories) > 0 {
		if expectedLength != len(diagram.XAxis.Categories) {
			errors = append(errors, &ValidationError{
				Line:     diagram.XAxis.Pos.Line,
				Column:   diagram.XAxis.Pos.Column,
				Message:  fmt.Sprintf("series have %d values but x-axis has %d categories", expectedLength, len(diagram.XAxis.Categories)),
				Severity: SeverityWarning,
			})
		}
	}

	if !diagram.YAxis.IsNumeric && len(diagram.YAxis.Categories) > 0 {
		if expectedLength != len(diagram.YAxis.Categories) {
			errors = append(errors, &ValidationError{
				Line:     diagram.YAxis.Pos.Line,
				Column:   diagram.YAxis.Pos.Column,
				Message:  fmt.Sprintf("series have %d values but y-axis has %d categories", expectedLength, len(diagram.YAxis.Categories)),
				Severity: SeverityWarning,
			})
		}
	}

	return errors
}

// XYChartValidOrientationRule checks that orientation is valid.
type XYChartValidOrientationRule struct{}

// Validate checks that the orientation is either "horizontal" or "vertical".
func (r *XYChartValidOrientationRule) Validate(diagram *ast.XYChartDiagram) []*ValidationError {
	if diagram.Orientation != "horizontal" && diagram.Orientation != "vertical" {
		return []*ValidationError{
			{
				Line:     diagram.Pos.Line,
				Column:   diagram.Pos.Column,
				Message:  fmt.Sprintf("invalid orientation %q, must be 'horizontal' or 'vertical'", diagram.Orientation),
				Severity: SeverityError,
			},
		}
	}
	return nil
}
