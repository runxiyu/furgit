package files

import (
	"codeberg.org/lindenii/furgit/ref/store"
)

type Transaction struct {
	store *Store
	ops   []queuedUpdate
}

var _ refstore.Transaction = (*Transaction)(nil)
