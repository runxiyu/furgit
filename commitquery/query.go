package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// query stores one mutable reusable worker and its cached node arena.
//
// Labels: MT-Unsafe.
type query struct {
	store objectstore.ReadingStore
	graph *commitgraphread.Reader

	nodes []node

	byOID      map[objectid.ObjectID]nodeIndex
	byGraphPos map[commitgraphread.Position]nodeIndex

	markPhase uint32
	touched   []nodeIndex
}
