// Package ast defines the Abstract Syntax Tree types for all Mermaid diagram types.
package ast

// Diagram is the base interface that all diagram types implement.
type Diagram interface {
	// GetType returns the diagram type (e.g., "flowchart", "sequence", "class").
	GetType() string
	// GetPosition returns the position in the source where this diagram starts.
	GetPosition() Position
}

// Position represents a location in the source text.
type Position struct {
	Line   int // Line number (1-indexed)
	Column int // Column number (1-indexed)
}
