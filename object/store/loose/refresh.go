package loose

// Refresh is a no-op for loose object stores.
func (store *Store) Refresh() error {
	return nil
}
