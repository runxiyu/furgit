package reading

import objecttype "codeberg.org/lindenii/furgit/object/type"

// deltaResolveContent resolves one object's content bytes from its pack location.
func (store *Store) deltaResolveContent(start location) (objecttype.Type, []byte, error) {
	chain, err := store.deltaBuildChain(start)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	pack, meta, err := store.entryMetaAt(start)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	declaredSize := meta.size
	if !meta.ty.IsBaseObject() {
		declaredSize, err = deltaDeclaredSizeAt(pack, meta.dataOffset)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}
	}

	return store.deltaResolveChain(chain, declaredSize)
}
