#ifndef MMDR_H
#define MMDR_H

/* C declarations for the mmdr-go cgo binding.
 * Keep in sync with shim/src/lib.rs. */

/* Status codes returned by mmdr_render. */
#define MMDR_OK 0           /* success; *out_svg owns the SVG string        */
#define MMDR_RENDER_ERROR 1 /* parse/render failure; *out_err owns message  */
#define MMDR_PANIC 2        /* renderer panicked; *out_err owns message     */
#define MMDR_BAD_INPUT 3    /* input was null or not valid UTF-8            */

/* Render a Mermaid diagram to SVG.
 *
 * Returns one of the MMDR_* codes. out_svg and out_err are set to NULL first,
 * so they may be read unconditionally. On MMDR_OK, *out_svg owns a C string.
 * On an error code, *out_err may own a message string. Every non-NULL pointer
 * written to *out_svg / *out_err must be released with mmdr_free. */
int mmdr_render(const char *input, char **out_svg, char **out_err);

/* Render a Mermaid diagram to SVG with options.
 *
 * Same status codes, out-param contract, and ownership rules as mmdr_render.
 * Conventions: a null or empty theme/aspect_ratio means unset/intrinsic;
 * width/height <= 0 means intrinsic; fast_text != 0 enables fast text metrics.
 * theme accepts "" (engine default), "default"/"modern", "dark", "neutral",
 * "forest" (case-insensitive); unknown names keep the engine default.
 * aspect_ratio accepts "W:H", "W/H", or a decimal. When both width and height
 * are positive and no aspect_ratio is given, width/height is used as the
 * preferred aspect ratio (the SVG path has no fixed-pixel sizing). */
int mmdr_render_with_options(const char *input, const char *theme, int fast_text,
                             int width, int height, const char *aspect_ratio,
                             char **out_svg, char **out_err);

/* Free a string returned by mmdr_render or mmdr_version. NULL is a no-op. */
void mmdr_free(char *ptr);

/* Return the upstream mermaid-rs-renderer version. Caller frees with mmdr_free. */
char *mmdr_version(void);

#endif /* MMDR_H */
