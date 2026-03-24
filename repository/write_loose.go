package repository

import (
	objectloose "codeberg.org/lindenii/furgit/objectstore/loose"
)

// LooseStoreForWriting returns the repository's loose-object writer.
//
// The returned store is owned by Repository and borrows repository-managed
// resources. Callers must not close it directly, and it must not be used after
// Close.
func (repo *Repository) LooseStoreForWriting() *objectloose.Store {
	return repo.objectsLoose
}
