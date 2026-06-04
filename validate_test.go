package mmdr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSwallowedCases(t *testing.T) {
	// These are the inputs mmdr's renderer accepts silently (no error, a
	// best-effort SVG). Validate should surface a signal for each.
	t.Run("unknown_type_is_warning", func(t *testing.T) {
		for _, src := range []string{
			"totally not mermaid @#$",    // garbage -> unknown type
			"seqenceDiagram\n A->>B: hi", // typo'd diagram type
		} {
			ds := Validate(src)
			if len(ds) == 0 {
				t.Fatalf("expected a diagnostic for %q", src)
			}
			if ds[0].Severity != SeverityWarning {
				t.Errorf("%q: got %s, want warning (unrecognized type must not read as error)", src, ds[0].Severity)
			}
		}
	})

	t.Run("structural_error_has_line", func(t *testing.T) {
		ds := Validate("flowchart TD\n subgraph X\n A-->B") // unclosed subgraph on line 3
		if len(ds) == 0 {
			t.Fatal("expected a diagnostic for unclosed subgraph")
		}
		var found bool
		for _, d := range ds {
			if d.Severity == SeverityError && d.Line == 3 {
				found = true
			}
		}
		if !found {
			t.Errorf("expected an error diagnostic on line 3, got %v", ds)
		}
	})
}

// TestValidateCleanCorpus asserts Validate raises no error-severity diagnostics
// for any diagram mmdr renders — i.e. the validator does not falsely reject
// renderable input.
func TestValidateCleanCorpus(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(corpusDir, "*.mmd"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no corpus files: %v", err)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		name := filepath.Base(f)
		for _, d := range Validate(string(b)) {
			if d.Severity == SeverityError {
				t.Errorf("%s: unexpected error diagnostic: %s", name, d)
			}
		}
	}
}

// TestKnownDivergence pins cases where mermaid-check and mmdr disagree, so we
// notice if an upstream bump changes the behavior. These are mermaid-check
// limitations (valid mermaid that mmdr renders but the validator rejects), not
// mmdr bugs. Tracked for an upstream report.
func TestKnownDivergence(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// Inline-semicolon header — common in mermaid docs (`graph TD; A-->B;`),
		// rendered by mmdr, but mermaid-check wants the header on its own line.
		{"inline_semicolon_header", "flowchart LR; A-->B-->C"},
	}
	for _, tc := range cases {
		ds := Validate(tc.src)
		hasErr := false
		for _, d := range ds {
			if d.Severity == SeverityError {
				hasErr = true
			}
		}
		if !hasErr {
			t.Errorf("%s: expected the known false-positive to still occur — if it's gone, "+
				"mermaid-check fixed it; remove this case. (%q)", tc.name, tc.src)
		}
		// Confirm mmdr renders it regardless, i.e. the diagram really is valid.
		if svg, err := Render(tc.src); err != nil || !strings.Contains(svg, "<svg") {
			t.Errorf("%s: mmdr should render this valid diagram (err=%v)", tc.name, err)
		}
	}
}

func TestRenderChecked(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		// Multi-line form (header on its own line) — accepted by both engines.
		// NOTE: the inline-semicolon form "flowchart LR; A-->B" is rejected by
		// mermaid-check v0.0.4 though mmdr renders it; see TestKnownDivergence.
		res, err := RenderChecked("flowchart LR\n A-->B-->C")
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		if !strings.Contains(res.SVG, "<svg") {
			t.Error("expected SVG output")
		}
		if res.HasErrors() {
			t.Errorf("clean diagram reported errors: %v", res.Diagnostics)
		}
	})

	t.Run("flawed_still_renders_but_flags", func(t *testing.T) {
		res, err := RenderChecked("flowchart TD\n subgraph X\n A-->B")
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		// mmdr is lenient — it renders something even though the source is broken.
		if !strings.Contains(res.SVG, "<svg") {
			t.Error("expected best-effort SVG even for flawed input")
		}
		if !res.HasErrors() {
			t.Errorf("flawed diagram should report an error, got %v", res.Diagnostics)
		}
	})
}
