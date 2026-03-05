package ingest

import (
	"io"
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

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
	fixThin bool,
	writeRev bool,
	base objectstore.Store,
) (Result, error) {
	state, err := newIngestState(src, destination, algo, fixThin, writeRev, base)
	if err != nil {
		return Result{}, err
	}

	return ingest(state)
}
