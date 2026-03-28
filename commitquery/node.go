package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// nodeIndex identifies one internal query node.
type nodeIndex int

// node stores one mutable commit traversal node.
type node struct {
	id objectid.ObjectID

	parents []nodeIndex

	commitTime int64
	generation uint64

	hasGeneration bool
	hasGraphPos   bool
	loaded        bool

	graphPos commitgraphread.Position
	marks    markBits

	touchedPhase uint32
}

// newNode allocates one empty internal node.
func (query *query) newNode(id objectid.ObjectID) nodeIndex {
	count := len(query.nodes)

	idx := nodeIndex(count)

	query.nodes = append(query.nodes, node{id: id})

	return idx
}
