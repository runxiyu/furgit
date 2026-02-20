package loose_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objectstore/loose"
	"codeberg.org/lindenii/furgit/objecttype"
)

func openLooseStore(t *testing.T, repoPath string, algo objectid.Algorithm) *loose.Store {
	t.Helper()
	objectsPath := filepath.Join(repoPath, "objects")
	root, err := os.OpenRoot(objectsPath)
	if err != nil {
		t.Fatalf("OpenRoot(%q): %v", objectsPath, err)
	}
	t.Cleanup(func() { _ = root.Close() })

	store, err := loose.New(root, algo)
	if err != nil {
		t.Fatalf("loose.New: %v", err)
	}
	return store
}

func mustReadAllAndClose(t *testing.T, reader io.ReadCloser) []byte {
	t.Helper()
	data, err := io.ReadAll(reader)
	if err != nil {
		_ = reader.Close()
		t.Fatalf("ReadAll: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return data
}

func expectedRawObject(t *testing.T, testRepo *testgit.TestRepo, id objectid.ObjectID) (objecttype.Type, []byte, []byte) {
	t.Helper()

	typeName := testRepo.Run(t, "cat-file", "-t", id.String())
	ty, ok := objecttype.ParseName(typeName)
	if !ok {
		t.Fatalf("ParseName(%q) failed", typeName)
	}
	body := testRepo.CatFile(t, typeName, id)
	header, ok := objectheader.Encode(ty, int64(len(body)))
	if !ok {
		t.Fatalf("objectheader.Encode failed")
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return ty, body, raw
}

func TestLooseStoreReadAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
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
				if got := mustReadAllAndClose(t, fullReader); !bytes.Equal(got, wantRaw) {
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
				if got := mustReadAllAndClose(t, contentReader); !bytes.Equal(got, wantBody) {
					t.Fatalf("ReadReaderContent stream mismatch")
				}
			})
		}
	})
}

func TestLooseStoreErrors(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
		store := openLooseStore(t, testRepo.Dir(), algo)

		notFoundID, err := objectid.ParseHex(algo, strings.Repeat("0", algo.HexLen()))
		if err != nil {
			t.Fatalf("ParseHex(notFoundID): %v", err)
		}

		if _, err := store.ReadBytesFull(notFoundID); !errors.Is(err, objectstore.ErrObjectNotFound) {
			t.Fatalf("ReadBytesFull not-found error = %v", err)
		}
		if _, _, err := store.ReadBytesContent(notFoundID); !errors.Is(err, objectstore.ErrObjectNotFound) {
			t.Fatalf("ReadBytesContent not-found error = %v", err)
		}
		if _, err := store.ReadReaderFull(notFoundID); !errors.Is(err, objectstore.ErrObjectNotFound) {
			t.Fatalf("ReadReaderFull not-found error = %v", err)
		}
		if _, _, _, err := store.ReadReaderContent(notFoundID); !errors.Is(err, objectstore.ErrObjectNotFound) {
			t.Fatalf("ReadReaderContent not-found error = %v", err)
		}
		if _, _, err := store.ReadHeader(notFoundID); !errors.Is(err, objectstore.ErrObjectNotFound) {
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

		if _, err := store.ReadBytesFull(otherID); err == nil || !strings.Contains(err.Error(), "algorithm mismatch") {
			t.Fatalf("ReadBytesFull algorithm-mismatch error = %v", err)
		}
	})
}

func TestLooseStoreNewValidation(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer func() { _ = root.Close() }()

	if _, err := loose.New(root, objectid.AlgorithmUnknown); err == nil {
		t.Fatalf("loose.New(root, unknown) expected error")
	}
}
