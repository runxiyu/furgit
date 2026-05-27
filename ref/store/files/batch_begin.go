package files

import refstore "lindenii.org/go/furgit/ref/store"

// BeginBatch creates one new files batch.
//
//nolint:ireturn
func (store *Store) BeginBatch() (refstore.Batch, error) {
	return &Batch{
		store: store,
		ops:   make([]queuedUpdate, 0, 8),
	}, nil
}
