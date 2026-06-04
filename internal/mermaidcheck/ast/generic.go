package ast

// GenericDiagram represents any Mermaid diagram type that doesn't have a specific parser yet.
// It stores the raw source and provides basic validation capabilities.
type GenericDiagram struct {
	DiagramType string   // The type of diagram (e.g., "sequence", "class", "gantt")
	Source      string   // Raw diagram source
	Lines       []string // Split lines for line-based validation
	Pos         Position // Position in source
}

// GetType returns the diagram type.
func (g *GenericDiagram) GetType() string { return g.DiagramType }

// GetPosition returns the position of this diagram in the source.
func (g *GenericDiagram) GetPosition() Position { return g.Pos }

// NewGenericDiagram creates a new generic diagram from source.
func NewGenericDiagram(diagramType, source string, pos Position) *GenericDiagram {
	lines := splitLines(source)
	return &GenericDiagram{
		DiagramType: diagramType,
		Source:      source,
		Lines:       lines,
		Pos:         pos,
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
