package repository

import "codeberg.org/lindenii/furgit/object/resolve"

// Resolver returns an object resolver backed by the repository's object store.
//
// The returned resolver is ready for use and does not take ownership of the
// repository or its underlying object store.
func (repo *Repository) Resolver() *resolve.Resolver {
	return resolve.New(repo.objects)
}
