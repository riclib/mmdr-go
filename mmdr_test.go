package mmdr

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// corpusDir holds one .mmd fixture per diagram type. Loaded by the tests below.
const corpusDir = "testdata/diagrams"

// requiredTypes are the diagram families validated against the production
// artifact corpus in S-710 (10 flowcharts, 2 sequence, 1 gantt). These MUST
// render; any failure is a regression. Other fixtures are exercised on a
// best-effort basis since upstream support across the 23 types varies.
var requiredTypes = map[string]bool{
	"flowchart_lr":          true,
	"flowchart_td_subgraph": true,
	"sequence":              true,
	"gantt":                 true,
}

func loadCorpus(t *testing.T) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("read corpus dir: %v", err)
	}
	out := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".mmd") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(corpusDir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		name := strings.TrimSuffix(e.Name(), ".mmd")
		out[name] = string(b)
	}
	if len(out) == 0 {
		t.Fatal("corpus is empty")
	}
	return out
}

// assertValidSVG fails if s is not a well-formed-looking SVG document.
func assertValidSVG(t *testing.T, name, s string) {
	t.Helper()
	if !strings.Contains(s, "<svg") {
		t.Errorf("%s: output has no <svg element", name)
	}
	if !strings.Contains(s, "</svg>") {
		t.Errorf("%s: output has no closing </svg>", name)
	}
}

func TestRenderBasic(t *testing.T) {
	svg, err := Render("flowchart LR; A-->B-->C")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	assertValidSVG(t, "basic", svg)
}

// TestRenderCorpus renders every fixture. Required types must succeed; the rest
// are reported so we can see the breadth of support without flaking the suite
// when upstream lacks a given diagram type.
func TestRenderCorpus(t *testing.T) {
	corpus := loadCorpus(t)

	names := make([]string, 0, len(corpus))
	for name := range corpus {
		names = append(names, name)
	}
	sort.Strings(names)

	var supported, unsupported []string
	for _, name := range names {
		svg, err := Render(corpus[name])
		switch {
		case err != nil && requiredTypes[name]:
			t.Errorf("required diagram %q failed to render: %v", name, err)
		case err != nil:
			unsupported = append(unsupported, name)
			t.Logf("optional diagram %q not rendered: %v", name, err)
		default:
			assertValidSVG(t, name, svg)
			supported = append(supported, name)
		}
	}
	t.Logf("rendered %d/%d fixtures: %s", len(supported), len(corpus), strings.Join(supported, ", "))
	if len(unsupported) > 0 {
		t.Logf("not rendered by upstream %s: %s", Version(), strings.Join(unsupported, ", "))
	}
}

func TestVersion(t *testing.T) {
	if v := Version(); v != "0.3.1" {
		t.Errorf("Version() = %q, want %q", v, "0.3.1")
	}
}
