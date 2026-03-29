package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

type query struct {
	store objectstore.ReadingStore
	graph *commitgraphread.Reader

	nodes []node

	byOID      map[objectid.ObjectID]nodeIndex
	byGraphPos map[commitgraphread.Position]nodeIndex

	markPhase uint32
	touched   []nodeIndex
}

func newQuery(store objectstore.ReadingStore, graph *commitgraphread.Reader) *query {
	return &query{
		store:      store,
		graph:      graph,
		byOID:      make(map[objectid.ObjectID]nodeIndex),
		byGraphPos: make(map[commitgraphread.Position]nodeIndex),
	}
}

func (query *query) resetForReuse() {
	for _, idx := range query.touched {
		query.nodes[idx].marks = 0
	}

	query.touched = query.touched[:0]
}
