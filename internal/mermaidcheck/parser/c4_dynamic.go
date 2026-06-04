package parser

import "github.com/riclib/mmdr-go/internal/mermaidcheck/ast"

// C4DynamicParser parses C4 Dynamic diagrams.
type C4DynamicParser struct{}

// NewC4DynamicParser creates a new C4 Dynamic parser.
func NewC4DynamicParser() *C4DynamicParser {
	return &C4DynamicParser{}
}

// Parse parses a C4 Dynamic diagram and returns a C4Diagram AST.
func (p *C4DynamicParser) Parse(source string) (ast.Diagram, error) {
	return parseC4Diagram(source, "c4Dynamic", "C4Dynamic")
}

// SupportedTypes returns the diagram types this parser supports.
func (p *C4DynamicParser) SupportedTypes() []string {
	return []string{"c4Dynamic"}
}
