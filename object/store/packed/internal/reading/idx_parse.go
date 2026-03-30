package reading

import (
	"encoding/binary"
	"fmt"
)

const (
	idxMagicV2   = 0xff744f63
	idxVersionV2 = 2
)

// parse validates mapped idx v2 structure and stores table boundaries.
func (index *idxFile) parse() error {
	hashSize := index.algo.Size()
	if hashSize <= 0 {
		return fmt.Errorf("objectstore/packed: idx %q has invalid hash algorithm", index.idxName)
	}

	minLen := 8 + 256*4 + 2*hashSize
	if len(index.data) < minLen {
		return fmt.Errorf("objectstore/packed: idx %q too short", index.idxName)
	}

	if binary.BigEndian.Uint32(index.data[:4]) != idxMagicV2 {
		return fmt.Errorf("objectstore/packed: idx %q invalid magic", index.idxName)
	}

	if binary.BigEndian.Uint32(index.data[4:8]) != idxVersionV2 {
		return fmt.Errorf("objectstore/packed: idx %q unsupported version", index.idxName)
	}

	prev := uint32(0)

	for i := range 256 {
		base := 8 + i*4

		cur := binary.BigEndian.Uint32(index.data[base : base+4])
		if cur < prev {
			return fmt.Errorf("objectstore/packed: idx %q has non-monotonic fanout table", index.idxName)
		}

		index.fanout[i] = cur
		prev = cur
	}

	index.numObjects = int(index.fanout[255])
	if index.numObjects < 0 {
		return fmt.Errorf("objectstore/packed: idx %q has invalid object count", index.idxName)
	}

	namesBytes := index.numObjects * hashSize
	crcBytes := index.numObjects * 4
	offset32Bytes := index.numObjects * 4

	minSize := 8 + 256*4 + namesBytes + crcBytes + offset32Bytes + 2*hashSize
	if minSize < 0 || len(index.data) < minSize {
		return fmt.Errorf("objectstore/packed: idx %q has truncated tables", index.idxName)
	}

	index.namesOffset = 8 + 256*4
	index.offset32Offset = index.namesOffset + namesBytes + crcBytes
	index.offset64Offset = index.offset32Offset + offset32Bytes

	offset64Bytes := len(index.data) - index.offset64Offset - 2*hashSize
	if offset64Bytes < 0 || offset64Bytes%8 != 0 {
		return fmt.Errorf("objectstore/packed: idx %q has malformed 64-bit offset table", index.idxName)
	}

	index.offset64Count = offset64Bytes / 8

	maxOffset64Count := max(index.numObjects-1, 0)
	if index.offset64Count > maxOffset64Count {
		return fmt.Errorf("objectstore/packed: idx %q has oversized 64-bit offset table", index.idxName)
	}

	return nil
}
