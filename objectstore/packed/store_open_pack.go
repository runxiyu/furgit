package packed

import "fmt"

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

// verifyPackMatchesIndexes checks that one opened pack's trailer hash matches
// every loaded index that references the same pack name.
func (store *Store) verifyPackMatchesIndexes(pack *packFile) error {
	err := store.ensureCandidates()
	if err != nil {
		return err
	}

	candidate, ok := store.candidateForPack(pack.name)
	if !ok {
		return fmt.Errorf("objectstore/packed: missing index for pack %q", pack.name)
	}

	index, err := store.openIndex(candidate)
	if err != nil {
		return err
	}

	err = verifyMappedPackMatchesMappedIdx(pack.data, index.data, store.algo)
	if err != nil {
		return fmt.Errorf("objectstore/packed: pack %q does not match idx %q: %w", pack.name, index.idxName, err)
	}

	return nil
}

// entryMetaAt parses one pack entry header at location.
func (store *Store) entryMetaAt(loc location) (*packFile, entryMeta, error) {
	pack, err := store.openPack(loc.packName)
	if err != nil {
		return nil, entryMeta{}, err
	}

	meta, err := parseEntryMeta(pack, store.algo, loc.offset)
	if err != nil {
		return nil, entryMeta{}, err
	}

	return pack, meta, nil
}
