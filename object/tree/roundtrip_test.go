package tree_test

import (
	"bytes"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree"
	"lindenii.org/go/furgit/object/typ"
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			entries := mixedEntries(t, repo)
			tr := buildTree(t, entries)

			rawBody, err := tr.AppendWithoutHeader(nil)
			if err != nil {
				t.Fatalf("AppendWithoutHeader: %v", err)
			}

			treeID, err := repo.HashObject(t, typ.TypeTree, bytes.NewReader(rawBody))
			if err != nil {
				t.Fatalf("HashObject(tree): %v", err)
			}

			err = repo.Fsck(t, testgit.FsckOptions{
				Strict:     true,
				NoDangling: true,
			}, treeID)
			if err != nil {
				t.Fatalf("Fsck: %v", err)
			}

			gitBody, err := repo.CatFile(t, typ.TypeTree, treeID)
			if err != nil {
				t.Fatalf("CatFile: %v", err)
			}

			parsed, err := tree.Parse(gitBody, objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			assertEntriesEqual(t, parsed.Entries(), tr.Entries())
		})
	}
}

func assertEntriesEqual(t *testing.T, got []tree.Entry, want []tree.Entry) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("entry count = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i].Mode != want[i].Mode {
			t.Fatalf("entry[%d] mode = %o, want %o", i, got[i].Mode, want[i].Mode)
		}

		if got[i].Name != want[i].Name {
			t.Fatalf("entry[%d] name = %q, want %q", i, got[i].Name, want[i].Name)
		}

		if got[i].ID != want[i].ID {
			t.Fatalf("entry[%d] id = %s, want %s", i, got[i].ID, want[i].ID)
		}
	}
}
