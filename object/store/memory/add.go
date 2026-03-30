package memory

import (
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// AddObject stores one object body and returns its object ID.
//
//go:fix inline
func (store *Store) AddObject(ty objecttype.Type, body []byte) objectid.ObjectID {
	id, err := store.WriteBytesContent(ty, body)
	if err != nil {
		panic(err)
	}

	return id
}
