package ancestor_test

import (
	"errors"
	"testing"

	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"

	"codeberg.org/lindenii/furgit/ancestor"
)

func TestIsMatchesGitMergeBase(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, tree1 := testRepo.MakeSingleFileTree(t, "one.txt", []byte("one\n"))
		c1 := testRepo.CommitTree(t, tree1, "c1")

		_, tree2 := testRepo.MakeSingleFileTree(t, "two.txt", []byte("two\n"))
		c2 := testRepo.CommitTree(t, tree2, "c2", c1)

		_, tree3 := testRepo.MakeSingleFileTree(t, "three.txt", []byte("three\n"))
		c3 := testRepo.CommitTree(t, tree3, "c3", c2)

		tag := testRepo.TagAnnotated(t, "tip", c2, "tip")

		store := testRepo.OpenObjectStore(t)

		got, err := ancestor.Is(store, nil, c1, tag)
		if err != nil {
			t.Fatalf("Is(c1, tag): %v", err)
		}

		want := gitMergeBaseIsAncestor(t, testRepo, c1, c2)
		if got != want {
			t.Fatalf("Is(c1, tag)=%v, want %v", got, want)
		}

		got, err = ancestor.Is(store, nil, c3, c2)
		if err != nil {
			t.Fatalf("Is(c3, c2): %v", err)
		}

		want = gitMergeBaseIsAncestor(t, testRepo, c3, c2)
		if got != want {
			t.Fatalf("Is(c3, c2)=%v, want %v", got, want)
		}
	})
}

func TestIsMatchesGitMergeBaseWithCommitGraph(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, tree1 := testRepo.MakeSingleFileTree(t, "one.txt", []byte("one\n"))
		c1 := testRepo.CommitTree(t, tree1, "c1")

		_, tree2 := testRepo.MakeSingleFileTree(t, "two.txt", []byte("two\n"))
		c2 := testRepo.CommitTree(t, tree2, "c2", c1)

		testRepo.UpdateRef(t, "refs/heads/main", c2)
		testRepo.SymbolicRef(t, "HEAD", "refs/heads/main")
		testRepo.CommitGraphWrite(t, "--reachable")

		store := testRepo.OpenObjectStore(t)
		graph := testRepo.OpenCommitGraph(t)

		got, err := ancestor.Is(store, graph, c1, c2)
		if err != nil {
			t.Fatalf("Is(c1, c2): %v", err)
		}

		want := gitMergeBaseIsAncestor(t, testRepo, c1, c2)
		if got != want {
			t.Fatalf("Is(c1, c2)=%v, want %v", got, want)
		}
	})
}

func TestIsMissingObject(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, treeID, commitID := testRepo.MakeCommit(t, "missing")

		testRepo.RemoveLooseObject(t, treeID)

		store := testRepo.OpenObjectStore(t)

		_, err := ancestor.Is(store, nil, treeID, commitID)
		if err == nil {
			t.Fatal("expected error")
		}

		missing, ok := errors.AsType[*giterrors.ObjectMissingError](err)
		if !ok {
			t.Fatalf("expected ObjectMissingError, got %T (%v)", err, err)
		}

		if missing.OID != treeID {
			t.Fatalf("missing oid = %s, want %s", missing.OID, treeID)
		}
	})
}

// gitMergeBaseIsAncestor reports Git's merge-base ancestry answer.
func gitMergeBaseIsAncestor(t *testing.T, testRepo *testgit.TestRepo, left, right objectid.ObjectID) bool {
	t.Helper()

	out := testRepo.Run(t, "merge-base", left.String(), right.String())

	return out == left.String()
}
