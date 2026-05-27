package commitquery

import (
	commitgraphread "lindenii.org/go/furgit/format/commitgraph/read"
	objectfetch "lindenii.org/go/furgit/object/fetch"
	objectid "lindenii.org/go/furgit/object/id"
)

// query stores one mutable reusable worker and its cached node arena.
//
// Labels: MT-Unsafe.
type query struct {
	fetcher *objectfetch.Fetcher
	graph   *commitgraphread.Reader

	nodes []node

	byOID      map[objectid.ObjectID]nodeIndex
	byGraphPos map[commitgraphread.Position]nodeIndex

	markPhase uint32
	touched   []nodeIndex
}
