package furgit

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTreeWrite(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	blobData := []byte("file content")
	blobHash := gitHashObject(t, repoPath, "blob", blobData)

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	blobHashObj, _ := repo.ParseHash(blobHash)
	tree := &Tree{
		Entries: []TreeEntry{
			{Mode: 0o100644, Name: []byte("file.txt"), ID: blobHashObj},
		},
	}

	treeHash, err := repo.WriteLooseObject(tree)
	if err != nil {
		t.Fatalf("WriteLooseObject failed: %v", err)
	}

	gitType := string(gitCatFile(t, repoPath, "-t", treeHash.String()))
	if gitType != "tree" {
		t.Errorf("git type: got %q, want %q", gitType, "tree")
	}

	gitLsTree := gitCmd(t, repoPath, "ls-tree", treeHash.String())
	if !strings.Contains(gitLsTree, "file.txt") {
		t.Errorf("git ls-tree doesn't contain file.txt: %s", gitLsTree)
	}
	if !strings.Contains(gitLsTree, blobHash) {
		t.Errorf("git ls-tree doesn't contain blob hash: %s", gitLsTree)
	}
}

func TestTreeRead(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "a.txt"), []byte("content a"), 0o644)
	if err != nil {
		t.Fatalf("failed to write a.txt: %v", err)
	}
	err = os.WriteFile(filepath.Join(workDir, "b.txt"), []byte("content b"), 0o644)
	if err != nil {
		t.Fatalf("failed to write b.txt: %v", err)
	}
	err = os.WriteFile(filepath.Join(workDir, "c.txt"), []byte("content c"), 0o644)
	if err != nil {
		t.Fatalf("failed to write c.txt: %v", err)
	}

	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	treeHash := gitCmd(t, repoPath, "--work-tree="+workDir, "write-tree")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hash, _ := repo.ParseHash(treeHash)
	obj, err := repo.ReadObject(hash)
	if err != nil {
		t.Fatalf("ReadObject failed: %v", err)
	}

	tree, ok := obj.(*StoredTree)
	if !ok {
		t.Fatalf("expected *StoredTree, got %T", obj)
	}

	if len(tree.Entries) != 3 {
		t.Fatalf("entries count: got %d, want 3", len(tree.Entries))
	}

	expectedNames := []string{"a.txt", "b.txt", "c.txt"}
	for i, expected := range expectedNames {
		if string(tree.Entries[i].Name) != expected {
			t.Errorf("entry[%d] name: got %q, want %q", i, tree.Entries[i].Name, expected)
		}
	}

	if tree.ObjectType() != ObjectTypeTree {
		t.Errorf("ObjectType(): got %d, want %d", tree.ObjectType(), ObjectTypeTree)
	}
}

func TestTreeEntry(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "a.txt"), []byte("content a"), 0o644)
	if err != nil {
		t.Fatalf("failed to write a.txt: %v", err)
	}
	err = os.WriteFile(filepath.Join(workDir, "b.txt"), []byte("content b"), 0o644)
	if err != nil {
		t.Fatalf("failed to write b.txt: %v", err)
	}
	err = os.WriteFile(filepath.Join(workDir, "c.txt"), []byte("content c"), 0o644)
	if err != nil {
		t.Fatalf("failed to write c.txt: %v", err)
	}

	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	treeHash := gitCmd(t, repoPath, "--work-tree="+workDir, "write-tree")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hash, _ := repo.ParseHash(treeHash)
	obj, _ := repo.ReadObject(hash)
	tree := obj.(*StoredTree)

	entry := tree.Entry([]byte("b.txt"))
	if entry == nil {
		t.Fatal("Entry returned nil for existing entry")
	}
	if !bytes.Equal(entry.Name, []byte("b.txt")) {
		t.Errorf("entry name: got %q, want %q", entry.Name, "b.txt")
	}

	notFound := tree.Entry([]byte("notfound.txt"))
	if notFound != nil {
		t.Error("Entry returned non-nil for non-existing entry")
	}
}

func TestTreeEntryRecursive(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.MkdirAll(filepath.Join(workDir, "dir"), 0o755)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(workDir, "file1.txt"), []byte("file1"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file1.txt: %v", err)
	}
	err = os.WriteFile(filepath.Join(workDir, "file2.txt"), []byte("file2"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file2.txt: %v", err)
	}
	err = os.WriteFile(filepath.Join(workDir, "dir", "nested.txt"), []byte("nested"), 0o644)
	if err != nil {
		t.Fatalf("failed to write dir/nested.txt: %v", err)
	}

	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	treeHash := gitCmd(t, repoPath, "--work-tree="+workDir, "write-tree")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hash, _ := repo.ParseHash(treeHash)
	obj, _ := repo.ReadObject(hash)
	tree := obj.(*StoredTree)

	entry, err := tree.EntryRecursive(repo, [][]byte{[]byte("file1.txt")})
	if err != nil {
		t.Fatalf("EntryRecursive file1.txt failed: %v", err)
	}
	if !bytes.Equal(entry.Name, []byte("file1.txt")) {
		t.Errorf("entry name: got %q, want %q", entry.Name, "file1.txt")
	}

	gitShow := string(gitCatFile(t, repoPath, "blob", entry.ID.String()))
	if gitShow != "file1" {
		t.Errorf("file1 content from git: got %q, want %q", gitShow, "file1")
	}

	nestedEntry, err := tree.EntryRecursive(repo, [][]byte{[]byte("dir"), []byte("nested.txt")})
	if err != nil {
		t.Fatalf("EntryRecursive dir/nested.txt failed: %v", err)
	}
	if !bytes.Equal(nestedEntry.Name, []byte("nested.txt")) {
		t.Errorf("nested entry name: got %q, want %q", nestedEntry.Name, "nested.txt")
	}

	gitShowNested := string(gitCatFile(t, repoPath, "blob", nestedEntry.ID.String()))
	if gitShowNested != "nested" {
		t.Errorf("nested content from git: got %q, want %q", gitShowNested, "nested")
	}

	_, err = tree.EntryRecursive(repo, [][]byte{[]byte("nonexistent.txt")})
	if err == nil {
		t.Error("expected error for nonexistent path")
	}

	_, err = tree.EntryRecursive(repo, [][]byte{})
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestTreeLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large tree test in short mode")
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
	treeHash := gitCmd(t, repoPath, "--work-tree="+workDir, "write-tree")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hash, _ := repo.ParseHash(treeHash)
	obj, _ := repo.ReadObject(hash)
	tree := obj.(*StoredTree)

	if len(tree.Entries) != numFiles {
		t.Errorf("tree entries: got %d, want %d", len(tree.Entries), numFiles)
	}

	gitCount := gitCmd(t, repoPath, "ls-tree", treeHash)
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
