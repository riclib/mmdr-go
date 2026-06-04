package ast

// SankeyDiagram represents a Sankey diagram AST.
type SankeyDiagram struct {
	Type   string       // Always "sankey"
	Links  []SankeyLink // Flow links between nodes
	Source string       // Original source
	Pos    Position     // Position in source
}

// SankeyLink represents a flow link between two nodes.
type SankeyLink struct {
	Source string   // Source node name
	Target string   // Target node name
	Value  float64  // Flow value (must be positive)
	Pos    Position // Position in source
}

// GetType returns the diagram type.
func (d *SankeyDiagram) GetType() string {
	return d.Type
}

// GetSource returns the original source.
func (d *SankeyDiagram) GetSource() string {
	return d.Source
}

// GetPosition returns the position in source.
func (d *SankeyDiagram) GetPosition() Position {
	return d.Pos
}
