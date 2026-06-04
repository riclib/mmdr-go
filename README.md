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

> **Status: pre-release.** Verified end-to-end on **darwin/arm64** and
> **linux/amd64** (built + full `-race` suite on glibc Ubuntu 24.04). linux/arm64
> and a tagged `v1.0.0` are next. See [Project status](#project-status).

## Why

| Option | Runtime dependency | Per-render cost | Notes |
| --- | --- | --- | --- |
| `mermaid.js` (headless browser) | Node + Chromium | ~100s of ms + spawn | heavy, but reference-quality |
| `mmdr` CLI (subprocess) | the `mmdr` binary on `PATH` | ~5–8 ms + spawn | simple, but process spawn per diagram |
| **`mmdr-go` (this library)** | **none at build/run time** | **~0.03–7 ms** | native call, no spawn; see [Performance](#performance) |

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

## Command-line demo

A small CLI under `cmd/mmdr-demo` renders a `.mmd` file (or stdin) to SVG:

```sh
go run ./cmd/mmdr-demo diagram.mmd > diagram.svg
echo 'flowchart LR; A-->B-->C' | go run ./cmd/mmdr-demo -o out.svg -t
go run ./cmd/mmdr-demo -version
```

It reads from the file argument or stdin, writes to stdout or `-o`, and exits
`2` on an invalid diagram, `3` on a renderer panic — handy for shell pipelines.

## Validating input

mmdr's renderer is **lenient**: bad Mermaid never returns an error — it silently
produces a best-effort (often wrong) diagram. There's no error return and no
error graphic, so you can't tell from the render alone that the input was wrong.

For an error signal — say, to feed back to an LLM that generated the diagram —
use `Validate`, which checks the source with a vendored, pure-Go Mermaid parser
([`mermaid-check`](https://github.com/sammcj/mermaid-check), Apache-2.0, under
`internal/mermaidcheck/`) and returns line-numbered diagnostics:

```go
res, err := mmdr.RenderChecked(src) // validate + render in one call
if err != nil { /* actual render failure */ }
if res.HasErrors() {
	for _, d := range res.Diagnostics {
		fmt.Printf("line %d: %s\n", d.Line, d.Message) // hand back to author/LLM
	}
}
use(res.SVG) // mmdr renders best-effort even on warnings; your call whether to trust it
```

Notes:
* **Validation is ~free** — 0.5–12 µs vs. a render of 30 µs–7 ms (see below), so
  `RenderChecked` ≈ `Render`.
* **Coverage is a subset.** mermaid-check validates ~16 of the diagram types mmdr
  renders. For a type it doesn't recognize, `Validate` returns a *Warning* (not
  an error) so a renderable-but-unvalidated diagram isn't flagged as invalid.
* **The two parsers can disagree.** They're independent implementations. Known
  case: mermaid-check rejects the inline-semicolon header `flowchart LR; A-->B`
  that mmdr renders fine — prefer the multi-line form when validating. Divergences
  are pinned in `validate_test.go`.

Validate-only and no cgo? Use upstream
[`mermaid-check`](https://github.com/sammcj/mermaid-check) directly — that's
exactly what it's for; this binding vendors it only to keep the render+validate
loop dependency-free and under one version.

## API

```go
// Render parses Mermaid source and returns the rendered SVG.
func Render(source string) (string, error)

// Validate checks the source with mermaid-check and returns diagnostics (nil
// when clean). It does not render.
func Validate(source string) []Diagnostic

// RenderChecked validates and renders in one call.
func RenderChecked(source string) (CheckedResult, error)

type Diagnostic struct {
	Line, Column int
	Severity     Severity // "error" | "warning" | "info"
	Message      string
}
type CheckedResult struct {
	SVG         string
	Diagnostics []Diagnostic
}
func (CheckedResult) HasErrors() bool

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
* **Go ≥ 1.24.**
* **No Rust toolchain, no `mmdr` binary, and no external Go modules.** The static
  archive is committed under `lib/<os>_<arch>/` and selected by build tag, so a
  consumer binary embeds only its own architecture's archive (~9 MB). Input
  validation is vendored (pure Go) under `internal/mermaidcheck/`, so `go.sum` is
  empty — zero third-party dependencies.

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

Measured on an M-series Mac, after the one-time font-DB / regex warm-up:

* **Most diagrams: 30–210 µs.** Sequence, class, gantt, pie, state, mindmap,
  small flowcharts without edge labels all land here.
* **Edge labels are expensive.** Each labeled flowchart edge (`A -->|text| B`)
  adds **~0.7–1.3 ms** — a 6-node flowchart with two labels is ~6.6 ms vs.
  ~200 µs for the same flowchart unlabeled. An upstream rendering cost, not the
  binding's. If you render many labeled flowcharts, budget for it.
* Layout cost is also **superlinear in edge count** — a flowchart with thousands
  of edges can take seconds.
* Still no browser, no Node, no subprocess — even the 6.6 ms case beats
  mermaid.js comfortably.
* No memory leak: 20,000 renders grow RSS < 1 MB.

(Validation via `Validate`/`RenderChecked` is 0.5–12 µs — negligible next to any
render.)

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

| Platform | State |
| --- | --- |
| darwin/arm64 | ✅ built + full `-race` suite (Apple Silicon) |
| darwin/amd64 | ✅ built + tests (via Rosetta) |
| linux/amd64 (glibc) | ✅ built + full `-race` suite — Ubuntu 24.04 **and RHEL** |
| linux/arm64 (glibc) | ✅ cross-built; hardware verification pending |

The static archives are **cross-built locally** (a `staticlib` needs no cross
linker) via `make libs` and committed — no CI service required. Each Linux
archive uses the `*-gnu` triple, whose old glibc baseline links into any modern
distro (Ubuntu, Debian, **RHEL 8/9**); build your binary on the target as usual.
`v1.0.0` adds `RenderWithOptions` and a tagged release.

See [`docs/poc-findings.md`](docs/poc-findings.md) for the verification writeup.

## Acknowledgements

This library is a thin wrapper. All the rendering work is done by
[`1jehuang/mermaid-rs-renderer`](https://github.com/1jehuang/mermaid-rs-renderer)
(MIT) — please star the upstream project. Input validation is built on
[`sammcj/mermaid-check`](https://github.com/sammcj/mermaid-check) (Apache-2.0), a
pure-Go Mermaid parser/validator, vendored under `internal/mermaidcheck/` (see
its `PROVENANCE.md`). Thanks to both authors.

## License

[MIT](LICENSE) © 2026 Ricardo Liberato, except `internal/mermaidcheck/`, which is
a vendored copy of `mermaid-check` under the Apache-2.0 license (see
`internal/mermaidcheck/LICENSE`).
