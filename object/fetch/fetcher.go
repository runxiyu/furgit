package fetch

import objectstore "lindenii.org/go/furgit/object/store"

// Fetcher provides ordinary object access above an object store.
//
// It exposes object metadata, typed object loading, tree-ish and commit-ish
// peeling, path resolution, one-tree fs views, and blob content streaming.
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
