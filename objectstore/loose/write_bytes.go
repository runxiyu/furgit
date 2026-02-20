package loose

import (
	"bytes"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// WriteBytesFull writes a full serialized object as "type size\\x00content".
func (store *Store) WriteBytesFull(raw []byte) (objectid.ObjectID, error) {
	var zero objectid.ObjectID

	writer, finalize, err := store.WriteWriterFull()
	if err != nil {
		return zero, err
	}
	if _, err := bytes.NewReader(raw).WriteTo(writer); err != nil {
		_ = writer.Close()
		return zero, err
	}
	if err := writer.Close(); err != nil {
		return zero, err
	}
	return finalize()
}

// WriteBytesContent writes typed content bytes as a loose object.
func (store *Store) WriteBytesContent(ty objecttype.Type, content []byte) (objectid.ObjectID, error) {
	var zero objectid.ObjectID

	writer, finalize, err := store.WriteWriterContent(ty, int64(len(content)))
	if err != nil {
		return zero, err
	}
	if _, err := bytes.NewReader(content).WriteTo(writer); err != nil {
		_ = writer.Close()
		return zero, err
	}
	if err := writer.Close(); err != nil {
		return zero, err
	}
	return finalize()
}
