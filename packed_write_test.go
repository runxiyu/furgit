package furgit

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/bufpool"
	"codeberg.org/lindenii/furgit/internal/zlibx"
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

	const (
		fileCount = 1000
		fileSize  = 1024
	)
	buf := make([]byte, fileSize)
	for i := 0; i < fileCount; i++ {
		if _, err := rand.Read(buf); err != nil {
			t.Fatalf("rand.Read failed: %v", err)
		}
		name := filepath.Join(workDir, fmt.Sprintf("file%04d.bin", i))
		if err := os.WriteFile(name, buf, 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "Test commit")
	commitHash := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")

	commitBody := gitCatFile(t, repoPath, "commit", commitHash)
	lines := bytes.Split(commitBody, []byte{'\n'})
	if len(lines) == 0 || !bytes.HasPrefix(lines[0], []byte("tree ")) {
		t.Fatalf("commit missing tree header")
	}
	treeHash := strings.TrimSpace(string(bytes.TrimPrefix(lines[0], []byte("tree "))))

	lsTree := gitCmd(t, repoPath, nil, "ls-tree", "-r", treeHash)
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
	if _, err := repo.packWrite(pf, objects, packWriteOptions{}, nil); err != nil {
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

	if err := checkPackStream(packPath, repo.hashAlgo, len(objects)); err != nil {
		t.Fatalf("pack stream invalid: %v", err)
	}

	_ = gitCmd(t, repoPath, nil, "index-pack", "-o", idxPath, packPath)

	verifyOut := gitCmd(t, repoPath, nil, "verify-pack", "-v", idxPath)
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
		_ = gitCmd(t, repoPath, nil, "cat-file", "-p", oid)
	}

	_ = gitCmd(t, repoPath, nil, "fsck", "--full", "--strict")
}

func TestPackWriteDeltas(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	const (
		fileCount = 200
		fileSize  = 2048
	)
	base := bytes.Repeat([]byte("delta-base-"), fileSize/10)
	for i := 0; i < fileCount; i++ {
		buf := make([]byte, len(base))
		copy(buf, base)
		buf[i%len(buf)] ^= byte(i)
		name := filepath.Join(workDir, fmt.Sprintf("delta%04d.txt", i))
		if err := os.WriteFile(name, buf, 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "Delta commit")
	commitHash := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")

	commitBody := gitCatFile(t, repoPath, "commit", commitHash)
	lines := bytes.Split(commitBody, []byte{'\n'})
	if len(lines) == 0 || !bytes.HasPrefix(lines[0], []byte("tree ")) {
		t.Fatalf("commit missing tree header")
	}
	treeHash := strings.TrimSpace(string(bytes.TrimPrefix(lines[0], []byte("tree "))))

	lsTree := gitCmd(t, repoPath, nil, "ls-tree", "-r", treeHash)
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
	pf, err := os.CreateTemp(packDir, "furgit-delta-test-*.pack")
	if err != nil {
		t.Fatalf("failed to create pack file: %v", err)
	}
	packPath := pf.Name()
	idxPath := strings.TrimSuffix(packPath, ".pack") + ".idx"
	if _, err := repo.packWrite(pf, objects, packWriteOptions{
		EnableDeltas:    true,
		MinDeltaSavings: 1,
	}, nil); err != nil {
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

	if err := checkPackStream(packPath, repo.hashAlgo, len(objects)); err != nil {
		t.Fatalf("pack stream invalid: %v", err)
	}

	_ = gitCmd(t, repoPath, nil, "index-pack", "-o", idxPath, packPath)

	verifyOut := gitCmd(t, repoPath, nil, "verify-pack", "-v", idxPath)
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
		_ = gitCmd(t, repoPath, nil, "cat-file", "-p", oid)
	}

	_ = gitCmd(t, repoPath, nil, "fsck", "--full", "--strict")
}

func TestPackWriteThinPackReachable(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	base := bytes.Repeat([]byte("A"), 16384)
	if err := os.WriteFile(filepath.Join(workDir, "file.txt"), base, 0o644); err != nil {
		t.Fatalf("write base file: %v", err)
	}
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "base")
	haveHash := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")

	mod := append([]byte(nil), base...)
	mod[1024] = 'B'
	if err := os.WriteFile(filepath.Join(workDir, "file.txt"), mod, 0o644); err != nil {
		t.Fatalf("write mod file: %v", err)
	}
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "target")
	wantHash := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	wantID, _ := repo.ParseHash(wantHash)
	haveID, _ := repo.ParseHash(haveHash)

	query := ReachabilityQuery{
		Wants:       []Hash{wantID},
		Haves:       []Hash{haveID},
		Mode:        ReachabilityAllObjects,
		StopAtHaves: true,
	}
	var buf bytes.Buffer
	if _, err := repo.packWriteReachable(&buf, query, packWriteOptions{
		EnableDeltas:    true,
		EnableThinPack:  true,
		MinDeltaSavings: 1,
	}); err != nil {
		t.Fatalf("packWriteReachable failed: %v", err)
	}

	thinSeen, err := checkThinPackStream(buf.Bytes(), repo)
	if err != nil {
		t.Fatalf("thin pack stream invalid: %v", err)
	}
	if !thinSeen {
		t.Fatalf("expected thin pack with ref-delta base outside pack")
	}

	packDir := filepath.Join(repoPath, "objects", "pack")
	if err := os.MkdirAll(packDir, 0o755); err != nil {
		t.Fatalf("failed to create pack dir: %v", err)
	}
	packPath := filepath.Join(packDir, "furgit-thin-test.pack")
	idxPath := strings.TrimSuffix(packPath, ".pack") + ".idx"
	_ = os.Remove(packPath)
	_ = os.Remove(idxPath)

	_ = gitCmd(t, repoPath, buf.Bytes(), "index-pack", "--stdin", "--fix-thin", "-o", idxPath, packPath)

	_ = gitCmd(t, repoPath, nil, "cat-file", "-p", wantHash)
	_ = gitCmd(t, repoPath, nil, "fsck", "--full", "--strict")

	_ = os.Remove(packPath)
	_ = os.Remove(idxPath)
}

func checkPackStream(path string, algo hashAlgorithm, objectCount int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) < 12 {
		return ErrInvalidObject
	}
	if binary.BigEndian.Uint32(data[0:4]) != packMagic || binary.BigEndian.Uint32(data[4:8]) != packVersion2 {
		return ErrInvalidObject
	}
	pos := 12
	hashSize := algo.Size()
	type objEntry struct {
		offset uint64
		ty     ObjectType
		body   []byte
	}
	byOffset := make(map[uint64]objEntry, objectCount)
	byHash := make(map[string]objEntry, objectCount)
	for i := 0; i < objectCount; i++ {
		objOffset := uint64(pos)
		ty, size, consumed, err := packHeaderParse(data[pos:])
		if err != nil {
			return fmt.Errorf("obj %d header at %d: %v", i, pos, err)
		}
		pos += consumed
		baseOffset := uint64(0)
		baseTy := ObjectTypeInvalid
		var baseBody []byte
		var baseHash Hash
		switch ty {
		case ObjectTypeOfsDelta:
			dist, distConsumed, err := packDeltaReadOfsDistance(data[pos:])
			if err != nil {
				return fmt.Errorf("obj %d ofs at %d: %v", i, pos, err)
			}
			pos += distConsumed
			if dist == 0 || dist > objOffset {
				return fmt.Errorf("obj %d ofs at %d: invalid dist", i, pos)
			}
			baseOffset = objOffset - dist
			base, ok := byOffset[baseOffset]
			if !ok {
				return fmt.Errorf("obj %d ofs at %d: missing base", i, pos)
			}
			baseTy = base.ty
			baseBody = base.body
		case ObjectTypeRefDelta:
			if pos+hashSize > len(data) {
				return ErrInvalidObject
			}
			copy(baseHash.data[:], data[pos:pos+hashSize])
			baseHash.algo = algo
			baseEntry, ok := byHash[baseHash.String()]
			if !ok {
				return fmt.Errorf("obj %d ref base not found", i)
			}
			baseTy = baseEntry.ty
			baseBody = baseEntry.body
			pos += hashSize
		default:
		}

		payloadBuf, zconsumed, err := zlibx.DecompressSized(data[pos:], size)
		if err != nil {
			return fmt.Errorf("obj %d zlib at %d: %v", i, pos, err)
		}
		payload := append([]byte(nil), payloadBuf.Bytes()...)
		payloadBuf.Release()
		pos += zconsumed
		switch ty {
		case ObjectTypeOfsDelta:
			if baseBody == nil {
				return fmt.Errorf("obj %d missing base body", i)
			}
			pos := 0
			baseSize, err := packVarintRead(payload, &pos)
			if err != nil {
				return fmt.Errorf("obj %d delta base size: %v", i, err)
			}
			resultSize, err := packVarintRead(payload, &pos)
			if err != nil {
				return fmt.Errorf("obj %d delta result size: %v", i, err)
			}
			if baseSize != len(baseBody) {
				return fmt.Errorf("obj %d delta base size mismatch: got %d want %d", i, baseSize, len(baseBody))
			}
			out, err := packDeltaApply(bufpool.FromOwned(baseBody), bufpool.FromOwned(payload))
			if err != nil {
				return fmt.Errorf("obj %d delta apply: %v", i, err)
			}
			body := append([]byte(nil), out.Bytes()...)
			out.Release()
			if resultSize != len(body) {
				return fmt.Errorf("obj %d delta result size mismatch: got %d want %d", i, len(body), resultSize)
			}
			byOffset[objOffset] = objEntry{offset: objOffset, ty: baseTy, body: body}
		case ObjectTypeRefDelta:
			if baseBody == nil {
				return fmt.Errorf("obj %d missing ref base body", i)
			}
			pos := 0
			baseSize, err := packVarintRead(payload, &pos)
			if err != nil {
				return fmt.Errorf("obj %d ref delta base size: %v", i, err)
			}
			resultSize, err := packVarintRead(payload, &pos)
			if err != nil {
				return fmt.Errorf("obj %d ref delta result size: %v", i, err)
			}
			if baseSize != len(baseBody) {
				return fmt.Errorf("obj %d ref delta base size mismatch: got %d want %d", i, baseSize, len(baseBody))
			}
			out, err := packDeltaApply(bufpool.FromOwned(baseBody), bufpool.FromOwned(payload))
			if err != nil {
				return fmt.Errorf("obj %d ref delta apply: %v", i, err)
			}
			body := append([]byte(nil), out.Bytes()...)
			out.Release()
			if resultSize != len(body) {
				return fmt.Errorf("obj %d ref delta result size mismatch: got %d want %d", i, len(body), resultSize)
			}
			byOffset[objOffset] = objEntry{offset: objOffset, ty: baseTy, body: body}
		default:
			if size >= 0 && len(payload) != size {
				return fmt.Errorf("obj %d size mismatch: got %d want %d", i, len(payload), size)
			}
			body := append([]byte(nil), payload...)
			byOffset[objOffset] = objEntry{offset: objOffset, ty: ty, body: body}
		}

		entry := byOffset[objOffset]
		if entry.body != nil && entry.ty != ObjectTypeInvalid {
			hdr, err := headerForType(entry.ty, entry.body)
			if err != nil {
				return err
			}
			raw := append(hdr, entry.body...)
			hash := algo.Sum(raw)
			byHash[hash.String()] = entry
		}
	}
	return nil
}

func checkThinPackStream(data []byte, repo *Repository) (bool, error) {
	if repo == nil {
		return false, ErrInvalidObject
	}
	if len(data) < 12 {
		return false, ErrInvalidObject
	}
	if binary.BigEndian.Uint32(data[0:4]) != packMagic || binary.BigEndian.Uint32(data[4:8]) != packVersion2 {
		return false, ErrInvalidObject
	}
	count := int(binary.BigEndian.Uint32(data[8:12]))
	pos := 12
	hashSize := repo.hashAlgo.Size()
	type objEntry struct {
		offset uint64
		ty     ObjectType
		body   []byte
	}
	byOffset := make(map[uint64]objEntry, count)
	byHash := make(map[string]objEntry, count)
	thinSeen := false

	for i := 0; i < count; i++ {
		objOffset := uint64(pos)
		ty, size, consumed, err := packHeaderParse(data[pos:])
		if err != nil {
			return thinSeen, fmt.Errorf("obj %d header at %d: %v", i, pos, err)
		}
		pos += consumed
		baseTy := ObjectTypeInvalid
		var baseBody []byte
		switch ty {
		case ObjectTypeOfsDelta:
			dist, distConsumed, err := packDeltaReadOfsDistance(data[pos:])
			if err != nil {
				return thinSeen, fmt.Errorf("obj %d ofs at %d: %v", i, pos, err)
			}
			pos += distConsumed
			if dist == 0 || dist > objOffset {
				return thinSeen, fmt.Errorf("obj %d ofs at %d: invalid dist", i, pos)
			}
			baseOffset := objOffset - dist
			base, ok := byOffset[baseOffset]
			if !ok {
				return thinSeen, fmt.Errorf("obj %d ofs at %d: missing base", i, pos)
			}
			baseTy = base.ty
			baseBody = base.body
		case ObjectTypeRefDelta:
			if pos+hashSize > len(data) {
				return thinSeen, ErrInvalidObject
			}
			var baseHash Hash
			copy(baseHash.data[:], data[pos:pos+hashSize])
			baseHash.algo = repo.hashAlgo
			baseEntry, ok := byHash[baseHash.String()]
			if ok {
				baseTy = baseEntry.ty
				baseBody = baseEntry.body
			} else {
				thinSeen = true
				ty, body, err := repo.ReadObjectTypeRaw(baseHash)
				if err != nil {
					return thinSeen, err
				}
				baseTy = ty
				baseBody = body
			}
			pos += hashSize
		default:
		}

		payloadBuf, zconsumed, err := zlibx.DecompressSized(data[pos:], size)
		if err != nil {
			return thinSeen, fmt.Errorf("obj %d zlib at %d: %v", i, pos, err)
		}
		payload := append([]byte(nil), payloadBuf.Bytes()...)
		payloadBuf.Release()
		pos += zconsumed
		switch ty {
		case ObjectTypeOfsDelta, ObjectTypeRefDelta:
			if baseBody == nil {
				return thinSeen, fmt.Errorf("obj %d missing base body", i)
			}
			pos := 0
			baseSize, err := packVarintRead(payload, &pos)
			if err != nil {
				return thinSeen, fmt.Errorf("obj %d delta base size: %v", i, err)
			}
			resultSize, err := packVarintRead(payload, &pos)
			if err != nil {
				return thinSeen, fmt.Errorf("obj %d delta result size: %v", i, err)
			}
			if baseSize != len(baseBody) {
				return thinSeen, fmt.Errorf("obj %d delta base size mismatch: got %d want %d", i, baseSize, len(baseBody))
			}
			out, err := packDeltaApply(bufpool.FromOwned(baseBody), bufpool.FromOwned(payload))
			if err != nil {
				return thinSeen, fmt.Errorf("obj %d delta apply: %v", i, err)
			}
			body := append([]byte(nil), out.Bytes()...)
			out.Release()
			if resultSize != len(body) {
				return thinSeen, fmt.Errorf("obj %d delta result size mismatch: got %d want %d", i, len(body), resultSize)
			}
			byOffset[objOffset] = objEntry{offset: objOffset, ty: baseTy, body: body}
		default:
			if size >= 0 && len(payload) != size {
				return thinSeen, fmt.Errorf("obj %d size mismatch: got %d want %d", i, len(payload), size)
			}
			body := append([]byte(nil), payload...)
			byOffset[objOffset] = objEntry{offset: objOffset, ty: ty, body: body}
		}

		entry := byOffset[objOffset]
		if entry.body != nil && entry.ty != ObjectTypeInvalid {
			hdr, err := headerForType(entry.ty, entry.body)
			if err != nil {
				return thinSeen, err
			}
			raw := append(hdr, entry.body...)
			hash := repo.hashAlgo.Sum(raw)
			byHash[hash.String()] = entry
		}
	}
	return thinSeen, nil
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
