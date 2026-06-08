package loose

import (
	"bytes"
	"fmt"
	"io"

	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// WriteBytesFull writes a full serialized object as "type size\0content".
func (loose *Loose) WriteBytesFull(raw []byte) (id.ObjectID, error) {
	return loose.WriteReaderFull(bytes.NewReader(raw))
}

// WriteBytesContent writes typed content bytes as a loose object.
func (loose *Loose) WriteBytesContent(ty typ.Type, content []byte) (id.ObjectID, error) {
	return loose.WriteReaderContent(ty, uint64(len(content)), bytes.NewReader(content))
}

// WriteReaderContent writes one loose object from typed content bytes read from src.
// src must provide exactly size bytes.
// size is required because loose object headers are "type size\0content",
// so the header must be emitted before streaming content without buffering.
func (loose *Loose) WriteReaderContent(ty typ.Type, size uint64, src io.Reader) (id.ObjectID, error) {
	headerBytes := header.Append(nil, ty, size)

	writer, err := loose.newStreamWriter(false)
	if err != nil {
		return id.ObjectID{}, err
	}

	writer.headerDone = true
	writer.expectedContentLeft = size

	err = writer.writeRawChunk(headerBytes)
	if err != nil {
		_ = writer.Close()
		_ = loose.root.Remove(writer.tmpRelPath)

		return id.ObjectID{}, err
	}

	return writeReaderIntoStreamWriter(writer, src)
}

// WriteReaderFull writes one loose object from raw bytes "type size\0content" read from src.
func (loose *Loose) WriteReaderFull(src io.Reader) (id.ObjectID, error) {
	writer, err := loose.newStreamWriter(true)
	if err != nil {
		return id.ObjectID{}, err
	}

	return writeReaderIntoStreamWriter(writer, src)
}

// writeReaderIntoStreamWriter copies src into writer and publishes the object.
func writeReaderIntoStreamWriter(writer *streamWriter, src io.Reader) (id.ObjectID, error) {
	_, err := io.Copy(writer, src)
	if err != nil {
		_ = writer.Close()
		_ = writer.loose.root.Remove(writer.tmpRelPath)

		return id.ObjectID{}, fmt.Errorf("object/store/loose: %w", err)
	}

	err = writer.Close()
	if err != nil {
		_ = writer.loose.root.Remove(writer.tmpRelPath)

		return id.ObjectID{}, err
	}

	objectID, err := writer.finalize()
	if err != nil {
		_ = writer.loose.root.Remove(writer.tmpRelPath)

		return id.ObjectID{}, err
	}

	return objectID, nil
}
