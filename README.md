# gofar

**Go Far Archive Extraction Tool**

`gofar` is a Go-based command-line tool and library for extracting files from `.far` archives. This proprietary archive format was notably used by Maxis in early games like *The Sims 1* and *The Sims Online*.

The implementation supports reading and extracting from **Version 1A**, **Version 1B**, and **Version 3** archives.

## Features

* **CLI Extraction:** Easily extract all contents of a `.far` file to a specified directory.
* **Version Auto-detection:** Automatically differentiates between Version 1A (4-byte filename length) and Version 1B (2-byte filename length) archives.
* **Reusable Library:** Includes a clean `far` Go package that accepts standard `io.ReadSeeker` interfaces for seamless integration into other Go projects.

## Installation

Make sure you have Go installed, then clone the repository and build the binary:

```bash
git clone <repository_url>
cd gofar
go build -o gofar main.go
```

## CLI Usage

```bash
./gofar <file.far> [output_dir]
```

* `<file.far>`: The path to the FAR archive you want to extract.
* `[output_dir]`: (Optional) The directory to extract the files into. Defaults to `extracted`.

**Example:**
```bash
./gofar samples/Turkey.far extracted_turkey
```

## Library Usage

You can also use `gofar` programmatically in your own Go applications:

```go
package main

import (
	"fmt"
	"log"
	"gofar/far"
)

func main() {
	// Open the FAR archive
	reader, err := far.Open("example.far")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	fmt.Printf("Detected FAR Version: %s\n", reader.Version)
	fmt.Printf("Total files: %d\n", len(reader.Entries))

	// Extract all files to a directory
	for _, entry := range reader.Entries {
		err := reader.Extract(entry, "output_directory")
		if err != nil {
			log.Printf("Failed to extract %s: %v", entry.Filename, err)
		}
	}
}
```

## Credits

Based on code on the following GitHub repo: https://github.com/FaithBeam/Sims.Far
