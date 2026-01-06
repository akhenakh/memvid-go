# memvid-go

Go bindings for [Memvid](https://github.com/memvid/memvid), a crash-safe, deterministic, single-file memory layer for AI agents.

This experimental package allows Go applications to create, write to, search, and manage `.mv2` memory files using the high-performance Rust core via CGO.

## Prerequisites

*   **Go**: 1.20 or newer.
*   **Rust**: 1.85+ (required to build the FFI layer).
*   **C Compiler**: GCC or Clang (required for CGO).

## Installation

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/akhenakh/memvid-go.git
    cd memvid-go
    ```

2.  **Build the Rust FFI library**:
    The Go bindings rely on the Rust static library. You must build it before running Go code.
    ```bash
    cd rust
    cargo build --release
    cd ..
    ```
    This creates `rust/target/release/libmemvid_ffi.a` (and `.so`/`.dylib`), which CGO is configured to link against.

## Usage

### Basic Example

The example demonstrates creating a memory file, adding a document, committing changes, and performing a search.

File: `cmd/basic/main.go`


 ### PDF Ingestion Example

This example demonstrates how to ingest a PDF document (automatically downloads the "Attention Is All You Need" paper), extract its text, and perform search queries against it.

File: `cmd/pdf-ingestion/main.go`

Run it with:
```bash
go run ./cmd/pdf-ingestion
```

## Project Structure

*   `memvid.go`: The Go API wrapper.
*   `memvid_wrapper.h`: C header definition for CGO.
*   `rust/`: Contains the `memvid-ffi` Rust crate that bridges `memvid-core` to C.
*   `cmd/`: Go example applications.

## License

Apache License 2.0
