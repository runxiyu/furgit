package packed

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ReadBytesContent reads an object's type and content bytes.
func (store *Store) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	loc, err := store.lookup(id)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	return store.deltaResolveContent(loc)
}

// ReadBytesFull reads a full serialized object as "type size\0content".
func (store *Store) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	ty, content, err := store.ReadBytesContent(id)
	if err != nil {
		return nil, err
	}

	header, ok := objectheader.Encode(ty, int64(len(content)))
	if !ok {
		return nil, fmt.Errorf("objectstore/packed: failed to encode object header for type %d", ty)
	}

	out := make([]byte, len(header)+len(content))
	copy(out, header)
	copy(out[len(header):], content)

	return out, nil
}
