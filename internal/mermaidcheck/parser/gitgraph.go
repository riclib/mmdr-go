package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// GitGraphParser handles parsing of git graph diagrams.
type GitGraphParser struct{}

// NewGitGraphParser creates a new git graph parser.
func NewGitGraphParser() *GitGraphParser {
	return &GitGraphParser{}
}

var (
	gitGraphHeaderRegex   = regexp.MustCompile(`^gitGraph\s*$`)
	gitGraphThemeRegex    = regexp.MustCompile(`^\s*%%\{init:\s*\{\s*'theme'\s*:\s*'([^']+)'\s*\}\s*\}%%\s*$`)
	gitGraphCommitRegex   = regexp.MustCompile(`^\s*commit(?:\s+id:\s*"([^"]+)")?(?:\s+tag:\s*"([^"]+)")?(?:\s+type:\s*(NORMAL|REVERSE|HIGHLIGHT))?\s*$`)
	gitGraphBranchRegex   = regexp.MustCompile(`^\s*branch\s+([\w-]+)(?:\s+order:\s*(\d+))?\s*$`)
	gitGraphCheckoutRegex = regexp.MustCompile(`^\s*checkout\s+([\w-]+)\s*$`)
	gitGraphMergeRegex    = regexp.MustCompile(`^\s*merge\s+([\w-]+)(?:\s+id:\s*"([^"]+)")?(?:\s+tag:\s*"([^"]+)")?(?:\s+type:\s*(NORMAL|REVERSE|HIGHLIGHT))?\s*$`)
	gitGraphCherryRegex   = regexp.MustCompile(`^\s*cherry-pick\s+id:\s*"([^"]+)"(?:\s+tag:\s*"([^"]+)")?\s*$`)
	gitGraphOptionRegex   = regexp.MustCompile(`^\s*(mainBranchName|mainBranchOrder)\s*:\s*(.+)\s*$`)
)

// Parse parses a git graph diagram source.
func (p *GitGraphParser) Parse(source string) (ast.Diagram, error) {
	lines := strings.Split(source, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diagram source")
	}

	diagram := &ast.GitGraphDiagram{
		Type:       "gitGraph",
		Source:     source,
		Operations: []ast.GitOperation{},
		Pos:        ast.Position{Line: 1, Column: 1},
	}

	// Find header line, skipping config comments
	headerIdx := -1
	for i := range lines {
		trimmed := strings.TrimSpace(lines[i])
		// Skip empty lines
		if trimmed == "" {
			continue
		}
		// Check for theme config comment
		if themeMatches := gitGraphThemeRegex.FindStringSubmatch(trimmed); themeMatches != nil {
			diagram.Theme = themeMatches[1]
			continue
		}
		// Skip regular comments
		if strings.HasPrefix(trimmed, "%%") {
			continue
		}
		// Found first non-comment, non-empty line - should be header
		if !gitGraphHeaderRegex.MatchString(trimmed) {
			return nil, fmt.Errorf("invalid gitGraph header: %s", trimmed)
		}
		headerIdx = i
		break
	}

	if headerIdx == -1 {
		return nil, fmt.Errorf("no gitGraph header found")
	}

	// Parse operations
	for i := headerIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		pos := ast.Position{Line: i + 1, Column: 1}

		// Try to match commit
		if matches := gitGraphCommitRegex.FindStringSubmatch(trimmed); matches != nil {
			op := ast.GitOperation{
				Type: "commit",
				ID:   matches[1],
				Tag:  matches[2],
				Pos:  pos,
			}
			if matches[3] != "" {
				op.CommitType = matches[3]
			}
			diagram.Operations = append(diagram.Operations, op)
			continue
		}

		// Try to match branch
		if matches := gitGraphBranchRegex.FindStringSubmatch(trimmed); matches != nil {
			op := ast.GitOperation{
				Type:       "branch",
				BranchName: matches[1],
				Pos:        pos,
			}
			if matches[2] != "" {
				order, err := strconv.Atoi(matches[2])
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid branch order: %s", i+1, matches[2])
				}
				op.Order = order
			}
			diagram.Operations = append(diagram.Operations, op)
			continue
		}

		// Try to match checkout
		if matches := gitGraphCheckoutRegex.FindStringSubmatch(trimmed); matches != nil {
			op := ast.GitOperation{
				Type:       "checkout",
				BranchName: matches[1],
				Pos:        pos,
			}
			diagram.Operations = append(diagram.Operations, op)
			continue
		}

		// Try to match merge
		if matches := gitGraphMergeRegex.FindStringSubmatch(trimmed); matches != nil {
			op := ast.GitOperation{
				Type:       "merge",
				BranchName: matches[1],
				ID:         matches[2],
				Tag:        matches[3],
				Pos:        pos,
			}
			if matches[4] != "" {
				op.CommitType = matches[4]
			}
			diagram.Operations = append(diagram.Operations, op)
			continue
		}

		// Try to match cherry-pick
		if matches := gitGraphCherryRegex.FindStringSubmatch(trimmed); matches != nil {
			op := ast.GitOperation{
				Type:     "cherry-pick",
				ParentID: matches[1],
				Tag:      matches[2],
				Pos:      pos,
			}
			diagram.Operations = append(diagram.Operations, op)
			continue
		}

		// Try to match options
		if matches := gitGraphOptionRegex.FindStringSubmatch(trimmed); matches != nil {
			option := matches[1]
			value := strings.Trim(matches[2], `"' `)

			switch option {
			case "mainBranchName":
				diagram.MainBranchName = value
			case "mainBranchOrder":
				order, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid mainBranchOrder: %s", i+1, value)
				}
				diagram.MainBranchOrder = order
			}
			continue
		}

		return nil, fmt.Errorf("line %d: unrecognised gitGraph syntax: %s", i+1, trimmed)
	}

	if len(diagram.Operations) == 0 {
		return nil, fmt.Errorf("gitGraph must have at least one operation")
	}

	return diagram, nil
}

// SupportedTypes returns the diagram types this parser supports.
func (p *GitGraphParser) SupportedTypes() []string {
	return []string{"gitGraph"}
}
