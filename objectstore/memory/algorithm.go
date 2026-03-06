package memory

import "codeberg.org/lindenii/furgit/objectid"

// Algorithm returns the object ID algorithm used by the store.
func (store *Store) Algorithm() objectid.Algorithm {
	return store.algo
}
