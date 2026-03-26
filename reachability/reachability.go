// Package reachability traverses the object graph to test relationships and emit object lists.
package reachability

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectstorer "codeberg.org/lindenii/furgit/object/storer"
)

// Reachability provides graph traversal over objects in one object store.
//
// It is not safe for concurrent use.
type Reachability struct {
	store objectstorer.Store
	graph *commitgraphread.Reader
}

// New builds a Reachability  over one object store.
func New(store objectstorer.Store) *Reachability {
	return &Reachability{store: store}
}

// NewWithCommitGraph builds a Reachability over one object store with an
// optional commit-graph reader for faster commit-domain traversal.
func NewWithCommitGraph(store objectstorer.Store, graph *commitgraphread.Reader) *Reachability {
	return &Reachability{store: store, graph: graph}
}
