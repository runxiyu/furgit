// Package commitquery provides private commit-domain query routines.
package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/commitgraph/read"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Context owns the mutable node arena for one commit query.
type Context struct {
	store objectstore.Store
	graph *commitgraphread.Reader

	nodes []node

	byOID      map[objectid.ObjectID]NodeIndex
	byGraphPos map[commitgraphread.Position]NodeIndex

	markPhase uint32
	touched   []NodeIndex
}

// NewContext builds one empty query context over one object store and optional commit-graph reader.
func NewContext(store objectstore.Store, graph *commitgraphread.Reader) *Context {
	return &Context{
		store:      store,
		graph:      graph,
		byOID:      make(map[objectid.ObjectID]NodeIndex),
		byGraphPos: make(map[commitgraphread.Position]NodeIndex),
	}
}
