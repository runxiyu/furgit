package packrev

import (
	"encoding/binary"
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/lgo/intconv"
)

// ErrMalformedReverseIndex reports that
// a pack reverse index is truncated,
// has a bad signature, version, or hash function,
// or contains invalid index positions.
var ErrMalformedReverseIndex = errors.New("internal/format/packrev: malformed pack reverse index")

const (
	signature = 0x52494458 // "RIDX"
	version   = 1

	headerLen = 12
)

// hashFunctionID returns the on-disk hash function identifier
// for one object format.
func hashFunctionID(objectFormat id.ObjectFormat) (uint32, error) {
	switch objectFormat {
	case id.ObjectFormatSHA1:
		return 1, nil
	case id.ObjectFormatSHA256:
		return 2, nil
	case id.ObjectFormatUnknown:
	}

	return 0, id.ErrInvalidObjectFormat
}

// Packrev is a parsed pack reverse index view over borrowed bytes.
//
// Labels: Deps-Borrowed, Life-Parent, MT-Safe.
type Packrev struct {
	// data is the entire pack reverse index payload.
	data []byte
	// hashSize is the object ID size of the object format.
	hashSize int
	// numObjects is the number of index position entries.
	numObjects int
}

// Parse parses a pack reverse index from data.
func Parse(data []byte, objectFormat id.ObjectFormat) (Packrev, error) {
	var zero Packrev

	wantHashID, err := hashFunctionID(objectFormat)
	if err != nil {
		return zero, err
	}

	hashSize := objectFormat.Size()

	if len(data) < headerLen+2*hashSize {
		return zero, fmt.Errorf("%w: truncated", ErrMalformedReverseIndex)
	}

	if binary.BigEndian.Uint32(data) != signature {
		return zero, fmt.Errorf("%w: bad signature", ErrMalformedReverseIndex)
	}

	if binary.BigEndian.Uint32(data[4:]) != version {
		return zero, fmt.Errorf("%w: unsupported version", ErrMalformedReverseIndex)
	}

	if binary.BigEndian.Uint32(data[8:]) != wantHashID {
		return zero, fmt.Errorf("%w: hash function mismatch", ErrMalformedReverseIndex)
	}

	positionBytes := len(data) - headerLen - 2*hashSize
	if positionBytes%4 != 0 {
		return zero, fmt.Errorf("%w: position table size not a 32-bit multiple", ErrMalformedReverseIndex)
	}

	return Packrev{
		data:       data,
		hashSize:   hashSize,
		numObjects: positionBytes / 4,
	}, nil
}

// NumObjects returns the number of index position entries.
func (rev *Packrev) NumObjects() int {
	return rev.numObjects
}

// PackHash returns the pack hash recorded in the trailer.
//
// Labels: Life-Parent, Mut-No.
func (rev *Packrev) PackHash() []byte {
	return rev.data[len(rev.data)-2*rev.hashSize : len(rev.data)-rev.hashSize]
}

// PositionAt returns the pack index position
// of the object at a pack offset order position.
//
// PositionAt panics when packOrder is out of range,
// and errors when the stored position is not a valid index position.
func (rev *Packrev) PositionAt(packOrder int) (int, error) {
	if packOrder < 0 || packOrder >= rev.numObjects {
		panic("internal/format/packrev: pack order position out of range")
	}

	stored := binary.BigEndian.Uint32(rev.data[headerLen+4*packOrder:])

	position, err := intconv.Uint32ToInt(stored)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrMalformedReverseIndex, err)
	}

	if position >= rev.numObjects {
		return 0, fmt.Errorf("%w: index position out of range", ErrMalformedReverseIndex)
	}

	return position, nil
}
