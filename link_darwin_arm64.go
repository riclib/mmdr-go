//go:build darwin && arm64

package mmdr

// Link the prebuilt static archive for this platform. Build tags ensure a
// consumer binary embeds only the matching architecture's ~9 MB archive.
//
// `cargo rustc -- --print native-static-libs` reports `-liconv -lSystem -lc -lm`
// for the aarch64-apple-darwin staticlib. On macOS the Go linker already pulls
// in libSystem (which provides libc and libm), so only -liconv must be added
// explicitly; listing the rest again triggers "ignoring duplicate libraries".

/*
#cgo LDFLAGS: -L${SRCDIR}/lib/darwin_arm64 -lmmdr -liconv
*/
import "C"
