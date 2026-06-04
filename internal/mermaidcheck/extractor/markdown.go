// Package extractor provides utilities for extracting Mermaid diagrams from various file formats.
package extractor

import (
	"bufio"
	"fmt"
	"strings"
)

// DiagramBlock represents a Mermaid diagram extracted from a source file.
type DiagramBlock struct {
	// Source contains the raw Mermaid diagram syntax
	Source string
	// LineOffset is the line number in the original file where this diagram starts (1-indexed)
	LineOffset int
	// EndLine is the line number in the original file where this diagram ends (1-indexed)
	EndLine int
	// DiagramType is the type of Mermaid diagram (e.g., "flowchart", "sequence", "graph")
	DiagramType string
}

// ExtractFromMarkdown extracts all Mermaid code blocks from markdown content.
// It returns a slice of DiagramBlock, each containing the diagram source and its position
// in the original markdown file for accurate error reporting.
func ExtractFromMarkdown(markdown string) ([]DiagramBlock, error) {
	var blocks []DiagramBlock
	scanner := bufio.NewScanner(strings.NewReader(markdown))

	var (
		inMermaidBlock bool
		currentBlock   strings.Builder
		blockStartLine int
		lineNum        int
	)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for accidentally escaped backticks that look like code fence attempts
		// Only flag if the escaped backticks appear at the start of the line (after whitespace)
		// and aren't embedded within other text (like inline code or examples)
		if (strings.HasPrefix(trimmed, "\\`\\`\\`mermaid") || strings.HasPrefix(trimmed, "\\`\\`\\`")) &&
			!strings.Contains(line, "`\\`\\`\\`") { // Ignore if it's in inline code like `\`\`\``
			return nil, fmt.Errorf("line %d: escaped backticks found (\\`\\`\\`). Remove backslashes to use proper markdown code fences: ```", lineNum)
		}

		// Check for start of Mermaid code block
		if !inMermaidBlock && (trimmed == "```mermaid" || strings.HasPrefix(trimmed, "```mermaid ")) {
			inMermaidBlock = true
			blockStartLine = lineNum + 1 // Content starts on next line
			currentBlock.Reset()
			continue
		}

		// Check for end of code block
		if inMermaidBlock && trimmed == "```" {
			inMermaidBlock = false
			source := currentBlock.String()

			// Only add non-empty blocks
			if strings.TrimSpace(source) != "" {
				diagramType := detectDiagramType(source)
				blocks = append(blocks, DiagramBlock{
					Source:      source,
					LineOffset:  blockStartLine,
					EndLine:     lineNum - 1, // Content ends on previous line before closing fence
					DiagramType: diagramType,
				})
			}
			continue
		}

		// Collect lines within Mermaid block
		if inMermaidBlock {
			if currentBlock.Len() > 0 {
				currentBlock.WriteByte('\n')
			}
			currentBlock.WriteString(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Handle unclosed block at end of file
	if inMermaidBlock {
		source := currentBlock.String()
		if strings.TrimSpace(source) != "" {
			diagramType := detectDiagramType(source)
			blocks = append(blocks, DiagramBlock{
				Source:      source,
				LineOffset:  blockStartLine,
				EndLine:     lineNum, // Content ends at last line
				DiagramType: diagramType,
			})
		}
	}

	return blocks, nil
}

// detectDiagramType attempts to determine the diagram type from the source.
func detectDiagramType(source string) string {
	lines := strings.SplitSeq(source, "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue // Skip empty lines and comments
		}

		// Check for diagram type keywords in order of specificity
		// State diagrams - check v2 first to avoid matching base stateDiagram
		if strings.HasPrefix(trimmed, "stateDiagram-v2") {
			return "stateDiagram-v2"
		}
		if strings.HasPrefix(trimmed, "stateDiagram") {
			return "state"
		}

		// C4 diagrams - multiple variants
		if strings.HasPrefix(trimmed, "C4Context") {
			return "c4Context"
		}
		if strings.HasPrefix(trimmed, "C4Container") {
			return "c4Container"
		}
		if strings.HasPrefix(trimmed, "C4Component") {
			return "c4Component"
		}
		if strings.HasPrefix(trimmed, "C4Dynamic") {
			return "c4Dynamic"
		}
		if strings.HasPrefix(trimmed, "C4Deployment") {
			return "c4Deployment"
		}

		// Other diagram types
		if strings.HasPrefix(trimmed, "sequenceDiagram") {
			return "sequence"
		}
		if strings.HasPrefix(trimmed, "classDiagram") {
			return "class"
		}
		if strings.HasPrefix(trimmed, "erDiagram") {
			return "er"
		}
		if strings.HasPrefix(trimmed, "gantt") {
			return "gantt"
		}
		if strings.HasPrefix(trimmed, "pie") {
			return "pie"
		}
		if strings.HasPrefix(trimmed, "journey") {
			return "journey"
		}
		if strings.HasPrefix(trimmed, "gitGraph") {
			return "gitGraph"
		}
		if strings.HasPrefix(trimmed, "mindmap") {
			return "mindmap"
		}
		if strings.HasPrefix(trimmed, "timeline") {
			return "timeline"
		}
		if strings.HasPrefix(trimmed, "sankey-beta") {
			return "sankey"
		}
		if strings.HasPrefix(trimmed, "quadrantChart") {
			return "quadrantChart"
		}
		if strings.HasPrefix(trimmed, "xychart-beta") {
			return "xyChart"
		}
		if strings.HasPrefix(trimmed, "flowchart") {
			return "flowchart"
		}
		if strings.HasPrefix(trimmed, "graph") {
			return "graph"
		}

		// If we found a non-empty, non-comment line, stop looking
		break
	}

	return "unknown"
}
