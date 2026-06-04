package mmdr

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// rssKB returns this process's resident set size in KB via ps. The leaked
// memory in a broken FFI binding lives in the C/Rust allocator, invisible to
// runtime.MemStats, so we must observe the OS-level RSS.
func rssKB(t *testing.T) int {
	t.Helper()
	out, err := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		t.Skipf("ps unavailable, skipping RSS measurement: %v", err)
	}
	kb, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		t.Skipf("could not parse ps output %q: %v", out, err)
	}
	return kb
}

// TestNoLeak renders a large number of diagrams and asserts RSS plateaus. A
// missing mmdr_free (or a double-counted allocation) would grow RSS roughly
// linearly with the iteration count.
//
// The font cache and lazy regexes allocate once, so we warm them and take the
// baseline AFTER a priming phase, then measure growth across the bulk phase.
func TestNoLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping leak test in -short mode")
	}

	const (
		prime = 2000
		bulk  = 20000
	)
	diagrams := []string{
		"flowchart LR; A-->B-->C-->D; B-->E",
		"sequenceDiagram\n U->>A: req\n A-->>U: resp",
		"flowchart TD; subgraph S\n X-->Y\n end\n Y-->Z",
		"pie title outcomes\n \"ok\": 9\n \"err\": 1",
	}
	render := func(n int) {
		for i := 0; i < n; i++ {
			if _, err := Render(diagrams[i%len(diagrams)]); err != nil {
				t.Fatalf("render failed mid-loop: %v", err)
			}
		}
	}

	render(prime)
	runtime.GC()
	baseline := rssKB(t)

	render(bulk)
	runtime.GC()
	after := rssKB(t)

	growthKB := after - baseline
	// Leaking just the ~2.5 KB SVG per iteration would add ~50 MB over the bulk
	// phase. A healthy binding stays within normal allocator slack.
	const budgetKB = 15 * 1024
	t.Logf("RSS baseline=%d KB, after %d more renders=%d KB, growth=%d KB (budget %d KB)",
		baseline, bulk, after, growthKB, budgetKB)
	if growthKB > budgetKB {
		t.Errorf("RSS grew %d KB over %d renders (budget %d KB) — likely a leak",
			growthKB, bulk, budgetKB)
	}
}
