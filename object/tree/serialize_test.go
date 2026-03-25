package tree_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/tree"
)

func TestTreeSerialize(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		entries := adversarialRootEntries(t, testRepo)
		obj := &tree.Tree{}

		for i := len(entries) - 1; i >= 0; i-- {
			err := obj.InsertEntry(entries[i])
			if err != nil {
				t.Fatalf("InsertEntry(%q): %v", entries[i].Name, err)
			}
		}

		if len(obj.Entries) < 32 {
			t.Fatalf("expected at least 32 entries, got %d", len(obj.Entries))
		}

		dup := obj.Entries[0]

		err := obj.InsertEntry(dup)
		if err == nil {
			t.Fatalf("duplicate InsertEntry should fail")
		}

		removed := obj.Entries[len(obj.Entries)/2]

		err = obj.RemoveEntry(removed.Name)
		if err != nil {
			t.Fatalf("RemoveEntry(%q): %v", removed.Name, err)
		}

		if obj.Entry(removed.Name) != nil {
			t.Fatalf("Entry(%q) should be nil after remove", removed.Name)
		}

		err = obj.RemoveEntry([]byte("no-such-entry"))
		if err == nil {
			t.Fatalf("RemoveEntry missing entry should fail")
		}

		err = obj.InsertEntry(removed)
		if err != nil {
			t.Fatalf("re-InsertEntry(%q): %v", removed.Name, err)
		}

		if obj.Entry(removed.Name) == nil {
			t.Fatalf("Entry(%q) should exist after reinsert", removed.Name)
		}

		wantTreeID := testRepo.Mktree(t, buildGitMktreeInput(obj.Entries))

		rawObj, err := obj.SerializeWithHeader()
		if err != nil {
			t.Fatalf("SerializeWithHeader: %v", err)
		}

		gotTreeID := algo.Sum(rawObj)
		if gotTreeID != wantTreeID {
			t.Fatalf("tree id mismatch: got %s want %s", gotTreeID, wantTreeID)
		}
	})
}
