package files

import refstore "codeberg.org/lindenii/furgit/ref/store"

// Batch stages files-store updates for one non-atomic apply.
type Batch struct {
	store *Store
	ops   []queuedUpdate
}

var _ refstore.Batch = (*Batch)(nil)
