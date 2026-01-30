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

	gitLsTree := gitCmd(t, repoPath, nil, "ls-tree", treeHash.String())
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

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	treeHash := gitCmd(t, repoPath, nil, "--work-tree="+workDir, "write-tree")

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

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	treeHash := gitCmd(t, repoPath, nil, "--work-tree="+workDir, "write-tree")

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

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	treeHash := gitCmd(t, repoPath, nil, "--work-tree="+workDir, "write-tree")

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

	gitCmd(t, repoPath, nil, "config", "gc.auto", "0")

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

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	treeHash := gitCmd(t, repoPath, nil, "--work-tree="+workDir, "write-tree")

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

	gitCount := gitCmd(t, repoPath, nil, "ls-tree", treeHash)
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

func TestTreeInsertEntry(t *testing.T) {
	tree := &Tree{
		Entries: []TreeEntry{
			{Mode: FileModeRegular, Name: []byte("alpha"), ID: Hash{}},
			{Mode: FileModeRegular, Name: []byte("gamma"), ID: Hash{}},
		},
	}

	if err := tree.InsertEntry(TreeEntry{Mode: FileModeRegular, Name: []byte("beta"), ID: Hash{}}); err != nil {
		t.Fatalf("InsertEntry failed: %v", err)
	}
	if len(tree.Entries) != 3 {
		t.Fatalf("entries count: got %d, want 3", len(tree.Entries))
	}
	if string(tree.Entries[1].Name) != "beta" {
		t.Fatalf("inserted order mismatch: got %q, want %q", tree.Entries[1].Name, "beta")
	}

	if err := tree.InsertEntry(TreeEntry{Mode: FileModeRegular, Name: []byte("beta"), ID: Hash{}}); err == nil {
		t.Fatal("expected duplicate insert error")
	}

	var nilTree *Tree
	if err := nilTree.InsertEntry(TreeEntry{Mode: FileModeRegular, Name: []byte("x"), ID: Hash{}}); err == nil {
		t.Fatal("expected error for nil tree")
	}
}

func TestTreeRemoveEntry(t *testing.T) {
	tree := &Tree{
		Entries: []TreeEntry{
			{Mode: FileModeRegular, Name: []byte("alpha"), ID: Hash{}},
			{Mode: FileModeRegular, Name: []byte("beta"), ID: Hash{}},
			{Mode: FileModeRegular, Name: []byte("gamma"), ID: Hash{}},
		},
	}

	if err := tree.RemoveEntry([]byte("beta")); err != nil {
		t.Fatalf("RemoveEntry failed: %v", err)
	}
	if len(tree.Entries) != 2 {
		t.Fatalf("entries count: got %d, want 2", len(tree.Entries))
	}
	if string(tree.Entries[0].Name) != "alpha" || string(tree.Entries[1].Name) != "gamma" {
		t.Fatalf("remove order mismatch: got %q, %q", tree.Entries[0].Name, tree.Entries[1].Name)
	}

	if err := tree.RemoveEntry([]byte("beta")); err == nil {
		t.Fatal("expected ErrNotFound for missing entry")
	}

	var nilTree *Tree
	if err := nilTree.RemoveEntry([]byte("alpha")); err == nil {
		t.Fatal("expected error for nil tree")
	}
}

func TestTreeEntryNameCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		entryName    []byte
		entryMode    FileMode
		searchName   []byte
		searchIsTree bool
		want         int
	}{
		{
			name:       "equal file names",
			entryName:  []byte("alpha"),
			entryMode:  FileModeRegular,
			searchName: []byte("alpha"),
			want:       0,
		},
		{
			name:         "equal tree names",
			entryName:    []byte("dir"),
			entryMode:    FileModeDir,
			searchName:   []byte("dir"),
			searchIsTree: true,
			want:         0,
		},
		{
			name:       "lexicographic less",
			entryName:  []byte("alpha"),
			entryMode:  FileModeRegular,
			searchName: []byte("beta"),
			want:       -1,
		},
		{
			name:       "lexicographic greater",
			entryName:  []byte("gamma"),
			entryMode:  FileModeRegular,
			searchName: []byte("beta"),
			want:       1,
		},
		{
			name:         "file sorts before same-name dir",
			entryName:    []byte("same"),
			entryMode:    FileModeRegular,
			searchName:   []byte("same"),
			searchIsTree: true,
			want:         -1,
		},
		{
			name:         "dir sorts after same-name file",
			entryName:    []byte("same"),
			entryMode:    FileModeDir,
			searchName:   []byte("same"),
			searchIsTree: false,
			want:         1,
		},
		{
			name:         "dir sorts before longer file",
			entryName:    []byte("a"),
			entryMode:    FileModeDir,
			searchName:   []byte("ab"),
			searchIsTree: false,
			want:         -1,
		},
		{
			name:       "file sorts before longer file",
			entryName:  []byte("a"),
			entryMode:  FileModeRegular,
			searchName: []byte("ab"),
			want:       -1,
		},
		{
			name:         "search tree compares after exact file name",
			entryName:    []byte("a"),
			entryMode:    FileModeRegular,
			searchName:   []byte("a"),
			searchIsTree: true,
			want:         -1,
		},
		{
			name:         "entry tree compares after exact search file",
			entryName:    []byte("a"),
			entryMode:    FileModeDir,
			searchName:   []byte("a"),
			searchIsTree: false,
			want:         1,
		},
		{
			name:         "slash impact mid-compare",
			entryName:    []byte("a"),
			entryMode:    FileModeDir,
			searchName:   []byte("a0"),
			searchIsTree: false,
			want:         -1,
		},
		{
			name:         "file sorts after same prefix dir",
			entryName:    []byte("a0"),
			entryMode:    FileModeRegular,
			searchName:   []byte("a"),
			searchIsTree: true,
			want:         1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := TreeEntryNameCompare(tt.entryName, tt.entryMode, tt.searchName, tt.searchIsTree)
			if got < 0 {
				got = -1
			} else if got > 0 {
				got = 1
			}
			if got != tt.want {
				t.Fatalf("compare(%q,%v,%q,%v) = %d, want %d", tt.entryName, tt.entryMode, tt.searchName, tt.searchIsTree, got, tt.want)
			}
		})
	}
}
