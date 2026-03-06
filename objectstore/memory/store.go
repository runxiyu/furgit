package memory

import (
	"codeberg.org/lindenii/furgit/objectid"
)

// Store is one in-memory object store.
type Store struct {
	algo    objectid.Algorithm
	objects map[objectid.ObjectID]storedObject
}

// New builds one empty in-memory store for one object format.
func New(algo objectid.Algorithm) *Store {
	return &Store{
		algo:    algo,
		objects: make(map[objectid.ObjectID]storedObject),
	}
}

// Close closes the in-memory store.
func (store *Store) Close() error {
	return nil
}
