package packed

import (
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ReadHeader reads an object's type and declared content size.
func (store *Store) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	loc, err := store.lookup(id)
	if err != nil {
		return objecttype.TypeInvalid, 0, err
	}
	return store.resolveHeaderAt(loc)
}
