package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

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
