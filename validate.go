package mmdr

import (
	"fmt"
	"strings"

	mermaidcheck "github.com/riclib/mmdr-go/internal/mermaidcheck"
	"github.com/riclib/mmdr-go/internal/mermaidcheck/validator"
)

// Severity classifies a Diagnostic.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Diagnostic is a single issue found by Validate. Positions are 1-indexed; Line
// is 0 when the underlying error carries no position. The fields are flat and
// JSON-friendly so they can be surfaced to a human or fed back to an LLM.
type Diagnostic struct {
	Line     int
	Column   int
	Severity Severity
	Message  string
}

func (d Diagnostic) String() string {
	if d.Line > 0 {
		return fmt.Sprintf("line %d: %s: %s", d.Line, d.Severity, d.Message)
	}
	return fmt.Sprintf("%s: %s", d.Severity, d.Message)
}

// Validate parses and semantically checks Mermaid source using the pure-Go
// mermaid-check parser. It returns any diagnostics found (nil when clean) and
// does NOT render — pair it with Render, or use RenderChecked to do both.
//
// This is the error signal the renderer itself cannot give you: mmdr's engine
// is lenient and never reports syntax errors (bad input becomes a best-effort
// diagram), whereas mermaid-check builds a real AST and reports problems with
// line/column.
//
// Coverage caveat: mermaid-check validates a subset of the diagram types mmdr
// can render. For a type it does not recognize, Validate returns a single
// Warning (not an Error) noting that validation was skipped — so a
// renderable-but-unvalidated diagram is never mistaken for an invalid one.
func Validate(source string) []Diagnostic {
	diagram, err := mermaidcheck.Parse(source)
	if err != nil {
		msg := err.Error()
		sev := SeverityError
		// A diagram type mmdr may support but mermaid-check doesn't: this is not
		// an error — mmdr can still render it; we simply can't validate it.
		if strings.Contains(msg, "unknown or unsupported diagram type") {
			sev = SeverityWarning
		}
		return []Diagnostic{{Line: lineFromError(msg), Severity: sev, Message: msg}}
	}

	verrs := mermaidcheck.Validate(diagram, false)
	if len(verrs) == 0 {
		return nil
	}
	out := make([]Diagnostic, len(verrs))
	for i, v := range verrs {
		out[i] = Diagnostic{
			Line:     v.Line,
			Column:   v.Column,
			Severity: mapSeverity(v.Severity),
			Message:  v.Message,
		}
	}
	return out
}

// CheckedResult bundles a render with the diagnostics from Validate.
type CheckedResult struct {
	SVG         string
	Diagnostics []Diagnostic
}

// HasErrors reports whether any diagnostic is error-severity. Warnings and info
// (including the "validation skipped for this diagram type" warning) do not
// count, so HasErrors is a safe gate for an automated render/feedback loop.
func (r CheckedResult) HasErrors() bool {
	for _, d := range r.Diagnostics {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

// RenderChecked validates and renders in one call, returning both the SVG and
// any diagnostics. It always attempts the render — mmdr usually produces a
// best-effort diagram even for flawed input — so inspect Diagnostics (or
// HasErrors) to decide whether to trust the SVG or feed the messages back to
// the author/LLM. The error return is reserved for an actual render failure
// (see Render), not for validation findings.
func RenderChecked(source string) (CheckedResult, error) {
	diags := Validate(source)
	svg, err := Render(source)
	return CheckedResult{SVG: svg, Diagnostics: diags}, err
}

func mapSeverity(s validator.Severity) Severity {
	switch s {
	case validator.SeverityWarning:
		return SeverityWarning
	case validator.SeverityInfo:
		return SeverityInfo
	default:
		return SeverityError
	}
}

// lineFromError best-effort extracts a leading line number from a parse error
// message such as "line 3: unclosed subgraph". Returns 0 if absent.
func lineFromError(msg string) int {
	const p = "line "
	i := strings.Index(msg, p)
	if i < 0 {
		return 0
	}
	n, found := 0, false
	for _, r := range msg[i+len(p):] {
		if r < '0' || r > '9' {
			break
		}
		n, found = n*10+int(r-'0'), true
	}
	if !found {
		return 0
	}
	return n
}
