GO ?= go
CARGO ?= cargo

RUST_MANIFEST := rust/Cargo.toml

APP_BASIC := basic
APP_INGESTPDF := ingestpdf
BIN_DIR := bin

.PHONY: cargo build-dynamic build-static clean

# Default target: dynamic linking (requires libmemvid_ffi.so at runtime).
build-dynamic: cargo
	@mkdir -p ./$(BIN_DIR)
	$(GO) build -trimpath -o ./$(BIN_DIR)/$(APP_BASIC) ./cmd/basic
	$(GO) build -trimpath -o ./$(BIN_DIR)/$(APP_INGESTPDF) ./cmd/ingestpdf

# Static Rust linking (links libmemvid_ffi.a into Go binaries; no libmemvid_ffi.so needed at runtime).
build-static: cargo
	@mkdir -p ./$(BIN_DIR)
	$(GO) build -trimpath -tags memvid_static -o ./$(BIN_DIR)/$(APP_BASIC) ./cmd/basic
	$(GO) build -trimpath -tags memvid_static -o ./$(BIN_DIR)/$(APP_INGESTPDF) ./cmd/ingestpdf

cargo:
	$(CARGO) build --release --manifest-path $(RUST_MANIFEST)

clean:
	@rm -rf ./$(BIN_DIR)
	@rm -rf ./rust/target
