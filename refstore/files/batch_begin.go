package files

import "codeberg.org/lindenii/furgit/refstore"

// BeginBatch creates one new files batch.
//
//nolint:ireturn
func (store *Store) BeginBatch() (refstore.Batch, error) {
	return &Batch{
		store: store,
		ops:   make([]txOp, 0, 8),
	}, nil
}
