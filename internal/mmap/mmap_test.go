//go:build unix

package mmap_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"lindenii.org/go/furgit/internal/mmap"
)

// makeTempFileWithContents creates one file with content and opens it for reading.
func makeTempFileWithContents(t *testing.T, content []byte) *os.File {
	t.Helper()

	path := filepath.Join(t.TempDir(), "data")

	err := os.WriteFile(path, content, 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	file, err := os.Open(path) //nolint:gosec
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(func() { _ = file.Close() })

	return file
}

// makeLargeTempFileLongerThanPage creates and opens one file
// whose content repeats seed
// until it spans more than one OS page.
func makeLargeTempFileLongerThanPage(t *testing.T, seed []byte) (*os.File, []byte) {
	t.Helper()

	content := bytes.Repeat(seed, os.Getpagesize()/len(seed)+2)

	return makeTempFileWithContents(t, content), content
}

func TestOpenData(t *testing.T) {
	t.Parallel()

	file, content := makeLargeTempFileLongerThanPage(t, []byte("mapped bytes\n"))

	mapping, err := mmap.Open(file)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	// The mapping must survive closing the originating file.
	err = file.Close()
	if err != nil {
		t.Fatalf("file Close: %v", err)
	}

	if !bytes.Equal(mapping.Data(), content) {
		t.Fatalf("Data mismatch")
	}

	err = mapping.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestOpenEmpty(t *testing.T) {
	t.Parallel()

	file := makeTempFileWithContents(t, nil)

	mapping, err := mmap.Open(file)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if len(mapping.Data()) != 0 {
		t.Fatalf("Data = %d bytes, want empty", len(mapping.Data()))
	}

	err = mapping.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestCloseIdempotent(t *testing.T) {
	t.Parallel()

	file, _ := makeLargeTempFileLongerThanPage(t, []byte("once"))

	mapping, err := mmap.Open(file)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	err = mapping.Close()
	if err != nil {
		t.Fatalf("first Close: %v", err)
	}

	err = mapping.Close()
	if err != nil {
		t.Fatalf("second Close: %v", err)
	}
}
