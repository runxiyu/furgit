package memory

import objectid "lindenii.org/go/furgit/object/id"

// Algorithm returns the object ID algorithm used by the store.
func (store *Store) Algorithm() objectid.Algorithm {
	return store.algo
}
