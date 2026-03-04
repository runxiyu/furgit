package loose

import (
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// WriteReaderContent writes one loose object from typed content bytes read from src.
// src must provide exactly size bytes.
// size is required because loose object headers are "type size\0content", so the
// header must be emitted before streaming content without buffering.
func (store *Store) WriteReaderContent(ty objecttype.Type, size int64, src io.Reader) (objectid.ObjectID, error) {
	if size < 0 {
		return objectid.ObjectID{}, fmt.Errorf("objectstore/loose: negative content size: %d", size)
	}

	header, ok := objectheader.Encode(ty, size)
	if !ok {
		return objectid.ObjectID{}, fmt.Errorf("objectstore/loose: failed to encode object header for type %v", ty)
	}

	writer, err := store.newStreamWriter(false)
	if err != nil {
		return objectid.ObjectID{}, err
	}

	writer.headerDone = true
	writer.expectedContentLeft = size

	err = writer.writeRawChunk(header)
	if err != nil {
		_ = writer.Close()
		_ = store.root.Remove(writer.tmpRelPath)

		return objectid.ObjectID{}, err
	}

	return writeReaderIntoStreamWriter(writer, src)
}

// WriteReaderFull writes one loose object from raw bytes "type size\0content"
// read from src.
func (store *Store) WriteReaderFull(src io.Reader) (objectid.ObjectID, error) {
	writer, err := store.newStreamWriter(true)
	if err != nil {
		return objectid.ObjectID{}, err
	}

	return writeReaderIntoStreamWriter(writer, src)
}

// writeReaderIntoStreamWriter copies src into writer and publishes the object.
func writeReaderIntoStreamWriter(writer *streamWriter, src io.Reader) (objectid.ObjectID, error) {
	_, err := io.Copy(writer, src)
	if err != nil {
		_ = writer.Close()
		_ = writer.store.root.Remove(writer.tmpRelPath)

		return objectid.ObjectID{}, err
	}

	err = writer.Close()
	if err != nil {
		_ = writer.store.root.Remove(writer.tmpRelPath)

		return objectid.ObjectID{}, err
	}

	id, err := writer.finalize()
	if err != nil {
		_ = writer.store.root.Remove(writer.tmpRelPath)

		return objectid.ObjectID{}, err
	}

	return id, nil
}
