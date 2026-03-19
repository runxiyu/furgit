package resolve

import "codeberg.org/lindenii/furgit/objectstore"

// Resolver resolves parsed and streamed objects from an object store.
//
// A Resolver does not take ownership of the store and does not close it.
type Resolver struct {
	store objectstore.Store
}

// New returns a Resolver that reads objects from store.
//
// The returned Resolver does not take ownership of store.
func New(store objectstore.Store) *Resolver {
	return &Resolver{store: store}
}
