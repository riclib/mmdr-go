// Package parser provides parsing functionality for all Mermaid diagram types.
package parser

import (
	"fmt"
	"strings"

	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// DiagramParser defines the interface all diagram parsers must implement.
type DiagramParser interface {
	// Parse parses the source and returns a Diagram AST.
	Parse(source string) (ast.Diagram, error)
	// SupportedTypes returns the diagram types this parser handles.
	SupportedTypes() []string
}

// Parse parses a Mermaid diagram from source and returns a Diagram.
// It automatically detects the diagram type and uses the appropriate parser.
func Parse(source string) (ast.Diagram, error) {
	if strings.TrimSpace(source) == "" {
		return nil, fmt.Errorf("empty diagram source")
	}

	diagType := detectDiagramType(source)

	// Direct parser instantiation based on type
	var parser DiagramParser
	switch diagType {
	case "flowchart", "graph":
		parser = NewFlowchartParser()
	case "sequence":
		parser = NewSequenceParser()
	case "class":
		parser = NewClassParser()
	case "state", "stateDiagram-v2":
		parser = NewStateParser()
	case "er":
		parser = NewERParser()
	case "gantt":
		parser = NewGanttParser()
	case "pie":
		parser = NewPieParser()
	case "journey":
		parser = NewJourneyParser()
	case "timeline":
		parser = NewTimelineParser()
	case "gitGraph":
		parser = NewGitGraphParser()
	case "mindmap":
		parser = NewMindmapParser()
	case "sankey":
		parser = NewSankeyParser()
	case "quadrantChart":
		parser = NewQuadrantParser()
	case "xyChart":
		parser = NewXYChartParser()
	case "c4Context":
		parser = NewC4ContextParser()
	case "c4Container":
		parser = NewC4ContainerParser()
	case "c4Component":
		parser = NewC4ComponentParser()
	case "c4Dynamic":
		parser = NewC4DynamicParser()
	case "c4Deployment":
		parser = NewC4DeploymentParser()
	default:
		// Fallback to GenericDiagram for known types without specific parsers
		if isKnownDiagramType(diagType) {
			return ast.NewGenericDiagram(diagType, source, ast.Position{Line: 1, Column: 1}), nil
		}
		supportedTypes := "flowchart, graph, sequence, class, state, stateDiagram-v2, er, gantt, pie, journey, gitGraph, mindmap, timeline, sankey, quadrantChart, xyChart, c4Context, c4Container, c4Component, c4Dynamic, c4Deployment"
		return nil, fmt.Errorf("unknown or unsupported diagram type %q: expected one of: %s", diagType, supportedTypes)
	}

	return parser.Parse(source)
}

// diagramTypeMapping maps Mermaid diagram prefixes to normalized type names.
// Ordered by specificity (more specific prefixes first).
var diagramTypeMapping = []struct {
	prefix string
	typeID string
}{
	{"stateDiagram-v2", "stateDiagram-v2"},
	{"stateDiagram", "state"},
	{"sequenceDiagram", "sequence"},
	{"classDiagram", "class"},
	{"erDiagram", "er"},
	{"C4Context", "c4Context"},
	{"C4Container", "c4Container"},
	{"C4Component", "c4Component"},
	{"C4Dynamic", "c4Dynamic"},
	{"C4Deployment", "c4Deployment"},
	{"quadrantChart", "quadrantChart"},
	{"xychart-beta", "xyChart"},
	{"sankey-beta", "sankey"},
	{"gitGraph", "gitGraph"},
	{"timeline", "timeline"},
	{"mindmap", "mindmap"},
	{"journey", "journey"},
	{"flowchart", "flowchart"},
	{"gantt", "gantt"},
	{"graph", "graph"},
	{"pie", "pie"},
}

// detectDiagramType detects the diagram type from the source.
func detectDiagramType(source string) string {
	lines := strings.SplitSeq(source, "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue // Skip empty lines and comments
		}

		// Check for diagram type keywords in order of specificity
		for _, mapping := range diagramTypeMapping {
			if strings.HasPrefix(trimmed, mapping.prefix) {
				return mapping.typeID
			}
		}

		// If we found a non-empty, non-comment line, stop looking
		break
	}

	return "unknown"
}

// isKnownDiagramType returns true if the type is a known Mermaid diagram type.
func isKnownDiagramType(diagType string) bool {
	for _, mapping := range diagramTypeMapping {
		if mapping.typeID == diagType {
			return true
		}
	}
	return false
}
