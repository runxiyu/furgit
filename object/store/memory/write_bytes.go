package memory

import (
	"bytes"

	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// WriteBytesContent writes one typed object content byte slice.
func (store *Store) WriteBytesContent(ty objecttype.Type, content []byte) (objectid.ObjectID, error) {
	id := store.algo.Sum(buildRawObject(ty, content))
	store.objects[id] = storedObject{ty: ty, content: append([]byte(nil), content...)}

	return id, nil
}

// WriteBytesFull writes one full serialized object byte slice as "type size\0content".
func (store *Store) WriteBytesFull(raw []byte) (objectid.ObjectID, error) {
	return store.WriteReaderFull(bytes.NewReader(raw))
}

func buildRawObject(ty objecttype.Type, body []byte) []byte {
	header, ok := objectheader.Encode(ty, int64(len(body)))
	if !ok {
		panic("failed to encode object header")
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)

	return raw
}
