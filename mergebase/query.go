package mergebase

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Query builds one single-use merge-base query over two commit roots.
//
// Both inputs are peeled through annotated tags before commit traversal.
func Query(
	store objectstore.Store,
	graph *commitgraphread.Reader,
	left objectid.ObjectID,
	right objectid.ObjectID,
) *Bases {
	return &Bases{
		store: store,
		graph: graph,
		left:  left,
		right: right,
	}
}
