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
# Usage: shim/build-libs.sh [target-key ...]   (default: all)
set -euo pipefail
cd "$(dirname "$0")"

# target-key  ->  "rust-triple goos_goarch"
declare -A TARGETS=(
  [darwin-arm64]="aarch64-apple-darwin darwin_arm64"
  [darwin-amd64]="x86_64-apple-darwin darwin_amd64"
  [linux-amd64]="x86_64-unknown-linux-gnu linux_amd64"
  [linux-arm64]="aarch64-unknown-linux-gnu linux_arm64"
)

keys=("$@")
[ ${#keys[@]} -eq 0 ] && keys=(darwin-arm64 linux-amd64 linux-arm64)

for key in "${keys[@]}"; do
  entry="${TARGETS[$key]:?unknown target '$key'}"
  triple="${entry%% *}"; libdir="${entry##* }"
  echo "==> $key ($triple)"
  rustup target add "$triple" >/dev/null
  cargo build --release --target "$triple"
  mkdir -p "../lib/$libdir"
  cp "target/$triple/release/libmmdr.a" "../lib/$libdir/libmmdr.a"
  echo "    -> lib/$libdir/libmmdr.a"
done
echo "done. Remember to run 'cargo rustc --release --target <triple> -- --print native-static-libs'"
echo "for any NEW target and put the result in its link_<goos>_<goarch>.go LDFLAGS."
