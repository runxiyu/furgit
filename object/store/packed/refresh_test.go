package packed_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/packed"
)

func TestRefreshIsExplicit(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			_, prefix, seeded := makeGitPack(t, objectFormat)
			oids := seeded.All()

			dir := t.TempDir()

			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatalf("OpenRoot: %v", err)
			}

			t.Cleanup(func() { _ = root.Close() })

			packedStore, err := packed.New(root, objectFormat)
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			t.Cleanup(func() { _ = packedStore.Close() })

			_, _, err = packedStore.ReadBytesContent(oids[0])
			if !errors.Is(err, store.ErrObjectNotFound) {
				t.Fatalf("empty store read error = %v, want ErrObjectNotFound", err)
			}

			t.Helper()
			base := filepath.Base(prefix)
			cp(t, prefix+".pack", filepath.Join(dir, base+".pack"))
			cp(t, prefix+".idx", filepath.Join(dir, base+".idx"))

			// New packs must stay invisible until an explicit Refresh.
			_, _, err = packedStore.ReadBytesContent(oids[0])
			if !errors.Is(err, store.ErrObjectNotFound) {
				t.Fatalf("pre-Refresh read error = %v, want ErrObjectNotFound", err)
			}

			err = packedStore.Refresh()
			if err != nil {
				t.Fatalf("Refresh: %v", err)
			}

			for _, oid := range oids {
				_, _, err := packedStore.ReadBytesContent(oid)
				if err != nil {
					t.Fatalf("post-Refresh ReadBytesContent(%s): %v", oid, err)
				}
			}
		})
	}
}

func TestRefreshRejectsIndexWithoutPack(t *testing.T) {
	t.Parallel()

	objectFormat := id.ObjectFormatSHA256

	_, prefix, _ := makeGitPack(t, objectFormat)

	dir := t.TempDir()
	cp(t, prefix+".idx", filepath.Join(dir, filepath.Base(prefix)+".idx"))

	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}

	t.Cleanup(func() { _ = root.Close() })

	_, err = packed.New(root, objectFormat)
	if err == nil {
		t.Fatalf("New with orphan index: expected error")
	}
}

// cp copies one file from src to dst.
func cp(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src) //nolint:gosec
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	err = os.WriteFile(dst, data, 0o600) //nolint:gosec
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
