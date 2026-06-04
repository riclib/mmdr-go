package ast

// SequenceDiagram represents a complete Mermaid sequence diagram.
type SequenceDiagram struct {
	Type       string    // "sequence"
	Statements []SeqStmt // All statements in the diagram
	Source     string    // Original source
	Pos        Position  // Position in source
}

// GetType returns the diagram type.
func (s *SequenceDiagram) GetType() string { return s.Type }

// GetPosition returns the position of this diagram in the source.
func (s *SequenceDiagram) GetPosition() Position { return s.Pos }

// GetSource returns the original source.
func (s *SequenceDiagram) GetSource() string { return s.Source }

// SeqStmt represents any statement in a sequence diagram.
type SeqStmt interface {
	seqStmt()
	GetPosition() Position
}

// Participant represents a participant declaration.
type Participant struct {
	ID    string // Participant identifier
	Alias string // Display name (optional)
	Type  string // "participant", "actor", "boundary", "control", "entity", "database", "collections", "queue"
	Pos   Position
}

func (p *Participant) seqStmt() {}

// GetPosition returns the position of this participant in the source.
func (p *Participant) GetPosition() Position { return p.Pos }

// Message represents a message between participants.
type Message struct {
	From       string // Source participant ID
	To         string // Target participant ID
	Arrow      string // Arrow type: "->", "-->", "->>", "-->>", "-x", "--x", "-)", "--)", "<<->>", "<<-->>"
	Text       string // Message text (optional)
	Activate   bool   // Activate target on this message
	Deactivate bool   // Deactivate source on this message
	Pos        Position
}

func (m *Message) seqStmt() {}

// GetPosition returns the position of this message in the source.
func (m *Message) GetPosition() Position { return m.Pos }

// Activation represents explicit activation/deactivation.
type Activation struct {
	Participant string // Participant ID
	Active      bool   // true for activate, false for deactivate
	Pos         Position
}

func (a *Activation) seqStmt() {}

// GetPosition returns the position of this activation in the source.
func (a *Activation) GetPosition() Position { return a.Pos }

// Loop represents a loop block.
type Loop struct {
	Label      string    // Loop description
	Statements []SeqStmt // Nested statements
	Pos        Position
}

func (l *Loop) seqStmt() {}

// GetPosition returns the position of this loop in the source.
func (l *Loop) GetPosition() Position { return l.Pos }

// Alt represents an alternative (if/else) block.
type Alt struct {
	Conditions []AltCondition // Alt/else branches
	Pos        Position
}

// AltCondition represents one branch of an alt block.
type AltCondition struct {
	Label      string    // Condition description
	Statements []SeqStmt // Statements in this branch
	IsElse     bool      // true for else branch
}

func (a *Alt) seqStmt() {}

// GetPosition returns the position of this alt block in the source.
func (a *Alt) GetPosition() Position { return a.Pos }

// Opt represents an optional block.
type Opt struct {
	Label      string    // Condition description
	Statements []SeqStmt // Nested statements
	Pos        Position
}

func (o *Opt) seqStmt() {}

// GetPosition returns the position of this opt block in the source.
func (o *Opt) GetPosition() Position { return o.Pos }

// Par represents a parallel block.
type Par struct {
	Branches []ParBranch // Parallel branches
	Pos      Position
}

// ParBranch represents one parallel execution path.
type ParBranch struct {
	Label      string    // Branch description
	Statements []SeqStmt // Statements in this branch
}

func (p *Par) seqStmt() {}

// GetPosition returns the position of this par block in the source.
func (p *Par) GetPosition() Position { return p.Pos }

// Critical represents a critical region block.
type Critical struct {
	Label      string           // Description
	Options    []CriticalOption // Critical option branches
	Statements []SeqStmt        // Main statements
	Pos        Position
}

// CriticalOption represents an option branch in a critical block.
type CriticalOption struct {
	Label      string    // Option description
	Statements []SeqStmt // Statements in this option
}

func (c *Critical) seqStmt() {}

// GetPosition returns the position of this critical block in the source.
func (c *Critical) GetPosition() Position { return c.Pos }

// Break represents a break block.
type Break struct {
	Label      string    // Break description
	Statements []SeqStmt // Nested statements
	Pos        Position
}

func (b *Break) seqStmt() {}

// GetPosition returns the position of this break block in the source.
func (b *Break) GetPosition() Position { return b.Pos }

// Note represents a note attached to participants.
type Note struct {
	Position     string   // "left of", "right of", "over"
	Participants []string // Participant IDs
	Text         string   // Note content
	Pos          Position
}

func (n *Note) seqStmt() {}

// GetPosition returns the position of this note in the source.
func (n *Note) GetPosition() Position { return n.Pos }

// Box represents a grouping box around participants.
type Box struct {
	Colour       string        // Box colour (optional)
	Label        string        // Box label
	Participants []Participant // Participants in this box
	Pos          Position
}

func (b *Box) seqStmt() {}

// GetPosition returns the position of this box in the source.
func (b *Box) GetPosition() Position { return b.Pos }

// Autonumber represents the autonumber directive.
type Autonumber struct {
	Enabled bool // Enable/disable autonumbering
	Pos     Position
}

func (a *Autonumber) seqStmt() {}

// GetPosition returns the position of this autonumber directive in the source.
func (a *Autonumber) GetPosition() Position { return a.Pos }

// SeqComment represents a comment in the sequence diagram.
type SeqComment struct {
	Text string
	Pos  Position
}

func (c *SeqComment) seqStmt() {}

// GetPosition returns the position of this comment in the source.
func (c *SeqComment) GetPosition() Position { return c.Pos }
