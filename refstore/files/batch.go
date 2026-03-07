package files

import "codeberg.org/lindenii/furgit/refstore"

type Batch struct {
	store  *Store
	ops    []txOp
	closed bool
}

var _ refstore.Batch = (*Batch)(nil)
