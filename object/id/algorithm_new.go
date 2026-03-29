package objectid

import "hash"

// New returns a new hash.Hash for this algorithm.
func (algo Algorithm) New() (hash.Hash, error) {
	newFn := algo.info().new
	if newFn == nil {
		return nil, ErrInvalidAlgorithm
	}

	return newFn(), nil
}
