package furgit

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommitWrite(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	blobHash := gitHashObject(t, repoPath, "blob", []byte("content"))

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	blobHashObj, _ := repo.ParseHash(blobHash)
	tree := &Tree{
		Entries: []TreeEntry{
			{Mode: 0o100644, Name: []byte("file.txt"), ID: blobHashObj},
		},
	}
	treeHash, _ := repo.WriteLooseObject(tree)

	whenUnix := time.Date(2023, 11, 16, 12, 0, 0, 0, time.UTC).Unix()
	commit := &Commit{
		Tree: treeHash,
		Author: Ident{
			Name:          []byte("Test Author"),
			Email:         []byte("test@example.org"),
			WhenUnix:      whenUnix,
			OffsetMinutes: 0,
		},
		Committer: Ident{
			Name:          []byte("Test Committer"),
			Email:         []byte("committer@example.org"),
			WhenUnix:      whenUnix,
			OffsetMinutes: 0,
		},
		Message: []byte("Initial commit\n"),
	}

	commitHash, err := repo.WriteLooseObject(commit)
	if err != nil {
		t.Fatalf("WriteLooseObject failed: %v", err)
	}

	gitType := string(gitCatFile(t, repoPath, "-t", commitHash.String()))
	if gitType != "commit" {
		t.Errorf("git type: got %q, want %q", gitType, "commit")
	}

	readObj, err := repo.ReadObject(commitHash)
	if err != nil {
		t.Fatalf("ReadObject failed after write: %v", err)
	}
	readCommit, ok := readObj.(*StoredCommit)
	if !ok {
		t.Fatalf("expected *StoredCommit, got %T", readObj)
	}

	if !bytes.HasPrefix(readCommit.Author.Name, []byte("Test Author")) {
		t.Errorf("author name: got %q, want prefix %q", readCommit.Author.Name, "Test Author")
	}
	if !bytes.Equal(readCommit.Message, []byte("Initial commit\n")) {
		t.Errorf("message: got %q, want %q", readCommit.Message, "Initial commit\n")
	}
}

func TestCommitRead(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "Test commit")
	commitHash := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	hash, _ := repo.ParseHash(commitHash)
	obj, err := repo.ReadObject(hash)
	if err != nil {
		t.Fatalf("ReadObject failed: %v", err)
	}

	commit, ok := obj.(*StoredCommit)
	if !ok {
		t.Fatalf("expected *StoredCommit, got %T", obj)
	}

	if !bytes.HasPrefix(commit.Author.Name, []byte("Test Author")) {
		t.Errorf("author name: got %q", commit.Author.Name)
	}
	if !bytes.Equal(commit.Author.Email, []byte("test@example.org")) {
		t.Errorf("author email: got %q", commit.Author.Email)
	}
	if !bytes.Equal(commit.Message, []byte("Test commit\n")) {
		t.Errorf("message: got %q", commit.Message)
	}
	if commit.ObjectType() != ObjectTypeCommit {
		t.Errorf("ObjectType(): got %d, want %d", commit.ObjectType(), ObjectTypeCommit)
	}
}

func TestCommitWithParents(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "file1.txt"), []byte("content1"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file1.txt: %v", err)
	}
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "First commit")
	parent1Hash := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")

	err = os.WriteFile(filepath.Join(workDir, "file2.txt"), []byte("content2"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file2.txt: %v", err)
	}
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "Second commit")
	parent2Hash := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")

	err = os.WriteFile(filepath.Join(workDir, "file3.txt"), []byte("content3"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file3.txt: %v", err)
	}
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	treeHash := gitCmd(t, repoPath, nil, "--work-tree="+workDir, "write-tree")

	mergeCommitData := fmt.Sprintf("tree %s\nparent %s\nparent %s\nauthor Test Author <test@example.org> 1234567890 +0000\ncommitter Test Committer <committer@example.org> 1234567890 +0000\n\nMerge commit\n",
		treeHash, parent1Hash, parent2Hash)

	cmd := gitHashObject(t, repoPath, "commit", []byte(mergeCommitData))
	mergeHash := cmd

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	hash, _ := repo.ParseHash(mergeHash)
	obj, _ := repo.ReadObject(hash)
	commit := obj.(*StoredCommit)

	if len(commit.Parents) != 2 {
		t.Fatalf("parents count: got %d, want 2", len(commit.Parents))
	}

	p1, _ := repo.ParseHash(parent1Hash)
	p2, _ := repo.ParseHash(parent2Hash)

	if commit.Parents[0] != p1 {
		t.Errorf("parent[0]: got %s, want %s", commit.Parents[0], parent1Hash)
	}
	if commit.Parents[1] != p2 {
		t.Errorf("parent[1]: got %s, want %s", commit.Parents[1], parent2Hash)
	}
}
