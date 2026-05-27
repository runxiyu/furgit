package commitquery

import (
	commitgraphread "lindenii.org/go/furgit/format/commitgraph/read"
	objectfetch "lindenii.org/go/furgit/object/fetch"
	objectid "lindenii.org/go/furgit/object/id"
)

// newQuery builds one empty mutable worker over one object fetcher and graph.
//
// Labels: Deps-Borrowed, Life-Parent.
func newQuery(fetcher *objectfetch.Fetcher, graph *commitgraphread.Reader) *query {
	return &query{
		fetcher:    fetcher,
		graph:      graph,
		byOID:      make(map[objectid.ObjectID]nodeIndex),
		byGraphPos: make(map[commitgraphread.Position]nodeIndex),
	}
}
