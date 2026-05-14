package main

import (
	"fmt"
	"os"

	"gofar/far"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gofar <file.far> [output_dir]")
		return
	}

	farPath := os.Args[1]
	destDir := "extracted"
	if len(os.Args) > 2 {
		destDir = os.Args[2]
	}

	reader, err := far.Open(farPath)
	if err != nil {
		fmt.Printf("Error opening FAR: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	fmt.Printf("FAR Version: %s, Files: %d\n", reader.Version, len(reader.Entries))

	for _, entry := range reader.Entries {
		fmt.Printf("Extracting %s (%d bytes)...\n", entry.Filename, entry.DecompressedSize)
		if err := reader.Extract(entry, destDir); err != nil {
			fmt.Printf("Error extracting %s: %v\n", entry.Filename, err)
		}
	}
	fmt.Println("Done.")
}
