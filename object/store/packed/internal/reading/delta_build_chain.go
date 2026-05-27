package reading

import (
	"fmt"

	objecttype "lindenii.org/go/furgit/object/type"
)

// deltaBuildChain walks one object's chain and builds a reconstruction chain.
func (store *Store) deltaBuildChain(start location) (deltaChain, error) {
	visited := make(map[location]struct{})
	current := start

	var chain deltaChain

	for {
		if _, ok := visited[current]; ok {
			return deltaChain{}, fmt.Errorf("objectstore/packed: delta cycle while resolving object")
		}

		visited[current] = struct{}{}

		_, meta, err := store.entryMetaAt(current)
		if err != nil {
			return deltaChain{}, err
		}

		if meta.ty.IsBaseObject() {
			chain.baseLoc = current
			chain.baseType = meta.ty

			return chain, nil
		}

		switch meta.ty {
		case objecttype.TypeRefDelta:
			chain.deltas = append(chain.deltas, deltaNode{
				loc:        current,
				dataOffset: meta.dataOffset,
			})

			next, err := store.lookup(meta.baseRefID)
			if err != nil {
				return deltaChain{}, err
			}

			current = next
		case objecttype.TypeOfsDelta:
			chain.deltas = append(chain.deltas, deltaNode{
				loc:        current,
				dataOffset: meta.dataOffset,
			})
			current = location{
				packName: current.packName,
				offset:   meta.baseOfs,
			}
		case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
			return deltaChain{}, fmt.Errorf("objectstore/packed: internal invariant violation for base type %d", meta.ty)
		case objecttype.TypeInvalid, objecttype.TypeFuture:
			return deltaChain{}, fmt.Errorf("objectstore/packed: unsupported pack type %d", meta.ty)
		default:
			return deltaChain{}, fmt.Errorf("objectstore/packed: unsupported pack type %d", meta.ty)
		}
	}
}
