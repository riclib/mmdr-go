package ast

// XYChartDiagram represents an XY chart diagram AST.
type XYChartDiagram struct {
	Type        string          // Always "xyChart"
	Orientation string          // "horizontal" or "vertical" (default "vertical")
	Title       string          // Optional title
	XAxis       XYChartAxis     // X-axis configuration
	YAxis       XYChartAxis     // Y-axis configuration
	Series      []XYChartSeries // Data series (bar, line)
	Source      string          // Original source
	Pos         Position        // Position in source
}

// XYChartAxis represents an axis configuration in an XY chart.
type XYChartAxis struct {
	Label      string   // Axis label (optional)
	Categories []string // Category labels (for categorical axis)
	Min        float64  // Minimum value (for numeric axis)
	Max        float64  // Maximum value (for numeric axis)
	IsNumeric  bool     // True if numeric, false if categorical
	Pos        Position // Position in source
}

// XYChartSeries represents a data series in an XY chart.
type XYChartSeries struct {
	Type   string    // "bar" or "line"
	Values []float64 // Data values
	Pos    Position  // Position in source
}

// GetType returns the diagram type.
func (d *XYChartDiagram) GetType() string {
	return d.Type
}

// GetSource returns the original source.
func (d *XYChartDiagram) GetSource() string {
	return d.Source
}

// GetPosition returns the position in source.
func (d *XYChartDiagram) GetPosition() Position {
	return d.Pos
}
