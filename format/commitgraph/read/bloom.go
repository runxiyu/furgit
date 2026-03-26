package read

import (
	"encoding/binary"

	"codeberg.org/lindenii/furgit/format/commitgraph/bloom"
	"codeberg.org/lindenii/furgit/internal/intconv"
)

// HasBloom reports whether any layer has changed-path Bloom data.
func (reader *Reader) HasBloom() bool {
	for i := range reader.layers {
		layer := &reader.layers[i]
		if layer.chunkBloomIndex != nil && layer.chunkBloomData != nil && layer.bloomSettings != nil {
			return true
		}
	}

	return false
}

// BloomVersion returns the changed-path Bloom hash version, or 0 if absent.
func (reader *Reader) BloomVersion() uint8 {
	for i := len(reader.layers) - 1; i >= 0; i-- {
		layer := &reader.layers[i]
		if layer.bloomSettings != nil {
			version, err := intconv.Uint32ToUint8(layer.bloomSettings.HashVersion)
			if err != nil {
				return 0
			}

			return version
		}
	}

	return 0
}

// BloomFilterAt returns one commit's changed-path Bloom filter.
//
// The returned filter borrows reader-owned mapped commit-graph data and is
// only valid until the reader is closed.
//
// Returns BloomUnavailableError when this commit graph has no Bloom data.
func (reader *Reader) BloomFilterAt(pos Position) (bloom.Filter, error) {
	layer, err := reader.layerByPosition(pos)
	if err != nil {
		return bloom.Filter{}, err
	}

	if layer.chunkBloomIndex == nil || layer.chunkBloomData == nil || layer.bloomSettings == nil {
		return bloom.Filter{}, &BloomUnavailableError{Pos: pos}
	}

	start, end, err := bloomRange(layer, pos.Index)
	if err != nil {
		return bloom.Filter{}, err
	}

	filter := bloom.NewFilter(
		layer.chunkBloomData[bloom.DataHeaderSize+start:bloom.DataHeaderSize+end],
		*layer.bloomSettings,
	)

	return filter, nil
}

func bloomRange(layer *layer, commitIndex uint32) (int, int, error) {
	off64 := uint64(commitIndex) * 4

	off, err := intconv.Uint64ToInt(off64)
	if err != nil {
		return 0, 0, err
	}

	end := binary.BigEndian.Uint32(layer.chunkBloomIndex[off : off+4])

	var start uint32

	if commitIndex > 0 {
		prevOff64 := uint64(commitIndex-1) * 4

		prevOff, err := intconv.Uint64ToInt(prevOff64)
		if err != nil {
			return 0, 0, err
		}

		start = binary.BigEndian.Uint32(layer.chunkBloomIndex[prevOff : prevOff+4])
	}

	if end < start {
		return 0, 0, &MalformedError{Path: layer.path, Reason: "invalid BIDX range"}
	}

	bdatLen := len(layer.chunkBloomData) - bloom.DataHeaderSize

	bdatLenU32, err := intconv.IntToUint32(bdatLen)
	if err != nil {
		return 0, 0, err
	}

	if end > bdatLenU32 {
		return 0, 0, &MalformedError{Path: layer.path, Reason: "BIDX range out of BDAT bounds"}
	}

	startInt, err := intconv.Uint64ToInt(uint64(start))
	if err != nil {
		return 0, 0, err
	}

	endInt, err := intconv.Uint64ToInt(uint64(end))
	if err != nil {
		return 0, 0, err
	}

	return startInt, endInt, nil
}
