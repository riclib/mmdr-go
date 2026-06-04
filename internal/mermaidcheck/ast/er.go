package ast

// ERDiagram represents an Entity Relationship diagram AST.
type ERDiagram struct {
	Type          string           // Always "erDiagram"
	Direction     string           // Optional: TB, BT, LR, RL
	Entities      []EREntity       // Entity definitions
	Relationships []ERRelationship // Relationships between entities
	Source        string           // Original source
	Pos           Position         // Position in source
}

// EREntity represents an entity in an ER diagram.
type EREntity struct {
	Name       string        // Entity name
	Alias      string        // Optional alias
	Attributes []ERAttribute // Entity attributes
	Pos        Position      // Position in source
}

// ERAttribute represents an entity attribute.
type ERAttribute struct {
	Type    string   // Attribute type
	Name    string   // Attribute name
	Keys    []string // Key indicators: PK, FK, UK
	Comment string   // Optional comment
	Pos     Position // Position in source
}

// ERRelationship represents a relationship between entities.
type ERRelationship struct {
	From     string   // Source entity
	To       string   // Target entity
	FromCard string   // Source cardinality (||, |o, }|, }o)
	ToCard   string   // Target cardinality
	Type     string   // Identifying (--) or non-identifying (..)
	Label    string   // Optional relationship label
	Pos      Position // Position in source
}

// GetType returns the diagram type.
func (d *ERDiagram) GetType() string {
	return d.Type
}

// GetSource returns the original source.
func (d *ERDiagram) GetSource() string {
	return d.Source
}

// GetPosition returns the position in source.
func (d *ERDiagram) GetPosition() Position {
	return d.Pos
}
