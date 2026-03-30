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

// ObjectQuarantine represents one quarantined object-wise write.
type ObjectQuarantine interface {
	Quarantine
	ObjectWriter
}

// ObjectQuarantineOptions controls the options for one object quarantine creation.
type ObjectQuarantineOptions struct{}

// ObjectQuarantiner creates quarantines for object-wise writes.
type ObjectQuarantiner interface {
	BeginObjectQuarantine(opts ObjectQuarantineOptions) (ObjectQuarantine, error)
}
