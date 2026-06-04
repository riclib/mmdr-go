package ast

// JourneyDiagram represents a user journey diagram AST.
type JourneyDiagram struct {
	Type     string    // Always "journey"
	Title    string    // Optional title
	Sections []Section // Journey sections
	Source   string    // Original source
	Pos      Position  // Position in source
}

// Section represents a section within a user journey diagram.
type Section struct {
	Name  string   // Section name
	Tasks []Task   // Tasks within this section
	Pos   Position // Position in source
}

// Task represents a task within a journey section.
type Task struct {
	Name   string   // Task description
	Score  int      // Task score (1-5)
	Actors []string // Actors involved in the task
	Pos    Position // Position in source
}

// GetType returns the diagram type.
func (d *JourneyDiagram) GetType() string {
	return d.Type
}

// GetSource returns the original source.
func (d *JourneyDiagram) GetSource() string {
	return d.Source
}

// GetPosition returns the position in source.
func (d *JourneyDiagram) GetPosition() Position {
	return d.Pos
}
