package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/akhenakh/memvid-go"
)

func main() {
	// Create a temporary directory for our example
	tempDir, err := os.MkdirTemp("", "memvid-example")
	if err != nil {
		log.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // clean up

	dbPath := filepath.Join(tempDir, "example.mv2")

	fmt.Println("=== Memvid Core Basic Usage Example ===")

	// ========================================
	// 1. CREATE a new memory file
	// ========================================
	fmt.Printf("1. Creating memory file at %s\n", dbPath)
	mem, err := memvid.Create(dbPath)
	if err != nil {
		log.Fatalf("Failed to create memory: %v", err)
	}
	// Ensure we close the handle when done, though we will close explicitly later to reopen
	// defer mem.Close()

	// Enable features (optional, but ensures indexes are active)
	_ = mem.EnableLex()
	_ = mem.EnableVec()

	fmt.Println("   Memory created successfully!")

	// ========================================
	// 2. PUT documents into the memory
	// ========================================
	fmt.Println("2. Adding documents to memory...")

	// Simple put with just bytes
	seq1, err := mem.PutBytes([]byte("Hello, Memvid! This is a simple text document."), memvid.PutOptions{})
	if err != nil {
		log.Fatalf("Put failed: %v", err)
	}
	fmt.Printf("   Added document 1, sequence: %d\n", seq1)

	// Put with options (title, URI, tags)
	opts2 := memvid.PutOptions{
		Title: "Getting Started Guide",
		URI:   "mv2://docs/getting-started.md",
		ExtraMetadata: map[string]string{
			"category": "documentation",
			"version":  "2.0",
		},
		AutoTag: true,
	}
	seq2, err := mem.PutBytes(
		[]byte("This guide covers the basics of using Memvid for AI memory storage."),
		opts2,
	)
	if err != nil {
		log.Fatalf("Put failed: %v", err)
	}
	fmt.Printf("   Added document 2 (with metadata), sequence: %d\n", seq2)

	// Add more documents
	opts3 := memvid.PutOptions{
		Title: "API Reference",
		URI:   "mv2://docs/api-reference.md",
		ExtraMetadata: map[string]string{
			"category": "documentation",
		},
	}
	_, err = mem.PutBytes(
		[]byte("The Memvid API provides methods for create, put, find, and timeline operations."),
		opts3,
	)
	if err != nil {
		log.Fatal(err)
	}

	opts4 := memvid.PutOptions{
		Title: "FAQ",
		URI:   "mv2://docs/faq.md",
		ExtraMetadata: map[string]string{
			"category": "support",
		},
	}
	_, err = mem.PutBytes(
		[]byte("Frequently asked questions about Memvid memory files and search."),
		opts4,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Commit changes to persist them
	if err := mem.Commit(); err != nil {
		log.Fatalf("Commit failed: %v", err)
	}
	fmt.Println("   Committed all changes")

	// ========================================
	// 3. STATS - Check memory statistics
	// ========================================
	fmt.Println("3. Memory statistics:")
	stats, err := mem.Stats()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Frame count: %d\n", stats.FrameCount)
	fmt.Printf("   Has lexical index: %v\n", stats.HasLexIndex)
	fmt.Printf("   Has vector index: %v\n", stats.HasVecIndex)
	// Go struct might not expose HasTimeIndex if not added yet, but core has it
	fmt.Println()

	// ========================================
	// 4. FIND - Search for documents
	// ========================================
	fmt.Println("4. Searching for documents...")

	// Search for "memvid"
	req1 := memvid.SearchRequest{
		Query:        "memvid",
		TopK:         10,
		SnippetChars: 200,
		NoSketch:     false,
	}
	resp, err := mem.Search(req1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("   Query: 'memvid'\n")
	fmt.Printf("   Total hits: %d\n", resp.TotalHits)
	fmt.Printf("   Elapsed: %dms\n", resp.ElapsedMs)

	for _, hit := range resp.Hits {
		title := "Untitled"
		if hit.Title != nil {
			title = *hit.Title
		}

		// Truncate snippet for display like the Rust example
		snippet := hit.Text
		if len(snippet) > 60 {
			snippet = snippet[:60]
		}

		fmt.Printf("   - [%d] %s (score: %.3f)\n", hit.FrameID, title, hit.Score)
		fmt.Printf("     Snippet: %s...\n", snippet)
	}
	fmt.Println()

	// Search within a scope
	scope := "mv2://docs/"
	req2 := memvid.SearchRequest{
		Query:        "documentation",
		TopK:         10,
		SnippetChars: 100,
		Scope:        &scope,
	}
	resp2, err := mem.Search(req2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Query: 'documentation' (scope: mv2://docs/)\n")
	fmt.Printf("   Total hits: %d\n", resp2.TotalHits)
	fmt.Println()

	// ========================================
	// 5. TIMELINE - Browse documents chronologically
	// ========================================
	fmt.Println("5. Timeline (chronological view):")
	timeline, err := mem.Timeline(memvid.TimelineQuery{})
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range timeline {
		uri := entry.URI
		if uri == "" {
			uri = "(no uri)"
		}

		preview := entry.Preview
		if len(preview) > 40 {
			preview = preview[:40]
		}

		fmt.Printf("   [%d] %s - %s\n", entry.FrameID, uri, preview)
	}
	fmt.Println()

	// ========================================
	// 6. REOPEN - Close and reopen the memory
	// ========================================
	fmt.Println("6. Closing and reopening memory...")
	mem.Close() // Explicitly close to release lock

	reopened, err := memvid.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to reopen: %v", err)
	}

	statsReopened, err := reopened.Stats()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("   Reopened successfully!")
	fmt.Printf("   Frame count after reopen: %d\n", statsReopened.FrameCount)
	fmt.Println()

	// ========================================
	// 7. VERIFY - Check file integrity
	// ========================================
	fmt.Println("7. Verifying file integrity...")
	reopened.Close() // Close the memory before verifying to release lock

	report, err := memvid.Verify(dbPath, false)
	if err != nil {
		log.Fatalf("Verify failed: %v", err)
	}

	fmt.Printf("   Verification status: %s\n", report.OverallStatus)
	fmt.Println()

	fmt.Println("=== Example completed successfully! ===")
}
