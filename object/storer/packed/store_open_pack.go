package packed

// openPack returns one opened and validated pack handle.
func (store *Store) openPack(name string) (*packFile, error) {
	store.stateMu.RLock()

	pack, ok := store.packs[name]
	if ok {
		store.stateMu.RUnlock()

		return pack, nil
	}

	store.stateMu.RUnlock()

	file, err := store.root.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	pack, err = openPackFile(name, file, info.Size())
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	err = store.verifyPackMatchesIndexes(pack)
	if err != nil {
		_ = pack.close()

		return nil, err
	}

	store.stateMu.Lock()

	existing, ok := store.packs[name]
	if ok {
		store.stateMu.Unlock()

		_ = pack.close()

		return existing, nil
	}

	store.packs[name] = pack
	store.stateMu.Unlock()

	return pack, nil
}
