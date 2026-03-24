package repository

import "codeberg.org/lindenii/furgit/refstore"

// Refs returns the configured ref store.
//
// The returned store is owned by Repository and borrows repository-managed
// resources. Callers must not close it directly, and it must not be used after
// Close.
//
//nolint:ireturn
func (repo *Repository) Refs() refstore.ReadWriteStore {
	return repo.refs
}
