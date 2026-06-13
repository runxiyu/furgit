package packed_test

import (
	"bytes"
	"io"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store/packed"
	"lindenii.org/go/furgit/object/typ"
)

func TestReadGitPack(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, prefix, seeded := makeGitPack(t, objectFormat)
			requireDeltas(t, repo, prefix, objectFormat)

			packedStore := openPackedStore(t, prefix, objectFormat)

			groups := []struct {
				ty   typ.Type
				oids []id.ObjectID
			}{
				{ty: typ.Blob, oids: seeded.Blobs},
				{ty: typ.Tree, oids: seeded.Trees},
				{ty: typ.Commit, oids: seeded.Commits},
				{ty: typ.Tag, oids: seeded.Tags},
			}

			for _, group := range groups {
				for _, oid := range group.oids {
					wantContent, err := repo.CatFile(t, group.ty, oid)
					if err != nil {
						t.Fatalf("CatFile(%s): %v", oid, err)
					}

					ty, content, err := packedStore.ReadBytesContent(oid)
					if err != nil {
						t.Fatalf("ReadBytesContent(%s): %v", oid, err)
					}

					if ty != group.ty {
						t.Fatalf("ReadBytesContent(%s) type = %v, want %v", oid, ty, group.ty)
					}

					if !bytes.Equal(content, wantContent) {
						t.Fatalf("ReadBytesContent(%s) content mismatch", oid)
					}

					raw, err := packedStore.ReadBytesFull(oid)
					if err != nil {
						t.Fatalf("ReadBytesFull(%s): %v", oid, err)
					}

					if got := objectFormat.Sum(raw); got != oid {
						t.Fatalf("ReadBytesFull(%s) hashes to %s", oid, got)
					}

					ty, size, err := packedStore.ReadHeader(oid)
					if err != nil {
						t.Fatalf("ReadHeader(%s): %v", oid, err)
					}

					if ty != group.ty {
						t.Fatalf("ReadHeader(%s) type = %v, want %v", oid, ty, group.ty)
					}

					if size != len(wantContent) {
						t.Fatalf("ReadHeader(%s) size = %d, want %d", oid, size, len(wantContent))
					}

					size, err = packedStore.ReadSize(oid)
					if err != nil {
						t.Fatalf("ReadSize(%s): %v", oid, err)
					}

					if size != len(wantContent) {
						t.Fatalf("ReadSize(%s) = %d, want %d", oid, size, len(wantContent))
					}

					checkReaderContent(t, packedStore, oid, group.ty, wantContent)
					checkReaderFull(t, packedStore, oid, objectFormat)
				}
			}
		})
	}
}

func checkReaderContent(t *testing.T, packedStore *packed.Packed, oid id.ObjectID, wantType typ.Type, wantContent []byte) {
	t.Helper()

	ty, size, reader, err := packedStore.ReadReaderContent(oid)
	if err != nil {
		t.Fatalf("ReadReaderContent(%s): %v", oid, err)
	}

	defer func() { _ = reader.Close() }()

	if ty != wantType {
		t.Fatalf("ReadReaderContent(%s) type = %v, want %v", oid, ty, wantType)
	}

	if size != len(wantContent) {
		t.Fatalf("ReadReaderContent(%s) size = %d, want %d", oid, size, len(wantContent))
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadReaderContent(%s) read: %v", oid, err)
	}

	if !bytes.Equal(content, wantContent) {
		t.Fatalf("ReadReaderContent(%s) content mismatch", oid)
	}
}

func checkReaderFull(t *testing.T, packedStore *packed.Packed, oid id.ObjectID, objectFormat id.ObjectFormat) {
	t.Helper()

	reader, err := packedStore.ReadReaderFull(oid)
	if err != nil {
		t.Fatalf("ReadReaderFull(%s): %v", oid, err)
	}

	defer func() { _ = reader.Close() }()

	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadReaderFull(%s) read: %v", oid, err)
	}

	if got := objectFormat.Sum(raw); got != oid {
		t.Fatalf("ReadReaderFull(%s) hashes to %s", oid, got)
	}
}
