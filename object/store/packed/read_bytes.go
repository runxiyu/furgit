package packed

import (
	"fmt"

	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadBytesContent reads an object's type and content bytes.
//
// It fully resolves the requested object bytes. For base pack entries, this
// includes verifying that the zlib stream inflates to exactly the declared
// object size and reaches its Adler-32 trailer.
func (store *Store) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	loc, err := store.lookup(id)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	return store.deltaResolveContent(loc)
}

// ReadBytesFull reads a full serialized object as "type size\0content".
//
// Like ReadBytesContent, it fully resolves the requested object bytes. For
// base pack entries, this includes verifying that the zlib stream inflates to
// exactly the declared object size and reaches its Adler-32 trailer.
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
