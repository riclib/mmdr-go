package ast

// TimelineDiagram represents a timeline diagram AST.
type TimelineDiagram struct {
	Type     string            // Always "timeline"
	Title    string            // Optional title
	Sections []TimelineSection // Sections containing periods
	Source   string            // Original source
	Pos      Position          // Position in source
}

// TimelineSection represents a section in a timeline diagram.
type TimelineSection struct {
	Name    string           // Section name (empty for default section)
	Periods []TimelinePeriod // Time periods in this section
	Pos     Position         // Position in source
}

// TimelinePeriod represents a time period with associated events.
type TimelinePeriod struct {
	TimePeriod string   // Time period text (e.g., "1940s", "Early Stage")
	Events     []string // Events that occurred in this period
	Pos        Position // Position in source
}

// GetType returns the diagram type.
func (d *TimelineDiagram) GetType() string {
	return d.Type
}

// GetSource returns the original source.
func (d *TimelineDiagram) GetSource() string {
	return d.Source
}

// GetPosition returns the position in source.
func (d *TimelineDiagram) GetPosition() Position {
	return d.Pos
}
