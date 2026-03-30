package repository_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func TestWriteLooseBytesContent(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		repo := repoHarness.OpenRepository(t)

		content := []byte("write-loose-bytes-content\n")

		gotID, err := repo.Objects().WriteBytesContent(objecttype.TypeBlob, content)
		if err != nil {
			t.Fatalf("WriteLooseBytesContent: %v", err)
		}

		wantID := repoHarness.HashObject(t, "blob", content)
		if gotID != wantID {
			t.Fatalf("WriteLooseBytesContent id = %s, want %s", gotID, wantID)
		}

		ty, gotContent, err := repo.Objects().ReadBytesContent(gotID)
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

		repo := repoHarness.OpenRepository(t)

		content := []byte("write-loose-reader-content\n")

		gotID, err := repo.Objects().WriteReaderContent(objecttype.TypeBlob, int64(len(content)), bytes.NewReader(content))
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

		repo := repoHarness.OpenRepository(t)

		raw, err := repo.Objects().ReadBytesFull(commitID)
		if err != nil {
			t.Fatalf("ReadStoredBytesFull: %v", err)
		}

		idFromBytes, err := repo.Objects().WriteBytesFull(raw)
		if err != nil {
			t.Fatalf("WriteLooseBytesFull: %v", err)
		}

		if idFromBytes != commitID {
			t.Fatalf("WriteLooseBytesFull id = %s, want %s", idFromBytes, commitID)
		}

		idFromReader, err := repo.Objects().WriteReaderFull(bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("WriteLooseReaderFull: %v", err)
		}

		if idFromReader != commitID {
			t.Fatalf("WriteLooseReaderFull id = %s, want %s", idFromReader, commitID)
		}
	})
}
