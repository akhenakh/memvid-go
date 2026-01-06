package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/akhenakh/memvid-go"
)

func main() {
	// Create a temporary directory for our memory file
	tempDir, err := os.MkdirTemp("", "memvid-pdf-example")
	if err != nil {
		log.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // clean up

	mv2Path := filepath.Join(tempDir, "paper.mv2")
	pdfPath := filepath.Join(tempDir, "1706.03762v7.pdf")

	fmt.Println("=== Memvid PDF Ingestion Example ===")

	// DOWNLOAD the PDF file (if not exists)
	// Since we are not inside the Rust repo, let's download the paper for the example
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		fmt.Println("Downloading Attention Is All You Need paper...")
		if err := downloadFile(pdfPath, "https://arxiv.org/pdf/1706.03762.pdf"); err != nil {
			log.Fatalf("Failed to download PDF: %v", err)
		}
	}

	// CREATE a new memory file
	fmt.Println("Creating memory file...")
	mem, err := memvid.Create(mv2Path)
	if err != nil {
		log.Fatalf("Failed to create memory: %v", err)
	}
	// Note: Explicit close handled later

	// Enable lexical index for text search
	if err := mem.EnableLex(); err != nil {
		log.Printf("Warning: failed to enable lex index: %v", err)
	}

	fmt.Printf("   Memory created at %s\n\n", mv2Path)

	// INGEST the PDF file
	fmt.Printf("Ingesting PDF: %s\n", pdfPath)

	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		log.Fatalf("Failed to read PDF file: %v", err)
	}
	fmt.Printf("   PDF size: %d bytes\n", len(pdfBytes))

	// Put the PDF with metadata
	opts := memvid.PutOptions{
		Title: "Attention Is All You Need",
		URI:   "mv2://papers/transformer.pdf",
		ExtraMetadata: map[string]string{
			"author": "Vaswani et al.",
			"year":   "2017",
			"topic":  "transformer",
		},
		AutoTag: true, // Let Memvid extract tags automatically
	}

	frameID, err := mem.PutBytes(pdfBytes, opts)
	if err != nil {
		log.Fatalf("Ingestion failed: %v", err)
	}
	fmt.Printf("   Ingested as frame: %d\n", frameID)

	// Commit changes
	if err := mem.Commit(); err != nil {
		log.Fatalf("Commit failed: %v", err)
	}
	fmt.Println("   Committed successfully!")

	// CHECK memory statistics
	fmt.Println("Memory statistics:")
	stats, err := mem.Stats()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Frame count: %d\n", stats.FrameCount)
	fmt.Printf("   Has lexical index: %v\n", stats.HasLexIndex)
	fmt.Println()

	// SEARCH the ingested PDF
	fmt.Println("Searching the paper...")

	queries := []string{
		"attention mechanism",
		"transformer architecture",
		"self-attention",
		"encoder decoder",
		"positional encoding",
	}

	for _, query := range queries {
		req := memvid.SearchRequest{
			Query:        query,
			TopK:         3,
			SnippetChars: 150,
			NoSketch:     false,
		}

		resp, err := mem.Search(req)
		if err != nil {
			log.Printf("Search failed for '%s': %v", query, err)
			continue
		}

		fmt.Printf("   Query: '%s'\n", query)
		fmt.Printf("   Hits: %d (%dms)\n", resp.TotalHits, resp.ElapsedMs)

		for i, hit := range resp.Hits {
			if i >= 2 {
				break
			}

			// Clean up snippet for display (remove newlines)
			snippet := strings.ReplaceAll(hit.Text, "\n", " ")
			if len(snippet) > 100 {
				snippet = snippet[:100]
			}

			fmt.Printf("   %d. %s...\n", i+1, snippet)
		}
		fmt.Println()
	}

	// VERIFY file integrity
	fmt.Println("Verifying file integrity...")
	mem.Close() // Close before verify to release lock

	report, err := memvid.Verify(mv2Path, false)
	if err != nil {
		log.Fatalf("Verify failed: %v", err)
	}
	fmt.Printf("   Status: %s\n", report.OverallStatus)
	fmt.Println()

	fmt.Println("=== PDF ingestion example completed! ===")
}

// Helper to download a file
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
