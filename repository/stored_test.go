package repository_test

import (
	"fmt"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/repository"
)

func TestReadStoredTyped(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		blobID, treeID, commitID := repoHarness.MakeCommit(t, "stored types")

		repo, err := repository.Open(repoHarness.Dir())
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}
		defer func() { _ = repo.Close() }()

		blob, err := repo.ReadStoredBlob(blobID)
		if err != nil {
			t.Fatalf("ReadStoredBlob: %v", err)
		}
		if blob.ID() != blobID {
			t.Fatalf("blob ID = %s, want %s", blob.ID(), blobID)
		}
		if string(blob.Blob().Data) != "commit-body\n" {
			t.Fatalf("blob body = %q, want %q", blob.Blob().Data, "commit-body\n")
		}

		tree, err := repo.ReadStoredTree(treeID)
		if err != nil {
			t.Fatalf("ReadStoredTree: %v", err)
		}
		if tree.ID() != treeID {
			t.Fatalf("tree ID = %s, want %s", tree.ID(), treeID)
		}
		if len(tree.Tree().Entries) != 1 {
			t.Fatalf("tree entries = %d, want 1", len(tree.Tree().Entries))
		}

		commit, err := repo.ReadStoredCommit(commitID)
		if err != nil {
			t.Fatalf("ReadStoredCommit: %v", err)
		}
		if commit.ID() != commitID {
			t.Fatalf("commit ID = %s, want %s", commit.ID(), commitID)
		}
		if commit.Commit().Tree != treeID {
			t.Fatalf("commit tree = %s, want %s", commit.Commit().Tree, treeID)
		}
	})
}

func TestResolveTreeEntry(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		blobID := repoHarness.HashObject(t, "blob", []byte("nested-file\n"))
		childTreeID := repoHarness.Mktree(t, fmt.Sprintf("100644 blob %s\tleaf.txt\n", blobID))
		rootTreeID := repoHarness.Mktree(t, fmt.Sprintf("040000 tree %s\tdir\n", childTreeID))

		repo, err := repository.Open(repoHarness.Dir())
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}
		defer func() { _ = repo.Close() }()

		rootTree, err := repo.ReadStoredTree(rootTreeID)
		if err != nil {
			t.Fatalf("ReadStoredTree(root): %v", err)
		}

		entry, err := repo.ResolveTreeEntry(rootTree, [][]byte{[]byte("dir"), []byte("leaf.txt")})
		if err != nil {
			t.Fatalf("ResolveTreeEntry: %v", err)
		}
		if entry.Mode != object.FileModeRegular {
			t.Fatalf("ResolveTreeEntry mode = %o, want %o", entry.Mode, object.FileModeRegular)
		}
		if entry.ID != blobID {
			t.Fatalf("ResolveTreeEntry id = %s, want %s", entry.ID, blobID)
		}
	})
}

func TestResolveTreeEntryErrors(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		t.Run("missing path component", func(t *testing.T) {
			t.Parallel()
			repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
				ObjectFormat: algo,
				Bare:         true,
				RefFormat:    "files",
			})
			blobID := repoHarness.HashObject(t, "blob", []byte("body\n"))
			rootTreeID := repoHarness.Mktree(t, fmt.Sprintf("100644 blob %s\tfile.txt\n", blobID))

			repo, err := repository.Open(repoHarness.Dir())
			if err != nil {
				t.Fatalf("repository.Open: %v", err)
			}
			defer func() { _ = repo.Close() }()

			rootTree, err := repo.ReadStoredTree(rootTreeID)
			if err != nil {
				t.Fatalf("ReadStoredTree(root): %v", err)
			}

			_, err = repo.ResolveTreeEntry(rootTree, [][]byte{[]byte("missing")})
			if err == nil || !strings.Contains(err.Error(), "not found") {
				t.Fatalf("ResolveTreeEntry missing: err = %v, want not found error", err)
			}
		})

		t.Run("non-tree intermediate", func(t *testing.T) {
			t.Parallel()
			repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
				ObjectFormat: algo,
				Bare:         true,
				RefFormat:    "files",
			})
			blobID := repoHarness.HashObject(t, "blob", []byte("body\n"))
			rootTreeID := repoHarness.Mktree(t, fmt.Sprintf("100644 blob %s\tdir\n", blobID))

			repo, err := repository.Open(repoHarness.Dir())
			if err != nil {
				t.Fatalf("repository.Open: %v", err)
			}
			defer func() { _ = repo.Close() }()

			rootTree, err := repo.ReadStoredTree(rootTreeID)
			if err != nil {
				t.Fatalf("ReadStoredTree(root): %v", err)
			}

			_, err = repo.ResolveTreeEntry(rootTree, [][]byte{[]byte("dir"), []byte("leaf")})
			if err == nil || !strings.Contains(err.Error(), "is not a tree") {
				t.Fatalf("ResolveTreeEntry non-tree: err = %v, want non-tree error", err)
			}
		})
	})
}
