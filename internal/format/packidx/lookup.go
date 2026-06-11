package packidx

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Lookup searches the index for one object ID
// and returns its pack file offset.
//
// oid must be exactly the index's hash size;
// Lookup panics otherwise.
func (idx *Packidx) Lookup(oid []byte) (offset uint64, found bool, err error) {
	if len(oid) != idx.hashSize {
		panic("internal/format/packidx: invalid object ID length")
	}

	lo, hi := idx.fanoutRange(oid[0])

	for lo < hi {
		mid := lo + (hi-lo)/2

		cmp := bytes.Compare(oid, idx.OIDAt(mid))

		switch {
		case cmp == 0:
			offset, err = idx.OffsetAt(mid)
			if err != nil {
				return 0, false, err
			}

			return offset, true, nil
		case cmp < 0:
			hi = mid
		default:
			lo = mid + 1
		}
	}

	return 0, false, nil
}

// OffsetAt returns the pack file offset at one index position.
//
// OffsetAt panics when pos is out of range.
func (idx *Packidx) OffsetAt(pos int) (uint64, error) {
	idx.checkPos(pos)

	raw := binary.BigEndian.Uint32(idx.data[idx.off32Off+4*pos:])
	if raw&largeOffsetFlag == 0 {
		return uint64(raw), nil
	}

	slot := raw &^ largeOffsetFlag
	if uint64(slot) >= idx.off64Count {
		return 0, fmt.Errorf("%w: 64-bit offset reference out of range", ErrMalformedPackIndex)
	}

	return binary.BigEndian.Uint64(idx.data[idx.off64Off+8*int(slot):]), nil
}

// fanoutRange returns the index position range [lo, hi)
// of object IDs beginning with first.
func (idx *Packidx) fanoutRange(first byte) (lo, hi int) {
	hi = int(binary.BigEndian.Uint32(idx.data[headerLen+4*int(first):]))

	if first > 0 {
		lo = int(binary.BigEndian.Uint32(idx.data[headerLen+4*(int(first)-1):]))
	}

	return lo, hi
}
