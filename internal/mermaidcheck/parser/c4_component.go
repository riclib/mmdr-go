package parser

import "github.com/riclib/mmdr-go/internal/mermaidcheck/ast"

// C4ComponentParser parses C4 Component diagrams.
type C4ComponentParser struct{}

// NewC4ComponentParser creates a new C4 Component parser.
func NewC4ComponentParser() *C4ComponentParser {
	return &C4ComponentParser{}
}

// Parse parses a C4 Component diagram and returns a C4Diagram AST.
func (p *C4ComponentParser) Parse(source string) (ast.Diagram, error) {
	return parseC4Diagram(source, "c4Component", "C4Component")
}

// SupportedTypes returns the diagram types this parser supports.
func (p *C4ComponentParser) SupportedTypes() []string {
	return []string{"c4Component"}
}
