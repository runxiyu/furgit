package store

import (
	"context"
	"errors"
	"io"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
	"lindenii.org/go/lgo/iowrap"
)

// ErrInvalidObject indicates a malformed object passed to a write.
var ErrInvalidObject = errors.New("object/store: invalid object")

// ErrObjectTooLarge indicates that an object exceeds
// the size limit configured for the write.
var ErrObjectTooLarge = errors.New("object/store: object too large")

// ObjectWriter writes individual Git objects.
type ObjectWriter interface {
	// WriteBytesFull writes one full serialized object byte slice as "type size\x00content".
	WriteBytesFull(raw []byte) (id.ObjectID, error)

	// WriteBytesContent writes one typed object content byte slice.
	WriteBytesContent(ty typ.Type, content []byte) (id.ObjectID, error)

	// WriteReaderFull writes one full serialized object stream as "type size\x00content".
	WriteReaderFull(src io.Reader) (id.ObjectID, error)

	// WriteReaderContent writes one typed object content stream.
	WriteReaderContent(ty typ.Type, size int, src io.Reader) (id.ObjectID, error)
}

// PackWriter writes Git pack streams.
type PackWriter interface {
	// WritePack ingests one pack stream,
	// such that the objects contained therein
	// become available in the relevant store.
	WritePack(ctx context.Context, src io.Reader, opts PackWriteOptions) error
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
	ThinBase ObjectReader

	// Progress receives human-readable progress messages
	// for packfile ingestion.
	//
	// When nil, no progress output is emitted.
	Progress iowrap.WriteFlusher

	// MaxObjectSize rejects ingestion of any object
	// whose declared inflated size or delta result size exceeds it,
	// bounding the memory spent reconstructing a single object.
	//
	// Zero or negative means no limit.
	MaxObjectSize int
}
