package reading

import (
	objectid "lindenii.org/go/furgit/object/id"
	objecttype "lindenii.org/go/furgit/object/type"
)

// ReadHeader reads an object's type and declared content size.
//
// It resolves header metadata only. It does not verify that the full pack entry
// payload is readable and does not verify any zlib Adler-32 trailer for
// compressed entry data.
func (store *Store) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	loc, err := store.lookup(id)
	if err != nil {
		return objecttype.TypeInvalid, 0, err
	}

	return store.resolveHeaderAt(loc)
}
