package memory

import (
	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// AddObject stores one object body and returns its object ID.
func (store *Store) AddObject(ty objecttype.Type, body []byte) objectid.ObjectID {
	header, ok := objectheader.Encode(ty, int64(len(body)))
	if !ok {
		panic("failed to encode object header")
	}

	raw := append(append([]byte(nil), header...), body...)
	id := store.algo.Sum(raw)
	store.objects[id] = storedObject{ty: ty, content: append([]byte(nil), body...)}

	return id
}
