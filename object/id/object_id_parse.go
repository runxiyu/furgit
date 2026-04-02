package id

import (
	"encoding/hex"
	"fmt"
)

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

// FromHex parses an object ID from hex for the specified algorithm.
func FromHex(algo Algorithm, s string) (ObjectID, error) {
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
