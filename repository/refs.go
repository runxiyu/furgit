package repository

import "codeberg.org/lindenii/furgit/refstore"

// Refs returns the configured ref store.
//
//nolint:ireturn
func (repo *Repository) Refs() refstore.ReadWriteStore {
	return repo.refs
}
