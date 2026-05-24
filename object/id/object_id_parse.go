package id

import (
	"encoding/hex"
	"fmt"
)

// FromBytes builds an object ID from raw bytes for the specified object format.
func FromBytes(objectFormat ObjectFormat, b []byte) (ObjectID, error) {
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

// FromHex parses an object ID from hex for the specified object format.
func FromHex(objectFormat ObjectFormat, s string) (ObjectID, error) {
	var id ObjectID
	if objectFormat.Size() == 0 {
		return id, ErrInvalidObjectFormat
	}

	if len(s)%2 != 0 {
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
