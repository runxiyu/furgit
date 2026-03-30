package reading

import (
	"fmt"

	deltaapply "codeberg.org/lindenii/furgit/format/packfile/delta/apply"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

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
