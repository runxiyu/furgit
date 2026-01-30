package furgit

import (
	"encoding/binary"

	"codeberg.org/lindenii/furgit/internal/bloom"
)

const (
	commitGraphParentNone      = 0x70000000
	commitGraphParentExtraMask = 0x80000000
	commitGraphParentLastMask  = 0x7fffffff

	correctedGenerationOverflow = 0x80000000
)

type commitGraphCommit struct {
	Position    uint32
	Tree        Hash
	Parents     []uint32
	CommitTime  uint64
	Generation  uint64
	Graph       *commitGraph
	graphLexPos uint32
}

func commitGraphAppendParent(parents []uint32, g *commitGraph, edge uint32) ([]uint32, error) {
	if edge == commitGraphParentNone {
		return parents, nil
	}
	if (edge & commitGraphParentExtraMask) == 0 {
		return append(parents, edge), nil
	}
	if g.chunkExtraEdges == nil {
		return nil, ErrInvalidObject
	}
	idx := edge & commitGraphParentLastMask
	for {
		off := int(idx) * 4
		if off+4 > len(g.chunkExtraEdges) {
			return nil, ErrInvalidObject
		}
		val := binary.BigEndian.Uint32(g.chunkExtraEdges[off : off+4])
		parents = append(parents, val&commitGraphParentLastMask)
		idx++
		if (val & commitGraphParentExtraMask) != 0 {
			break
		}
	}
	return parents, nil
}

func (cg *commitGraph) CommitPosition(id Hash) (uint32, bool) {
	return cg.lookupOID(id)
}

func (cg *commitGraph) CommitAt(pos uint32) (*commitGraphCommit, error) {
	return cg.commitDataAt(pos)
}

func (cg *commitGraph) OIDAt(pos uint32) (Hash, error) {
	return cg.loadOID(pos)
}

func (c *commitGraphCommit) HasBloom() bool {
	return c != nil && c.Graph != nil && c.Graph.chunkBloomIndex != nil && c.Graph.chunkBloomData != nil
}

func (cg *commitGraph) BloomSettings() *bloom.Settings {
	if cg == nil {
		return nil
	}
	return cg.bloomSettings
}

func (cg *commitGraph) BloomFilter(pos uint32) (*bloom.Filter, bool) {
	g := cg
	for g != nil && pos < g.numCommitsInBase {
		g = g.base
	}
	if g == nil || g.chunkBloomIndex == nil || g.chunkBloomData == nil || g.bloomSettings == nil {
		return nil, false
	}
	if pos >= g.numCommits+g.numCommitsInBase {
		return nil, false
	}
	lex := pos - g.numCommitsInBase
	if int(lex)*4+4 > len(g.chunkBloomIndex) {
		return nil, false
	}
	end := binary.BigEndian.Uint32(g.chunkBloomIndex[int(lex)*4:])
	var start uint32
	if lex > 0 {
		start = binary.BigEndian.Uint32(g.chunkBloomIndex[int(lex-1)*4:])
	}
	if end < start {
		return nil, false
	}
	if int(end) > len(g.chunkBloomData)-bloom.DataHeaderSize ||
		int(start) > len(g.chunkBloomData)-bloom.DataHeaderSize {
		return nil, false
	}
	data := g.chunkBloomData[bloom.DataHeaderSize+start : bloom.DataHeaderSize+end]
	return &bloom.Filter{Data: data, Version: g.bloomSettings.HashVersion}, true
}

func (c *commitGraphCommit) BloomFilter() (*bloom.Filter, *bloom.Settings, bool) {
	if c == nil || c.Graph == nil {
		return nil, nil, false
	}
	filter, ok := c.Graph.BloomFilter(c.Position)
	if !ok {
		return nil, nil, false
	}
	return filter, c.Graph.bloomSettings, true
}

func (c *commitGraphCommit) ChangedPathMightContain(path []byte) bool {
	filter, settings, ok := c.BloomFilter()
	if !ok {
		return false
	}
	return filter.MightContain(path, settings)
}
