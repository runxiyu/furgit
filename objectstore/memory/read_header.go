package memory

import (
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadHeader reads one object header.
func (store *Store) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	obj, ok := store.objects[id]
	if !ok {
		return objecttype.TypeInvalid, 0, objectstore.ErrObjectNotFound
	}

	return obj.ty, int64(len(obj.content)), nil
}
