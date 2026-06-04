package ast

// QuadrantDiagram represents a quadrant chart diagram AST.
type QuadrantDiagram struct {
	Type           string          // Always "quadrantChart"
	Title          string          // Optional title
	XAxis          QuadrantAxis    // X-axis configuration
	YAxis          QuadrantAxis    // Y-axis configuration
	QuadrantLabels [4]string       // Labels for quadrants 1-4 (indexed 0-3)
	Points         []QuadrantPoint // Data points
	Source         string          // Original source
	Pos            Position        // Position in source
}

// QuadrantAxis represents an axis definition in a quadrant chart.
type QuadrantAxis struct {
	Min string // Left/bottom label
	Max string // Right/top label
}

// QuadrantPoint represents a data point in a quadrant chart.
type QuadrantPoint struct {
	Name string   // Point name
	X    float64  // X coordinate (0.0-1.0)
	Y    float64  // Y coordinate (0.0-1.0)
	Pos  Position // Position in source
}

// GetType returns the diagram type.
func (d *QuadrantDiagram) GetType() string {
	return d.Type
}

// GetSource returns the original source.
func (d *QuadrantDiagram) GetSource() string {
	return d.Source
}

// GetPosition returns the position in source.
func (d *QuadrantDiagram) GetPosition() Position {
	return d.Pos
}
