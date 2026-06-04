package mmdr

import (
	"errors"
	"strings"
	"testing"
)

// TestAdversarialInputs feeds malformed and hostile inputs and asserts the
// binding returns cleanly (error or best-effort SVG) without crashing the Go
// process. The test itself surviving to completion is the assertion: a panic
// unwinding across FFI, a segfault, or an abort would take the test binary down.
func TestAdversarialInputs(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"whitespace", "   \n\t  "},
		{"garbage", "this is definitely not mermaid syntax !!! @#$"},
		{"truncated_flowchart", "flowchart LR\n    A -->"},
		{"unterminated_subgraph", "flowchart TD\n  subgraph X\n    A-->B"},
		{"unknown_type", "wobbleDiagram\n  foo --> bar"},
		{"nul_midstring", "flowchart LR; A\x00-->B"},
		// Layout cost is superlinear in edge count (see README perf notes), so
		// keep this modest — it proves robustness, not throughput.
		{"deep_nesting", "flowchart LR;" + strings.Repeat(" A-->B;", 800)},
		{"huge_label", "flowchart LR; A[" + strings.Repeat("x", 1<<20) + "]-->B"},
		{"unicode", "flowchart LR; 日本語-->🚀-->Ω"},
		{"only_arrows", "------>>>>----"},
		{"control_chars", "flowchart LR; A[\x01\x02\x03]-->B"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svg, err := Render(tc.in)
			// Either outcome is acceptable; we only require no crash and, on
			// error, a typed sentinel that callers can match.
			if err != nil {
				if !errors.Is(err, ErrInvalidInput) && !errors.Is(err, ErrPanic) {
					t.Errorf("error %v wraps neither ErrInvalidInput nor ErrPanic", err)
				}
				return
			}
			if !strings.Contains(svg, "<svg") {
				t.Errorf("non-error result is not SVG: %.80q", svg)
			}
		})
	}
}

// TestErrorIsMatching documents the errors.Is contract callers rely on.
func TestErrorIsMatching(t *testing.T) {
	re := &RenderError{sentinel: ErrInvalidInput, Detail: "boom"}
	if !errors.Is(re, ErrInvalidInput) {
		t.Error("RenderError should match its sentinel via errors.Is")
	}
	if errors.Is(re, ErrPanic) {
		t.Error("RenderError matched the wrong sentinel")
	}
	if !strings.Contains(re.Error(), "boom") {
		t.Errorf("Error() should include detail, got %q", re.Error())
	}
}
