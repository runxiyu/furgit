package memory

// Refresh is a no-op for in-memory object stores.
func (store *Store) Refresh() error {
	return nil
}
