package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	"codeberg.org/lindenii/furgit/objectid"
)

// NodeIndex identifies one internal query node.
type NodeIndex int

// node stores one mutable commit traversal node.
type node struct {
	id objectid.ObjectID

	parents []NodeIndex

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
func (ctx *Context) newNode(id objectid.ObjectID) NodeIndex {
	count := len(ctx.nodes)

	idx := NodeIndex(count)

	ctx.nodes = append(ctx.nodes, node{id: id})

	return idx
}
