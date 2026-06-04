package validator

import (
	"fmt"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// SequenceRule defines validation rules for sequence diagrams.
type SequenceRule interface {
	// Name returns the name of the rule.
	Name() string
	// ValidateSequence checks a sequence diagram and returns any validation errors.
	ValidateSequence(diagram *ast.SequenceDiagram) []ValidationError
}

// ValidParticipantReferences checks that all message participants are defined or implicitly created.
type ValidParticipantReferences struct{}

// Name returns the name of this validation rule.
func (r *ValidParticipantReferences) Name() string { return "valid-participant-references" }

// ValidateSequence checks participant references.
func (r *ValidParticipantReferences) ValidateSequence(diagram *ast.SequenceDiagram) []ValidationError {
	var errors []ValidationError

	// Collect all defined and referenced participants
	defined := make(map[string]bool)
	referenced := make(map[string]*ast.Position)

	// Collect participants from all statements
	r.collectParticipants(diagram.Statements, defined, referenced)

	// All referenced participants are implicitly defined in sequence diagrams
	// So we don't need to check if they're defined - this is valid Mermaid behaviour
	// However, we can warn about unused explicit participants

	return errors
}

func (r *ValidParticipantReferences) collectParticipants(statements []ast.SeqStmt, defined map[string]bool, referenced map[string]*ast.Position) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.Participant:
			defined[s.ID] = true

		case *ast.Message:
			if _, exists := referenced[s.From]; !exists {
				pos := s.Pos
				referenced[s.From] = &pos
			}
			if _, exists := referenced[s.To]; !exists {
				pos := s.Pos
				referenced[s.To] = &pos
			}

		case *ast.Activation:
			if _, exists := referenced[s.Participant]; !exists {
				pos := s.Pos
				referenced[s.Participant] = &pos
			}

		case *ast.Note:
			for _, p := range s.Participants {
				if _, exists := referenced[p]; !exists {
					pos := s.Pos
					referenced[p] = &pos
				}
			}

		case *ast.Loop:
			r.collectParticipants(s.Statements, defined, referenced)

		case *ast.Alt:
			for _, cond := range s.Conditions {
				r.collectParticipants(cond.Statements, defined, referenced)
			}

		case *ast.Opt:
			r.collectParticipants(s.Statements, defined, referenced)

		case *ast.Par:
			for _, branch := range s.Branches {
				r.collectParticipants(branch.Statements, defined, referenced)
			}

		case *ast.Critical:
			r.collectParticipants(s.Statements, defined, referenced)
			for _, opt := range s.Options {
				r.collectParticipants(opt.Statements, defined, referenced)
			}

		case *ast.Break:
			r.collectParticipants(s.Statements, defined, referenced)

		case *ast.Box:
			for _, p := range s.Participants {
				defined[p.ID] = true
			}
		}
	}
}

// NoDuplicateParticipants checks that participant IDs are unique.
type NoDuplicateParticipants struct{}

// Name returns the name of this validation rule.
func (r *NoDuplicateParticipants) Name() string { return "no-duplicate-participants" }

// ValidateSequence checks for duplicate participant IDs.
func (r *NoDuplicateParticipants) ValidateSequence(diagram *ast.SequenceDiagram) []ValidationError {
	var errors []ValidationError
	seen := make(map[string]ast.Position)

	r.checkDuplicates(diagram.Statements, seen, &errors)

	return errors
}

func (r *NoDuplicateParticipants) checkDuplicates(statements []ast.SeqStmt, seen map[string]ast.Position, errors *[]ValidationError) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.Participant:
			if firstPos, exists := seen[s.ID]; exists {
				*errors = append(*errors, ValidationError{
					Line:     s.Pos.Line,
					Column:   s.Pos.Column,
					Message:  fmt.Sprintf("duplicate participant ID '%s', first defined at line %d", s.ID, firstPos.Line),
					Severity: SeverityError,
				})
			} else {
				seen[s.ID] = s.Pos
			}

		case *ast.Box:
			for _, p := range s.Participants {
				if firstPos, exists := seen[p.ID]; exists {
					*errors = append(*errors, ValidationError{
						Line:     p.Pos.Line,
						Column:   p.Pos.Column,
						Message:  fmt.Sprintf("duplicate participant ID '%s', first defined at line %d", p.ID, firstPos.Line),
						Severity: SeverityError,
					})
				} else {
					seen[p.ID] = p.Pos
				}
			}

		case *ast.Loop:
			r.checkDuplicates(s.Statements, seen, errors)

		case *ast.Alt:
			for _, cond := range s.Conditions {
				r.checkDuplicates(cond.Statements, seen, errors)
			}

		case *ast.Opt:
			r.checkDuplicates(s.Statements, seen, errors)

		case *ast.Par:
			for _, branch := range s.Branches {
				r.checkDuplicates(branch.Statements, seen, errors)
			}

		case *ast.Critical:
			r.checkDuplicates(s.Statements, seen, errors)
			for _, opt := range s.Options {
				r.checkDuplicates(opt.Statements, seen, errors)
			}

		case *ast.Break:
			r.checkDuplicates(s.Statements, seen, errors)
		}
	}
}

// ValidMessageArrows checks that message arrows are valid.
type ValidMessageArrows struct{}

// Name returns the name of this validation rule.
func (r *ValidMessageArrows) Name() string { return "valid-message-arrows" }

// ValidateSequence checks message arrow syntax.
func (r *ValidMessageArrows) ValidateSequence(diagram *ast.SequenceDiagram) []ValidationError {
	var errors []ValidationError
	validArrows := map[string]bool{
		"->":     true,
		"-->":    true,
		"->>":    true,
		"-->>":   true,
		"-x":     true,
		"--x":    true,
		"-)":     true,
		"--)":    true,
		"<<->>":  true,
		"<<-->>": true,
	}

	r.checkArrows(diagram.Statements, validArrows, &errors)

	return errors
}

func (r *ValidMessageArrows) checkArrows(statements []ast.SeqStmt, validArrows map[string]bool, errors *[]ValidationError) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.Message:
			if !validArrows[s.Arrow] {
				*errors = append(*errors, ValidationError{
					Line:     s.Pos.Line,
					Column:   s.Pos.Column,
					Message:  fmt.Sprintf("invalid message arrow '%s'", s.Arrow),
					Severity: SeverityError,
				})
			}

		case *ast.Loop:
			r.checkArrows(s.Statements, validArrows, errors)

		case *ast.Alt:
			for _, cond := range s.Conditions {
				r.checkArrows(cond.Statements, validArrows, errors)
			}

		case *ast.Opt:
			r.checkArrows(s.Statements, validArrows, errors)

		case *ast.Par:
			for _, branch := range s.Branches {
				r.checkArrows(branch.Statements, validArrows, errors)
			}

		case *ast.Critical:
			r.checkArrows(s.Statements, validArrows, errors)
			for _, opt := range s.Options {
				r.checkArrows(opt.Statements, validArrows, errors)
			}

		case *ast.Break:
			r.checkArrows(s.Statements, validArrows, errors)
		}
	}
}

// ValidNotePositions checks that notes reference valid participants.
type ValidNotePositions struct{}

// Name returns the name of this validation rule.
func (r *ValidNotePositions) Name() string { return "valid-note-positions" }

// ValidateSequence checks note participant references.
func (r *ValidNotePositions) ValidateSequence(diagram *ast.SequenceDiagram) []ValidationError {
	var errors []ValidationError

	// Collect all participants (explicit and implicit)
	participants := make(map[string]bool)
	r.collectAllParticipants(diagram.Statements, participants)

	// Check notes
	r.checkNotes(diagram.Statements, participants, &errors)

	return errors
}

func (r *ValidNotePositions) collectAllParticipants(statements []ast.SeqStmt, participants map[string]bool) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.Participant:
			participants[s.ID] = true

		case *ast.Message:
			participants[s.From] = true
			participants[s.To] = true

		case *ast.Activation:
			participants[s.Participant] = true

		case *ast.Box:
			for _, p := range s.Participants {
				participants[p.ID] = true
			}

		case *ast.Loop:
			r.collectAllParticipants(s.Statements, participants)

		case *ast.Alt:
			for _, cond := range s.Conditions {
				r.collectAllParticipants(cond.Statements, participants)
			}

		case *ast.Opt:
			r.collectAllParticipants(s.Statements, participants)

		case *ast.Par:
			for _, branch := range s.Branches {
				r.collectAllParticipants(branch.Statements, participants)
			}

		case *ast.Critical:
			r.collectAllParticipants(s.Statements, participants)
			for _, opt := range s.Options {
				r.collectAllParticipants(opt.Statements, participants)
			}

		case *ast.Break:
			r.collectAllParticipants(s.Statements, participants)
		}
	}
}

func (r *ValidNotePositions) checkNotes(statements []ast.SeqStmt, participants map[string]bool, errors *[]ValidationError) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.Note:
			for _, p := range s.Participants {
				if !participants[p] {
					*errors = append(*errors, ValidationError{
						Line:     s.Pos.Line,
						Column:   s.Pos.Column,
						Message:  fmt.Sprintf("note references undefined participant '%s'", p),
						Severity: SeverityWarning,
					})
				}
			}

		case *ast.Loop:
			r.checkNotes(s.Statements, participants, errors)

		case *ast.Alt:
			for _, cond := range s.Conditions {
				r.checkNotes(cond.Statements, participants, errors)
			}

		case *ast.Opt:
			r.checkNotes(s.Statements, participants, errors)

		case *ast.Par:
			for _, branch := range s.Branches {
				r.checkNotes(branch.Statements, participants, errors)
			}

		case *ast.Critical:
			r.checkNotes(s.Statements, participants, errors)
			for _, opt := range s.Options {
				r.checkNotes(opt.Statements, participants, errors)
			}

		case *ast.Break:
			r.checkNotes(s.Statements, participants, errors)
		}
	}
}

// SequenceDefaultRules returns default validation rules for sequence diagrams.
func SequenceDefaultRules() []SequenceRule {
	return []SequenceRule{
		&ValidParticipantReferences{},
		&NoDuplicateParticipants{},
		&ValidMessageArrows{},
		&ValidNotePositions{},
	}
}

// SequenceStrictRules returns strict validation rules for sequence diagrams.
func SequenceStrictRules() []SequenceRule {
	return SequenceDefaultRules()
}
