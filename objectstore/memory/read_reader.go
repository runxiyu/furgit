package memory

import (
	"bytes"
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadReaderFull reads one full object through a reader.
func (store *Store) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	raw, err := store.ReadBytesFull(id)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(raw)), nil
}

// ReadReaderContent reads one object body through a reader.
func (store *Store) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	ty, content, err := store.ReadBytesContent(id)
	if err != nil {
		return objecttype.TypeInvalid, 0, nil, err
	}

	return ty, int64(len(content)), io.NopCloser(bytes.NewReader(content)), nil
}
