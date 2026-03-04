package repository_test

import (
	"bytes"
	"os"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
	"codeberg.org/lindenii/furgit/repository"
)

func TestWriteLooseBytesContent(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

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

		content := []byte("write-loose-bytes-content\n")

		gotID, err := repo.WriteLooseBytesContent(objecttype.TypeBlob, content)
		if err != nil {
			t.Fatalf("WriteLooseBytesContent: %v", err)
		}

		wantID := repoHarness.HashObject(t, "blob", content)
		if gotID != wantID {
			t.Fatalf("WriteLooseBytesContent id = %s, want %s", gotID, wantID)
		}

		ty, gotContent, err := repo.ReadStoredBytesContent(gotID)
		if err != nil {
			t.Fatalf("ReadStoredBytesContent: %v", err)
		}

		if ty != objecttype.TypeBlob {
			t.Fatalf("ReadStoredBytesContent type = %v, want %v", ty, objecttype.TypeBlob)
		}

		if !bytes.Equal(gotContent, content) {
			t.Fatalf("ReadStoredBytesContent content mismatch")
		}
	})
}

func TestWriteLooseReaderContent(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

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

		content := []byte("write-loose-reader-content\n")

		gotID, err := repo.WriteLooseReaderContent(objecttype.TypeBlob, int64(len(content)), bytes.NewReader(content))
		if err != nil {
			t.Fatalf("WriteLooseReaderContent: %v", err)
		}

		wantID := repoHarness.HashObject(t, "blob", content)
		if gotID != wantID {
			t.Fatalf("WriteLooseReaderContent id = %s, want %s", gotID, wantID)
		}
	})
}

func TestWriteLooseFull(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})
		_, _, commitID := repoHarness.MakeCommit(t, "write-loose-full")

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

		raw, err := repo.ReadStoredBytesFull(commitID)
		if err != nil {
			t.Fatalf("ReadStoredBytesFull: %v", err)
		}

		idFromBytes, err := repo.WriteLooseBytesFull(raw)
		if err != nil {
			t.Fatalf("WriteLooseBytesFull: %v", err)
		}

		if idFromBytes != commitID {
			t.Fatalf("WriteLooseBytesFull id = %s, want %s", idFromBytes, commitID)
		}

		idFromReader, err := repo.WriteLooseReaderFull(bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("WriteLooseReaderFull: %v", err)
		}

		if idFromReader != commitID {
			t.Fatalf("WriteLooseReaderFull id = %s, want %s", idFromReader, commitID)
		}
	})
}
