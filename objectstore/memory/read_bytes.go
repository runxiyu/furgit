package memory

import (
	"codeberg.org/lindenii/furgit/object/header"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ReadBytesFull reads one full object, including the object header.
func (store *Store) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	obj, ok := store.objects[id]
	if !ok {
		return nil, objectstore.ErrObjectNotFound
	}

	header, ok := header.Encode(obj.ty, int64(len(obj.content)))
	if !ok {
		panic("failed to encode object header")
	}

	raw := make([]byte, len(header)+len(obj.content))
	copy(raw, header)
	copy(raw[len(header):], obj.content)

	return raw, nil
}

// ReadBytesContent reads one object body.
func (store *Store) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	obj, ok := store.objects[id]
	if !ok {
		return objecttype.TypeInvalid, nil, objectstore.ErrObjectNotFound
	}

	return obj.ty, append([]byte(nil), obj.content...), nil
}
