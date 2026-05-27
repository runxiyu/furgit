package memory_test

import (
	"bytes"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	objectheader "lindenii.org/go/furgit/object/header"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store/memory"
	objecttype "lindenii.org/go/furgit/object/type"
)

func TestStoreWriteReaderContent(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		content := []byte("memory-content\n")

		gotID, err := store.WriteReaderContent(objecttype.TypeBlob, int64(len(content)), bytes.NewReader(content))
		if err != nil {
			t.Fatalf("WriteReaderContent: %v", err)
		}

		wantID := algo.Sum(buildRawObject(t, objecttype.TypeBlob, content))
		if gotID != wantID {
			t.Fatalf("WriteReaderContent id = %s, want %s", gotID, wantID)
		}

		gotType, gotContent, err := store.ReadBytesContent(gotID)
		if err != nil {
			t.Fatalf("ReadBytesContent: %v", err)
		}

		if gotType != objecttype.TypeBlob {
			t.Fatalf("ReadBytesContent type = %v, want %v", gotType, objecttype.TypeBlob)
		}

		if !bytes.Equal(gotContent, content) {
			t.Fatalf("ReadBytesContent content mismatch")
		}
	})
}

func TestStoreWriteReaderFull(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		content := []byte("memory-full\n")
		raw := buildRawObject(t, objecttype.TypeBlob, content)

		gotID, err := store.WriteReaderFull(bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("WriteReaderFull: %v", err)
		}

		wantID := algo.Sum(raw)
		if gotID != wantID {
			t.Fatalf("WriteReaderFull id = %s, want %s", gotID, wantID)
		}

		gotRaw, err := store.ReadBytesFull(gotID)
		if err != nil {
			t.Fatalf("ReadBytesFull: %v", err)
		}

		if !bytes.Equal(gotRaw, raw) {
			t.Fatalf("ReadBytesFull mismatch")
		}
	})
}

func TestStoreWriteBytes(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		content := []byte("memory-bytes\n")

		gotID, err := store.WriteBytesContent(objecttype.TypeBlob, content)
		if err != nil {
			t.Fatalf("WriteBytesContent: %v", err)
		}

		wantID := algo.Sum(buildRawObject(t, objecttype.TypeBlob, content))
		if gotID != wantID {
			t.Fatalf("WriteBytesContent id = %s, want %s", gotID, wantID)
		}

		raw := buildRawObject(t, objecttype.TypeBlob, content)

		gotID2, err := store.WriteBytesFull(raw)
		if err != nil {
			t.Fatalf("WriteBytesFull: %v", err)
		}

		if gotID2 != wantID {
			t.Fatalf("WriteBytesFull id = %s, want %s", gotID2, wantID)
		}
	})
}

func TestStoreWriteReaderValidationErrors(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		t.Run("content overflow", func(t *testing.T) {
			t.Parallel()

			store := memory.New(algo)

			_, err := store.WriteReaderContent(objecttype.TypeBlob, 1, bytes.NewReader([]byte("hello")))
			if err == nil {
				t.Fatalf("expected error after overflow")
			}
		})

		t.Run("content short", func(t *testing.T) {
			t.Parallel()

			store := memory.New(algo)

			_, err := store.WriteReaderContent(objecttype.TypeBlob, 5, bytes.NewReader([]byte("x")))
			if err == nil {
				t.Fatalf("expected error for short content")
			}
		})

		t.Run("full malformed header", func(t *testing.T) {
			t.Parallel()

			store := memory.New(algo)

			_, err := store.WriteReaderFull(bytes.NewReader([]byte("not-a-header")))
			if err == nil {
				t.Fatalf("expected error for malformed header")
			}
		})

		t.Run("full size mismatch", func(t *testing.T) {
			t.Parallel()

			store := memory.New(algo)

			_, err := store.WriteReaderFull(bytes.NewReader([]byte("blob 1\x00hello")))
			if err == nil {
				t.Fatalf("expected error after mismatch")
			}
		})

		t.Run("bytes malformed header", func(t *testing.T) {
			t.Parallel()

			store := memory.New(algo)

			_, err := store.WriteBytesFull([]byte("not-a-header"))
			if err == nil {
				t.Fatalf("expected error for malformed byte header")
			}
		})
	})
}

func TestBuildRawObjectMatchesObjectHeaderEncode(t *testing.T) {
	t.Parallel()

	content := []byte("body")
	raw := buildRawObject(t, objecttype.TypeBlob, content)

	header, ok := objectheader.Encode(objecttype.TypeBlob, int64(len(content)))
	if !ok {
		t.Fatalf("objectheader.Encode failed")
	}

	want := append(append([]byte(nil), header...), content...)
	if !bytes.Equal(raw, want) {
		t.Fatalf("buildRawObject mismatch")
	}
}

func buildRawObject(tb testing.TB, ty objecttype.Type, body []byte) []byte { //nolint:unparam
	tb.Helper()

	header, ok := objectheader.Encode(ty, int64(len(body)))
	if !ok {
		tb.Fatalf("objectheader.Encode(%v, %d) failed", ty, len(body))
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)

	return raw
}
