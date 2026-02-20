package furgit

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestTreeNestedDeep(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	depth := 50
	currentDir := workDir
	for i := 0; i < depth; i++ {
		currentDir = filepath.Join(currentDir, fmt.Sprintf("level%d", i))
		err := os.MkdirAll(currentDir, 0o755)
		if err != nil {
			t.Fatalf("failed to create directory %s: %v", currentDir, err)
		}
	}
	err := os.WriteFile(filepath.Join(currentDir, "deep.txt"), []byte("deep content"), 0o644)
	if err != nil {
		t.Fatalf("failed to create deep.txt: %v", err)
	}

	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	treeHash := gitCmd(t, repoPath, "--work-tree="+workDir, "write-tree")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	hash, _ := repo.ParseHash(treeHash)
	obj, _ := repo.ReadObject(hash)
	tree := obj.(*StoredTree)

	path := make([][]byte, depth+1)
	for i := 0; i < depth; i++ {
		path[i] = []byte(fmt.Sprintf("level%d", i))
	}
	path[depth] = []byte("deep.txt")

	entry, err := tree.EntryRecursive(repo, path)
	if err != nil {
		t.Fatalf("EntryRecursive failed for deep path: %v", err)
	}

	blobObj, _ := repo.ReadObject(entry.ID)
	blob := blobObj.(*StoredBlob)

	if !bytes.Equal(blob.Data, []byte("deep content")) {
		t.Errorf("deep file content: got %q, want %q", blob.Data, "deep content")
	}
}

func TestTreeMixedModes(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "normal.txt"), []byte("normal"), 0o644)
	if err != nil {
		t.Fatalf("failed to create normal.txt: %v", err)
	}
	err = os.WriteFile(filepath.Join(workDir, "executable.sh"), []byte("#!/bin/sh\necho test"), 0o755)
	if err != nil {
		t.Fatalf("failed to create executable.sh: %v", err)
	}
	err = os.Symlink("normal.txt", filepath.Join(workDir, "link.txt"))
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	treeHash := gitCmd(t, repoPath, "--work-tree="+workDir, "write-tree")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	hash, _ := repo.ParseHash(treeHash)
	obj, _ := repo.ReadObject(hash)
	tree := obj.(*StoredTree)

	modes := make(map[string]FileMode)
	for _, entry := range tree.Entries {
		modes[string(entry.Name)] = entry.Mode
	}

	if modes["normal.txt"] != 0o100644 {
		t.Errorf("normal.txt mode: got %o, want %o", modes["normal.txt"], 0o100644)
	}
	if modes["executable.sh"] != 0o100755 {
		t.Errorf("executable.sh mode: got %o, want %o", modes["executable.sh"], 0o100755)
	}
	if modes["link.txt"] != 0o120000 {
		t.Errorf("link.txt mode: got %o, want %o", modes["link.txt"], 0o120000)
	}
}

func TestCommitChain(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	numCommits := 100
	var commits []string

	for i := 0; i < numCommits; i++ {
		filename := filepath.Join(workDir, fmt.Sprintf("file%d.txt", i))
		err := os.WriteFile(filename, []byte(fmt.Sprintf("content %d", i)), 0o644)
		if err != nil {
			t.Fatalf("failed to create %s: %v", filename, err)
		}
		gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
		gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", fmt.Sprintf("Commit %d", i))
		commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")
		commits = append(commits, commitHash)
	}

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	hash, _ := repo.ParseHash(commits[len(commits)-1])
	for i := numCommits - 1; i >= 0; i-- {
		obj, err := repo.ReadObject(hash)
		if err != nil {
			t.Fatalf("failed to read commit %d: %v", i, err)
		}

		commit, ok := obj.(*StoredCommit)
		if !ok {
			t.Fatalf("expected *StoredCommit at %d, got %T", i, obj)
		}

		expectedMsg := fmt.Sprintf("Commit %d\n", i)
		if !bytes.Equal(commit.Message, []byte(expectedMsg)) {
			t.Errorf("commit %d message: got %q, want %q", i, commit.Message, expectedMsg)
		}

		if i > 0 {
			if len(commit.Parents) != 1 {
				t.Fatalf("commit %d should have 1 parent, got %d", i, len(commit.Parents))
			}
			hash = commit.Parents[0]
		} else {
			if len(commit.Parents) != 0 {
				t.Errorf("first commit should have 0 parents, got %d", len(commit.Parents))
			}
		}
	}
}

func TestMultipleTags(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to create file.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "Tagged commit")
	commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")

	tags := []string{"v1.0.0", "v1.0.1", "v1.1.0", "v2.0.0"}
	for _, tagName := range tags {
		gitCmd(t, repoPath, "tag", "-a", "-m", fmt.Sprintf("Release %s", tagName), tagName, commitHash)
	}

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	for _, tagName := range tags {
		tagHash := gitCmd(t, repoPath, "rev-parse", tagName)
		hash, _ := repo.ParseHash(tagHash)
		obj, err := repo.ReadObject(hash)
		if err != nil {
			t.Errorf("failed to read tag %s: %v", tagName, err)
			continue
		}

		tag, ok := obj.(*StoredTag)
		if !ok {
			t.Errorf("tag %s: expected *StoredTag, got %T", tagName, obj)
			continue
		}

		if !bytes.Equal(tag.Name, []byte(tagName)) {
			t.Errorf("tag name: got %q, want %q", tag.Name, tagName)
		}
	}
}

func TestPackfileAfterMultipleRepacks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multiple repack test in short mode")
	}

	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	gitCmd(t, repoPath, "config", "gc.auto", "0")

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	for i := 0; i < 5; i++ {
		err := os.WriteFile(filepath.Join(workDir, fmt.Sprintf("file%d.txt", i)), []byte(fmt.Sprintf("content %d", i)), 0o644)
		if err != nil {
			t.Fatalf("failed to create file%d.txt: %v", i, err)
		}
		gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
		gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", fmt.Sprintf("Commit %d", i))
		gitCmd(t, repoPath, "repack", "-d")
	}

	gitCmd(t, repoPath, "repack", "-a", "-d")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	headHash := gitCmd(t, repoPath, "rev-parse", "HEAD")
	hash, _ := repo.ParseHash(headHash)

	obj, err := repo.ReadObject(hash)
	if err != nil {
		t.Fatalf("failed to read HEAD from final packfile: %v", err)
	}

	commit := obj.(*StoredCommit)
	if !bytes.Contains(commit.Message, []byte("Commit 4")) {
		t.Errorf("HEAD commit message incorrect: got %q", commit.Message)
	}
}
