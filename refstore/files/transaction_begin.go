package files

import "codeberg.org/lindenii/furgit/refstore"

// BeginTransaction creates one new files transaction.
//
//nolint:ireturn
func (store *Store) BeginTransaction() (refstore.Transaction, error) {
	return &Transaction{
		store: store,
		ops:   make([]txOp, 0, 8),
	}, nil
}
