package ast

// StateDiagram represents a Mermaid state diagram (stateDiagram or stateDiagram-v2).
type StateDiagram struct {
	Type       string      // "state" or "stateDiagram-v2"
	Statements []StateStmt // All statements in the diagram
	Source     string      // Original source
	Pos        Position    // Position in source
}

// StateStmt is the interface for all state diagram statements.
type StateStmt interface {
	stateStmt()
	GetPosition() Position
}

// State represents a state in the diagram.
type State struct {
	ID          string      // State ID
	Description string      // State description/label
	IsComposite bool        // true if this state contains nested states
	Nested      []StateStmt // Nested statements (for composite states)
	Pos         Position
}

func (s *State) stateStmt() {}

// GetPosition returns the position in source.
func (s *State) GetPosition() Position { return s.Pos }

// Transition represents a transition between states.
type Transition struct {
	From  string // Source state ID
	To    string // Target state ID
	Label string // Transition label/condition
	Pos   Position
}

func (t *Transition) stateStmt() {}

// GetPosition returns the position in source.
func (t *Transition) GetPosition() Position { return t.Pos }

// StartState represents the start state [*].
type StartState struct {
	To  string // Target state after start
	Pos Position
}

func (s *StartState) stateStmt() {}

// GetPosition returns the position in source.
func (s *StartState) GetPosition() Position { return s.Pos }

// EndState represents the end state transition.
type EndState struct {
	From string // State transitioning to end
	Pos  Position
}

func (e *EndState) stateStmt() {}

// GetPosition returns the position in source.
func (e *EndState) GetPosition() Position { return e.Pos }

// Fork represents a fork node for concurrent states.
type Fork struct {
	ID  string // Fork ID
	Pos Position
}

func (f *Fork) stateStmt() {}

// GetPosition returns the position in source.
func (f *Fork) GetPosition() Position { return f.Pos }

// Join represents a join node for concurrent states.
type Join struct {
	ID  string // Join ID
	Pos Position
}

func (j *Join) stateStmt() {}

// GetPosition returns the position in source.
func (j *Join) GetPosition() Position { return j.Pos }

// Choice represents a choice node (conditional).
type Choice struct {
	ID  string // Choice ID
	Pos Position
}

func (c *Choice) stateStmt() {}

// GetPosition returns the position in source.
func (c *Choice) GetPosition() Position { return c.Pos }

// StateNote represents a note attached to a state.
type StateNote struct {
	StateID  string // State the note is attached to
	Text     string // Note text
	Position string // "left of", "right of"
	Pos      Position
}

func (n *StateNote) stateStmt() {}

// GetPosition returns the position in source.
func (n *StateNote) GetPosition() Position { return n.Pos }

// StateComment represents a comment in the state diagram.
type StateComment struct {
	Text string // Comment text (without %%)
	Pos  Position
}

func (c *StateComment) stateStmt() {}

// GetPosition returns the position in source.
func (c *StateComment) GetPosition() Position { return c.Pos }

// GetType returns the diagram type.
func (d *StateDiagram) GetType() string {
	return d.Type
}

// GetPosition returns the position in source.
func (d *StateDiagram) GetPosition() Position {
	return d.Pos
}

// GetSource returns the original source.
func (d *StateDiagram) GetSource() string {
	return d.Source
}
