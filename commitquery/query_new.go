package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// newQuery builds one empty mutable worker over one object store and graph.
//
// Labels: Deps-Borrowed, Life-Parent.
func newQuery(store objectstore.ReadingStore, graph *commitgraphread.Reader) *query {
	return &query{
		store:      store,
		graph:      graph,
		byOID:      make(map[objectid.ObjectID]nodeIndex),
		byGraphPos: make(map[commitgraphread.Position]nodeIndex),
	}
}
