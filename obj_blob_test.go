package furgit

import (
	"bytes"
	"fmt"
	"testing"
)

func TestBlobRead(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	testData := []byte("Hello, Furgit!\nThis is test blob data.")
	gitHash := gitHashObject(t, repoPath, "blob", testData)

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	hash, _ := repo.ParseHash(gitHash)
	obj, err := repo.ReadObject(hash)
	if err != nil {
		t.Fatalf("ReadObject failed: %v", err)
	}

	blob, ok := obj.(*StoredBlob)
	if !ok {
		t.Fatalf("expected *StoredBlob, got %T", obj)
	}

	if !bytes.Equal(blob.Data, testData) {
		t.Errorf("Data mismatch: got %q, want %q", blob.Data, testData)
	}
	if blob.Hash() != hash {
		t.Errorf("Hash(): got %s, want %s", blob.Hash(), hash)
	}
	if blob.ObjectType() != ObjectTypeBlob {
		t.Errorf("ObjectType(): got %d, want %d", blob.ObjectType(), ObjectTypeBlob)
	}

	gitData := gitCatFile(t, repoPath, "blob", gitHash)
	if !bytes.Equal(blob.Data, gitData) {
		t.Error("furgit data doesn't match git data")
	}
}

func TestBlobWrite(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	testData := []byte("Test data written by furgit")
	blob := &Blob{Data: testData}

	hash, err := repo.WriteLooseObject(blob)
	if err != nil {
		t.Fatalf("WriteLooseObject failed: %v", err)
	}

	gitType := string(gitCatFile(t, repoPath, "-t", hash.String()))
	if gitType != "blob" {
		t.Errorf("git type: got %q, want %q", gitType, "blob")
	}

	gitData := gitCatFile(t, repoPath, "blob", hash.String())
	if !bytes.Equal(gitData, testData) {
		t.Error("git data doesn't match written data")
	}

	gitSize := string(gitCatFile(t, repoPath, "-s", hash.String()))
	if gitSize != fmt.Sprintf("%d", len(testData)) {
		t.Errorf("git size: got %s, want %d", gitSize, len(testData))
	}
}

func TestBlobRoundtrip(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	testData := []byte("Roundtrip test data")
	blob := &Blob{Data: testData}

	hash, err := repo.WriteLooseObject(blob)
	if err != nil {
		t.Fatalf("WriteLooseObject failed: %v", err)
	}

	obj, err := repo.ReadObject(hash)
	if err != nil {
		t.Fatalf("ReadObject failed: %v", err)
	}

	readBlob, ok := obj.(*StoredBlob)
	if !ok {
		t.Fatalf("expected *StoredBlob, got %T", obj)
	}

	if !bytes.Equal(readBlob.Data, testData) {
		t.Error("roundtrip data mismatch")
	}
}
