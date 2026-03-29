package files

// Close releases resources associated with the store.
//
// Labels: MT-Unsafe.
func (store *Store) Close() error {
	return store.commonRoot.Close()
}
