package ingest

import (
	"io"
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

const (
	defaultDeltaBaseCacheMaxBytes = 32 << 20
)

// ingestState holds mutable state for one Ingest call.
type ingestState struct {
	src         io.Reader
	destination *os.Root
	algo        objectid.Algorithm
	fixThin     bool
	writeRev    bool
	base        objectstore.Store

	packFile    *os.File
	packTmpName string
	idxFile     *os.File
	idxTmpName  string
	revFile     *os.File
	revTmpName  string

	stream *streamCopier

	records             []objectRecord
	ofsDeltas           []ofsDeltaRef
	refDeltas           []refDeltaRef
	unresolvedRefDeltas []int
	offsetToRecord      map[uint64]int
	objectToRecord      map[objectid.ObjectID]int

	baseCache *deltaBaseCache
	packHash  objectid.ObjectID

	objectCountHeader uint32
	thinFixed         bool
}

// newIngestState constructs one call-local ingest state.
func newIngestState(
	src io.Reader,
	destination *os.Root,
	algo objectid.Algorithm,
	fixThin bool,
	writeRev bool,
	base objectstore.Store,
) (*ingestState, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	return &ingestState{
		src:            src,
		destination:    destination,
		algo:           algo,
		fixThin:        fixThin,
		writeRev:       writeRev,
		base:           base,
		offsetToRecord: make(map[uint64]int),
		objectToRecord: make(map[objectid.ObjectID]int),
		baseCache:      newDeltaBaseCache(defaultDeltaBaseCacheMaxBytes),
	}, nil
}
