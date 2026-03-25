package fetch

import objectstorer "codeberg.org/lindenii/furgit/object/storer"

// Fetcher resolves parsed and streamed objects from an object store.
//
// A Fetcher does not take ownership of the store and does not close it.
type Fetcher struct {
	store objectstorer.Store
}

// New returns a Fetcher that reads objects from store.
//
// The returned Fetcher does not take ownership of store.
func New(store objectstorer.Store) *Fetcher {
	return &Fetcher{store: store}
}
