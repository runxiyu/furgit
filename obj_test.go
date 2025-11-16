package furgit

import (
	"fmt"
	"testing"
)

func TestObjectTypeSize(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	testData := []byte("Test data for size check")
	gitHash := gitHashObject(t, repoPath, "blob", testData)

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hash, _ := repo.ParseHash(gitHash)
	ty, size, err := repo.ReadObjectTypeSize(hash)
	if err != nil {
		t.Fatalf("ReadObjectTypeSize failed: %v", err)
	}

	if ty != ObjectTypeBlob {
		t.Errorf("type: got %d, want %d", ty, ObjectTypeBlob)
	}

	gitSize := string(gitCatFile(t, repoPath, "-s", gitHash))
	if size != int64(len(testData)) || gitSize != fmt.Sprintf("%d", size) {
		t.Errorf("size mismatch: furgit=%d git=%s expected=%d", size, gitSize, len(testData))
	}
}

func TestReadObjectInvalid(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	invalidHash, _ := repo.ParseHash("0000000000000000000000000000000000000000000000000000000000000000")
	_, err = repo.ReadObject(invalidHash)
	if err == nil {
		t.Error("expected error for invalid object")
	}
}
