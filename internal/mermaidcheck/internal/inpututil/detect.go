// Package inpututil provides utilities for detecting and processing input file types.
package inpututil

import (
	"path/filepath"
	"strings"
)

// FileType represents the type of input file.
type FileType int

const (
	// FileTypeUnknown represents an unknown file type.
	FileTypeUnknown FileType = iota
	// FileTypeMermaid represents a pure Mermaid file (.mmd).
	FileTypeMermaid
	// FileTypeMarkdown represents a markdown file (.md, .markdown, .mdx).
	FileTypeMarkdown
)

// DetectFileType determines the file type from the file extension.
func DetectFileType(filename string) FileType {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".mmd":
		return FileTypeMermaid
	case ".md", ".markdown", ".mdx":
		return FileTypeMarkdown
	default:
		return FileTypeUnknown
	}
}

// IsMarkdownExtension returns true if the extension indicates a markdown file.
func IsMarkdownExtension(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".md" || ext == ".markdown" || ext == ".mdx"
}

// IsMermaidExtension returns true if the extension indicates a Mermaid file.
func IsMermaidExtension(ext string) bool {
	return strings.ToLower(ext) == ".mmd"
}
