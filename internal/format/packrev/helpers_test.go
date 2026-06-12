package packrev_test

import (
	"cmp"
	"slices"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/lgo/intconv"
)

// makeGitPack seeds a repository,
// packs the seeded objects with git pack-objects
// including a reverse index,
// and returns the artifact path prefix.
func makeGitPack(t *testing.T, objectFormat id.ObjectFormat) string {
	t.Helper()

	repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	seeded, err := repo.SeedHistory(t)
	if err != nil {
		t.Fatalf("SeedHistory: %v", err)
	}

	prefix, err := repo.PackObjects(t, slices.Values(seeded.All()), testgit.PackObjectsOptions{RevIndex: true})
	if err != nil {
		t.Fatalf("PackObjects: %v", err)
	}

	return prefix
}

// packOrderPositions derives the pack-offset-order index positions
// from one parsed pack index.
func packOrderPositions(t *testing.T, idx *packidx.Packidx) []uint32 {
	t.Helper()

	type pair struct {
		offset   uint64
		position uint32
	}

	pairs := make([]pair, 0, idx.NumObjects())

	for pos := range idx.NumObjects() {
		offset, err := idx.OffsetAt(pos)
		if err != nil {
			t.Fatalf("OffsetAt(%d): %v", pos, err)
		}

		position, err := intconv.IntToUint32(pos)
		if err != nil {
			t.Fatalf("IntToUint32(%d): %v", pos, err)
		}

		pairs = append(pairs, pair{offset: offset, position: position})
	}

	slices.SortFunc(pairs, func(a, b pair) int {
		return cmp.Compare(a.offset, b.offset)
	})

	positions := make([]uint32, 0, len(pairs))
	for _, p := range pairs {
		positions = append(positions, p.position)
	}

	return positions
}
