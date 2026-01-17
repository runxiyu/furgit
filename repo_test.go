package furgit

import (
	"testing"
)

func TestRepositoryOpen(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	if repo.rootPath != repoPath {
		t.Errorf("rootPath: got %q, want %q", repo.rootPath, repoPath)
	}
	hashSize := repo.hashAlgo.size()
	if hashSize != 32 && hashSize != 20 {
		t.Errorf("hashSize: got %d, want 32 (SHA-256) or 20 (SHA-1)", hashSize)
	}
}

func TestRepositoryOpenInvalid(t *testing.T) {
	_, err := OpenRepository("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestRepositoryClose(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}

	if err := repo.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if err := repo.Close(); err != nil {
		t.Fatalf("second Close failed: %v", err)
	}
}
