package object_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestCommitParseFromGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		repo := testgit.NewBareRepo(t, algo)
		_, treeID, commitID := repo.MakeCommit(t, "subject\n\nbody")

		rawBody := repo.CatFile(t, "commit", commitID)
		commit, err := object.ParseCommit(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseCommit: %v", err)
		}
		if commit.Tree != treeID {
			t.Fatalf("tree id mismatch: got %s want %s", commit.Tree, treeID)
		}
		if len(commit.Parents) != 0 {
			t.Fatalf("parent count = %d, want 0", len(commit.Parents))
		}
		if !bytes.Equal(commit.Author.Name, []byte("Test Author")) {
			t.Fatalf("author name = %q, want %q", commit.Author.Name, "Test Author")
		}
		if !bytes.Equal(commit.Committer.Name, []byte("Test Committer")) {
			t.Fatalf("committer name = %q, want %q", commit.Committer.Name, "Test Committer")
		}
		if !bytes.Contains(commit.Message, []byte("subject")) {
			t.Fatalf("commit message missing subject: %q", commit.Message)
		}
	})
}
