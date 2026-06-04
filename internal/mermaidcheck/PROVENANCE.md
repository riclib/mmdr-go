# Vendored: mermaid-check

This directory is a vendored, modified copy of
**[github.com/sammcj/mermaid-check](https://github.com/sammcj/mermaid-check)**
at **v0.0.4** (commit `605c28f`), a pure-Go Mermaid parser/validator.

Copyright 2025 Sam McLeod. Licensed under the **Apache License 2.0** — see
[`LICENSE`](LICENSE) in this directory. The rest of `mmdr-go` is MIT; this
subtree retains its original Apache-2.0 license.

## Why vendored

`mmdr-go` uses this code on its critical input-validation path. To keep that
path under our control (and to extend it to diagram types the renderer supports
but upstream does not yet validate), the source is vendored here rather than
taken as an external module dependency.

## Modifications from upstream (Apache-2.0 §4(2) change notice)

- Import paths rewritten from `github.com/sammcj/mermaid-check/...` to
  `github.com/riclib/mmdr-go/internal/mermaidcheck/...`.
- Root facade package renamed from `mermaid` to `mermaidcheck`.
- Upstream tests and the `cmd/` CLI (and its `fatih/color` dependency) omitted.
- Subsequent local changes (e.g. added diagram-type parsers/validators) are
  tracked in this repo's git history.

To diff against upstream, compare this tree with `mermaid-check` at the commit
above.
