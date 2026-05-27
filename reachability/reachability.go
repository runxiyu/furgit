package reachability

import (
	commitgraphread "lindenii.org/go/furgit/format/commitgraph/read"
	objectfetch "lindenii.org/go/furgit/object/fetch"
)

// Reachability provides graph traversal over objects in one object store.
//
// Labels: MT-Safe.
type Reachability struct {
	fetcher *objectfetch.Fetcher
	graph   *commitgraphread.Reader
}

// New builds a Reachability over one object fetcher with an optional
// commit-graph reader for faster commit-domain traversal.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(fetcher *objectfetch.Fetcher, graph *commitgraphread.Reader) *Reachability {
	return &Reachability{fetcher: fetcher, graph: graph}
}
