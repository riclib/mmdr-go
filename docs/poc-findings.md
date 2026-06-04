# PoC findings — cgo binding + thread-safety verification

**Milestone:** PoC: cgo binding + thread safety verification (darwin-arm64)
**Platform verified:** darwin/arm64 (Apple Silicon), Go 1.26, Rust 1.96
**Upstream pinned:** `mermaid-rs-renderer = "=0.2.2"` (MIT)
**Outcome:** ✅ all definition-of-done criteria met.

## Definition of done — results

| Criterion | Result |
| --- | --- |
| Working `Render()` on darwin-arm64 | ✅ renders valid SVG |
| Output valid across diagram corpus | ✅ 13/13 fixtures render (see `testdata/diagrams/`) |
| Documented thread-safety conclusion, `-race` backed | ✅ **safe to call concurrently** (below) |
| No memory leak after bulk renders | ✅ RSS grew < 1 MB over 20,000 renders |
| No crash on adversarial inputs | ✅ 12 hostile inputs return cleanly |

## Architecture as built

```
mmdr.go                 -- public API: Render, Version, RenderError + sentinels
mmdr.h                  -- C declarations (status codes + 3 functions)
link_darwin_arm64.go    -- //go:build darwin && arm64; cgo LDFLAGS
lib/darwin_arm64/libmmdr.a   -- 9.3 MB prebuilt static archive
shim/                   -- Rust C-ABI crate (built in CI, not by consumers)
  Cargo.toml            -- crate-type=["staticlib"], pins mermaid-rs-renderer
  src/lib.rs            -- mmdr_render / mmdr_free / mmdr_version + catch_unwind
testdata/diagrams/*.mmd -- one fixture per diagram type
```

### ABI choice: richer than the original sketch

The project sketch proposed `mmdr_render(input) -> *mut c_char` returning null on
failure. That cannot distinguish a parse error from a panic, which the `v1.0.0`
API (`ErrInvalidInput` vs `ErrPanic`) requires. The shim instead returns a
status code with out-params:

```c
int mmdr_render(const char *input, char **out_svg, char **out_err);
```

`0=ok, 1=render error, 2=panic, 3=bad input`. This carries the renderer's error
message back to Go and maps cleanly onto the sentinel errors. Forward-compatible
with `v1.0.0`.

## Thread safety — the load-bearing question

**Conclusion: `Render` is safe to call concurrently. No Go-side mutex is needed.**

Verified two ways:

1. **Source audit of `mermaid-rs-renderer` 0.2.2.** The only mutable global is:

   ```rust
   static TEXT_MEASURER: Lazy<Mutex<TextMeasurer>> =
       Lazy::new(|| Mutex::new(TextMeasurer::new()));
   ```

   Every access is `TEXT_MEASURER.lock()`, so the internal `HashMap` font cache
   is mutated only under the mutex — data-race-free. Every other `static` in the
   crate is an immutable `Lazy<Regex>` (parser regexes); `Regex` is `Sync` and
   they are read-only after one-time init. There is **no `static mut`** and no
   unsynchronized interior mutability anywhere in the crate's 45 source files.

2. **`go test -race`** with 64 goroutines × 200 renders (12,800 total) — clean,
   zero failures.

**Caveat (throughput, not correctness):** concurrent renders serialize on the
text-measurement mutex. Fine for typical workloads; a future optimization could
make the cache per-thread or lock-free, ideally upstream.

## Memory

* `into_raw` / `from_raw` ownership: the shim hands the SVG to Go via
  `CString::into_raw`; Go returns it with `mmdr_free` (`CString::from_raw`).
  Input strings are Go-allocated (`C.CString`) and freed with `C.free`. The two
  allocators never cross.
* **Leak test:** prime 2,000 renders → baseline RSS; then 20,000 more →
  growth **912 KB** (budget 15 MB). No per-render leak.

## Adversarial inputs

All of these return a typed error or best-effort SVG, none crash the process:
empty, whitespace-only, garbage, truncated flowchart, unterminated subgraph,
unknown diagram type, **NUL byte mid-string**, 800-edge graph, **1 MB node
label**, multi-byte Unicode/emoji, arrow-only, control characters.

`catch_unwind` in the shim means a Rust panic becomes `MMDR_PANIC` → `ErrPanic`,
never an unwind across FFI. (No input in the corpus actually triggered a panic;
the path is exercised structurally and unit-tested on the Go side.)

## Surprises that affect the v1.0.0 milestone

1. **Per-render cost is bimodal.** Most diagrams are 30–210 µs (better than the
   3–5 ms estimate), BUT flowcharts with **edge labels** (`A -->|text| B`) cost
   ~0.7–1.3 ms *per labeled edge* — a 6-node, 2-label flowchart is ~6.6 ms.
   (An earlier note here claimed a flat ~0.25 ms; that was a label-free diagram.)
   Upstream rendering characteristic, not the binding. Candidate for the v1.0.0
   upstream coordination issue alongside the superlinear-edge-count note.
2. **Layout is superlinear in edge count.** A 5,000-edge flowchart took ~74 s.
   Not a bug (it's the upstream layout algorithm), but: (a) document it,
   (b) keep stress fixtures modest, (c) a consumer rendering user-supplied
   diagrams should consider a size cap / timeout. Worth a note to upstream.
3. **All 13 fixtures across diverse types rendered on 0.2.2** — broader than the
   "flowchart/sequence/gantt" production mix from S-710. The `v1.0.0` snapshot
   suite can safely cover the wider set.
4. **macOS link libs:** `cargo rustc -- --print native-static-libs` reports
   `-liconv -lSystem -lc -lm`; on macOS the Go linker already pulls libSystem
   (libc/libm), so only `-liconv` must be added — listing the rest triggers
   "ignoring duplicate libraries". For the Linux musl targets in `v1.0.0`,
   re-run that command per target to get the right set.
5. **`opt-level="s"` + LTO** keeps the archive at 9.3 MB (vs the ~12 MB
   estimate).

## Out of scope for this PoC (→ v1.0.0)

Linux musl builds, CI matrix + reproducible builds, `RenderWithOptions`/`Options`/
`Result`, snapshot/golden tests, README polish for release, upstream
coordination issue, and the Solid v5 consumer swap.
