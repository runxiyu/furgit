package repository_test

import (
	"bytes"
	"io"
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

		repo, err := repository.Open(repoHarness.Dir())
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

func TestWriteLooseWriterContent(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		repo, err := repository.Open(repoHarness.Dir())
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}
		defer func() { _ = repo.Close() }()

		content := []byte("write-loose-writer-content\n")
		writer, finalize, err := repo.WriteLooseWriterContent(objecttype.TypeBlob, int64(len(content)))
		if err != nil {
			t.Fatalf("WriteLooseWriterContent: %v", err)
		}

		if _, err := writer.Write(content[:6]); err != nil {
			t.Fatalf("WriteLooseWriterContent first write: %v", err)
		}
		if _, err := writer.Write(content[6:]); err != nil {
			t.Fatalf("WriteLooseWriterContent second write: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("WriteLooseWriterContent close: %v", err)
		}
		gotID, err := finalize()
		if err != nil {
			t.Fatalf("WriteLooseWriterContent finalize: %v", err)
		}

		wantID := repoHarness.HashObject(t, "blob", content)
		if gotID != wantID {
			t.Fatalf("WriteLooseWriterContent id = %s, want %s", gotID, wantID)
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

		repo, err := repository.Open(repoHarness.Dir())
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

		writer, finalize, err := repo.WriteLooseWriterFull()
		if err != nil {
			t.Fatalf("WriteLooseWriterFull: %v", err)
		}
		if _, err := io.Copy(writer, bytes.NewReader(raw)); err != nil {
			t.Fatalf("WriteLooseWriterFull copy: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("WriteLooseWriterFull close: %v", err)
		}
		idFromWriter, err := finalize()
		if err != nil {
			t.Fatalf("WriteLooseWriterFull finalize: %v", err)
		}
		if idFromWriter != commitID {
			t.Fatalf("WriteLooseWriterFull id = %s, want %s", idFromWriter, commitID)
		}
	})
}
