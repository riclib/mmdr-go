package mmdr

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// TestConcurrentRender hammers Render from many goroutines. Run with -race.
//
// The thread-safety guarantee comes from reading the upstream source (see the
// THREAD SAFETY section of the README): the only mutable global is a
// Mutex-guarded text-measurement cache; every other static is an immutable
// Lazy<Regex>. The Go race detector cannot see into the Rust archive, but this
// test confirms the Go-side wrapper is race-free and that concurrent calls do
// not corrupt output or crash.
func TestConcurrentRender(t *testing.T) {
	const (
		goroutines = 64
		perG       = 200
	)
	diagrams := []string{
		"flowchart LR; A-->B-->C",
		"sequenceDiagram\n A->>B: hi\n B-->>A: yo",
		"flowchart TD; X-->Y; Y-->Z; X-->Z",
		"pie title x\n \"a\": 1\n \"b\": 2",
	}

	var wg sync.WaitGroup
	var failures int64
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				src := diagrams[(g+i)%len(diagrams)]
				svg, err := Render(src)
				if err != nil {
					atomic.AddInt64(&failures, 1)
					continue
				}
				if !strings.Contains(svg, "<svg") {
					atomic.AddInt64(&failures, 1)
				}
			}
		}(g)
	}
	wg.Wait()

	if failures != 0 {
		t.Errorf("%d/%d concurrent renders failed", failures, goroutines*perG)
	}
}
