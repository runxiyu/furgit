// Package objectid provides utilities around object IDs and hash algorithms.
package objectid

import (
	//#nosec G505

	"encoding/hex"
	"fmt"
)

// ObjectID represents a Git object ID.
//
//nolint:recvcheck
type ObjectID struct {
	algo Algorithm
	data [maxObjectIDSize]byte
}

// Algorithm returns the object ID's hash algorithm.
func (id ObjectID) Algorithm() Algorithm {
	return id.algo
}

// Size returns the object ID size in bytes.
func (id ObjectID) Size() int {
	return id.algo.Size()
}

// String returns the canonical hex representation.
func (id ObjectID) String() string {
	size := id.Size()

	return hex.EncodeToString(id.data[:size])
}

// Bytes returns a copy of the object ID bytes.
func (id ObjectID) Bytes() []byte {
	size := id.Size()

	return append([]byte(nil), id.data[:size]...)
}

// RawBytes returns a direct byte slice view of the object ID bytes.
//
// The returned slice aliases the object ID's internal storage. Callers MUST
// treat it as read-only and MUST NOT modify its contents.
//
// Use Bytes when an independent copy is required.
func (id *ObjectID) RawBytes() []byte {
	size := id.Size()

	return id.data[:size:size]
}

// ParseHex parses an object ID from hex for the specified algorithm.
func ParseHex(algo Algorithm, s string) (ObjectID, error) {
	var id ObjectID
	if algo.Size() == 0 {
		return id, ErrInvalidAlgorithm
	}

	if len(s)%2 != 0 {
		return id, fmt.Errorf("%w: odd hex length %d", ErrInvalidObjectID, len(s))
	}

	if len(s) != algo.HexLen() {
		return id, fmt.Errorf("%w: got %d chars, expected %d", ErrInvalidObjectID, len(s), algo.HexLen())
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return id, fmt.Errorf("%w: decode: %w", ErrInvalidObjectID, err)
	}

	copy(id.data[:], decoded)
	id.algo = algo

	return id, nil
}

// FromBytes builds an object ID from raw bytes for the specified algorithm.
func FromBytes(algo Algorithm, b []byte) (ObjectID, error) {
	var id ObjectID
	if algo.Size() == 0 {
		return id, ErrInvalidAlgorithm
	}

	if len(b) != algo.Size() {
		return id, fmt.Errorf("%w: got %d bytes, expected %d", ErrInvalidObjectID, len(b), algo.Size())
	}

	copy(id.data[:], b)
	id.algo = algo

	return id, nil
}
