package ast

// GitGraphDiagram represents a git graph diagram AST.
type GitGraphDiagram struct {
	Type            string         // Always "gitGraph"
	Theme           string         // Optional theme
	MainBranchName  string         // Optional main branch name (default "main")
	MainBranchOrder int            // Optional main branch order
	Operations      []GitOperation // All git operations (commits, branches, merges, etc.)
	Source          string         // Original source
	Pos             Position       // Position in source
}

// GitOperation represents a single git operation.
type GitOperation struct {
	Type       string   // "commit", "branch", "checkout", "merge", "cherry-pick"
	ID         string   // Commit ID or branch name
	Tag        string   // Optional tag name
	CommitType string   // NORMAL, REVERSE, HIGHLIGHT (for commits)
	BranchName string   // Branch name (for branch, checkout, merge operations)
	Order      int      // Branch order (for branch operation)
	ParentID   string   // Parent commit ID (for cherry-pick)
	Pos        Position // Position in source
}

// GetType returns the diagram type.
func (d *GitGraphDiagram) GetType() string {
	return d.Type
}

// GetSource returns the original source.
func (d *GitGraphDiagram) GetSource() string {
	return d.Source
}

// GetPosition returns the position in source.
func (d *GitGraphDiagram) GetPosition() Position {
	return d.Pos
}
