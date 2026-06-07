package tree_test

import (
	"bytes"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

func TestAppend(t *testing.T) {
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

			assertGitDecode(t, repo, treeID, tr.Entries())
		})
	}
}
