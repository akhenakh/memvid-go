// main.go
package main

import (
	"fmt"
	"log"

	"github.com/akhenakh/memvid-go"
)

func main() {
	mem, err := memvid.Create("test.mv2")
	if err != nil {
		log.Fatal(err)
	}
	defer mem.Close()

	mem.EnableLex()

	opts := memvid.PutOptions{
		Title:   "Hello World",
		AutoTag: true,
	}
	seq, err := mem.PutBytes([]byte("This is some content for memvid."), opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted frame: %d\n", seq)

	mem.Commit()

	res, err := mem.Search(memvid.SearchRequest{
		Query:        "memvid",
		TopK:         5,
		SnippetChars: 100,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d hits\n", res.TotalHits)
	for _, hit := range res.Hits {
		fmt.Printf("- %s (Score: %.2f)\n", hit.Text, hit.Score)
	}
}
