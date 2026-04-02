package store

import (
	"io"

	"codeberg.org/lindenii/furgit/common/iowrap"
	"codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/typ"
)

// Writer represents a store
// that could perform both pack ingestions
// and individual object writes.
type Writer interface {
	PackWriter
	ObjectWriter
}

// ObjectWriter writes individual Git objects.
type ObjectWriter interface {
	// WriteReaderContent writes one typed object content stream.
	WriteReaderContent(ty typ.Type, size int64, src io.Reader) (id.ObjectID, error)

	// WriteReaderFull writes one full serialized object stream as "type size\0content".
	WriteReaderFull(src io.Reader) (id.ObjectID, error)

	// WriteBytesContent writes one typed object content byte slice.
	WriteBytesContent(ty typ.Type, content []byte) (id.ObjectID, error)

	// WriteBytesFull writes one full serialized object byte slice as "type size\0content".
	WriteBytesFull(raw []byte) (id.ObjectID, error)
}

// PackWriter writes Git pack streams.
type PackWriter interface {
	// WritePack ingests one pack stream,
	// such that the objects contained therein
	// become available in the relevant store.
	WritePack(src io.Reader, opts PackWriteOptions) error
}

// PackWriteOptions controls one pack write operation.
type PackWriteOptions struct {
	// ThinBase supplies the wider object reader
	// used to complete thin packs during ingestion.
	//
	// This is an option for the write operation
	// rather than on a particular pack-backed store
	// because any pack-accepting store
	// is not generally expected to know
	// the entire repository object universe around it.
	// In a normal repository,
	// thin bases usually come from a broader view
	// such as mix(loose, packed),
	// and should not be treated as
	// a property of the destination pack-accepting store.
	// Thus, in almost all pack-ingesting operations
	// where thin bases are relevant,
	// a thin base reader would be required,
	// and hence it is included here.
	//
	// When nil,
	// external thin-base repair is disabled,
	// and unresolved thin deltas fail ingestion.
	//
	// TODO: Define the errors here?
	ThinBase Reader

	// Progress receives human-readable progress messages
	// for packfile ingestion.
	//
	// When nil, no progress output is emitted.
	Progress iowrap.WriteFlusher
}
