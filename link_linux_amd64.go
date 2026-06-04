//go:build linux && amd64

package mmdr

// `cargo rustc --release --target x86_64-unknown-linux-gnu -- --print
// native-static-libs` reports the glibc set below. -lgcc_s provides the
// unwinder the shim's catch_unwind needs; the rest are standard glibc libs the
// Go linker mostly pulls in already (harmless to repeat on GNU ld).

/*
#cgo LDFLAGS: -L${SRCDIR}/lib/linux_amd64 -lmmdr -lgcc_s -lm -ldl -lpthread
*/
import "C"
