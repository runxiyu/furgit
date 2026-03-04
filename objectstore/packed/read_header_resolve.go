package packed

import (
	"fmt"

	packfmt "codeberg.org/lindenii/furgit/format/pack"
	"codeberg.org/lindenii/furgit/objecttype"
)

// resolveHeaderAt resolves one object's canonical type and declared content size.
func (store *Store) resolveHeaderAt(start location) (objecttype.Type, int64, error) {
	visited := make(map[location]struct{})
	current := start
	declaredSize := int64(-1)

	for {
		if _, ok := visited[current]; ok {
			return objecttype.TypeInvalid, 0, fmt.Errorf("objectstore/packed: delta cycle while resolving object header")
		}

		visited[current] = struct{}{}

		pack, meta, err := store.entryMetaAt(current)
		if err != nil {
			return objecttype.TypeInvalid, 0, err
		}

		if declaredSize < 0 {
			if packfmt.IsBaseObjectType(meta.ty) {
				declaredSize = meta.size
			} else {
				size, err := deltaDeclaredSizeAt(pack, meta.dataOffset)
				if err != nil {
					return objecttype.TypeInvalid, 0, err
				}

				declaredSize = size
			}
		}

		if packfmt.IsBaseObjectType(meta.ty) {
			return meta.ty, declaredSize, nil
		}

		switch meta.ty {
		case objecttype.TypeRefDelta:
			next, err := store.lookup(meta.baseRefID)
			if err != nil {
				return objecttype.TypeInvalid, 0, err
			}

			current = next
		case objecttype.TypeOfsDelta:
			current = location{
				packName: current.packName,
				offset:   meta.baseOfs,
			}
		case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
			return objecttype.TypeInvalid, 0, fmt.Errorf("objectstore/packed: internal invariant violation for base type %d", meta.ty)
		case objecttype.TypeInvalid, objecttype.TypeFuture:
			return objecttype.TypeInvalid, 0, fmt.Errorf("objectstore/packed: unsupported pack type %d", meta.ty)
		default:
			return objecttype.TypeInvalid, 0, fmt.Errorf("objectstore/packed: unsupported pack type %d", meta.ty)
		}
	}
}
