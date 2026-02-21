package packed

import (
	"fmt"

	packfmt "codeberg.org/lindenii/furgit/format/pack"
	"codeberg.org/lindenii/furgit/objecttype"
)

// deltaResolveBase materializes the base object body for one delta plan.
func (store *Store) deltaResolveBase(plan deltaPlan) (objecttype.Type, []byte, error) {
	cacheKey := deltaBaseKey{
		packName: plan.baseLoc.packName,
		offset:   plan.baseLoc.offset,
	}

	store.cacheMu.RLock()
	if ty, content, ok := store.deltaCache.get(cacheKey); ok {
		store.cacheMu.RUnlock()
		return ty, content, nil
	}
	store.cacheMu.RUnlock()

	pack, meta, err := store.entryMetaAt(plan.baseLoc)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}
	if !packfmt.IsBaseObjectType(meta.ty) {
		return objecttype.TypeInvalid, nil, fmt.Errorf("objectstore/packed: delta plan base is not a base object")
	}
	base, err := inflateAt(pack, meta.dataOffset, meta.size)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	store.cacheMu.Lock()
	store.deltaCache.add(cacheKey, meta.ty, base)
	store.cacheMu.Unlock()
	return meta.ty, base, nil
}
