package furgit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveRef(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "test.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "test")
	commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")
	gitCmd(t, repoPath, "update-ref", "refs/heads/main", commitHash)

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hashObj, _ := repo.ParseHash(commitHash)
	resolved, err := repo.ResolveRef("refs/heads/main")
	if err != nil {
		t.Fatalf("ResolveRef failed: %v", err)
	}

	if resolved.Kind != RefKindDetached {
		t.Fatalf("expected detached ref, got %v", resolved.Kind)
	}
	if resolved.Hash != hashObj {
		t.Errorf("resolved hash: got %s, want %s", resolved.Hash, hashObj)
	}

	gitRevParse := gitCmd(t, repoPath, "rev-parse", "refs/heads/main")
	if resolved.Hash.String() != gitRevParse {
		t.Errorf("furgit resolved %s, git resolved %s", resolved.Hash, gitRevParse)
	}

	_, err = repo.ResolveRef("refs/heads/nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent ref")
	}
}

func TestResolveHEAD(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "test.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write test.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "test")
	commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")
	gitCmd(t, repoPath, "update-ref", "refs/heads/main", commitHash)
	gitCmd(t, repoPath, "symbolic-ref", "HEAD", "refs/heads/main")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	ref, err := repo.ResolveRef("HEAD")
	if err != nil {
		t.Fatalf("ResolveRef(HEAD) failed: %v", err)
	}

	if ref.Kind != RefKindSymbolic {
		t.Fatalf("HEAD kind: got %v, want %v", ref.Kind, RefKindSymbolic)
	}

	if ref.Ref != "refs/heads/main" {
		t.Errorf("HEAD symbolic ref: got %q, want %q", ref.Ref, "refs/heads/main")
	}

	gitSymRef := gitCmd(t, repoPath, "symbolic-ref", "HEAD")
	if ref.Ref != gitSymRef {
		t.Errorf("furgit resolved %v, git resolved %s", ref.Ref, gitSymRef)
	}
}

func TestPackedRefs(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "test.txt"), []byte("content1"), 0o644)
	if err != nil {
		t.Fatalf("failed to write test.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "commit1")
	commit1Hash := gitCmd(t, repoPath, "rev-parse", "HEAD")

	err = os.WriteFile(filepath.Join(workDir, "test2.txt"), []byte("content2"), 0o644)
	if err != nil {
		t.Fatalf("failed to write test2.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "commit2")
	commit2Hash := gitCmd(t, repoPath, "rev-parse", "HEAD")

	gitCmd(t, repoPath, "update-ref", "refs/heads/branch1", commit1Hash)
	gitCmd(t, repoPath, "update-ref", "refs/heads/branch2", commit2Hash)
	gitCmd(t, repoPath, "update-ref", "refs/tags/v1.0", commit1Hash)

	gitCmd(t, repoPath, "pack-refs", "--all")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hash1, _ := repo.ParseHash(commit1Hash)
	hash2, _ := repo.ParseHash(commit2Hash)

	resolved1, err := repo.ResolveRef("refs/heads/branch1")
	if err != nil {
		t.Fatalf("ResolveRef branch1 failed: %v", err)
	}
	if resolved1.Kind != RefKindDetached || resolved1.Hash != hash1 {
		t.Errorf("branch1: got %s, want %s", resolved1.Hash, hash1)
	}

	gitResolved1 := gitCmd(t, repoPath, "rev-parse", "refs/heads/branch1")
	if resolved1.Hash.String() != gitResolved1 {
		t.Errorf("furgit resolved %s, git resolved %s", resolved1.Hash, gitResolved1)
	}

	resolved2, err := repo.ResolveRef("refs/heads/branch2")
	if err != nil {
		t.Fatalf("ResolveRef branch2 failed: %v", err)
	}
	if resolved2.Kind != RefKindDetached || resolved2.Hash != hash2 {
		t.Errorf("branch2: got %s, want %s", resolved2.Hash, hash2)
	}

	resolvedTag, err := repo.ResolveRef("refs/tags/v1.0")
	if err != nil {
		t.Fatalf("ResolveRef tag failed: %v", err)
	}
	if resolvedTag.Kind != RefKindDetached || resolvedTag.Hash != hash1 {
		t.Errorf("tag: got %s, want %s", resolvedTag.Hash, hash1)
	}
}

func TestResolveRefFully(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	// Create an initial commit
	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "init")
	commit := gitCmd(t, repoPath, "rev-parse", "HEAD")

	// Create two layers of symbolic refs
	gitCmd(t, repoPath, "symbolic-ref", "refs/heads/level1", "refs/heads/level2")
	gitCmd(t, repoPath, "symbolic-ref", "refs/heads/level2", "refs/heads/main")
	gitCmd(t, repoPath, "update-ref", "refs/heads/main", commit)

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	commitHash, err := repo.ParseHash(commit)
	if err != nil {
		t.Fatalf("ParseHash failed: %v", err)
	}

	resolved, err := repo.ResolveRefFully("refs/heads/level1")
	if err != nil {
		t.Fatalf("ResolveRefFully failed: %v", err)
	}

	if resolved != commitHash {
		t.Errorf("ResolveRefFully: got hash %s, want %s", resolved, commitHash)
	}
}

func TestResolveRefFullySymbolicCycle(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	gitCmd(t, repoPath, "symbolic-ref", "refs/heads/A", "refs/heads/B")
	gitCmd(t, repoPath, "symbolic-ref", "refs/heads/B", "refs/heads/A")

	_, err = repo.ResolveRefFully("refs/heads/A")
	if err == nil {
		t.Fatalf("ResolveRefFully should fail on a symbolic cycle")
	}

	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("unexpected error for symbolic cycle: %v", err)
	}
}

func TestResolveRefHashInput(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "init")

	commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hashObj, err := repo.ParseHash(commitHash)
	if err != nil {
		t.Fatalf("ParseHash failed: %v", err)
	}

	ref, err := repo.ResolveRef(commitHash)
	if err != nil {
		t.Fatalf("ResolveRef(hash) failed: %v", err)
	}
	if ref.Kind != RefKindDetached {
		t.Fatalf("expected RefKindDetached, got %v", ref.Kind)
	}
	if ref.Hash != hashObj {
		t.Fatalf("hash mismatch: got %s, want %s", ref.Hash, hashObj)
	}

	hash, err := repo.ResolveRefFully(commitHash)
	if err != nil {
		t.Fatalf("ResolveRefFully(hash) failed: %v", err)
	}
	if hash != hashObj {
		t.Fatalf("hash mismatch: got %s, want %s", hash, hashObj)
	}

	_, err = repo.ResolveRef("this_is_not_a_hash")
	if err == nil {
		t.Fatalf("expected error for invalid hash input")
	}
}
