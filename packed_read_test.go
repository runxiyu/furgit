package furgit

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackfileRead(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	gitCmd(t, repoPath, "config", "gc.auto", "0")

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "file1.txt"), []byte("content1"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file1.txt: %v", err)
	}
	err = os.WriteFile(filepath.Join(workDir, "file2.txt"), []byte("content2"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file2.txt: %v", err)
	}

	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "Test commit")
	commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")

	gitCmd(t, repoPath, "repack", "-a", "-d")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hashObj, _ := repo.ParseHash(commitHash)
	obj, err := repo.ReadObject(hashObj)
	if err != nil {
		t.Fatalf("ReadObject from pack failed: %v", err)
	}

	commit, ok := obj.(*StoredCommit)
	if !ok {
		t.Fatalf("expected *StoredCommit, got %T", obj)
	}

	treeObj, err := repo.ReadObject(commit.Tree)
	if err != nil {
		t.Fatalf("ReadObject tree failed: %v", err)
	}

	tree, ok := treeObj.(*StoredTree)
	if !ok {
		t.Fatalf("expected *StoredTree, got %T", treeObj)
	}

	if len(tree.Entries) != 2 {
		t.Errorf("tree entries: got %d, want 2", len(tree.Entries))
	}

	gitLsTree := gitCmd(t, repoPath, "ls-tree", commit.Tree.String())
	for _, entry := range tree.Entries {
		if !strings.Contains(gitLsTree, string(entry.Name)) {
			t.Errorf("git ls-tree doesn't contain %s", entry.Name)
		}
	}
}

func TestPackfileLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large packfile test in short mode")
	}

	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	gitCmd(t, repoPath, "config", "gc.auto", "0")

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	numFiles := 1000
	for i := 0; i < numFiles; i++ {
		filename := filepath.Join(workDir, fmt.Sprintf("file%04d.txt", i))
		content := fmt.Sprintf("Content for file %d\n", i)
		err := os.WriteFile(filename, []byte(content), 0o644)
		if err != nil {
			t.Fatalf("failed to write %s: %v", filename, err)
		}
	}

	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "Large commit")
	commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")

	gitCmd(t, repoPath, "repack", "-a", "-d")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hashObj, _ := repo.ParseHash(commitHash)
	obj, _ := repo.ReadObject(hashObj)
	commit := obj.(*StoredCommit)

	treeObj, _ := repo.ReadObject(commit.Tree)
	tree := treeObj.(*StoredTree)

	if len(tree.Entries) != numFiles {
		t.Errorf("tree entries: got %d, want %d", len(tree.Entries), numFiles)
	}

	gitCount := gitCmd(t, repoPath, "ls-tree", commit.Tree.String())
	gitLines := strings.Count(gitCount, "\n") + 1
	if len(tree.Entries) != gitLines {
		t.Errorf("furgit found %d entries, git found %d", len(tree.Entries), gitLines)
	}

	for i := 0; i < 10; i++ {
		idx := i * (numFiles / 10)
		expectedName := fmt.Sprintf("file%04d.txt", idx)
		entry := tree.Entry([]byte(expectedName))
		if entry == nil {
			t.Errorf("expected to find entry %s", expectedName)
			continue
		}

		blobObj, _ := repo.ReadObject(entry.ID)
		blob := blobObj.(*StoredBlob)

		expectedContent := fmt.Sprintf("Content for file %d\n", idx)
		if string(blob.Data) != expectedContent {
			t.Errorf("blob %s: got %q, want %q", expectedName, blob.Data, expectedContent)
		}

		gitData := gitCatFile(t, repoPath, "blob", entry.ID.String())
		if !bytes.Equal(blob.Data, gitData) {
			t.Errorf("blob %s: furgit data doesn't match git data", expectedName)
		}
	}
}
