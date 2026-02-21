package packed_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objectstore/packed"
)

func TestPackedStoreReadAgainstGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo, ids := createPackedFixtureRepo(t, algo)
		store := openPackedStore(t, testRepo.Dir(), algo)

		for _, id := range ids {
			t.Run(id.String(), func(t *testing.T) {
				wantType, wantBody, wantRaw := expectedRawObject(t, testRepo, id)

				gotHeaderType, gotHeaderSize, err := store.ReadHeader(id)
				if err != nil {
					t.Fatalf("ReadHeader: %v", err)
				}
				if gotHeaderType != wantType {
					t.Fatalf("ReadHeader type = %v, want %v", gotHeaderType, wantType)
				}
				if gotHeaderSize != int64(len(wantBody)) {
					t.Fatalf("ReadHeader size = %d, want %d", gotHeaderSize, len(wantBody))
				}

				gotRaw, err := store.ReadBytesFull(id)
				if err != nil {
					t.Fatalf("ReadBytesFull: %v", err)
				}
				if !bytes.Equal(gotRaw, wantRaw) {
					t.Fatalf("ReadBytesFull mismatch")
				}

				gotType, gotBody, err := store.ReadBytesContent(id)
				if err != nil {
					t.Fatalf("ReadBytesContent: %v", err)
				}
				if gotType != wantType {
					t.Fatalf("ReadBytesContent type = %v, want %v", gotType, wantType)
				}
				if !bytes.Equal(gotBody, wantBody) {
					t.Fatalf("ReadBytesContent mismatch")
				}

				fullReader, err := store.ReadReaderFull(id)
				if err != nil {
					t.Fatalf("ReadReaderFull: %v", err)
				}
				if got := mustReadAllAndClose(t, fullReader); !bytes.Equal(got, wantRaw) {
					t.Fatalf("ReadReaderFull mismatch")
				}

				contentType, contentSize, contentReader, err := store.ReadReaderContent(id)
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
					t.Fatalf("ReadReaderContent mismatch")
				}
			})
		}
	})
}

func TestPackedStoreErrors(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo, _ := createPackedFixtureRepo(t, algo)
		store := openPackedStore(t, testRepo.Dir(), algo)

		notFoundID, err := objectid.ParseHex(algo, strings.Repeat("0", algo.HexLen()))
		if err != nil {
			t.Fatalf("ParseHex(notFound): %v", err)
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
		for _, candidate := range objectid.SupportedAlgorithms() {
			if candidate != algo {
				otherAlgo = candidate
				break
			}
		}
		if otherAlgo != objectid.AlgorithmUnknown {
			mismatchID, err := objectid.ParseHex(otherAlgo, strings.Repeat("0", otherAlgo.HexLen()))
			if err != nil {
				t.Fatalf("ParseHex(mismatch): %v", err)
			}
			if _, err := store.ReadBytesFull(mismatchID); err == nil || !strings.Contains(err.Error(), "algorithm mismatch") {
				t.Fatalf("ReadBytesFull algorithm-mismatch error = %v", err)
			}
		}
	})
}

func TestPackedStoreNewValidation(t *testing.T) {
	t.Parallel()
	testRepo, _ := createPackedFixtureRepo(t, objectid.AlgorithmSHA1)
	store := openPackedStore(t, testRepo.Dir(), objectid.AlgorithmSHA1)
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close second: %v", err)
	}
}

func TestPackedStoreInvalidAlgorithm(t *testing.T) {
	t.Parallel()
	testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectid.AlgorithmSHA1, Bare: true})
	root, err := os.OpenRoot(testRepo.Dir())
	if err != nil {
		t.Fatalf("OpenRoot(%q): %v", testRepo.Dir(), err)
	}
	t.Cleanup(func() { _ = root.Close() })

	if _, err := packed.New(root, objectid.AlgorithmUnknown); !errors.Is(err, objectid.ErrInvalidAlgorithm) {
		t.Fatalf("packed.New invalid algorithm error = %v", err)
	}
}

func TestPackedStoreReadHeaderUsesResolvedObjectSizeForDelta(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})

		var parent objectid.ObjectID
		for i := range 96 {
			content := strings.Repeat("common-line-"+strconv.Itoa(i%7)+"\n", 384) + fmt.Sprintf("tail-%03d\n", i)
			_, treeID := testRepo.MakeSingleFileTree(t, "file.txt", []byte(content))
			if i == 0 {
				parent = testRepo.CommitTree(t, treeID, "delta-header-size-0")
				continue
			}
			parent = testRepo.CommitTree(t, treeID, fmt.Sprintf("delta-header-size-%03d", i), parent)
		}
		testRepo.UpdateRef(t, "refs/heads/main", parent)
		testRepo.Repack(t, "-a", "-d", "-f", "--window=128", "--depth=128")

		deltaID, wantResolvedSize := findDeltaObjectWithResolvedSizeMismatch(t, testRepo, algo)
		store := openPackedStore(t, testRepo.Dir(), algo)

		_, gotSize, err := store.ReadHeader(deltaID)
		if err != nil {
			t.Fatalf("ReadHeader(%s): %v", deltaID, err)
		}
		if gotSize != wantResolvedSize {
			t.Fatalf("ReadHeader(%s) size = %d, want resolved size %d", deltaID, gotSize, wantResolvedSize)
		}
	})
}

func findDeltaObjectWithResolvedSizeMismatch(t *testing.T, testRepo *testgit.TestRepo, algo objectid.Algorithm) (objectid.ObjectID, int64) {
	t.Helper()

	idxFiles, err := filepath.Glob(filepath.Join(testRepo.Dir(), "objects", "pack", "*.idx"))
	if err != nil {
		t.Fatalf("Glob idx: %v", err)
	}
	if len(idxFiles) == 0 {
		t.Fatalf("no idx files found")
	}

	verifyOut := testRepo.Run(t, "verify-pack", "-v", idxFiles[0])
	for _, line := range strings.Split(strings.TrimSpace(verifyOut), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		idHex := fields[0]
		deltaStreamSize, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			continue
		}

		resolvedSizeStr := testRepo.Run(t, "cat-file", "-s", idHex)
		resolvedSize, err := strconv.ParseInt(strings.TrimSpace(resolvedSizeStr), 10, 64)
		if err != nil {
			t.Fatalf("parse cat-file size for %s: %v", idHex, err)
		}
		if deltaStreamSize == resolvedSize {
			continue
		}

		id, err := objectid.ParseHex(algo, idHex)
		if err != nil {
			t.Fatalf("ParseHex(%s): %v", idHex, err)
		}
		return id, resolvedSize
	}

	t.Fatalf("did not find a delta object with mismatched stream/resolved size")
	return objectid.ObjectID{}, 0
}
