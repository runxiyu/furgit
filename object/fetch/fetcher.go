package fetch

import objectstore "codeberg.org/lindenii/furgit/object/store"

// Fetcher resolves parsed and streamed objects from an object store.
//
// Labels: MT-Safe.
type Fetcher struct {
	store objectstore.ReadingStore
}

// New returns a Fetcher that reads objects from store.
//
// Labels: Deps-Borrowed.
func New(store objectstore.ReadingStore) *Fetcher {
	return &Fetcher{store: store}
}
