package ingest

import (
	"io"
	"os"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

const (
	defaultDeltaBaseCacheMaxBytes = 32 << 20
)

// ingestState holds mutable state for one Ingest call.
type ingestState struct {
	src         io.Reader
	destination *os.Root
	algo        objectid.Algorithm
	opts        Options

	packHeaderRaw [packHeaderSize]byte

	packFile    *os.File
	packTmpName string
	idxFile     *os.File
	idxTmpName  string
	revFile     *os.File
	revTmpName  string

	stream *streamScanner

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
	opts Options,
	header HeaderInfo,
	headerRaw [packHeaderSize]byte,
) (*ingestState, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	return &ingestState{
		src:               src,
		destination:       destination,
		algo:              algo,
		opts:              opts,
		packHeaderRaw:     headerRaw,
		objectCountHeader: header.ObjectCount,
		offsetToRecord:    make(map[uint64]int),
		objectToRecord:    make(map[objectid.ObjectID]int),
		baseCache:         newDeltaBaseCache(defaultDeltaBaseCacheMaxBytes),
	}, nil
}
