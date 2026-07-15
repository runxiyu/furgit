package repository

import "lindenii.org/go/furgit/object/fetch"

// Fetcher returns an object fetcher backed by the object store
// of the repository.
//
// Labels: Life-Parent.
func (repo *Repository) Fetcher() *fetch.Fetcher {
	return repo.fetcher
}
