package mmdr

/*
#include <stdlib.h>
#include "mmdr.h"
*/
import "C"

import (
	"math"
	"strconv"
	"strings"
	"unsafe"
)

// Options tunes a RenderWithOptions call. The zero value renders exactly like
// Render: engine-default theme, accurate text metrics, intrinsic dimensions.
type Options struct {
	// Theme selects a color palette (case-insensitive). "" or "default"/"modern"
	// use the engine default; "neutral" is upstream's classic mermaid palette.
	// "dark" and "forest" are mmdr-go's OWN palettes — the upstream engine ships
	// no dark/forest themes, so these are defined here and are NOT intended to
	// match mermaid.js's same-named themes.
	//
	// "solid-light" and "solid-dark" are first-party palettes tuned to the
	// Solid/V4 app's design tokens, so diagrams blend into its shell. Both use a
	// TRANSPARENT background (the root <rect> is fill="none"), inheriting the host
	// surface's background rather than painting their own.
	//
	// Unrecognized names fall back to the engine default.
	Theme string
	// FastText skips the font-database load and uses approximate text widths.
	// Faster, with slightly less accurate text sizing.
	FastText bool
	// Width and Height are intrinsic when <= 0. The SVG renderer has no
	// fixed-pixel sizing knob upstream; when BOTH are positive they are used as
	// a preferred aspect ratio (Width/Height). A single dimension is ignored.
	Width  int
	Height int
	// AspectRatio hints a preferred output ratio, e.g. "16:9", "4/3", or "1.5".
	// "" leaves the ratio intrinsic. It takes precedence over Width/Height.
	AspectRatio string
}

// Result is the output of RenderWithOptions: the SVG plus the rendered root
// element's width and height, rounded to the nearest integer. Width or Height
// is 0 when the SVG root carries a non-numeric dimension (e.g. "100%").
type Result struct {
	SVG    string
	Width  int
	Height int
}

// RenderWithOptions renders Mermaid source to SVG with the given Options and
// reports the rendered dimensions.
//
// Error behavior matches Render: a *RenderError wrapping ErrInvalidInput on a
// parse/render failure, or ErrPanic on a renderer panic.
func RenderWithOptions(source string, opts Options) (Result, error) {
	cInput := C.CString(source)
	defer C.free(unsafe.Pointer(cInput))

	cTheme := C.CString(opts.Theme)
	defer C.free(unsafe.Pointer(cTheme))

	cAspect := C.CString(opts.AspectRatio)
	defer C.free(unsafe.Pointer(cAspect))

	var fastText C.int
	if opts.FastText {
		fastText = 1
	}

	var cSVG *C.char
	var cErr *C.char

	status := C.mmdr_render_with_options(
		cInput,
		cTheme,
		fastText,
		C.int(opts.Width),
		C.int(opts.Height),
		cAspect,
		&cSVG,
		&cErr,
	)
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
		svg := C.GoString(cSVG)
		w, h := svgRootDimensions(svg)
		return Result{SVG: svg, Width: w, Height: h}, nil
	case C.MMDR_PANIC:
		return Result{}, &RenderError{sentinel: ErrPanic, Detail: C.GoString(cErr)}
	default: // MMDR_RENDER_ERROR, MMDR_BAD_INPUT
		return Result{}, &RenderError{sentinel: ErrInvalidInput, Detail: C.GoString(cErr)}
	}
}

// svgRootDimensions extracts the width and height attributes of the first <svg>
// element and rounds them to the nearest int. A missing or non-numeric value
// (e.g. "100%") yields 0 for that dimension.
func svgRootDimensions(svg string) (width, height int) {
	start := strings.Index(svg, "<svg")
	if start < 0 {
		return 0, 0
	}
	end := strings.IndexByte(svg[start:], '>')
	if end < 0 {
		return 0, 0
	}
	tag := svg[start : start+end]
	return roundAttr(attrValue(tag, "width")), roundAttr(attrValue(tag, "height"))
}

// attrValue returns the value of a double-quoted attribute within an opening
// tag, or "" if absent. It matches whole attribute names (avoiding e.g.
// "stroke-width" when looking for "width").
func attrValue(tag, name string) string {
	from := 0
	for {
		i := strings.Index(tag[from:], name+"=\"")
		if i < 0 {
			return ""
		}
		pos := from + i
		// Ensure the char before the name is a boundary (space or tag start),
		// so "width" does not match inside "stroke-width".
		if pos == 0 || isAttrBoundary(tag[pos-1]) {
			rest := tag[pos+len(name)+2:]
			if j := strings.IndexByte(rest, '"'); j >= 0 {
				return rest[:j]
			}
			return ""
		}
		from = pos + len(name)
	}
}

func isAttrBoundary(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// roundAttr parses a numeric attribute value and rounds it to the nearest int.
// Returns 0 for empty or non-numeric values (e.g. "100%").
func roundAttr(v string) int {
	if v == "" {
		return 0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return int(math.Round(f))
}
