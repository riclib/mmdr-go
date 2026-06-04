# Changelog

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
- linux/arm64 (glibc) — cross-built; hardware verification pending.

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
- linux/arm64 hardware verification; upstream coordination issue.
