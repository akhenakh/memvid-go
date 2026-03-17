//go:build memvid_static

package memvid

/*
#cgo LDFLAGS: -L./rust/target/release -Wl,-Bstatic -lmemvid_ffi -Wl,-Bdynamic -lstdc++ -lssl -lcrypto -ldl -lpthread -lm
*/
import "C"

// Static linking mode.
//
// This file is selected when the `memvid_static` build tag is provided.
//
// It links the Rust static archive `libmemvid_ffi.a` into the Go binary so the
// `libmemvid_ffi.so` is not required at runtime.
//
// Note: Some dependencies (e.g. OpenSSL, libstdc++) may still be dynamically
// linked unless you also provide static versions of those libraries and adjust
// link flags / toolchain accordingly.
