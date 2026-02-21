package object_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestTreeSerialize(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		entries := adversarialRootEntries(t, testRepo)
		tree := &object.Tree{}

		for i := len(entries) - 1; i >= 0; i-- {
			if err := tree.InsertEntry(entries[i]); err != nil {
				t.Fatalf("InsertEntry(%q): %v", entries[i].Name, err)
			}
		}
		if len(tree.Entries) < 32 {
			t.Fatalf("expected at least 32 entries, got %d", len(tree.Entries))
		}

		dup := tree.Entries[0]
		if err := tree.InsertEntry(dup); err == nil {
			t.Fatalf("duplicate InsertEntry should fail")
		}

		removed := tree.Entries[len(tree.Entries)/2]
		if err := tree.RemoveEntry(removed.Name); err != nil {
			t.Fatalf("RemoveEntry(%q): %v", removed.Name, err)
		}
		if tree.Entry(removed.Name) != nil {
			t.Fatalf("Entry(%q) should be nil after remove", removed.Name)
		}
		if err := tree.RemoveEntry([]byte("no-such-entry")); err == nil {
			t.Fatalf("RemoveEntry missing entry should fail")
		}
		if err := tree.InsertEntry(removed); err != nil {
			t.Fatalf("re-InsertEntry(%q): %v", removed.Name, err)
		}
		if tree.Entry(removed.Name) == nil {
			t.Fatalf("Entry(%q) should exist after reinsert", removed.Name)
		}

		wantTreeID := testRepo.Mktree(t, buildGitMktreeInput(tree.Entries))

		rawObj, err := tree.SerializeWithHeader()
		if err != nil {
			t.Fatalf("SerializeWithHeader: %v", err)
		}
		gotTreeID := algo.Sum(rawObj)
		if gotTreeID != wantTreeID {
			t.Fatalf("tree id mismatch: got %s want %s", gotTreeID, wantTreeID)
		}
	})
}
