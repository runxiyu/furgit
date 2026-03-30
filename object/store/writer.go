package objectstore

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ObjectWriter writes individual Git objects.
type ObjectWriter interface {
	// WriteContent writes one typed object content stream.
	WriteContent(ty objecttype.Type, size int64, src io.Reader) (objectid.ObjectID, error)

	// WriteFull writes one full serialized object stream as "type size\0content".
	WriteFull(src io.Reader) (objectid.ObjectID, error)
}

// PackWriteOptions controls one pack write operation.
type PackWriteOptions struct{}

// PackWriter writes Git pack streams.
type PackWriter interface {
	// WritePack ingests one pack stream.
	WritePack(src io.Reader, opts PackWriteOptions) error
}
