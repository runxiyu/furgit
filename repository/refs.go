package repository

import refstore "codeberg.org/lindenii/furgit/ref/store"

// Refs returns the configured ref store.
//
// Labels: Life-Parent, Close-No.
//
//nolint:ireturn
func (repo *Repository) Refs() refstore.ReadWriteStore {
	return repo.refs
}
