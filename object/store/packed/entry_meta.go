package packed

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
