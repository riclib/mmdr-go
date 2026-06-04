package validator

import (
	"fmt"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// GenericRule defines validation rules that work across all diagram types.
type GenericRule interface {
	// Name returns the name of the rule.
	Name() string
	// Validate checks a generic diagram and returns any validation errors.
	ValidateGeneric(diagram *ast.GenericDiagram) []ValidationError
}

// ValidComments checks that all comments use proper %% syntax.
type ValidComments struct{}

// Name returns the name of this validation rule.
func (r *ValidComments) Name() string { return "valid-comments" }

// ValidateGeneric checks comment syntax.
func (r *ValidComments) ValidateGeneric(diagram *ast.GenericDiagram) []ValidationError {
	var errors []ValidationError
	for i, line := range diagram.Lines {
		trimmed := strings.TrimSpace(line)
		// Check for invalid comment syntax (single % instead of %%)
		if strings.HasPrefix(trimmed, "%") && !strings.HasPrefix(trimmed, "%%") {
			errors = append(errors, ValidationError{
				Line:     diagram.Pos.Line + i,
				Column:   strings.Index(line, "%") + 1,
				Message:  "invalid comment syntax: use '%%' for comments, not '%'",
				Severity: SeverityError,
			})
		}
	}
	return errors
}

// NoTrailingWhitespace checks for trailing whitespace on lines.
type NoTrailingWhitespace struct{}

// Name returns the name of this validation rule.
func (r *NoTrailingWhitespace) Name() string { return "no-trailing-whitespace" }

// ValidateGeneric checks for trailing whitespace.
func (r *NoTrailingWhitespace) ValidateGeneric(diagram *ast.GenericDiagram) []ValidationError {
	var errors []ValidationError
	for i, line := range diagram.Lines {
		if len(line) > 0 && (line[len(line)-1] == ' ' || line[len(line)-1] == '\t') {
			errors = append(errors, ValidationError{
				Line:     diagram.Pos.Line + i,
				Column:   len(line),
				Message:  "trailing whitespace on line",
				Severity: SeverityWarning,
			})
		}
	}
	return errors
}

// NoParenthesesInText is a generic version that works on any diagram type.
type NoParenthesesInText struct{}

// Name returns the name of this validation rule.
func (r *NoParenthesesInText) Name() string { return "no-parentheses-in-text" }

// ValidateGeneric checks for parentheses in text content.
func (r *NoParenthesesInText) ValidateGeneric(diagram *ast.GenericDiagram) []ValidationError {
	var errors []ValidationError
	for i, line := range diagram.Lines {
		trimmed := strings.TrimSpace(line)
		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		// Skip the first line (diagram type declaration)
		if i == 0 {
			continue
		}

		// Check for parentheses in text (excluding function-like syntax)
		if strings.Contains(line, "(") || strings.Contains(line, ")") {
			// Allow certain patterns like arrows `-->` or technical notation
			if !isAllowedParenthesesContext(line) {
				errors = append(errors, ValidationError{
					Line:     diagram.Pos.Line + i,
					Column:   1,
					Message:  "text contains parentheses; use <br/> for line breaks instead",
					Severity: SeverityWarning,
				})
			}
		}
	}
	return errors
}

func isAllowedParenthesesContext(line string) bool {
	// Allow parentheses in certain contexts like method signatures in class diagrams
	trimmed := strings.TrimSpace(line)
	// Class diagram method signatures (must start with +, -, or #)
	if len(trimmed) > 0 && (trimmed[0] == '+' || trimmed[0] == '-' || trimmed[0] == '#') {
		return true
	}
	// ER diagram notation
	if strings.Contains(trimmed, "||") || strings.Contains(trimmed, "}|") || strings.Contains(trimmed, "|{") {
		return true
	}
	return false
}

// ValidDiagramHeader checks that the diagram starts with a valid header.
type ValidDiagramHeader struct{}

// Name returns the name of this validation rule.
func (r *ValidDiagramHeader) Name() string { return "valid-diagram-header" }

// ValidateGeneric checks the diagram header.
func (r *ValidDiagramHeader) ValidateGeneric(diagram *ast.GenericDiagram) []ValidationError {
	if len(diagram.Lines) == 0 {
		return []ValidationError{{
			Line:     diagram.Pos.Line,
			Column:   1,
			Message:  "diagram is empty",
			Severity: SeverityError,
		}}
	}

	firstLine := ""
	for _, line := range diagram.Lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "%%") {
			firstLine = trimmed
			break
		}
	}

	if firstLine == "" {
		return []ValidationError{{
			Line:     diagram.Pos.Line,
			Column:   1,
			Message:  "diagram has no content (only comments or whitespace)",
			Severity: SeverityError,
		}}
	}

	// Verify the header matches the diagram type
	expectedPrefixes := getValidHeaderPrefixes(diagram.DiagramType)
	hasValidPrefix := false
	for _, prefix := range expectedPrefixes {
		if strings.HasPrefix(firstLine, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return []ValidationError{{
			Line:     diagram.Pos.Line,
			Column:   1,
			Message:  fmt.Sprintf("invalid diagram header for type '%s'", diagram.DiagramType),
			Severity: SeverityError,
		}}
	}

	return nil
}

func getValidHeaderPrefixes(diagramType string) []string {
	switch diagramType {
	case "flowchart":
		return []string{"flowchart"}
	case "graph":
		return []string{"graph"}
	case "sequence":
		return []string{"sequenceDiagram"}
	case "class":
		return []string{"classDiagram"}
	case "state":
		return []string{"stateDiagram"}
	case "stateDiagram-v2":
		return []string{"stateDiagram-v2"}
	case "er":
		return []string{"erDiagram"}
	case "gantt":
		return []string{"gantt"}
	case "pie":
		return []string{"pie"}
	case "journey":
		return []string{"journey"}
	case "gitGraph":
		return []string{"gitGraph"}
	case "mindmap":
		return []string{"mindmap"}
	case "timeline":
		return []string{"timeline"}
	case "sankey":
		return []string{"sankey-beta"}
	case "quadrantChart":
		return []string{"quadrantChart"}
	case "xyChart":
		return []string{"xychart-beta"}
	case "c4Context":
		return []string{"C4Context"}
	case "c4Container":
		return []string{"C4Container"}
	case "c4Component":
		return []string{"C4Component"}
	case "c4Dynamic":
		return []string{"C4Dynamic"}
	case "c4Deployment":
		return []string{"C4Deployment"}
	default:
		return []string{}
	}
}

// GenericDefaultRules returns default validation rules for generic diagrams.
func GenericDefaultRules() []GenericRule {
	return []GenericRule{
		&ValidDiagramHeader{},
		&ValidComments{},
	}
}

// GenericStrictRules returns strict validation rules for generic diagrams.
func GenericStrictRules() []GenericRule {
	return []GenericRule{
		&ValidDiagramHeader{},
		&ValidComments{},
		&NoParenthesesInText{},
		&NoTrailingWhitespace{},
	}
}
