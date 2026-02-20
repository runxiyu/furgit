package furgit

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackHeaderEncodeParseRoundtrip(t *testing.T) {
	cases := []struct {
		ty    ObjectType
		sizes []int
	}{
		{ObjectTypeCommit, []int{0, 1, 15, 16, 127, 128, 1024, 1 << 20}},
		{ObjectTypeTree, []int{0, 3, 31, 32, 255, 256, 4096}},
		{ObjectTypeBlob, []int{0, 7, 63, 64, 511, 512, 99999}},
		{ObjectTypeTag, []int{0, 2, 14, 15, 16, 127, 128}},
	}

	for _, c := range cases {
		for _, size := range c.sizes {
			encoded, err := packHeaderEncode(c.ty, size)
			if err != nil {
				t.Fatalf("packHeaderEncode(%v,%d) error: %v", c.ty, size, err)
			}
			gotTy, gotSize, consumed, err := packHeaderParse(encoded)
			if err != nil {
				t.Fatalf("packHeaderParse error: %v", err)
			}
			if gotTy != c.ty || gotSize != size {
				t.Fatalf("roundtrip mismatch: got (%v,%d), want (%v,%d)", gotTy, gotSize, c.ty, size)
			}
			if consumed != len(encoded) {
				t.Fatalf("consumed=%d, encoded=%d", consumed, len(encoded))
			}
		}
	}
}

func TestPackVarintEncodeRoundtrip(t *testing.T) {
	values := []int{0, 1, 2, 7, 8, 127, 128, 129, 255, 1024, 1 << 20}
	for _, v := range values {
		encoded, err := packVarintEncode(v)
		if err != nil {
			t.Fatalf("packVarintEncode(%d) error: %v", v, err)
		}
		pos := 0
		got, err := packVarintRead(encoded, &pos)
		if err != nil {
			t.Fatalf("packVarintRead error: %v", err)
		}
		if got != v {
			t.Fatalf("roundtrip mismatch: got %d, want %d", got, v)
		}
		if pos != len(encoded) {
			t.Fatalf("pos=%d, encoded=%d", pos, len(encoded))
		}
	}
}

func TestPackOfsEncodeRoundtrip(t *testing.T) {
	values := []uint64{1, 2, 7, 8, 9, 0x7f, 0x80, 0x81, 0x1000, 0x12345}
	for _, v := range values {
		encoded, err := packOfsEncode(v)
		if err != nil {
			t.Fatalf("packOfsEncode(%d) error: %v", v, err)
		}
		dist, consumed, err := packDeltaReadOfsDistance(encoded)
		if err != nil {
			t.Fatalf("packDeltaReadOfsDistance error: %v", err)
		}
		if dist != v {
			t.Fatalf("roundtrip mismatch: got %d, want %d", dist, v)
		}
		if consumed != len(encoded) {
			t.Fatalf("consumed=%d, encoded=%d", consumed, len(encoded))
		}
	}
}

func TestPackWriteNoDeltas(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	if err := os.WriteFile(filepath.Join(workDir, "file1.txt"), []byte("content1"), 0o644); err != nil {
		t.Fatalf("failed to write file1.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "file2.txt"), []byte("content2"), 0o644); err != nil {
		t.Fatalf("failed to write file2.txt: %v", err)
	}

	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "Test commit")
	commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")

	commitBody := gitCatFile(t, repoPath, "commit", commitHash)
	lines := bytes.Split(commitBody, []byte{'\n'})
	if len(lines) == 0 || !bytes.HasPrefix(lines[0], []byte("tree ")) {
		t.Fatalf("commit missing tree header")
	}
	treeHash := strings.TrimSpace(string(bytes.TrimPrefix(lines[0], []byte("tree "))))

	lsTree := gitCmd(t, repoPath, "ls-tree", "-r", treeHash)
	var blobHashes []string
	for _, line := range strings.Split(lsTree, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			t.Fatalf("unexpected ls-tree line: %q", line)
		}
		blobHashes = append(blobHashes, fields[2])
	}

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	var objects []Hash
	commitID, _ := repo.ParseHash(commitHash)
	objects = append(objects, commitID)
	treeID, _ := repo.ParseHash(treeHash)
	objects = append(objects, treeID)
	for _, bh := range blobHashes {
		id, _ := repo.ParseHash(bh)
		objects = append(objects, id)
	}
	expectedOids := append([]string{commitHash, treeHash}, blobHashes...)

	packDir := filepath.Join(repoPath, "objects", "pack")
	if err := os.MkdirAll(packDir, 0o755); err != nil {
		t.Fatalf("failed to create pack dir: %v", err)
	}
	pf, err := os.CreateTemp(packDir, "furgit-test-*.pack")
	if err != nil {
		t.Fatalf("failed to create pack file: %v", err)
	}
	packPath := pf.Name()
	idxPath := strings.TrimSuffix(packPath, ".pack") + ".idx"
	if _, err := repo.packWrite(pf, objects, packWriteOptions{}); err != nil {
		_ = pf.Close()
		t.Fatalf("packWrite failed: %v", err)
	}
	if err := pf.Close(); err != nil {
		t.Fatalf("failed to close pack file: %v", err)
	}

	defer func() {
		_ = os.Remove(packPath)
		_ = os.Remove(idxPath)
	}()

	_ = gitCmd(t, repoPath, "index-pack", "-o", idxPath, packPath)

	verifyOut := gitCmd(t, repoPath, "verify-pack", "-v", idxPath)
	seen := make(map[string]struct{})
	for _, line := range strings.Split(verifyOut, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "chain length") || strings.HasPrefix(line, "non delta") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		seen[parts[0]] = struct{}{}
	}
	for _, oid := range expectedOids {
		if _, ok := seen[oid]; !ok {
			t.Fatalf("verify-pack missing object %s", oid)
		}
	}

	for _, oid := range expectedOids {
		if err := removeLooseObject(repoPath, oid); err != nil {
			t.Fatalf("remove loose object %s: %v", oid, err)
		}
	}
	for _, oid := range expectedOids {
		_ = gitCmd(t, repoPath, "cat-file", "-p", oid)
	}

	_ = gitCmd(t, repoPath, "fsck", "--full", "--strict")
}

func TestPackWriteDeltasUnimplemented(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	buf := new(bytes.Buffer)
	_, err = repo.packWrite(buf, nil, packWriteOptions{EnableDeltas: true})
	if !errors.Is(err, errPackDeltaUnimplemented) {
		t.Fatalf("expected errPackDeltaUnimplemented, got %v", err)
	}
}

func removeLooseObject(repoPath, oid string) error {
	if len(oid) < 2 {
		return ErrInvalidObject
	}
	path := filepath.Join(repoPath, "objects", oid[:2], oid[2:])
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}
