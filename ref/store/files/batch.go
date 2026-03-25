package files

import refstore "codeberg.org/lindenii/furgit/ref/store"

type Batch struct {
	store *Store
	ops   []queuedUpdate
}

var _ refstore.Batch = (*Batch)(nil)
