package loose_test

import (
	"bytes"
	"io"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

func TestLooseStoreWriteWriterContentAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
		store := openLooseStore(t, testRepo.Dir(), algo)

		content := []byte("written-by-content-writer\n")
		expectedHex := testRepo.RunInput(t, content, "hash-object", "-t", "blob", "--stdin")
		expectedID, err := objectid.ParseHex(algo, expectedHex)
		if err != nil {
			t.Fatalf("ParseHex(expected): %v", err)
		}

		writer, finalize, err := store.WriteWriterContent(objecttype.TypeBlob, int64(len(content)))
		if err != nil {
			t.Fatalf("WriteWriterContent: %v", err)
		}
		if _, err := io.Copy(writer, bytes.NewReader(content)); err != nil {
			t.Fatalf("WriteWriterContent write: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("WriteWriterContent close: %v", err)
		}
		writtenID, err := finalize()
		if err != nil {
			t.Fatalf("WriteWriterContent finalize: %v", err)
		}
		if writtenID != expectedID {
			t.Fatalf("WriteWriterContent id = %s, want %s", writtenID, expectedID)
		}

			gotBody := testRepo.CatFile(t, "blob", writtenID)
			if !bytes.Equal(gotBody, content) {
				t.Fatalf("git cat-file body mismatch")
			}

			// Writing the same object again should succeed and return the same ID.
			writer, finalize, err = store.WriteWriterContent(objecttype.TypeBlob, int64(len(content)))
			if err != nil {
				t.Fatalf("WriteWriterContent second: %v", err)
			}
			if _, err := io.Copy(writer, bytes.NewReader(content)); err != nil {
				t.Fatalf("WriteWriterContent second write: %v", err)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("WriteWriterContent second close: %v", err)
			}
			writtenID2, err := finalize()
			if err != nil {
				t.Fatalf("WriteWriterContent second finalize: %v", err)
			}
			if writtenID2 != expectedID {
				t.Fatalf("WriteWriterContent second id = %s, want %s", writtenID2, expectedID)
			}
		})
}

func TestLooseStoreWriteWriterFullAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
		store := openLooseStore(t, testRepo.Dir(), algo)

		body := []byte("full-writer-body\n")
		header, ok := objectheader.Encode(objecttype.TypeBlob, int64(len(body)))
		if !ok {
			t.Fatalf("objectheader.Encode failed")
		}
		raw := make([]byte, len(header)+len(body))
		copy(raw, header)
		copy(raw[len(header):], body)

		wantID := algo.Sum(raw)
		writer, finalize, err := store.WriteWriterFull()
		if err != nil {
			t.Fatalf("WriteWriterFull: %v", err)
		}
		if _, err := io.Copy(writer, bytes.NewReader(raw)); err != nil {
			t.Fatalf("WriteWriterFull write: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("WriteWriterFull close: %v", err)
		}
		gotID, err := finalize()
		if err != nil {
			t.Fatalf("WriteWriterFull finalize: %v", err)
		}
		if gotID != wantID {
			t.Fatalf("WriteWriterFull id = %s, want %s", gotID, wantID)
		}

		gotBody := testRepo.CatFile(t, "blob", gotID)
		if !bytes.Equal(gotBody, body) {
			t.Fatalf("git cat-file body mismatch")
		}
	})
}

func TestLooseStoreWriterValidationErrors(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
		store := openLooseStore(t, testRepo.Dir(), algo)

		t.Run("content overflow", func(t *testing.T) {
			writer, finalize, err := store.WriteWriterContent(objecttype.TypeBlob, 1)
			if err != nil {
				t.Fatalf("WriteWriterContent: %v", err)
			}
			if _, err := writer.Write([]byte("hello")); err == nil {
				t.Fatalf("expected overflow error")
			}
			_ = writer.Close()
			if _, err := finalize(); err == nil {
				t.Fatalf("expected finalize error after overflow")
			}
		})

		t.Run("content short", func(t *testing.T) {
			writer, finalize, err := store.WriteWriterContent(objecttype.TypeBlob, 5)
			if err != nil {
				t.Fatalf("WriteWriterContent: %v", err)
			}
			if _, err := writer.Write([]byte("x")); err != nil {
				t.Fatalf("write short: %v", err)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("close short: %v", err)
			}
			if _, err := finalize(); err == nil {
				t.Fatalf("expected finalize error for short content")
			}
		})

		t.Run("full malformed header", func(t *testing.T) {
			writer, finalize, err := store.WriteWriterFull()
			if err != nil {
				t.Fatalf("WriteWriterFull: %v", err)
			}
			if _, err := writer.Write([]byte("not-a-header")); err != nil {
				t.Fatalf("write malformed header bytes unexpectedly failed: %v", err)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("close malformed header: %v", err)
			}
			if _, err := finalize(); err == nil {
				t.Fatalf("expected finalize error for malformed header")
			}
		})

		t.Run("full size mismatch", func(t *testing.T) {
			writer, finalize, err := store.WriteWriterFull()
			if err != nil {
				t.Fatalf("WriteWriterFull: %v", err)
			}
			raw := []byte("blob 1\x00hello")
			if _, err := io.Copy(writer, bytes.NewReader(raw)); err == nil {
				t.Fatalf("expected overflow error")
			}
			_ = writer.Close()
			if _, err := finalize(); err == nil {
				t.Fatalf("expected finalize error after mismatch")
			}
		})
	})
}
