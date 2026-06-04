package ast

// PieDiagram represents a pie chart diagram AST.
type PieDiagram struct {
	Type        string     // Always "pie"
	Title       string     // Optional title
	ShowData    bool       // Whether to show data values
	DataEntries []PieEntry // Data entries
	Source      string     // Original source
	Pos         Position   // Position in source
}

// PieEntry represents a single data entry in a pie chart.
type PieEntry struct {
	Label string   // Entry label (must be quoted in source)
	Value float64  // Numeric value (must be positive)
	Pos   Position // Position in source
}

// GetType returns the diagram type.
func (d *PieDiagram) GetType() string {
	return d.Type
}

// GetSource returns the original source.
func (d *PieDiagram) GetSource() string {
	return d.Source
}

// GetPosition returns the position in source.
func (d *PieDiagram) GetPosition() Position {
	return d.Pos
}
