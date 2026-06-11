package packidx

import (
	"encoding/binary"
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/lgo/intconv"
)

// ErrMalformedPackIndex reports that
// a pack index is truncated,
// has a bad signature or unsupported version,
// or has inconsistent tables.
var ErrMalformedPackIndex = errors.New("internal/format/packidx: malformed pack index")

const (
	signature = 0xff744f63
	version   = 2

	headerLen = 8
	fanoutLen = 256 * 4

	// largeOffsetFlag marks one 32-bit offset table entry
	// as an index into the 64-bit offset table.
	largeOffsetFlag = 0x80000000
)

// Packidx is one parsed pack index view over borrowed bytes.
//
// Labels: Deps-Borrowed, Life-Parent, MT-Safe.
type Packidx struct {
	// data is the entire pack index payload.
	data []byte
	// hashSize is the object ID size of the index's object format.
	hashSize int

	// numObjects is the object count from the last fanout entry.
	numObjects int

	// namesOff, crcOff, off32Off, and off64Off are
	// the byte offsets of the object ID, CRC32,
	// 32-bit offset, and 64-bit offset tables.
	namesOff int
	crcOff   int
	off32Off int
	off64Off int
	// off64Count is the number of 64-bit offset table entries.
	off64Count uint32
}

// Parse parses one pack index from data.
//
// hashSize must be the object ID size
// of the pack's object format;
// Parse panics on implausible hash sizes.
func Parse(data []byte, hashSize int) (Packidx, error) {
	var zero Packidx

	if hashSize <= 0 || hashSize > id.MaxObjectIDSize {
		panic("internal/format/packidx: invalid hash size")
	}

	if len(data) < headerLen+fanoutLen+2*hashSize {
		return zero, fmt.Errorf("%w: truncated", ErrMalformedPackIndex)
	}

	if binary.BigEndian.Uint32(data) != signature {
		return zero, fmt.Errorf("%w: bad signature", ErrMalformedPackIndex)
	}

	if binary.BigEndian.Uint32(data[4:]) != version {
		return zero, fmt.Errorf("%w: unsupported version", ErrMalformedPackIndex)
	}

	prev := uint32(0)
	for i := range 256 {
		count := binary.BigEndian.Uint32(data[headerLen+4*i:])
		if count < prev {
			return zero, fmt.Errorf("%w: non-monotonic fanout", ErrMalformedPackIndex)
		}

		prev = count
	}

	numObjects := uint64(prev)
	hashSize64 := uint64(hashSize)

	namesOff := uint64(headerLen + fanoutLen)
	crcOff := namesOff + numObjects*hashSize64
	off32Off := crcOff + 4*numObjects
	off64Off := off32Off + 4*numObjects

	minTotal := off64Off + 2*hashSize64

	dataLen, err := intconv.IntToUint64(len(data))
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrMalformedPackIndex, err)
	}

	if dataLen < minTotal {
		return zero, fmt.Errorf("%w: tables exceed index size", ErrMalformedPackIndex)
	}

	off64Bytes := dataLen - minTotal
	if off64Bytes%8 != 0 {
		return zero, fmt.Errorf("%w: trailing table size not a 64-bit offset multiple", ErrMalformedPackIndex)
	}

	off64Count := off64Bytes / 8
	if off64Count > numObjects {
		return zero, fmt.Errorf("%w: more 64-bit offsets than objects", ErrMalformedPackIndex)
	}

	idx := Packidx{
		data:       data,
		hashSize:   hashSize,
		off64Count: uint32(off64Count),
	}

	idx.numObjects, err = intconv.Uint64ToInt(numObjects)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrMalformedPackIndex, err)
	}

	idx.namesOff, err = intconv.Uint64ToInt(namesOff)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrMalformedPackIndex, err)
	}

	idx.crcOff, err = intconv.Uint64ToInt(crcOff)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrMalformedPackIndex, err)
	}

	idx.off32Off, err = intconv.Uint64ToInt(off32Off)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrMalformedPackIndex, err)
	}

	idx.off64Off, err = intconv.Uint64ToInt(off64Off)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrMalformedPackIndex, err)
	}

	return idx, nil
}

// NumObjects returns the number of objects in the index.
func (idx *Packidx) NumObjects() int {
	return idx.numObjects
}

// PackHash returns the pack hash recorded in the index trailer.
//
// Labels: Life-Parent, Mut-No.
func (idx *Packidx) PackHash() []byte {
	return idx.data[len(idx.data)-2*idx.hashSize : len(idx.data)-idx.hashSize]
}

// OIDAt returns the object ID bytes at one index position.
// Positions follow object ID sort order.
//
// OIDAt panics when pos is out of range.
//
// Labels: Life-Parent, Mut-No.
func (idx *Packidx) OIDAt(pos int) []byte {
	idx.checkPos(pos)

	start := idx.namesOff + pos*idx.hashSize

	return idx.data[start : start+idx.hashSize]
}

// CRCAt returns the CRC32 of the packed entry data
// at one index position.
//
// CRCAt panics when pos is out of range.
func (idx *Packidx) CRCAt(pos int) uint32 {
	idx.checkPos(pos)

	return binary.BigEndian.Uint32(idx.data[idx.crcOff+4*pos:])
}

// checkPos panics when pos is not a valid index position.
//
// An out-of-range position is a caller bug
// that slice bounds checking would not catch,
// since the tables share one data slice;
// an unchecked access would silently read other tables' bytes.
func (idx *Packidx) checkPos(pos int) {
	if pos < 0 || pos >= idx.numObjects {
		panic("internal/format/packidx: index position out of range")
	}
}
