package furgit

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/bloom"
)

func TestCommitGraphReadBasic(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	files := []string{"a.txt", "b.txt", "c.txt"}
	for i, name := range files {
		path := filepath.Join(workDir, name)
		content := fmt.Sprintf("content-%d\n", i)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
		gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", fmt.Sprintf("commit %d", i))
	}

	gitCmd(t, repoPath, nil, "repack", "-a", "-d")
	gitCmd(t, repoPath, nil, "commit-graph", "write", "--reachable", "--changed-paths")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	cg, err := repo.CommitGraph()
	if err != nil {
		t.Fatalf("CommitGraph failed: %v", err)
	}

	revList := gitCmd(t, repoPath, nil, "rev-list", "--all")
	commits := strings.Fields(revList)
	if len(commits) == 0 {
		t.Fatal("no commits found")
	}

	falsePositives := 0
	for _, oidStr := range commits {
		id, err := repo.ParseHash(oidStr)
		if err != nil {
			t.Fatalf("ParseHash failed: %v", err)
		}
		pos, ok := cg.CommitPosition(id)
		if !ok {
			t.Fatalf("commit %s not found in commit-graph", oidStr)
		}
		cc, err := cg.CommitAt(pos)
		if err != nil {
			t.Fatalf("CommitAt failed: %v", err)
		}

		treeStr := string(gitCmd(t, repoPath, nil, "show", "-s", "--format=%T", oidStr))
		tree, err := repo.ParseHash(treeStr)
		if err != nil {
			t.Fatalf("ParseHash tree failed: %v", err)
		}
		if cc.Tree != tree {
			t.Fatalf("tree mismatch for %s: got %s want %s", oidStr, cc.Tree.String(), tree.String())
		}

		parentStr := string(gitCmd(t, repoPath, nil, "show", "-s", "--format=%P", oidStr))
		parents := strings.Fields(parentStr)
		if len(parents) != len(cc.Parents) {
			t.Fatalf("parent count mismatch for %s: got %d want %d", oidStr, len(cc.Parents), len(parents))
		}
		for i, p := range parents {
			pid, err := repo.ParseHash(p)
			if err != nil {
				t.Fatalf("ParseHash parent failed: %v", err)
			}
			got, err := cg.OIDAt(cc.Parents[i])
			if err != nil {
				t.Fatalf("OIDAt parent failed: %v", err)
			}
			if got != pid {
				t.Fatalf("parent mismatch for %s: got %s want %s", oidStr, got.String(), pid.String())
			}
		}

		ctStr := gitCmd(t, repoPath, nil, "show", "-s", "--format=%ct", oidStr)
		ct, err := strconv.ParseInt(ctStr, 10, 64)
		if err != nil {
			t.Fatalf("parse commit time: %v", err)
		}
		if cc.CommitTime != uint64(ct) {
			t.Fatalf("commit time mismatch for %s: got %d want %d", oidStr, cc.CommitTime, ct)
		}

		changed := gitCmd(t, repoPath, nil, "diff-tree", "--no-commit-id", "--name-only", "-r", "--root", oidStr)
		for _, path := range strings.Fields(changed) {
			if path == "" {
				continue
			}
			if !cc.ChangedPathMightContain([]byte(path)) {
				t.Fatalf("bloom filter missing changed path %q in %s", path, oidStr)
			}
		}
		if cc.ChangedPathMightContain([]byte("this-path-should-not-exist-abcdefg")) {
			falsePositives++
		}
	}
	if falsePositives > 1 {
		t.Fatalf("unexpected bloom false positive rate: %d", falsePositives)
	}
}

func TestCommitGraphReadExtraEdges(t *testing.T) {
	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	gitCmd(t, workDir, nil, "init")

	if err := os.WriteFile(filepath.Join(workDir, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatalf("write base: %v", err)
	}
	gitCmd(t, workDir, nil, "add", ".")
	gitCmd(t, workDir, nil, "commit", "-m", "base")

	gitCmd(t, workDir, nil, "checkout", "-b", "branch1")
	if err := os.WriteFile(filepath.Join(workDir, "b1.txt"), []byte("b1\n"), 0o644); err != nil {
		t.Fatalf("write b1: %v", err)
	}
	gitCmd(t, workDir, nil, "add", ".")
	gitCmd(t, workDir, nil, "commit", "-m", "b1")

	gitCmd(t, workDir, nil, "checkout", "master")
	gitCmd(t, workDir, nil, "checkout", "-b", "branch2")
	if err := os.WriteFile(filepath.Join(workDir, "b2.txt"), []byte("b2\n"), 0o644); err != nil {
		t.Fatalf("write b2: %v", err)
	}
	gitCmd(t, workDir, nil, "add", ".")
	gitCmd(t, workDir, nil, "commit", "-m", "b2")

	gitCmd(t, workDir, nil, "checkout", "master")
	gitCmd(t, workDir, nil, "checkout", "-b", "branch3")
	if err := os.WriteFile(filepath.Join(workDir, "b3.txt"), []byte("b3\n"), 0o644); err != nil {
		t.Fatalf("write b3: %v", err)
	}
	gitCmd(t, workDir, nil, "add", ".")
	gitCmd(t, workDir, nil, "commit", "-m", "b3")

	gitCmd(t, workDir, nil, "checkout", "master")
	gitCmd(t, workDir, nil, "merge", "--no-ff", "-m", "octopus", "branch1", "branch2", "branch3")

	gitCmd(t, workDir, nil, "repack", "-a", "-d")
	gitCmd(t, workDir, nil, "commit-graph", "write", "--reachable")

	repo, err := OpenRepository(filepath.Join(workDir, ".git"))
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	cg, err := repo.CommitGraph()
	if err != nil {
		t.Fatalf("CommitGraph failed: %v", err)
	}

	mergeHash := gitCmd(t, workDir, nil, "rev-parse", "HEAD")
	id, _ := repo.ParseHash(mergeHash)
	pos, ok := cg.CommitPosition(id)
	if !ok {
		t.Fatalf("merge commit not found")
	}
	cc, err := cg.CommitAt(pos)
	if err != nil {
		t.Fatalf("CommitAt failed: %v", err)
	}
	parentStr := gitCmd(t, workDir, nil, "show", "-s", "--format=%P", mergeHash)
	parents := strings.Fields(parentStr)
	if len(cc.Parents) != len(parents) {
		t.Fatalf("octopus parent mismatch: got %d want %d", len(cc.Parents), len(parents))
	}
}

func TestCommitGraphReadSplitChain(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	for i := 0; i < 5; i++ {
		path := filepath.Join(workDir, fmt.Sprintf("file%d.txt", i))
		if err := os.WriteFile(path, []byte(fmt.Sprintf("v%d\n", i)), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
		gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", fmt.Sprintf("commit %d", i))
	}

	gitCmd(t, repoPath, nil, "repack", "-a", "-d")
	gitCmd(t, repoPath, nil, "commit-graph", "write", "--reachable", "--changed-paths", "--split=no-merge")

	// more commits needed to get a split chain with base layer
	for i := 5; i < 8; i++ {
		path := filepath.Join(workDir, fmt.Sprintf("file%d.txt", i))
		if err := os.WriteFile(path, []byte(fmt.Sprintf("v%d\n", i)), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
		gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", fmt.Sprintf("commit %d", i))
	}
	gitCmd(t, repoPath, nil, "repack", "-a", "-d")
	gitCmd(t, repoPath, nil, "commit-graph", "write", "--reachable", "--changed-paths", "--split=no-merge", "--append")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	cg, err := repo.CommitGraph()
	if err != nil {
		t.Fatalf("CommitGraph failed: %v", err)
	}
	chainPath := filepath.Join(repoPath, "objects", "info", "commit-graphs", "commit-graph-chain")
	if _, err := os.Stat(chainPath); err != nil {
		t.Fatalf("commit-graph chain not created by git: %v", err)
	}
	if cg.base == nil {
		t.Fatalf("expected split commit-graph chain")
	}
	if cg.numCommitsInBase == 0 {
		t.Fatalf("expected non-zero base commit count")
	}
	revList := gitCmd(t, repoPath, nil, "rev-list", "--max-parents=0", "HEAD")
	id, _ := repo.ParseHash(revList)
	if _, ok := cg.CommitPosition(id); !ok {
		t.Fatalf("base commit not found in commit-graph chain")
	}
}

func TestCommitGraphTrailerChecksum(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()
	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	data := buildCommitGraphBytes(t, repo, 1, nil, nil, nil)
	data[len(data)-1] ^= 0xff
	if _, err := parseCommitGraph(repo, "corrupt.graph", data); err == nil {
		t.Fatalf("expected checksum error")
	}
}

func TestCommitGraphFanoutMonotonicity(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()
	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	oids := [][]byte{
		bytes.Repeat([]byte{0x10}, repo.hashAlgo.Size()),
		bytes.Repeat([]byte{0x20}, repo.hashAlgo.Size()),
	}
	data := buildCommitGraphBytes(t, repo, 2, oids, nil, nil)
	start, end, ok := chunkRange(t, repo, data, commitGraphChunkOIDF)
	if !ok {
		t.Fatalf("OIDF chunk not found")
	}
	fanout := make([]byte, end-start)
	copy(fanout, data[start:end])
	binary.BigEndian.PutUint32(fanout[0:4], 2)
	binary.BigEndian.PutUint32(fanout[4:8], 1)
	data = replaceChunk(t, repo, data, commitGraphChunkOIDF, fanout)
	if _, err := parseCommitGraph(repo, "bad-fanout.graph", data); err == nil {
		t.Fatalf("expected fanout monotonicity error")
	}
}

func TestCommitGraphBIDXMonotonicity(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()
	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	oids := [][]byte{
		bytes.Repeat([]byte{0x10}, repo.hashAlgo.Size()),
		bytes.Repeat([]byte{0x20}, repo.hashAlgo.Size()),
	}
	bidx := []uint32{8, 4}
	data := buildCommitGraphBytes(t, repo, 2, oids, bidx, []byte{0, 0, 0, 0, 0, 0, 0, 0})
	if _, err := parseCommitGraph(repo, "bad-bidx.graph", data); err == nil {
		t.Fatalf("expected bidx monotonicity error")
	}
}

func TestCommitGraphExtraEdgesBounds(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()
	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	oids := [][]byte{
		bytes.Repeat([]byte{0x10}, repo.hashAlgo.Size()),
		bytes.Repeat([]byte{0x20}, repo.hashAlgo.Size()),
		bytes.Repeat([]byte{0x30}, repo.hashAlgo.Size()),
	}
	extraEdges := []uint32{0x80000000 | 100}
	data := buildCommitGraphBytes(t, repo, 3, oids, nil, nil)

	// inject EDGE chunk with an out-of-range index
	data = injectChunk(t, repo, data, commitGraphChunkEDGE, u32SliceToBytes(extraEdges))

	// force the first commit to use the extra edges list
	start, end, ok := chunkRange(t, repo, data, commitGraphChunkCDAT)
	if !ok {
		t.Fatalf("CDAT chunk not found")
	}
	cdat := make([]byte, end-start)
	copy(cdat, data[start:end])
	hashSize := repo.hashAlgo.Size()
	binary.BigEndian.PutUint32(cdat[hashSize+4:hashSize+8], 0x80000000|100)
	data = replaceChunk(t, repo, data, commitGraphChunkCDAT, cdat)

	cg, err := parseCommitGraph(repo, "edge.graph", data)
	if err != nil {
		t.Fatalf("parseCommitGraph failed: %v", err)
	}
	pos, ok := cg.CommitPosition(hashFromBytes(repo, oids[0]))
	if !ok {
		t.Fatalf("commit not found")
	}
	if _, err := cg.CommitAt(pos); err == nil {
		t.Fatalf("expected extra-edges bounds error")
	}
}

func buildCommitGraphBytes(t *testing.T, repo *Repository, n int, oids [][]byte, bidx []uint32, bdat []byte) []byte {
	t.Helper()
	if n <= 0 {
		t.Fatalf("invalid num commits")
	}
	hashSize := repo.hashAlgo.Size()
	if len(oids) == 0 {
		for i := 0; i < n; i++ {
			b := bytes.Repeat([]byte{byte(0x10 + i)}, hashSize)
			oids = append(oids, b)
		}
	}
	sort.Slice(oids, func(i, j int) bool {
		return bytes.Compare(oids[i], oids[j]) < 0
	})

	var chunks []chunkSpec
	oidf := make([]byte, commitGraphFanoutSize)
	for _, oid := range oids {
		idx := int(oid[0])
		for i := idx; i < 256; i++ {
			v := binary.BigEndian.Uint32(oidf[i*4 : i*4+4])
			binary.BigEndian.PutUint32(oidf[i*4:i*4+4], v+1)
		}
	}
	chunks = append(chunks, chunkSpec{commitGraphChunkOIDF, oidf})

	oidl := make([]byte, 0, n*hashSize)
	for _, oid := range oids {
		oidl = append(oidl, oid...)
	}
	chunks = append(chunks, chunkSpec{commitGraphChunkOIDL, oidl})

	cdat := make([]byte, n*commitGraphCommitDataSize(repo.hashAlgo))
	for i := 0; i < n; i++ {
		base := i * commitGraphCommitDataSize(repo.hashAlgo)
		copy(cdat[base:base+hashSize], bytes.Repeat([]byte{0xaa}, hashSize))
		binary.BigEndian.PutUint32(cdat[base+hashSize:], commitGraphParentNone)
		binary.BigEndian.PutUint32(cdat[base+hashSize+4:], commitGraphParentNone)
		binary.BigEndian.PutUint32(cdat[base+hashSize+8:], 4)
		binary.BigEndian.PutUint32(cdat[base+hashSize+12:], 0)
	}
	chunks = append(chunks, chunkSpec{commitGraphChunkCDAT, cdat})

	if bidx != nil && bdat != nil {
		bidxBytes := u32SliceToBytes(bidx)
		bdatHeader := make([]byte, bloom.DataHeaderSize)
		binary.BigEndian.PutUint32(bdatHeader[0:4], 2)
		binary.BigEndian.PutUint32(bdatHeader[4:8], 7)
		binary.BigEndian.PutUint32(bdatHeader[8:12], 10)
		bdatFull := append(bdatHeader, bdat...)
		chunks = append(chunks, chunkSpec{commitGraphChunkBIDX, bidxBytes})
		chunks = append(chunks, chunkSpec{commitGraphChunkBDAT, bdatFull})
	}

	return assembleCommitGraphBytes(t, repo, chunks)
}

type chunkSpec struct {
	id   uint32
	data []byte
}

func assembleCommitGraphBytes(t *testing.T, repo *Repository, chunks []chunkSpec) []byte {
	t.Helper()
	var buf bytes.Buffer
	buf.Grow(1024)
	var header [commitGraphHeaderSize]byte
	binary.BigEndian.PutUint32(header[0:4], commitGraphSignature)
	header[4] = commitGraphVersion
	header[5] = hashAlgoVersion(repo.hashAlgo)
	header[6] = byte(len(chunks))
	header[7] = 0
	buf.Write(header[:])

	tocStart := buf.Len()
	tocLen := (len(chunks) + 1) * commitGraphChunkEntrySize
	buf.Grow(tocLen)
	buf.Write(make([]byte, tocLen))

	offset := tocStart + tocLen
	for i, ch := range chunks {
		pos := tocStart + i*commitGraphChunkEntrySize
		binary.BigEndian.PutUint32(buf.Bytes()[pos:pos+4], ch.id)
		binary.BigEndian.PutUint64(buf.Bytes()[pos+4:pos+12], uint64(offset))
		offset += len(ch.data)
	}
	pos := tocStart + len(chunks)*commitGraphChunkEntrySize
	binary.BigEndian.PutUint32(buf.Bytes()[pos:pos+4], 0)
	binary.BigEndian.PutUint64(buf.Bytes()[pos+4:pos+12], uint64(offset))

	for _, ch := range chunks {
		buf.Write(ch.data)
	}

	h, err := repo.hashAlgo.New()
	if err != nil {
		t.Fatalf("hash.New failed: %v", err)
	}
	_, _ = h.Write(buf.Bytes())
	buf.Write(h.Sum(nil))
	return buf.Bytes()
}

func chunkRange(t *testing.T, repo *Repository, data []byte, id uint32) (int, int, bool) {
	t.Helper()
	hashSize := repo.hashAlgo.Size()
	numChunks := int(data[6])
	tocStart := commitGraphHeaderSize
	trailerStart := len(data) - hashSize

	for i := 0; i < numChunks; i++ {
		pos := tocStart + i*commitGraphChunkEntrySize
		chID := binary.BigEndian.Uint32(data[pos : pos+4])
		if chID == 0 {
			break
		}
		start := int(binary.BigEndian.Uint64(data[pos+4 : pos+12]))
		end := trailerStart
		if i+1 < numChunks {
			end = int(binary.BigEndian.Uint64(data[pos+commitGraphChunkEntrySize+4 : pos+commitGraphChunkEntrySize+12]))
		}
		if chID == id {
			return start, end, true
		}
	}
	return 0, 0, false
}

func replaceChunk(t *testing.T, repo *Repository, data []byte, id uint32, payload []byte) []byte {
	t.Helper()
	hashSize := repo.hashAlgo.Size()
	numChunks := int(data[6])
	tocStart := commitGraphHeaderSize
	trailerStart := len(data) - hashSize

	chunks := make([]chunkSpec, 0, numChunks)
	for i := 0; i < numChunks; i++ {
		pos := tocStart + i*commitGraphChunkEntrySize
		chID := binary.BigEndian.Uint32(data[pos : pos+4])
		if chID == 0 {
			break
		}
		start := int(binary.BigEndian.Uint64(data[pos+4 : pos+12]))
		end := trailerStart
		if i+1 < numChunks {
			end = int(binary.BigEndian.Uint64(data[pos+commitGraphChunkEntrySize+4 : pos+commitGraphChunkEntrySize+12]))
		}
		if chID == id {
			chunks = append(chunks, chunkSpec{chID, payload})
		} else {
			chunks = append(chunks, chunkSpec{chID, data[start:end]})
		}
	}
	if len(chunks) == 0 {
		t.Fatalf("no chunks found")
	}
	return assembleCommitGraphBytes(t, repo, chunks)
}

func injectChunk(t *testing.T, repo *Repository, data []byte, id uint32, payload []byte) []byte {
	t.Helper()
	hashSize := repo.hashAlgo.Size()
	numChunks := int(data[6])
	tocStart := commitGraphHeaderSize
	trailerStart := len(data) - hashSize

	chunks := make([]chunkSpec, 0, numChunks+1)
	for i := 0; i < numChunks; i++ {
		pos := tocStart + i*commitGraphChunkEntrySize
		chID := binary.BigEndian.Uint32(data[pos : pos+4])
		if chID == 0 {
			break
		}
		start := int(binary.BigEndian.Uint64(data[pos+4 : pos+12]))
		end := trailerStart
		if i+1 < numChunks {
			end = int(binary.BigEndian.Uint64(data[pos+commitGraphChunkEntrySize+4 : pos+commitGraphChunkEntrySize+12]))
		}
		chunks = append(chunks, chunkSpec{chID, data[start:end]})
	}
	chunks = append(chunks, chunkSpec{id, payload})
	return assembleCommitGraphBytes(t, repo, chunks)
}

func u32SliceToBytes(vals []uint32) []byte {
	out := make([]byte, len(vals)*4)
	for i, v := range vals {
		binary.BigEndian.PutUint32(out[i*4:i*4+4], v)
	}
	return out
}

func hashFromBytes(repo *Repository, b []byte) Hash {
	var h Hash
	copy(h.data[:], b)
	h.algo = repo.hashAlgo
	return h
}
