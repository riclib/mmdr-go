// Package mmdr renders Mermaid diagrams to SVG using the native Rust
// mermaid-rs-renderer engine, linked in as a static library via cgo.
//
// Build requirement: CGO_ENABLED=1 (the default on most platforms). No Rust
// toolchain and no mmdr binary are needed at consumer build time — prebuilt
// static archives ship in this module's lib/ directory and are selected by
// build tag, so only the matching architecture is embedded.
//
// Thread safety: Render is safe to call concurrently from multiple goroutines.
// See the THREAD SAFETY section in the README for the basis of that claim.
package mmdr

/*
#include <stdlib.h>
#include "mmdr.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// Sentinel errors wrapped by the errors returned from Render. Match them with
// errors.Is, e.g. errors.Is(err, mmdr.ErrInvalidInput).
var (
	// ErrInvalidInput indicates the Mermaid source failed to parse or render,
	// or was not valid UTF-8.
	ErrInvalidInput = errors.New("mmdr: invalid mermaid source")
	// ErrPanic indicates the underlying Rust renderer panicked. The panic was
	// caught at the FFI boundary, so the Go process is unharmed, but no SVG is
	// available.
	ErrPanic = errors.New("mmdr: rust panic during render")
)

// RenderError carries the message reported by the renderer and wraps one of the
// sentinel errors so errors.Is works.
type RenderError struct {
	sentinel error
	// Detail is the message reported by the Rust renderer, if any.
	Detail string
}

func (e *RenderError) Error() string {
	if e.Detail == "" {
		return e.sentinel.Error()
	}
	return e.sentinel.Error() + ": " + e.Detail
}

func (e *RenderError) Unwrap() error { return e.sentinel }

// Render parses Mermaid source and returns the rendered SVG.
//
// On a parse/render failure it returns a *RenderError wrapping ErrInvalidInput;
// on a renderer panic it returns a *RenderError wrapping ErrPanic.
func Render(source string) (string, error) {
	cInput := C.CString(source)
	defer C.free(unsafe.Pointer(cInput))

	var cSVG *C.char
	var cErr *C.char

	status := C.mmdr_render(cInput, &cSVG, &cErr)
	// The shim owns these via Rust's allocator; release them with mmdr_free
	// (never C.free). Safe to register both unconditionally — nil is a no-op.
	if cSVG != nil {
		defer C.mmdr_free(cSVG)
	}
	if cErr != nil {
		defer C.mmdr_free(cErr)
	}

	switch status {
	case C.MMDR_OK:
		return C.GoString(cSVG), nil
	case C.MMDR_PANIC:
		return "", &RenderError{sentinel: ErrPanic, Detail: C.GoString(cErr)}
	default: // MMDR_RENDER_ERROR, MMDR_BAD_INPUT
		return "", &RenderError{sentinel: ErrInvalidInput, Detail: C.GoString(cErr)}
	}
}

// Version returns the upstream mermaid-rs-renderer version this binding wraps,
// e.g. "0.2.2".
func Version() string {
	cVer := C.mmdr_version()
	if cVer == nil {
		return ""
	}
	defer C.mmdr_free(cVer)
	return C.GoString(cVer)
}
