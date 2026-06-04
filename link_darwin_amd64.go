//go:build darwin && amd64

package mmdr

// native-static-libs for x86_64-apple-darwin is `-liconv -lSystem -lc -lm`;
// macOS Go already links libSystem (libc/libm), so only -liconv is added.

/*
#cgo LDFLAGS: -L${SRCDIR}/lib/darwin_amd64 -lmmdr -liconv
*/
import "C"
