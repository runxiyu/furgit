package packed

import (
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/objecttype"
)

// deltaFrame describes one delta payload to apply during reconstruction.
type deltaFrame struct {
	// packName identifies where the delta payload lives.
	packName string
	// dataOffset points to the start of the delta zlib payload in pack.
	dataOffset int
}

// deltaPlan describes how to reconstruct one requested object.
type deltaPlan struct {
	// declaredSize is the target object's declared content size.
	declaredSize int64
	// baseLoc points to the innermost base object.
	baseLoc location
	// baseType is the canonical object type resolved from baseLoc.
	baseType objecttype.Type
	// frames contains deltas from target down toward base.
	frames []deltaFrame
}

// deltaPlanFor walks one object's chain and builds a delta reconstruction plan.
func (store *Store) deltaPlanFor(start location) (deltaPlan, error) {
	visited := make(map[location]struct{})
	current := start

	var plan deltaPlan
	plan.declaredSize = -1

	for {
		if _, ok := visited[current]; ok {
			return deltaPlan{}, fmt.Errorf("objectstore/packed: delta cycle while resolving object")
		}
		visited[current] = struct{}{}

		pack, meta, err := store.entryMetaAt(current)
		if err != nil {
			return deltaPlan{}, err
		}
		if plan.declaredSize < 0 {
			if isBaseObjectType(meta.ty) {
				plan.declaredSize = meta.size
			} else {
				declaredSize, err := deltaDeclaredSizeAt(pack, meta.dataOffset)
				if err != nil {
					return deltaPlan{}, err
				}
				plan.declaredSize = declaredSize
			}
		}

		if isBaseObjectType(meta.ty) {
			plan.baseLoc = current
			plan.baseType = meta.ty
			return plan, nil
		}

		switch meta.ty {
		case objecttype.TypeRefDelta:
			plan.frames = append(plan.frames, deltaFrame{
				packName:   current.packName,
				dataOffset: meta.dataOffset,
			})
			next, err := store.lookup(meta.baseRefID)
			if err != nil {
				return deltaPlan{}, err
			}
			current = next
		case objecttype.TypeOfsDelta:
			plan.frames = append(plan.frames, deltaFrame{
				packName:   current.packName,
				dataOffset: meta.dataOffset,
			})
			current = location{
				packName: current.packName,
				offset:   meta.baseOfs,
			}
		case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
			return deltaPlan{}, fmt.Errorf("objectstore/packed: internal invariant violation for base type %d", meta.ty)
		case objecttype.TypeInvalid, objecttype.TypeFuture:
			return deltaPlan{}, fmt.Errorf("objectstore/packed: unsupported pack type %d", meta.ty)
		default:
			return deltaPlan{}, fmt.Errorf("objectstore/packed: unsupported pack type %d", meta.ty)
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

	_, size, err := readDeltaHeaderSizes(reader)
	if err != nil {
		return 0, err
	}
	return int64(size), nil
}

// readDeltaHeaderSizes reads the first two varints in one inflated delta stream.
func readDeltaHeaderSizes(reader io.Reader) (int, int, error) {
	// Two Git varints are read here. Each can take up to 10 bytes.
	var buf [20]byte
	n := 0

	for {
		if n >= len(buf) {
			return 0, 0, fmt.Errorf("objectstore/packed: malformed delta varint")
		}
		if _, err := io.ReadFull(reader, buf[n:n+1]); err != nil {
			return 0, 0, fmt.Errorf("objectstore/packed: malformed delta varint: %w", err)
		}
		n++
		if buf[n-1]&0x80 == 0 {
			break
		}
	}
	pos := 0
	srcSize, err := readDeltaVarint(buf[:n], &pos)
	if err != nil {
		return 0, 0, err
	}

	for {
		if n >= len(buf) {
			return 0, 0, fmt.Errorf("objectstore/packed: malformed delta varint")
		}
		if _, err := io.ReadFull(reader, buf[n:n+1]); err != nil {
			return 0, 0, fmt.Errorf("objectstore/packed: malformed delta varint: %w", err)
		}
		n++
		if buf[n-1]&0x80 == 0 {
			break
		}
	}
	dstSize, err := readDeltaVarint(buf[:n], &pos)
	if err != nil {
		return 0, 0, err
	}
	return srcSize, dstSize, nil
}
