package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

var (
	// Class diagram patterns
	classHeaderPattern  = regexp.MustCompile(`^classDiagram\s*$`)
	classCommentPattern = regexp.MustCompile(`^%%(.*)$`)

	// Class declaration patterns
	classDeclPattern      = regexp.MustCompile(`^class\s+(\w+)(?:\s*<<(.+)>>)?\s*$`)
	classBodyStartPattern = regexp.MustCompile(`^class\s+(\w+)(?:\s*<<(.+)>>)?\s*\{\s*$`)
	classBodyEndPattern   = regexp.MustCompile(`^\}\s*$`)

	// Member patterns
	memberPattern = regexp.MustCompile(`^([+\-#~])(\w+)(?:\(([^)]*)\))?(?:\s+(.+))?\s*$`)

	// Relationship patterns
	// Inheritance: --|>, <|--
	// Composition: --*, *--
	// Aggregation: --o, o--
	// Association: --, -->
	// Dependency: .., ..>, <..
	// Realization: ..|>, <|..
	relationshipPattern = regexp.MustCompile(`^(\w+)\s+(?:"([^"]+)"\s+)?([<*o])?(-{2}|\.{2})([>|*o]?)\s+(?:"([^"]+)"\s+)?(\w+)(?:\s*:\s*(.+))?\s*$`)

	// Note pattern
	classNotePattern = regexp.MustCompile(`^note\s+for\s+(\w+)\s+"([^"]+)"\s*$`)
)

// ClassParser parses Mermaid class diagrams.
type ClassParser struct{}

// NewClassParser creates a new class diagram parser.
func NewClassParser() *ClassParser {
	return &ClassParser{}
}

// SupportedTypes returns the diagram types this parser handles.
func (p *ClassParser) SupportedTypes() []string {
	return []string{"class"}
}

// Parse parses a Mermaid class diagram from a string.
func (p *ClassParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram")
	}

	// Parse header
	header := strings.TrimSpace(lines[0])
	if !classHeaderPattern.MatchString(header) {
		return nil, fmt.Errorf("invalid class diagram header: expected 'classDiagram'")
	}

	diagram := &ast.ClassDiagram{
		Type:   "class",
		Source: source,
		Pos:    ast.Position{Line: 1, Column: 1},
	}

	// Parse statements
	statements, err := p.parseStatements(lines[1:], 1)
	if err != nil {
		return nil, err
	}
	diagram.Statements = statements

	return diagram, nil
}

func (p *ClassParser) parseStatements(lines []string, startLine int) ([]ast.ClassStmt, error) {
	var statements []ast.ClassStmt
	lineNum := startLine

	for i := 0; i < len(lines); i++ {
		lineNum++
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Handle comments
		if matches := classCommentPattern.FindStringSubmatch(trimmed); matches != nil {
			statements = append(statements, &ast.ClassComment{
				Text: strings.TrimSpace(matches[1]),
				Pos:  ast.Position{Line: lineNum, Column: 1},
			})
			continue
		}

		// Handle class with body
		if matches := classBodyStartPattern.FindStringSubmatch(trimmed); matches != nil {
			className := matches[1]
			stereotype := ""
			if len(matches) > 2 && matches[2] != "" {
				stereotype = matches[2]
			}

			// Find closing brace
			members, consumed, err := p.parseClassBody(lines[i+1:], lineNum+1)
			if err != nil {
				return nil, err
			}

			class := &ast.Class{
				Name:       className,
				Stereotype: stereotype,
				Members:    members,
				Pos:        ast.Position{Line: lineNum, Column: 1},
			}
			statements = append(statements, class)

			i += consumed
			lineNum += consumed
			continue
		}

		// Handle simple class declaration
		if matches := classDeclPattern.FindStringSubmatch(trimmed); matches != nil {
			className := matches[1]
			stereotype := ""
			if len(matches) > 2 && matches[2] != "" {
				stereotype = matches[2]
			}

			class := &ast.Class{
				Name:       className,
				Stereotype: stereotype,
				Members:    []ast.ClassMember{},
				Pos:        ast.Position{Line: lineNum, Column: 1},
			}
			statements = append(statements, class)
			continue
		}

		// Handle relationships
		if matches := relationshipPattern.FindStringSubmatch(trimmed); matches != nil {
			from := matches[1]
			fromCard := matches[2]
			leftSymbol := matches[3]
			linkType := matches[4]
			rightSymbol := matches[5]
			toCard := matches[6]
			to := matches[7]
			label := ""
			if len(matches) > 8 {
				label = matches[8]
			}

			relType := p.determineRelationshipType(leftSymbol, linkType, rightSymbol)

			relationship := &ast.Relationship{
				From:            from,
				To:              to,
				Type:            relType,
				Label:           label,
				FromCardinality: fromCard,
				ToCardinality:   toCard,
				Pos:             ast.Position{Line: lineNum, Column: 1},
			}
			statements = append(statements, relationship)
			continue
		}

		// Handle notes
		if matches := classNotePattern.FindStringSubmatch(trimmed); matches != nil {
			note := &ast.ClassNote{
				ClassName: matches[1],
				Text:      matches[2],
				Pos:       ast.Position{Line: lineNum, Column: 1},
			}
			statements = append(statements, note)
			continue
		}

		// Skip lines we can't parse (for now)
		continue
	}

	return statements, nil
}

func (p *ClassParser) parseClassBody(lines []string, startLine int) ([]ast.ClassMember, int, error) {
	var members []ast.ClassMember
	lineNum := startLine

	for i, line := range lines {
		lineNum++
		trimmed := strings.TrimSpace(line)

		// Check for end of class body
		if classBodyEndPattern.MatchString(trimmed) {
			return members, i + 1, nil
		}

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Parse member
		if matches := memberPattern.FindStringSubmatch(trimmed); matches != nil {
			visibility := matches[1]
			name := matches[2]
			params := matches[3]
			typ := ""
			if len(matches) > 4 {
				typ = matches[4]
			}

			member := ast.ClassMember{
				Visibility: visibility,
				Name:       name,
				Type:       typ,
				IsMethod:   params != "",
				Pos:        ast.Position{Line: lineNum, Column: 1},
			}

			if params != "" {
				// Parse parameters
				if params != "" {
					paramList := strings.Split(params, ",")
					for i := range paramList {
						paramList[i] = strings.TrimSpace(paramList[i])
					}
					member.Parameters = paramList
				}
			}

			members = append(members, member)
		}
	}

	return nil, 0, fmt.Errorf("line %d: unclosed class body", startLine)
}

func (p *ClassParser) determineRelationshipType(left, link, right string) string {
	// Inheritance: --|>, <|--
	if link == "--" && (right == "|>" || left == "<|") {
		return "inheritance"
	}

	// Realization: ..|>, <|..
	if link == ".." && (right == "|>" || left == "<|") {
		return "realization"
	}

	// Composition: --*, *--
	if link == "--" && (right == "*" || left == "*") {
		return "composition"
	}

	// Aggregation: --o, o--
	if link == "--" && (right == "o" || left == "o") {
		return "aggregation"
	}

	// Dependency: .., ..>, <..
	if link == ".." {
		return "dependency"
	}

	// Association: --, -->
	return "association"
}
