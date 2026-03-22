package packed

// Close releases mapped pack/index resources associated with the store.
//
// Store borrows its root, so Close does not close it.
// Close releases cached pack/index mappings retained by the store.
//
// Repeated calls to Close are undefined behavior.
func (store *Store) Close() error {
	store.stateMu.Lock()
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

	return closeErr
}
