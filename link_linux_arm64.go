//go:build linux && arm64

package mmdr

// native-static-libs for aarch64-unknown-linux-gnu matches the amd64 gnu set;
// -lgcc_s is the unwinder catch_unwind needs, the rest are standard glibc libs.

/*
#cgo LDFLAGS: -L${SRCDIR}/lib/linux_arm64 -lmmdr -lgcc_s -lm -ldl -lpthread
*/
import "C"
