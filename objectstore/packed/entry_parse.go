package packed

import (
	"fmt"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
	objecttype "codeberg.org/lindenii/furgit/object/type"
	packfmt "codeberg.org/lindenii/furgit/packfile"
)

// entryMeta describes one parsed pack entry header.
type entryMeta struct {
	// ty is the pack entry type tag.
	ty objecttype.Type
	// size is the declared resulting content size.
	size int64
	// dataOffset points to the zlib payload start.
	dataOffset int
	// baseRefID is set for ref-delta entries.
	baseRefID objectid.ObjectID
	// baseOfs is set for ofs-delta entries.
	baseOfs uint64
}

// parseEntryMeta parses one pack entry header at offset.
func parseEntryMeta(pack *packFile, algo objectid.Algorithm, offset uint64) (entryMeta, error) {
	var zero entryMeta
	if offset >= uint64(len(pack.data)) {
		return zero, fmt.Errorf("objectstore/packed: pack %q offset %d out of bounds", pack.name, offset)
	}

	pos, err := intconv.Uint64ToInt(offset)
	if err != nil {
		return zero, fmt.Errorf("objectstore/packed: pack %q offset conversion: %w", pack.name, err)
	}

	entry, err := packfmt.ParseEntry(pack.data[pos:], algo.Size())
	if err != nil {
		return zero, fmt.Errorf("objectstore/packed: pack %q: %w", pack.name, err)
	}

	meta := entryMeta{
		ty:         entry.Type,
		size:       entry.Size,
		dataOffset: pos + entry.DataOffset,
	}
	switch meta.ty {
	case objecttype.TypeRefDelta:
		baseID, err := objectid.FromBytes(algo, entry.RefBaseID)
		if err != nil {
			return zero, fmt.Errorf("objectstore/packed: pack %q invalid ref-delta base id: %w", pack.name, err)
		}

		meta.baseRefID = baseID
	case objecttype.TypeOfsDelta:
		if offset <= entry.OfsBaseDistance {
			return zero, fmt.Errorf("objectstore/packed: pack %q has invalid ofs-delta base", pack.name)
		}

		meta.baseOfs = offset - entry.OfsBaseDistance
	case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
		// Base object types do not have delta base metadata.
	case objecttype.TypeInvalid, objecttype.TypeFuture:
		return zero, fmt.Errorf("objectstore/packed: pack %q has unsupported entry type %d", pack.name, meta.ty)
	default:
		return zero, fmt.Errorf("objectstore/packed: pack %q has unsupported entry type %d", pack.name, meta.ty)
	}

	return meta, nil
}
