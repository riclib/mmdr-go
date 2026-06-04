//! C-ABI shim over [`mermaid-rs-renderer`], exposing a tiny, panic-safe surface
//! consumed by the `github.com/riclib/mmdr-go` cgo binding.
//!
//! Invariants (see CONTRIBUTING.md):
//!   * Every render call is wrapped in `catch_unwind` — a Rust panic returns a
//!     status code, never unwinds across the FFI boundary (which would be UB).
//!   * The shim allocates every returned string with Rust's allocator; the
//!     caller must release it with `mmdr_free`. Never free a shim pointer with
//!     C `free`, and never pass a C-allocated pointer to `mmdr_free`.

use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use std::panic::{self, AssertUnwindSafe};
use std::ptr;

use mermaid_rs_renderer::{render, render_with_options, RenderOptions, Theme};

/// Upstream `mermaid-rs-renderer` version this shim is pinned to.
/// Must match the `=x.y.z` requirement in Cargo.toml and the `.mmdr-version` file.
const UPSTREAM_VERSION: &str = "0.2.2";

// Status codes returned by `mmdr_render`. Mirrored in mmdr.h and mmdr.go.
const MMDR_OK: i32 = 0;
const MMDR_RENDER_ERROR: i32 = 1;
const MMDR_PANIC: i32 = 2;
const MMDR_BAD_INPUT: i32 = 3;

/// Render a Mermaid diagram to SVG.
///
/// Returns one of the `MMDR_*` status codes:
///   * `MMDR_OK` — `*out_svg` owns a heap C string with the SVG.
///   * `MMDR_RENDER_ERROR` — parse/render failure; `*out_err` owns the message.
///   * `MMDR_PANIC` — the renderer panicked; `*out_err` owns the panic message.
///     No UB crosses the boundary; the host process is unharmed.
///   * `MMDR_BAD_INPUT` — `input` was null or not valid UTF-8.
///
/// `out_svg` and `out_err` are always initialized to null first, so a caller
/// may read them unconditionally. Any non-null pointer they receive is owned by
/// the caller and must be released with `mmdr_free`.
///
/// # Safety
/// `input` must be a valid NUL-terminated C string or null. `out_svg` and
/// `out_err` must be valid pointers to `*mut c_char` (or null to discard).
#[no_mangle]
pub extern "C" fn mmdr_render(
    input: *const c_char,
    out_svg: *mut *mut c_char,
    out_err: *mut *mut c_char,
) -> i32 {
    // Initialize out-params so callers never read uninitialized memory.
    unsafe {
        if !out_svg.is_null() {
            *out_svg = ptr::null_mut();
        }
        if !out_err.is_null() {
            *out_err = ptr::null_mut();
        }
    }

    if input.is_null() {
        return MMDR_BAD_INPUT;
    }

    let source = match unsafe { CStr::from_ptr(input) }.to_str() {
        Ok(s) => s,
        Err(_) => return MMDR_BAD_INPUT,
    };

    match panic::catch_unwind(AssertUnwindSafe(|| render(source))) {
        Ok(Ok(svg)) => match CString::new(svg) {
            Ok(c) => {
                write_str(out_svg, c);
                MMDR_OK
            }
            // The SVG contained an interior NUL byte — surface as a render error.
            Err(_) => {
                write_err(out_err, "rendered SVG contained an interior NUL byte");
                MMDR_RENDER_ERROR
            }
        },
        Ok(Err(e)) => {
            write_err(out_err, &format!("{e}"));
            MMDR_RENDER_ERROR
        }
        Err(payload) => {
            write_err(out_err, &panic_message(payload.as_ref()));
            MMDR_PANIC
        }
    }
}

/// Render a Mermaid diagram to SVG with options (theme, fast-text, dimensions,
/// aspect ratio).
///
/// Same status codes, out-param contract, allocator ownership, and panic safety
/// as [`mmdr_render`]; only the option arguments differ.
///
/// Option conventions (decoded here, applied to `RenderOptions`):
///   * `theme` — null or empty = engine default (`Theme::modern`). `"default"`,
///     `"modern"`, `"dark"`, `"neutral"`, `"forest"` select a palette (see
///     `theme_for`). Unknown non-empty names fall back to the engine default.
///   * `fast_text` — nonzero enables approximate (font-DB-free) text metrics.
///   * `width` / `height` — `<= 0` means intrinsic. When BOTH are positive they
///     are translated into a preferred aspect ratio (`width/height`); the SVG
///     path has no fixed-pixel sizing (that is a PNG-only knob upstream), so the
///     diagram keeps its intrinsic scale but is fitted to the requested ratio.
///     A single dimension alone yields no ratio and is left intrinsic.
///   * `aspect_ratio` — null/empty = intrinsic; otherwise `"W:H"`, `"W/H"`, or a
///     decimal. Takes precedence over a width/height-derived ratio.
///
/// # Safety
/// `input`, `theme`, and `aspect_ratio` must each be a valid NUL-terminated C
/// string or null. `out_svg` and `out_err` must be valid pointers to
/// `*mut c_char` (or null to discard).
#[no_mangle]
pub extern "C" fn mmdr_render_with_options(
    input: *const c_char,
    theme: *const c_char,
    fast_text: i32,
    width: i32,
    height: i32,
    aspect_ratio: *const c_char,
    out_svg: *mut *mut c_char,
    out_err: *mut *mut c_char,
) -> i32 {
    // Initialize out-params so callers never read uninitialized memory.
    unsafe {
        if !out_svg.is_null() {
            *out_svg = ptr::null_mut();
        }
        if !out_err.is_null() {
            *out_err = ptr::null_mut();
        }
    }

    if input.is_null() {
        return MMDR_BAD_INPUT;
    }

    let source = match unsafe { CStr::from_ptr(input) }.to_str() {
        Ok(s) => s,
        Err(_) => return MMDR_BAD_INPUT,
    };

    // Optional C strings: null = unset. Non-UTF-8 is treated as bad input, like
    // the source itself.
    let theme_str = match opt_str(theme) {
        Ok(s) => s,
        Err(_) => return MMDR_BAD_INPUT,
    };
    let aspect_str = match opt_str(aspect_ratio) {
        Ok(s) => s,
        Err(_) => return MMDR_BAD_INPUT,
    };

    // Build options outside catch_unwind (pure data assembly, no upstream call).
    let mut options = RenderOptions::default();
    if let Some(t) = theme_for(theme_str) {
        options.theme = t;
    }
    if fast_text != 0 {
        options.layout.fast_text_metrics = true;
    }
    // aspect_ratio wins over a width/height-derived ratio.
    if let Some(ratio) = aspect_str.and_then(parse_aspect_ratio) {
        options.layout.preferred_aspect_ratio = Some(ratio);
    } else if width > 0 && height > 0 {
        options.layout.preferred_aspect_ratio = Some(width as f32 / height as f32);
    }

    match panic::catch_unwind(AssertUnwindSafe(|| render_with_options(source, options))) {
        Ok(Ok(svg)) => match CString::new(svg) {
            Ok(c) => {
                write_str(out_svg, c);
                MMDR_OK
            }
            Err(_) => {
                write_err(out_err, "rendered SVG contained an interior NUL byte");
                MMDR_RENDER_ERROR
            }
        },
        Ok(Err(e)) => {
            write_err(out_err, &format!("{e}"));
            MMDR_RENDER_ERROR
        }
        Err(payload) => {
            write_err(out_err, &panic_message(payload.as_ref()));
            MMDR_PANIC
        }
    }
}

/// Free a string previously produced by `mmdr_render` (via `out_svg`/`out_err`)
/// or `mmdr_version`. Passing null is a no-op; double-free is undefined.
///
/// # Safety
/// `ptr` must be null or a pointer returned by this shim and not yet freed.
#[no_mangle]
pub extern "C" fn mmdr_free(ptr: *mut c_char) {
    if ptr.is_null() {
        return;
    }
    // Reclaim ownership and drop, freeing with Rust's allocator.
    unsafe { drop(CString::from_raw(ptr)) };
}

/// Return the upstream `mermaid-rs-renderer` version as a heap C string, or
/// null on allocation failure. Caller frees with `mmdr_free`.
#[no_mangle]
pub extern "C" fn mmdr_version() -> *mut c_char {
    match CString::new(UPSTREAM_VERSION) {
        Ok(c) => c.into_raw(),
        Err(_) => ptr::null_mut(),
    }
}

/// Hand ownership of `value` to the caller through `out`. No-op if `out` is null.
fn write_str(out: *mut *mut c_char, value: CString) {
    if out.is_null() {
        return;
    }
    unsafe { *out = value.into_raw() };
}

/// Write an error message to `out`, sanitizing interior NULs so allocation of
/// the C string cannot fail on the error path.
fn write_err(out: *mut *mut c_char, msg: &str) {
    if out.is_null() {
        return;
    }
    let sanitized = msg.replace('\0', " ");
    if let Ok(c) = CString::new(sanitized) {
        unsafe { *out = c.into_raw() };
    }
}

/// Best-effort extraction of a human-readable message from a panic payload.
fn panic_message(payload: &(dyn std::any::Any + Send)) -> String {
    if let Some(s) = payload.downcast_ref::<&str>() {
        (*s).to_string()
    } else if let Some(s) = payload.downcast_ref::<String>() {
        s.clone()
    } else {
        "unknown panic".to_string()
    }
}

/// Decode an optional C string. `None` for a null pointer; `Err` for non-UTF-8.
///
/// # Safety
/// `ptr` must be null or a valid NUL-terminated C string.
fn opt_str<'a>(ptr: *const c_char) -> Result<Option<&'a str>, ()> {
    if ptr.is_null() {
        return Ok(None);
    }
    match unsafe { CStr::from_ptr(ptr) }.to_str() {
        Ok(s) => Ok(Some(s)),
        Err(_) => Err(()),
    }
}

/// Map a theme name to a `Theme`, or `None` to keep the engine default.
///
/// The upstream crate ships only two theme constructors — `Theme::modern` (the
/// engine default) and `Theme::mermaid_default`. The dark/neutral/forest
/// palettes are synthesized here by overriding the public color fields of a base
/// theme; they are genuine distinct palettes, not aliases. Matching is
/// case-insensitive; an empty or unrecognized name keeps the default.
fn theme_for(name: Option<&str>) -> Option<Theme> {
    let name = name?.trim();
    if name.is_empty() {
        return None;
    }
    match name.to_ascii_lowercase().as_str() {
        // Engine default; returning None leaves RenderOptions::default() in place.
        "default" | "modern" => None,
        // The classic Mermaid palette — upstream's only other real theme. Exposed
        // under "neutral" as the closest light, non-modern option.
        "neutral" => Some(Theme::mermaid_default()),
        "dark" => Some(dark_theme()),
        "forest" => Some(forest_theme()),
        // Unknown name: keep the engine default rather than guess.
        _ => None,
    }
}

/// A dark palette synthesized from the modern theme by overriding its public
/// color fields. Upstream has no dark constructor.
fn dark_theme() -> Theme {
    let mut t = Theme::modern();
    t.background = "#1E1E2E".to_string();
    t.primary_color = "#313244".to_string();
    t.primary_text_color = "#CDD6F4".to_string();
    t.primary_border_color = "#585B70".to_string();
    t.line_color = "#A6ADC8".to_string();
    t.secondary_color = "#45475A".to_string();
    t.tertiary_color = "#313244".to_string();
    t.text_color = "#CDD6F4".to_string();
    t.edge_label_background = "#1E1E2E".to_string();
    t.cluster_background = "#181825".to_string();
    t.cluster_border = "#585B70".to_string();
    t
}

/// A green-tinted palette synthesized from the classic theme. Upstream has no
/// forest constructor.
fn forest_theme() -> Theme {
    let mut t = Theme::mermaid_default();
    t.background = "#F1F8F1".to_string();
    t.primary_color = "#CDE8CD".to_string();
    t.primary_border_color = "#2E7D32".to_string();
    t.line_color = "#2E7D32".to_string();
    t.secondary_color = "#A5D6A7".to_string();
    t.tertiary_color = "#E8F5E9".to_string();
    t.cluster_background = "#DCEDC8".to_string();
    t.cluster_border = "#558B2F".to_string();
    t
}

/// Parse an aspect ratio of the form `"W:H"`, `"W/H"`, or a bare decimal.
/// Returns `None` for empty, malformed, or non-positive/non-finite values.
/// Mirrors the CLI's `parse_aspect_ratio_value`.
fn parse_aspect_ratio(raw: &str) -> Option<f32> {
    let value = raw.trim();
    if value.is_empty() {
        return None;
    }
    let parse_pair = |w: &str, h: &str| -> Option<f32> {
        let w = w.trim().parse::<f32>().ok()?;
        let h = h.trim().parse::<f32>().ok()?;
        if !w.is_finite() || !h.is_finite() || w <= 0.0 || h <= 0.0 {
            return None;
        }
        Some(w / h)
    };
    if let Some((w, h)) = value.split_once(':') {
        return parse_pair(w, h);
    }
    if let Some((w, h)) = value.split_once('/') {
        return parse_pair(w, h);
    }
    let ratio = value.parse::<f32>().ok()?;
    if !ratio.is_finite() || ratio <= 0.0 {
        return None;
    }
    Some(ratio)
}
