package files

import (
	"codeberg.org/lindenii/furgit/refstore"
)

type Transaction struct {
	store *Store
	ops   []txOp
}

var _ refstore.Transaction = (*Transaction)(nil)
