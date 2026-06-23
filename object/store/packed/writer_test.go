package packed_test

import (
	"bytes"
	"os"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/packed"
	"lindenii.org/go/furgit/object/typ"
)

// TestWritePack verifies that writing a pack through the store
// makes its objects readable without a manual refresh.
func TestWritePack(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			seeded, err := repo.SeedHistory(t)
			if err != nil {
				t.Fatalf("SeedHistory: %v", err)
			}

			stream, err := repo.PackObjectsStdout(t, seeded.All(), testgit.PackObjectsStdoutOptions{
				Revs:    false,
				Thin:    false,
				Exclude: nil,
			})
			if err != nil {
				t.Fatalf("PackObjectsStdout: %v", err)
			}

			packedStore := openEmptyStore(t, objectFormat)

			err = packedStore.WritePack(t.Context(), bytes.NewReader(stream), store.PackWriteOptions{
				ThinBase: nil,
				Progress: nil,
			})
			if err != nil {
				t.Fatalf("WritePack: %v", err)
			}

			probes := []struct {
				ty  typ.Type
				oid id.ObjectID
			}{
				{typ.Blob, seeded.Blobs[0]},
				{typ.Tree, seeded.Trees[0]},
				{typ.Commit, seeded.Commits[len(seeded.Commits)-1]},
				{typ.Tag, seeded.Tags[0]},
			}

			for _, probe := range probes {
				want, err := repo.CatFile(t, probe.ty, probe.oid)
				if err != nil {
					t.Fatalf("CatFile(%s): %v", probe.oid, err)
				}

				ty, content, err := packedStore.ReadBytesContent(probe.oid)
				if err != nil {
					t.Fatalf("ReadBytesContent(%s): %v", probe.oid, err)
				}

				if ty != probe.ty {
					t.Fatalf("ReadBytesContent(%s) type = %v, want %v", probe.oid, ty, probe.ty)
				}

				if !bytes.Equal(content, want) {
					t.Fatalf("ReadBytesContent(%s) content mismatch", probe.oid)
				}
			}
		})
	}
}

// openEmptyStore opens a packed store over a fresh empty directory.
func openEmptyStore(t *testing.T, objectFormat id.ObjectFormat) *packed.Packed {
	t.Helper()

	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}

	t.Cleanup(func() { _ = root.Close() })

	packedStore, err := packed.New(root, objectFormat)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	t.Cleanup(func() { _ = packedStore.Close() })

	return packedStore
}
