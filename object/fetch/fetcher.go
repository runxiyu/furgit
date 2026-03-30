package fetch

import objectstore "codeberg.org/lindenii/furgit/object/store"

// Fetcher resolves parsed and streamed objects from an object store.
//
// Labels: MT-Safe.
type Fetcher struct {
	store objectstore.Reader
}

// New returns a Fetcher that reads objects from store.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(store objectstore.Reader) *Fetcher {
	return &Fetcher{store: store}
}
