package packed

// Close releases mapped pack/index resources associated with the store.
func (store *Store) Close() error {
	store.stateMu.Lock()

	if store.closed {
		store.stateMu.Unlock()

		return nil
	}

	store.closed = true
	root := store.root
	packs := store.packs
	store.stateMu.Unlock()
	store.idxMu.RLock()
	indexes := store.idxByPack
	store.idxMu.RUnlock()

	var closeErr error

	for _, pack := range packs {
		err := pack.close()
		if err != nil && closeErr == nil {
			closeErr = err
		}
	}

	for _, index := range indexes {
		err := index.close()
		if err != nil && closeErr == nil {
			closeErr = err
		}
	}

	store.cacheMu.Lock()
	store.deltaCache.clear()
	store.cacheMu.Unlock()

	err := root.Close()
	if err != nil && closeErr == nil {
		closeErr = err
	}

	return closeErr
}
