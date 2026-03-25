package packed

import (
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
	packfmt "codeberg.org/lindenii/furgit/packfile"
)

// ReadSize reads an object's declared content size.
//
// Like ReadHeader, it resolves header metadata only. It does not verify that
// the full pack entry payload is readable and does not verify any zlib
// Adler-32 trailer for compressed entry data.
func (store *Store) ReadSize(id objectid.ObjectID) (int64, error) {
	loc, err := store.lookup(id)
	if err != nil {
		return 0, err
	}

	return store.resolveSizeAt(loc)
}

// resolveSizeAt resolves one object's declared content size from location.
func (store *Store) resolveSizeAt(start location) (int64, error) {
	pack, meta, err := store.entryMetaAt(start)
	if err != nil {
		return 0, err
	}

	if packfmt.IsBaseObjectType(meta.ty) {
		return meta.size, nil
	}

	switch meta.ty {
	case objecttype.TypeRefDelta, objecttype.TypeOfsDelta:
		return deltaDeclaredSizeAt(pack, meta.dataOffset)
	case objecttype.TypeInvalid, objecttype.TypeFuture:
		return 0, fmt.Errorf("objectstore/packed: unsupported pack type %d", meta.ty)
	case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
		return 0, fmt.Errorf("objectstore/packed: internal invariant violation for base type %d", meta.ty)
	default:
		return 0, fmt.Errorf("objectstore/packed: unsupported pack type %d", meta.ty)
	}
}
