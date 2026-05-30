package id

import (
	"encoding/hex"
	"fmt"
	"hash"
)

// HexLen returns the encoded hexadecimal length.
func (objectFormat ObjectFormat) HexLen() int {
	return objectFormat.Size() * 2
}

// Size returns the hash size in bytes.
func (objectFormat ObjectFormat) Size() int {
	return objectFormat.details().size
}

// New returns a new hash.Hash for this object format.
func (objectFormat ObjectFormat) New() (hash.Hash, error) {
	newFn := objectFormat.details().new
	if newFn == nil {
		return nil, ErrInvalidObjectFormat
	}

	return newFn(), nil
}

// FromBytes builds an object ID from raw bytes for this object format.
func (objectFormat ObjectFormat) FromBytes(b []byte) (ObjectID, error) {
	var id ObjectID
	if objectFormat.Size() == 0 {
		return id, ErrInvalidObjectFormat
	}

	if len(b) != objectFormat.Size() {
		return id, fmt.Errorf("%w: got %d bytes, expected %d", ErrInvalidObjectID, len(b), objectFormat.Size())
	}

	copy(id.data[:], b)
	id.objectFormat = objectFormat

	return id, nil
}

// FromString parses an object ID from its canonical hex representation.
func (objectFormat ObjectFormat) FromString(s string) (ObjectID, error) {
	var id ObjectID
	if objectFormat.Size() == 0 {
		return id, ErrInvalidObjectFormat
	}

	if len(s)&1 != 0 {
		return id, fmt.Errorf("%w: odd hex length %d", ErrInvalidObjectID, len(s))
	}

	if len(s) != objectFormat.HexLen() {
		return id, fmt.Errorf("%w: got %d chars, expected %d", ErrInvalidObjectID, len(s), objectFormat.HexLen())
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return id, fmt.Errorf("%w: decode: %w", ErrInvalidObjectID, err)
	}

	copy(id.data[:], decoded)
	id.objectFormat = objectFormat

	return id, nil
}

// String returns the canonical object format name.
func (objectFormat ObjectFormat) String() string {
	return objectFormat.details().name
}

// Sum computes an object ID from raw data using the selected object format.
func (objectFormat ObjectFormat) Sum(data []byte) ObjectID {
	return objectFormat.details().sum(data)
}

// Zero returns the all-zero object ID for this object format.
func (objectFormat ObjectFormat) Zero() ObjectID {
	return ObjectID{
		objectFormat: objectFormat,
		data:         [maxObjectIDSize]byte{},
	}
}
