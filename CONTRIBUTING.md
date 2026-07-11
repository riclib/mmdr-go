# Contributing to mmdr-go

This module wraps the Rust [`mermaid-rs-renderer`](https://github.com/1jehuang/mermaid-rs-renderer)
crate as a cgo binding with **prebuilt static archives**. Consumers need only a
C toolchain; the Rust toolchain is required only to *rebuild* those archives.

## Layout

```
mmdr.go, mmdr.h, link_*.go   -- the Go binding (per-platform link files)
lib/<os>_<arch>/libmmdr.a    -- committed prebuilt archives (binary)
shim/                        -- the Rust C-ABI crate that produces them
.mmdr-version                -- the single source of truth for the upstream pin
```

## Rebuilding the static archive (host platform)

You need a Rust toolchain (`rustup`, stable). Then:

```sh
make lib        # builds shim for the host triple, copies libmmdr.a into lib/
make test       # go test ./...
make test-race  # go test -race (skips the long leak test)
```

`make lib` runs `cargo build --release` in `shim/` and copies the resulting
`libmmdr.a` into the matching `lib/<os>_<arch>/` directory. It then runs
`make stamp`, and that step is not optional.

> **Go's build cache does not notice a rebuilt archive.** cgo names `libmmdr.a`
> by path in `#cgo LDFLAGS`, and Go hashes the LDFLAGS *text*, not the archive's
> contents — so the package's cache key is unchanged and `go test` will happily
> relink the **stale** archive. You get a green suite for the engine you just
> replaced. `make stamp` rewrites the digest in `archive_stamp.go`, which changes
> a source file in the package and forces the relink. If you ever build an archive
> without going through `make lib`/`make libs`, run `make stamp` (or
> `go clean -cache`) before you trust a test result.

After building, capture the native link libs for any **new** platform:

```sh
cd shim && cargo rustc --release -- --print native-static-libs
```

and put them in that platform's `link_<os>_<arch>.go` `#cgo LDFLAGS`. (On macOS,
the Go linker already provides libSystem, so only `-liconv` is added beyond
`-lmmdr`.)

## C-ABI invariants (do not break these)

The boundary in `shim/src/lib.rs` is small but load-bearing:

1. **Panic safety.** Every call into the renderer is wrapped in `catch_unwind`.
   A Rust panic must return `MMDR_PANIC`, never unwind across the FFI boundary
   (that is undefined behavior). Do not add a code path that calls upstream
   outside the `catch_unwind`.
2. **Allocator ownership.** Strings returned to Go are allocated by Rust
   (`CString::into_raw`) and must be freed by Rust (`mmdr_free` →
   `CString::from_raw`). Never free a shim pointer with C `free`, and never pass
   a C-allocated pointer to `mmdr_free`. The Go side already follows this.
3. **Out-params are pre-nulled.** `mmdr_render` sets `*out_svg`/`*out_err` to
   null before doing anything, so callers may read them unconditionally.
4. **No interior NULs escape.** Error messages are sanitized; an SVG containing
   a NUL is reported as a render error rather than truncated.

The status codes (`MMDR_OK`/`RENDER_ERROR`/`PANIC`/`BAD_INPUT`) are defined in
three places that must stay in sync: `shim/src/lib.rs`, `mmdr.h`, and the
`switch` in `mmdr.go`.

## Bumping the upstream mermaid-rs-renderer version

1. Update the pin in **three** places: `.mmdr-version`, `shim/Cargo.toml`
   (`mermaid-rs-renderer = "=x.y.z"`), and `UPSTREAM_VERSION` in
   `shim/src/lib.rs`.
2. `make lib` on each supported platform (CI does this for the release).
3. Re-run the thread-safety source audit (grep the new crate version for
   `static`/`Lazy`/`Mutex`/interior mutability) — the safety claim in the README
   is version-specific.
4. `make test-race`, then the full `make test` (including the leak test).
5. Note the new version in `CHANGELOG.md` and bump the binding's own version per
   the policy below.

## Versioning policy

Module path `github.com/riclib/mmdr-go` is stable; renaming it is a breaking
change for every consumer. For `vX.Y.Z`:

* `Z` — upstream patch bump or an internal wrapper fix.
* `Y` — upstream minor bump or new wrapper API surface.
* `X` — a breaking change to *our* Go API (independent of upstream's major).

Each release pins one upstream version and records the mapping in the changelog.

## Vendored validator (`internal/mermaidcheck/`)

Input validation (`Validate`, `RenderChecked`) is built on a vendored, modified
copy of [`mermaid-check`](https://github.com/sammcj/mermaid-check) (Apache-2.0)
under `internal/mermaidcheck/`. It's vendored — not a module dependency — because
it's on the critical path and we extend it to diagram types the renderer supports
but upstream doesn't yet validate. See `internal/mermaidcheck/PROVENANCE.md` for
the source commit, license, and the change notice required by Apache-2.0 §4.

- **Adding a diagram type:** add a parser under `internal/mermaidcheck/parser/`,
  its AST under `ast/`, and a validator under `validator/`, then wire the type
  into the `Parse` dispatch (`parser/parser.go`) and the `Validate` type switch
  (`mermaidcheck.Validate`). Mirror an existing type as a template.
- **Keep the Apache notices intact**; record local changes in git history (the
  PROVENANCE change-notice covers the bulk move + rename).

## Tests

* `go test ./...` runs the corpus, error, concurrency, and leak tests.
* `go test -race -short ./...` for the race detector without the slow leak test.
* New diagram fixtures go in `testdata/diagrams/<name>.mmd`; they are picked up
  automatically.
