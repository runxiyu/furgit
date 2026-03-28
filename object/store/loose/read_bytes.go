package loose

import (
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// readBytesParsed reads, inflates, and parses a loose object in one pass.
// It returns the full raw payload and its parsed type and content.
func (store *Store) readBytesParsed(id objectid.ObjectID) ([]byte, objecttype.Type, []byte, error) {
	file, err := store.openObject(id)
	if err != nil {
		return nil, objecttype.TypeInvalid, nil, err
	}

	defer func() { _ = file.Close() }()

	raw, err := decodeAll(file)
	if err != nil {
		return nil, objecttype.TypeInvalid, nil, err
	}

	ty, content, err := parseRaw(raw)
	if err != nil {
		return nil, objecttype.TypeInvalid, nil, err
	}

	return raw, ty, content, nil
}

// ReadBytesFull reads a full serialized object as "type size\0content".
//
// It inflates and parses the full loose object, including verifying the zlib
// Adler-32 trailer.
func (store *Store) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	raw, _, _, err := store.readBytesParsed(id)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// ReadBytesContent reads an object's type and content bytes.
//
// Like ReadBytesFull, it inflates and parses the full loose object, including
// verifying the zlib Adler-32 trailer.
func (store *Store) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	_, ty, content, err := store.readBytesParsed(id)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	return ty, content, nil
}
