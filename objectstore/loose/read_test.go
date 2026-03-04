package loose_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objectstore/loose"
)

func TestLooseStoreReadAgainstGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		blobID := testRepo.HashObject(t, "blob", []byte("blob body\n"))
		_, treeID, commitID := testRepo.MakeCommit(t, "subject\n\nbody")
		tagID := testRepo.TagAnnotated(t, "v1", commitID, "tag message")

		store := openLooseStore(t, testRepo.Dir(), algo)

		tests := []struct {
			name string
			id   objectid.ObjectID
		}{
			{name: "blob", id: blobID},
			{name: "tree", id: treeID},
			{name: "commit", id: commitID},
			{name: "tag", id: tagID},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				wantType, wantBody, wantRaw := expectedRawObject(t, testRepo, tt.id)

				gotRaw, err := store.ReadBytesFull(tt.id)
				if err != nil {
					t.Fatalf("ReadBytesFull: %v", err)
				}

				if !bytes.Equal(gotRaw, wantRaw) {
					t.Fatalf("ReadBytesFull mismatch")
				}

				gotType, gotBody, err := store.ReadBytesContent(tt.id)
				if err != nil {
					t.Fatalf("ReadBytesContent: %v", err)
				}

				if gotType != wantType {
					t.Fatalf("ReadBytesContent type = %v, want %v", gotType, wantType)
				}

				if !bytes.Equal(gotBody, wantBody) {
					t.Fatalf("ReadBytesContent body mismatch")
				}

				headType, headSize, err := store.ReadHeader(tt.id)
				if err != nil {
					t.Fatalf("ReadHeader: %v", err)
				}

				if headType != wantType {
					t.Fatalf("ReadHeader type = %v, want %v", headType, wantType)
				}

				if headSize != int64(len(wantBody)) {
					t.Fatalf("ReadHeader size = %d, want %d", headSize, len(wantBody))
				}

				fullReader, err := store.ReadReaderFull(tt.id)
				if err != nil {
					t.Fatalf("ReadReaderFull: %v", err)
				}

				got := mustReadAllAndClose(t, fullReader)
				if !bytes.Equal(got, wantRaw) {
					t.Fatalf("ReadReaderFull stream mismatch")
				}

				contentType, contentSize, contentReader, err := store.ReadReaderContent(tt.id)
				if err != nil {
					t.Fatalf("ReadReaderContent: %v", err)
				}

				if contentType != wantType {
					t.Fatalf("ReadReaderContent type = %v, want %v", contentType, wantType)
				}

				if contentSize != int64(len(wantBody)) {
					t.Fatalf("ReadReaderContent size = %d, want %d", contentSize, len(wantBody))
				}

				got = mustReadAllAndClose(t, contentReader)
				if !bytes.Equal(got, wantBody) {
					t.Fatalf("ReadReaderContent stream mismatch")
				}
			})
		}
	})
}

func TestLooseStoreErrors(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		store := openLooseStore(t, testRepo.Dir(), algo)

		notFoundID, err := objectid.ParseHex(algo, strings.Repeat("0", algo.HexLen()))
		if err != nil {
			t.Fatalf("ParseHex(notFoundID): %v", err)
		}

		_, err = store.ReadBytesFull(notFoundID)
		if !errors.Is(err, objectstore.ErrObjectNotFound) {
			t.Fatalf("ReadBytesFull not-found error = %v", err)
		}

		_, _, err = store.ReadBytesContent(notFoundID)
		if !errors.Is(err, objectstore.ErrObjectNotFound) {
			t.Fatalf("ReadBytesContent not-found error = %v", err)
		}

		_, err = store.ReadReaderFull(notFoundID)
		if !errors.Is(err, objectstore.ErrObjectNotFound) {
			t.Fatalf("ReadReaderFull not-found error = %v", err)
		}

		_, _, _, err = store.ReadReaderContent(notFoundID)
		if !errors.Is(err, objectstore.ErrObjectNotFound) {
			t.Fatalf("ReadReaderContent not-found error = %v", err)
		}

		_, _, err = store.ReadHeader(notFoundID)
		if !errors.Is(err, objectstore.ErrObjectNotFound) {
			t.Fatalf("ReadHeader not-found error = %v", err)
		}

		var otherAlgo objectid.Algorithm
		if algo == objectid.AlgorithmSHA1 {
			otherAlgo = objectid.AlgorithmSHA256
		} else {
			otherAlgo = objectid.AlgorithmSHA1
		}

		otherID, err := objectid.ParseHex(otherAlgo, strings.Repeat("1", otherAlgo.HexLen()))
		if err != nil {
			t.Fatalf("ParseHex(otherID): %v", err)
		}

		_, err = store.ReadBytesFull(otherID)
		if err == nil || !strings.Contains(err.Error(), "algorithm mismatch") {
			t.Fatalf("ReadBytesFull algorithm-mismatch error = %v", err)
		}
	})
}

func TestLooseStoreNewValidation(t *testing.T) {
	t.Parallel()

	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}

	defer func() { _ = root.Close() }()

	_, err = loose.New(root, objectid.AlgorithmUnknown)
	if err == nil {
		t.Fatalf("loose.New(root, unknown) expected error")
	}
}
