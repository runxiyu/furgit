package repository

import "codeberg.org/lindenii/furgit/object/fetch"

// Fetcher returns an object fetcher backed by the repository's object store.
//
// The returned fetcher is ready for use, borrows the repository's object
// store, does not need closing, and must not be used after Close.
func (repo *Repository) Fetcher() *fetch.Fetcher {
	return fetch.New(repo.objects)
}
