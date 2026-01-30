package furgit

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"codeberg.org/lindenii/furgit/internal/bloom"
)

const (
	commitGraphSignature = 0x43475048 // "CGPH"
	commitGraphVersion   = 1
)

const (
	commitGraphChunkOIDF = 0x4f494446 // "OIDF"
	commitGraphChunkOIDL = 0x4f49444c // "OIDL"
	commitGraphChunkCDAT = 0x43444154 // "CDAT"
	commitGraphChunkGDA2 = 0x47444132 // "GDA2"
	commitGraphChunkGDO2 = 0x47444f32 // "GDO2"
	commitGraphChunkEDGE = 0x45444745 // "EDGE"
	commitGraphChunkBIDX = 0x42494458 // "BIDX"
	commitGraphChunkBDAT = 0x42444154 // "BDAT"
	commitGraphChunkBASE = 0x42415345 // "BASE"
)

const (
	commitGraphHeaderSize     = 8
	commitGraphChunkEntrySize = 12
	commitGraphFanoutSize     = 256 * 4
)

type commitGraph struct {
	repo *Repository

	filename string
	data     []byte
	dataLen  int

	numChunks uint8
	numBase   uint8
	hashAlgo  hashAlgorithm

	numCommits       uint32
	numCommitsInBase uint32

	base *commitGraph

	chunkOIDFanout []byte
	chunkOIDLookup []byte
	chunkCommit    []byte

	chunkGeneration   []byte
	chunkGenerationOv []byte
	readGeneration    bool

	chunkExtraEdges []byte

	chunkBloomIndex []byte
	chunkBloomData  []byte

	chunkBaseGraphs []byte

	bloomSettings *bloom.Settings

	closeOnce sync.Once
}

func (cg *commitGraph) Close() error {
	if cg == nil {
		return nil
	}
	var closeErr error
	cg.closeOnce.Do(func() {
		if len(cg.data) > 0 {
			if err := syscall.Munmap(cg.data); closeErr == nil {
				closeErr = err
			}
			cg.data = nil
		}
	})
	if cg.base != nil {
		_ = cg.base.Close()
	}
	return closeErr
}

func (repo *Repository) CommitGraph() (*commitGraph, error) {
	if repo == nil {
		return nil, ErrInvalidObject
	}
	repo.commitGraphOnce.Do(func() {
		repo.commitGraph, repo.commitGraphErr = repo.loadCommitGraph()
	})
	return repo.commitGraph, repo.commitGraphErr
}

func (repo *Repository) loadCommitGraph() (*commitGraph, error) {
	cg, err := repo.loadCommitGraphSingle()
	if err == nil {
		return cg, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	return repo.loadCommitGraphChain()
}

func (repo *Repository) loadCommitGraphSingle() (*commitGraph, error) {
	path := repo.repoPath(filepath.Join("objects", "info", "commit-graph"))
	return repo.loadCommitGraphFile(path)
}

func (repo *Repository) loadCommitGraphChain() (*commitGraph, error) {
	chainPath := repo.repoPath(filepath.Join("objects", "info", "commit-graphs", "commit-graph-chain"))
	f, err := os.Open(chainPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var hashes []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		hashes = append(hashes, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(hashes) == 0 {
		return nil, ErrNotFound
	}

	var chain *commitGraph
	for i, h := range hashes {
		graphPath := repo.repoPath(filepath.Join("objects", "info", "commit-graphs", fmt.Sprintf("graph-%s.graph", h)))
		g, err := repo.loadCommitGraphFile(graphPath)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			if g.chunkBaseGraphs == nil {
				return nil, ErrInvalidObject
			}
			if len(g.chunkBaseGraphs) < i*repo.hashAlgo.Size() {
				return nil, ErrInvalidObject
			}
			for j := 0; j < i; j++ {
				want := hashes[j]
				start := j * repo.hashAlgo.Size()
				end := start + repo.hashAlgo.Size()
				got := hexEncode(g.chunkBaseGraphs[start:end])
				if got != want {
					return nil, ErrInvalidObject
				}
			}
		}
		if chain != nil {
			if chain.numCommits+chain.numCommitsInBase > ^uint32(0) {
				return nil, ErrInvalidObject
			}
			g.numCommitsInBase = chain.numCommits + chain.numCommitsInBase
		}
		g.base = chain
		chain = g
	}
	validateCommitGraphChain(chain)
	return chain, nil
}

func validateCommitGraphChain(head *commitGraph) {
	if head == nil {
		return
	}
	readGeneration := true
	for g := head; g != nil; g = g.base {
		readGeneration = readGeneration && g.readGeneration
	}
	if !readGeneration {
		for g := head; g != nil; g = g.base {
			g.readGeneration = false
		}
	}

	var settings *bloom.Settings
	for g := head; g != nil; g = g.base {
		if g.bloomSettings == nil {
			continue
		}
		if settings == nil {
			settings = g.bloomSettings
			continue
		}
		if settings.HashVersion != g.bloomSettings.HashVersion ||
			settings.NumHashes != g.bloomSettings.NumHashes ||
			settings.BitsPerEntry != g.bloomSettings.BitsPerEntry {
			g.chunkBloomIndex = nil
			g.chunkBloomData = nil
			g.bloomSettings = nil
		}
	}
}

func (repo *Repository) loadCommitGraphFile(path string) (*commitGraph, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if stat.Size() < int64(commitGraphHeaderSize+commitGraphFanoutSize) {
		_ = f.Close()
		return nil, ErrInvalidObject
	}
	region, err := syscall.Mmap(
		int(f.Fd()),
		0,
		int(stat.Size()),
		syscall.PROT_READ,
		syscall.MAP_PRIVATE,
	)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	err = f.Close()
	if err != nil {
		_ = syscall.Munmap(region)
		return nil, err
	}
	cg, err := parseCommitGraph(repo, path, region)
	if err != nil {
		_ = syscall.Munmap(region)
		return nil, err
	}
	cg.data = region
	return cg, nil
}

func parseCommitGraph(repo *Repository, filename string, data []byte) (*commitGraph, error) {
	if len(data) < commitGraphHeaderSize {
		return nil, ErrInvalidObject
	}
	sig := binary.BigEndian.Uint32(data[0:4])
	if sig != commitGraphSignature {
		return nil, ErrInvalidObject
	}
	version := data[4]
	if version != commitGraphVersion {
		return nil, ErrInvalidObject
	}
	hashVersion := data[5]
	if hashVersion != hashAlgoVersion(repo.hashAlgo) {
		return nil, ErrInvalidObject
	}
	numChunks := data[6]
	numBase := data[7]

	tocStart := commitGraphHeaderSize
	tocLen := (int(numChunks) + 1) * commitGraphChunkEntrySize
	if len(data) < tocStart+tocLen+repo.hashAlgo.Size() {
		return nil, ErrInvalidObject
	}

	trailerStart := len(data) - repo.hashAlgo.Size()

	type chunkEntry struct {
		id     uint32
		offset uint64
	}
	entries := make([]chunkEntry, 0, numChunks+1)
	for i := 0; i < int(numChunks)+1; i++ {
		pos := tocStart + i*commitGraphChunkEntrySize
		id := binary.BigEndian.Uint32(data[pos : pos+4])
		offset := binary.BigEndian.Uint64(data[pos+4 : pos+12])
		entries = append(entries, chunkEntry{id: id, offset: offset})
	}

	chunks := map[uint32][]byte{}
	for i := 0; i < len(entries)-1; i++ {
		id := entries[i].id
		if id == 0 {
			break
		}
		start := int(entries[i].offset)
		end := int(entries[i+1].offset)
		if start < tocStart+tocLen || end < start || end > trailerStart {
			return nil, ErrInvalidObject
		}
		chunks[id] = data[start:end]
	}

	cg := &commitGraph{
		repo:      repo,
		filename:  filename,
		dataLen:   len(data),
		hashAlgo:  repo.hashAlgo,
		numChunks: numChunks,
		numBase:   numBase,
	}

	oidf := chunks[commitGraphChunkOIDF]
	if len(oidf) != commitGraphFanoutSize {
		return nil, ErrInvalidObject
	}
	cg.chunkOIDFanout = oidf
	cg.numCommits = binary.BigEndian.Uint32(oidf[len(oidf)-4:])
	for i := 0; i < 255; i++ {
		if binary.BigEndian.Uint32(oidf[i*4:(i+1)*4]) >
			binary.BigEndian.Uint32(oidf[(i+1)*4:(i+2)*4]) {
			return nil, ErrInvalidObject
		}
	}

	oidl := chunks[commitGraphChunkOIDL]
	if len(oidl) != int(cg.numCommits)*repo.hashAlgo.Size() {
		return nil, ErrInvalidObject
	}
	cg.chunkOIDLookup = oidl

	cdat := chunks[commitGraphChunkCDAT]
	if len(cdat) != int(cg.numCommits)*commitGraphCommitDataSize(repo.hashAlgo) {
		return nil, ErrInvalidObject
	}
	cg.chunkCommit = cdat

	if gda2 := chunks[commitGraphChunkGDA2]; len(gda2) != 0 {
		if len(gda2) != int(cg.numCommits)*4 {
			return nil, ErrInvalidObject
		}
		cg.chunkGeneration = gda2
		cg.readGeneration = true
	}
	if gdo2 := chunks[commitGraphChunkGDO2]; len(gdo2) != 0 {
		cg.chunkGenerationOv = gdo2
	}

	if edge := chunks[commitGraphChunkEDGE]; len(edge) != 0 {
		if len(edge)%4 != 0 {
			return nil, ErrInvalidObject
		}
		cg.chunkExtraEdges = edge
	}

	if base := chunks[commitGraphChunkBASE]; len(base) != 0 {
		if numBase == 0 {
			return nil, ErrInvalidObject
		}
		if len(base) != int(numBase)*repo.hashAlgo.Size() {
			return nil, ErrInvalidObject
		}
		cg.chunkBaseGraphs = base
	} else if numBase != 0 {
		return nil, ErrInvalidObject
	}

	if bidx := chunks[commitGraphChunkBIDX]; len(bidx) != 0 {
		if len(bidx) != int(cg.numCommits)*4 {
			return nil, ErrInvalidObject
		}
		cg.chunkBloomIndex = bidx
	}
	if bdat := chunks[commitGraphChunkBDAT]; len(bdat) != 0 {
		if len(bdat) < bloom.DataHeaderSize {
			return nil, ErrInvalidObject
		}
		cg.chunkBloomData = bdat
		settings, err := bloom.ParseSettings(bdat)
		if err != nil {
			return nil, err
		}
		cg.bloomSettings = settings
	}

	if (cg.chunkBloomIndex != nil) != (cg.chunkBloomData != nil) {
		cg.chunkBloomIndex = nil
		cg.chunkBloomData = nil
		cg.bloomSettings = nil
	}
	if cg.chunkBloomIndex != nil && cg.chunkBloomData != nil {
		for i := 0; i < int(cg.numCommits); i++ {
			cur := binary.BigEndian.Uint32(cg.chunkBloomIndex[i*4 : i*4+4])
			if i > 0 {
				prev := binary.BigEndian.Uint32(cg.chunkBloomIndex[(i-1)*4 : i*4])
				if cur < prev {
					return nil, ErrInvalidObject
				}
			}
			if int(cur) > len(cg.chunkBloomData)-bloom.DataHeaderSize {
				return nil, ErrInvalidObject
			}
		}
	}

	if err := verifyCommitGraphTrailer(repo, data); err != nil {
		return nil, err
	}

	return cg, nil
}

func commitGraphCommitDataSize(algo hashAlgorithm) int {
	return algo.Size() + 16
}

func hashAlgoVersion(algo hashAlgorithm) byte {
	switch algo {
	case hashAlgoSHA1:
		return 1
	case hashAlgoSHA256:
		return 2
	default:
		return 0
	}
}

func verifyCommitGraphTrailer(repo *Repository, data []byte) error {
	if repo == nil {
		return ErrInvalidObject
	}
	hashSize := repo.hashAlgo.Size()
	if len(data) < hashSize {
		return ErrInvalidObject
	}
	want := data[len(data)-hashSize:]
	h, err := repo.hashAlgo.New()
	if err != nil {
		return err
	}
	_, _ = h.Write(data[:len(data)-hashSize])
	got := h.Sum(nil)
	if !bytes.Equal(got, want) {
		return ErrInvalidObject
	}
	return nil
}

func hexEncode(b []byte) string {
	const hex = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hex[v>>4]
		out[i*2+1] = hex[v&0x0f]
	}
	return string(out)
}

func (cg *commitGraph) lookupOID(id Hash) (uint32, bool) {
	if cg == nil {
		return 0, false
	}
	return commitGraphBsearchOID(cg, id)
}

func commitGraphBsearchOID(cg *commitGraph, id Hash) (uint32, bool) {
	if cg == nil || len(cg.chunkOIDLookup) == 0 {
		return 0, false
	}
	hashSize := cg.hashAlgo.Size()
	first := int(id.data[0])
	fanout := cg.chunkOIDFanout
	var start uint32
	if first > 0 {
		start = binary.BigEndian.Uint32(fanout[(first-1)*4 : first*4])
	}
	end := binary.BigEndian.Uint32(fanout[first*4 : (first+1)*4])
	lo := int(start)
	hi := int(end) - 1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		pos := mid * hashSize
		cmp := bytes.Compare(cg.chunkOIDLookup[pos:pos+hashSize], id.data[:hashSize])
		if cmp == 0 {
			return uint32(mid) + cg.numCommitsInBase, true
		}
		if cmp < 0 {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	if cg.base != nil {
		return commitGraphBsearchOID(cg.base, id)
	}
	return 0, false
}

func (cg *commitGraph) loadOID(pos uint32) (Hash, error) {
	g := cg
	for g != nil && pos < g.numCommitsInBase {
		g = g.base
	}
	if g == nil {
		return Hash{}, ErrInvalidObject
	}
	if pos >= g.numCommits+g.numCommitsInBase {
		return Hash{}, ErrInvalidObject
	}
	lex := pos - g.numCommitsInBase
	hashSize := g.hashAlgo.Size()
	start := int(lex) * hashSize
	end := start + hashSize
	var out Hash
	copy(out.data[:], g.chunkOIDLookup[start:end])
	out.algo = g.hashAlgo
	return out, nil
}

func (cg *commitGraph) commitDataAt(pos uint32) (*commitGraphCommit, error) {
	g := cg
	for g != nil && pos < g.numCommitsInBase {
		g = g.base
	}
	if g == nil {
		return nil, ErrInvalidObject
	}
	if pos >= g.numCommits+g.numCommitsInBase {
		return nil, ErrInvalidObject
	}
	lex := pos - g.numCommitsInBase
	stride := commitGraphCommitDataSize(g.hashAlgo)
	start := int(lex) * stride
	end := start + stride
	if end > len(g.chunkCommit) {
		return nil, ErrInvalidObject
	}
	buf := g.chunkCommit[start:end]

	var tree Hash
	hashSize := g.hashAlgo.Size()
	copy(tree.data[:], buf[:hashSize])
	tree.algo = g.hashAlgo

	p1 := binary.BigEndian.Uint32(buf[hashSize : hashSize+4])
	p2 := binary.BigEndian.Uint32(buf[hashSize+4 : hashSize+8])

	genTime := binary.BigEndian.Uint32(buf[hashSize+8 : hashSize+12])
	timeLow := binary.BigEndian.Uint32(buf[hashSize+12 : hashSize+16])

	timeHigh := genTime & 0x3
	commitTime := (uint64(timeHigh) << 32) | uint64(timeLow)
	generation := uint64(genTime >> 2)

	if g.readGeneration {
		genOffset := binary.BigEndian.Uint32(g.chunkGeneration[int(lex)*4:])
		if genOffset&correctedGenerationOverflow != 0 {
			off := genOffset ^ correctedGenerationOverflow
			idx := int(off) * 8
			if g.chunkGenerationOv == nil || idx+8 > len(g.chunkGenerationOv) {
				return nil, ErrInvalidObject
			}
			offset := binary.BigEndian.Uint64(g.chunkGenerationOv[idx : idx+8])
			generation = commitTime + offset
		} else {
			generation = commitTime + uint64(genOffset)
		}
	}

	var parents []uint32
	var err error
	parents, err = commitGraphAppendParent(parents, g, p1)
	if err != nil {
		return nil, err
	}
	parents, err = commitGraphAppendParent(parents, g, p2)
	if err != nil {
		return nil, err
	}

	return &commitGraphCommit{
		Position:    pos,
		Tree:        tree,
		Parents:     parents,
		CommitTime:  commitTime,
		Generation:  generation,
		Graph:       g,
		graphLexPos: lex,
	}, nil
}
