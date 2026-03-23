package files

import "codeberg.org/lindenii/furgit/refstore"

type Batch struct {
	store *Store
	ops   []queuedUpdate
}

var _ refstore.Batch = (*Batch)(nil)
