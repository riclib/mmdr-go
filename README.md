# mmdr-go

Native [Mermaid](https://mermaid.js.org/) diagram rendering for Go, with **no
browser, no Node, and no subprocess**. It links the Rust
[`mermaid-rs-renderer`](https://github.com/1jehuang/mermaid-rs-renderer) engine
in as a static library via cgo and exposes a one-line API:

```go
svg, err := mmdr.Render("flowchart LR; A-->B-->C")
```

```go
import "github.com/riclib/mmdr-go"
```

> **Status: proof-of-concept.** The binding is verified end-to-end on
> **darwin/arm64** only. Linux (amd64/arm64) and a tagged `v1.0.0` release are
> the next milestone. See [Project status](#project-status).

## Why

| Option | Runtime dependency | Per-render cost | Notes |
| --- | --- | --- | --- |
| `mermaid.js` (headless browser) | Node + Chromium | ~100s of ms + spawn | heavy, but reference-quality |
| `mmdr` CLI (subprocess) | the `mmdr` binary on `PATH` | ~5–8 ms + spawn | simple, but process spawn per diagram |
| **`mmdr-go` (this library)** | **none at build/run time** | **~0.25 ms** | native call, no spawn, batch/stream friendly |

If you already build with cgo (e.g. you use `mattn/go-sqlite3` or
`marcboeker/go-duckdb`), this adds no new toolchain tax — the Rust engine ships
**prebuilt** in this module.

## Quick start

```go
package main

import (
	"fmt"
	"os"

	mmdr "github.com/riclib/mmdr-go"
)

func main() {
	svg, err := mmdr.Render(`flowchart LR
	    A[Start] --> B{OK?}
	    B -->|yes| C[Ship]
	    B -->|no| D[Fix]`)
	if err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		os.Exit(1)
	}
	os.WriteFile("diagram.svg", []byte(svg), 0o644)
}
```

## API

```go
// Render parses Mermaid source and returns the rendered SVG.
func Render(source string) (string, error)

// Version returns the upstream mermaid-rs-renderer version, e.g. "0.2.2".
func Version() string

// Sentinel errors, matchable with errors.Is.
var (
	ErrInvalidInput = errors.New("mmdr: invalid mermaid source")
	ErrPanic        = errors.New("mmdr: rust panic during render")
)
```

Errors returned by `Render` are `*RenderError`, which wraps one of the sentinels
and carries the renderer's message:

```go
svg, err := mmdr.Render(src)
if errors.Is(err, mmdr.ErrInvalidInput) {
	// bad diagram source — surface to the user
}
```

> `RenderWithOptions(source, Options)` (themes, fast-text mode, fixed
> dimensions) is part of the `v1.0.0` API and not yet implemented in this PoC.

## Build requirements

* **`CGO_ENABLED=1`** (the default on the supported platforms).
* A C toolchain (clang/gcc) — already present on a normal dev box.
* **No Rust toolchain and no `mmdr` binary are required.** The static archive is
  committed under `lib/<os>_<arch>/` and selected by build tag, so a consumer
  binary embeds only its own architecture's archive (~9 MB).

## Thread safety

**`Render` is safe to call concurrently from multiple goroutines.** This is not
just an empirical observation — it follows from the upstream source
(`mermaid-rs-renderer` 0.2.2):

* The only mutable global is the text-measurement font cache,
  `static TEXT_MEASURER: Lazy<Mutex<TextMeasurer>>`. Every access goes through
  `.lock()`, so cache mutation is mutually excluded — data-race-free, merely
  *serialized* during text measurement.
* Every other `static` is an immutable `Lazy<Regex>` (parser regexes). `Regex`
  is `Sync`; these are read-only after one-time initialization.
* There is no `static mut` and no unsynchronized interior mutability anywhere in
  the crate.

A concurrency test drives `Render` from 64 goroutines under `go test -race`
(12,800 renders) with no failures. The one caveat is throughput, not
correctness: heavy concurrent rendering contends on the text-measurement mutex.

## Performance

* **~0.25 ms** per typical small diagram on an M-series Mac (after the
  one-time font-DB / regex warm-up on the first call).
* Layout cost is **superlinear in edge count** — a single flowchart with
  thousands of edges can take seconds. This is an upstream layout
  characteristic; budget accordingly for pathologically large diagrams.
* No memory leak: 20,000 renders grow RSS < 1 MB.

## Supported diagrams

This PoC renders all of the following with `mermaid-rs-renderer` 0.2.2
(see `testdata/diagrams/`): flowchart, sequence, gantt, class, state, ER, pie,
journey, mindmap, gitGraph, timeline, quadrant. Upstream advertises 23 diagram
types; layout quality "may not match mermaid-cli in all cases" — see the
[upstream README](https://github.com/1jehuang/mermaid-rs-renderer).

## Memory model (for the curious)

The Rust shim allocates each returned string with Rust's allocator; the Go
wrapper releases it with `mmdr_free` (never C `free`). Render calls are wrapped
in `catch_unwind`, so a panic in the renderer becomes an `ErrPanic` return
rather than unwinding across the FFI boundary (which would be undefined
behavior). Details and invariants live in
[`CONTRIBUTING.md`](CONTRIBUTING.md) and the shim source (`shim/src/lib.rs`).

## Project status

| Milestone | State |
| --- | --- |
| PoC: cgo binding + thread-safety verification (darwin/arm64) | ✅ done |
| v1.0.0: Linux musl builds, CI matrix, `RenderWithOptions`, tagged release | ⏳ next |

See [`docs/poc-findings.md`](docs/poc-findings.md) for the full PoC writeup.

## Acknowledgements

This library is a thin wrapper. All the rendering work is done by
[`1jehuang/mermaid-rs-renderer`](https://github.com/1jehuang/mermaid-rs-renderer)
— please star the upstream project. Both it and this binding are MIT licensed.

## License

[MIT](LICENSE) © 2026 Ricardo Liberato.
