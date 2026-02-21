package packed

import (
	"fmt"

	deltaapply "codeberg.org/lindenii/furgit/format/delta/apply"
	packfmt "codeberg.org/lindenii/furgit/format/pack"
	"codeberg.org/lindenii/furgit/objecttype"
)

// deltaNode describes one delta object in a reconstruction chain.
type deltaNode struct {
	// loc identifies the delta object's pack location.
	loc location
	// dataOffset points to the start of the delta zlib payload in pack.
	dataOffset int
}

// deltaChain describes how to reconstruct one requested object.
type deltaChain struct {
	// baseLoc points to the innermost base object.
	baseLoc location
	// baseType is the canonical object type resolved from baseLoc.
	baseType objecttype.Type
	// deltas contains delta objects from target down toward base.
	deltas []deltaNode
}

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

		if packfmt.IsBaseObjectType(meta.ty) {
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

// deltaDeclaredSizeAt returns the resolved object size declared by one delta
// stream header at dataOffset.
func deltaDeclaredSizeAt(pack *packFile, dataOffset int) (int64, error) {
	reader, err := zlibReaderAt(pack, dataOffset)
	if err != nil {
		return 0, err
	}
	defer func() { _ = reader.Close() }()

	_, size, err := deltaapply.ReadHeaderSizes(reader)
	if err != nil {
		return 0, err
	}
	return int64(size), nil
}
