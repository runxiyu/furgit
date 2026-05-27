package memory

import (
	objectheader "lindenii.org/go/furgit/object/header"
	objectid "lindenii.org/go/furgit/object/id"
	objectstore "lindenii.org/go/furgit/object/store"
	objecttype "lindenii.org/go/furgit/object/type"
)

// ReadBytesFull reads one full object, including the object header.
func (store *Store) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	obj, ok := store.objects[id]
	if !ok {
		return nil, objectstore.ErrObjectNotFound
	}

	header, ok := objectheader.Encode(obj.ty, int64(len(obj.content)))
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
