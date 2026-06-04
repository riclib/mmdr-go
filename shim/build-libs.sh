#!/usr/bin/env bash
# Build the prebuilt static archives for all supported platforms and copy them
# into ../lib/<goos>_<goarch>/libmmdr.a.
#
# This replaces a CI matrix: a `staticlib` has no final link step, so every
# target cross-compiles from one host (e.g. macOS) with just `rustup target add`
# — no cross C toolchain required. Run this on a machine with Rust installed,
# commit the resulting archives, and consumers need only a C compiler.
#
# Linux targets use the *-gnu* triples (not musl): Rust's gnu std references an
# old glibc baseline, so the archive links into any modern-glibc Go cgo binary
# (Ubuntu, Debian, RHEL 8/9, ...). musl would add an -lunwind dependency that
# isn't present by default on most distros.
#
# Usage: shim/build-libs.sh [target-key ...]   (default: the 3 core targets)
# Keys: darwin-arm64 darwin-amd64 linux-amd64 linux-arm64
#
# Kept POSIX/bash-3.2 friendly (macOS ships bash 3.2 — no associative arrays).
set -euo pipefail
cd "$(dirname "$0")"

triple_for() {
  case "$1" in
    darwin-arm64) echo "aarch64-apple-darwin" ;;
    darwin-amd64) echo "x86_64-apple-darwin" ;;
    linux-amd64)  echo "x86_64-unknown-linux-gnu" ;;
    linux-arm64)  echo "aarch64-unknown-linux-gnu" ;;
    *) echo "unknown target '$1'" >&2; exit 1 ;;
  esac
}
libdir_for() {
  case "$1" in
    darwin-arm64) echo "darwin_arm64" ;;
    darwin-amd64) echo "darwin_amd64" ;;
    linux-amd64)  echo "linux_amd64" ;;
    linux-arm64)  echo "linux_arm64" ;;
  esac
}

keys=("$@")
[ ${#keys[@]} -eq 0 ] && keys=(darwin-arm64 linux-amd64 linux-arm64)

for key in "${keys[@]}"; do
  triple="$(triple_for "$key")"
  libdir="$(libdir_for "$key")"
  echo "==> $key ($triple)"
  rustup target add "$triple" >/dev/null
  cargo build --release --target "$triple"
  mkdir -p "../lib/$libdir"
  cp "target/$triple/release/libmmdr.a" "../lib/$libdir/libmmdr.a"
  echo "    -> lib/$libdir/libmmdr.a"
done
echo "done. For a NEW target, also run:"
echo "  cargo rustc --release --target <triple> -- --print native-static-libs"
echo "and put the result in its link_<goos>_<goarch>.go LDFLAGS."
