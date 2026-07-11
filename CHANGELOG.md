# Changelog

## v0.4.0 — 2026-07-11

Upstream bump: `mermaid-rs-renderer` **0.2.2 → 0.3.1**. All four prebuilt archives
rebuilt. The C ABI and the Go API are unchanged.

### Changed
- **Every diagram's rendered geometry changes.** Upstream 0.3.0 overhauled
  subgraph containment and edge routing: sibling subgraph membership now matches
  mermaid-js (`flowDb.makeUniq`), edges detour around subgraph boxes they don't
  belong to, and non-member nodes that visually landed inside a subgraph box are
  evicted. Real fix, visibly better output — in `testdata/flowchart_td_subgraph`
  the `Query` node no longer overlaps the `Store` box — but **any golden-SVG
  snapshot downstream will churn**. 63 of our 91 render fixtures also change
  dimensions. Also in the bump: architecture-beta port routing, C4 connector
  routing, gantt/sequence/block-beta parse fixes, pie legend and CJK title fixes.

### Fixed
- Rendering stayed **lenient**, which took deliberate work. As of upstream 0.3.0,
  `render`/`render_with_options` route through the new `parse_mermaid_strict`,
  which runs a preflight validator and hard-fails the render. That would have
  broken this library twice over: it contradicts mmdr-go's documented contract
  (flawed input yields a best-effort SVG; `Validate` is the error channel, and
  `RenderChecked` depends on the render still happening), and the validator has a
  false positive — it only recognizes a `subgraph` opener when the trimmed line
  *starts with* `subgraph`, so an opener after an inline `;` separator
  (`flowchart TD; subgraph S` … `end`) is missed and the matching `end` is then
  reported as unbalanced, rejecting **valid** Mermaid that 0.2.2 renders fine.
  The shim now composes the lenient pipeline itself from still-public upstream
  parts (`parse_mermaid` → `compute_layout` → `render_svg`) — exactly what
  upstream's own `render_with_options` did through 0.2.2 — so we get every layout
  improvement without the strict path.

### Notes
- One narrow behavior change survives, on genuinely invalid input: upstream's
  *parser* (not the preflight validator) now rejects a line starting with a bare
  arrow (`--> B`), where 0.2.2 rendered it best-effort. It returns
  `ErrInvalidInput` rather than a wrong diagram, which is the better outcome.
- Themes are unchanged. Upstream gained its own `dark`/`forest`/`neutral` presets
  in 0.3.0; mmdr-go keeps its own same-named palettes so an upgrade never silently
  recolors an existing diagram. `"neutral"` still maps to the classic mermaid
  palette.
- `Result.Width/Height` still reports `0x0` for `pie`, `mindmap`, and `gitgraph`
  (they emit `width="100%"` and no numeric height, and dimensions are parsed from
  the SVG root). Pre-existing, not caused by this bump. Upstream 0.3.0 added
  `measure_svg_dimensions`, which returns exact dimensions from the layout and
  would fix this properly — it needs a new C-ABI entry point.
- Maintainer gotcha: Go's build cache does not re-hash the contents of
  `lib/*/libmmdr.a`, so after rebuilding an archive it will happily link the stale
  one. Run `go clean -cache` before testing, or you will validate the old engine.

## v0.3.0 — 2026-06-04

### Added
- Two first-party named themes tuned to the Solid/V4 app's design tokens:
  `"solid-light"` and `"solid-dark"`. Both render with a TRANSPARENT background
  (the root canvas `<rect>` is `fill="none"`) so diagrams blend into the host
  surface. Synthesized from `Theme::modern()` in the shim; the compiled archives
  for all four platforms were rebuilt (the C ABI is unchanged).

## v0.2.0 — 2026-06-04

### Added
- `RenderWithOptions(source, Options) (Result, error)` — render with a theme,
  fast-text mode, and/or a preferred aspect ratio; `Result` reports the rendered
  SVG dimensions. Backed by a new `mmdr_render_with_options` C-ABI entry point;
  all four prebuilt archives rebuilt and re-verified (darwin arm64/amd64, linux
  amd64/arm64 via OrbStack).

### Notes
- Themes: upstream ships only `default`/`modern` and `neutral`. `"dark"` and
  `"forest"` are mmdr-go's own palettes, not mermaid.js's same-named themes.
- `Width`/`Height` act as a preferred aspect ratio (the SVG path has no
  fixed-pixel sizing upstream); `Result.Width/Height` report the real size.

## v0.1.0 — 2026-06-04

First public pre-release. Wraps `mermaid-rs-renderer` 0.2.2.

### Added
- `Render(source) (string, error)` — native Mermaid → SVG via a cgo binding over
  the Rust `mermaid-rs-renderer` engine. No browser, Node, or subprocess.
- `Validate(source) []Diagnostic` and `RenderChecked(source) (CheckedResult, error)`
  — line-numbered input validation, the error signal the renderer itself can't
  give. Built on a vendored, pure-Go copy of `mermaid-check` (Apache-2.0) under
  `internal/mermaidcheck/`.
- `Version()` — the upstream renderer version.
- `cmd/mmdr-demo` — a small CLI (file/stdin → SVG).
- `shim/build-libs.sh` / `make libs` — cross-build every platform's static
  archive from one host (no CI service, no cross C toolchain).

### Platforms
Prebuilt static archives committed under `lib/`, selected by build tag:
- darwin/arm64, darwin/amd64 — verified.
- linux/amd64 (glibc) — verified on Ubuntu 24.04 and RHEL.
- linux/arm64 (glibc) — verified on native aarch64 Linux.

### Notes
- **Zero external Go module dependencies** (`go.sum` is empty).
- Requires `CGO_ENABLED=1` and **Go ≥ 1.24**.
- The renderer is lenient: invalid Mermaid returns a best-effort SVG, not an
  error — use `Validate` for diagnostics. One known parser divergence is pinned
  in tests (the inline-semicolon header `flowchart LR; A-->B`).
- Render performance is bimodal: ~30–210 µs typical, but edge-labeled flowcharts
  cost ~0.7–1.3 ms per labeled edge.

### Not yet (planned for v1.0.0)
- `RenderWithOptions` (themes, dimensions).
- Upstream coordination issue.
