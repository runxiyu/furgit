package files

import (
	refstore "lindenii.org/go/furgit/ref/store"
)

// Transaction stages files-store updates for one atomic commit.
type Transaction struct {
	store *Store
	ops   []queuedUpdate
}

var _ refstore.Transaction = (*Transaction)(nil)
