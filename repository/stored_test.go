package repository_test

import (
	"fmt"
	"os"
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

		root, err := os.OpenRoot(repoHarness.Dir())
		if err != nil {
			t.Fatalf("os.OpenRoot: %v", err)
		}

		defer func() { _ = root.Close() }()

		repo, err := repository.Open(root)
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

		root, err := os.OpenRoot(repoHarness.Dir())
		if err != nil {
			t.Fatalf("os.OpenRoot: %v", err)
		}

		defer func() { _ = root.Close() }()

		repo, err := repository.Open(root)
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

			root, err := os.OpenRoot(repoHarness.Dir())
			if err != nil {
				t.Fatalf("os.OpenRoot: %v", err)
			}

			defer func() { _ = root.Close() }()

			repo, err := repository.Open(root)
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

			root, err := os.OpenRoot(repoHarness.Dir())
			if err != nil {
				t.Fatalf("os.OpenRoot: %v", err)
			}

			defer func() { _ = root.Close() }()

			repo, err := repository.Open(root)
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

func TestResolveTreeEntryDeepPath(t *testing.T) {
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

		root, err := os.OpenRoot(repoHarness.Dir())
		if err != nil {
			t.Fatalf("os.OpenRoot: %v", err)
		}

		defer func() { _ = root.Close() }()

		repo, err := repository.Open(root)
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}

		defer func() { _ = repo.Close() }()

		rootTree, err := repo.ReadStoredTree(currentTree)
		if err != nil {
			t.Fatalf("ReadStoredTree(root): %v", err)
		}

		entry, err := repo.ResolveTreeEntry(rootTree, parts)
		if err != nil {
			t.Fatalf("ResolveTreeEntry(deep): %v", err)
		}

		if entry.Mode != object.FileModeRegular {
			t.Fatalf("ResolveTreeEntry(deep) mode = %o, want %o", entry.Mode, object.FileModeRegular)
		}

		if entry.ID != leafBlobID {
			t.Fatalf("ResolveTreeEntry(deep) id = %s, want %s", entry.ID, leafBlobID)
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

		root, err := os.OpenRoot(repoHarness.Dir())
		if err != nil {
			t.Fatalf("os.OpenRoot: %v", err)
		}

		defer func() { _ = root.Close() }()

		repo, err := repository.Open(root)
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}

		defer func() { _ = repo.Close() }()

		rootTree, err := repo.ReadStoredTree(rootTreeID)
		if err != nil {
			t.Fatalf("ReadStoredTree(root): %v", err)
		}

		expect := map[string]object.FileMode{
			"normal.txt": object.FileModeRegular,
			"run.sh":     object.FileModeExecutable,
			"link.txt":   object.FileModeSymlink,
			"dir":        object.FileModeDir,
		}

		for name, wantMode := range expect {
			entry := rootTree.Tree().Entry([]byte(name))

			if entry == nil {
				t.Fatalf("Entry(%q) returned nil", name)
			}

			if entry.Mode != wantMode {
				t.Fatalf("Entry(%q) mode = %o, want %o", name, entry.Mode, wantMode)
			}
		}
	})
}
