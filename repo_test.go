package furgit

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func writeLooseBlob(t *testing.T, root string, data []byte) Hash {
	header, err := headerForType(ObjBlob, data)
	if err != nil {
		t.Fatalf("headerForType: %v", err)
	}
	raw := append(append([]byte(nil), header...), data...)
	id := computeRawHash(raw)
	path := filepath.Join(root, loosePath(id))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for loose object: %v", err)
	}
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(raw); err != nil {
		t.Fatalf("compress: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zlib: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write loose object: %v", err)
	}
	return id
}

func TestOpenRepositoryAndLooseRead(t *testing.T) {
	root := t.TempDir()
	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	id := writeLooseBlob(t, root, []byte("loose blob payload"))
	obj, err := repo.looseRead(id)
	if err != nil {
		t.Fatalf("looseRead error: %v", err)
	}
	blob, ok := obj.(*Blob)
	if !ok {
		t.Fatalf("expected Blob, got %T", obj)
	}
	if string(blob.Data) != "loose blob payload" {
		t.Fatalf("blob data mismatch: %q", blob.Data)
	}
}

func TestResolveRefLooseAndPacked(t *testing.T) {
	root := t.TempDir()
	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	looseID := hashWithByte(0xa0)
	loosePath := filepath.Join(root, "refs", "heads")
	if err := os.MkdirAll(loosePath, 0o755); err != nil {
		t.Fatalf("mkdir refs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(loosePath, "master"), []byte(looseID.String()+"\n"), 0o644); err != nil {
		t.Fatalf("write ref: %v", err)
	}
	id, err := repo.ResolveRef("refs/heads/master")
	if err != nil || id != looseID {
		t.Fatalf("ResolveRef loose mismatch (id=%v err=%v)", id, err)
	}

	packedID := hashWithByte(0xb0)
	packed := fmt.Sprintf("%s refs/tags/v1\n", packedID.String())
	if err := os.WriteFile(filepath.Join(root, "packed-refs"), []byte(packed), 0o644); err != nil {
		t.Fatalf("write packed refs: %v", err)
	}
	id, err = repo.resolvePackedRef("refs/tags/v1")
	if err != nil || id != packedID {
		t.Fatalf("resolvePackedRef direct mismatch (id=%v err=%v)", id, err)
	}
	id, err = repo.ResolveRef("refs/tags/v1")
	if err != nil || id != packedID {
		t.Fatalf("ResolveRef packed mismatch (id=%v err=%v)", id, err)
	}

	if _, err := repo.ResolveRef("refs/heads/missing"); !errors.Is(err, ErrInvalidObject) {
		t.Fatalf("expected ErrInvalidObject for missing ref, got %v", err)
	}
}

func TestResolveHEAD(t *testing.T) {
	root := t.TempDir()
	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	headPath := filepath.Join(root, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/master\n"), 0o644); err != nil {
		t.Fatalf("write HEAD: %v", err)
	}
	ref, err := repo.ResolveHEAD()
	if err != nil || ref != "refs/heads/master" {
		t.Fatalf("ResolveHEAD mismatch (ref=%q err=%v)", ref, err)
	}
	if err := os.WriteFile(headPath, []byte("detached\n"), 0o644); err != nil {
		t.Fatalf("write HEAD detached: %v", err)
	}
	if _, err := repo.ResolveHEAD(); err == nil {
		t.Fatal("expected error for detached HEAD")
	}
}
