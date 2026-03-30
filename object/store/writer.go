package objectstore

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ObjectWriter writes individual Git objects.
type ObjectWriter interface {
	// WriteReaderContent writes one typed object content stream.
	WriteReaderContent(ty objecttype.Type, size int64, src io.Reader) (objectid.ObjectID, error)

	// WriteReaderFull writes one full serialized object stream as "type size\0content".
	WriteReaderFull(src io.Reader) (objectid.ObjectID, error)

	// WriteBytesContent writes one typed object content byte slice.
	WriteBytesContent(ty objecttype.Type, content []byte) (objectid.ObjectID, error)

	// WriteBytesFull writes one full serialized object byte slice as "type size\0content".
	WriteBytesFull(raw []byte) (objectid.ObjectID, error)
}

// PackWriteOptions controls one pack write operation.
type PackWriteOptions struct{}

// PackWriter writes Git pack streams.
type PackWriter interface {
	// WritePack ingests one pack stream.
	WritePack(src io.Reader, opts PackWriteOptions) error
}
