package repository

import "codeberg.org/lindenii/furgit/object/fetch"

// Fetcher returns an object fetcher backed by the repository's object store.
//
// The returned fetcher is ready for use, borrows the repository's object
// store, must not be closed directly, and must not be used after repository Close.
func (repo *Repository) Fetcher() *fetch.Fetcher {
	return fetch.New(repo.objects)
}
