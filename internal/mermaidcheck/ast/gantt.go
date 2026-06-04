package ast

// GanttDiagram represents a Gantt chart diagram AST.
type GanttDiagram struct {
	Type        string         // Always "gantt"
	Title       string         // Optional title
	DateFormat  string         // Date format (default YYYY-MM-DD)
	AxisFormat  string         // Optional axis format for display
	Excludes    string         // Excluded days (weekends, holidays, etc.)
	TodayMarker string         // "on", "off", or colour value
	Sections    []GanttSection // Sections with tasks
	Source      string         // Original source
	Pos         Position       // Position in source
}

// GanttSection represents a section within a Gantt chart.
type GanttSection struct {
	Name  string      // Section name
	Tasks []GanttTask // Tasks within this section
	Pos   Position    // Position in source
}

// GanttTask represents a task within a Gantt section.
type GanttTask struct {
	Name         string   // Task name/description
	ID           string   // Optional task ID
	Status       string   // done, active, crit, milestone, etc.
	Dependencies []string // Task IDs this depends on (after syntax)
	StartDate    string   // Start date or dependency reference
	EndDate      string   // End date or duration (e.g., "10d")
	Pos          Position // Position in source
}

// GetType returns the diagram type.
func (d *GanttDiagram) GetType() string {
	return d.Type
}

// GetSource returns the original source.
func (d *GanttDiagram) GetSource() string {
	return d.Source
}

// GetPosition returns the position in source.
func (d *GanttDiagram) GetPosition() Position {
	return d.Pos
}
