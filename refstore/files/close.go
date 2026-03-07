package files

// Close releases resources associated with the store.
func (store *Store) Close() error {
	err := store.gitRoot.Close()
	commonErr := store.commonRoot.Close()

	if err != nil {
		return err
	}

	return commonErr
}
