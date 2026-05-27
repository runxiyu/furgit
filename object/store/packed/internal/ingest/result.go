package ingest

import objectid "lindenii.org/go/furgit/object/id"

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
