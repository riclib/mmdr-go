# mmdr-go — rebuild the prebuilt static archive and run tests.
#
# Consumers do NOT need this Makefile or a Rust toolchain; the archives are
# committed under lib/. This is for maintainers rebuilding them.

# Map the host to a lib/<os>_<arch> directory.
GOOS  := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
LIBDIR := lib/$(GOOS)_$(GOARCH)

# sha256sum on Linux, shasum on macOS.
SHA256 := $(shell command -v sha256sum >/dev/null 2>&1 && echo 'sha256sum' || echo 'shasum -a 256')

.PHONY: lib libs stamp demo test test-race vet fmt clean

# Cross-build the static archives for all supported platforms into lib/.
# Needs Rust; a staticlib cross-compiles from any host with no cross C toolchain.
libs:
	shim/build-libs.sh
	@$(MAKE) --no-print-directory stamp

# Build the CLI demo into ./bin.
demo:
	go build -o bin/mmdr-demo ./cmd/mmdr-demo
	@echo "built bin/mmdr-demo"

# Build the Rust shim for the host platform and copy libmmdr.a into lib/.
lib:
	cd shim && cargo build --release
	mkdir -p $(LIBDIR)
	cp shim/target/release/libmmdr.a $(LIBDIR)/libmmdr.a
	@echo "copied libmmdr.a -> $(LIBDIR)"
	@$(MAKE) --no-print-directory stamp

# Refresh archive_stamp.go from the current archives.
#
# Without this, Go's build cache does NOT notice a rebuilt libmmdr.a — cgo names
# the archive by path and Go hashes the LDFLAGS text, not the archive bytes, so
# `go test` silently relinks the stale one. Touching a source file in the package
# changes its cache key and forces the relink. See archive_stamp.go.
stamp:
	@sum=$$(cat lib/*/libmmdr.a | $(SHA256) | cut -d' ' -f1); \
	sed -i.bak -e "s|^const archiveStamp = .*|const archiveStamp = \"sha256:$$sum\"|" archive_stamp.go; \
	rm -f archive_stamp.go.bak; \
	echo "stamped archive_stamp.go -> sha256:$$sum"

test:
	go test ./...

# Race detector without the long leak test.
test-race:
	go test -race -short ./...

vet:
	go vet ./...

fmt:
	go fmt ./...
	cd shim && cargo fmt

clean:
	cd shim && cargo clean
	go clean -testcache
