// Package reachability traverses the object graph to test relationships and emit object lists.
package reachability

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// Reachability provides graph traversal over objects in one object store.
//
// It is not safe for concurrent use.
type Reachability struct {
	store objectstore.Store
	graph *commitgraphread.Reader
}

// New builds a Reachability  over one object store.
func New(store objectstore.Store) *Reachability {
	return &Reachability{store: store}
}

// NewWithCommitGraph builds a Reachability over one object store with an
// optional commit-graph reader for faster commit-domain traversal.
func NewWithCommitGraph(store objectstore.Store, graph *commitgraphread.Reader) *Reachability {
	return &Reachability{store: store, graph: graph}
}
