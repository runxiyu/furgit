package objectid

import "fmt"

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
