package loose_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

func TestLooseStoreWriteBytesContentAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
		store := openLooseStore(t, testRepo.Dir(), algo)

		content := []byte("written-by-loose-store\n")
		expectedHex := testRepo.RunInput(t, content, "hash-object", "-t", "blob", "--stdin")
		expectedID, err := objectid.ParseHex(algo, expectedHex)
		if err != nil {
			t.Fatalf("ParseHex(expected): %v", err)
		}

		writtenID, err := store.WriteBytesContent(objecttype.TypeBlob, content)
		if err != nil {
			t.Fatalf("WriteBytesContent: %v", err)
		}
		if writtenID != expectedID {
			t.Fatalf("WriteBytesContent id = %s, want %s", writtenID, expectedID)
		}

		gotBody := testRepo.CatFile(t, "blob", writtenID)
		if !bytes.Equal(gotBody, content) {
			t.Fatalf("git cat-file body mismatch")
		}

		writtenID2, err := store.WriteBytesContent(objecttype.TypeBlob, content)
		if err != nil {
			t.Fatalf("WriteBytesContent second write: %v", err)
		}
		if writtenID2 != expectedID {
			t.Fatalf("WriteBytesContent second id = %s, want %s", writtenID2, expectedID)
		}
	})
}

func TestLooseStoreWriteBytesFullAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
		store := openLooseStore(t, testRepo.Dir(), algo)

		body := []byte("full-write-body\n")
		header, ok := objectheader.Encode(objecttype.TypeBlob, int64(len(body)))
		if !ok {
			t.Fatalf("objectheader.Encode failed")
		}
		raw := make([]byte, len(header)+len(body))
		copy(raw, header)
		copy(raw[len(header):], body)

		wantID := algo.Sum(raw)
		gotID, err := store.WriteBytesFull(raw)
		if err != nil {
			t.Fatalf("WriteBytesFull: %v", err)
		}
		if gotID != wantID {
			t.Fatalf("WriteBytesFull id = %s, want %s", gotID, wantID)
		}

		gotBody := testRepo.CatFile(t, "blob", gotID)
		if !bytes.Equal(gotBody, body) {
			t.Fatalf("git cat-file body mismatch")
		}
	})
}

func TestLooseStoreWriteValidationErrors(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
		store := openLooseStore(t, testRepo.Dir(), algo)

		if _, err := store.WriteBytesFull([]byte("blob 1\x00hello")); err == nil {
			t.Fatalf("WriteBytesFull expected size/content mismatch error")
		}
		if _, err := store.WriteBytesFull([]byte("not-a-header")); err == nil {
			t.Fatalf("WriteBytesFull expected malformed header error")
		}
		if _, err := store.WriteBytesContent(objecttype.TypeInvalid, []byte("x")); err == nil {
			t.Fatalf("WriteBytesContent expected invalid type error")
		}
	})
}
