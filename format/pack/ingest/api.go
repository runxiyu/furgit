package ingest

import (
	"io"
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Options controls one pack ingest operation.
type Options struct {
	// FixThin appends missing local bases for thin packs.
	FixThin bool
	// WriteRev writes a .rev alongside the .pack and .idx.
	WriteRev bool
	// Base supplies existing objects for thin-pack fixup.
	Base objectstore.Store
	// RequireTrailingEOF requires the source to hit EOF after the pack trailer.
	//
	// This is suitable for exact pack-file readers, but should be disabled for
	// full-duplex transport streams like receive-pack where the peer keeps the
	// connection open to read the server response.
	RequireTrailingEOF bool
}

// Result describes one successful ingest transaction.
type Result struct {
	// PackName is the destination-relative filename of the written .pack.
	PackName string
	// IdxName is the destination-relative filename of the written .idx.
	IdxName string
	// RevName is the destination-relative filename of the written .rev.
	//
	// RevName is empty when writeRev is false.
	RevName string
	// PackHash is the final pack hash (same hash embedded in .idx/.rev trailers).
	PackHash objectid.ObjectID
	// ObjectCount is the final object count in the resulting pack.
	//
	// If thin fixup appends objects, this includes appended base objects.
	ObjectCount uint32
	// ThinFixed reports whether thin fixup appended local bases.
	ThinFixed bool
}

// Ingest ingests one pack stream from src into destination.
//
// Ingest performs streaming pack read/write/verification, delta resolution,
// optional thin fixup, then writes .idx and optionally .rev.
//
// destination ownership and lifecycle are managed by the caller.
// Ingest does not perform quarantine promotion/migration.
func Ingest(
	src io.Reader,
	destination *os.Root,
	algo objectid.Algorithm,
	opts Options,
) (Result, error) {
	state, err := newIngestState(src, destination, algo, opts)
	if err != nil {
		return Result{}, err
	}

	return ingest(state)
}
