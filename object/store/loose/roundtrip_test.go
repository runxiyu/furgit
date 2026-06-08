package loose_test

import (
	"bytes"
	"io"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		ty      typ.Type
		content []byte
	}{
		{name: "blob", ty: typ.TypeBlob, content: []byte("roundtrip blob\n")},
		{name: "empty blob", ty: typ.TypeBlob, content: []byte{}},
		{name: "tree", ty: typ.TypeTree, content: []byte("roundtrip tree bytes")},
		{name: "commit", ty: typ.TypeCommit, content: []byte("roundtrip commit bytes")},
		{name: "tag", ty: typ.TypeTag, content: []byte("roundtrip tag bytes")},
	}

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			looseStore := openLooseStore(t, repo)

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					wantRaw := header.Append(nil, tc.ty, uint64(len(tc.content)))
					wantRaw = append(wantRaw, tc.content...)

					objectID, err := looseStore.WriteBytesContent(tc.ty, tc.content)
					if err != nil {
						t.Fatalf("WriteBytesContent: %v", err)
					}

					gotRaw, err := looseStore.ReadBytesFull(objectID)
					if err != nil {
						t.Fatalf("ReadBytesFull: %v", err)
					}

					if !bytes.Equal(gotRaw, wantRaw) {
						t.Fatalf("ReadBytesFull mismatch")
					}

					gotType, gotBody, err := looseStore.ReadBytesContent(objectID)
					if err != nil {
						t.Fatalf("ReadBytesContent: %v", err)
					}

					if gotType != tc.ty {
						t.Fatalf("ReadBytesContent type = %v, want %v", gotType, tc.ty)
					}

					if !bytes.Equal(gotBody, tc.content) {
						t.Fatalf("ReadBytesContent body mismatch")
					}

					headType, headSize, err := looseStore.ReadHeader(objectID)
					if err != nil {
						t.Fatalf("ReadHeader: %v", err)
					}

					if headType != tc.ty {
						t.Fatalf("ReadHeader type = %v, want %v", headType, tc.ty)
					}

					if headSize != uint64(len(tc.content)) {
						t.Fatalf("ReadHeader size = %d, want %d", headSize, len(tc.content))
					}

					fullReader, err := looseStore.ReadReaderFull(objectID)
					if err != nil {
						t.Fatalf("ReadReaderFull: %v", err)
					}

					gotFull, err := io.ReadAll(fullReader)
					if err != nil {
						_ = fullReader.Close()

						t.Fatalf("ReadReaderFull ReadAll: %v", err)
					}

					err = fullReader.Close()
					if err != nil {
						t.Fatalf("ReadReaderFull Close: %v", err)
					}

					if !bytes.Equal(gotFull, wantRaw) {
						t.Fatalf("ReadReaderFull mismatch")
					}

					contentType, contentSize, contentReader, err := looseStore.ReadReaderContent(objectID)
					if err != nil {
						t.Fatalf("ReadReaderContent: %v", err)
					}

					gotContent, err := io.ReadAll(contentReader)
					if err != nil {
						_ = contentReader.Close()

						t.Fatalf("ReadReaderContent ReadAll: %v", err)
					}

					err = contentReader.Close()
					if err != nil {
						t.Fatalf("ReadReaderContent Close: %v", err)
					}

					if contentType != tc.ty {
						t.Fatalf("ReadReaderContent type = %v, want %v", contentType, tc.ty)
					}

					if contentSize != uint64(len(tc.content)) {
						t.Fatalf("ReadReaderContent size = %d, want %d", contentSize, len(tc.content))
					}

					if !bytes.Equal(gotContent, tc.content) {
						t.Fatalf("ReadReaderContent mismatch")
					}
				})
			}
		})
	}
}
