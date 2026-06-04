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

use mermaid_rs_renderer::render;

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
