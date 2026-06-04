package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

var (
	// Sequence diagram patterns
	seqHeaderPattern  = regexp.MustCompile(`^sequenceDiagram\s*$`)
	seqCommentPattern = regexp.MustCompile(`^%%(.*)$`)

	// Participant patterns
	participantPattern = regexp.MustCompile(`^(participant|actor)\s+(\w+)(?:\s+as\s+(.+))?$`)

	// Activation patterns
	activatePattern   = regexp.MustCompile(`^activate\s+(\w+)$`)
	deactivatePattern = regexp.MustCompile(`^deactivate\s+(\w+)$`)

	// Block patterns
	loopPattern     = regexp.MustCompile(`^loop\s+(.+)$`)
	altPattern      = regexp.MustCompile(`^alt\s+(.+)$`)
	elsePattern     = regexp.MustCompile(`^else(?:\s+(.+))?$`)
	optPattern      = regexp.MustCompile(`^opt\s+(.+)$`)
	parPattern      = regexp.MustCompile(`^par\s+(.+)$`)
	andPattern      = regexp.MustCompile(`^and(?:\s+(.+))?$`)
	criticalPattern = regexp.MustCompile(`^critical\s+(.+)$`)
	optionPattern   = regexp.MustCompile(`^option\s+(.+)$`)
	breakPattern    = regexp.MustCompile(`^break\s+(.+)$`)
	endPattern      = regexp.MustCompile(`^end\s*$`)

	// Note patterns (case-insensitive to match Mermaid spec)
	noteLeftPattern  = regexp.MustCompile(`(?i)^note\s+left\s+of\s+(\w+)\s*:\s*(.+)$`)
	noteRightPattern = regexp.MustCompile(`(?i)^note\s+right\s+of\s+(\w+)\s*:\s*(.+)$`)
	noteOverPattern  = regexp.MustCompile(`(?i)^note\s+over\s+([\w,\s]+)\s*:\s*(.+)$`)

	// Box pattern
	boxPattern = regexp.MustCompile(`^box\s+(?:(\w+)\s+)?(.+)$`)

	// Autonumber pattern
	autonumberPattern = regexp.MustCompile(`^autonumber\s*$`)
)

// SequenceParser parses Mermaid sequence diagrams.
type SequenceParser struct{}

// NewSequenceParser creates a new sequence diagram parser.
func NewSequenceParser() *SequenceParser {
	return &SequenceParser{}
}

// Parse parses a Mermaid sequence diagram from a string.
func (p *SequenceParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")

	// Check header
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty sequence diagram")
	}

	// Find first non-comment, non-empty line
	headerLine := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "%%") {
			headerLine = i
			break
		}
	}

	if headerLine == -1 {
		return nil, fmt.Errorf("sequence diagram has no content")
	}

	trimmedHeader := strings.TrimSpace(lines[headerLine])
	if !seqHeaderPattern.MatchString(trimmedHeader) {
		return nil, fmt.Errorf("line %d: invalid sequence diagram header, expected 'sequenceDiagram'", headerLine+1)
	}

	diagram := &ast.SequenceDiagram{
		Type:   "sequence",
		Source: source,
		Pos:    ast.Position{Line: 1, Column: 1},
	}

	// Parse statements
	statements, err := p.parseStatements(lines[headerLine+1:], headerLine+2)
	if err != nil {
		return nil, err
	}

	diagram.Statements = statements

	return diagram, nil
}

// SupportedTypes returns the diagram types this parser handles.
func (p *SequenceParser) SupportedTypes() []string {
	return []string{"sequence"}
}

func (p *SequenceParser) parseStatements(lines []string, startLine int) ([]ast.SeqStmt, error) {
	var statements []ast.SeqStmt
	lineNum := startLine

	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		pos := ast.Position{Line: lineNum, Column: 1}
		lineNum++

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Skip comments
		if seqCommentPattern.MatchString(trimmed) {
			continue
		}

		// Try to parse statement
		stmt, consumed, err := p.parseStatement(lines[i:], pos, i+startLine)
		if err != nil {
			return nil, err
		}

		if stmt != nil {
			statements = append(statements, stmt)
		}

		// Skip consumed lines
		if consumed > 1 {
			i += consumed - 1
			lineNum += consumed - 1
		}
	}

	return statements, nil
}

func (p *SequenceParser) parseStatement(lines []string, pos ast.Position, lineNum int) (ast.SeqStmt, int, error) {
	if len(lines) == 0 {
		return nil, 0, nil
	}

	trimmed := strings.TrimSpace(lines[0])

	// Participant/actor
	if matches := participantPattern.FindStringSubmatch(trimmed); matches != nil {
		return &ast.Participant{
			ID:    matches[2],
			Alias: matches[3],
			Type:  matches[1],
			Pos:   pos,
		}, 1, nil
	}

	// Activation
	if matches := activatePattern.FindStringSubmatch(trimmed); matches != nil {
		return &ast.Activation{
			Participant: matches[1],
			Active:      true,
			Pos:         pos,
		}, 1, nil
	}

	if matches := deactivatePattern.FindStringSubmatch(trimmed); matches != nil {
		return &ast.Activation{
			Participant: matches[1],
			Active:      false,
			Pos:         pos,
		}, 1, nil
	}

	// Loop block
	if matches := loopPattern.FindStringSubmatch(trimmed); matches != nil {
		blockLines, consumed, err := p.extractBlock(lines[1:], lineNum+1)
		if err != nil {
			return nil, 0, err
		}

		statements, err := p.parseStatements(blockLines, lineNum+1)
		if err != nil {
			return nil, 0, err
		}

		return &ast.Loop{
			Label:      matches[1],
			Statements: statements,
			Pos:        pos,
		}, consumed + 1, nil
	}

	// Alt block
	if matches := altPattern.FindStringSubmatch(trimmed); matches != nil {
		return p.parseAltBlock(lines, pos, lineNum, matches[1])
	}

	// Opt block
	if matches := optPattern.FindStringSubmatch(trimmed); matches != nil {
		blockLines, consumed, err := p.extractBlock(lines[1:], lineNum+1)
		if err != nil {
			return nil, 0, err
		}

		statements, err := p.parseStatements(blockLines, lineNum+1)
		if err != nil {
			return nil, 0, err
		}

		return &ast.Opt{
			Label:      matches[1],
			Statements: statements,
			Pos:        pos,
		}, consumed + 1, nil
	}

	// Par block
	if matches := parPattern.FindStringSubmatch(trimmed); matches != nil {
		return p.parseParBlock(lines, pos, lineNum, matches[1])
	}

	// Critical block
	if matches := criticalPattern.FindStringSubmatch(trimmed); matches != nil {
		return p.parseCriticalBlock(lines, pos, lineNum, matches[1])
	}

	// Break block
	if matches := breakPattern.FindStringSubmatch(trimmed); matches != nil {
		blockLines, consumed, err := p.extractBlock(lines[1:], lineNum+1)
		if err != nil {
			return nil, 0, err
		}

		statements, err := p.parseStatements(blockLines, lineNum+1)
		if err != nil {
			return nil, 0, err
		}

		return &ast.Break{
			Label:      matches[1],
			Statements: statements,
			Pos:        pos,
		}, consumed + 1, nil
	}

	// Box
	if matches := boxPattern.FindStringSubmatch(trimmed); matches != nil {
		return p.parseBoxBlock(lines, pos, lineNum, matches[1], matches[2])
	}

	// Notes
	if matches := noteLeftPattern.FindStringSubmatch(trimmed); matches != nil {
		return &ast.Note{
			Position:     "left of",
			Participants: []string{matches[1]},
			Text:         matches[2],
			Pos:          pos,
		}, 1, nil
	}

	if matches := noteRightPattern.FindStringSubmatch(trimmed); matches != nil {
		return &ast.Note{
			Position:     "right of",
			Participants: []string{matches[1]},
			Text:         matches[2],
			Pos:          pos,
		}, 1, nil
	}

	if matches := noteOverPattern.FindStringSubmatch(trimmed); matches != nil {
		participants := strings.Split(strings.ReplaceAll(matches[1], " ", ""), ",")
		return &ast.Note{
			Position:     "over",
			Participants: participants,
			Text:         matches[2],
			Pos:          pos,
		}, 1, nil
	}

	// Autonumber
	if autonumberPattern.MatchString(trimmed) {
		return &ast.Autonumber{
			Enabled: true,
			Pos:     pos,
		}, 1, nil
	}

	// Message (try this last as it's more permissive)
	if msg := p.parseMessage(trimmed, pos); msg != nil {
		return msg, 1, nil
	}

	// Unknown statement
	return nil, 0, fmt.Errorf("line %d: unknown sequence diagram statement: %s", pos.Line, trimmed)
}

func (p *SequenceParser) parseMessage(line string, pos ast.Position) *ast.Message {
	// Try different arrow patterns
	arrows := []string{
		"<<-->>", "<<->>", // Bidirectional
		"-->>", "->>", "--x", "-x", "--)", "-)", "-->", "->", // Unidirectional
	}

	for _, arrow := range arrows {
		if before, after, ok := strings.Cut(line, arrow); ok {
			from := strings.TrimSpace(before)
			rest := strings.TrimSpace(after)

			// Check for activation/deactivation markers
			activate := strings.HasSuffix(rest, "+")
			deactivate := strings.HasSuffix(rest, "-")
			if activate || deactivate {
				rest = strings.TrimSuffix(strings.TrimSuffix(rest, "+"), "-")
				rest = strings.TrimSpace(rest)
			}

			// Split on colon for message text
			parts := strings.SplitN(rest, ":", 2)
			to := strings.TrimSpace(parts[0])
			text := ""
			if len(parts) > 1 {
				text = strings.TrimSpace(parts[1])
			}

			// Validate participant IDs
			if !isValidID(from) || !isValidID(to) {
				continue
			}

			return &ast.Message{
				From:       from,
				To:         to,
				Arrow:      arrow,
				Text:       text,
				Activate:   activate,
				Deactivate: deactivate,
				Pos:        pos,
			}
		}
	}

	return nil
}

func (p *SequenceParser) extractBlock(lines []string, startLine int) ([]string, int, error) {
	var blockLines []string //nolint:prealloc // Size cannot be determined beforehand
	depth := 1
	consumed := 0

	for _, line := range lines {
		consumed++
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			blockLines = append(blockLines, line)
			continue
		}

		// Check for nested blocks
		if loopPattern.MatchString(trimmed) || altPattern.MatchString(trimmed) ||
			optPattern.MatchString(trimmed) || parPattern.MatchString(trimmed) ||
			criticalPattern.MatchString(trimmed) || breakPattern.MatchString(trimmed) {
			depth++
		}

		// Check for end
		if endPattern.MatchString(trimmed) {
			depth--
			if depth == 0 {
				return blockLines, consumed, nil
			}
		}

		blockLines = append(blockLines, line)
	}

	return nil, 0, fmt.Errorf("line %d: unclosed block, missing 'end'", startLine)
}

func (p *SequenceParser) parseAltBlock(lines []string, pos ast.Position, lineNum int, firstLabel string) (ast.SeqStmt, int, error) {
	var conditions []ast.AltCondition
	currentCondition := ast.AltCondition{
		Label:  firstLabel,
		IsElse: false,
	}

	consumed := 1
	depth := 1
	var currentLines []string

	for i := 1; i < len(lines); i++ {
		consumed++
		trimmed := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			currentLines = append(currentLines, lines[i])
			continue
		}

		// Check for nested blocks
		if loopPattern.MatchString(trimmed) || altPattern.MatchString(trimmed) ||
			optPattern.MatchString(trimmed) || parPattern.MatchString(trimmed) ||
			criticalPattern.MatchString(trimmed) || breakPattern.MatchString(trimmed) {
			depth++
			currentLines = append(currentLines, lines[i])
			continue
		}

		// Check for else at same depth
		if depth == 1 && elsePattern.MatchString(trimmed) {
			// Save current condition
			statements, err := p.parseStatements(currentLines, lineNum+1)
			if err != nil {
				return nil, 0, err
			}
			currentCondition.Statements = statements
			conditions = append(conditions, currentCondition)

			// Start else condition
			matches := elsePattern.FindStringSubmatch(trimmed)
			currentCondition = ast.AltCondition{
				Label:  matches[1],
				IsElse: true,
			}
			currentLines = nil
			continue
		}

		// Check for end
		if endPattern.MatchString(trimmed) {
			depth--
			if depth == 0 {
				// Save last condition
				statements, err := p.parseStatements(currentLines, lineNum+1)
				if err != nil {
					return nil, 0, err
				}
				currentCondition.Statements = statements
				conditions = append(conditions, currentCondition)

				return &ast.Alt{
					Conditions: conditions,
					Pos:        pos,
				}, consumed, nil
			}
			currentLines = append(currentLines, lines[i])
			continue
		}

		currentLines = append(currentLines, lines[i])
	}

	return nil, 0, fmt.Errorf("line %d: unclosed alt block, missing 'end'", lineNum)
}

func (p *SequenceParser) parseParBlock(lines []string, pos ast.Position, lineNum int, firstLabel string) (ast.SeqStmt, int, error) {
	var branches []ast.ParBranch
	currentBranch := ast.ParBranch{
		Label: firstLabel,
	}

	consumed := 1
	depth := 1
	var currentLines []string

	for i := 1; i < len(lines); i++ {
		consumed++
		trimmed := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			currentLines = append(currentLines, lines[i])
			continue
		}

		// Check for nested blocks
		if loopPattern.MatchString(trimmed) || altPattern.MatchString(trimmed) ||
			optPattern.MatchString(trimmed) || parPattern.MatchString(trimmed) ||
			criticalPattern.MatchString(trimmed) || breakPattern.MatchString(trimmed) {
			depth++
			currentLines = append(currentLines, lines[i])
			continue
		}

		// Check for and at same depth
		if depth == 1 && andPattern.MatchString(trimmed) {
			// Save current branch
			statements, err := p.parseStatements(currentLines, lineNum+1)
			if err != nil {
				return nil, 0, err
			}
			currentBranch.Statements = statements
			branches = append(branches, currentBranch)

			// Start new branch
			matches := andPattern.FindStringSubmatch(trimmed)
			currentBranch = ast.ParBranch{
				Label: matches[1],
			}
			currentLines = nil
			continue
		}

		// Check for end
		if endPattern.MatchString(trimmed) {
			depth--
			if depth == 0 {
				// Save last branch
				statements, err := p.parseStatements(currentLines, lineNum+1)
				if err != nil {
					return nil, 0, err
				}
				currentBranch.Statements = statements
				branches = append(branches, currentBranch)

				return &ast.Par{
					Branches: branches,
					Pos:      pos,
				}, consumed, nil
			}
			currentLines = append(currentLines, lines[i])
			continue
		}

		currentLines = append(currentLines, lines[i])
	}

	return nil, 0, fmt.Errorf("line %d: unclosed par block, missing 'end'", lineNum)
}

func (p *SequenceParser) parseCriticalBlock(lines []string, pos ast.Position, lineNum int, label string) (ast.SeqStmt, int, error) {
	var options []ast.CriticalOption
	var mainStatements []ast.SeqStmt

	consumed := 1
	depth := 1
	var currentLines []string
	inOption := false
	var currentOption ast.CriticalOption

	for i := 1; i < len(lines); i++ {
		consumed++
		trimmed := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			currentLines = append(currentLines, lines[i])
			continue
		}

		// Check for nested blocks
		if loopPattern.MatchString(trimmed) || altPattern.MatchString(trimmed) ||
			optPattern.MatchString(trimmed) || parPattern.MatchString(trimmed) ||
			criticalPattern.MatchString(trimmed) || breakPattern.MatchString(trimmed) {
			depth++
			currentLines = append(currentLines, lines[i])
			continue
		}

		// Check for option at same depth
		if depth == 1 && optionPattern.MatchString(trimmed) {
			if !inOption {
				// Save main statements
				statements, err := p.parseStatements(currentLines, lineNum+1)
				if err != nil {
					return nil, 0, err
				}
				mainStatements = statements
				inOption = true
			} else {
				// Save previous option
				statements, err := p.parseStatements(currentLines, lineNum+1)
				if err != nil {
					return nil, 0, err
				}
				currentOption.Statements = statements
				options = append(options, currentOption)
			}

			// Start new option
			matches := optionPattern.FindStringSubmatch(trimmed)
			currentOption = ast.CriticalOption{
				Label: matches[1],
			}
			currentLines = nil
			continue
		}

		// Check for end
		if endPattern.MatchString(trimmed) {
			depth--
			if depth == 0 {
				if inOption {
					// Save last option
					statements, err := p.parseStatements(currentLines, lineNum+1)
					if err != nil {
						return nil, 0, err
					}
					currentOption.Statements = statements
					options = append(options, currentOption)
				} else {
					// No options, save as main statements
					statements, err := p.parseStatements(currentLines, lineNum+1)
					if err != nil {
						return nil, 0, err
					}
					mainStatements = statements
				}

				return &ast.Critical{
					Label:      label,
					Options:    options,
					Statements: mainStatements,
					Pos:        pos,
				}, consumed, nil
			}
			currentLines = append(currentLines, lines[i])
			continue
		}

		currentLines = append(currentLines, lines[i])
	}

	return nil, 0, fmt.Errorf("line %d: unclosed critical block, missing 'end'", lineNum)
}

func (p *SequenceParser) parseBoxBlock(lines []string, pos ast.Position, lineNum int, colour, label string) (ast.SeqStmt, int, error) {
	var participants []ast.Participant
	consumed := 1

	for i := 1; i < len(lines); i++ {
		consumed++
		trimmed := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		// Check for end
		if endPattern.MatchString(trimmed) {
			return &ast.Box{
				Colour:       colour,
				Label:        label,
				Participants: participants,
				Pos:          pos,
			}, consumed, nil
		}

		// Parse participant
		if matches := participantPattern.FindStringSubmatch(trimmed); matches != nil {
			participants = append(participants, ast.Participant{
				ID:    matches[2],
				Alias: matches[3],
				Type:  matches[1],
				Pos:   ast.Position{Line: lineNum + i, Column: 1},
			})
		}
	}

	return nil, 0, fmt.Errorf("line %d: unclosed box, missing 'end'", lineNum)
}

func isValidID(id string) bool {
	if id == "" {
		return false
	}
	// Check if ID contains only alphanumeric and underscore
	for _, ch := range id {
		if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') &&
			(ch < '0' || ch > '9') && ch != '_' {
			return false
		}
	}
	return true
}
