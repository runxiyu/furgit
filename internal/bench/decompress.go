package main

import (
	"fmt"
	"io"
	"os"

	"git.sr.ht/~runxiyu/furgit/internal/zlib"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.zlib> <output>\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	in, err := os.Open(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer in.Close()

	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	reader, err := zlib.NewReader(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating zlib reader: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	_, err = io.Copy(out, reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decompressing data: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Successfully decompressed %s to %s\n", inputFile, outputFile)
}
