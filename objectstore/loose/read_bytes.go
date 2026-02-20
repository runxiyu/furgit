package loose

import (
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ReadBytesFull reads a full serialized object as "type size\0content".
func (store *Store) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	file, err := store.openObject(id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return decodeAll(file)
}

// ReadBytesContent reads an object's type and content bytes.
func (store *Store) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	raw, err := store.ReadBytesFull(id)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}
	ty, content, err := parseRaw(raw)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}
	return ty, content, nil
}
