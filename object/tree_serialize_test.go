package object_test

import (
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/oid"
)

func TestTreeSerialize(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo oid.Algorithm) {
		repo := testgit.NewBareRepo(t, algo)
		entries := adversarialRootEntries(t, repo)
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
		if err := tree.RemoveEntry([]byte("no-such-entry")); !errors.Is(err, object.ErrNotFound) {
			t.Fatalf("RemoveEntry missing err = %v, want ErrNotFound", err)
		}
		if err := tree.InsertEntry(removed); err != nil {
			t.Fatalf("re-InsertEntry(%q): %v", removed.Name, err)
		}
		if tree.Entry(removed.Name) == nil {
			t.Fatalf("Entry(%q) should exist after reinsert", removed.Name)
		}

		wantTreeID := repo.Mktree(t, buildGitMktreeInput(tree.Entries))

		rawObj, err := tree.Serialize()
		if err != nil {
			t.Fatalf("Serialize: %v", err)
		}
		gotTreeID := algo.Sum(rawObj)
		if gotTreeID != wantTreeID {
			t.Fatalf("tree id mismatch: got %s want %s", gotTreeID, wantTreeID)
		}
	})
}
