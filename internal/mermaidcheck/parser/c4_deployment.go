package parser

import "github.com/riclib/mmdr-go/internal/mermaidcheck/ast"

// C4DeploymentParser parses C4 Deployment diagrams.
type C4DeploymentParser struct{}

// NewC4DeploymentParser creates a new C4 Deployment parser.
func NewC4DeploymentParser() *C4DeploymentParser {
	return &C4DeploymentParser{}
}

// Parse parses a C4 Deployment diagram and returns a C4Diagram AST.
func (p *C4DeploymentParser) Parse(source string) (ast.Diagram, error) {
	return parseC4Diagram(source, "c4Deployment", "C4Deployment")
}

// SupportedTypes returns the diagram types this parser supports.
func (p *C4DeploymentParser) SupportedTypes() []string {
	return []string{"c4Deployment"}
}
