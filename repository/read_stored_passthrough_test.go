package repository_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
	"codeberg.org/lindenii/furgit/repository"
)

func TestReadStoredPassThroughs(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, _, commitID := repoHarness.MakeCommit(t, "pass-through")

		root, err := os.OpenRoot(repoHarness.Dir())
		if err != nil {
			t.Fatalf("os.OpenRoot: %v", err)
		}
		defer func() { _ = root.Close() }()

		repo, err := repository.Open(root)
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}
		defer func() { _ = repo.Close() }()

		headerTy, headerSize, err := repo.ReadStoredHeader(commitID)
		if err != nil {
			t.Fatalf("ReadStoredHeader: %v", err)
		}
		if headerTy != objecttype.TypeCommit {
			t.Fatalf("ReadStoredHeader type = %v, want %v", headerTy, objecttype.TypeCommit)
		}
		if headerSize <= 0 {
			t.Fatalf("ReadStoredHeader size = %d, want > 0", headerSize)
		}

		full, err := repo.ReadStoredBytesFull(commitID)
		if err != nil {
			t.Fatalf("ReadStoredBytesFull: %v", err)
		}
		if len(full) == 0 {
			t.Fatalf("ReadStoredBytesFull returned empty payload")
		}

		contentTy, content, err := repo.ReadStoredBytesContent(commitID)
		if err != nil {
			t.Fatalf("ReadStoredBytesContent: %v", err)
		}
		if contentTy != objecttype.TypeCommit {
			t.Fatalf("ReadStoredBytesContent type = %v, want %v", contentTy, objecttype.TypeCommit)
		}
		if len(content) == 0 {
			t.Fatalf("ReadStoredBytesContent returned empty content")
		}

		fullReader, err := repo.ReadStoredReaderFull(commitID)
		if err != nil {
			t.Fatalf("ReadStoredReaderFull: %v", err)
		}
		fullReaderBytes, readErr := io.ReadAll(fullReader)
		closeErr := fullReader.Close()
		if readErr != nil {
			t.Fatalf("ReadStoredReaderFull read: %v", readErr)
		}
		if closeErr != nil {
			t.Fatalf("ReadStoredReaderFull close: %v", closeErr)
		}
		if !bytes.Equal(fullReaderBytes, full) {
			t.Fatalf("ReadStoredReaderFull bytes mismatch against ReadStoredBytesFull")
		}

		readerTy, readerSize, contentReader, err := repo.ReadStoredReaderContent(commitID)
		if err != nil {
			t.Fatalf("ReadStoredReaderContent: %v", err)
		}
		if readerTy != objecttype.TypeCommit {
			t.Fatalf("ReadStoredReaderContent type = %v, want %v", readerTy, objecttype.TypeCommit)
		}
		if readerSize != int64(len(content)) {
			t.Fatalf("ReadStoredReaderContent size = %d, want %d", readerSize, len(content))
		}
		readerContentBytes, readErr := io.ReadAll(contentReader)
		closeErr = contentReader.Close()
		if readErr != nil {
			t.Fatalf("ReadStoredReaderContent read: %v", readErr)
		}
		if closeErr != nil {
			t.Fatalf("ReadStoredReaderContent close: %v", closeErr)
		}
		if !bytes.Equal(readerContentBytes, content) {
			t.Fatalf("ReadStoredReaderContent bytes mismatch against ReadStoredBytesContent")
		}
	})
}
