package packed_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store/packed"
)

// makeGitPack seeds a repository,
// packs the seeded objects with git pack-objects,
// and returns the repository, the artifact path prefix,
// and the seeded objects.
func makeGitPack(t *testing.T, objectFormat id.ObjectFormat) (*testgit.Repo, string, testgit.Seeded) {
	t.Helper()

	repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	seeded, err := repo.SeedHistory(t)
	if err != nil {
		t.Fatalf("SeedHistory: %v", err)
	}

	prefix, err := repo.PackObjects(t, slices.Values(seeded.All()), testgit.PackObjectsOptions{})
	if err != nil {
		t.Fatalf("PackObjects: %v", err)
	}

	return repo, prefix, seeded
}

// requireDeltas asserts that the pack at prefix
// contains at least one deltified entry,
// so that tests really do exercise delta resolution.
func requireDeltas(t *testing.T, repo *testgit.Repo, prefix string, objectFormat id.ObjectFormat) {
	t.Helper()

	out, err := repo.VerifyPack(t, prefix+".idx")
	if err != nil {
		t.Fatalf("VerifyPack: %v", err)
	}

	hexLen := objectFormat.HexLen()

	for line := range strings.Lines(string(out)) {
		fields := strings.Fields(line)
		if len(fields) >= 7 && len(fields[0]) == hexLen {
			return
		}
	}

	t.Fatalf("fixture pack contains no deltified entries")
}

// openPackedStore opens a packed store
// over the directory containing prefix's pack artifacts.
func openPackedStore(t *testing.T, prefix string, objectFormat id.ObjectFormat) *packed.Packed {
	t.Helper()

	root, err := os.OpenRoot(filepath.Dir(prefix))
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
