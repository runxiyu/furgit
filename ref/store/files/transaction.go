package files

import (
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

// Transaction stages files-store updates for one atomic commit.
type Transaction struct {
	store *Store
	ops   []queuedUpdate
}

var _ refstore.Transaction = (*Transaction)(nil)
