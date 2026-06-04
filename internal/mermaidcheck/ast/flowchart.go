// Package ast defines the Abstract Syntax Tree types for Mermaid diagrams.
package ast

// Flowchart represents a complete Mermaid flowchart or graph diagram.
type Flowchart struct {
	Type       string      // "flowchart" or "graph"
	Direction  string      // TB, TD, BT, RL, LR
	Statements []Statement // All statements in the diagram
	Source     string      // Original source
	Pos        Position    // Position in source
}

// GetType returns the diagram type.
func (f *Flowchart) GetType() string { return f.Type }

// GetPosition returns the position of this diagram in the source.
func (f *Flowchart) GetPosition() Position { return f.Pos }

// GetSource returns the original source.
func (f *Flowchart) GetSource() string { return f.Source }

// Statement represents any statement in the flowchart.
type Statement interface {
	statement()
	GetPosition() Position
}

// NodeDef represents a node definition.
type NodeDef struct {
	ID    string // Node identifier
	Shape string // Shape type (bracket style)
	Label string // Node label/text
	Pos   Position
}

func (n *NodeDef) statement() {}

// GetPosition returns the position of this node definition in the source.
func (n *NodeDef) GetPosition() Position { return n.Pos }

// Link represents a link between nodes.
type Link struct {
	From  string // Source node ID
	To    string // Target node ID
	Arrow string // Arrow type (-->, -.>, ==>, etc.)
	Label string // Link label (optional)
	BiDir bool   // Bidirectional arrow
	Pos   Position
}

func (l *Link) statement() {}

// GetPosition returns the position of this link in the source.
func (l *Link) GetPosition() Position { return l.Pos }

// Subgraph represents a subgraph block.
type Subgraph struct {
	Title      string      // Subgraph title
	Statements []Statement // Nested statements
	Pos        Position
}

func (s *Subgraph) statement() {}

// GetPosition returns the position of this subgraph in the source.
func (s *Subgraph) GetPosition() Position { return s.Pos }

// ClassDef represents a class definition for styling.
type ClassDef struct {
	Name   string            // Class name
	Styles map[string]string // CSS properties
	Pos    Position
}

func (c *ClassDef) statement() {}

// GetPosition returns the position of this class definition in the source.
func (c *ClassDef) GetPosition() Position { return c.Pos }

// ClassAssignment represents assigning classes to nodes.
type ClassAssignment struct {
	NodeIDs   []string // Node IDs to apply class to
	ClassName string   // Class name to apply
	Pos       Position
}

func (c *ClassAssignment) statement() {}

// GetPosition returns the position of this class assignment in the source.
func (c *ClassAssignment) GetPosition() Position { return c.Pos }

// Comment represents a comment line.
type Comment struct {
	Text string
	Pos  Position
}

func (c *Comment) statement() {}

// GetPosition returns the position of this comment in the source.
func (c *Comment) GetPosition() Position { return c.Pos }
