package repository

import (
	objectloose "codeberg.org/lindenii/furgit/objectstore/loose"
)

func (repo *Repository) LooseStoreForWriting() *objectloose.Store {
	return repo.objectsLoose
}
