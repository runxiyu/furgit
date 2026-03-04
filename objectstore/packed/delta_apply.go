package packed

import (
	"fmt"

	deltaapply "codeberg.org/lindenii/furgit/format/delta/apply"
	packfmt "codeberg.org/lindenii/furgit/format/pack"
	"codeberg.org/lindenii/furgit/objecttype"
)

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
	if !packfmt.IsBaseObjectType(meta.ty) {
		declaredSize, err = deltaDeclaredSizeAt(pack, meta.dataOffset)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}
	}

	return store.deltaResolveChain(chain, declaredSize)
}

// deltaResolveChain resolves one object chain into content bytes.
func (store *Store) deltaResolveChain(chain deltaChain, declaredSize int64) (objecttype.Type, []byte, error) {
	ty, out, nextDelta, err := store.deltaResolveChainStart(chain)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	for i := nextDelta; i >= 0; i-- {
		node := chain.deltas[i]

		pack, err := store.openPack(node.loc.packName)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}

		delta, err := inflateAt(pack, node.dataOffset, -1)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}

		out, err = deltaapply.Apply(out, delta)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}

		store.cacheMu.Lock()
		store.deltaCache.add(
			deltaBaseKey{packName: node.loc.packName, offset: node.loc.offset},
			ty,
			out,
		)
		store.cacheMu.Unlock()
	}

	if int64(len(out)) != declaredSize {
		return objecttype.TypeInvalid, nil, fmt.Errorf(
			"objectstore/packed: resolved content size mismatch: got %d want %d",
			len(out),
			declaredSize,
		)
	}

	if ty != chain.baseType {
		return objecttype.TypeInvalid, nil, fmt.Errorf(
			"objectstore/packed: resolved content type mismatch: got %d want %d",
			ty,
			chain.baseType,
		)
	}

	return ty, out, nil
}

// deltaResolveChainStart finds the nearest cached chain node or inflates the
// innermost base object. It returns the starting bytes and the next delta index
// to apply in reverse order.
func (store *Store) deltaResolveChainStart(chain deltaChain) (objecttype.Type, []byte, int, error) {
	for i, node := range chain.deltas {
		store.cacheMu.RLock()
		ty, out, ok := store.deltaCache.get(
			deltaBaseKey{packName: node.loc.packName, offset: node.loc.offset},
		)
		store.cacheMu.RUnlock()

		if ok {
			return ty, out, i - 1, nil
		}
	}

	store.cacheMu.RLock()
	ty, out, ok := store.deltaCache.get(
		deltaBaseKey{packName: chain.baseLoc.packName, offset: chain.baseLoc.offset},
	)
	store.cacheMu.RUnlock()

	if ok {
		return ty, out, len(chain.deltas) - 1, nil
	}

	pack, meta, err := store.entryMetaAt(chain.baseLoc)
	if err != nil {
		return objecttype.TypeInvalid, nil, 0, err
	}

	if !packfmt.IsBaseObjectType(meta.ty) {
		return objecttype.TypeInvalid, nil, 0, fmt.Errorf("objectstore/packed: delta chain base is not a base object")
	}

	base, err := inflateAt(pack, meta.dataOffset, meta.size)
	if err != nil {
		return objecttype.TypeInvalid, nil, 0, err
	}

	store.cacheMu.Lock()
	store.deltaCache.add(
		deltaBaseKey{packName: chain.baseLoc.packName, offset: chain.baseLoc.offset},
		meta.ty,
		base,
	)
	store.cacheMu.Unlock()

	return meta.ty, base, len(chain.deltas) - 1, nil
}
