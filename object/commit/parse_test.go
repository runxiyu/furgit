package commit_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/commit"
	"lindenii.org/go/furgit/object/id"
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

			blobID, err := repo.HashObject(t, typ.TypeBlob, strings.NewReader("content\n"))
			if err != nil {
				t.Fatalf("HashObject: %v", err)
			}

			treeID, err := repo.MkTree(t, []testgit.MkTreeEntry{
				{Mode: "100644", Type: typ.TypeBlob, OID: blobID, Name: "file.txt"},
			})
			if err != nil {
				t.Fatalf("MkTree: %v", err)
			}

			commitID, err := repo.CommitTree(t, treeID, "subject\n\nbody")
			if err != nil {
				t.Fatalf("CommitTree: %v", err)
			}

			rawBody, err := repo.CatFile(t, typ.TypeCommit, commitID)
			if err != nil {
				t.Fatalf("CatFile: %v", err)
			}

			parsed, err := commit.Parse(rawBody, objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			if parsed.Tree != treeID {
				t.Fatalf("tree id mismatch: got %s want %s", parsed.Tree, treeID)
			}

			if len(parsed.Parents) != 0 {
				t.Fatalf("parent count = %d, want 0", len(parsed.Parents))
			}

			if !bytes.Equal(parsed.Author.Name, []byte("Test Author")) {
				t.Fatalf("author name = %q, want %q", parsed.Author.Name, "Test Author")
			}

			if !bytes.Equal(parsed.Committer.Name, []byte("Test Committer")) {
				t.Fatalf("committer name = %q, want %q", parsed.Committer.Name, "Test Committer")
			}

			if !bytes.Contains(parsed.Message, []byte("subject")) {
				t.Fatalf("commit message missing subject: %q", parsed.Message)
			}
		})
	}
}

func TestParseMultipleParents(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			blobID, err := repo.HashObject(t, typ.TypeBlob, strings.NewReader("merge-content\n"))
			if err != nil {
				t.Fatalf("HashObject: %v", err)
			}

			treeID, err := repo.MkTree(t, []testgit.MkTreeEntry{
				{Mode: "100644", Type: typ.TypeBlob, OID: blobID, Name: "file.txt"},
			})
			if err != nil {
				t.Fatalf("MkTree: %v", err)
			}

			parent1, err := repo.CommitTree(t, treeID, "parent-one")
			if err != nil {
				t.Fatalf("CommitTree(parent1): %v", err)
			}

			parent2, err := repo.CommitTree(t, treeID, "parent-two", parent1)
			if err != nil {
				t.Fatalf("CommitTree(parent2): %v", err)
			}

			rawCommit := fmt.Sprintf(
				"tree %s\nparent %s\nparent %s\nauthor Test Author <test@example.org> 1234567890 +0000\ncommitter Test Committer <committer@example.org> 1234567890 +0000\n\nMerge commit\n",
				treeID,
				parent1,
				parent2,
			)

			mergeID, err := repo.HashObject(t, typ.TypeCommit, strings.NewReader(rawCommit))
			if err != nil {
				t.Fatalf("HashObject(merge): %v", err)
			}

			rawBody, err := repo.CatFile(t, typ.TypeCommit, mergeID)
			if err != nil {
				t.Fatalf("CatFile: %v", err)
			}

			parsed, err := commit.Parse(rawBody, objectFormat)
			if err != nil {
				t.Fatalf("Parse(merge): %v", err)
			}

			if parsed.Tree != treeID {
				t.Fatalf("merge tree = %s, want %s", parsed.Tree, treeID)
			}

			if len(parsed.Parents) != 2 {
				t.Fatalf("merge parent count = %d, want 2", len(parsed.Parents))
			}

			if parsed.Parents[0] != parent1 {
				t.Fatalf("merge parent[0] = %s, want %s", parsed.Parents[0], parent1)
			}

			if parsed.Parents[1] != parent2 {
				t.Fatalf("merge parent[1] = %s, want %s", parsed.Parents[1], parent2)
			}

			if !bytes.Equal(parsed.Message, []byte("Merge commit\n")) {
				t.Fatalf("merge message = %q, want %q", parsed.Message, "Merge commit\n")
			}
		})
	}
}
