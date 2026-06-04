# mmdr-go — rebuild the prebuilt static archive and run tests.
#
# Consumers do NOT need this Makefile or a Rust toolchain; the archives are
# committed under lib/. This is for maintainers rebuilding them.

# Map the host to a lib/<os>_<arch> directory.
GOOS  := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
LIBDIR := lib/$(GOOS)_$(GOARCH)

.PHONY: lib demo test test-race vet fmt clean

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
