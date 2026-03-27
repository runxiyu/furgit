package fetch

import objectstore "codeberg.org/lindenii/furgit/object/store"

// Fetcher resolves parsed and streamed objects from an object store.
//
// A Fetcher does not take ownership of the store and does not close it.
type Fetcher struct {
	store objectstore.Store
}

// New returns a Fetcher that reads objects from store.
//
// The returned Fetcher does not take ownership of store.
func New(store objectstore.Store) *Fetcher {
	return &Fetcher{store: store}
}
