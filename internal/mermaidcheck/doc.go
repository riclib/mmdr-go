// Package mermaid provides parsing, validation, and linting for Mermaid diagram syntax.
//
// This package allows you to parse Mermaid diagrams from raw text or extract them from
// markdown files, then validate the syntax and apply custom linting rules.
//
// # Basic Usage
//
// Parse a raw Mermaid diagram:
//
//	diagram, err := mermaid.Parse(source)
//	if err != nil {
//	    // Handle parse error
//	}
//
// Extract diagrams from markdown:
//
//	diagrams, err := mermaid.ExtractFromMarkdown(markdownContent)
//	for _, diagram := range diagrams {
//	    // Process each diagram
//	}
//
// Validate with custom rules:
//
//	errors := mermaid.Validate(diagram, mermaid.NoParenthesesInLabels)
//
// # Supported Diagram Types
//
// Currently supports flowchart and graph diagrams. Additional diagram types
// (sequence, ER, class, state, gantt, pie, etc.) will be added in future releases.
package mermaidcheck
