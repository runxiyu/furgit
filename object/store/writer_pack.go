package objectstore

import (
	"io"

	"lindenii.org/go/furgit/common/iowrap"
)

// PackWriteOptions controls one pack write operation.
type PackWriteOptions struct {
	// ThinBase supplies the wider object reader used to complete thin packs
	// during ingestion.
	//
	// This is an option for the write operation rather than on a particular
	// pack-backed store because any pack-accepting store is not generally
	// expected to know the entire repository object universe around it.
	// In a normal repository, thin bases usually come from a broader view
	// such as mix(loose, packed), and should not be treated as a property of
	// the destination pack-accepting store. Thus, in almost all pack-ingesting
	// operations, a thin base reader would be required, and hence it is
	// included here.
	//
	// When nil, external thin-base repair is disabled and unresolved thin deltas
	// fail ingestion.
	ThinBase Reader

	// Progress receives human-readable progress messages.
	//
	// When nil, no progress output is emitted.
	Progress iowrap.WriteFlusher

	// RequireTrailingEOF requires the source to hit EOF after the pack trailer.
	//
	// This is suitable for exact pack-file readers, but should be disabled for
	// full-duplex transport streams like receive-pack where the peer keeps the
	// connection open to read the server response.
	RequireTrailingEOF bool
}

// PackWriter writes Git pack streams.
type PackWriter interface {
	// WritePack ingests one pack stream.
	WritePack(src io.Reader, opts PackWriteOptions) error
}

// PackQuarantine represents one quarantined pack-wise write.
type PackQuarantine interface {
	BaseQuarantine
	PackWriter
}

// PackQuarantineOptions controls the options for one pack quarantine creation.
type PackQuarantineOptions struct{}

// PackQuarantiner creates quarantines for pack-wise writes.
type PackQuarantiner interface {
	BeginPackQuarantine(opts PackQuarantineOptions) (PackQuarantine, error)
}
