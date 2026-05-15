package id

import "hash"

// HexLen returns the encoded hexadecimal length.
func (algo Algorithm) HexLen() int {
	return algo.Size() * 2
}

// Size returns the hash size in bytes.
func (algo Algorithm) Size() int {
	return algo.details().size
}

// New returns a new hash.Hash for this algorithm.
func (algo Algorithm) New() (hash.Hash, error) {
	newFn := algo.details().new
	if newFn == nil {
		return nil, ErrInvalidAlgorithm
	}

	return newFn(), nil
}

// String returns the canonical algorithm name.
func (algo Algorithm) String() string {
	return algo.details().name
}

// Sum computes an object ID from raw data using the selected algorithm.
func (algo Algorithm) Sum(data []byte) ObjectID {
	return algo.details().sum(data)
}

// Zero returns the all-zero object ID for this algorithm.
func (algo Algorithm) Zero() ObjectID {
	return ObjectID{
		algo: algo,
		data: [maxObjectIDSize]byte{},
	}
}
