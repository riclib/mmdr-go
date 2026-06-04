package ast

// C4Diagram represents any C4 diagram (Context, Container, Component, Dynamic, Deployment).
// All C4 diagram types share the same AST structure with common elements.
type C4Diagram struct {
	DiagramType   string           // "c4Context", "c4Container", "c4Component", "c4Dynamic", "c4Deployment"
	Title         string           // Optional title
	Elements      []C4Element      // All elements (Person, System, Container, Component, Node)
	Boundaries    []C4Boundary     // Boundary elements (can be nested)
	Relationships []C4Relationship // All relationships (Rel, BiRel, etc.)
	Styles        []C4Style        // Style overrides
	Source        string           // Original source
	Pos           Position         // Position in source
}

// GetType implements the Diagram interface.
func (c *C4Diagram) GetType() string {
	return c.DiagramType
}

// GetSource implements the Diagram interface.
func (c *C4Diagram) GetSource() string {
	return c.Source
}

// GetPosition implements the Diagram interface.
func (c *C4Diagram) GetPosition() Position {
	return c.Pos
}

// C4Element represents a C4 diagram element (Person, System, Container, Component, Node).
type C4Element struct {
	ElementType string   // "Person", "System", "Container", "Component", "Deployment_Node", "Node"
	ID          string   // Element identifier
	Label       string   // Display label (required)
	Technology  string   // Optional technology (for Container/Component)
	Description string   // Optional description
	Sprite      string   // Optional sprite
	Tags        string   // Optional tags
	Link        string   // Optional link
	External    bool     // True for _Ext variants (Person_Ext, System_Ext)
	Database    bool     // True for Db variants (ContainerDb, ComponentDb)
	Queue       bool     // True for Queue variants (ContainerQueue, ComponentQueue)
	Pos         Position // Position in source
}

// C4Boundary represents a boundary element that can contain other elements.
type C4Boundary struct {
	BoundaryType string       // "Boundary", "Enterprise_Boundary", "System_Boundary", "Container_Boundary"
	ID           string       // Boundary identifier
	Label        string       // Display label
	Type         string       // Optional type (for generic Boundary)
	Elements     []C4Element  // Nested elements
	Boundaries   []C4Boundary // Nested boundaries
	Pos          Position     // Position in source
}

// C4Relationship represents a relationship between elements.
type C4Relationship struct {
	RelType     string   // "Rel", "Rel_Back", "Rel_Neighbor", "Rel_Down", "Rel_Up", "Rel_Left", "Rel_Right", "BiRel"
	From        string   // Source element ID
	To          string   // Target element ID
	Label       string   // Relationship label
	Technology  string   // Optional technology
	Description string   // Optional description
	Sprite      string   // Optional sprite
	Tags        string   // Optional tags
	Link        string   // Optional link
	Pos         Position // Position in source
}

// C4Style represents a style override for elements or relationships.
type C4Style struct {
	StyleType   string   // "UpdateElementStyle" or "UpdateRelStyle"
	ElementID   string   // For element styles (UpdateElementStyle)
	From        string   // For relationship styles (UpdateRelStyle)
	To          string   // For relationship styles (UpdateRelStyle)
	BgColor     string   // Background colour
	FontColor   string   // Font colour
	BorderColor string   // Border colour
	Shadowing   string   // Shadowing setting
	Shape       string   // Shape override
	TextColor   string   // Text colour (for relationships)
	LineColor   string   // Line colour (for relationships)
	OffsetX     string   // X offset (for relationships)
	OffsetY     string   // Y offset (for relationships)
	Pos         Position // Position in source
}
