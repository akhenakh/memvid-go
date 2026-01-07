package memvid

/*
#include <stdlib.h>
#include "memvid_wrapper.h"
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"runtime"
	"unsafe"
)

// Memvid is the main handle to the memory file.
type Memvid struct {
	ptr *C.MemvidHandle
}

// --- Types mirroring Rust structs (JSON compatible) ---

type PutOptions struct {
	Title         string            `json:"title,omitempty"`
	URI           string            `json:"uri,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	ExtraMetadata map[string]string `json:"extra_metadata,omitempty"`
	AutoTag       bool              `json:"auto_tag"`
	ExtractDates  bool              `json:"extract_dates"`
}

type SearchRequest struct {
	Query        string  `json:"query"`
	TopK         int     `json:"top_k"`
	SnippetChars int     `json:"snippet_chars"`
	URI          *string `json:"uri,omitempty"`
	Scope        *string `json:"scope,omitempty"`
	NoSketch     bool    `json:"no_sketch"`
}

type SearchResponse struct {
	Query     string      `json:"query"`
	TotalHits int         `json:"total_hits"`
	ElapsedMs int         `json:"elapsed_ms"`
	Hits      []SearchHit `json:"hits"`
}

type SearchHit struct {
	FrameID uint64  `json:"frame_id"`
	URI     string  `json:"uri"`
	Title   *string `json:"title"`
	Text    string  `json:"text"`
	Score   float64 `json:"score"`
}

type TimelineQuery struct {
	Limit   *int `json:"limit,omitempty"`
	Reverse bool `json:"reverse"`
}

type TimelineEntry struct {
	FrameID   uint64 `json:"frame_id"`
	Timestamp int64  `json:"timestamp"`
	Preview   string `json:"preview"`
	URI       string `json:"uri"`
}

type Stats struct {
	FrameCount   uint64 `json:"frame_count"`
	ActiveFrames uint64 `json:"active_frame_count"`
	HasLexIndex  bool   `json:"has_lex_index"`
	HasVecIndex  bool   `json:"has_vec_index"`
}

type VerificationReport struct {
	OverallStatus string `json:"overall_status"`
	// Add checks field if detailed reporting needed
}

// --- Helper Functions ---

func makeError(cErr *C.char) error {
	if cErr == nil {
		return nil
	}
	defer C.memvid_free_string(cErr)
	return fmt.Errorf("memvid error: %s", C.GoString(cErr))
}

func jsonToC(v interface{}) (*C.char, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return C.CString(string(b)), nil
}

// --- API ---

// Create creates a new memory file at path.
func Create(path string) (*Memvid, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var ptr *C.MemvidHandle
	var cErr *C.char

	res := C.memvid_create(cPath, &ptr, &cErr)
	if res != 0 {
		return nil, makeError(cErr)
	}

	m := &Memvid{ptr: ptr}
	runtime.SetFinalizer(m, func(x *Memvid) { x.Close() })
	return m, nil
}

// Open opens an existing memory file.
func Open(path string) (*Memvid, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var ptr *C.MemvidHandle
	var cErr *C.char

	res := C.memvid_open(cPath, &ptr, &cErr)
	if res != 0 {
		return nil, makeError(cErr)
	}

	m := &Memvid{ptr: ptr}
	runtime.SetFinalizer(m, func(x *Memvid) { x.Close() })
	return m, nil
}

// Close releases the underlying Rust memory.
func (m *Memvid) Close() {
	if m.ptr != nil {
		C.memvid_close(m.ptr)
		m.ptr = nil
	}
}

// PutBytes adds a frame to the memory.
func (m *Memvid) PutBytes(data []byte, opts PutOptions) (uint64, error) {
	if m.ptr == nil {
		return 0, fmt.Errorf("memvid handle is closed")
	}

	cOpts, err := jsonToC(opts)
	if err != nil {
		return 0, err
	}
	defer C.free(unsafe.Pointer(cOpts))

	// Handle empty payload case safely
	var cData *C.uint8_t
	if len(data) > 0 {
		cData = (*C.uint8_t)(unsafe.Pointer(&data[0]))
	}

	var seq C.uint64_t
	var cErr *C.char

	res := C.memvid_put(m.ptr, cData, C.size_t(len(data)), cOpts, &seq, &cErr)
	if res != 0 {
		return 0, makeError(cErr)
	}
	return uint64(seq), nil
}

// Commit persists pending writes (WAL to TOC).
func (m *Memvid) Commit() error {
	if m.ptr == nil {
		return fmt.Errorf("memvid handle is closed")
	}
	var cErr *C.char
	res := C.memvid_commit(m.ptr, &cErr)
	if res != 0 {
		return makeError(cErr)
	}
	return nil
}

// Search queries the memory.
func (m *Memvid) Search(req SearchRequest) (*SearchResponse, error) {
	if m.ptr == nil {
		return nil, fmt.Errorf("memvid handle is closed")
	}

	cReq, err := jsonToC(req)
	if err != nil {
		return nil, err
	}
	defer C.free(unsafe.Pointer(cReq))

	var cResp *C.char
	var cErr *C.char

	res := C.memvid_search(m.ptr, cReq, &cResp, &cErr)
	if res != 0 {
		return nil, makeError(cErr)
	}
	defer C.memvid_free_string(cResp)

	var resp SearchResponse
	if err := json.Unmarshal([]byte(C.GoString(cResp)), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Timeline retrieves frames in chronological order.
func (m *Memvid) Timeline(query TimelineQuery) ([]TimelineEntry, error) {
	if m.ptr == nil {
		return nil, fmt.Errorf("memvid handle is closed")
	}

	cQuery, err := jsonToC(query)
	if err != nil {
		return nil, err
	}
	defer C.free(unsafe.Pointer(cQuery))

	var cResp *C.char
	var cErr *C.char

	res := C.memvid_timeline(m.ptr, cQuery, &cResp, &cErr)
	if res != 0 {
		return nil, makeError(cErr)
	}
	defer C.memvid_free_string(cResp)

	var entries []TimelineEntry
	if err := json.Unmarshal([]byte(C.GoString(cResp)), &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// Stats returns memory statistics.
func (m *Memvid) Stats() (*Stats, error) {
	if m.ptr == nil {
		return nil, fmt.Errorf("memvid handle is closed")
	}

	var cResp *C.char
	var cErr *C.char
	res := C.memvid_stats(m.ptr, &cResp, &cErr)
	if res != 0 {
		return nil, makeError(cErr)
	}
	defer C.memvid_free_string(cResp)

	var stats Stats
	if err := json.Unmarshal([]byte(C.GoString(cResp)), &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// EnableLex enables the lexical (Tantivy) index.
func (m *Memvid) EnableLex() error {
	var cErr *C.char
	if C.memvid_enable_lex(m.ptr, &cErr) != 0 {
		return makeError(cErr)
	}
	return nil
}

// EnableVec enables the vector (HNSW) index.
func (m *Memvid) EnableVec() error {
	var cErr *C.char
	if C.memvid_enable_vec(m.ptr, &cErr) != 0 {
		return makeError(cErr)
	}
	return nil
}

// Verify checks the file integrity.
func Verify(path string, deep bool) (*VerificationReport, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var deepInt C.int = 0
	if deep {
		deepInt = 1
	}

	var cResp *C.char
	var cErr *C.char

	res := C.memvid_verify(cPath, deepInt, &cResp, &cErr)
	if res != 0 {
		return nil, makeError(cErr)
	}
	defer C.memvid_free_string(cResp)

	var report VerificationReport
	if err := json.Unmarshal([]byte(C.GoString(cResp)), &report); err != nil {
		return nil, err
	}

	return &report, nil
}
