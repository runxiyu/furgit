package loose

import (
	"bytes"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// WriteBytesFull writes a full serialized object as "type size\0content".
func (store *Store) WriteBytesFull(raw []byte) (objectid.ObjectID, error) {
	return store.WriteReaderFull(bytes.NewReader(raw))
}

// WriteBytesContent writes typed content bytes as a loose object.
func (store *Store) WriteBytesContent(ty objecttype.Type, content []byte) (objectid.ObjectID, error) {
	return store.WriteReaderContent(ty, int64(len(content)), bytes.NewReader(content))
}
