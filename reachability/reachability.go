// Package reachability traverses the object graph to test relationships and emit object lists.
package reachability

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// Reachability provides graph traversal over objects in one object store.
//
// Labels: MT-Unsafe.
type Reachability struct {
	store objectstore.ReadingStore
	graph *commitgraphread.Reader
}

// New builds a Reachability  over one object store.
//
// Labels: Deps-Borrowed.
func New(store objectstore.ReadingStore) *Reachability {
	return &Reachability{store: store}
}

// NewWithCommitGraph builds a Reachability over one object store with an
// optional commit-graph reader for faster commit-domain traversal.
//
// Labels: Deps-Borrowed.
func NewWithCommitGraph(store objectstore.ReadingStore, graph *commitgraphread.Reader) *Reachability {
	return &Reachability{store: store, graph: graph}
}
