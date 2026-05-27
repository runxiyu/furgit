package repository

import refstore "codeberg.org/lindenii/furgit/ref/store"

// RefStore returns the configured ref store.
//
// Use RefStore when starting from branch names, tags, HEAD, or other references.
// A common pattern is to resolve a reference first and then pass the resulting
// object ID to [Repository.Fetcher] or [Repository.ObjectStore].
//
// Labels: Life-Parent.
//
//nolint:ireturn
func (repo *Repository) RefStore() interface {
	refstore.Reader
	refstore.Transactioner
	refstore.Batcher
} {
	return repo.refs
}
