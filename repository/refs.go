package repository

import refstore "codeberg.org/lindenii/furgit/ref/store"

// Refs returns the configured ref store.
//
// Use Refs when starting from branch names, tags, HEAD, or other references.
// A common pattern is to resolve a reference first and then pass the resulting
// object ID to [Repository.Fetcher] or [Repository.Objects].
//
// Labels: Life-Parent.
//
//nolint:ireturn
func (repo *Repository) Refs() interface {
	refstore.ReadingStore
	refstore.TransactionalStore
	refstore.BatchStore
} {
	return repo.refs
}
