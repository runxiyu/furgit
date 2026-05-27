package repository_test

import (
	"fmt"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree"
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

		repo := repoHarness.OpenRepository(t)

		blob, err := repo.Fetcher().ExactBlob(blobID)
		if err != nil {
			t.Fatalf("ExactBlob: %v", err)
		}

		if blob.ID() != blobID {
			t.Fatalf("blob ID = %s, want %s", blob.ID(), blobID)
		}

		if string(blob.Object().Data) != "commit-body\n" {
			t.Fatalf("blob body = %q, want %q", blob.Object().Data, "commit-body\n")
		}

		tree, err := repo.Fetcher().ExactTree(treeID)
		if err != nil {
			t.Fatalf("ExactTree: %v", err)
		}

		if tree.ID() != treeID {
			t.Fatalf("tree ID = %s, want %s", tree.ID(), treeID)
		}

		if len(tree.Object().Entries) != 1 {
			t.Fatalf("tree entries = %d, want 1", len(tree.Object().Entries))
		}

		commit, err := repo.Fetcher().ExactCommit(commitID)
		if err != nil {
			t.Fatalf("ExactCommit: %v", err)
		}

		if commit.ID() != commitID {
			t.Fatalf("commit ID = %s, want %s", commit.ID(), commitID)
		}

		if commit.Object().Tree != treeID {
			t.Fatalf("commit tree = %s, want %s", commit.Object().Tree, treeID)
		}
	})
}

func TestResolverPath(t *testing.T) {
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

		repo := repoHarness.OpenRepository(t)

		entry, err := repo.Fetcher().Path(rootTreeID, [][]byte{[]byte("dir"), []byte("leaf.txt")})
		if err != nil {
			t.Fatalf("Path: %v", err)
		}

		if entry.Mode != tree.FileModeRegular {
			t.Fatalf("Path mode = %o, want %o", entry.Mode, tree.FileModeRegular)
		}

		if entry.ID != blobID {
			t.Fatalf("Path id = %s, want %s", entry.ID, blobID)
		}
	})
}

func TestResolverPathErrors(t *testing.T) {
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

			repo := repoHarness.OpenRepository(t)

			_, err := repo.Fetcher().Path(rootTreeID, [][]byte{[]byte("missing")})
			if err == nil || !strings.Contains(err.Error(), "not found") {
				t.Fatalf("Path missing: err = %v, want not found error", err)
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

			repo := repoHarness.OpenRepository(t)

			_, err := repo.Fetcher().Path(rootTreeID, [][]byte{[]byte("dir"), []byte("leaf")})
			if err == nil || !strings.Contains(err.Error(), "is not a tree") {
				t.Fatalf("Path non-tree: err = %v, want non-tree error", err)
			}
		})
	})
}

func TestResolverPathDeepPath(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		const depth = 50

		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		leafBlobID := repoHarness.HashObject(t, "blob", []byte("deep-content\n"))
		currentTree := repoHarness.Mktree(t, fmt.Sprintf("100644 blob %s\tleaf.txt\n", leafBlobID))

		parts := make([][]byte, 0, depth+1)
		for i := depth - 1; i >= 0; i-- {
			name := fmt.Sprintf("level%02d", i)
			currentTree = repoHarness.Mktree(t, fmt.Sprintf("040000 tree %s\t%s\n", currentTree, name))
			parts = append([][]byte{[]byte(name)}, parts...)
		}

		parts = append(parts, []byte("leaf.txt"))

		repo := repoHarness.OpenRepository(t)

		entry, err := repo.Fetcher().Path(currentTree, parts)
		if err != nil {
			t.Fatalf("Path(deep): %v", err)
		}

		if entry.Mode != tree.FileModeRegular {
			t.Fatalf("Path(deep) mode = %o, want %o", entry.Mode, tree.FileModeRegular)
		}

		if entry.ID != leafBlobID {
			t.Fatalf("Path(deep) id = %s, want %s", entry.ID, leafBlobID)
		}
	})
}

func TestReadStoredTreeMixedModes(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		normalID := repoHarness.HashObject(t, "blob", []byte("normal-file\n"))
		execID := repoHarness.HashObject(t, "blob", []byte("#!/bin/sh\necho hi\n"))
		symID := repoHarness.HashObject(t, "blob", []byte("normal.txt"))
		nestedBlobID := repoHarness.HashObject(t, "blob", []byte("nested\n"))
		nestedTreeID := repoHarness.Mktree(t, fmt.Sprintf("100644 blob %s\tleaf.txt\n", nestedBlobID))

		rootTreeID := repoHarness.Mktree(t,
			fmt.Sprintf(
				"100644 blob %s\tnormal.txt\n100755 blob %s\trun.sh\n120000 blob %s\tlink.txt\n040000 tree %s\tdir\n",
				normalID,
				execID,
				symID,
				nestedTreeID,
			),
		)

		repo := repoHarness.OpenRepository(t)

		rootTree, err := repo.Fetcher().ExactTree(rootTreeID)
		if err != nil {
			t.Fatalf("ExactTree(root): %v", err)
		}

		expect := map[string]tree.FileMode{
			"normal.txt": tree.FileModeRegular,
			"run.sh":     tree.FileModeExecutable,
			"link.txt":   tree.FileModeSymlink,
			"dir":        tree.FileModeDir,
		}

		for name, wantMode := range expect {
			entry := rootTree.Object().Entry([]byte(name))

			if entry == nil {
				t.Fatalf("Entry(%q) returned nil", name)
			}

			if entry.Mode != wantMode {
				t.Fatalf("Entry(%q) mode = %o, want %o", name, entry.Mode, wantMode)
			}
		}
	})
}
