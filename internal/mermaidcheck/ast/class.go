package ast

// ClassDiagram represents a Mermaid class diagram.
type ClassDiagram struct {
	Type       string      // "class"
	Statements []ClassStmt // All statements in the diagram
	Source     string      // Original source
	Pos        Position    // Position in source
}

// ClassStmt is the interface for all class diagram statements.
type ClassStmt interface {
	classStmt()
	GetPosition() Position
}

// Class represents a class definition.
type Class struct {
	Name        string        // Class name
	Stereotype  string        // Optional stereotype (e.g., "interface", "abstract")
	Members     []ClassMember // Class members (attributes and methods)
	Annotations []string      // Annotations like <<interface>>
	Pos         Position
}

func (c *Class) classStmt() {}

// GetPosition returns the position in source.
func (c *Class) GetPosition() Position { return c.Pos }

// ClassMember represents an attribute or method in a class.
type ClassMember struct {
	Visibility string   // +, -, #, ~ (public, private, protected, package)
	Name       string   // Member name
	Type       string   // Type for attributes, return type for methods
	IsMethod   bool     // true if method, false if attribute
	Parameters []string // Method parameters (if IsMethod)
	IsStatic   bool     // Class-level member
	IsAbstract bool     // Abstract method
	Pos        Position
}

// Relationship represents a relationship between classes.
type Relationship struct {
	From             string // Source class name
	To               string // Target class name
	Type             string // Relationship type (inheritance, composition, aggregation, association, dependency, realization)
	Label            string // Optional relationship label
	FromMultiplicity string // Multiplicity on source end
	ToMultiplicity   string // Multiplicity on target end
	FromCardinality  string // Cardinality on source end (alternative to multiplicity)
	ToCardinality    string // Cardinality on target end
	Pos              Position
}

func (r *Relationship) classStmt() {}

// GetPosition returns the position in source.
func (r *Relationship) GetPosition() Position { return r.Pos }

// ClassNote represents a note attached to a class.
type ClassNote struct {
	ClassName string // Class the note is attached to
	Text      string // Note text
	Pos       Position
}

func (n *ClassNote) classStmt() {}

// GetPosition returns the position in source.
func (n *ClassNote) GetPosition() Position { return n.Pos }

// ClassComment represents a comment in the class diagram.
type ClassComment struct {
	Text string // Comment text (without %%)
	Pos  Position
}

func (c *ClassComment) classStmt() {}

// GetPosition returns the position in source.
func (c *ClassComment) GetPosition() Position { return c.Pos }

// GetType returns the diagram type.
func (d *ClassDiagram) GetType() string {
	return d.Type
}

// GetPosition returns the position in source.
func (d *ClassDiagram) GetPosition() Position {
	return d.Pos
}

// GetSource returns the original source.
func (d *ClassDiagram) GetSource() string {
	return d.Source
}
