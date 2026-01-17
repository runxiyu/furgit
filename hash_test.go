package furgit

import (
	"testing"
)

func TestHashParse(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	var validHash string
	var expectedSize int
	if repo.hashAlgo.size() == 32 {
		validHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
		expectedSize = 32
	} else {
		validHash = "0123456789abcdef0123456789abcdef01234567"
		expectedSize = 20
	}

	hash, err := repo.ParseHash(validHash)
	if err != nil {
		t.Fatalf("ParseHash failed: %v", err)
	}
	if hash.String() != validHash {
		t.Errorf("String(): got %q, want %q", hash.String(), validHash)
	}
	if hash.Size() != expectedSize {
		t.Errorf("Size(): got %d, want %d", hash.Size(), expectedSize)
	}

	hashBytes := hash.Bytes()
	if len(hashBytes) != expectedSize {
		t.Errorf("Bytes() length: got %d, want %d", len(hashBytes), expectedSize)
	}
}

func TestHashParseErrors(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	tests := []struct {
		name string
		hash string
	}{
		{"invalid chars", "invalid"},
		{"wrong length", "0123456789abcdef"},
		{"non-hex", "0123456789abcdefg123456789abcdef0123456789abcdef0123456789abcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.ParseHash(tt.hash)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}
