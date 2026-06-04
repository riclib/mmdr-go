package mmdr

import (
	"errors"
	"strings"
	"testing"
)

const optsDiagram = "flowchart LR\n A-->B-->C"

func TestRenderWithOptions_Default(t *testing.T) {
	r, err := RenderWithOptions(optsDiagram, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(r.SVG, "<svg") {
		t.Fatalf("output is not SVG: %q", truncate(r.SVG))
	}
	if r.Width <= 0 || r.Height <= 0 {
		t.Fatalf("expected positive dimensions, got W=%d H=%d", r.Width, r.Height)
	}
}

func TestRenderWithOptions_ThemeChangesOutput(t *testing.T) {
	def, err := RenderWithOptions(optsDiagram, Options{})
	if err != nil {
		t.Fatalf("default render failed: %v", err)
	}
	dark, err := RenderWithOptions(optsDiagram, Options{Theme: "dark"})
	if err != nil {
		t.Fatalf("dark render failed: %v", err)
	}
	if dark.SVG == def.SVG {
		t.Fatal("Theme:\"dark\" produced identical SVG to the default theme")
	}
}

func TestRenderWithOptions_FastText(t *testing.T) {
	r, err := RenderWithOptions(optsDiagram, Options{FastText: true})
	if err != nil {
		t.Fatalf("fast-text render failed: %v", err)
	}
	if !strings.Contains(r.SVG, "<svg") {
		t.Fatalf("fast-text output is not SVG: %q", truncate(r.SVG))
	}
	if r.Width <= 0 || r.Height <= 0 {
		t.Fatalf("expected positive dimensions, got W=%d H=%d", r.Width, r.Height)
	}
}

func TestRenderWithOptions_AspectRatio(t *testing.T) {
	// A wide aspect ratio should reshape the rendered dimensions relative to the
	// intrinsic render, and the ratio hint should appear in the SVG.
	intrinsic, err := RenderWithOptions(optsDiagram, Options{})
	if err != nil {
		t.Fatalf("intrinsic render failed: %v", err)
	}
	wide, err := RenderWithOptions(optsDiagram, Options{AspectRatio: "16:9"})
	if err != nil {
		t.Fatalf("aspect-ratio render failed: %v", err)
	}
	if !strings.Contains(wide.SVG, "aspect-ratio") {
		t.Errorf("expected aspect-ratio hint in SVG, got root: %q", svgRoot(wide.SVG))
	}
	if wide.Width <= 0 || wide.Height <= 0 {
		t.Fatalf("expected positive dimensions, got W=%d H=%d", wide.Width, wide.Height)
	}
	// 16:9 is wider than the intrinsic ~4.25:1? Regardless, the fitted ratio must
	// differ from the intrinsic shape unless they already matched.
	if wide.Width == intrinsic.Width && wide.Height == intrinsic.Height {
		t.Errorf("aspect ratio did not change dimensions: %dx%d", wide.Width, wide.Height)
	}
}

func TestRenderWithOptions_WidthHeightRatio(t *testing.T) {
	// Both dimensions positive => preferred aspect ratio width/height. A 1:1
	// request should square up the result relative to the intrinsic shape.
	square, err := RenderWithOptions(optsDiagram, Options{Width: 400, Height: 400})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if square.Width <= 0 || square.Height <= 0 {
		t.Fatalf("expected positive dimensions, got W=%d H=%d", square.Width, square.Height)
	}
	if !strings.Contains(square.SVG, "aspect-ratio") {
		t.Errorf("expected aspect-ratio hint from width/height, root: %q", svgRoot(square.SVG))
	}
}

func TestRenderWithOptions_InvalidInput(t *testing.T) {
	// The render engine is lenient and rarely rejects malformed Mermaid text, so
	// drive the guaranteed error path: invalid UTF-8 cannot cross the C boundary
	// and is reported as bad input (wrapping ErrInvalidInput).
	bad := string([]byte{0xff, 0xfe, 'f', 'l', 'o', 'w'})
	_, err := RenderWithOptions(bad, Options{})
	if err == nil {
		t.Fatal("expected an error for invalid UTF-8 input")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestRenderWithOptions_SolidThemes(t *testing.T) {
	def, err := RenderWithOptions(optsDiagram, Options{})
	if err != nil {
		t.Fatalf("default render failed: %v", err)
	}
	dark, err := RenderWithOptions(optsDiagram, Options{Theme: "solid-dark"})
	if err != nil {
		t.Fatalf("solid-dark render failed: %v", err)
	}
	light, err := RenderWithOptions(optsDiagram, Options{Theme: "solid-light"})
	if err != nil {
		t.Fatalf("solid-light render failed: %v", err)
	}

	// Both must be valid, positively-sized SVG.
	for name, r := range map[string]Result{"solid-dark": dark, "solid-light": light} {
		if !strings.Contains(r.SVG, "<svg") {
			t.Fatalf("%s output is not SVG: %q", name, truncate(r.SVG))
		}
		if r.Width <= 0 || r.Height <= 0 {
			t.Fatalf("%s expected positive dimensions, got W=%d H=%d", name, r.Width, r.Height)
		}
		// Transparent background: the root full-canvas <rect> must be fill="none".
		i := strings.Index(r.SVG, "<rect")
		if i < 0 {
			t.Fatalf("%s has no <rect> background element", name)
		}
		j := strings.IndexByte(r.SVG[i:], '>')
		root := r.SVG[i : i+j+1]
		if !strings.Contains(root, `fill="none"`) {
			t.Errorf("%s root background rect not transparent: %q", name, root)
		}
	}

	// The two Solid themes must differ from each other and from the default.
	if dark.SVG == light.SVG {
		t.Error("solid-dark and solid-light produced identical SVG")
	}
	if dark.SVG == def.SVG {
		t.Error("solid-dark produced identical SVG to the default theme")
	}
	if light.SVG == def.SVG {
		t.Error("solid-light produced identical SVG to the default theme")
	}
}

func TestRenderWithOptions_UnknownThemeFallsBack(t *testing.T) {
	// An unknown theme keeps the engine default rather than erroring.
	def, err := RenderWithOptions(optsDiagram, Options{})
	if err != nil {
		t.Fatalf("default render failed: %v", err)
	}
	unknown, err := RenderWithOptions(optsDiagram, Options{Theme: "no-such-theme"})
	if err != nil {
		t.Fatalf("unknown-theme render failed: %v", err)
	}
	if unknown.SVG != def.SVG {
		t.Error("unknown theme should fall back to the engine default")
	}
}

func truncate(s string) string {
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

func svgRoot(svg string) string {
	if i := strings.IndexByte(svg, '>'); i >= 0 {
		return svg[:i+1]
	}
	return svg
}
