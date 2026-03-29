package repository

import "codeberg.org/lindenii/furgit/object/fetch"

// Fetcher returns an object fetcher backed by the repository's object store.
//
// Use Fetcher when you want typed commits, trees, blobs, or tags, when you
// need to peel through annotated tags, or when you want path-based access
// within trees.
//
// Labels: Life-Parent.
func (repo *Repository) Fetcher() *fetch.Fetcher {
	return repo.fetcher
}
