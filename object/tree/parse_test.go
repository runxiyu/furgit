package tree_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/tree"
)

func TestTreeParseFromGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		entries := adversarialRootEntries(t, testRepo)

		inserted := &tree.Tree{}
		for _, entry := range entries {
			err := inserted.InsertEntry(entry)
			if err != nil {
				t.Fatalf("InsertEntry(%q): %v", entry.Name, err)
			}
		}

		treeID := testRepo.Mktree(t, buildGitMktreeInput(inserted.Entries))

		rawBody := testRepo.CatFile(t, "tree", treeID)

		parsed, err := tree.Parse(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseTree: %v", err)
		}

		if len(parsed.Entries) != len(inserted.Entries) {
			t.Fatalf("entry count = %d, want %d", len(parsed.Entries), len(inserted.Entries))
		}

		for i := range inserted.Entries {
			got := parsed.Entries[i]

			want := inserted.Entries[i]
			if got.Mode != want.Mode || got.ID != want.ID || !bytes.Equal(got.Name, want.Name) {
				t.Fatalf("entry[%d] mismatch: got (%o,%q,%s) want (%o,%q,%s)",
					i, got.Mode, got.Name, got.ID, want.Mode, want.Name, want.ID)
			}
		}

		lsNames := gitLsTreeNames(testRepo.RunBytes(t, "ls-tree", "--name-only", "-z", treeID.String()))
		if len(lsNames) != len(parsed.Entries) {
			t.Fatalf("ls-tree names = %d, want %d", len(lsNames), len(parsed.Entries))
		}

		for i := range lsNames {
			if !bytes.Equal(lsNames[i], parsed.Entries[i].Name) {
				t.Fatalf("ordering mismatch at %d: git=%q parsed=%q", i, lsNames[i], parsed.Entries[i].Name)
			}
		}

		for _, want := range inserted.Entries {
			got := parsed.Entry(want.Name)

			if got == nil {
				t.Fatalf("Entry(%q) returned nil", want.Name)
			}

			if got.Mode != want.Mode || got.ID != want.ID {
				t.Fatalf("Entry(%q) mismatch", want.Name)
			}
		}

		if parsed.Entry([]byte("does-not-exist")) != nil {
			t.Fatalf("Entry on missing name should be nil")
		}
	})
}

func TestTreeInsertEntryCopiesName(t *testing.T) {
	t.Parallel()

	var tr tree.Tree
	name := []byte("alpha")
	entry := tree.TreeEntry{
		Mode: tree.FileModeRegular,
		Name: name,
		ID:   objectid.ObjectID{},
	}

	if err := tr.InsertEntry(entry); err != nil {
		t.Fatalf("InsertEntry: %v", err)
	}

	name[0] = 'b'

	got := tr.Entry([]byte("alpha"))
	if got == nil {
		t.Fatalf("Entry(alpha) returned nil")
	}

	if !bytes.Equal(got.Name, []byte("alpha")) {
		t.Fatalf("stored name = %q, want %q", got.Name, []byte("alpha"))
	}

	if tr.Entry([]byte("blpha")) != nil {
		t.Fatalf("mutating caller name should not affect stored entry")
	}
}
