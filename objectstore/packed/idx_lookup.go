package packed

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"codeberg.org/lindenii/furgit/objectid"
)

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
