package loose_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object/header"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

func TestLooseStoreWriteReaderContentAgainstGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		store := openLooseStore(t, testRepo, algo)

		content := []byte("written-by-content-reader\n")
		expectedHex := testRepo.RunInput(t, content, "hash-object", "-t", "blob", "--stdin")

		expectedID, err := objectid.ParseHex(algo, expectedHex)
		if err != nil {
			t.Fatalf("ParseHex(expected): %v", err)
		}

		writtenID, err := store.WriteReaderContent(objecttype.TypeBlob, int64(len(content)), bytes.NewReader(content))
		if err != nil {
			t.Fatalf("WriteReaderContent: %v", err)
		}

		if writtenID != expectedID {
			t.Fatalf("WriteReaderContent id = %s, want %s", writtenID, expectedID)
		}

		gotBody := testRepo.CatFile(t, "blob", writtenID)
		if !bytes.Equal(gotBody, content) {
			t.Fatalf("git cat-file body mismatch")
		}

		// Writing the same object again should succeed and return the same ID.
		writtenID2, err := store.WriteReaderContent(objecttype.TypeBlob, int64(len(content)), bytes.NewReader(content))
		if err != nil {
			t.Fatalf("WriteReaderContent second: %v", err)
		}

		if writtenID2 != expectedID {
			t.Fatalf("WriteReaderContent second id = %s, want %s", writtenID2, expectedID)
		}
	})
}

func TestLooseStoreWriteReaderFullAgainstGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		store := openLooseStore(t, testRepo, algo)

		body := []byte("full-reader-body\n")

		header, ok := header.Encode(objecttype.TypeBlob, int64(len(body)))
		if !ok {
			t.Fatalf("header.Encode failed")
		}

		raw := make([]byte, len(header)+len(body))
		copy(raw, header)
		copy(raw[len(header):], body)

		wantID := algo.Sum(raw)

		gotID, err := store.WriteReaderFull(bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("WriteReaderFull: %v", err)
		}

		if gotID != wantID {
			t.Fatalf("WriteReaderFull id = %s, want %s", gotID, wantID)
		}

		gotBody := testRepo.CatFile(t, "blob", gotID)
		if !bytes.Equal(gotBody, body) {
			t.Fatalf("git cat-file body mismatch")
		}
	})
}

func TestLooseStoreReaderValidationErrors(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		t.Run("content overflow", func(t *testing.T) {
			t.Parallel()
			testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
			store := openLooseStore(t, testRepo, algo)

			_, err := store.WriteReaderContent(objecttype.TypeBlob, 1, bytes.NewReader([]byte("hello")))
			if err == nil {
				t.Fatalf("expected error after overflow")
			}
		})

		t.Run("content short", func(t *testing.T) {
			t.Parallel()
			testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
			store := openLooseStore(t, testRepo, algo)

			_, err := store.WriteReaderContent(objecttype.TypeBlob, 5, bytes.NewReader([]byte("x")))
			if err == nil {
				t.Fatalf("expected error for short content")
			}
		})

		t.Run("full malformed header", func(t *testing.T) {
			t.Parallel()
			testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
			store := openLooseStore(t, testRepo, algo)

			_, err := store.WriteReaderFull(bytes.NewReader([]byte("not-a-header")))
			if err == nil {
				t.Fatalf("expected error for malformed header")
			}
		})

		t.Run("full size mismatch", func(t *testing.T) {
			t.Parallel()
			testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
			store := openLooseStore(t, testRepo, algo)

			raw := []byte("blob 1\x00hello")

			_, err := store.WriteReaderFull(bytes.NewReader(raw))
			if err == nil {
				t.Fatalf("expected error after mismatch")
			}
		})
	})
}
