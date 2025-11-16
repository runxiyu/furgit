package furgit

import (
	"os"
	"path/filepath"
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

	if resolved != hashObj {
		t.Errorf("resolved hash: got %s, want %s", resolved, hashObj)
	}

	gitRevParse := gitCmd(t, repoPath, "rev-parse", "refs/heads/main")
	if resolved.String() != gitRevParse {
		t.Errorf("furgit resolved %s, git resolved %s", resolved, gitRevParse)
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

	ref, err := repo.ResolveHead()
	if err != nil {
		t.Fatalf("ResolveHEAD failed: %v", err)
	}

	switch ref.Kind {
	case HeadKindSymbolic:
		if ref.Ref != "refs/heads/main" {
			t.Errorf("HEAD ref: got %q, want %q", ref, "refs/heads/main")
		}
		gitSymRef := gitCmd(t, repoPath, "symbolic-ref", "HEAD")
		if ref.Ref != gitSymRef {
			t.Errorf("furgit resolved %v, git resolved %s", ref, gitSymRef)
		}
	default:
		t.Errorf("HEAD kind: got %v, want %v", ref.Kind, HeadKindSymbolic)
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
	if resolved1 != hash1 {
		t.Errorf("branch1: got %s, want %s", resolved1, hash1)
	}

	gitResolved1 := gitCmd(t, repoPath, "rev-parse", "refs/heads/branch1")
	if resolved1.String() != gitResolved1 {
		t.Errorf("furgit resolved %s, git resolved %s", resolved1, gitResolved1)
	}

	resolved2, err := repo.ResolveRef("refs/heads/branch2")
	if err != nil {
		t.Fatalf("ResolveRef branch2 failed: %v", err)
	}
	if resolved2 != hash2 {
		t.Errorf("branch2: got %s, want %s", resolved2, hash2)
	}

	resolvedTag, err := repo.ResolveRef("refs/tags/v1.0")
	if err != nil {
		t.Fatalf("ResolveRef tag failed: %v", err)
	}
	if resolvedTag != hash1 {
		t.Errorf("tag: got %s, want %s", resolvedTag, hash1)
	}
}
