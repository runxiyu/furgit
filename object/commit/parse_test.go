package commit_test

import (
	"bytes"
	"fmt"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/commit"
	objectid "lindenii.org/go/furgit/object/id"
)

func TestCommitParseFromGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, treeID, commitID := testRepo.MakeCommit(t, "subject\n\nbody")

		rawBody := testRepo.CatFile(t, "commit", commitID)

		parsed, err := commit.Parse(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseCommit: %v", err)
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

func TestCommitParseMultipleParents(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})

		_, treeID := testRepo.MakeSingleFileTree(t, "file.txt", []byte("merge-content\n"))
		parent1 := testRepo.CommitTree(t, treeID, "parent-one")
		parent2 := testRepo.CommitTree(t, treeID, "parent-two", parent1)

		rawCommit := fmt.Sprintf(
			"tree %s\nparent %s\nparent %s\nauthor Test Author <test@example.org> 1234567890 +0000\ncommitter Test Committer <committer@example.org> 1234567890 +0000\n\nMerge commit\n",
			treeID,
			parent1,
			parent2,
		)
		mergeID := testRepo.HashObject(t, "commit", []byte(rawCommit))
		rawBody := testRepo.CatFile(t, "commit", mergeID)

		parsed, err := commit.Parse(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseCommit(merge): %v", err)
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
