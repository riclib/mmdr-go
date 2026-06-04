# Changelog

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
