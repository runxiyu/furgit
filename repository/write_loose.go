package repository

import (
	objectloose "codeberg.org/lindenii/furgit/object/store/loose"
)

// LooseStoreForWriting returns the repository's loose-object writer.
//
// Labels: Life-Parent, Close-No.
func (repo *Repository) LooseStoreForWriting() *objectloose.Store {
	return repo.objectsLoose
}
