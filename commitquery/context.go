// Package commitquery answers commit ancestry and merge-base queries.
package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// Query owns the mutable node arena for commit-domain queries over one object
// store.
type Query struct {
	store objectstore.ReadingStore
	graph *commitgraphread.Reader

	nodes []node

	byOID      map[objectid.ObjectID]nodeIndex
	byGraphPos map[commitgraphread.Position]nodeIndex

	markPhase uint32
	touched   []nodeIndex
}

// New builds one reusable commit query arena over one object store and optional
// commit-graph reader.
func New(store objectstore.ReadingStore, graph *commitgraphread.Reader) *Query {
	return &Query{
		store:      store,
		graph:      graph,
		byOID:      make(map[objectid.ObjectID]nodeIndex),
		byGraphPos: make(map[commitgraphread.Position]nodeIndex),
	}
}
