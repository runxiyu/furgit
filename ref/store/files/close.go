package files

// Close releases resources associated with the store.
func (store *Store) Close() error {
	return store.commonRoot.Close()
}
