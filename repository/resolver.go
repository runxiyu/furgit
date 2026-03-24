package repository

import "codeberg.org/lindenii/furgit/object/resolve"

// Resolver returns an object resolver backed by the repository's object store.
//
// The returned resolver is ready for use, borrows the repository's object
// store, does not need closing, and must not be used after Close.
func (repo *Repository) Resolver() *resolve.Resolver {
	return resolve.New(repo.objects)
}
