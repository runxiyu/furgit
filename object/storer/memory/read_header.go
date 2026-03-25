package memory

import (
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstorer "codeberg.org/lindenii/furgit/object/storer"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadHeader reads one object header.
func (store *Store) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	obj, ok := store.objects[id]
	if !ok {
		return objecttype.TypeInvalid, 0, objectstorer.ErrObjectNotFound
	}

	return obj.ty, int64(len(obj.content)), nil
}
