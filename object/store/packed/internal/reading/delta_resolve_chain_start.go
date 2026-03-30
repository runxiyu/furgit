package reading

import (
	"fmt"

	objecttype "codeberg.org/lindenii/furgit/object/type"
)

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

	if !meta.ty.IsBaseObject() {
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
