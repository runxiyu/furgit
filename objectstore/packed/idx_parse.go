package packed

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"codeberg.org/lindenii/furgit/objectid"
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
	maxOffset64Count := index.numObjects - 1
	if maxOffset64Count < 0 {
		maxOffset64Count = 0
	}
	if index.offset64Count > maxOffset64Count {
		return fmt.Errorf("objectstore/packed: idx %q has oversized 64-bit offset table", index.idxName)
	}
	return nil
}

// lookup resolves one object ID to its pack offset within this index.
func (index *idxFile) lookup(id objectid.ObjectID) (uint64, bool, error) {
	if id.Algorithm() != index.algo {
		return 0, false, fmt.Errorf("objectstore/packed: object id algorithm mismatch")
	}
	idBytes := (&id).RawBytes()
	hashSize := len(idBytes)
	if hashSize != index.algo.Size() {
		return 0, false, fmt.Errorf("objectstore/packed: unexpected object id length")
	}

	first := int(idBytes[0])
	lo := 0
	if first > 0 {
		lo = int(index.fanout[first-1])
	}
	hi := int(index.fanout[first])
	if lo < 0 || hi < 0 || lo > hi || hi > index.numObjects {
		return 0, false, fmt.Errorf("objectstore/packed: idx %q has invalid fanout bounds", index.idxName)
	}

	for lo < hi {
		mid := lo + (hi-lo)/2
		nameOffset := index.namesOffset + mid*hashSize
		if nameOffset < 0 || nameOffset+hashSize > len(index.data) {
			return 0, false, fmt.Errorf("objectstore/packed: idx %q truncated name table", index.idxName)
		}
		cmp := bytes.Compare(index.data[nameOffset:nameOffset+hashSize], idBytes)
		if cmp == 0 {
			offset, err := index.offsetAt(mid)
			if err != nil {
				return 0, false, err
			}
			return offset, true, nil
		}
		if cmp < 0 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return 0, false, nil
}

// offsetAt resolves the pack offset for one object index entry.
func (index *idxFile) offsetAt(objectIndex int) (uint64, error) {
	if objectIndex < 0 || objectIndex >= index.numObjects {
		return 0, fmt.Errorf("objectstore/packed: idx %q offset index out of bounds", index.idxName)
	}
	wordOffset := index.offset32Offset + objectIndex*4
	if wordOffset < 0 || wordOffset+4 > len(index.data) {
		return 0, fmt.Errorf("objectstore/packed: idx %q truncated 32-bit offset table", index.idxName)
	}
	word := binary.BigEndian.Uint32(index.data[wordOffset : wordOffset+4])
	if word&0x80000000 == 0 {
		return uint64(word), nil
	}

	pos := int(word & 0x7fffffff)
	if pos < 0 || pos >= index.offset64Count {
		return 0, fmt.Errorf("objectstore/packed: idx %q invalid 64-bit offset position", index.idxName)
	}
	offOffset := index.offset64Offset + pos*8
	if offOffset < 0 || offOffset+8 > len(index.data)-2*index.algo.Size() {
		return 0, fmt.Errorf("objectstore/packed: idx %q truncated 64-bit offset table", index.idxName)
	}
	return binary.BigEndian.Uint64(index.data[offOffset : offOffset+8]), nil
}
