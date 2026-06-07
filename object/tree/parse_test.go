package tree_test

import (
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree"
	"lindenii.org/go/furgit/object/typ"
)

func TestParse(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			entries := mixedEntries(t, repo)

			treeID, err := repo.MkTree(t, mkTreeEntries(entries))
			if err != nil {
				t.Fatalf("MkTree: %v", err)
			}

			rawBody, err := repo.CatFile(t, typ.TypeTree, treeID)
			if err != nil {
				t.Fatalf("CatFile: %v", err)
			}

			parsed, err := tree.Parse(rawBody, objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			assertGitDecode(t, repo, treeID, parsed.Entries())
		})
	}
}
