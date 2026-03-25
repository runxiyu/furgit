package files

// Close releases resources associated with the store.
//
// Store borrows gitRoot, so Close does not close it.
// Transactions and batches borrowing the store are invalid after Close.
//
// Repeated calls to Close are undefined behavior.
func (store *Store) Close() error {
	return store.commonRoot.Close()
}
