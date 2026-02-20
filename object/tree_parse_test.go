package object_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestTreeParseFromGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
		entries := adversarialRootEntries(t, testRepo)
		inserted := &object.Tree{}
		for _, entry := range entries {
			if err := inserted.InsertEntry(entry); err != nil {
				t.Fatalf("InsertEntry(%q): %v", entry.Name, err)
			}
		}

		treeID := testRepo.Mktree(t, buildGitMktreeInput(inserted.Entries))

		rawBody := testRepo.CatFile(t, "tree", treeID)
		tree, err := object.ParseTree(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseTree: %v", err)
		}
		if len(tree.Entries) != len(inserted.Entries) {
			t.Fatalf("entry count = %d, want %d", len(tree.Entries), len(inserted.Entries))
		}

		for i := range inserted.Entries {
			got := tree.Entries[i]
			want := inserted.Entries[i]
			if got.Mode != want.Mode || got.ID != want.ID || !bytes.Equal(got.Name, want.Name) {
				t.Fatalf("entry[%d] mismatch: got (%o,%q,%s) want (%o,%q,%s)",
					i, got.Mode, got.Name, got.ID, want.Mode, want.Name, want.ID)
			}
		}

		lsNames := gitLsTreeNames(testRepo.RunBytes(t, "ls-tree", "--name-only", "-z", treeID.String()))
		if len(lsNames) != len(tree.Entries) {
			t.Fatalf("ls-tree names = %d, want %d", len(lsNames), len(tree.Entries))
		}
		for i := range lsNames {
			if !bytes.Equal(lsNames[i], tree.Entries[i].Name) {
				t.Fatalf("ordering mismatch at %d: git=%q parsed=%q", i, lsNames[i], tree.Entries[i].Name)
			}
		}

		for _, want := range inserted.Entries {
			got := tree.Entry(want.Name)
			if got == nil {
				t.Fatalf("Entry(%q) returned nil", want.Name)
			}
			if got.Mode != want.Mode || got.ID != want.ID {
				t.Fatalf("Entry(%q) mismatch", want.Name)
			}
		}
		if tree.Entry([]byte("does-not-exist")) != nil {
			t.Fatalf("Entry on missing name should be nil")
		}
	})
}
