package ast

// MindmapDiagram represents a mindmap diagram AST.
type MindmapDiagram struct {
	Type   string       // Always "mindmap"
	Root   *MindmapNode // Root node (required)
	Source string       // Original source
	Pos    Position     // Position in source
}

// MindmapNode represents a node in a mindmap diagram.
type MindmapNode struct {
	Text     string         // Node text content
	Shape    string         // Node shape: "()", "(())", "[]", "{{}}", "))((" or "" for default
	Icon     string         // Optional icon (e.g., "fa fa-book")
	Level    int            // Indentation level (0 for root)
	Children []*MindmapNode // Child nodes
	Pos      Position       // Position in source
}

// GetType returns the diagram type.
func (d *MindmapDiagram) GetType() string {
	return d.Type
}

// GetSource returns the original source.
func (d *MindmapDiagram) GetSource() string {
	return d.Source
}

// GetPosition returns the position in source.
func (d *MindmapDiagram) GetPosition() Position {
	return d.Pos
}
