package ingest

import "lindenii.org/go/furgit/object/id"

// Result describes one finalized pack write.
type Result struct {
	// PackName is the destination-relative name of the written pack.
	PackName string

	// IdxName is the destination-relative name of the written index.
	IdxName string

	// RevName is the destination-relative name of the written reverse index.
	RevName string

	// PackHash is the pack trailer hash
	// shared by the pack, index, and reverse index.
	PackHash id.ObjectID

	// ObjectCount is the number of objects in the finalized pack,
	// including any bases appended during thin completion.
	ObjectCount uint32

	// ThinFixed reports whether thin completion appended local bases.
	ThinFixed bool
}
