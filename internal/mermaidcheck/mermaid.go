package mermaidcheck

import (
	"fmt"
	"io"
	"os"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
	"github.com/riclib/mmdr-go/internal/mermaidcheck/extractor"
	"github.com/riclib/mmdr-go/internal/mermaidcheck/internal/inpututil"
	"github.com/riclib/mmdr-go/internal/mermaidcheck/parser"
	"github.com/riclib/mmdr-go/internal/mermaidcheck/validator"
)

// Parse parses a raw Mermaid diagram from a string.
// Returns a Diagram interface that can be a Flowchart or GenericDiagram depending on type.
func Parse(source string) (ast.Diagram, error) {
	return parser.Parse(source)
}

// ParseReader parses a raw Mermaid diagram from an io.Reader.
// Returns a Diagram interface that can be a Flowchart or GenericDiagram depending on type.
func ParseReader(r io.Reader) (ast.Diagram, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Parse(string(data))
}

// ParseFlowchart parses a flowchart/graph diagram specifically.
// Use this if you need the full Flowchart AST.
func ParseFlowchart(source string) (*ast.Flowchart, error) {
	p := parser.NewFlowchartParser()
	diagram, err := p.Parse(source)
	if err != nil {
		return nil, err
	}
	flowchart, ok := diagram.(*ast.Flowchart)
	if !ok {
		return nil, fmt.Errorf("parsed diagram is not a flowchart: %T", diagram)
	}
	return flowchart, nil
}

// ParseFile parses a file containing Mermaid diagram(s).
// It auto-detects the file type based on extension and content:
// - .mmd files are parsed as raw Mermaid (unless they contain markdown code fences)
// - .md, .markdown, .mdx files are parsed as markdown and all Mermaid blocks are extracted
// - If a .mmd file contains markdown code fences, it's treated as markdown
//
// Returns a slice of diagrams (potentially multiple for markdown files).
func ParseFile(path string) ([]ast.Diagram, error) {
	data, err := os.ReadFile(path) //nolint:gosec // User-provided file path is intentional
	if err != nil {
		return nil, err
	}

	content := string(data)
	fileType := inpututil.DetectFileType(path)

	// Check if .mmd file contains markdown code fences
	if fileType == inpututil.FileTypeMermaid && containsMarkdownFences(content) {
		fileType = inpututil.FileTypeMarkdown
	}

	switch fileType {
	case inpututil.FileTypeMermaid:
		// Parse as raw Mermaid
		diagram, err := Parse(content)
		if err != nil {
			return nil, err
		}
		return []ast.Diagram{diagram}, nil

	case inpututil.FileTypeMarkdown:
		// Extract and parse Mermaid blocks from markdown
		blocks, err := extractor.ExtractFromMarkdown(content)
		if err != nil {
			return nil, err
		}

		var diagrams []ast.Diagram
		for _, block := range blocks {
			diagram, err := Parse(block.Source)
			if err != nil {
				return nil, fmt.Errorf("error parsing Mermaid block at line %d: %w", block.LineOffset, err)
			}
			diagrams = append(diagrams, diagram)
		}
		return diagrams, nil

	default:
		return nil, fmt.Errorf("unsupported file type for %s", path)
	}
}

// containsMarkdownFences checks if the content contains markdown code fences.
func containsMarkdownFences(content string) bool {
	// Check for ```mermaid or ~~~mermaid code fences
	return len(content) > 10 && (contains(content, "```mermaid") ||
		contains(content, "~~~mermaid") ||
		contains(content, "``` mermaid"))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ExtractFromMarkdown extracts all Mermaid diagram blocks from markdown content.
func ExtractFromMarkdown(markdown string) ([]extractor.DiagramBlock, error) {
	return extractor.ExtractFromMarkdown(markdown)
}

// Validate validates any diagram using the appropriate validator.
// Automatically detects diagram type and applies corresponding rules.
func Validate(diagram ast.Diagram, strict bool) []validator.ValidationError {
	switch d := diagram.(type) {
	case *ast.Flowchart:
		var rules []validator.Rule
		if strict {
			rules = validator.StrictRules()
		} else {
			rules = validator.DefaultRules()
		}
		v := validator.New(rules...)
		return v.Validate(d)

	case *ast.SequenceDiagram:
		var rules []validator.SequenceRule
		if strict {
			rules = validator.SequenceStrictRules()
		} else {
			rules = validator.SequenceDefaultRules()
		}
		v := validator.NewSequence(rules...)
		return v.ValidateDiagram(diagram)

	case *ast.ClassDiagram:
		var rules []validator.ClassRule
		if strict {
			rules = validator.ClassStrictRules()
		} else {
			rules = validator.ClassDefaultRules()
		}
		v := validator.NewClass(rules...)
		return v.ValidateDiagram(diagram)

	case *ast.StateDiagram:
		var rules []validator.StateRule
		if strict {
			rules = validator.StateStrictRules()
		} else {
			rules = validator.StateDefaultRules()
		}
		v := validator.NewState(rules...)
		return v.ValidateDiagram(diagram)

	case *ast.PieDiagram:
		errors := validator.ValidatePie(d, strict)
		var validationErrors []validator.ValidationError
		for _, err := range errors {
			validationErrors = append(validationErrors, *err)
		}
		return validationErrors

	case *ast.ERDiagram:
		errors := validator.ValidateER(d, strict)
		var validationErrors []validator.ValidationError
		for _, err := range errors {
			validationErrors = append(validationErrors, *err)
		}
		return validationErrors

	case *ast.JourneyDiagram:
		errors := validator.ValidateJourney(d, strict)
		var validationErrors []validator.ValidationError
		for _, err := range errors {
			validationErrors = append(validationErrors, *err)
		}
		return validationErrors

	case *ast.TimelineDiagram:
		errors := validator.ValidateTimeline(d, strict)
		var validationErrors []validator.ValidationError
		for _, err := range errors {
			validationErrors = append(validationErrors, *err)
		}
		return validationErrors

	case *ast.GanttDiagram:
		errors := validator.ValidateGantt(d, strict)
		var validationErrors []validator.ValidationError
		for _, err := range errors {
			validationErrors = append(validationErrors, *err)
		}
		return validationErrors

	case *ast.GitGraphDiagram:
		errors := validator.ValidateGitGraph(d, strict)
		var validationErrors []validator.ValidationError
		for _, err := range errors {
			validationErrors = append(validationErrors, *err)
		}
		return validationErrors

	case *ast.MindmapDiagram:
		errors := validator.ValidateMindmap(d, strict)
		var validationErrors []validator.ValidationError
		for _, err := range errors {
			validationErrors = append(validationErrors, *err)
		}
		return validationErrors

	case *ast.SankeyDiagram:
		errors := validator.ValidateSankey(d, strict)
		var validationErrors []validator.ValidationError
		for _, err := range errors {
			validationErrors = append(validationErrors, *err)
		}
		return validationErrors

	case *ast.QuadrantDiagram:
		errors := validator.ValidateQuadrant(d, strict)
		var validationErrors []validator.ValidationError
		for _, err := range errors {
			validationErrors = append(validationErrors, *err)
		}
		return validationErrors

	case *ast.XYChartDiagram:
		errors := validator.ValidateXYChart(d, strict)
		var validationErrors []validator.ValidationError
		for _, err := range errors {
			validationErrors = append(validationErrors, *err)
		}
		return validationErrors

	case *ast.C4Diagram:
		var rules []validator.C4Rule
		if strict {
			rules = validator.StrictC4Rules()
		} else {
			rules = validator.DefaultC4Rules()
		}
		return validator.ValidateC4(d, rules)

	case *ast.GenericDiagram:
		var rules []validator.GenericRule
		if strict {
			rules = validator.GenericStrictRules()
		} else {
			rules = validator.GenericDefaultRules()
		}
		v := validator.NewGeneric(rules...)
		return v.ValidateDiagram(diagram)

	default:
		return []validator.ValidationError{{
			Line:     1,
			Column:   1,
			Message:  fmt.Sprintf("unsupported diagram type for validation: %T", diagram),
			Severity: validator.SeverityError,
		}}
	}
}

// ValidateFlowchart validates a flowchart diagram using the provided rules.
// If no rules are provided, uses default rules.
func ValidateFlowchart(diagram *ast.Flowchart, rules ...validator.Rule) []validator.ValidationError {
	if len(rules) == 0 {
		rules = validator.DefaultRules()
	}
	v := validator.New(rules...)
	return v.Validate(diagram)
}

// Exported validation rules for convenience
var (
	// NoParenthesesInLabels is a validation rule that checks node labels don't contain parentheses.
	NoParenthesesInLabels = &validator.NoParenthesesInLabels{}
	// ValidDirection checks if the flowchart direction is valid.
	ValidDirection = &validator.ValidDirection{}
	// NoUndefinedNodes checks that all referenced nodes are defined.
	NoUndefinedNodes = &validator.NoUndefinedNodes{}
	// NoDuplicateNodeIDs checks that node IDs are unique.
	NoDuplicateNodeIDs = &validator.NoDuplicateNodeIDs{}
)

// DefaultRules returns the default set of validation rules.
func DefaultRules() []validator.Rule {
	return validator.DefaultRules()
}

// StrictRules returns a strict set of validation rules including style checks.
func StrictRules() []validator.Rule {
	return validator.StrictRules()
}
