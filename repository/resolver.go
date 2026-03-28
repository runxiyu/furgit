package repository

import "codeberg.org/lindenii/furgit/object/fetch"

// Fetcher returns an object fetcher backed by the repository's object store.
//
// Labels: Life-Parent, Close-No.
func (repo *Repository) Fetcher() *fetch.Fetcher {
	return fetch.New(repo.objects)
}
