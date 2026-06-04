package parser

import "github.com/riclib/mmdr-go/internal/mermaidcheck/ast"

// C4ContainerParser parses C4 Container diagrams.
type C4ContainerParser struct{}

// NewC4ContainerParser creates a new C4 Container parser.
func NewC4ContainerParser() *C4ContainerParser {
	return &C4ContainerParser{}
}

// Parse parses a C4 Container diagram and returns a C4Diagram AST.
func (p *C4ContainerParser) Parse(source string) (ast.Diagram, error) {
	return parseC4Diagram(source, "c4Container", "C4Container")
}

// SupportedTypes returns the diagram types this parser supports.
func (p *C4ContainerParser) SupportedTypes() []string {
	return []string{"c4Container"}
}
