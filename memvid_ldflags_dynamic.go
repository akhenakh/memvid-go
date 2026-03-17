//go:build !memvid_static

package memvid

/*
#cgo LDFLAGS: -L./rust/target/release -lmemvid_ffi -ldl -lpthread -lm
*/
import "C"

// Dynamic (default) linking mode.
//
// This file is selected when the `memvid_static` build tag is NOT provided.
//
// It links against the Rust shared object `libmemvid_ffi.so`, which must be
// available to the dynamic loader at runtime (via system paths, rpath, or
// LD_LIBRARY_PATH).
